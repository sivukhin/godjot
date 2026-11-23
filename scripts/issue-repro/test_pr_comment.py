import contextlib
import io
import json
import unittest
from unittest.mock import patch

import pr_comment


class PRComment(unittest.TestCase):
    def setUp(self):
        self.report = {
            "revision": "b" * 40,
            "counts": {"unchanged": 8, "fixes": 1, "regressions": 0, "new": 2, "inconclusive": 3},
            "fixes": [18], "regressions": [],
        }
        self.api = self.enterContext(patch("pr_comment.api"))
        self.output = self.enterContext(contextlib.redirect_stdout(io.StringIO()))

    def publish(self, run="123", pr=88, **opts):
        pr_comment.publish(json.dumps(self.report).encode(), "org/repo", run, pr, **opts)

    def comment(self, body=pr_comment.PR_MARKER, bot=pr_comment.BOT_ID):
        return {"id": 50, "body": body, "user": {"id": bot, "type": "Bot"}}

    def test_create_pr_comment_ignores_other_authors_and_embedded_markers(self):
        self.api.return_value = [self.comment(bot=1, body=pr_comment.PR_MARKER),
                                 self.comment(body="See below\n" + pr_comment.PR_MARKER),
                                 self.comment(body="other bot comment")]
        self.publish(pr=88)
        self.assertTrue(self.api.call_args_list[0].args[1].startswith("repos/org/repo/issues/88/comments?"))
        method, endpoint, payload = self.api.call_args.args
        self.assertEqual((method, endpoint), ("POST", "repos/org/repo/issues/88/comments"))
        self.assertTrue(payload["body"].startswith(pr_comment.PR_MARKER + "\n\n### Issue replay results"))
        self.assertIn("actions/runs/123)", payload["body"])

    def test_existing_pr_comment_paginates_and_updates_latest_run(self):
        self.api.side_effect = [[self.comment(body="discussion")] * 100,
                                [dict(self.comment(body=pr_comment.PR_MARKER), id=12345)], None]
        self.publish(pr=88)
        method, endpoint, payload = self.api.call_args.args
        self.assertEqual((method, endpoint), ("PATCH", "repos/org/repo/issues/comments/12345"))
        self.assertIn("page=2", self.api.call_args_list[1].args[1])
        self.api.reset_mock()
        self.api.side_effect = [[dict(self.comment(body=payload["body"]), id=12345)], None]
        self.publish(run="456", pr=88)
        self.assertEqual(self.api.call_args.args[:2], (method, endpoint))
        self.assertIn("actions/runs/456)", self.api.call_args.args[2]["body"])

    def test_pr_number_is_validated_before_contacting_github(self):
        for pr in (0, -2, True, "5/../6"):
            with self.subTest(pr=pr), self.assertRaises(ValueError):
                self.publish(pr=pr)
        self.api.assert_not_called()

    def test_dry_run_prints_without_gh(self):
        self.publish(dry_run=True)
        self.api.assert_not_called()
        self.assertIn("**Fixed** 1 · **Unchanged** 8 · **Regressed** 0", self.output.getvalue())

    def test_optional_outcomes_sum_to_conclusive_cases(self):
        self.report["outcomes"] = {"pass": 4, "fail": 7}
        self.publish(pr=88, dry_run=True)
        self.assertIn("First result for 2 cases.", self.output.getvalue())
        for invalid in ({"pass": 11, "fail": True}, {"pass": 0, "fail": 14},
                        {"pass": -1, "fail": 12}, {"pass": 11}):
            with self.subTest(outcomes=invalid), self.assertRaises(ValueError):
                self.report["outcomes"] = invalid
                self.publish(pr=88)
        self.api.assert_not_called()

    def test_duplicate_comments_fail_closed(self):
        self.api.return_value = [self.comment(), dict(self.comment(), id=51)]
        with self.assertRaises(ValueError):
            self.publish()
        self.assertEqual(self.api.call_count, 1)

    def test_reject_untrusted_fields_and_counts_before_any_api_call(self):
        for altered in (
            {"revision": "a" * 40 + "\n@anyone"},
            {"body": "attacker content"},
            {"counts": dict(self.report["counts"], fixes=True)},
            {"fixes": ["18"]},
            {"fixes": [19, 20]},
            {"fixes": [0]},
            {"counts": dict(self.report["counts"], regressions=1), "regressions": [18]},
            {"fixes": list(range(5000))},
        ):
            with self.subTest(change=list(altered)), self.assertRaises(ValueError):
                data = json.dumps(dict(self.report, **altered)).encode()
                pr_comment.publish(data, "org/repo", "123", 88)
        self.api.assert_not_called()


if __name__ == "__main__":
    unittest.main()
