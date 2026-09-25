"""Check that the contract comparator rejects changed public observations."""

import contextlib
import copy
import io
import json
from pathlib import Path
import sys
import tempfile
import unittest
from unittest import mock

import compare_contract
from compare_contract import compare


class ContractComparison(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name) / "base"
        self.candidate = Path(self.temp.name) / "candidate"
        self.fixture = {"id": "article/html", "result": {"content": "<p>Article</p>", "title": "Article"}}
        self.error = {"id": "denied", "result": None, "code": 3, "error": "blocked", "cause": "private network"}
        for directory in (self.base, self.candidate):
            directory.mkdir()
            (directory / "api.txt").write_text("func New(opts ...Option) *Client\n")
            self.write(directory, "fixtures.jsonl", [self.fixture])
            self.write(directory, "contracts.jsonl", [self.error])

    @staticmethod
    def write(directory, name, rows):
        (directory / name).write_text("".join(json.dumps(row) + "\n" for row in rows))

    def test_equal_observations(self):
        result = compare(self.base, self.candidate)
        self.assertEqual(result["differences"], [])
        self.assertEqual(result["cases"], {"contracts.jsonl": 1, "fixtures.jsonl": 1})

    def test_content_error_and_api_changes_fail(self):
        changed = copy.deepcopy(self.fixture)
        changed["result"]["content"] = "<p>Different article</p>"
        self.write(self.candidate, "fixtures.jsonl", [changed])
        changed_error = {**self.error, "code": 1}
        self.write(self.candidate, "contracts.jsonl", [changed_error])
        (self.candidate / "api.txt").write_text("func New() *Client\n")
        differences = compare(self.base, self.candidate)["differences"]
        self.assertEqual(len(differences), 3)
        self.assertIn("article/html: result", differences)
        self.assertIn("denied: code", differences)

    def test_missing_case_fails(self):
        self.write(self.candidate, "fixtures.jsonl", [])
        self.assertIn("fixtures.jsonl: case IDs or order differ", compare(self.base, self.candidate)["differences"])


class FixtureCapture(unittest.TestCase):
    DNS = "\tresolver := &net.Resolver{}\n\taddrs, err := resolver.LookupIPAddr(ctx, hostname)\n"
    # Source-order loops no longer contain the map-range text the former control patched.
    CONVERT = "func ConvertNodeTo() {\n\tfor _, attr := range node.Nodes[0].Attr {\n\t}\n}\n"

    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)

    def repo(self, name, convert):
        repo = self.root / name
        for relative, text in (("internal/validation/url.go", self.DNS),
                               ("internal/utils/dom/convert.go", convert)):
            (repo / relative).parent.mkdir(parents=True, exist_ok=True)
            (repo / relative).write_text(text)
        (repo / "internal" / "fixtures").mkdir()
        return repo

    def test_fixture_overlay_substitutes_dns_only(self):
        repo = self.repo("repo", self.CONVERT)
        overlays = []

        def run(args, cwd, output):
            if "-overlay" in args:
                # The overlay lives in a temporary directory that capture removes on return.
                overlays.append(json.loads(Path(args[args.index("-overlay") + 1]).read_text())["Replace"])

        with mock.patch.object(compare_contract, "run", run):
            compare_contract.capture(repo, repo / "internal" / "fixtures", self.root / "out")
        validation = repo / "internal" / "validation" / "url.go"
        self.assertEqual(len(overlays), 1)
        self.assertEqual(list(overlays[0]), [str(validation)])
        self.assertEqual(validation.read_text(), self.DNS)

    def test_missing_dns_seam_fails(self):
        repo = self.repo("repo", self.CONVERT)
        (repo / "internal" / "validation" / "url.go").write_text("package validation\n")
        with mock.patch.object(compare_contract, "run"), \
                self.assertRaisesRegex(RuntimeError, "DNS comparison seam changed"):
            compare_contract.capture(repo, repo / "internal" / "fixtures", self.root / "out")

    def test_changed_dom_source_is_compared(self):
        base = self.repo("base", "func ConvertNodeTo() {}\n")
        candidate = self.repo("candidate", self.CONVERT)
        output = self.root / "comparison"

        def capture(repo, fixtures, destination, pattern):
            destination.mkdir(parents=True)
            (destination / "api.txt").write_text("func New(opts ...Option) *Client\n")
            ContractComparison.write(destination, "contracts.jsonl", [{"id": "denied"}])
            ContractComparison.write(destination, "fixtures.jsonl", [{"id": "article/html"}])

        argv = ["compare_contract.py", "--base", str(base), "--candidate", str(candidate), "--output", str(output)]
        with mock.patch.object(compare_contract, "capture", capture), mock.patch.object(sys, "argv", argv), \
                contextlib.redirect_stdout(io.StringIO()), self.assertRaises(SystemExit) as exit:
            compare_contract.main()
        self.assertFalse(exit.exception.code)
        self.assertEqual(json.loads((output / "comparison.json").read_text())["differences"], [])


if __name__ == "__main__":
    unittest.main()
