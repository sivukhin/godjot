import unittest
from unittest.mock import patch

from cases import Invalid, parse
from sandbox import check


def block(command):
    return f"```bash\n# issue-repro v1\n{command}\n```"


class ReferenceChecks(unittest.TestCase):
    def test_single_argument_with_newlines_is_literal(self):
        source = "Quoted:\n$(echo hello) | text"
        self.assertEqual(parse(block(f"diff-tool render '{source}'")),
                         {"stdin": source + "\n", "reference": True})

    def test_renderer_and_arity_are_fixed(self):
        for command in ("diff-tool render", "diff-tool other input", "diff-tool render input extra",
                        "diff-tool render input\necho ok", "diff-tool render input > output",
                        "diff-tool render input --version=other"):
            with self.subTest(command=command), self.assertRaises(Invalid):
                parse(block(command))

    def test_input_byte_limit_and_nul_shared_with_expected_mode(self):
        for prefix, suffix in (("diff-tool render ", ""), ("expect-tool render ", " out")):
            boundary = "é" * 8192
            self.assertEqual(parse(block(f"{prefix}'{boundary}'{suffix}"))["stdin"], boundary + "\n")
            for invalid in (boundary + "x", "a\0b"):
                with self.subTest(helper=prefix), self.assertRaises(Invalid):
                    parse(block(f"{prefix}'{invalid}'{suffix}"))

    def test_unusable_reference_is_inconclusive(self):
        case = parse(block("diff-tool render input"))
        for failure in (("fail", "nonzero exit", b""), ("inconclusive", "timeout", b"")):
            with self.subTest(result=failure[0]), patch("sandbox.run_renderer", return_value=failure) as run:
                self.assertEqual(check("godjot", case, reference="upstream")[0], "inconclusive")
                run.assert_called_once_with("upstream", b"input\n")

    def test_compare_exact_bytes_from_independent_runs(self):
        case = parse(block("diff-tool render input"))
        for actual, status in ((b"<p>x</p>\n", "pass"), (b"<p>x</p>", "fail")):
            results = [("ok", "", b"<p>x</p>\n"), ("ok", "", actual)]
            with self.subTest(expected=status), patch("sandbox.run_renderer", side_effect=results) as run:
                self.assertEqual(check("godjot", case, reference="upstream")[0], status)
                self.assertEqual([c.args for c in run.call_args_list],
                                 [("upstream", b"input\n"), ("godjot", b"input\n")])


if __name__ == "__main__":
    unittest.main()
