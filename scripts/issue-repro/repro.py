#!/usr/bin/env python3
"""Scan GitHub issues or replay a saved selection against this checkout."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import tomllib

from cases import parse
from github import scan
from sandbox import build, check

HERE = Path(__file__).resolve().parent
PROJECT = HERE.parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cache", type=Path, default=PROJECT / "out/issue-repro/cache")
    commands = parser.add_subparsers(dest="command", required=True)
    fetch = commands.add_parser("scan", help="cache the newest approved check for each issue")
    fetch.add_argument("--state", choices=("open", "closed", "all"), default="all")
    fetch.add_argument("--label", action="append", default=[])
    fetch.add_argument("--jobs", type=int, choices=range(1, 17), default=4)
    run = commands.add_parser("run", help="build this checkout and replay cached cases")
    run.add_argument("--show-rows", action="store_true", help="show bounded differences")
    for sub in (fetch, run):
        sub.add_argument("--issue", type=int, action="append", default=[])
    local = commands.add_parser("check-file", help="try a Markdown comment locally")
    local.add_argument("file", type=Path)
    args = parser.parse_args()
    config = tomllib.loads((PROJECT / ".issue-repro.toml").read_text())
    if args.command == "check-file":
        with args.file.open("rb") as file:
            body = file.read(65537)
        if len(body) > 65536:
            parser.error("comment exceeds 64 KiB")
        case = parse(body.decode())
        if case is None:
            parser.error("no marked render check")
        status, detail = check(build(config, PROJECT), case, show_rows=True)
        print(status, detail)
        return int(status != "pass")

    args.cache.mkdir(parents=True, exist_ok=True)
    snapshot = args.cache / "cases.json"
    # A saved case cannot carry approval across a change to the allowlist.
    policy = digest([config[k] for k in ("repository", "maintainer_ids", "bot_ids")])
    if args.command == "scan":
        entries = scan(config["repository"], args.state, args.label, set(args.issue),
                       args.jobs, config["maintainer_ids"], config["bot_ids"])
        with tempfile.NamedTemporaryFile(mode="w", dir=args.cache, delete=False) as file:
            json.dump({"policy": policy, "entries": entries}, file)
        os.replace(file.name, snapshot)
        print(f"Cached {len(entries)} marked issues in {snapshot}")
        return 0

    if not snapshot.exists():
        parser.error("no cached selection; run scan first")
    data = json.loads(snapshot.read_bytes())
    if data["policy"] != policy:
        parser.error("approval configuration changed; run scan again")
    entries = [e for e in data["entries"] if not args.issue or e["number"] in args.issue]
    missing = set(args.issue) - {e["number"] for e in entries}
    if missing:
        parser.error(f"no cached reproducer for issues {sorted(missing)}")
    if not entries:
        publish("No marked reproducer matched this scan.")
        return 0
    return replay(config, args, entries)


def replay(config, args, entries):
    image = build(config, PROJECT) if any("case" in e for e in entries) else None
    revision = git("rev-parse", "HEAD")
    branch = git("rev-parse", "--abbrev-ref", "HEAD")
    if git("status", "--porcelain", "--untracked-files=no"):
        revision += "-dirty"
    runner = digest([config["build"], *[(HERE / name).read_text() for name in
                                       ("Dockerfile", "cases.py", "sandbox.py")]])
    rows = ["| Issue | Result | Evidence |", "| --- | --- | --- |"]
    changes = []
    unsuccessful = False
    with sqlite3.connect(args.cache / "history.db") as history:
        history.execute("CREATE TABLE IF NOT EXISTS results "
                        "(branch,issue,digest,status,revision, PRIMARY KEY(branch,issue,digest))")
        for entry in entries:
            n = entry["number"]
            if "invalid" in entry:
                status, detail = "invalid", entry["invalid"]
            else:
                if args.show_rows:
                    print(f"Checking #{n}")
                case = entry["case"]
                status, detail = check(image, case, args.show_rows)
                if status in ("pass", "fail"):
                    key = (branch, n, digest([config["repository"], runner, case]))
                    previous = history.execute("SELECT status,revision FROM results WHERE "
                                               "branch=? AND issue=? AND digest=?", key).fetchone()
                    history.execute("INSERT OR REPLACE INTO results VALUES (?,?,?,?,?)",
                                    (*key, status, revision))
                    if previous and status != previous[0]:
                        change = "fix candidate" if status == "pass" else "regression candidate"
                        changes.append(f"#{n} {change}")
                        detail += f"; {change} since {previous[1]}"
                        unsuccessful |= status == "fail"
            link = f"https://github.com/{config['repository']}/issues/{n}#issuecomment-{entry['comment']}"
            rows.append(f"| [#{n}]({link}) | {status} | {detail} |")
            unsuccessful |= (status in ("invalid", "inconclusive") or
                             status == "fail" and entry["state"] == "CLOSED")
    publish("\n".join([f"godjot revision `{revision}`", "", *changes, "", *rows]))
    return int(unsuccessful)


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True).encode()).hexdigest()


def git(*args):
    return subprocess.check_output(["git", *args], cwd=PROJECT, text=True).strip()


def publish(text):
    print(text)
    if summary := os.getenv("GITHUB_STEP_SUMMARY"):
        with open(summary, "a") as file:
            print(text, file=file)


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError, subprocess.SubprocessError) as exc:
        raise SystemExit(str(exc)) from None
