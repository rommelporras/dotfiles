#!/usr/bin/env python3
"""Integration test for ai-sandbox containers.

Verifies: tools installed, shell experience, security isolation, persistence.
"""

from __future__ import annotations

import argparse
import subprocess
import sys
from dataclasses import dataclass, field

from distrobox_lib import console

TEST_PROJECT = "test-sandbox"
CONTAINER_NAME = f"sandbox-{TEST_PROJECT}"
IMAGE_NAME = "localhost/sandbox-base:latest"


@dataclass
class TestState:
    passed: int = 0
    failed: int = 0
    errors: list[str] = field(default_factory=list)


def run_in_sandbox(cmd: str) -> subprocess.CompletedProcess[str]:
    """Run a command in a temporary sandbox container."""
    return subprocess.run(
        [
            "podman", "run", "--rm",
            "--name", f"{CONTAINER_NAME}-test-{id(cmd) % 10000}",
            "--userns=keep-id",
            IMAGE_NAME,
            cmd,
        ],
        capture_output=True,
        text=True,
        timeout=60,
    )


def run_in_sandbox_with_home(home_dir: str, cmd: str) -> subprocess.CompletedProcess[str]:
    """Run a command in a sandbox container with persistent home."""
    return subprocess.run(
        [
            "podman", "run", "--rm",
            "--name", f"{CONTAINER_NAME}-test-{id(cmd) % 10000}",
            "--volume", f"{home_dir}:/home/developer:z",
            "--userns=keep-id",
            IMAGE_NAME,
            cmd,
        ],
        capture_output=True,
        text=True,
        timeout=60,
    )


class SandboxTest:
    """Assertion helpers for sandbox container tests."""

    def __init__(self, state: TestState) -> None:
        self.state = state

    def _record(self, desc: str, success: bool, detail: str = "") -> None:
        if success:
            console.print(f"  [green]PASS:[/] {desc}")
            self.state.passed += 1
        else:
            msg = f"{desc}: {detail}" if detail else desc
            console.print(f"  [red]FAIL:[/] {msg}")
            self.state.failed += 1
            self.state.errors.append(f"[sandbox] {msg}")

    def assert_cmd(self, desc: str, cmd: str) -> None:
        result = run_in_sandbox(cmd)
        self._record(desc, result.returncode == 0, result.stderr.strip()[:100])

    def assert_cmd_fails(self, desc: str, cmd: str) -> None:
        result = run_in_sandbox(cmd)
        self._record(desc, result.returncode != 0)

    def assert_cmd_output(self, desc: str, cmd: str, expected: str) -> None:
        result = run_in_sandbox(cmd)
        self._record(
            desc,
            expected in result.stdout,
            f"expected '{expected}' in output, got: {result.stdout.strip()[:100]}",
        )

    # ── Tool verification ────────────────────────────────────────────

    def verify_tools(self) -> None:
        console.print("  --- Core tools ---")
        self.assert_cmd("node installed", "node --version")
        self.assert_cmd("npm installed", "npm --version")
        self.assert_cmd("npx installed", "npx --version")
        self.assert_cmd("bun installed", "bun --version")
        self.assert_cmd("python3 installed", "python3 --version")
        self.assert_cmd("uv installed", "uv --version")
        self.assert_cmd("git installed", "git --version")
        self.assert_cmd("zsh installed", "zsh --version")
        self.assert_cmd("sqlite3 installed", "sqlite3 --version")
        self.assert_cmd("psql installed", "psql --version")
        self.assert_cmd("claude installed", "claude --version")
        self.assert_cmd("curl installed", "curl --version")
        console.print("  --- Container tools ---")
        self.assert_cmd("podman installed", "podman --version")
        self.assert_cmd("podman-compose installed", "podman-compose --version")
        self.assert_cmd("host-spawn installed", "host-spawn --version")
        self.assert_cmd("jq installed", "jq --version")

    def verify_shell_tools(self) -> None:
        console.print("  --- Shell experience ---")
        self.assert_cmd("starship installed", "starship --version")
        self.assert_cmd("fzf installed", "fzf --version")
        self.assert_cmd("atuin installed", "atuin --version")
        self.assert_cmd("oh-my-zsh installed", "test -d /home/developer/.oh-my-zsh")
        self.assert_cmd("zsh-autosuggestions installed",
                        "test -d /home/developer/.oh-my-zsh/custom/plugins/zsh-autosuggestions")
        self.assert_cmd(".zshrc exists", "test -f /home/developer/.zshrc")
        self.assert_cmd(".zshrc has starship", "grep -q starship /home/developer/.zshrc")
        self.assert_cmd(".zshrc has fzf", "grep -q fzf /home/developer/.zshrc")
        self.assert_cmd(".zshrc has zsh-interactive-cd",
                        "grep -q zsh-interactive-cd /home/developer/.zshrc")
        self.assert_cmd(".zshrc has atuin", "grep -q atuin /home/developer/.zshrc")
        console.print("  --- Aliases ---")
        self.assert_cmd(".zshrc has docker alias",
                        "grep -q 'alias docker=podman' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has docker-compose alias",
                        "grep -q 'alias docker-compose=podman-compose' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has gitc alias",
                        "grep -q 'alias gitc=' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has gitm alias",
                        "grep -q 'alias gitm=' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has atuin login prompt",
                        "grep -q 'atuin login' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has ai-sandbox alias",
                        "grep -q 'alias ai-sandbox=' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has 1password ssh config",
                        "grep -q '/run/1password/agent.sock' /home/developer/.zshrc")
        console.print("  --- Host integration ---")
        self.assert_cmd("host-open script exists",
                        "test -x /home/developer/.local/bin/host-open")
        self.assert_cmd(".zshrc has BROWSER",
                        "grep -q 'BROWSER.*host-open' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has IDE forwarding",
                        "grep -q '_host_run code' /home/developer/.zshrc")
        self.assert_cmd("atuin sync_address configured",
                        "grep -q 'atuin.k8s.rommelporras.com' /home/developer/.config/atuin/config.toml")
        self.assert_cmd(".zshrc has claude plugin first-run",
                        "grep -q '.plugins-installed' /home/developer/.zshrc")
        self.assert_cmd(".zshrc has context7 mcp first-run",
                        "grep -q '.mcp-configured' /home/developer/.zshrc")
        self.assert_cmd("nerd font installed",
                        "test -d /home/developer/.local/share/fonts")
        self.assert_cmd("playwright browsers installed",
                        "test -d /opt/playwright")

    # ── Security isolation ───────────────────────────────────────────

    def verify_security(self) -> None:
        console.print("  --- Security isolation ---")
        self.assert_cmd_fails("cannot see host home",
                              "test -d /home/0xwsh")
        self.assert_cmd_fails("cannot see host SSH keys",
                              "ls /home/0xwsh/.ssh/")
        self.assert_cmd_fails("cannot see host 1Password socket",
                              "test -S /home/0xwsh/.1password/agent.sock")
        self.assert_cmd_fails("cannot see host dotfiles",
                              "ls /home/0xwsh/personal/dotfiles/")
        self.assert_cmd_fails("cannot see host .distrobox",
                              "ls /home/0xwsh/.distrobox/")
        self.assert_cmd("no SSH_AUTH_SOCK set",
                        "test -z \"${SSH_AUTH_SOCK:-}\"")
        self.assert_cmd("running as non-root",
                        "test \"$(whoami)\" != root")
        self.assert_cmd_output("home is /home/developer", "echo $HOME", "/home/developer")

    # ── Persistence ──────────────────────────────────────────────────

    def verify_persistence(self, home_dir: str) -> None:
        console.print("  --- Persistence ---")
        # Write a file in first container
        result = run_in_sandbox_with_home(
            home_dir,
            "echo 'persistence-test' > /home/developer/test-persist.txt",
        )
        self._record("write file to persistent home", result.returncode == 0)

        # Read it back in a new container
        result = run_in_sandbox_with_home(
            home_dir,
            "cat /home/developer/test-persist.txt",
        )
        self._record(
            "read file from persistent home",
            "persistence-test" in result.stdout,
        )


