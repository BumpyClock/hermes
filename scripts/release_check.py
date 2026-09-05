"""Read-only gates for source-only Hermes v1 releases."""

import argparse
import datetime
import json
import re
import subprocess
import sys
from pathlib import Path

REPO = "BumpyClock/hermes"
MODULE = "github.com/" + REPO
VERSION = r"(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)"


class CheckError(Exception):
    pass


def run(*args):
    try:
        result = subprocess.run(args, capture_output=True, text=True, check=False, timeout=30)
    except subprocess.TimeoutExpired:
        raise CheckError(f"{args[0]} {args[1]} timed out after 30 seconds") from None
    if result.returncode:
        # External diagnostics can contain credentials from remote configuration.
        raise CheckError(f"{args[0]} {args[1]} failed (exit {result.returncode})")
    return result.stdout.strip()


def require(condition, message):
    if not condition:
        raise CheckError(message)


def parse_version(value):
    require(re.fullmatch(VERSION, value), "Use a stable numeric version: X.Y.Z")
    return tuple(map(int, value.split(".")))


def github_identity():
    require(run("gh", "api", "user", "--jq", ".login").lower() == "bumpyclock",
            "Authenticate as BumpyClock before release checks")
    repo = json.loads(run("gh", "api", f"repos/{REPO}"))
    require(repo["full_name"].lower() == REPO.lower() and repo["default_branch"] == "master",
            "GitHub repository or default branch mismatch")


def source_version(root):
    source = (root / "cmd/hermes/main.go").read_text()
    versions = re.findall(r'fmt\.Println\("Hermes v(' + VERSION + r')"\)', source)
    require(len(versions) == 1, "Expected one CLI release version in cmd/hermes/main.go")
    return versions[0][0]


def check_artifacts(root, expected):
    changelog = (root / "CHANGELOG.md").read_text()
    headings = re.findall(r"^## \[([^\]]+)\](.*)$", changelog, re.MULTILINE)
    releases = [(v, suffix) for v, suffix in headings if re.fullmatch(VERSION, v)]
    require(releases and releases[0][0] == expected, "Latest changelog version does not match CLI version")
    date = re.fullmatch(r" - (\d{4}-\d{2}-\d{2})", releases[0][1])
    require(date is not None, "Latest changelog release needs a YYYY-MM-DD date")
    try:
        datetime.date.fromisoformat(date[1])
    except ValueError as exc:
        raise CheckError("Invalid changelog release date") from exc
    paths = run("git", "ls-files", "-z").split("\0")
    for path in filter(None, paths):
        absolute = str(root / path)
        kind = run("file", "-b", "--", absolute)
        if not any(marker in kind for marker in ("PE32", "Mach-O", "ELF")):
            continue
        versions = set(re.findall(r"Hermes v(" + VERSION + r")", run("strings", "--", absolute)))
        require({v[0] for v in versions} == {expected},
                f"Tracked binary {path} has absent or inconsistent Hermes version metadata")
        print(f"Binary: {path}: Hermes v{expected}")


def preflight(version, ready=False):
    target = parse_version(version)
    require(target[0] == 1, "This procedure supports v1 only; major releases need a module migration")
    root = Path(run("git", "rev-parse", "--show-toplevel"))
    require(Path.cwd().resolve() == root.resolve(), "Run from the repository root")
    require(run("git", "branch", "--show-current") == "master", "Release branch must be master")
    require(not run("git", "status", "--porcelain", "--untracked-files=all"), "Release requires a clean worktree and index")
    require(re.search(r"^module " + re.escape(MODULE) + r"\s*$", (root / "go.mod").read_text(), re.MULTILINE),
            "Unexpected Go module path")
    allowed = {f"https://github.com/{REPO}.git", f"https://github.com/{REPO}",
               f"git@github.com:{REPO}.git", f"ssh://git@github.com/{REPO}.git"}
    for args in (("git", "remote", "get-url", "--all", "origin"),
                 ("git", "remote", "get-url", "--push", "--all", "origin")):
        require(run(*args) in allowed, "origin must have one intended Hermes URL")
    github_identity()
    head = run("git", "rev-parse", "HEAD")
    remote = run("git", "ls-remote", "--heads", "origin", "refs/heads/master")
    require(remote.split() == [head, "refs/heads/master"], "HEAD must equal remote master")
    refs = run("git", "ls-remote", "--tags", "--refs", "origin").splitlines()
    tags = {line.split()[1].removeprefix("refs/tags/") for line in refs}
    tag = "v" + version
    require(tag not in tags, "Target tag already exists remotely; use the recovery procedure")
    require(not run("git", "tag", "--list", tag), "Target tag already exists locally; use the recovery procedure")
    # A successful paginated list distinguishes absence from authentication or network failure.
    releases = json.loads(run("gh", "api", "--paginate", "--slurp", f"repos/{REPO}/releases?per_page=100"))
    require(not any(release["tag_name"] == tag for page in releases for release in page),
            "GitHub release already exists; use the recovery procedure")
    stable = [parse_version(t[1:]) for t in tags if re.fullmatch("v" + VERSION, t)]
    require(stable and target > max(stable), "Target version must exceed the latest remote stable tag")
    current = source_version(root)
    require(current == version if ready else parse_version(current) < target,
            "CLI version must match target" if ready else "Target must exceed the current CLI version")
    check_artifacts(root, current)
    print(f"{'Ready' if ready else 'Preflight'} passed: {tag}, commit {head}. No publication performed.")


def check_ci(commit):
    require(re.fullmatch(r"[0-9a-f]{40}", commit), "Use the full 40-character release commit SHA")
    github_identity()
    runs = json.loads(run("gh", "run", "list", "--repo", REPO, "--workflow", "ci.yml",
                          "--branch", "master", "--event", "push", "--commit", commit,
                          "--limit", "100", "--json", "databaseId,headSha,status,conclusion,url"))
    require(runs and all(item["headSha"] == commit for item in runs), "No matching exact-commit CI evidence")
    latest = max(runs, key=lambda item: item["databaseId"])
    require(latest["status"] == "completed" and latest["conclusion"] == "success",
            "Latest exact-commit CI run is not successful")
    print(f"CI passed: {commit}: {latest['url']}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=("preflight", "ready", "ci"))
    parser.add_argument("value", help="numeric version for preflight/ready, full commit SHA for ci")
    args = parser.parse_args()
    try:
        if args.mode == "ci":
            check_ci(args.value)
        else:
            preflight(args.value, ready=args.mode == "ready")
    except (CheckError, OSError, ValueError, KeyError, TypeError) as exc:
        print(f"Release check failed: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
