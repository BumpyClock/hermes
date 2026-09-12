#!/usr/bin/env python3
"""Compare public API text and deterministic parser observations in two worktrees."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile


SOURCE = Path(__file__).resolve().parent / "contract-snapshot" / "main.go"


def run(args, cwd, output):
    with output.open("wb") as stream, output.with_suffix(output.suffix + ".stderr").open("wb") as errors:
        subprocess.run(args, cwd=cwd, stdout=stream, stderr=errors, check=True,
                       env={**os.environ, "TZ": "UTC", "HERMES_PARSER_DEBUG": "0"})


def capture(repo, fixtures, destination, map_order="sorted", pattern="*.html"):
    destination.mkdir(parents=True, exist_ok=True)
    run(["go", "doc", "-all", "."], repo, destination / "api.txt")
    run(["git", "rev-parse", "HEAD"], repo, destination / "revision.txt")
    run(["git", "diff", "HEAD", "--stat"], repo, destination / "diff.txt")
    with tempfile.TemporaryDirectory(prefix="hermes-contract-") as temporary:
        temp = Path(temporary)
        binary = temp / "snapshot"
        run(["go", "build", "-o", str(binary), str(SOURCE)], repo, destination / "build.txt")
        run([str(binary), "-contracts"], repo, destination / "contracts.jsonl")

        # The fixture pass substitutes DNS lookup for offline execution.
        # The separate contract pass above uses the unmodified build.
        validation = repo / "internal" / "validation" / "url.go"
        original = validation.read_text()
        needle = "resolver := &net.Resolver{}\n\taddrs, err := resolver.LookupIPAddr(ctx, hostname)"
        replacement = 'addrs, err := []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, error(nil)'
        if original.count(needle) != 1:
            raise RuntimeError("DNS comparison seam changed; inspect validation before comparison")
        substitute = temp / "url.go"
        substitute.write_text(original.replace(needle, replacement))
        overlay = temp / "overlay.json"
        replacements = {str(validation): str(substitute)}
        if map_order != "raw":
            for relative, loop, indent in (
                ("internal/utils/dom/convert.go", "for key, value := range attrs {", "\t"),
                ("internal/resource/dom.go", "for attrName, value := range attrs {", "\t\t"),
                ("internal/utils/dom/clean.go", "for attrName := range attrs {", "\t\t"),
            ):
                source = repo / relative
                code = source.read_text()
                if code.count(loop) != 1:
                    raise RuntimeError(f"Map-order comparison seam changed: {relative}")
                key = "key" if "for key," in loop else "attrName"
                ordered = ("keys := make([]string, 0, len(attrs))\n" + indent
                           + "for key := range attrs { keys = append(keys, key) }\n" + indent
                           + "sort.Strings(keys)\n" + indent
                           + f"for _, {key} := range keys {{\n" + indent + "\t"
                           + (f"value := attrs[{key}]" if ", value" in loop else ""))
                if map_order == "reverse":
                    ordered = ordered.replace("sort.Strings(keys)", "sort.Sort(sort.Reverse(sort.StringSlice(keys)))")
                code = code.replace('import (', 'import (\n\t"sort"', 1).replace(loop, ordered)
                patched = temp / source.name
                patched.write_text(code)
                replacements[str(source)] = str(patched)
        overlay.write_text(json.dumps({"Replace": replacements}))
        run(["go", "build", "-overlay", str(overlay), "-o", str(binary), str(SOURCE)],
            repo, destination / "fixture-build.txt")
        run([str(binary), "-fixtures", str(fixtures), "-match", pattern], repo, destination / "fixtures.jsonl")


def compare(base, candidate):
    differences = []
    counts = {}
    if (base / "api.txt").read_bytes() != (candidate / "api.txt").read_bytes():
        differences.append("public API declarations or documentation differ")
    for filename in ("contracts.jsonl", "fixtures.jsonl"):
        old = [json.loads(row) for row in (base / filename).read_text().splitlines()]
        new = [json.loads(row) for row in (candidate / filename).read_text().splitlines()]
        counts[filename] = len(old)
        if [row["id"] for row in old] != [row["id"] for row in new]:
            differences.append(f"{filename}: case IDs or order differ")
        for before, after in zip(old, new):
            if before != after:
                fields = sorted(key for key in before.keys() | after.keys()
                                if before.get(key) != after.get(key))
                differences.append(f'{before["id"]}: {", ".join(fields)}')
    return {"cases": counts, "differences": differences}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base", type=Path, required=True)
    parser.add_argument("--candidate", type=Path, default=Path.cwd())
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--map-order", choices=("sorted", "reverse", "raw"), default="sorted")
    parser.add_argument("--pattern", default="*.html", help="fixture filename glob")
    args = parser.parse_args()
    base, candidate, output = args.base.resolve(), args.candidate.resolve(), args.output.resolve()
    if output.exists():
        parser.error("output directory must not exist; preserve earlier comparison evidence")
    output.mkdir(parents=True)
    if args.map_order != "raw":
        for relative in ("internal/utils/dom/convert.go", "internal/resource/dom.go", "internal/utils/dom/clean.go"):
            if (base / relative).read_bytes() != (candidate / relative).read_bytes():
                parser.error(f"Map-order control requires unchanged source: {relative}")
    fixtures = base / "internal" / "fixtures"
    manifest = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(fixtures.glob(args.pattern))}
    (output / "fixture-manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    for name, repo in (("base", base), ("candidate", candidate)):
        print(f"Capture {name}: {repo}", flush=True)
        capture(repo, fixtures, output / name, args.map_order, args.pattern)
    result = compare(output / "base", output / "candidate")
    result["map_order"] = args.map_order
    result["dns"] = "fixed public IP for fixture pass only"
    (output / "comparison.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps(result, indent=2))
    raise SystemExit(bool(result["differences"]))


if __name__ == "__main__":
    main()
