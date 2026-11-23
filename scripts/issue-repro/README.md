# Issue reproducers

Keep a rendering check in the issue that describes the bug. A maintainer
posts a marked comment, and CI replays it against the default branch. The
report links to the comment when a case starts passing or failing.

## Write a comment

Include one fenced block with the marker on its first line. For example:

````markdown
```bash
# issue-repro v1
expect-tool render '*Hello*' '<p><strong>Hello</strong></p>'
```
````

`render` takes two arguments, the Djot input and expected HTML.
Use shell-style quoting to keep spaces or line breaks inside each argument. The runner adds
a trailing newline to both, sends the input to godjot through standard input,
and compares the output byte for byte. The fence accepts a fixed helper
grammar. It never invokes a shell.

Anyone can propose an example. The scanner accepts a published case from
an allowlisted maintainer (numeric GitHub user ID) or bot (GitHub GraphQL
node ID). The [repository configuration](../../.issue-repro.toml) defines
these publishers and the build command. Publish a new comment to revise a
case. An edited or malformed newest trusted case requires a replacement.

## Run locally

From the repository root on Linux, use Python 3.11+, Go 1.23+, Docker, and
an authenticated GitHub CLI for scanning. Try a Markdown file before posting,
then scan the tracker and run the saved cases:

```sh
scripts/issue-repro/repro.py check-file /tmp/comment.md
scripts/issue-repro/repro.py scan
scripts/issue-repro/repro.py run
```

By default the scan includes open and closed issues. To work with a subset,
filter by state or label, or request issue numbers directly. Use a separate
cache directory to keep multiple selections.

```sh
scripts/issue-repro/repro.py --cache out/issue-repro/open-bugs scan --state open --label bug --jobs 4
scripts/issue-repro/repro.py --cache out/issue-repro/open-bugs run
scripts/issue-repro/repro.py scan --issue 29
scripts/issue-repro/repro.py run --issue 29 --show-rows
```

A successful scan replaces the selected snapshot. Changes to the publisher allowlist require a fresh scan. The default cache directory
is `out/issue-repro/cache`. Switch branches or edit the renderer and use
`run` again to build the current checkout and replay the same cached inputs
without fetching issues. Initial builds can require network access to obtain
dependencies. Build commands run on the host, so use a trusted checkout.

The renderer runs as a non-root user in a fresh container with a read-only
root, bounded temporary storage, no host mounts or network, and limits on
CPU, memory, processes, output, and elapsed time. Only the scanner receives
the GitHub credential in CI.

## Read the report

A failure followed by a pass on the same case is a fix candidate. A pass
followed by a failure is a regression candidate. A first pass establishes a
baseline. An inconclusive check preserves the previous conclusive result.
The report includes the tested revision so a maintainer can review the
behavior and decide whether to close the issue.

The [nightly workflow](../../.github/workflows/issue-repros.yml) also
supports manual runs on the default branch. It writes a job summary and
caches run history between invocations. Cache eviction starts a fresh
baseline. Same-repository pull requests that change the renderer or tooling also replay
the approved cases and report current results. Only default-branch runs update
the shared nightly history. Fork pull requests run the offline tests.
