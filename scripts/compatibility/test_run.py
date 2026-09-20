import unittest

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


if __name__ == "__main__":
    unittest.main()
