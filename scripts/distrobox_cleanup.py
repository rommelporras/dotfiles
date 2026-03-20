#!/usr/bin/env python3
"""Clean up Distrobox containers — stop, remove, and optionally wipe home directory."""

from __future__ import annotations

import argparse
import os
import shutil
import sys

from distrobox_lib import (
    check_distrobox_available,
    check_not_wsl,
    console,
    container_rm,
    container_stop,
    remove_ptyxis_profile,
    validate_container_name,
)


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Clean up Distrobox containers.",
    )
    parser.add_argument(
        "containers",
        nargs="*",
        help=(
            "Container(s) to remove (work-eam, work-<name>, personal, "
            "personal-<project>). Required unless --all is used."
        ),
    )
    parser.add_argument(
        "--wipe-home",
        action="store_true",
        default=False,
        help="Also delete the container's home directory (~/.distrobox/<name>/).",
    )
    parser.add_argument(
        "--all",
        action="store_true",
        default=False,
        help="Remove ALL containers defined in distrobox.ini.",
    )
    return parser.parse_args(argv)


def get_all_containers() -> list[str]:
    """Get all container names from distrobox.ini."""
    from distrobox_lib import distrobox_ini_path, parse_distrobox_ini

    ini = parse_distrobox_ini(distrobox_ini_path())
    return list(ini.keys())


def cleanup_container(name: str, wipe_home: bool) -> None:
    """Stop, remove, and optionally wipe a single container."""
    console.print(f"\n[bold]--- Cleaning up: {name} ---[/]")

    console.print(f"  Stopping {name}...")
    container_stop(name)

    console.print(f"  Removing {name}...")
    container_rm(name)

    if wipe_home:
        home_dir = os.path.expanduser(f"~/.distrobox/{name}")
        if os.path.isdir(home_dir):
            console.print(f"  Wiping home directory: {home_dir}")
            shutil.rmtree(home_dir)
            console.print(f"  [green]Removed {home_dir}[/]")
        else:
            console.print(f"  [dim]No home directory at {home_dir}[/]")

    console.print(f"  Removing Ptyxis profile...")
    remove_ptyxis_profile(name)

    console.print(f"  [green]Done: {name}[/]")


def main(argv: list[str] | None = None) -> None:
    args = parse_args(argv)

    check_not_wsl()
    check_distrobox_available()

    if args.all:
        containers = get_all_containers()
        if not containers:
            console.print("[red]Error:[/] No containers found in distrobox.ini")
            sys.exit(1)
    elif args.containers:
        containers = args.containers
    else:
        console.print("[red]Error:[/] Specify container name(s) or use --all")
        sys.exit(1)

    # Validate all names first
    for name in containers:
        if not validate_container_name(name):
            console.print(
                f"[red]Error:[/] invalid container name [bold]{name}[/]. "
                "Choose: work-<name>, personal, personal-<project>"
            )
            sys.exit(1)

    action = "remove + wipe home" if args.wipe_home else "remove (keep home)"
    console.print(f"[bold]=== Distrobox Cleanup ===[/]")
    console.print(f"Containers: {', '.join(containers)}")
    console.print(f"Action: {action}")

    for name in containers:
        cleanup_container(name, args.wipe_home)

    console.print(f"\n[bold green]Cleanup complete: {len(containers)} container(s)[/]")


if __name__ == "__main__":
    main()
