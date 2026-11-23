"""Maintain closing references for approved reproducers in a pull request body."""

import re

from github import api

START = "<!-- godjot:issue-repro-fixes:start -->"
END = "<!-- godjot:issue-repro-fixes:end -->"
DIRECTIVE = re.compile(r"(?i)^ {0,3}(?:close[sd]?|fix(?:es|ed)?|resolve[sd]?):? +#([1-9][0-9]{0,9})[.!]? *$")


def sync_description(repo, pr, fixed, *, head_sha, base_sha=None):
    """Update a current open PR against its default branch. Return whether it changed."""
    if not re.fullmatch(r"[A-Za-z0-9_-]+/[A-Za-z0-9_][A-Za-z0-9._-]*", repo):
        raise ValueError("invalid repository")
    if type(pr) is not int or not 1 <= pr <= 2**31 - 1:
        raise ValueError("invalid pull request number")
    if not re.fullmatch("[a-f0-9]{40}", head_sha):
        raise ValueError("invalid PR head SHA")
    if base_sha is not None and not re.fullmatch("[a-f0-9]{40}", base_sha):
        raise ValueError("invalid PR base SHA")
    ids = set()
    for n in fixed:
        if type(n) is not int or not 1 <= n <= 2**31 - 1:
            raise ValueError("invalid fixed issue number")
        ids.add(n)

    endpoint = f"repos/{repo}/pulls/{pr}"
    # The pulls API replaces the entire body. Recheck the source immediately
    # before writing, and recompute from a newer body if the author edited it.
    previous = None
    for _ in range(3):
        pull = api("GET", endpoint)
        if (pull["state"] != "open" or pull["head"]["sha"] != head_sha or
                pull["base"]["ref"] != pull["base"]["repo"]["default_branch"] or
                (base_sha is not None and pull["base"]["sha"] != base_sha)):
            return False
        original = pull["body"] or ""
        body = update_footer(original, ids)
        if body == original:
            return False
        if original == previous:
            api("PATCH", endpoint, {"body": body})
            return True
        previous = original
    raise ValueError("PR description kept changing during replay publication")


def update_footer(body, fixed):
    author = body
    if START in body or END in body:
        author, _, rest = body.partition(START)
        managed, _, suffix = rest.partition(END)
        if (body.count(START) != 1 or body.count(END) != 1 or
                not re.fullmatch(r"(?:\r?\nFixes #[1-9][0-9]{0,9})+\r?\n", managed) or
                (suffix and not suffix.startswith(("\n", "\r\n")))):
            raise ValueError("malformed managed footer; preserving the PR description")
        if author.endswith("\r\n\r\n"):
            author = author[:-4]
        elif author.endswith("\n\n"):
            author = author[:-2]
        elif author:
            raise ValueError("managed footer must start after a blank line")
        # A maintainer may have continued the description below our markers.
        author += suffix

    remaining = fixed - authored_closures(author)
    if not remaining:
        return author
    footer = "\n".join([START, *(f"Fixes #{n}" for n in sorted(remaining)), END])
    return author + "\n\n" + footer if author else footer


def authored_closures(body):
    # A standalone closing line in a quoted code example does not close the issue.
    found = set()
    fence = None
    for line in re.sub(r"<!--.*?-->", "", body, flags=re.S).splitlines():
        if fence:
            if re.fullmatch(r" {0,3}" + fence[0] + "{" + str(len(fence)) + r",}[ \t]*", line):
                fence = None
        elif match := re.match(r" {0,3}(`{3,}|~{3,})", line):
            fence = match[1]
        elif match := DIRECTIVE.fullmatch(line):
            found.add(int(match[1]))
    return found
