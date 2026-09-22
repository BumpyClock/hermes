#!/usr/bin/env python3
"""Compare two configured-site extraction workloads, without fetching or DNS."""

import argparse
from datetime import datetime, timezone
import json
import os
from pathlib import Path
import shutil

import run


BASELINE_ADAPTER = '''package compatibility_test
import (
    "testing"

    "github.com/BumpyClock/hermes/internal/parser"
)
func configuredOptions(_ testing.TB) parser.ParserOptions { return *parser.DefaultParserOptions() }
func configuredExtractor(host string) string { return "custom:" + host }
'''

CANDIDATE_ADAPTER = '''package compatibility_test
import (
    "os"
    "testing"

    "github.com/BumpyClock/hermes/internal/definitions"
    "github.com/BumpyClock/hermes/internal/parser"
)
func configuredOptions(tb testing.TB) parser.ParserOptions {
    tb.Helper()
    snapshot, err := definitions.LoadDirectory(os.Getenv("HERMES_COMPAT_DEFINITIONS"))
    if err != nil { tb.Fatal(err) }
    for _, host := range []string{"www.nytimes.com", "arstechnica.com"} {
        if snapshot.Match(host) == nil { tb.Fatalf("definition not matched: %s", host) }
    }
    options := *parser.DefaultParserOptions()
    options.Definitions = snapshot
    options.DefinitionsConfigured = true
    return options
}
func configuredExtractor(host string) string {
    if host == "www.nytimes.com" { return "definition:nytimes" }
    return "definition:" + host
}
'''

EXPECTED_IDS = {f"{fixture}/{content_type}"
                for fixture in ("nytimes.html", "arstechnica.html")
                for content_type in ("html", "markdown", "text")}


def compare_observations(before, after):
    return run.exact_observations(before, after, EXPECTED_IDS)


def observations(log):
    result = {}
    for line in log.read_text().splitlines():
        if "OBSERVATION " not in line:
            continue
        row = json.loads(line.split("OBSERVATION ", 1)[1])
        if row["id"] in result:
            raise ValueError(f"duplicate configured observation: {row['id']}")
        result[row["id"]] = row
    return result


