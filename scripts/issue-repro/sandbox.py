"""Build the trusted renderer, then check its output in a fresh container."""
import json
from pathlib import Path
import resource
import shutil
import subprocess
import tempfile
import uuid

HERE = Path(__file__).resolve().parent
MAX_CAPTURE = 131072


def build(config, project):
    project = Path(project).resolve()
    subprocess.run(config["build"], cwd=project, check=True)
    binary = (project / "out/issue-repro/godjot").resolve(strict=True)
    if not binary.is_relative_to(project) or not binary.is_file():
        raise ValueError("expected a renderer binary inside the checkout")
    with tempfile.TemporaryDirectory() as tmp:
        context = Path(tmp)
        shutil.copy(HERE / "Dockerfile", context)
        shutil.copy(binary, context / "godjot")
        image_id = context / "image-id"
        subprocess.run(["docker", "build", "--network=none", "--iidfile", image_id, context],
                       check=True)
        return image_id.read_text().strip()


def check(image, case, show_rows=False):
    """Supply UTF-8 input via stdin, compare exact stdout, and discard the container."""
    name = "godjot-repro-" + uuid.uuid4().hex
    command = [
        "docker", "run", "--name", name, "--pull=never", "--network=none",
        "--user=65534:65534", "--read-only", "--cap-drop=ALL",
        "--security-opt=no-new-privileges", "--pids-limit=32", "--memory=128m",
        "--memory-swap=128m", "--cpus=1", "--ulimit", "core=0:0", "--ulimit",
        "nofile=64:64", "--log-driver=none", "-i", image,
    ]
    with tempfile.TemporaryFile() as out, tempfile.TemporaryFile() as err:
        try:
            result = subprocess.run(command, input=case["stdin"].encode(), stdout=out, stderr=err,
                                    timeout=20, preexec_fn=limit_capture)
            if out.tell() >= MAX_CAPTURE or err.tell() >= MAX_CAPTURE:
                return "inconclusive", "output limit reached"
            code = result.returncode
            if code < 0 or code in (125, 126, 127, 137):
                return "inconclusive", f"container exited with status {code}"
            if code:
                return "fail", f"renderer exited with status {code}"
            out.seek(0)
            stdout = out.read()
        except subprocess.TimeoutExpired:
            return "inconclusive", "renderer timed out"
        finally:
            subprocess.run(["docker", "rm", "-f", name], stdout=subprocess.DEVNULL,
                           stderr=subprocess.DEVNULL, timeout=30)
    if stdout == case["expected"].encode():
        return "pass", "stdout matched"
    if show_rows:
        print(f"  expected: {json.dumps(case['expected'])[:500]}")
        print(f"  actual: {json.dumps(stdout.decode(errors='backslashreplace'))[:500]}")
    return "fail", "expected and actual stdout differ"


def limit_capture():
    # Bound the host-side Docker client's stdout and stderr independently of
    # the container's own memory and CPU quotas.
    resource.setrlimit(resource.RLIMIT_FSIZE, (MAX_CAPTURE, MAX_CAPTURE))
