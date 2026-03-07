#!/usr/bin/env python3
"""Integration test for distrobox + chezmoi bootstrap.

Full lifecycle: delete → create → bootstrap → verify → delete.
"""

from __future__ import annotations

import argparse
import signal
import sys
from dataclasses import dataclass, field
from pathlib import Path

from distrobox_lib import (
    bootstrap_chezmoi,
    check_distrobox_available,
    check_not_in_distrobox,
    console,
    container_create,
    container_rm,
    container_stop,
    distrobox_ini_path,
    full_config_for,
    parse_distrobox_ini,
    repo_dir,
    run_in_container,
)

ALL_CONTAINERS = ["sandbox", "personal", "personal-fintrack", "work-eam"]
DEFAULT_CONTAINER = "personal-fintrack"


@dataclass
class TestState:
    passed: int = 0
    failed: int = 0
    errors: list[str] = field(default_factory=list)


class ContainerTest:
    """Assertion helpers for a single container."""

    def __init__(self, name: str, state: TestState) -> None:
        self.name = name
        self.state = state

    def _record(self, desc: str, success: bool) -> None:
        if success:
            console.print(f"  [green]PASS:[/] {desc}")
            self.state.passed += 1
        else:
            console.print(f"  [red]FAIL:[/] {desc}")
            self.state.failed += 1
            self.state.errors.append(f"[{self.name}] {desc}")

    def assert_exec(self, desc: str, cmd: str) -> None:
        result = run_in_container(self.name, cmd)
        self._record(desc, result.returncode == 0)

    def assert_file(self, desc: str, path: str) -> None:
        self.assert_exec(desc, f'test -f "{path}"')

    def assert_no_file(self, desc: str, path: str) -> None:
        self.assert_exec(desc, f'test ! -f "{path}"')

    def assert_dir(self, desc: str, path: str) -> None:
        self.assert_exec(desc, f'test -d "{path}"')

    def assert_no_dir(self, desc: str, path: str) -> None:
        self.assert_exec(desc, f'test ! -d "{path}"')

    def assert_symlink(self, desc: str, path: str) -> None:
        self.assert_exec(desc, f'test -L "{path}"')

    def assert_executable(self, desc: str, path: str) -> None:
        self.assert_exec(desc, f'test -x "{path}"')

    def assert_contains(self, desc: str, filepath: str, pattern: str) -> None:
        self.assert_exec(desc, f"grep -q '{pattern}' \"{filepath}\"")

    def assert_not_contains(self, desc: str, filepath: str, pattern: str) -> None:
        self.assert_exec(desc, f"! grep -q '{pattern}' \"{filepath}\"")

    # ── Per-container verification ──────────────────────────────────────

    def verify_common(self) -> None:
        console.print("  --- Common assertions ---")
        self.assert_file("chezmoi binary exists", "$HOME/bin/chezmoi")
        self.assert_file(".zshrc exists", "$HOME/.zshrc")
        self.assert_exec(".zshrc is not empty", "test -s $HOME/.zshrc")
        self.assert_file(".gitconfig exists", "$HOME/.gitconfig")
        self.assert_symlink("chezmoi source is symlink", "$HOME/.local/share/chezmoi")
        self.assert_exec(
            "NVM installed",
            "test -f $HOME/.nvm/nvm.sh || test -f $HOME/.config/nvm/nvm.sh",
        )
        self.assert_exec(
            "NVM functional",
            'export NVM_DIR=${NVM_DIR:-$HOME/.config/nvm} && . $NVM_DIR/nvm.sh && nvm --version',
        )
        self.assert_exec("fzf installed", "test -d $HOME/.fzf")

    def verify_personal(self) -> None:
        console.print("  --- personal-specific assertions ---")
        self.assert_executable("setup-creds is executable", "$HOME/.local/bin/setup-creds")
        self.assert_exec("glab installed", "command -v glab")
        self.assert_exec("kubectl installed", "command -v kubectl")
        self.assert_file("atuin installed", "$HOME/.atuin/bin/atuin")
        self.assert_file("atuin config exists", "$HOME/.config/atuin/config.toml")
        self.assert_contains("atuin sync_address set", "$HOME/.config/atuin/config.toml", "atuin.k8s.rommelporras.com")
        self.assert_contains("SSH_AUTH_SOCK = 1Password", "$HOME/.zshrc", "1password/agent.sock")
        self.assert_contains("invoicetron alias", "$HOME/.zshrc", "invoicetron")
        self.assert_contains("kubectl-homelab alias", "$HOME/.zshrc", "kubectl-homelab")
        self.assert_contains("OTEL env vars", "$HOME/.zshrc", "OTEL_METRICS_EXPORTER")
        self.assert_exec("ansible installed", "command -v ansible")
        self.assert_not_contains("no OP_BIOMETRIC in .zshrc", "$HOME/.zshrc", "OP_BIOMETRIC_UNLOCK_ENABLED")

    def verify_personal_fintrack(self) -> None:
        console.print("  --- personal-fintrack-specific assertions ---")
        self.assert_executable("setup-creds is executable", "$HOME/.local/bin/setup-creds")
        self.assert_exec("glab installed", "command -v glab")
        self.assert_file("atuin installed", "$HOME/.atuin/bin/atuin")
        self.assert_file("atuin config exists", "$HOME/.config/atuin/config.toml")
        self.assert_contains("atuin sync_address set", "$HOME/.config/atuin/config.toml", "atuin.k8s.rommelporras.com")
        self.assert_not_contains("no 1Password SSH socket", "$HOME/.zshrc", "1password/agent.sock")
        self.assert_no_dir(".kube/ does not exist", "$HOME/.kube")
        self.assert_exec("op CLI installed", "command -v op")
        self.assert_exec("bun installed", "$HOME/.bun/bin/bun --version")
        self.assert_contains("OP_BIOMETRIC in .zshrc", "$HOME/.zshrc", "OP_BIOMETRIC_UNLOCK_ENABLED")
        self.assert_contains("OTEL env vars", "$HOME/.zshrc", "OTEL_METRICS_EXPORTER")
        self.assert_not_contains("no invoicetron alias", "$HOME/.zshrc", "invoicetron")
        self.assert_not_contains("no kubectl-homelab alias", "$HOME/.zshrc", "kubectl-homelab")
        self.assert_contains("BUN_INSTALL in .zshrc", "$HOME/.zshrc", "BUN_INSTALL")

    def verify_work_eam(self) -> None:
        console.print("  --- work-eam-specific assertions ---")
        self.assert_executable("setup-creds is executable", "$HOME/.local/bin/setup-creds")
        self.assert_file("atuin installed", "$HOME/.atuin/bin/atuin")
        self.assert_file("atuin config exists", "$HOME/.config/atuin/config.toml")
        self.assert_contains("atuin sync_address set", "$HOME/.config/atuin/config.toml", "atuin.k8s.rommelporras.com")
        self.assert_contains("SSH_AUTH_SOCK = 1Password", "$HOME/.zshrc", "1password/agent.sock")
        self.assert_contains("terraform alias tfi", "$HOME/.zshrc", "tfi")
        self.assert_contains("EAM alias", "$HOME/.zshrc", "eam-sre")
        self.assert_exec("aws CLI installed", "command -v aws")
        self.assert_exec("kubectl installed", "command -v kubectl")
        self.assert_exec("terraform installed", "command -v terraform")
        self.assert_not_contains("no invoicetron alias", "$HOME/.zshrc", "invoicetron")
        self.assert_not_contains("no OP_BIOMETRIC", "$HOME/.zshrc", "OP_BIOMETRIC_UNLOCK_ENABLED")

    def verify_sandbox(self) -> None:
        console.print("  --- sandbox-specific assertions ---")
        self.assert_no_file("no setup-creds", "$HOME/.local/bin/setup-creds")
        self.assert_exec("no atuin binary", "! command -v atuin && test ! -f $HOME/.atuin/bin/atuin")
        self.assert_not_contains("no atuin sync_address", "$HOME/.config/atuin/config.toml", "sync_address")
        self.assert_contains("SSH_AUTH_SOCK unset", "$HOME/.zshrc", "unset SSH_AUTH_SOCK")
        self.assert_not_contains("no 1Password socket", "$HOME/.zshrc", "1password/agent.sock")
        self.assert_not_contains("no OTEL env vars", "$HOME/.zshrc", "OTEL_METRICS_EXPORTER")
        self.assert_not_contains("no invoicetron alias", "$HOME/.zshrc", "invoicetron")


