"""Publish validated counts from the read-only runner to one tracking issue."""

import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import sys

MARKER = "<!-- godjot:issue-repro-dashboard:v1 -->"
TITLE = "Issue reproducer dashboard"
BOT_ID = 41898282  # https://api.github.com/users/github-actions%5Bbot%5D
MAX_BYTES = 16 * 1024
MAX_ISSUES = 100
CATEGORIES = {"unchanged", "fixes", "regressions", "new", "inconclusive"}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--file", type=Path, help="read the report from this file instead of REPORT_JSON")
    parser.add_argument("--dry-run", action="store_true", help="print the dashboard without contacting GitHub")
    args = parser.parse_args()
    if args.file:
        with args.file.open("rb") as stream:
            data = stream.read(MAX_BYTES + 1)
    else:
        data = os.environ["REPORT_JSON"].encode()
    publish(data, os.environ["GITHUB_REPOSITORY"], os.environ["GITHUB_RUN_ID"], args.dry_run)


def publish(data, repo, run_id, dry_run=False):
    report = read_report(data)
    if not re.fullmatch(r"[A-Za-z0-9_-]+/[A-Za-z0-9_][A-Za-z0-9._-]*", repo):
        raise ValueError("invalid GitHub repository")
    if not re.fullmatch(r"[1-9][0-9]{0,19}", run_id):
        raise ValueError("invalid Actions run ID")

    issues = f"repos/{repo}/issues"
    summary = render_summary(report, repo)
    run_link = f"[Actions run {run_id}](https://github.com/{repo}/actions/runs/{run_id})"
    body = f"{summary}\n\n{run_link}"
    if dry_run:
        print(body)
        return
    dashboard = find_dashboard(issues)
    if dashboard is None:
        api("POST", issues, {"title": TITLE, "body": body})
        print("Created the issue reproducer dashboard.")
    else:
        # Preserve the linked run when a repeated check produces the same
        # report on the same revision. Avoid updating the issue each night.
        old_summary, _, last_link = dashboard["body"].rpartition("\n\n")
        expected_link = rf"\[Actions run ([1-9][0-9]*)\]\(https://github\.com/{re.escape(repo)}/actions/runs/\1\)"
        if old_summary == summary and re.fullmatch(expected_link, last_link):
            print(f"Dashboard #{dashboard['number']} already has this report.")
            return
        api("PATCH", f"{issues}/{dashboard['number']}", {"body": body})
        print(f"Updated dashboard #{dashboard['number']}.")


def read_report(data):
    if len(data) > MAX_BYTES:
        raise ValueError("report too large")
    report = json.loads(data, object_pairs_hook=unique_keys)
    if type(report) is not dict or set(report) != {"revision", "counts", "fixes", "regressions"}:
        raise ValueError("invalid report fields")
    revision = report["revision"]
    if not isinstance(revision, str) or not re.fullmatch("[a-f0-9]{40}", revision):
        raise ValueError("invalid revision")
    counts = report["counts"]
    if (type(counts) is not dict or set(counts) != CATEGORIES or
            any(type(n) is not int or not 0 <= n <= 2**31 - 1 for n in counts.values())):
        raise ValueError("invalid counts")
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


def find_dashboard(issues):
    page = 1
    found = None
    while True:
        # The REST issues listing also includes pull requests.
        batch = api("GET", f"{issues}?creator=github-actions%5Bbot%5D&state=all&per_page=100&page={page}")
        for issue in batch:
            if ("pull_request" in issue or issue["user"]["id"] != BOT_ID or
                    issue["user"]["type"] != "Bot" or
                    (issue.get("body") or "").partition("\n")[0] != MARKER):
                continue
            if found is not None:
                raise ValueError("multiple bot-authored dashboards found")
            found = issue
        if len(batch) < 100:
            return found
        page += 1


def render_summary(report, repo):
    counts = report["counts"]
    rev = report["revision"]
    lines = [MARKER, "", f"Reproducer results for [{rev[:12]}](https://github.com/{repo}/commit/{rev}).",
             "", "| Result | Cases |", "| --- | ---: |"]
    for key, label in (("fixes", "Fix candidates"), ("regressions", "Regression candidates"),
                       ("new", "New baselines"), ("unchanged", "Unchanged"), ("inconclusive", "Inconclusive")):
        lines.append(f"| {label} | {counts[key]} |")
    for key, label in (("fixes", "Review possible fixes"), ("regressions", "Review possible regressions")):
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
