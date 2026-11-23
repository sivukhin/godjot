import unittest
from unittest.mock import patch

from pr_description import END, START, sync_description, update_footer


class PRDescription(unittest.TestCase):
    sha = "a" * 40
    base = "c" * 40

    def pull(self, body="Description", sha=None, target="main", state="open", base_sha=None):
        return {"body": body, "state": state, "head": {"sha": sha or self.sha},
                "base": {"sha": base_sha or self.base, "ref": target, "repo": {"default_branch": "main"}}}

    def sync(self, ids):
        return sync_description("org/repo", 31, ids, head_sha=self.sha, base_sha=self.base)

    def test_append_replace_and_remove_preserving_authored_space(self):
        author = "Description.  \r\n\r\nFixes #29.\r\n\r\n"
        first = update_footer(author, {29, 31, 30})
        self.assertEqual(first, author + f"\n\n{START}\nFixes #30\nFixes #31\n{END}")
        self.assertEqual(update_footer(first, {29}), author)
        self.assertEqual(update_footer(first, {31}), author + f"\n\n{START}\nFixes #31\n{END}")

    def test_code_and_comments_do_not_count_as_authored_closures(self):
        input = "```markdown\nFixes #29\n```\n<!--\nFixes #30\n-->"
        self.assertIn(f"{START}\nFixes #29\nFixes #30\n{END}", update_footer(input, {29, 30}))

    def test_preserve_text_appended_below_the_footer(self):
        original = "My description"
        suffix = "\r\n\r\nMore context from review.  \r\nFixes #30"
        edited = update_footer(original, {29}) + suffix
        updated = update_footer(edited, {30, 31})
        self.assertEqual(updated, original + suffix + f"\n\n{START}\nFixes #31\n{END}")
        self.assertEqual(update_footer(updated, {30, 31}), updated)
        self.assertEqual(update_footer(edited, set()), original + suffix)

    def test_malformed_footer_never_overwrites_authored_text(self):
        for body in (START, END, START + END, f"{START}\nuser text\n{END}",
                     f"{START}\nFixes #29\n{END}inline", f"{START}\n{START}\nFixes #29\n{END}"):
            with self.subTest(body=body), self.assertRaises(ValueError):
                update_footer(body, {29})

    def test_sync_recomputes_after_an_author_edit(self):
        newer = self.pull("Updated description")
        with patch("pr_description.api", side_effect=[self.pull(), newer, newer, None]) as api:
            self.assertTrue(self.sync([30, 29, 29]))
        self.assertEqual(api.call_args.args, ("PATCH", "repos/org/repo/pulls/31",
                         {"body": f"Updated description\n\n{START}\nFixes #29\nFixes #30\n{END}"}))

    def test_skip_wrong_head_target_and_state_even_if_they_change_after_read(self):
        for change in ({"sha": "b" * 40}, {"base_sha": "d" * 40}, {"target": "release"}, {"state": "closed"}):
            with self.subTest(change=change), patch("pr_description.api", side_effect=[self.pull(), self.pull(**change)]) as api:
                self.assertFalse(self.sync([29]))
                self.assertEqual([call.args[0] for call in api.call_args_list], ["GET", "GET"])

    def test_idempotence_and_unsafe_inputs(self):
        with patch("pr_description.api", return_value=self.pull("Fixes #29")) as api:
            self.assertFalse(self.sync([29]))
            for invalid in (["29"], [True], [-1]):
                with self.subTest(invalid=invalid), self.assertRaises(ValueError):
                    self.sync(invalid)
            self.assertEqual(api.call_count, 1)
