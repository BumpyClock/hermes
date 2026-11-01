"""Exercise executable hooks in disposable repositories without outbound pushes."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

SOURCE = Path(__file__).resolve().parents[1]
ZERO = "0" * 40


class HookTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(prefix="hook fixture ")
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.repo = self.root / "checkout with spaces"
        self.repo.mkdir()
        self.tools = self.root / "tools"
        self.tools.mkdir()
        self.log = self.root / "calls.jsonl"
        for tool in ("git", "python3"):
            (self.tools / tool).symlink_to(shutil.which(tool))
        self.env = dict(os.environ, PATH=str(self.tools),
                        GIT_CONFIG_GLOBAL=os.devnull, GIT_CONFIG_NOSYSTEM="1",
                        HOOK_TEST_LOG=str(self.log))
        self.git("init", "-q")
        self.git("config", "user.name", "Hook fixture")
        self.git("config", "user.email", "fixture@example.invalid")
        self.git("config", "commit.gpgsign", "false")
        for path in (".gitattributes", ".githooks/pre-push", "scripts/pre_push.py", "scripts/install-hooks.sh"):
            target = self.repo / path
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(SOURCE / path, target)
        (self.repo / "go.mod").write_text("module fixture\n\ngo 1.27.1\n")
        (self.repo / "main.go").write_text("package main\n")
        (self.repo / "unchanged.go").write_text("package main\n")
        self.base = self.commit()
        for tool in ("go", "golangci-lint", "make"):
            path = self.tools / tool
            path.write_text("""#!/usr/bin/env python3
import json, os, pathlib, sys
name = pathlib.Path(sys.argv[0]).name
args = sys.argv[1:]
with open(os.environ['HOOK_TEST_LOG'], 'a') as f:
    f.write(json.dumps([name, *args]) + '\\n')
if name == 'go' and args == ['env', 'GOVERSION']:
    print(os.environ.get('HOOK_GO_VERSION', 'go1.27.1'))
elif name == 'golangci-lint' and args == ['version']:
    print('golangci-lint has version 2.13.2 built with go1.27.1')
elif name == 'golangci-lint' and args[:2] == ['fmt', '--diff'] and os.environ.get('HOOK_FORMAT_DIFF'):
    print('formatting diff')
if os.environ.get('HOOK_FAIL') == ' '.join([name, *args]):
    sys.exit(37)
