#!/usr/bin/env python3
"""Set up a WSL2 instance with chezmoi dotfiles and credentials.

Equivalent of distrobox_setup.py but for WSL — one command to go from a
fresh Ubuntu instance to a fully configured dev environment.

Prerequisites (must be done manually on Windows first):
  1. 1Password desktop app with SSH Agent enabled
  2. npiperelay + socat installed in WSL
  3. 1Password CLI installed in WSL
  4. SSH agent bridge running (~/.1password/agent.sock)

Usage:
  uv run python scripts/wsl_setup.py work-eam --work-email you@company.com
  uv run python scripts/wsl_setup.py personal --personal-email you@example.com
"""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
from pathlib import Path

from rich.console import Console

console = Console()

ATUIN_SYNC_ADDRESS = "https://atuin.k8s.rommelporras.com"
GITHUB_USER = "rommelporras"


def repo_dir() -> Path:
    """Return the repository root directory."""
    return Path(__file__).resolve().parent.parent


def _command_exists(cmd: str) -> bool:
    return shutil.which(cmd) is not None


def _run(cmd: list[str], *, check: bool = True, **kwargs) -> subprocess.CompletedProcess:
    """Run a command, printing it first."""
    console.print(f"  [dim]$ {' '.join(cmd)}[/]")
    return subprocess.run(cmd, check=check, **kwargs)


# ─── Validation ───────────────────────────────────────────────────────


def check_is_wsl() -> None:
    """Exit if not running on WSL."""
    try:
        with open("/proc/version") as f:
            if "microsoft" not in f.read().lower():
                raise FileNotFoundError
    except FileNotFoundError:
        console.print("[red]Error:[/] This script is for WSL only.")
        sys.exit(1)


def check_prerequisites() -> None:
    """Check that Windows-side prerequisites are in place."""
    errors = []

    if not _command_exists("socat"):
        errors.append("socat not installed — run: sudo apt install -y socat")

    if not _command_exists("npiperelay.exe"):
        errors.append(
            "npiperelay.exe not found — see docs/setup/wsl2.md step 1.7"
        )

    if not _command_exists("op.exe"):
        errors.append(
            "1Password CLI not installed on Windows — "
            "run: winget install AgileBits.1Password.CLI"
        )

    sock = Path.home() / ".1password" / "agent.sock"
    if not sock.exists():
        errors.append(
            f"SSH agent socket not found at {sock} — "
            "see docs/setup/wsl2.md step 1.8"
        )

    if errors:
        console.print("[red]Error:[/] Prerequisites missing:")
        for e in errors:
            console.print(f"  - {e}")
        console.print()
        console.print(
            "Complete step 1 of docs/setup/wsl2.md before running this script."
        )
        sys.exit(1)

    # Verify SSH agent has keys
    result = subprocess.run(
        ["ssh-add", "-l"],
        capture_output=True,
        text=True,
        env={**os.environ, "SSH_AUTH_SOCK": str(sock)},
    )
    if result.returncode != 0:
        console.print(
            "[yellow]Warning:[/] SSH agent has no keys — "
            "check 1Password SSH Agent settings"
        )


def validate_context(context: str) -> bool:
    """Validate context name matches expected patterns."""
    import re
    return bool(re.match(r"^(work-\w+|personal(-\w+)?)$", context))


# ─── Config generation ────────────────────────────────────────────────


def full_config(
    context: str,
    personal_email: str,
    work_email: str,
) -> str:
    """Generate a full chezmoi TOML config (non-interactive)."""
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
        atuin_account = context
        atuin_sync_address = ATUIN_SYNC_ADDRESS
        has_work_creds = "true"

    return f"""\
[data]
  platform = "wsl"
  context = "{context}"
  personal_email = "{personal_email}"
  work_email = "{work_email}"
  has_work_creds = {has_work_creds}
  has_homelab_creds = {has_homelab_creds}
  atuin_sync_address = "{atuin_sync_address}"
  atuin_account = "{atuin_account}"
"""


# ─── Setup steps ──────────────────────────────────────────────────────


