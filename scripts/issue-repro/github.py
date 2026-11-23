"""Collect the latest eligible reproducer from each selected GitHub issue."""

from concurrent.futures import ThreadPoolExecutor
import json
import re
import subprocess

from cases import Invalid, parse


COMMENTS = """
    pageInfo { hasPreviousPage startCursor }
    nodes { databaseId body lastEditedAt author { __typename ... on Bot { id } ... on User { databaseId } } }
"""
ISSUE = "number state comments(last:8) {" + COMMENTS + "}"


def api(method, endpoint, payload=None):
    argv = ["gh", "api", "--method", method, endpoint]
    data = None
    if payload is not None:
        argv.extend(["--input", "-"])
        data = json.dumps(payload).encode()
    response = subprocess.check_output(argv, input=data, timeout=45)
    if method == "GET":
        return json.loads(response)


def scan(repo, state, labels, numbers, jobs, maintainer_ids, bot_ids):
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", repo):
        raise ValueError("expected OWNER/REPO")
    owner, name = repo.split("/")
    states = ["OPEN", "CLOSED"] if state == "all" else [state.upper()]

    def query(declarations, selection, **variables):
        document = ("query($owner:String!,$name:String!" + declarations +
                    "){repository(owner:$owner,name:$name){" + selection + "}}")
        payload = {"query": document, "variables": dict(owner=owner, name=name, **variables)}
        response = json.loads(subprocess.check_output(
            ["gh", "api", "graphql", "--input", "-"], input=json.dumps(payload).encode(), timeout=90))
        if response.get("errors") or not response["data"]["repository"]:
            raise ValueError("GitHub returned an incomplete issue listing")
        return response["data"]["repository"]

    def collect(issue):
        comments = issue["comments"]
        while True:
            for comment in reversed(comments["nodes"]):
                if not trusted(comment["author"], maintainer_ids, bot_ids):
                    continue
                if "issue-repro v1" not in comment["body"]:
                    continue
                try:
                    case = parse(comment["body"])
                    if case is None:
                        continue
                    entry = {"case": case}
                except Invalid as exc:
                    entry = {"invalid": str(exc)}
                if comment["lastEditedAt"]:
                    entry = {"invalid": "edited reproducer; publish a fresh comment"}
                return dict(entry, number=issue["number"], state=issue["state"], comment=comment["databaseId"])
            page = comments["pageInfo"]
            if not page["hasPreviousPage"]:
                return None
            comments = query(",$n:Int!,$cursor:String!",
                             "issue(number:$n){comments(last:50,before:$cursor){" + COMMENTS + "}}",
                             n=issue["number"], cursor=page["startCursor"])["issue"]["comments"]

    def by_number(n):
        issue = query(",$n:Int!", "issue(number:$n){" + ISSUE + "}", n=n)["issue"]
        if issue is None:
            raise ValueError(f"issue #{n} does not exist")
        return collect(issue) if issue["state"] in states else None

    with ThreadPoolExecutor(max_workers=jobs) as pool:
        if numbers and not labels:
            return list(filter(None, pool.map(by_number, sorted(numbers))))
        entries = []
        cursor = None
        while True:
            page = query(",$states:[IssueState!],$labels:[String!],$cursor:String",
                         "issues(first:40,after:$cursor,states:$states,labels:$labels,"
                         "orderBy:{field:CREATED_AT,direction:ASC}){pageInfo{hasNextPage endCursor}"
                         " nodes{" + ISSUE + "}}",
                         states=states, labels=labels or None, cursor=cursor)["issues"]
            issues = [issue for issue in page["nodes"] if not numbers or issue["number"] in numbers]
            entries.extend(filter(None, pool.map(collect, issues)))
            if not page["pageInfo"]["hasNextPage"]:
                return entries
            cursor = page["pageInfo"]["endCursor"]


def trusted(author, maintainer_ids, bot_ids):
    if author is None:
        return False
    if author["__typename"] == "User":
        user_id = author.get("databaseId")
        return type(user_id) is int and user_id in maintainer_ids
    if author["__typename"] == "Bot":
        node_id = author.get("id")
        return isinstance(node_id, str) and node_id in bot_ids
    return False