""")
            path.chmod(0o755)

    def git(self, *args, cwd=None):
        return subprocess.check_output(["git", *args], cwd=cwd or self.repo, env=self.env, stderr=subprocess.DEVNULL).decode().strip()

    def commit(self):
        self.git("add", ".")
        self.git("-c", "core.hooksPath=/dev/null", "commit", "-qm", "fixture")
        return self.git("rev-parse", "HEAD")

    def update(self, name="file with spaces.go"):
        (self.repo / name).write_text("package main\n")
        return self.commit()

    def record(self, local=None, remote=None):
        return f"refs/heads/local {local or self.git('rev-parse', 'HEAD')} refs/heads/remote {remote or self.base}\n"

    def hook(self, stdin, **extra):
        return subprocess.run([str(self.repo / ".githooks/pre-push"), "origin", "unused"],
                              input=stdin, text=True, cwd=self.repo,
                              env=dict(self.env, **extra), capture_output=True)

    def calls(self):
        return [json.loads(line) for line in self.log.read_text().splitlines()] if self.log.exists() else []

    def install(self, cwd=None):
        return subprocess.run([str(self.repo / "scripts/install-hooks.sh")], cwd=cwd or self.repo,
                              env=self.env, capture_output=True, text=True)

    def test_changed_file_with_spaces_and_full_graph(self):
        self.update()
        result = self.hook(self.record())
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = self.calls()
        self.assertIn(["golangci-lint", "fmt", "--diff", "./file with spaces.go"], calls)
        for command in (["golangci-lint", "run"], ["go", "vet", "./..."],
                        ["go", "build", "./..."], ["make", "test-release", "test-compatibility"]):
            self.assertIn(command, calls)
        self.assertEqual(self.git("status", "--porcelain"), "")

    def test_new_branch_unknown_base_and_empty_input_broaden(self):
        self.update()
        for stdin in (self.record(remote=ZERO), self.record(remote="f" * 40), "", self.record(remote=self.git("rev-parse", "HEAD"))):
            with self.subTest(stdin=stdin):
                self.log.unlink(missing_ok=True)
                result = self.hook(stdin)
                self.assertEqual(result.returncode, 0, result.stderr)
                formatting = next(call for call in self.calls() if call[:3] == ["golangci-lint", "fmt", "--diff"])
                self.assertIn("./unchanged.go", formatting)

    def test_shared_config_rename_and_deletion(self):
        self.git("mv", "main.go", "renamed file.go")
        (self.repo / "unchanged.go").unlink()
        (self.repo / ".golangci.yml").write_text('version: "2"\n')
        self.commit()
        result = self.hook(self.record())
        self.assertEqual(result.returncode, 0, result.stderr)
        formatting = next(call for call in self.calls() if call[:3] == ["golangci-lint", "fmt", "--diff"])
        self.assertIn("./renamed file.go", formatting)
        self.assertNotIn("./main.go", formatting)
        self.assertNotIn("./unchanged.go", formatting)

    def test_multiple_refs_and_deletion_only(self):
        self.update()
        result = self.hook(self.record() + self.record(remote=ZERO))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.log.unlink()
        (self.repo / "dirty").write_text("untracked")
        result = self.hook(self.record(local=ZERO))
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.calls(), [])

    def test_unrelated_ref_dirty_and_malformed_fail_before_tools(self):
        self.update()
        for stdin in (self.record(local=self.base), self.record() + self.record(local=self.base), "bad record\n", self.record(local="x" * 40)):
            self.assertNotEqual(self.hook(stdin).returncode, 0)
            self.assertEqual(self.calls(), [])
        (self.repo / "untracked").write_text("dirty")
        self.assertNotEqual(self.hook(self.record()).returncode, 0)
        self.assertEqual(self.calls(), [])

    def test_format_diff_and_missing_tool_fail(self):
        self.update()
        self.assertNotEqual(self.hook(self.record(), HOOK_FORMAT_DIFF="1").returncode, 0)
        self.assertNotIn(["go", "vet", "./..."], self.calls())
        self.assertNotEqual(self.hook(self.record(), HOOK_GO_VERSION="go1.23.4").returncode, 0)
        (self.tools / "golangci-lint").unlink()
        result = self.hook(self.record())
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("missing golangci-lint", result.stderr)

    def test_nonzero_checks_propagate(self):
        self.update()
        for command in ("golangci-lint run", "go vet ./...", "go build ./...", "make test-release test-compatibility"):
            with self.subTest(command=command):
                self.assertEqual(self.hook(self.record(), HOOK_FAIL=command).returncode, 37)

    def test_git_status_failure_cannot_become_green(self):
        self.update()
        real_git = shutil.which("git")
        wrapper = self.tools / "git"
        wrapper.unlink()
        wrapper.write_text(f"#!/usr/bin/env python3\nimport os,sys\nif sys.argv[1:2] == ['status']: sys.exit(41)\nos.execv({real_git!r}, [{real_git!r}, *sys.argv[1:]])\n")
        wrapper.chmod(0o755)
        self.assertEqual(self.hook(self.record()).returncode, 41)
        self.assertEqual(self.calls(), [])

    def test_symlinked_default_hook_directory_is_preserved(self):
        hooks = self.repo / ".git/hooks"
        shutil.rmtree(hooks)
        existing = self.root / "existing hooks"
        existing.mkdir()
        active = existing / "pre-commit"
        active.write_text("existing unrelated hook")
        active.chmod(0o755)
        hooks.symlink_to(existing, target_is_directory=True)
        self.assertNotEqual(self.install().returncode, 0)
        self.assertEqual(active.read_text(), "existing unrelated hook")
        self.assertTrue(hooks.is_symlink())

    def test_autocrlf_checkout_keeps_shell_entrypoints_executable(self):
        self.git("config", "core.autocrlf", "true")
        self.git("checkout-index", "--force", "--all")
        for path in (".githooks/pre-push", "scripts/install-hooks.sh"):
            self.assertNotIn(b"\r\n", (self.repo / path).read_bytes())
        self.assertEqual(self.install().returncode, 0)
        result = self.hook(self.record(remote=ZERO))
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_installer_idempotent_and_refuses_active_hooks_or_config(self):
        self.assertEqual(self.install().returncode, 0)
        self.assertEqual(self.install().returncode, 0)
        self.assertEqual(self.git("config", "--get", "core.hooksPath"), ".githooks")
        self.git("config", "--unset", "core.hooksPath")
        active = self.repo / ".git/hooks/pre-commit"
        active.write_text("existing hook")
        self.assertNotEqual(self.install().returncode, 0)
        self.assertEqual(active.read_text(), "existing hook")
        self.git("config", "core.hooksPath", "unrelated")
        self.assertNotEqual(self.install().returncode, 0)
        self.assertEqual(self.git("config", "--get", "core.hooksPath"), "unrelated")

    def test_linked_worktree_installation(self):
        linked = self.root / "linked worktree"
        self.git("worktree", "add", "-qb", "fixture-linked", str(linked))
        self.assertEqual(self.install(cwd=linked).returncode, 0)
        self.assertEqual(self.git("config", "--get", "core.hooksPath", cwd=linked), ".githooks")
        self.assertTrue((linked / ".githooks/pre-push").is_file())
        result = subprocess.run([str(linked / ".githooks/pre-push"), "origin", "unused"],
                                cwd=linked, env=self.env, input=self.record(remote=ZERO),
                                capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn(["go", "vet", "./..."], self.calls())


if __name__ == "__main__":
    unittest.main()
