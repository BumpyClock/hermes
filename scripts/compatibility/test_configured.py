import unittest
from unittest.mock import Mock

from run_configured import observations, validate_samples


class ConfiguredCompatibilityTests(unittest.TestCase):
    def test_observations_ignore_test_logging(self):
        log = Mock()
        log.read_text.return_value = '=== RUN\n    x.go:10: OBSERVATION {"id":"nyt/html","result":{"title":"Title"}}\nPASS\n'
        self.assertEqual(observations(log), {"nyt/html": {"id": "nyt/html", "result": {"title": "Title"}}})

    def test_duplicate_observations_fail(self):
        log = Mock()
        log.read_text.return_value = 'OBSERVATION {"id":"same"}\nOBSERVATION {"id":"same"}\n'
        with self.assertRaisesRegex(ValueError, "duplicate"):
            observations(log)

    def test_benchmark_sample_counts_are_required(self):
        benchmarks = {version: {str(case): {"samples": 6} for case in range(6)}
                      for version in ("baseline", "candidate")}
        validate_samples(benchmarks, 6)
        benchmarks["candidate"]["0"]["samples"] = 5
        with self.assertRaisesRegex(ValueError, "sample count"):
            validate_samples(benchmarks, 6)

    def test_missing_benchmark_workload_fails(self):
        with self.assertRaisesRegex(ValueError, "six paired"):
            validate_samples({"baseline": {}, "candidate": {}}, 6)


if __name__ == "__main__":
    unittest.main()
