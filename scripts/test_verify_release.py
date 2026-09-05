import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("verify-release.sh")
STUB = r'''#!/usr/bin/env python3
import json
import os
from pathlib import Path
import sys

tool = Path(sys.argv[0]).name
args = sys.argv[1:]
with open(os.environ["CALL_LOG"], "a") as log:
    keys = ("GOWORK", "GOFLAGS", "GOENV", "GOOS", "GOARCH", "GOPROXY", "GOSUMDB",
            "GOPRIVATE", "GONOSUMDB", "GONOPROXY", "GOMODCACHE", "GOCACHE", "GOBIN")
    log.write(json.dumps({"tool": tool, "args": args, "cwd": os.getcwd(),
                         "env": {key: os.environ[key] for key in keys if key in os.environ}}) + "\n")
stage = tool + " " + " ".join(args[:2])
if os.environ.get("FAIL_STAGE") and stage.startswith(os.environ["FAIL_STAGE"]):
    sys.exit(7)
if tool == "curl":
    if "-o" in args:
        output = Path(args[args.index("-o") + 1])
        output.write_text(json.dumps({"Version": "v1.2.3"}) if output.suffix == ".info" else "artifact")
    elif os.environ.get("DOCS_PENDING"):
        sys.exit(22)
elif args[:2] == ["mod", "download"]:
    print(json.dumps({"Path": "github.com/BumpyClock/hermes", "Version": "v1.2.3",
                      "Sum": "" if os.environ.get("NO_SUM") else "h1:module",
                      "GoModSum": "h1:mod"}))
elif args[:2] == ["list", "-m"]:
    print(os.environ.get("SELECTED_TAG", "v1.2.3"))
elif args[:1] == ["install"]:
    binary = Path(os.environ["GOBIN"]) / "hermes"
    binary.write_text("#!/bin/sh\nprintf 'Hermes " + os.environ.get("CLI_TAG", "v1.2.3") + "\\nGo version: test\\n'\n")
    binary.chmod(0o755)
'''


class VerifyReleaseTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.bin = self.root / "tools"
        self.bin.mkdir()
        for tool in ("go", "curl"):
            path = self.bin / tool
            path.write_text(STUB)
            path.chmod(0o755)
        self.log = self.root / "calls.jsonl"
        self.env = dict(os.environ, PATH=f"{self.bin}:{os.environ['PATH']}",
                        TMPDIR=str(self.root), CALL_LOG=str(self.log),
                        GOWORK="/bad/work", GOFLAGS="-mod=vendor", GOENV="/bad/env",
                        GOOS="windows", GOARCH="arm64", GOPROXY="off",
                        GOSUMDB="off", GOPRIVATE="*", GONOSUMDB="*", GONOPROXY="*")

    def run_script(self, *args, **env):
        return subprocess.run(["bash", str(SCRIPT), *args], env=dict(self.env, **env),
                              text=True, capture_output=True)

    def calls(self):
        return [json.loads(line) for line in self.log.read_text().splitlines()]

    def test_invalid_input_has_no_external_calls(self):
        for args in ((), ("v2.0.0",), ("v1.02.3",), ("v1.2.3-rc1",),
                     ("v1.2.3", "extra"), ("1.2.3",)):
            with self.subTest(args=args):
                self.assertEqual(self.run_script(*args).returncode, 2)
                self.assertFalse(self.log.exists())

    def test_success_uses_isolated_consumer_and_tag(self):
        result = self.run_script("v1.2.3")
        self.assertEqual(result.returncode, 0, result.stderr)
        calls = self.calls()
        go_calls = [call for call in calls if call["tool"] == "go"]
        self.assertEqual([call["args"][:2] for call in go_calls], [
            ["mod", "init"], ["mod", "download"], ["mod", "edit"],
            ["mod", "tidy"], ["list", "-m"], ["build", "-o"],
            ["install", "github.com/BumpyClock/hermes/cmd/hermes@v1.2.3"]])
        for call in go_calls:
            env = call["env"]
            for key, value in {"GOWORK": "off", "GOENV": "off", "GOFLAGS": "",
                               "GOPROXY": "https://proxy.golang.org", "GOSUMDB": "sum.golang.org",
                               "GOPRIVATE": "", "GONOSUMDB": "", "GONOPROXY": ""}.items():
                self.assertEqual(env[key], value)
            self.assertNotIn("GOOS", env)
            self.assertNotIn("GOARCH", env)
            self.assertNotEqual(call["cwd"], str(SCRIPT.parent.parent))
            for key in ("GOMODCACHE", "GOCACHE", "GOBIN"):
                self.assertTrue(Path(env[key]).is_dir())
        evidence = Path(go_calls[0]["cwd"]).parent
        reported = result.stdout.splitlines()[0].removeprefix("Verification evidence retained at ")
        self.assertEqual(Path(reported).resolve(), evidence.resolve())
        self.assertTrue((evidence / "download.json").exists())
        self.assertTrue((evidence / "module/main.go").exists())
        self.assertIn("pkg.go.dev: available", result.stdout)

    def test_failures_stop_before_later_checks(self):
        for stage in ("curl", "go mod download", "go build", "go install"):
            with self.subTest(stage=stage):
                self.log.unlink(missing_ok=True)
                result = self.run_script("v1.2.3", FAIL_STAGE=stage)
                self.assertNotEqual(result.returncode, 0)
                last = self.calls()[-1]
                self.assertTrue((last["tool"] + " " + " ".join(last["args"][:2])).startswith(stage))
                self.assertIn("Verification evidence retained at", result.stdout)

    def test_missing_checksum_blocks_compile(self):
        result = self.run_script("v1.2.3", NO_SUM="1")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("authenticated checksums", result.stderr)
        self.assertEqual(self.calls()[-1]["args"][:2], ["mod", "download"])

    def test_cli_mismatch_blocks_success(self):
        result = self.run_script("v1.2.3", CLI_TAG="v1.0.0")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("CLI version mismatch", result.stderr)
        self.assertNotIn("checks passed", result.stdout)

    def test_consumer_version_mismatch_blocks_compile(self):
        result = self.run_script("v1.2.3", SELECTED_TAG="v1.2.4")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("different module version", result.stderr)
        self.assertEqual(self.calls()[-1]["args"][:2], ["list", "-m"])

    def test_docs_failure_is_nonfatal(self):
        result = self.run_script("v1.2.3", DOCS_PENDING="1")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("pkg.go.dev: pending or unavailable", result.stdout)


if __name__ == "__main__":
    unittest.main()
