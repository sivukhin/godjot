"""Read one renderer check from a marked GitHub comment."""
import re
import shlex


class Invalid(ValueError):
    pass


def parse(body):
    """Return stdin and expected stdout, or None when no fence opts in."""
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
                case = render_case("\n".join(lines[1:]))
            fence = None
        else:
            lines.append(line)
    if fence and marked(language, lines):
        raise Invalid("unterminated reproducer fence")
    return case


def marked(language, lines):
    return language.strip() == "bash" and lines and lines[0] == "# issue-repro v1"


def render_case(source):
    if len(source.encode()) > 49152:
        raise Invalid("reproducer exceeds 48 KiB")
    try:
        # Quotes group literal arguments, including newlines. The strings go
        # directly to the comparator and the renderer's stdin.
        args = shlex.split(source)
    except ValueError as exc:
        raise Invalid("unclosed quote or escape") from exc
    if len(args) != 4 or args[:2] != ["expect-tool", "render"]:
        raise Invalid("expected: expect-tool render 'input' 'expected output'")
    stdin, expected = args[2:]
    if any(len(value.encode()) > 16384 or "\0" in value for value in (stdin, expected)):
        raise Invalid("each argument must contain at most 16 KiB of text without NUL")
    return {"stdin": stdin + "\n", "expected": expected + "\n"}
