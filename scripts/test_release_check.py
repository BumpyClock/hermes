import contextlib
import io
import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import release_check as check


SHA = "a" * 40


class ReleaseChecks(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        (self.root / "cmd/hermes").mkdir(parents=True)
        (self.root / "go.mod").write_text("module github.com/BumpyClock/hermes\n")
        self.set_version("1.1.0")
        self.responses = {
            ("git", "rev-parse", "--show-toplevel"): str(self.root),
            ("git", "branch", "--show-current"): "master",
            ("git", "status", "--porcelain", "--untracked-files=all"): "",
            ("git", "remote", "get-url", "--all", "origin"): "https://github.com/BumpyClock/hermes.git",
            ("git", "remote", "get-url", "--push", "--all", "origin"): "https://github.com/BumpyClock/hermes.git",
            ("gh", "api", "user", "--jq", ".login"): "BumpyClock",
            ("gh", "api", "repos/BumpyClock/hermes"): json.dumps({"full_name": "BumpyClock/hermes", "default_branch": "master"}),
            ("git", "rev-parse", "HEAD"): SHA,
            ("git", "ls-remote", "--heads", "origin", "refs/heads/master"): SHA + "\trefs/heads/master",
            ("git", "ls-remote", "--tags", "--refs", "origin"): SHA + "\trefs/tags/v1.1.0",
            ("git", "tag", "--list", "v1.1.1"): "",
            ("gh", "api", "--paginate", "--slurp", "repos/BumpyClock/hermes/releases?per_page=100"): "[[]]",
            ("git", "ls-files", "-z"): "cmd/hermes/main.go\0",
            ("file", "-b", "--", str(self.root / "cmd/hermes/main.go")): "ASCII text",
        }
        self.calls = []
        self.enterContext(patch.object(check, "run", side_effect=self.run_command))
        self.enterContext(patch.object(Path, "cwd", return_value=self.root))
        self.enterContext(contextlib.redirect_stdout(io.StringIO()))

    def set_version(self, version):
        (self.root / "cmd/hermes/main.go").write_text(f'fmt.Println("Hermes v{version}")\n')
        (self.root / "CHANGELOG.md").write_text(f"## [{version}] - 2026-09-05\n\n### Fixed\n- Regression.\n")

    def run_command(self, *args):
        self.calls.append(args)
        response = self.responses[args]
        if isinstance(response, Exception):
            raise response
        return response

    def test_preflight_and_ready(self):
        check.preflight("1.1.1")
        self.set_version("1.1.1")
        check.preflight("1.1.1", ready=True)

    def test_invalid_versions_fail_before_commands(self):
        for version in ("v1.1.1", "01.1.1", "1.1.1-rc.1", "2.0.0", "1.1.1; touch /tmp/x"):
            with self.subTest(version=version), self.assertRaises(check.CheckError):
                check.preflight(version)
        self.assertEqual(self.calls, [])

    def test_unsafe_repository_states(self):
        cases = [
            (("git", "branch", "--show-current"), "feature", "branch"),
            (("git", "status", "--porcelain", "--untracked-files=all"), " M user.txt", "clean"),
            (("git", "remote", "get-url", "--push", "--all", "origin"), "https://example.com/other", "origin"),
            (("gh", "api", "user", "--jq", ".login"), "other", "Authenticate"),
            (("git", "ls-remote", "--heads", "origin", "refs/heads/master"), "b" * 40 + "\trefs/heads/master", "remote master"),
            (("git", "tag", "--list", "v1.1.1"), "v1.1.1", "locally"),
            (("git", "ls-remote", "--tags", "--refs", "origin"), SHA + "\trefs/tags/v1.1.1", "remotely"),
            (("gh", "api", "--paginate", "--slurp", "repos/BumpyClock/hermes/releases?per_page=100"), '[[{"tag_name":"v1.1.1"}]]', "release already exists"),
            (("gh", "api", "--paginate", "--slurp", "repos/BumpyClock/hermes/releases?per_page=100"), check.CheckError("network failure"), "network failure"),
            (("git", "ls-remote", "--tags", "--refs", "origin"), SHA + "\trefs/tags/v1.2.0", "exceed"),
        ]
        for command, response, message in cases:
            with self.subTest(command=command, response=str(response)):
                original = self.responses[command]
                self.responses[command] = response
                with self.assertRaisesRegex(check.CheckError, message):
                    check.preflight("1.1.1")
                self.responses[command] = original

    def test_ready_requires_target_source_version(self):
        with self.assertRaisesRegex(check.CheckError, "CLI version"):
            check.preflight("1.1.1", ready=True)

    def test_changelog_mismatch_and_invalid_date(self):
        for heading in ("## [1.0.0] - 2026-09-05", "## [1.1.0] - 2026-02-30", "## [1.1.0]"):
            (self.root / "CHANGELOG.md").write_text(heading)
            with self.assertRaisesRegex(check.CheckError, "changelog"):
                check.preflight("1.1.1")

    def test_stale_or_unversioned_binary_blocks(self):
        binary = str(self.root / "cmd/hermes/hermes")
        self.responses[("git", "ls-files", "-z")] = "cmd/hermes/hermes\0"
        self.responses[("file", "-b", "--", binary)] = "PE32+ executable"
        for contents in ("Hermes v0.1.0", "no version", "Hermes v1.1.0 Hermes v0.1.0"):
            self.responses[("strings", "--", binary)] = contents
            with self.assertRaisesRegex(check.CheckError, "Tracked binary"):
                check.preflight("1.1.1")
        self.responses[("strings", "--", binary)] = "Hermes v1.1.0"
        check.preflight("1.1.1")

    def ci_response(self, runs):
        self.responses[("gh", "run", "list", "--repo", "BumpyClock/hermes", "--workflow", "ci.yml",
                        "--branch", "master", "--event", "push", "--commit", SHA, "--limit", "100",
                        "--json", "databaseId,headSha,status,conclusion,url")] = json.dumps(runs)

    def test_ci_requires_latest_exact_sha_success(self):
        success = {"databaseId": 1, "headSha": SHA, "status": "completed", "conclusion": "success", "url": "https://github.com/run/1"}
        self.ci_response([success])
        check.check_ci(SHA)
        for runs in ([], [dict(success, headSha="b" * 40)],
                     [dict(success, conclusion="failure")], [dict(success, conclusion="skipped")],
                     [dict(success, status="in_progress")],
                     [success, dict(success, databaseId=2, conclusion="failure")]):
            self.ci_response(runs)
            with self.assertRaises(check.CheckError):
                check.check_ci(SHA)


class CommandFailures(unittest.TestCase):
    def test_command_failure_does_not_leak_remote_credentials(self):
        result = subprocess.CompletedProcess([], 128, "", "fatal: https://secret@example.com")
        with patch.object(subprocess, "run", return_value=result):
            with self.assertRaisesRegex(check.CheckError, r"git remote failed \(exit 128\)") as error:
                check.run("git", "remote", "get-url", "origin")
        self.assertNotIn("secret", str(error.exception))


if __name__ == "__main__":
    unittest.main()
