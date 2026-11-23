"""Publish validated counts from the read-only runner to one pull request comment."""

import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import sys

from summary import change_summary

PR_MARKER = "<!-- godjot:issue-repro-pr:v1 -->"
BOT_ID = 41898282  # https://api.github.com/users/github-actions%5Bbot%5D
MAX_BYTES = 16 * 1024
MAX_ISSUES = 100
CATEGORIES = {"unchanged", "fixes", "regressions", "new", "inconclusive"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--file", type=Path, help="read the report from this file instead of REPORT_JSON")
    parser.add_argument("--dry-run", action="store_true", help="print the PR comment without contacting GitHub")
    parser.add_argument("--pr", required=True, type=int, help="publish a result comment on this pull request")
    args = parser.parse_args()
    if args.file:
        with args.file.open("rb") as stream:
            data = stream.read(MAX_BYTES + 1)
    else:
        data = os.environ["REPORT_JSON"].encode()
    publish(data, os.environ["GITHUB_REPOSITORY"], os.environ["GITHUB_RUN_ID"], args.pr, args.dry_run)


def publish(data, repo, run_id, pr, dry_run=False):
    report = read_report(data)
    if not re.fullmatch(r"[A-Za-z0-9_-]+/[A-Za-z0-9_][A-Za-z0-9._-]*", repo):
        raise ValueError("invalid GitHub repository")
    if not re.fullmatch(r"[1-9][0-9]{0,19}", run_id):
        raise ValueError("invalid Actions run ID")

    if type(pr) is not int or not 1 <= pr <= 2**31 - 1:
        raise ValueError("invalid pull request number")
    issues = f"repos/{repo}/issues"
    summary = f"{PR_MARKER}\n\n### Issue replay results\n\n{render_summary(report, repo)}"
    run_link = f"[Actions run {run_id}](https://github.com/{repo}/actions/runs/{run_id})"
    body = f"{summary}\n\n{run_link}"
    if dry_run:
        print(body)
        return
    comments = f"{issues}/{pr}/comments"
    existing = find_comment(comments)
    if existing is None:
        api("POST", comments, {"body": body})
    elif existing["body"] != body:
        api("PATCH", f"{issues}/comments/{existing['id']}", {"body": body})


def read_report(data):
    if len(data) > MAX_BYTES:
        raise ValueError("report too large")
    report = json.loads(data, object_pairs_hook=unique_keys)
    fields = {"revision", "counts", "fixes", "regressions"}
    if type(report) is not dict or set(report) not in (fields, fields | {"outcomes"}):
        raise ValueError("invalid report fields")
    revision = report["revision"]
    if not isinstance(revision, str) or not re.fullmatch("[a-f0-9]{40}", revision):
        raise ValueError("invalid revision")
    counts = report["counts"]
    if (type(counts) is not dict or set(counts) != CATEGORIES or
            any(type(n) is not int or not 0 <= n <= 2**31 - 1 for n in counts.values())):
        raise ValueError("invalid counts")
    if "outcomes" in report:
        outcomes = report["outcomes"]
        if (type(outcomes) is not dict or set(outcomes) != {"pass", "fail"} or
                any(type(n) is not int or not 0 <= n <= 2**31 - 1 for n in outcomes.values()) or
                sum(outcomes.values()) != sum(counts.values()) - counts["inconclusive"]):
            raise ValueError("invalid outcome totals")
    seen = set()
    for name in ("fixes", "regressions"):
        issues = report[name]
        if (type(issues) is not list or len(issues) > min(MAX_ISSUES, counts[name]) or
                any(type(n) is not int or not 1 <= n <= 2**31 - 1 for n in issues)):
            raise ValueError("invalid issue list")
        if len(set(issues)) != len(issues) or not seen.isdisjoint(issues):
            raise ValueError("duplicate issue")
        seen.update(issues)
    return report


def unique_keys(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON field")
        result[key] = value
    return result


def find_comment(endpoint):
    page = 1
    found = None
    while True:
        batch = api("GET", f"{endpoint}?per_page=100&page={page}")
        for item in batch:
            author = item.get("user") or {}
            if (author.get("id") != BOT_ID or author.get("type") != "Bot" or
                    (item.get("body") or "").partition("\n")[0] != PR_MARKER):
                continue
            if found is not None:
                raise ValueError("multiple bot-authored entries with the same marker")
            found = item
        if len(batch) < 100:
            return found
        page += 1


def render_summary(report, repo):
    counts = report["counts"]
    rev = report["revision"]
    lines = [change_summary(report), "", f"Tested [{rev[:12]}](https://github.com/{repo}/commit/{rev})."]
    for key, label in (("fixes", "Fixed cases"), ("regressions", "Regressed cases")):
        if report[key]:
            refs = ", ".join(f"[#{n}](https://github.com/{repo}/issues/{n})" for n in sorted(report[key]))
            more = counts[key] - len(report[key])
            lines.extend(["", f"**{label}** · {refs}" + (f" (and {more} more)" if more else "")])
    return "\n".join(lines)


def api(method, endpoint, payload=None):
    argv = ["gh", "api", "--method", method, endpoint]
    if payload is not None:
        argv.extend(["--input", "-"])
    response = subprocess.check_output(argv, input=None if payload is None else json.dumps(payload).encode(),
                                       timeout=45)
    if method == "GET":
        return json.loads(response)


if __name__ == "__main__":
    try:
        main()
    except (KeyError, OSError, RecursionError, ValueError, subprocess.SubprocessError) as exc:
        sys.exit(str(exc))
