"""Build trusted executables and run individual cases in fresh containers."""
import json
from pathlib import Path
import re
import resource
import shutil
import subprocess
import tempfile
import uuid

HERE = Path(__file__).resolve().parent
MAX_CAPTURE = 131072
REFERENCE_FILES = ("Dockerfile", "package.json", "package-lock.json", "render.cjs")
RACE_TESTS = {"parser-state": "TestParserState", "parallel-render": "TestParallelRender"}


def build(config, project, *, race=False):
    project = Path(project).resolve()
    subprocess.run(config["race_build" if race else "build"], cwd=project, check=True)
    name = "concurrency.test" if race else "godjot"
    binary = (project / "out/issue-repro" / name).resolve(strict=True)
    if not binary.is_relative_to(project) or not binary.is_file():
        raise ValueError("expected a regular executable inside the checkout")
    with tempfile.TemporaryDirectory() as tmp:
        context = Path(tmp)
        shutil.copy(HERE / ("Dockerfile.race" if race else "Dockerfile"), context / "Dockerfile")
        shutil.copy(binary, context / name)
        if race:
            copy_race_libraries(binary, context / "runtime")
        image_id = context / "image-id"
        subprocess.run(["docker", "build", "--network=none", "--iidfile", image_id, context],
                       check=True)
        return image_id.read_text().strip()


def copy_race_libraries(binary, destination):
    # cgo uses the host loader and libc. Preserve their runtime paths while
    # limiting the Docker build context to native system libraries.
    libraries = subprocess.check_output(["ldd", str(binary)], env={"PATH": "/usr/bin:/bin", "LC_ALL": "C"},
                                        text=True, timeout=10)
    for line in libraries.splitlines():
        if line.lstrip().startswith(("linux-vdso", "linux-gate")):
            continue
        match = re.fullmatch(r"\s*(?:\S+ => )?(/\S+) \(0x[0-9a-fA-F]+\)\s*", line)
        if not match:
            raise ValueError("unresolved race runtime library")
        path = Path(match[1])
        source = path.resolve(strict=True)
        roots = (Path("/lib"), Path("/lib64"), Path("/usr/lib"), Path("/usr/lib64"))
        if (".." in path.parts or not any(path.is_relative_to(root) for root in roots)
                or not any(source.is_relative_to(root.resolve()) for root in roots)
                or not source.is_file()):
            raise ValueError("expected a race runtime library in a system library directory")
        target = destination / path.relative_to("/")
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy(source, target)


def build_reference(project):
    """Only the pinned upstream package is downloaded, at image build time."""
    reference = Path(project) / "scripts/issue-repro/reference"
    with tempfile.TemporaryDirectory() as tmp:
        context = Path(tmp)
        for name in REFERENCE_FILES:
            shutil.copy(reference / name, context / name)
        image_id = context / "image-id"
        subprocess.run(["docker", "build", "--iidfile", image_id, context], check=True)
        return image_id.read_text().strip()


def check(image, case, show_rows=False, reference=None):
    """Run the fixed race suite or compare exact renderer output on the host."""
    if "suite" in case:
        test = RACE_TESTS.get(case["suite"])
        if test is None:
            return "invalid", "unknown concurrency suite"
        try:
            status, detail, _ = run_renderer(image, b"", test=test)
        except (OSError, subprocess.SubprocessError):
            return "inconclusive", "race container unavailable"
        return status, detail
    stdin = case["stdin"].encode()
    if case.get("reference"):
        if reference is None:
            return "inconclusive", "reference renderer unavailable"
        try:
            status, detail, expected = run_renderer(reference, stdin)
        except (OSError, subprocess.SubprocessError):
            return "inconclusive", "reference container unavailable"
        if status != "ok":
            return "inconclusive", "reference: " + detail
    else:
        expected = case["expected"].encode()
    try:
        status, detail, actual = run_renderer(image, stdin)
    except (OSError, subprocess.SubprocessError):
        return "inconclusive", "renderer container unavailable"
    if status != "ok":
        return status, detail
    if actual == expected:
        return "pass", "stdout matched"
    if show_rows:
        print(f"  expected: {json.dumps(expected.decode(errors='backslashreplace'))[:500]}")
        print(f"  actual: {json.dumps(actual.decode(errors='backslashreplace'))[:500]}")
    return "fail", "expected and actual stdout differ"


def run_renderer(image, stdin, *, test=None):
    name = "godjot-repro-" + uuid.uuid4().hex
    memory = "512m" if test else "128m"
    command = [
        "docker", "run", "--name", name, "--pull=never", "--network=none",
        "--user=65534:65534", "--read-only", "--cap-drop=ALL",
        "--security-opt=no-new-privileges", "--pids-limit=64", "--memory", memory,
        "--memory-swap", memory, "--cpus", "2" if test else "1", "--ulimit", "core=0:0", "--ulimit",
        "nofile=64:64", "--log-driver=none", "-i", image,
    ]
    if test:
        command.extend(["-test.run", f"^{test}$", "-test.v", "-test.timeout=25s"])
    with tempfile.TemporaryFile() as out, tempfile.TemporaryFile() as err:
        try:
            result = subprocess.run(command, input=stdin, stdout=out, stderr=err,
                                    timeout=35 if test else 20, preexec_fn=limit_capture)
            if out.tell() >= MAX_CAPTURE or err.tell() >= MAX_CAPTURE:
                return "inconclusive", "output limit reached", b""
            code = result.returncode
            if code < 0 or code in (125, 126, 127, 137):
                return "inconclusive", f"container exited with status {code}", b""
            out.seek(0)
            stdout = out.read()
            if test:
                err.seek(0)
                status, detail = race_result(test, code, stdout, err.read())
                return status, detail, b""
        except subprocess.TimeoutExpired:
            return "inconclusive", "container timed out", b""
        finally:
            subprocess.run(["docker", "rm", "-f", name], stdout=subprocess.DEVNULL,
                           stderr=subprocess.DEVNULL, timeout=30)
    if code:
        return "fail", f"renderer exited with status {code}", b""
    return "ok", "", stdout


def race_result(test, code, stdout, stderr):
    name = test.encode()
    lines = stdout.splitlines()
    if b"=== RUN   " + name not in lines:
        return "inconclusive", "named test did not run"
    diagnostics = stdout + b"\n" + stderr
    if b"panic: test timed out" in diagnostics or b"FATAL: ThreadSanitizer" in diagnostics:
        return "inconclusive", "race runtime unavailable or test timed out"
    if b"WARNING: DATA RACE" in diagnostics.splitlines():
        return "fail", "Go race detector reported a data race"
    if code and any(line.startswith(b"--- FAIL: " + name + b" (") for line in lines):
        return "fail", f"{test} failed"
    if not code and b"PASS" in lines and any(line.startswith(b"--- PASS: " + name + b" (") for line in lines):
        return "pass", f"{test} passed"
    return "inconclusive", "test did not report a conclusive result"


def limit_capture():
    # Bound the host-side Docker client's stdout and stderr independently of
    # the container's own memory and CPU quotas.
    resource.setrlimit(resource.RLIMIT_FSIZE, (MAX_CAPTURE, MAX_CAPTURE))