def validate_samples(benchmarks, count):
    baseline, candidate = benchmarks["baseline"], benchmarks["candidate"]
    if len(baseline) != 6 or baseline.keys() != candidate.keys():
        raise ValueError("expected six paired configured benchmark workloads")
    if any(row["samples"] != count for version in benchmarks.values() for row in version.values()):
        raise ValueError("configured benchmark sample count differs from requested count")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", default="v1.1.1")
    candidate_selection = parser.add_mutually_exclusive_group(required=True)
    candidate_selection.add_argument("--candidate-ref")
    candidate_selection.add_argument("--candidate-worktree", action="store_true",
                                     help="capture dirty tracked/unignored source; fail on runtime drift")
    parser.add_argument("--baseline-mode", choices=("compiled", "definitions"), default="compiled",
                        help="compiled release adapter (default), or same YAML adapter as candidate")
    parser.add_argument("--require-equivalent", action="store_true",
                        help="require exact pairwise outputs; requires --baseline-mode definitions")
    parser.add_argument("--functional-only", action="store_true",
                        help="build and capture configured outputs without running benchmarks")
    parser.add_argument("--definitions", type=Path, required=True, help="immutable whole snapshot root")
    parser.add_argument("--definitions-source", required=True, help="snapshot source revision")
    parser.add_argument("--output", required=True, help="new directory under .compatibility-runs")
    parser.add_argument("--samples", type=int, default=6)
    parser.add_argument("--benchtime", default="1s")
    parser.add_argument("--reference", type=Path,
                        help="previous configured capture; require exact candidate observations")
    args = parser.parse_args()
    if args.require_equivalent and args.baseline_mode != "definitions":
        parser.error("--require-equivalent requires --baseline-mode definitions")
    output = (run.ROOT / args.output).resolve()
    if args.samples < 5 or output.exists() or not output.is_relative_to(run.ROOT / ".compatibility-runs"):
        parser.error("require >=5 samples and a new output directory under .compatibility-runs")
    output.mkdir(parents=True)
    work = output / "build-work"
    work.mkdir()
    env = {**os.environ, "GOPROXY": "off", "GOSUMDB": "off", "GOTOOLCHAIN": "local",
           "GOFLAGS": "-mod=readonly", "GOMAXPROCS": "1", "TZ": "UTC",
           "HERMES_PARSER_DEBUG": "0", "GOTMPDIR": str(work), "TMPDIR": str(work)}
    metadata = {
        "baseline_sha": run.command(["git", "rev-parse", args.baseline + "^{commit}"]).strip(),
        "candidate_sha": run.command(["git", "rev-parse", (args.candidate_ref or "HEAD") + "^{commit}"]).strip(),
        "candidate_source": "git archive" if args.candidate_ref else "tracked and unignored worktree files",
        "candidate_status": "immutable archive" if args.candidate_ref else run.command(["git", "status", "--porcelain"]),
        "baseline_mode": args.baseline_mode,
        "require_equivalent": args.require_equivalent,
        "functional_only": args.functional_only,
        "definitions_source": args.definitions_source,
        "definitions_path": str(args.definitions.resolve()),
        "go": run.command(["go", "version"], env=env).strip(),
        "environment": {key: env[key] for key in ("GOPROXY", "GOSUMDB", "GOTOOLCHAIN", "GOFLAGS", "GOMAXPROCS", "TZ")},
        "samples": args.samples, "benchtime": args.benchtime,
        "scope": f"internal parser extraction only; {args.baseline_mode} baseline versus explicit whole YAML snapshot",
        "dns": "identical build-only overlay replaces resolver lookup with a fixed public IP",
        "timers": "definition loading, fixture reads and parser construction excluded",
        "started_at": datetime.now(timezone.utc).isoformat(),
    }
    run.write_json(output / "environment.json", metadata)
    baseline, candidate = output / "baseline", output / "candidate"
    run.archive_source(metadata["baseline_sha"], baseline, output / "baseline.tar")
    live = None
    if args.candidate_ref:
        run.archive_source(metadata["candidate_sha"], candidate, output / "candidate.tar")
    else:
        live = run.capture_worktree(candidate)
    guards = {"baseline": run.source_guard(baseline, output, "baseline"),
              "candidate": run.source_guard(candidate, output, "candidate", live)}
    definitions = output / "definitions"
    original_definitions = args.definitions / "definitions"
    definition_manifest = run.file_manifest(original_definitions)
    if not definition_manifest or any(path.is_symlink() for path in original_definitions.rglob("*")):
        raise RuntimeError("definitions must be a nonempty regular-file snapshot")
    shutil.copytree(original_definitions, definitions)
    if (definition_manifest != run.file_manifest(original_definitions) or
            definition_manifest != run.file_manifest(definitions)):
        raise RuntimeError("definitions changed during capture; no recapture attempted")
    run.write_json(output / "definition-digests.json", definition_manifest)
    metadata["definitions_manifest_sha256"] = run.observation_digest(definition_manifest)
    env["HERMES_COMPAT_DEFINITIONS"] = str(definitions)
    fixtures = output / "fixtures"
    fixtures.mkdir()
    for name, original in (("nytimes.html", "www.nytimes.com.html"), ("arstechnica.html", "arstechnica.com.html")):
        shutil.copyfile(baseline / "internal" / "fixtures" / original, fixtures / name)
    env["HERMES_COMPAT_FIXTURES"] = str(fixtures)
    run.write_json(output / "fixture-digests.json",
                   {p.name: {"sha256": run.digest(p), "bytes": p.stat().st_size} for p in sorted(fixtures.iterdir())})
    tooling = {name: run.digest(run.TOOLS / name)
               for name in ("run.py", "run_configured.py", "configured_benchmark_test.go")}
    run.write_json(output / "tooling-digests.json", tooling)
    relative = Path("internal/validation/url.go")
    original = (baseline / relative).read_text()
    if original != (candidate / relative).read_text():
        raise RuntimeError("paired DNS seam requires identical source")
    needle = "resolver := &net.Resolver{}\n\taddrs, err := resolver.LookupIPAddr(ctx, hostname)"
    if original.count(needle) != 1:
        raise RuntimeError("DNS seam changed")
    replacement = 'addrs, err := []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, error(nil)'
    controlled = output / "validation-offline.go"
    controlled.write_text(original.replace(needle, replacement))
    metadata["original_validation_sha256"] = run.digest(baseline / relative)
    metadata["controlled_validation_sha256"] = run.digest(controlled)
    run.record_provenance(output, metadata, guards, tooling)
    baseline_adapter = CANDIDATE_ADAPTER if args.baseline_mode == "definitions" else BASELINE_ADAPTER
    snapshots = (("baseline", baseline, baseline_adapter), ("candidate", candidate, CANDIDATE_ADAPTER))
    for label, source, adapter in snapshots:
        print(f"Preparing configured {label}", flush=True)
        evidence = output / (label + "-evidence")
        evidence.mkdir()
        package = source / "scripts" / "configured-compatibility"
        package.mkdir(parents=True)
        shutil.copyfile(run.TOOLS / "configured_benchmark_test.go", package / "configured_benchmark_test.go")
        (package / "adapter_test.go").write_text(adapter)
        run.command(["gofmt", "-w", str(package / "adapter_test.go")], env=env)
        shutil.copyfile(package / "adapter_test.go", output / (label + "-adapter.go"))
        metadata.setdefault("adapter_sha256", {})[label] = run.digest(output / (label + "-adapter.go"))
        overlay = output / (label + "-overlay.json")
        run.write_json(overlay, {"Replace": {str(source / relative): str(controlled)}})
        binary = output / (label + "-bench")
        run.command(["go", "test", "-c", "-tags=compatibility_configured", "-overlay", str(overlay),
                     "-o", str(binary), "./scripts/configured-compatibility"],
                    cwd=source, env=env, output=evidence / "build.log")
        for repeat in ("", "-repeat"):
            run.command([str(binary), "-test.run=^TestConfiguredEvidence$", "-test.v"], cwd=source, env=env,
                        output=evidence / ("observations" + repeat + ".log"))
            run.write_json(evidence / ("observations" + repeat + ".json"),
                           observations(evidence / ("observations" + repeat + ".log")))
    old = observations(output / "baseline-evidence" / "observations.log")
    new = observations(output / "candidate-evidence" / "observations.log")
    report = {"observations_per_version": len(old), "differences": compare_observations(old, new),
              "require_equivalent": args.require_equivalent,
              "repeat_differences": {
                  label: run.differences(observations(output / (label + "-evidence") / "observations.log"),
                                        observations(output / (label + "-evidence") / "observations-repeat.log"))
                  for label, _, _ in snapshots}}
    if args.reference:
        reference = args.reference.resolve()
        reference_metadata = json.loads((reference / "environment.json").read_text())
        if run.source_capture_failures(reference, reference_metadata):
            raise RuntimeError("reference capture is incomplete or recorded runtime drift")
        previous = json.loads((reference / "candidate-evidence" / "observations.json").read_text())
        report["reference"] = str(reference)
        report["reference_candidate_differences"] = compare_observations(previous, new)
    run.write_json(output / "comparison.json", report)
    run.write_json(output / "environment.json", metadata)
    run.write_json(output / "commands.json", run.COMMANDS)
    run.check_sources(guards, output, "functional-complete")
    if run.file_manifest(definitions) != definition_manifest:
        raise RuntimeError("captured definitions changed; no recapture attempted")
    if (any(report["repeat_differences"].values()) or report.get("reference_candidate_differences")
            or (args.require_equivalent and report["differences"])):
        raise RuntimeError("configured observations are nondeterministic or differ from the exact reference")
    if args.functional_only:
        print(json.dumps({"output": str(output), "changed_fields": len(report["differences"]),
                          "reference_differences": len(report.get("reference_candidate_differences", [])),
                          "baseline_mode": args.baseline_mode}, indent=2))
        return
    print("Six nonempty configured cases verified per version; running alternating benchmarks", flush=True)
    timing = []
    for sample in range(args.samples):
        for label, source, _ in snapshots if sample % 2 == 0 else reversed(snapshots):
            run.check_sources(guards, output, f"benchmark-{sample}-{label}")
            if run.file_manifest(definitions) != definition_manifest:
                raise RuntimeError("captured definitions changed; no recapture attempted")
            start = datetime.now(timezone.utc).isoformat()
            run.command([str(output / (label + "-bench")), "-test.run=^$", "-test.bench=BenchmarkConfigured",
                         "-test.benchmem", "-test.cpu=1", "-test.count=1", "-test.benchtime=" + args.benchtime],
                        cwd=source, env=env, output=output / (label + "-evidence") / f"benchmark-{sample}.log")
            timing.append({"version": label, "sample": sample, "start": start,
                           "end": datetime.now(timezone.utc).isoformat()})
    run.write_json(output / "benchmark-times.json", timing)
    report["benchmarks"] = {}
    for label, _, _ in snapshots:
        evidence = output / (label + "-evidence")
        combined = evidence / "benchmarks.log"
        combined.write_text("".join((evidence / f"benchmark-{i}.log").read_text() for i in range(args.samples)))
        report["benchmarks"][label] = run.bench_summary(combined)
    validate_samples(report["benchmarks"], args.samples)
    report["percent_changes"] = {
        name: {metric: (values[metric]["median"] / report["benchmarks"]["baseline"][name][metric]["median"] - 1) * 100
               for metric in ("ns/op", "B/op", "allocs/op")}
        for name, values in report["benchmarks"]["candidate"].items()
    }
    run.write_json(output / "comparison.json", report)
    run.check_sources(guards, output, "benchmark-complete")
    if run.file_manifest(definitions) != definition_manifest:
        raise RuntimeError("captured definitions changed; no recapture attempted")
    run.write_json(output / "commands.json", run.COMMANDS)
    print(json.dumps({"output": str(output), "changed_fields": len(report["differences"]),
                      "repeat_differences": {key: len(value) for key, value in report["repeat_differences"].items()},
                      "percent_changes": report["percent_changes"]}, indent=2))


if __name__ == "__main__":
    main()