def test_container(
    container: str,
    state: TestState,
    ini_config: dict[str, dict[str, str]],
    repo: str | Path,
    *,
    keep: bool = False,
) -> None:
    """Run the full lifecycle test for a single container."""
    pass_before = state.passed
    fail_before = state.failed

    console.print()
    console.print("=" * 64)
    console.print(f"[bold]Testing: {container}[/]")
    console.print("=" * 64)

    # Phase 1: Clean slate
    console.print()
    console.print(f"Phase 1: Clean slate (removing {container} if exists)...")
    container_stop(container)
    container_rm(container)
    console.print("  Done.")

    # Phase 2: Create container
    console.print()
    console.print("Phase 2: Creating container...")
    if container not in ini_config:
        console.print(f"  [red]ERROR:[/] Container '{container}' not found in distrobox.ini")
        state.failed += 1
        state.errors.append(f"[{container}] Container not defined in distrobox.ini")
        return

    cfg = ini_config[container]
    image = cfg.get("image", "")
    home = cfg.get("home", "")
    packages = cfg.get("additional_packages", "").strip('"')

    try:
        container_create(container, image, home, packages)
        console.print("  Done.")

        # Phase 3: Bootstrap chezmoi (non-interactive)
        console.print()
        console.print("Phase 3: Bootstrapping chezmoi (non-interactive)...")
        config = full_config_for(container)
        rc = bootstrap_chezmoi(container, repo, config, clear_state=True)
        if rc != 0:
            console.print(f"  [yellow]Bootstrap exited {rc} — continuing to verify what was deployed[/]")
        else:
            console.print("  Done.")

        # Phase 4: Verify deployed state
        console.print()
        console.print("Phase 4: Verifying deployed state...")
        ct = ContainerTest(container, state)
        ct.verify_common()

        if container == "personal":
            ct.verify_personal()
        elif container == "personal-fintrack":
            ct.verify_personal_fintrack()
        elif container.startswith("personal-"):
            ct.verify_personal_fintrack()  # fallback for other personal-<project>
        elif container == "work-eam":
            ct.verify_work_eam()
        elif container.startswith("work-"):
            ct.verify_work_eam()  # fallback for other work-<name>
        elif container == "sandbox":
            ct.verify_sandbox()

    except Exception as exc:
        console.print(f"  [red]ERROR:[/] {exc}")
        state.failed += 1
        state.errors.append(f"[{container}] {exc}")

    finally:
        # Phase 5: Teardown
        console.print()
        if keep:
            console.print(f"Phase 5: Skipped (--keep flag). Container '{container}' is still running.")
        else:
            console.print("Phase 5: Teardown...")
            container_stop(container)
            container_rm(container)
            console.print("  Done.")

    p = state.passed - pass_before
    f = state.failed - fail_before
    console.print()
    console.print(f"  Container '{container}': {p} passed, {f} failed")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Integration test for distrobox + chezmoi bootstrap. "
            "Lifecycle: delete → create → bootstrap → verify → delete."
        ),
    )
    parser.add_argument(
        "--all",
        action="store_true",
        help="Test all containers (sandbox, personal, personal-fintrack, work-eam)",
    )
    parser.add_argument(
        "--keep",
        action="store_true",
        help="Keep container after test (for manual inspection)",
    )
    parser.add_argument(
        "container",
        nargs="*",
        default=[],
        help=f"Container(s) to test (default: {DEFAULT_CONTAINER})",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> None:
    args = parse_args(argv)

    # Preflight checks
    check_not_in_distrobox()
    check_distrobox_available()

    ini_path = distrobox_ini_path()
    if not ini_path.exists():
        console.print(f"[red]Error:[/] {ini_path} not found.")
        sys.exit(1)

    # Determine containers
    if args.all:
        containers = ALL_CONTAINERS
    elif args.container:
        containers = args.container
    else:
        containers = [DEFAULT_CONTAINER]

    state = TestState()
    ini_config = parse_distrobox_ini(ini_path)
    repo = repo_dir()
    tested: list[str] = []  # track containers we've touched for cleanup

    # Signal handler for cleanup
    def cleanup(signum: int, frame: object) -> None:
        if not args.keep and tested:
            console.print()
            console.print("[yellow]Interrupted — cleaning up containers...[/]")
            for c in tested:
                container_stop(c)
                container_rm(c)
        sys.exit(1)

    signal.signal(signal.SIGINT, cleanup)
    signal.signal(signal.SIGTERM, cleanup)

    console.print("[bold]=== Distrobox Integration Test ===[/]")
    console.print(f"Containers: {' '.join(containers)}")
    console.print(f"Keep after test: {args.keep}")

    for container in containers:
        tested.append(container)
        test_container(container, state, ini_config, repo, keep=args.keep)

    # Summary
    console.print()
    console.print("=" * 64)
    console.print("[bold]SUMMARY[/]")
    console.print("=" * 64)
    total = state.passed + state.failed
    console.print(f"Total: {total} assertions, {state.passed} passed, {state.failed} failed")

    if state.errors:
        console.print()
        console.print("Failures:")
        for err in state.errors:
            console.print(f"  - {err}")

    console.print()
    if state.failed == 0:
        console.print("[bold green]ALL TESTS PASSED[/]")
    else:
        console.print("[bold red]TESTS FAILED[/]")
        sys.exit(1)


if __name__ == "__main__":
    main()