def clone_repos() -> None:
    """Clone claude-config and dotfiles repos if not present."""
    personal_dir = Path.home() / "personal"
    personal_dir.mkdir(exist_ok=True)

    claude_config = personal_dir / "claude-config"
    if not claude_config.exists():
        console.print("Cloning claude-config...")
        _run([
            "git", "clone",
            f"git@github.com:{GITHUB_USER}/claude-config.git",
            str(claude_config),
        ])
    else:
        console.print(f"claude-config already at {claude_config}")

    dotfiles = personal_dir / "dotfiles"
    if not dotfiles.exists():
        console.print("Cloning dotfiles...")
        _run([
            "git", "clone",
            f"git@github.com:{GITHUB_USER}/dotfiles.git",
            str(dotfiles),
        ])
    else:
        console.print(f"dotfiles already at {dotfiles}")


def install_chezmoi() -> Path:
    """Install chezmoi if not already present. Returns binary path."""
    # Check if already in PATH
    existing = shutil.which("chezmoi")
    if existing:
        console.print(f"chezmoi already installed at {existing}")
        return Path(existing)

    # Check common locations
    for candidate in [
        Path.home() / ".local" / "bin" / "chezmoi",
        Path.home() / "bin" / "chezmoi",
        Path.home() / "personal" / "bin" / "chezmoi",
    ]:
        if candidate.exists() and os.access(candidate, os.X_OK):
            console.print(f"chezmoi found at {candidate}")
            return candidate

    # Install via get.chezmoi.io
    console.print("Installing chezmoi...")
    _run(["sh", "-c", 'curl -fsLS get.chezmoi.io | sh'], check=True)

    # Find where it was installed (usually ~/bin/chezmoi)
    bin_path = Path.home() / "bin" / "chezmoi"
    if bin_path.exists():
        return bin_path

    console.print("[red]Error:[/] chezmoi install succeeded but binary not found")
    sys.exit(1)


def link_chezmoi_source(dotfiles_path: Path) -> None:
    """Symlink chezmoi source to the local dotfiles repo."""
    chezmoi_source = Path.home() / ".local" / "share" / "chezmoi"
    chezmoi_source.parent.mkdir(parents=True, exist_ok=True)

    if chezmoi_source.is_symlink():
        target = chezmoi_source.resolve()
        if target == dotfiles_path.resolve():
            console.print(f"chezmoi source already linked → {dotfiles_path}")
            return
        chezmoi_source.unlink()
    elif chezmoi_source.exists():
        # Remove the separate clone
        console.print("Removing chezmoi's separate clone...")
        shutil.rmtree(chezmoi_source)

    chezmoi_source.symlink_to(dotfiles_path)
    console.print(f"Linked chezmoi source → {dotfiles_path}")


def ensure_chezmoi_in_path(chezmoi_bin: Path) -> None:
    """Symlink chezmoi into ~/.local/bin/ if not already there."""
    local_bin = Path.home() / ".local" / "bin"
    local_bin.mkdir(parents=True, exist_ok=True)
    target = local_bin / "chezmoi"

    if target.exists() or target.is_symlink():
        return

    target.symlink_to(chezmoi_bin)
    console.print(f"Symlinked chezmoi → {target}")


def write_chezmoi_config(config: str) -> None:
    """Write chezmoi config for non-interactive init."""
    config_dir = Path.home() / ".config" / "chezmoi"
    config_dir.mkdir(parents=True, exist_ok=True)
    config_file = config_dir / "chezmoi.toml"

    if config_file.exists():
        console.print(f"chezmoi config already exists at {config_file}")
        console.print("  [dim](delete it to re-run with new values)[/]")
        return

    config_file.write_text(config)
    console.print(f"Wrote chezmoi config → {config_file}")


def run_chezmoi_apply(chezmoi_bin: Path) -> int:
    """Run chezmoi init --apply."""
    console.print()
    console.print("[bold]Running chezmoi init --apply...[/]")

    # Cache sudo for bootstrap
    console.print("Caching sudo (bootstrap needs apt)...")
    sudo_result = subprocess.run(["sudo", "-v"])
    if sudo_result.returncode != 0:
        console.print("[red]Error:[/] sudo authentication failed")
        return 1

    result = subprocess.run(
        [str(chezmoi_bin), "init", "--apply"],
        timeout=900,
    )
    if result.returncode != 0:
        console.print(
            f"[yellow]Warning:[/] chezmoi apply exited with code {result.returncode}"
        )
    return result.returncode


