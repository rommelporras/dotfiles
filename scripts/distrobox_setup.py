#!/usr/bin/env python3
"""Set up Distrobox containers and bootstrap chezmoi inside them."""

from __future__ import annotations

import argparse
import os
import sys

from distrobox_lib import (
    bootstrap_chezmoi,
    check_distrobox_available,
    check_not_wsl,
    console,
    container_assemble,
    container_create,
    create_ptyxis_profile,
    distrobox_ini_path,
    full_config_for,
    parse_distrobox_ini,
    partial_config,
    repo_dir,
    run_setup_creds,
    validate_container_name,
)

DEFAULT_CONTAINERS = ["work-eam", "personal"]


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Set up Distrobox containers and bootstrap chezmoi inside them.",
    )
    parser.add_argument(
        "container",
        nargs="?",
        default=None,
        help=(
            "Container to set up (work-eam, work-<name>, personal, "
            "personal-<project>). "
            "With no argument, sets up all default containers."
        ),
    )
    parser.add_argument(
        "--personal-email",
        default=None,
        help="Personal git email (required for non-interactive personal/personal-*).",
    )
    parser.add_argument(
        "--work-email",
        default=None,
        help="Work git email (required for non-interactive work-* containers).",
    )
    return parser.parse_args(argv)


def _resolve_config(container: str, args: argparse.Namespace) -> str:
    """Determine chezmoi config for a container — non-interactive when possible.

    Each context only requires its relevant email:
    - personal, personal-*: needs --personal-email
    - work-*: needs --work-email (personal_email defaults to placeholder)
    Falls back to interactive (partial_config) when the required email is missing.
    """
    if container.startswith("work-"):
        if args.work_email:
            personal = args.personal_email or ""
            console.print("Config mode: non-interactive")
            return full_config_for(container, personal, args.work_email)
    else:
        # personal, personal-*
        if args.personal_email:
            work = args.work_email or "work@placeholder.local"
            console.print("Config mode: non-interactive")
            return full_config_for(container, args.personal_email, work)

    console.print("Config mode: interactive (missing required email flag)")
    return partial_config(container)


def main(argv: list[str] | None = None) -> None:
    args = parse_args(argv)

    # Preflight checks
    check_not_wsl()
    check_distrobox_available()

    # Determine containers
    if args.container is None:
        containers = DEFAULT_CONTAINERS
    else:
        if not validate_container_name(args.container):
            console.print(
                f"[red]Error:[/] unknown container [bold]{args.container}[/]. "
                "Choose: work-<name>, personal, personal-<project>"
            )
            sys.exit(1)
        containers = [args.container]

    console.print("[bold]=== Distrobox Container Setup ===[/]")

    ini = distrobox_ini_path()

    if args.container is None:
        # No arg: create all containers from ini, then bootstrap defaults
        container_assemble(ini)
    else:
        # Specific container: create only that one
        ini_config = parse_distrobox_ini(ini)
        if args.container not in ini_config:
            console.print(
                f"[red]Error:[/] '{args.container}' not found in {ini}. "
                "Add it to containers/distrobox.ini first."
            )
            sys.exit(1)
        cfg = ini_config[args.container]
        home = os.path.expanduser(cfg.get("home", ""))
        container_create(
            args.container,
            cfg.get("image", ""),
            home,
            cfg.get("additional_packages", "").strip('"'),
        )

    repo = repo_dir()

    for container in containers:
        console.print()
        console.print(f"[bold]--- Bootstrapping '{container}' container ---[/]")
        console.print(f"Context: {container} (platform auto-detected as distrobox)")
        console.print()

        config = _resolve_config(container, args)
        bootstrap_chezmoi(container, repo, config)

        console.print()
        run_setup_creds(container)

        console.print()
        console.print(f"[bold]--- '{container}' bootstrap complete ---[/]")

    # Create Ptyxis terminal profiles
    console.print()
    console.print("[bold]=== Ptyxis Profiles ===[/]")
    for container in containers:
        create_ptyxis_profile(container)
    console.print("[yellow]  Close all Ptyxis windows and reopen for new profiles to appear.[/]")

    console.print()
    console.print("[bold]=== Done ===[/]")
    console.print()
    for container in containers:
        console.print(f"  Enter container:  [bold]distrobox enter {container}[/]")
    console.print()
    console.print("[yellow]  Verify setup-creds ran correctly:[/]")
    console.print("    1. Enter the container")
    console.print("    2. Run [bold]atuin sync[/] — should sync without errors")
    console.print("    3. If it fails, unlock 1Password and re-run: [bold]setup-creds[/]")


if __name__ == "__main__":
    main()
