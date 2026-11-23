"""Describe changes since each reproducer's previous conclusive run."""


def change_summary(report):
    counts = report["counts"]
    lines = []
    if counts["fixes"] or counts["unchanged"] or counts["regressions"]:
        lines.append(f"**Fixed** {counts['fixes']} · **Unchanged** {counts['unchanged']} · "
                     f"**Regressed** {counts['regressions']}")
    if n := counts["new"]:
        note = f"First result for {n} {'case' if n == 1 else 'cases'}"
        if not lines and "outcomes" in report:
            note += f" ({report['outcomes']['pass']} passed, {report['outcomes']['fail']} failed)"
        lines.append(note + ".")
    if n := counts["inconclusive"]:
        lines.append(f"No verdict for {n} {'case' if n == 1 else 'cases'}.")
    return "\n\n".join(lines) or "No cases to replay."
