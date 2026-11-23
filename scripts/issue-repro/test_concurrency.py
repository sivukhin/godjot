import unittest

from cases import Invalid, parse


class ConcurrencyCases(unittest.TestCase):
    def test_named_suites(self):
        for suite in ("parser-state", "parallel-render"):
            with self.subTest(suite=suite):
                self.assertEqual(parse(block(f"check-concurrency {suite}")), {"suite": suite})

    def test_rejects_other_suites_or_arguments(self):
        for command in (
            "check-concurrency",
            "check-concurrency ./parallel-render",
            "check-concurrency unknown",
            "check-concurrency parser-state -count=100",
            "check-concurrency parser-state\ncheck-concurrency parallel-render",
            "check-concurrency '$(parallel-render)'",
        ):
            with self.subTest(command=command), self.assertRaises(Invalid):
                parse(block(command))


def block(command):
    return f"```bash\n# issue-repro v1\n{command}\n```"
