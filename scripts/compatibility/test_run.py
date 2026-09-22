import copy
import json
from pathlib import Path
import shutil
import tarfile
import unittest
from unittest.mock import patch
import uuid

import run
from run import classify_observations, differences, observation_digest


class CompatibilityTests(unittest.TestCase):
    def test_allowance_requires_exact_before_and_after(self):
        old, new = {"id": {"value": "old"}}, {"id": {"value": "new"}}
        allowed = {"id": [observation_digest(old["id"]), observation_digest(new["id"])]}
        self.assertEqual(classify_observations(old, new, allowed),
                         {"accepted": ["id"], "unexpected": [], "stale_allowances": []})
        new["id"]["other"] = True
        self.assertEqual(classify_observations(old, new, allowed)["unexpected"], ["id"])
        old["id"]["value"] = "different baseline"
        self.assertEqual(classify_observations(old, new, allowed)["unexpected"], ["id"])

    def test_unused_allowance_is_reported(self):
        self.assertEqual(classify_observations({"id": 1}, {"id": 1}, {"id": ["a", "b"]}),
                         {"accepted": [], "unexpected": [], "stale_allowances": ["id"]})

    def test_added_and_removed_cases_are_reported(self):
        self.assertEqual(classify_observations({"old": 1}, {"new": 1}, {})["unexpected"],
                         ["new", "old"])

    def test_nested_fields_are_not_hidden(self):
        self.assertEqual(differences({"result": {"title": "old"}}, {"result": {"title": "new"}}),
                         [{"path": "/result/title", "before": "old", "after": "new"}])

    def test_json_scalar_types_are_not_equated(self):
        self.assertTrue(differences({"helpers": [True]}, {"helpers": [1]}))
        self.assertTrue(differences({"result": {"count": 1}}, {"result": {"count": 1.0}}))
        self.assertEqual(classify_observations({"id": True}, {"id": 1}, {})["unexpected"], ["id"])

    def test_exact_observations_require_correct_key_identity(self):
        with self.assertRaisesRegex(ValueError, "IDs"):
            run.exact_observations({"one": {"id": "two"}}, {"one": {"id": "two"}})

    def test_exact_observations_reject_wrong_workload_ids(self):
        with self.assertRaisesRegex(ValueError, "expected workloads"):
            run.exact_observations({"wrong": {"id": "wrong"}}, {"wrong": {"id": "wrong"}}, {"right"})

    def test_exact_observations_report_added_removed_and_changed_cases(self):
        before = {"one": {"id": "one", "content": "old"}}
        after = {"one": {"id": "one", "content": "new"}, "two": {"id": "two"}}
        self.assertEqual({row["path"] for row in run.exact_observations(before, after)},
                         {"/one/content", "/two"})


