# Issue reproducers

Keep a rendering check in the issue that describes the bug. A maintainer
posts a marked comment, and CI replays it against the default branch. The
report links to the comment when a case starts passing or failing.

## Write a comment

Include one fenced block with the marker on its first line. Use an explicit
expectation when you know the HTML godjot should produce.

````markdown
```bash
# issue-repro v1
expect-tool render '*Hello*' '<p><strong>Hello</strong></p>'
```
````

`expect-tool render` takes the Djot input and the expected HTML. To compare
godjot with an independent renderer, use `diff-tool render` and supply only
the input. This example covers [the raw-inline rendering bug in #29](https://github.com/sivukhin/godjot/issues/29).

````markdown
```bash
# issue-repro v1
diff-tool render 'a`hello`{=latex}b'
```
````

The reference HTML is `<p>ab</p>` followed by a newline. The bug report shows
godjot displaying `hello` inside a code span. A differential check derives
the expected HTML from the pinned reference, so the comment can contain just
the source document, including for examples with long HTML output.

Both helpers accept shell-style quoting for spaces and line breaks, with a
16 KiB UTF-8 limit per argument. The runner appends a newline to each input
and to an explicit expectation. The comment supplies only literal data.

The reference is `@djot/djot` 0.3.2, the [TypeScript implementation](https://github.com/jgm/djot.js),
pinned in the [npm lockfile](reference/package-lock.json). Djot also has an
[original Lua implementation](https://github.com/jgm/djot.lua). Using one
fixed reference makes results comparable across runs. CI renders identical
input in separate containers and compares stdout byte for byte. Differences
in indentation, entity spelling, or a trailing newline can produce different
bytes even when the HTML has the same meaning. Use an explicit expectation
when godjot intentionally formats HTML differently. A reference error or
timeout leaves the result inconclusive.

Anyone can propose an example. The scanner accepts a published case from
an allowlisted maintainer (numeric GitHub user ID) or bot (GitHub GraphQL
node ID). The [repository configuration](../../.issue-repro.toml) defines
these publishers and the build command. Publish a new comment to revise a
case. An edited or malformed newest trusted case requires a replacement.

## Check concurrent use of the library

An issue can select a reviewed Go test by name:

````markdown
```bash
# issue-repro v1
check-concurrency parallel-render
```
````

`parallel-render` shares a renderer across goroutines, each with its own parsed
document and output writer. It checks expected HTML under the Go race detector.
A passing run covers only the interleavings exercised by that run.

`parser-state` checks the shared-state symptom reported in
[issue #27](https://github.com/sivukhin/godjot/issues/27). It observes whether
parsing one document changes the package's exported start-symbol set in a fresh
process. This is a deterministic diagnostic for mutation. It does not establish
whether a particular execution involved a data race.

These checks use a cgo-enabled Linux Go toolchain with race detector support
and a C compiler. The runner builds a dedicated test executable on demand;
the repository chooses the test, flags and fixtures.

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

After a completed run on the default branch, a separate publishing job creates
or updates one **Issue reproducer dashboard** issue. It shows counts for unchanged
cases, new observations, fixes, regressions, and inconclusive results. The
report links to the changed cases and to the Actions run. An empty scan
reports zero cases. A scan or build interrupted before the report is available
leaves the previous dashboard in place. Pull requests publish only a job
summary. The reporting job is the only job with write access to issues.
