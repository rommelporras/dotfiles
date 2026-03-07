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
    distrobox_ini_path,
    full_config_for,
    parse_distrobox_ini,
    partial_config,
    repo_dir,
    run_setup_creds,
    validate_container_name,
)

DEFAULT_CONTAINERS = ["work-eam", "personal", "sandbox"]


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
            "personal-<project>, gaming, sandbox). "
            "With no argument, sets up all default containers."
        ),
    )
    parser.add_argument(
        "--personal-email",
        default=None,
        help="Personal git email. When both emails are provided, skips interactive prompts.",
    )
    parser.add_argument(
        "--work-email",
        default=None,
        help="Work git email. When both emails are provided, skips interactive prompts.",
    )
    return parser.parse_args(argv)


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
                "Choose: work-<name>, personal, personal-<project>, gaming, sandbox"
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

    # Determine config mode: non-interactive (emails provided) or interactive
    non_interactive = args.personal_email is not None and args.work_email is not None
    if non_interactive:
        console.print(f"Config mode: non-interactive (emails provided)")

    for container in containers:
        console.print()
        console.print(f"[bold]--- Bootstrapping '{container}' container ---[/]")
        console.print(f"Context: {container} (platform auto-detected as distrobox)")
        console.print()

        if container == "sandbox":
            # Sandbox needs no real config — always non-interactive
            config = full_config_for(container, "sandbox@localhost", "sandbox@localhost")
        elif non_interactive:
            config = full_config_for(container, args.personal_email, args.work_email)
        else:
            config = partial_config(container)
        bootstrap_chezmoi(container, repo, config)

        # Seed credentials for non-sandbox containers
        if container != "sandbox":
            console.print()
            run_setup_creds(container)

        console.print()
        console.print(f"[bold]--- '{container}' bootstrap complete ---[/]")

    console.print()
    console.print("[bold]=== Done ===[/]")
    console.print(f"Enter a container: distrobox enter {' '.join(containers)}")


if __name__ == "__main__":
    main()