class CaptureTests(unittest.TestCase):
    def setUp(self):
        self.directory = run.ROOT / ".compatibility-runs" / ("tool-unit-" + uuid.uuid4().hex)
        self.directory.mkdir(parents=True)
        self.addCleanup(shutil.rmtree, self.directory)
        self.source = self.directory / "source"
        self.source.mkdir()
        (self.source / "parser.go").write_text("package hermes\n")

    def test_capture_freezes_dirty_regular_files(self):
        with patch("run.command", return_value="parser.go\0"):
            live = run.capture_worktree(self.directory / "candidate", self.source)
        self.assertEqual(live, {"parser.go": run.digest(self.source / "parser.go")})
        self.assertEqual((self.directory / "candidate/parser.go").read_bytes(),
                         (self.source / "parser.go").read_bytes())
        self.assertTrue(json.loads((self.directory / "candidate-capture.json").read_text())["passed"])
        with tarfile.open(self.directory / "candidate.tar") as archive:
            self.assertEqual(archive.extractfile("./parser.go").read(), b"package hermes\n")

    def test_capture_rejects_concurrent_edit_without_recapture(self):
        before = run.file_manifest(self.source)
        after = {**before, "parser.go": "changed"}
        with patch("run.worktree_manifest", side_effect=[before, after]) as capture:
            with self.assertRaisesRegex(RuntimeError, "no recapture"):
                run.capture_worktree(self.directory / "candidate", self.source)
        self.assertEqual(capture.call_count, 2)
        self.assertFalse(json.loads((self.directory / "candidate-capture.json").read_text())["passed"])

    def test_capture_rejects_copy_corruption(self):
        original_copy = shutil.copyfile

        def corrupt(source, destination):
            original_copy(source, destination)
            Path(destination).write_text("package changed\n")

        with patch("run.command", return_value="parser.go\0"), patch("run.shutil.copyfile", side_effect=corrupt):
            with self.assertRaisesRegex(RuntimeError, "no recapture"):
                run.capture_worktree(self.directory / "candidate", self.source)

    def test_runtime_drift_after_capture_fails(self):
        with patch("run.command", return_value="parser.go\0"):
            live = run.capture_worktree(self.directory / "candidate", self.source)
            guards = {"candidate": run.source_guard(self.directory / "candidate", self.directory,
                                                    "candidate", live)}
            run.check_sources(guards, self.directory, "before", self.source)
            (self.source / "parser.go").write_text("package changed\n")
            with self.assertRaisesRegex(RuntimeError, "runtime source drift"):
                run.check_sources(guards, self.directory, "after", self.source)
        self.assertFalse(json.loads((self.directory / "source-checks.json").read_text())[-1]["passed"])

    def test_captured_runtime_drift_fails_even_for_archive(self):
        guard = run.source_guard(self.source, self.directory, "baseline")
        (self.source / "parser.go").write_text("package changed\n")
        with self.assertRaisesRegex(RuntimeError, "baseline_snapshot"):
            run.check_sources({"baseline": guard}, self.directory, "after")

    def test_verification_cannot_ignore_drift_or_incomplete_capture(self):
        metadata = {"source_manifest_sha256": {"candidate": "digest"}}
        self.assertTrue(run.source_capture_failures(self.directory, metadata))
        run.write_json(self.directory / "source-checks.json",
                       [{"stage": "functional-complete", "passed": True}])
        self.assertEqual(run.source_capture_failures(self.directory, metadata), [])
        run.write_json(self.directory / "source-checks.json",
                       [{"stage": "functional-complete", "passed": True},
                        {"stage": "benchmark-0", "passed": False}])
        self.assertTrue(run.source_capture_failures(self.directory, metadata))
        self.assertEqual(run.source_capture_failures(self.directory, {}), [])

    def make_evidence(self):
        row = {"id": "case", "result": {"title": "same"}}
        for label in ("baseline", "candidate"):
            evidence = self.directory / (label + "-evidence")
            evidence.mkdir()
            for kind in ("contracts", "fixtures"):
                for repeat in ("", "-repeat"):
                    (evidence / f"{kind}{repeat}.jsonl").write_text(json.dumps(row) + "\n")
            (evidence / "api.json").write_text('{"New":"func New()"}')
        return row

    def test_pairwise_is_opt_in_and_uses_raw_evidence(self):
        self.make_evidence()
        self.assertTrue(run.acceptance(self.directory, {}, "current-sha", mode="pairwise")["passed"])
        with self.assertRaisesRegex(ValueError, "unknown comparison"):
            run.acceptance(self.directory, {}, "current-sha", mode="typo")

    def test_pairwise_rejects_changed_output_api_or_repeat(self):
        original = self.make_evidence()
        target = self.directory / "candidate-evidence"
        for filename in ("fixtures.jsonl", "contracts.jsonl", "fixtures-repeat.jsonl"):
            changed = copy.deepcopy(original)
            changed["result"]["title"] = "different"
            (target / filename).write_text(json.dumps(changed) + "\n")
            self.assertFalse(run.pairwise_acceptance(self.directory)["passed"], filename)
            (target / filename).write_text(json.dumps(original) + "\n")
        (target / "api.json").write_text('{"New":"func New(int)"}')
        self.assertFalse(run.pairwise_acceptance(self.directory)["passed"])

    def test_pairwise_rejects_incorrectly_matched_record(self):
        self.make_evidence()
        (self.directory / "candidate-evidence/fixtures.jsonl").write_text('{"id":"different","result":{}}\n')
        self.assertFalse(run.pairwise_acceptance(self.directory)["passed"])


if __name__ == "__main__":
    unittest.main()
