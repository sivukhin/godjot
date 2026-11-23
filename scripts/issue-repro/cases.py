"""Read one predefined check from a marked GitHub comment."""
import re
import shlex


CONCURRENCY_SUITES = ("parser-state", "parallel-render")


class Invalid(ValueError):
    pass


def parse(body):
    """Return a check, or None when no fence opts in."""
    case = None
    fence = None
    for line in body.replace("\r\n", "\n").split("\n"):
        if fence is None:
            opener = re.fullmatch(r" {0,3}(`{3,}|~{3,})(.*)", line)
            if opener:
                fence, language = opener.groups()
                lines = []
        elif re.fullmatch(r" {0,3}" + fence[0] + "{" + str(len(fence)) + r",}[ \t]*", line):
            if marked(language, lines):
                if case is not None:
                    raise Invalid("expected only one marked fence per comment")
                case = parse_command("\n".join(lines[1:]))
            fence = None
        else:
            lines.append(line)
    if fence and marked(language, lines):
        raise Invalid("unterminated reproducer fence")
    return case


def marked(language, lines):
    return language.strip() == "bash" and lines and lines[0] == "# issue-repro v1"


def parse_command(source):
    if len(source.encode()) > 49152:
        raise Invalid("reproducer exceeds 48 KiB")
    try:
        # Quotes group literal arguments, including newlines in rendering cases.
        args = shlex.split(source)
    except ValueError as exc:
        raise Invalid("unclosed quote or escape") from exc
    if args[:1] == ["check-concurrency"]:
        if len(args) != 2 or args[1] not in CONCURRENCY_SUITES:
            raise Invalid("expected: check-concurrency parser-state or check-concurrency parallel-render")
        return {"suite": args[1]}
    if len(args) == 4 and args[:2] == ["expect-tool", "render"]:
        case = {"stdin": args[2] + "\n", "expected": args[3] + "\n"}
    elif len(args) == 3 and args[:2] == ["diff-tool", "render"]:
        case = {"stdin": args[2] + "\n", "reference": True}
    else:
        raise Invalid("expected: expect-tool render 'input' 'expected output' or diff-tool render 'input'")
    if any(len(value.encode()) > 16384 or "\0" in value for value in args[2:]):
        raise Invalid("each argument must contain at most 16 KiB of text without NUL")
    return case