def run_setup_wsl_creds() -> int:
    """Run setup-wsl-creds to install Claude plugins and seed credentials."""
    script = Path.home() / ".local" / "bin" / "setup-wsl-creds"
    if not script.exists():
        console.print(
            "[yellow]Warning:[/] setup-wsl-creds not found — "
            "skipping credential seeding"
        )
        return 1

    console.print()
    console.print(
        "[bold]Running setup-wsl-creds[/] "
        "(requires 1Password unlock on Windows)..."
    )
    result = subprocess.run([str(script)])
    if result.returncode != 0:
        console.print(
            f"[yellow]Warning:[/] setup-wsl-creds exited with code {result.returncode}"
        )
    return result.returncode


# ─── CLI ──────────────────────────────────────────────────────────────


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Set up a WSL2 instance with chezmoi dotfiles and credentials.",
    )
    parser.add_argument(
        "context",
        help=(
            "Context to set up (personal, personal-<project>, "
            "work-eam, work-<name>)"
        ),
    )
    parser.add_argument(
        "--personal-email",
        default=None,
        help="Personal git email (required for personal/personal-* contexts).",
    )
    parser.add_argument(
        "--work-email",
        default=None,
        help="Work git email (required for work-* contexts).",
    )
    parser.add_argument(
        "--skip-creds",
        action="store_true",
        help="Skip credential seeding (setup-wsl-creds).",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    args = parse_args(argv)

    # ─── Preflight ────────────────────────────────────────────────
    check_is_wsl()

    if not validate_context(args.context):
        console.print(
            f"[red]Error:[/] invalid context [bold]{args.context}[/]. "
            "Choose: personal, personal-<project>, work-eam, work-<name>"
        )
        sys.exit(1)

    # Validate required email
    if args.context.startswith("work-") and not args.work_email:
        console.print("[red]Error:[/] --work-email required for work-* contexts")
        sys.exit(1)
    if not args.context.startswith("work-") and not args.personal_email:
        console.print(
            "[red]Error:[/] --personal-email required for personal contexts"
        )
        sys.exit(1)

    check_prerequisites()

    console.print()
    console.print("[bold]=== WSL2 Setup ===[/]")
    console.print(f"Context: [bold]{args.context}[/]")
    console.print()

    # ─── Clone repos ──────────────────────────────────────────────
    console.print("[bold]--- Cloning repositories ---[/]")
    clone_repos()

    dotfiles_path = Path.home() / "personal" / "dotfiles"

    # ─── Install chezmoi ──────────────────────────────────────────
    console.print()
    console.print("[bold]--- Installing chezmoi ---[/]")
    chezmoi_bin = install_chezmoi()
    ensure_chezmoi_in_path(chezmoi_bin)
    link_chezmoi_source(dotfiles_path)

    # ─── Configure and apply ──────────────────────────────────────
    console.print()
    console.print("[bold]--- Configuring chezmoi ---[/]")
    personal = args.personal_email or ""
    work = args.work_email or ""
    config = full_config(args.context, personal, work)
    write_chezmoi_config(config)

    rc = run_chezmoi_apply(chezmoi_bin)
    if rc != 0:
        console.print(
            "[yellow]Bootstrap had warnings — check output above.[/]"
        )

    # ─── Credentials ──────────────────────────────────────────────
    if not args.skip_creds:
        run_setup_wsl_creds()

    # ─── Done ─────────────────────────────────────────────────────
    console.print()
    console.print("[bold]=== Setup complete ===[/]")
    console.print()
    console.print("Remaining manual steps:")
    console.print("  1. [bold]exec zsh[/]                — reload shell")
    console.print("  2. [bold]gh auth login[/]            — GitHub CLI (browser OAuth)")
    console.print()
    console.print(
        "If setup-wsl-creds failed (1Password locked), re-run later:"
    )
    console.print("  [bold]setup-wsl-creds[/]")


if __name__ == "__main__":
    main()
