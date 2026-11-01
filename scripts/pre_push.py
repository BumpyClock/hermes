"""Read-only validation of the checked-out commit named by pre-push stdin."""
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys


def git(*args, check=True):
    return subprocess.run(["git", *args], check=check, stdout=subprocess.PIPE,
                          stderr=None if check else subprocess.DEVNULL)


def pushed_paths():
    head = git("rev-parse", "HEAD").stdout.strip().decode()
    paths = set()
    full = False
    updates = False
    records = sys.stdin.read().splitlines()
    if not records:
        full = True
        updates = True
    for record in records:
        fields = record.split()
        if len(fields) != 4:
            raise ValueError("invalid pre-push input; expected four fields per ref")
        _, local, _, remote = fields
        if not all(re.fullmatch(r"[0-9a-fA-F]{40}|[0-9a-fA-F]{64}", oid) for oid in (local, remote)):
            raise ValueError("invalid pre-push object ID")
        if set(local) == {"0"}:
            continue
        updates = True
        commit = git("rev-parse", "--verify", local + "^{commit}").stdout.strip().decode()
        if commit != head:
            raise ValueError("pushed ref is not HEAD; validate it in its own clean checkout")
        if set(remote) == {"0"}:
            full = True
            continue
        base = git("rev-parse", "--verify", remote + "^{commit}", check=False)
        if base.returncode:
            full = True
            continue
        diff = git("diff", "--name-only", "--no-renames", "-z", base.stdout.strip().decode(), head)
        paths.update(os.fsdecode(path) for path in diff.stdout.split(b"\0") if path)
    if not updates:
        return None
    status = git("status", "--porcelain=v1", "--untracked-files=normal").stdout
    if status:
        raise ValueError("checkout must be clean, including untracked files; never stash automatically")
    if full or not paths:
        return tracked_paths()
    return paths


def tracked_paths():
    return {os.fsdecode(path) for path in git("ls-files", "-z").stdout.split(b"\0") if path}


def run(*args, capture=False):
    print("pre-push:", " ".join(args), flush=True)
    result = subprocess.run(args, check=True, stdout=subprocess.PIPE if capture else None)
    return result.stdout


def main():
    root = git("rev-parse", "--show-toplevel").stdout.strip()
    os.chdir(os.fsdecode(root))
    paths = pushed_paths()
    if paths is None:
        print("pre-push: deletion-only update; no code to validate")
        return
    if sys.version_info < (3, 11):
        raise ValueError("prepare Python 3.11+ before pushing")
    for tool in ("go", "golangci-lint", "make"):
        if not shutil.which(tool):
            raise ValueError(f"missing {tool}; follow docs/development/contributing.md setup first")
    os.environ.update(GOTOOLCHAIN="local", GOPROXY="off", GOSUMDB="off", GOFLAGS="-mod=readonly", PYTHONDONTWRITEBYTECODE="1")
    version = re.search(r"^go (\S+)$", Path("go.mod").read_text(), re.MULTILINE).group(1)
    if run("go", "env", "GOVERSION", capture=True).strip().decode() != "go" + version:
        raise ValueError(f"prepare Go {version} on PATH; hook never downloads toolchains")
    lint_version = run("golangci-lint", "version", capture=True).decode()
    if not re.search(r"\bversion 2\.13\.2\b", lint_version):
        raise ValueError("prepare golangci-lint v2.13.2 built with the Go version in go.mod")
    shared = {"go.mod", "go.sum", ".golangci.yml", "Makefile"}
    if any(path in shared or path.startswith(("scripts/", ".githooks/")) or (not path.endswith((".go", ".md")) and not path.startswith("docs/")) for path in paths):
        paths = tracked_paths()
    files = sorted("./" + path for path in paths if path.endswith(".go") and Path(path).is_file())
    if files and run("golangci-lint", "fmt", "--diff", *files, capture=True):
        raise ValueError("formatting differs; run the documented formatter explicitly, then commit")
    # Full graph avoids an unproven reverse-dependency detector. CI retains race coverage.
    run("golangci-lint", "run")
    run("go", "vet", "./...")
    run("go", "build", "./...")
    run("make", "test-release", "test-compatibility")


if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as error:
        sys.exit(error.returncode if 0 < error.returncode < 256 else 1)
    except (ValueError, OSError) as error:
        print(f"pre-push: {error}", file=sys.stderr)
        sys.exit(1)
