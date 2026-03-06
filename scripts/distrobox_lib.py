"""Shared library for distrobox setup and integration testing."""

from __future__ import annotations

import configparser
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

from rich.console import Console

console = Console()


def normalize_path(p: str) -> str:
    """Normalize /var/home → /home (containers don't have the atomic symlink)."""
    return re.sub(r"^/var/home/", "/home/", p)


def repo_dir() -> Path:
    """Return the repository root directory (normalized)."""
    raw = Path(__file__).resolve().parent.parent
    return Path(normalize_path(str(raw)))


def distrobox_ini_path() -> Path:
    """Return the path to containers/distrobox.ini."""
    return repo_dir() / "containers" / "distrobox.ini"


def parse_distrobox_ini(path: Path) -> dict[str, dict[str, str]]:
    """Parse distrobox.ini and return a dict of container configs.

    Returns: {"container-name": {"image": ..., "home": ..., "additional_packages": ...}}
    """
    parser = configparser.ConfigParser()
    parser.read(path)
    result: dict[str, dict[str, str]] = {}
    for section in parser.sections():
        result[section] = dict(parser[section])
    return result


def check_not_in_distrobox() -> None:
    """Exit if running inside a distrobox container."""
    if os.environ.get("DISTROBOX_ENTER_PATH"):
        console.print("[red]Error:[/] Cannot run inside a distrobox container.")
        sys.exit(1)


def check_not_wsl() -> None:
    """Exit if running on WSL."""
    try:
        version = Path("/proc/version").read_text()
        if "microsoft" in version.lower():
            console.print("[red]Error:[/] This script is for Aurora/Bluefin hosts, not WSL.")
            console.print("On WSL, chezmoi manages dotfiles directly without Distrobox.")
            sys.exit(1)
    except FileNotFoundError:
        pass


def check_distrobox_available() -> None:
    """Exit if distrobox is not installed."""
    if not _command_exists("distrobox"):
        console.print("[red]Error:[/] distrobox not found. Is this an Aurora/Bluefin system?")
        sys.exit(1)


def _command_exists(cmd: str) -> bool:
    """Check if a command exists on PATH."""
    return shutil.which(cmd) is not None


def run_in_container(name: str, cmd: str) -> subprocess.CompletedProcess[str]:
    """Run a shell command inside a distrobox container."""
    return subprocess.run(
        ["distrobox", "enter", name, "--", "sh", "-c", cmd],
        capture_output=True,
        text=True,
    )


def container_stop(name: str) -> None:
    """Stop a distrobox container (ignore errors)."""
    subprocess.run(
        ["distrobox", "stop", "-Y", name],
        capture_output=True,
    )


def container_rm(name: str) -> None:
    """Remove a distrobox container (ignore errors)."""
    subprocess.run(
        ["distrobox", "rm", "--force", name],
        capture_output=True,
    )


def container_create(name: str, image: str, home: str, packages: str) -> None:
    """Create a single distrobox container."""
    # Expand ~ in home path
    home = os.path.expanduser(home)
    cmd = [
        "distrobox", "create",
        "--name", name,
        "--image", image,
        "--home", home,
        "--additional-packages", packages,
        "--yes",
    ]
    subprocess.run(cmd, check=True)


def container_assemble(ini_path: Path) -> None:
    """Create containers from a distrobox ini file."""
    console.print(f"Creating Distrobox containers from {ini_path}...")
    subprocess.run(
        ["distrobox", "assemble", "create", "--file", str(ini_path)],
        check=True,
    )


def partial_config(context: str) -> str:
    """Generate a 2-line chezmoi TOML config (interactive — prompts for remaining values)."""
    return f"""\
[data]
  platform = "distrobox"
  context = "{context}"
"""


ATUIN_SYNC_ADDRESS = "https://atuin.k8s.rommelporras.com"


def full_config_for(context: str) -> str:
    """Generate a full chezmoi TOML config (non-interactive — all values pre-filled)."""
    atuin_account = "none"
    atuin_sync_address = ""
    has_homelab_creds = "false"
    has_work_creds = "false"

    if context == "personal":
        atuin_account = "personal"
        atuin_sync_address = ATUIN_SYNC_ADDRESS
        has_homelab_creds = "true"
    elif context.startswith("personal-"):
        atuin_account = "personal"
        atuin_sync_address = ATUIN_SYNC_ADDRESS
    elif context.startswith("work-"):
        atuin_account = context  # work-eam, work-<name>, etc.
        atuin_sync_address = ATUIN_SYNC_ADDRESS
        has_work_creds = "true"

    # has_op_cli is derived: true for personal-<project> on distrobox
    has_op_cli = "true" if context.startswith("personal-") else "false"

    return f"""\
[data]
  platform = "distrobox"
  context = "{context}"
  personal_email = "test@example.com"
  work_email = "test@company.com"
  has_work_creds = {has_work_creds}
  has_homelab_creds = {has_homelab_creds}
  has_op_cli = {has_op_cli}
  atuin_sync_address = "{atuin_sync_address}"
  atuin_account = "{atuin_account}"
"""


def bootstrap_chezmoi(
    name: str,
    repo: str | Path,
    config: str,
    *,
    clear_state: bool = False,
    timeout: int = 900,
) -> int:
    """Install chezmoi, link source to host repo, configure, and apply inside a container."""
    clear_cmd = ""
    if clear_state:
        clear_cmd = '"$HOME/bin/chezmoi" state delete-bucket --bucket=scriptState || true'

    script = f"""\
cd "$HOME"

# Install chezmoi
if [ ! -x "$HOME/bin/chezmoi" ]; then
    echo 'Installing chezmoi...'
    curl -fsLS get.chezmoi.io | sh
fi

# Link chezmoi source to host repo (live edits without cloning)
mkdir -p "$HOME/.local/share"
rm -rf "$HOME/.local/share/chezmoi"
ln -s '{repo}' "$HOME/.local/share/chezmoi"
echo 'Linked chezmoi source → {repo}'

# Pre-seed config
mkdir -p "$HOME/.config/chezmoi"
cat > "$HOME/.config/chezmoi/chezmoi.toml" <<'TOML'
{config}TOML

{clear_cmd}

# Apply
"$HOME/bin/chezmoi" init --apply
"""
    try:
        result = subprocess.run(
            ["distrobox", "enter", name, "--", "sh", "-c", script],
            timeout=timeout,
        )
    except subprocess.TimeoutExpired:
        console.print(
            f"[red]Error:[/] chezmoi bootstrap timed out after {timeout}s"
        )
        return 1
    if result.returncode != 0:
        console.print(
            f"[yellow]Warning:[/] chezmoi bootstrap exited with code {result.returncode}"
        )
    return result.returncode


def run_setup_creds(container: str) -> int:
    """Run setup-creds inside a container. Returns exit code."""
    console.print(
        f"Seeding credentials for [bold]{container}[/] "
        "(requires 1Password unlock on host)..."
    )
    result = subprocess.run(
        ["distrobox", "enter", container, "--", "sh", "-c",
         '"$HOME/.local/bin/setup-creds"'],
    )
    if result.returncode != 0:
        console.print(
            f"[yellow]Warning:[/] setup-creds exited with code {result.returncode}"
        )
    return result.returncode


VALID_CONTAINER_RE = re.compile(r"^(work-\w+|personal(-\w+)?|gaming|sandbox)$")


def validate_container_name(name: str) -> bool:
    """Check if a container name matches the allowed pattern."""
    return bool(VALID_CONTAINER_RE.match(name))
