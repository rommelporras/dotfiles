#!/usr/bin/env python3
"""Set up Distrobox containers and bootstrap chezmoi inside them."""

from __future__ import annotations

import argparse
import sys

from distrobox_lib import (
    bootstrap_chezmoi,
    check_distrobox_available,
    check_not_wsl,
    console,
    container_assemble,
    distrobox_ini_path,
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

    # Create containers from ini file
    ini = distrobox_ini_path()
    container_assemble(ini)

    repo = repo_dir()

    for container in containers:
        console.print()
        console.print(f"[bold]--- Bootstrapping '{container}' container ---[/]")
        console.print(f"Context: {container} (platform auto-detected as distrobox)")
        console.print()

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