def main(argv: list[str] | None = None) -> None:
    parser = argparse.ArgumentParser(
        description="Integration test for ai-sandbox containers.",
    )
    parser.add_argument(
        "--skip-build",
        action="store_true",
        help="Skip image build (use existing image)",
    )
    args = parser.parse_args(argv)

    # Preflight: check image exists
    if not args.skip_build:
        console.print("[bold]=== Building sandbox-base image ===[/]")
        script_dir = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True,
        ).stdout.strip()
        result = subprocess.run(
            [
                "podman", "build",
                "-t", IMAGE_NAME,
                "-f", f"{script_dir}/containers/Containerfile.sandbox-base",
                f"{script_dir}/containers",
            ],
            timeout=600,
        )
        if result.returncode != 0:
            console.print("[red]Error:[/] Failed to build sandbox-base image")
            sys.exit(1)
        console.print("  Image built successfully")
    else:
        # Verify image exists
        result = subprocess.run(
            ["podman", "image", "exists", IMAGE_NAME],
            capture_output=True,
        )
        if result.returncode != 0:
            console.print(f"[red]Error:[/] Image {IMAGE_NAME} not found. Run without --skip-build.")
            sys.exit(1)

    state = TestState()
    st = SandboxTest(state)

    console.print()
    console.print("[bold]=== Sandbox Integration Test ===[/]")

    # Tool tests
    console.print()
    console.print("[bold]Phase 1: Tool verification[/]")
    st.verify_tools()

    # Shell experience tests
    console.print()
    console.print("[bold]Phase 2: Shell experience[/]")
    st.verify_shell_tools()

    # Security tests
    console.print()
    console.print("[bold]Phase 3: Security isolation[/]")
    st.verify_security()

    # Persistence tests
    console.print()
    console.print("[bold]Phase 4: Persistence[/]")
    import os
    import shutil
    test_home = os.path.expanduser("~/.sandbox/test-sandbox")
    os.makedirs(test_home, exist_ok=True)
    try:
        st.verify_persistence(test_home)
    finally:
        shutil.rmtree(test_home, ignore_errors=True)

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
