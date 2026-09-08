"""Check that the contract comparator rejects changed public observations."""

import copy
import json
from pathlib import Path
import tempfile
import unittest

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


if __name__ == "__main__":
    unittest.main()
