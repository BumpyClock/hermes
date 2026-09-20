#!/usr/bin/env python3
"""Offline released-versus-source-snapshot compatibility evidence; never changes Git state."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import platform
import re
import shutil
import statistics
import subprocess
import tarfile


ROOT = Path(__file__).resolve().parents[2]
TOOLS = ROOT / "scripts" / "compatibility"
COMMANDS = []


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def archive_source(revision, destination, archive):
    COMMANDS.append({"argv": ["git", "archive", revision], "cwd": str(ROOT)})
    with archive.open("wb") as stream:
        subprocess.run(["git", "archive", revision], cwd=ROOT, stdout=stream, check=True)
    destination.mkdir()
    with tarfile.open(archive) as source:
        source.extractall(destination, filter="data")


def write_json(path, data):
    path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n")


def command(args, cwd=ROOT, env=None, output=None):
    COMMANDS.append({"argv": args, "cwd": str(cwd)})
    result = subprocess.run(args, cwd=cwd, env=env, capture_output=True)
    if output:
        output.write_bytes(result.stdout)
        output.with_suffix(output.suffix + ".stderr").write_bytes(result.stderr)
    if result.returncode:
        raise RuntimeError(f"{args!r} failed ({result.returncode}): {result.stderr.decode()}")
    return result.stdout.decode()


def differences(before, after, path=""):
    if isinstance(before, dict) and isinstance(after, dict):
        result = []
        for key in sorted(before.keys() | after.keys()):
            child = f"{path}/{key}"
            if key not in before or key not in after:
                result.append({"path": child, "before": before.get(key), "after": after.get(key)})
            else:
                result.extend(differences(before[key], after[key], child))
        return result
    return [] if before == after else [{"path": path, "before": before, "after": after}]


def rows(path):
    result = {}
    for line in path.read_text().splitlines():
        row = json.loads(line)
        if row["id"] in result:
            raise ValueError(f"duplicate observation: {row['id']}")
        result[row["id"]] = row
    return result


def observation_digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(",", ":")).encode()).hexdigest()


def classify_observations(before, after, allowed):
    accepted, unexpected = [], []
    for key in sorted(before.keys() | after.keys()):
        if key in before and key in after and before[key] == after[key]:
            continue
        pair = [observation_digest(before.get(key)), observation_digest(after.get(key))]
        (accepted if allowed.get(key) == pair else unexpected).append(key)
    return {"accepted": accepted, "unexpected": unexpected,
            "stale_allowances": sorted(allowed.keys() - set(accepted))}


def acceptance(output, report, baseline_sha):
    allowed = json.loads((TOOLS / "intentional-changes.json").read_text())
    result = classify_observations(
        rows(output / "baseline-evidence" / "fixtures.jsonl"),
        rows(output / "candidate-evidence" / "fixtures.jsonl"), allowed["observations"])
    reasons = allowed["reasons"]
    result["reasons"] = {
        key: reasons["text" if "/text/" in key else
                     "literal-markdown" if key.startswith("literal.") else "arstechnica-markdown"]
        for key in result["accepted"]
    }
    failures = []
    if baseline_sha != allowed["baseline_sha"]:
        failures.append("baseline SHA differs from reviewed release")
    if report["api"]["removed"] or report["api"]["changed"]:
        failures.append("stable exported API changed")
    if report["api"]["added"] != sorted(allowed["added_api"]):
        failures.append("exported API additions differ from reviewed list")
    for kind in ("contracts", "fixtures"):
        if any(report[kind]["repeat_differences"].values()):
            failures.append(f"{kind} observations are nondeterministic")
    if report["contracts"]["differences"]:
        failures.append("deterministic public contracts changed")
    if result["unexpected"] or result["stale_allowances"]:
        failures.append("fixture changes differ from exact reviewed observations")
    result["failures"] = failures
    result["passed"] = not failures
    return result


def bench_summary(output):
    values = {}
    for line in output.read_text().splitlines():
        match = re.match(r"(Benchmark\S+)\s+\d+\s+([\d.]+) ns/op.*?\s+(\d+) B/op\s+(\d+) allocs/op", line)
        if match:
            name, ns, size, allocs = match.groups()
            values.setdefault(name, []).append([float(ns), int(size), int(allocs)])
    return {
        name: {
            "samples": len(samples),
            **{metric: {"median": statistics.median(column), "min": min(column), "max": max(column)}
               for metric, column in zip(("ns/op", "B/op", "allocs/op"), zip(*samples))},
        } for name, samples in values.items()
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", default="v1.1.1")
    parser.add_argument("--candidate-ref", help="archive this immutable commit instead of the dirty worktree")
    parser.add_argument("--output", required=True, help="new directory under .compatibility-runs/")
    parser.add_argument("--samples", type=int, default=6)
    parser.add_argument("--benchtime", default="1s")
    parser.add_argument("--verify-only", action="store_true",
                        help="check existing evidence against exact reviewed changes without recapturing")
    parser.add_argument("--functional-only", action="store_true",
                        help="capture API, contracts and fixtures without compiling or running benchmarks")
    args = parser.parse_args()
    if args.samples < 5:
        parser.error("at least five benchmark samples are required")
    output = (ROOT / args.output).resolve()
    if not output.is_relative_to(ROOT / ".compatibility-runs"):
        parser.error("output must be under .compatibility-runs/")
    if args.verify_only:
        report = json.loads((output / "comparison.json").read_text())
        metadata = json.loads((output / "environment.json").read_text())
        result = acceptance(output, report, metadata["baseline_sha"])
        write_json(output / "acceptance.json", result)
        print(json.dumps(result, indent=2))
        raise SystemExit(not result["passed"])
    if output.exists():
        parser.error("output must be a new directory under .compatibility-runs/")
    source_paths = command(["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"]).split("\0")
    source_paths = sorted({p for p in source_paths if p and not p.startswith(".compatibility-runs/")
                           and (ROOT / p).is_file()})
    output.mkdir(parents=True)
    build_dir = output / "build-work"
    build_dir.mkdir()
    env = {**os.environ, "GOPROXY": "off", "GOSUMDB": "off", "GOTOOLCHAIN": "local",
           "GOFLAGS": "-mod=readonly", "GOMAXPROCS": "1", "TZ": "UTC",
           "HERMES_PARSER_DEBUG": "0", "GOTMPDIR": str(build_dir), "TMPDIR": str(build_dir)}
    metadata = {
        "baseline_sha": command(["git", "rev-parse", args.baseline + "^{commit}"]).strip(),
        "candidate_head": command(["git", "rev-parse", (args.candidate_ref or "HEAD") + "^{commit}"]).strip(),
        "candidate_status": "immutable archive" if args.candidate_ref else command(["git", "status", "--porcelain"]),
        "candidate_source": "git archive" if args.candidate_ref else "tracked and unignored worktree files",
        "go": command(["go", "version"], env=env).strip(),
        "platform": platform.platform(),
        "cpu": command(["sysctl", "-n", "machdep.cpu.brand_string"]).strip() if platform.system() == "Darwin" else platform.processor(),
        "environment": {key: env[key] for key in ("GOPROXY", "GOSUMDB", "GOTOOLCHAIN", "GOFLAGS", "GOMAXPROCS", "TZ")},
        "samples": args.samples, "benchtime": args.benchtime,
        "functional_only": args.functional_only,
        "benchmark_order": "alternating baseline/candidate; reversed on odd samples",
    }
    write_json(output / "environment.json", metadata)
    baseline, candidate = output / "baseline", output / "candidate"
    archive_source(metadata["baseline_sha"], baseline, output / "baseline.tar")
    if args.candidate_ref:
        archive_source(metadata["candidate_head"], candidate, output / "candidate.tar")
    else:
        candidate.mkdir()
        for relative in source_paths:
            destination = candidate / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(ROOT / relative, destination)
    write_json(output / "candidate-files.json",
               {str(path.relative_to(candidate)): digest(path)
                for path in sorted(candidate.rglob("*")) if path.is_file()})
    write_json(output / "tooling-files.json",
               {str(path.relative_to(ROOT)): digest(path)
                for path in sorted(TOOLS.rglob("*")) if path.is_file()}
               | {"scripts/contract-snapshot/main.go": digest(ROOT / "scripts" / "contract-snapshot" / "main.go")})
    fixtures = output / "fixtures"
    shutil.copytree(TOOLS / "fixtures", fixtures)
    for name, original in (("nytimes.html", "www.nytimes.com.html"), ("arstechnica.html", "arstechnica.com.html")):
        shutil.copyfile(baseline / "internal" / "fixtures" / original, fixtures / name)
    write_json(output / "fixture-digests.json", {p.name: {"sha256": digest(p), "bytes": p.stat().st_size}
                                               for p in sorted(fixtures.glob("*.html"))})
    env["HERMES_COMPAT_FIXTURES"] = str(fixtures)
    # Keep the existing deterministic contracts unchanged; restrict fixture URLs
    # to literal public IPs instead of replacing production DNS or DOM code.
    probe = (ROOT / "scripts" / "contract-snapshot" / "main.go").read_text()
    needle = 'for _, route := range []string{host, "93.184.216.34"}'
    if probe.count(needle) != 1:
        raise RuntimeError("contract runner fixture loop changed")
    probe = probe.replace(needle, 'for _, route := range []string{"93.184.216.34"}')
    probe = probe.replace('host := strings.SplitN', '_ = strings.SplitN')
    (output / "snapshot.go").write_text(probe)
    api_binary = output / "api-shapes"
    command(["go", "build", "-o", str(api_binary), str(TOOLS / "api" / "main.go")], env=env,
            output=output / "api-build.log")
    snapshots = (("baseline", baseline), ("candidate", candidate))
    for label, source in snapshots:
        print(f"Capturing {label}", flush=True)
        evidence = output / (label + "-evidence")
        evidence.mkdir()
        command([str(api_binary)], cwd=source, env=env, output=evidence / "api.json")
        binary = output / (label + "-contracts")
        command(["go", "build", "-o", str(binary), str(output / "snapshot.go")],
                cwd=source, env=env, output=evidence / "build.log")
        for repeat in ("", "-repeat"):
            command([str(binary), "-contracts"], cwd=source, env=env,
                    output=evidence / f"contracts{repeat}.jsonl")
            command([str(binary), "-fixtures", str(fixtures)], cwd=source, env=env,
                    output=evidence / f"fixtures{repeat}.jsonl")
        # Compile released consumers unchanged against both implementations.
        if label == "candidate":
            shutil.rmtree(source / "examples")
            shutil.copytree(baseline / "examples", source / "examples")
        command(["go", "test", "-run", "^$", "./examples/..."], cwd=source, env=env,
                output=evidence / "consumer-compile.log")
        if not args.functional_only:
            bench = source / "scripts" / "compatibility-bench"
            bench.mkdir(parents=True)
            shutil.copyfile(TOOLS / "benchmark_test.go", bench / "benchmark_test.go")
            command(["go", "test", "-c", "-o", str(output / (label + "-bench")), "./scripts/compatibility-bench"],
                    cwd=source, env=env, output=evidence / "benchmark-build.log")
    report = {"contracts": {}, "fixtures": {}, "api": {}}
    for kind in ("contracts", "fixtures"):
        old = rows(output / "baseline-evidence" / f"{kind}.jsonl")
        new = rows(output / "candidate-evidence" / f"{kind}.jsonl")
        report[kind] = {"baseline_count": len(old), "candidate_count": len(new),
                        "differences": differences(old, new), "repeat_differences": {}}
        for label, _ in snapshots:
            evidence = output / (label + "-evidence")
            report[kind]["repeat_differences"][label] = differences(
                rows(evidence / f"{kind}.jsonl"), rows(evidence / f"{kind}-repeat.jsonl"))
    old = json.loads((output / "baseline-evidence" / "api.json").read_text())
    new = json.loads((output / "candidate-evidence" / "api.json").read_text())
    report["api"] = {"baseline_count": len(old), "candidate_count": len(new),
                     "removed": sorted(old.keys() - new.keys()), "added": sorted(new.keys() - old.keys()),
                     "changed": differences(old, {key: new[key] for key in old if key in new})}
    report["acceptance"] = acceptance(output, report, metadata["baseline_sha"])
    write_json(output / "comparison.json", report)
    write_json(output / "commands.json", COMMANDS)
    if args.functional_only:
        print(json.dumps({"output": str(output), "api": report["api"],
                          "contract_differences": len(report["contracts"]["differences"]),
                          "fixture_differences": len(report["fixtures"]["differences"]),
                          "acceptance": report["acceptance"]}, indent=2))
        raise SystemExit(not report["acceptance"]["passed"])
    print("Functional/API comparison recorded; running alternating benchmarks", flush=True)
    for sample in range(args.samples):
        for label, source in snapshots if sample % 2 == 0 else reversed(snapshots):
            evidence = output / (label + "-evidence")
            command([str(output / (label + "-bench")), "-test.run=^$", "-test.bench=BenchmarkGeneric",
                     "-test.benchmem", "-test.cpu=1", "-test.count=1", "-test.benchtime=" + args.benchtime],
                    cwd=source, env=env, output=evidence / f"benchmark-{sample}.log")
    for label, _ in snapshots:
        evidence = output / (label + "-evidence")
        combined = evidence / "benchmarks.log"
        combined.write_text("".join((evidence / f"benchmark-{i}.log").read_text() for i in range(args.samples)))
        report.setdefault("benchmarks", {})[label] = bench_summary(combined)
    report["benchmark_percent_changes"] = {
        name: {metric: (values[metric]["median"] / report["benchmarks"]["baseline"][name][metric]["median"] - 1) * 100
               for metric in ("ns/op", "B/op", "allocs/op")}
        for name, values in report["benchmarks"]["candidate"].items()
    }
    write_json(output / "comparison.json", report)
    write_json(output / "commands.json", COMMANDS)
    print(json.dumps({"output": str(output), "api": report["api"],
                      "contract_differences": len(report["contracts"]["differences"]),
                      "fixture_differences": len(report["fixtures"]["differences"]),
                      "benchmark_percent_changes": report["benchmark_percent_changes"]}, indent=2))
    raise SystemExit(not report["acceptance"]["passed"])


if __name__ == "__main__":
    main()
