import contextlib
import io
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

from cases import Invalid, parse
from github import scan
import repro


def block(command):
    return f"```bash\n# issue-repro v1\n{command}\n```"


class Cases(unittest.TestCase):
    def test_render_treats_quoted_arguments_as_data(self):
        case = parse(block("expect-tool render '$(command) | text' '<p>text</p>'"))
        self.assertEqual(case, {"stdin": "$(command) | text\n", "expected": "<p>text</p>\n"})

    def test_other_helpers_and_shell_operations_rejected(self):
        for command in ("echo ok", "expect-tool other in out", "expect-tool render in out > file"):
            with self.subTest(command=command), self.assertRaises(Invalid):
                parse(block(command))

    def test_marker_must_be_first_line_of_eligible_fence(self):
        command = "expect-tool render in out"
        for comment in ("\n".join(["````markdown", block(command), "````"]),
                        block(command).replace("# issue", "extra\n# issue"),
                        block(command).replace("issue-repro", "turso-repro")):
            self.assertIsNone(parse(comment))


class IssueCollection(unittest.TestCase):
    def test_only_configured_actor_ids_are_eligible(self):
        def user(uid):
            return {"__typename": "User", "databaseId": uid}

        def marked(comment_id, author):
            return dict(databaseId=comment_id, body=block("expect-tool render x y"),
                        lastEditedAt=None, author=author)

        comments = [marked(10, user(100)), marked(11, user(200)),
                    marked(12, {"__typename": "Bot", "id": "100"}), marked(13, None)]
        issue = dict(number=19, state="CLOSED",
                     comments=dict(nodes=comments, pageInfo=dict(hasPreviousPage=False)))
        repo = {"owner": {"databaseId": 200}, "issue": issue}
        response = json.dumps({"data": {"repository": repo}})
        with patch("github.subprocess.check_output", return_value=response):
            entries = scan("org/repo", "closed", [], {19}, 2, [100], [])
            self.assertEqual([e["comment"] for e in entries], [10])
            self.assertEqual(scan("org/repo", "closed", [], {19}, 2, [], []), [])

    def test_edited_bot_reproducer_requires_fresh_comment(self):
        def comment(bot_id, edited=None):
            return dict(databaseId=123, body=block("expect-tool render x y"), lastEditedAt=edited,
                        author={"id": bot_id, "__typename": "Bot"})
        newest = dict(nodes=[comment("other")], pageInfo=dict(hasPreviousPage=True, startCursor="older"))
        older = dict(nodes=[comment("bot"), comment("bot", "yesterday")], pageInfo=dict(hasPreviousPage=False))
        pages = [dict(number=7, state="OPEN", comments=newest), dict(comments=older)]
        def response(*args, **kwargs):
            if len(pages) == 1:
                self.assertEqual(json.loads(kwargs["input"])["variables"]["cursor"], "older")
            return json.dumps({"data": {"repository": {"issue": pages.pop(0)}}})
        with patch("github.subprocess.check_output", side_effect=response):
            entries = scan("org/repo", "all", [], {7}, 4, [], ["bot"])
        self.assertIn("invalid", entries[0])
        self.assertNotIn("case", entries[0])

    def test_issue_pages_and_filters(self):
        comment = dict(databaseId=99, body=block("expect-tool render x y"), lastEditedAt=None,
                       author={"id": "publisher", "__typename": "Bot"})
        issue = dict(number=3, state="OPEN", comments=dict(nodes=[comment], pageInfo=dict(hasPreviousPage=False)))
        pages = [dict(nodes=[], pageInfo=dict(hasNextPage=True, endCursor="next")),
                 dict(nodes=[issue], pageInfo=dict(hasNextPage=False))]
        cursors = []
        def response(*args, **kwargs):
            variables = json.loads(kwargs["input"])["variables"]
            self.assertEqual(variables["states"], ["OPEN"])
            self.assertEqual(variables["labels"], ["bug"])
            cursors.append(variables["cursor"])
            return json.dumps({"data": {"repository": {"issues": pages.pop(0)}}})
        with patch("github.subprocess.check_output", side_effect=response):
            entries = scan("org/repo", "open", ["bug"], set(), 2, [], ["publisher"])
        self.assertEqual(cursors, [None, "next"])
        self.assertEqual([e["number"] for e in entries], [3])


class LocalRunner(unittest.TestCase):
    def setUp(self):
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.project = Path(tmp.name) / "godjot"
        self.cache = Path(tmp.name) / "cache"
        self.project.mkdir()
        self.git("init", "--quiet", "--initial-branch=main")
        self.git("-c", "user.name=Tester", "-c", "user.email=test@example.org",
                 "commit", "--quiet", "--allow-empty", "-m", "initial")
        self.git("remote", "add", "origin", "https://github.com/sivukhin/godjot.git")
        (self.project / ".issue-repro.toml").write_text(
            'repository = "sivukhin/godjot"\nbuild = ["go", "build"]\n'
            'maintainer_ids = [100]\nbot_ids = []\n')
        self.project_patch = patch.object(repro, "PROJECT", self.project)
        self.project_patch.start()
        self.addCleanup(self.project_patch.stop)
        case = parse(block("expect-tool render 'one' '<p>one</p>'"))
        entry = dict(number=7, state="OPEN", comment=42, case=case)
        with patch("repro.scan", return_value=[entry]):
            self.assertEqual(self.cli("scan")[0], 0)

    def git(self, *args):
        subprocess.run(["git", *args], cwd=self.project, check=True, capture_output=True)

    def cli(self, *args):
        argv = ["repro.py", "--cache", str(self.cache), *args]
        out = io.StringIO()
        with patch.object(sys, "argv", argv), contextlib.redirect_stdout(out):
            status = repro.main()
        return status, out.getvalue()

    def test_removed_publisher_invalidates_cached_cases(self):
        config = self.project / ".issue-repro.toml"
        config.write_text(config.read_text().replace("maintainer_ids = [100]", "maintainer_ids = []"))
        with patch("repro.build") as build, contextlib.redirect_stderr(io.StringIO()):
            with self.assertRaises(SystemExit):
                self.cli("run")
            build.assert_not_called()

    def test_failed_scan_keeps_last_snapshot(self):
        snapshot = self.cache / "cases.json"
        previous = snapshot.read_bytes()
        with patch("repro.scan", side_effect=subprocess.TimeoutExpired("gh", 45)):
            with self.assertRaises(subprocess.TimeoutExpired):
                self.cli("scan")
        self.assertEqual(snapshot.read_bytes(), previous)

    def test_inconclusive_keeps_history_and_branches_have_separate_baselines(self):
        with patch("repro.build", return_value="image"), patch("repro.check") as engine:
            def run(status):
                engine.return_value = (status, "test observation")
                return self.cli("run")

            self.assertNotIn("candidate", run("fail")[1])
            self.assertEqual(run("inconclusive")[0], 1)
            self.assertIn("fix candidate", run("pass")[1])
            self.git("switch", "--quiet", "-c", "topic")
            self.assertNotIn("candidate", run("fail")[1])
            self.git("switch", "--quiet", "main")
            code, report = run("fail")
            self.assertEqual(code, 1)
            self.assertIn("regression candidate", report)
            self.assertIn("#issuecomment-42", report)


if __name__ == "__main__":
    unittest.main()
