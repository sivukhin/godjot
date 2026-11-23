import argparse
import contextlib
import io
import json
import os
from pathlib import Path
import sys
import tempfile
import unittest
from unittest.mock import patch

import repro


class ReportStats(unittest.TestCase):
    def setUp(self):
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.cache = Path(tmp.name)
        self.output = self.cache / "output"
        self.summary = self.cache / "summary"
        self.revision = "a" * 40
        self.config = {"repository": "org/repo", "build": [], "maintainer_ids": [], "bot_ids": []}
        self.stack = contextlib.ExitStack()
        self.addCleanup(self.stack.close)
        self.stack.enter_context(patch.dict(os.environ, GITHUB_OUTPUT=str(self.output),
                                            GITHUB_STEP_SUMMARY=str(self.summary)))
        self.stack.enter_context(contextlib.redirect_stdout(io.StringIO()))
        self.stack.enter_context(patch("repro.git", side_effect=lambda cmd, *args:
                                      "modified" if cmd == "status" else
                                      ("main" if "--abbrev-ref" in args else self.revision)))
        self.stack.enter_context(patch("repro.build", return_value="image"))

    def run_cases(self, statuses, invalid=False):
        entries = [dict(number=i, state="OPEN", comment=42, case={"stdin": "x", "expected": "x"})
                   for i in range(1, len(statuses) + 1)]
        if invalid:
            entries.append(dict(number=42, state="OPEN", comment=43, invalid="bad case"))
        results = [(s, "observation") if isinstance(s, str) else s for s in statuses]
        with patch("repro.check", side_effect=results):
            return repro.replay(self.config, argparse.Namespace(cache=self.cache, show_rows=False),
                                entries)

    def reports(self):
        reports = []
        for line in self.output.read_text().splitlines():
            key, data = line.split("=", 1)
            self.assertEqual(key, "stats")
            reports.append(json.loads(data))
        return reports

    def test_transitions_with_invalid_and_operational_errors(self):
        self.run_cases(["pass", "fail", "fail", "pass", "pass", "inconclusive"])
        self.assertEqual(self.reports()[0]["counts"]["new"], 5)
        self.assertEqual(self.reports()[0]["fixes"], [])
        code = self.run_cases(["pass", "fail", "pass", "fail", "inconclusive", "pass"], invalid=True)
        self.assertEqual(code, 1)
        self.assertEqual(self.reports()[1], {
            "revision": self.revision,
            "counts": {"new": 1, "unchanged": 2, "fixes": 1, "regressions": 1, "inconclusive": 2},
            "fixes": [3], "regressions": [4],
        })
        self.assertIn("2 unchanged", self.summary.read_text())

    def test_partial_run_cannot_publish_or_advance_baseline(self):
        with self.assertRaisesRegex(RuntimeError, "interrupted"):
            self.run_cases(["fail", RuntimeError("interrupted")])
        self.assertFalse(self.output.exists())
        self.run_cases(["pass"])
        self.assertEqual(self.reports()[0]["counts"]["new"], 1)

    def test_cap_examples_preserves_full_counts(self):
        stats = repro.new_stats(self.revision)
        for n in range(1, 103):
            repro.record_result(stats, n, "pass", ("fail", self.revision))
        self.assertEqual(stats["fixes"], list(range(1, 101)))
        self.assertEqual(stats["counts"]["fixes"], 102)

    def test_empty_cached_selection_reports_zeroes(self):
        (self.cache / ".issue-repro.toml").write_text(
            'repository="org/repo"\nmaintainer_ids=[]\nbot_ids=[]\nbuild=[]\n')
        (self.cache / "cases.json").write_text(json.dumps({
            "policy": repro.digest(["org/repo", [], []]), "entries": [],
        }))
        with patch.object(repro, "PROJECT", self.cache), patch.object(
            sys, "argv", ["repro.py", "--cache", str(self.cache), "run"]
        ):
            self.assertEqual(repro.main(), 0)
        self.assertEqual(sum(self.reports()[0]["counts"].values()), 0)
