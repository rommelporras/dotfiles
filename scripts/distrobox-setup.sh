#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
DISTROBOX_INI="$REPO_DIR/containers/distrobox.ini"

usage() {
    echo "Usage: $(basename "$0") [work-eam|work-<name>|personal|sandbox]"
    echo ""
    echo "Set up Distrobox containers and bootstrap chezmoi inside them."
    echo "With no argument, sets up all containers defined in distrobox.ini."
    echo "With an argument, sets up only the specified container."
    exit 1
}

# Verify we're on Aurora/Bluefin (not WSL)
if grep -qi microsoft /proc/version 2>/dev/null; then
    echo "Error: This script is for Aurora/Bluefin hosts, not WSL."
    echo "On WSL, chezmoi manages dotfiles directly without Distrobox."
    exit 1
fi

if ! command -v distrobox &>/dev/null; then
    echo "Error: distrobox not found. Is this an Aurora/Bluefin system?"
    exit 1
fi

# Determine which containers to set up
if [ $# -eq 0 ]; then
    CONTAINERS=(work-eam personal sandbox)
elif [ $# -eq 1 ]; then
    case "$1" in
        work-*|personal|sandbox) CONTAINERS=("$1") ;;
        -h|--help) usage ;;
        *) echo "Error: unknown container '$1'. Choose: work-<name>, personal, sandbox"; exit 1 ;;
    esac
else
    usage
fi

echo "=== Distrobox Container Setup ==="

# Create containers from ini file
echo "Creating Distrobox containers from $DISTROBOX_INI..."
distrobox assemble create --file "$DISTROBOX_INI"

# Bootstrap dotfiles inside each container
for container in "${CONTAINERS[@]}"; do
    echo ""
    echo "--- Bootstrapping '$container' container ---"
    echo "Context: $container (platform auto-detected as distrobox)"
    echo ""
    distrobox enter "$container" -- sh -c "curl -fsLS get.chezmoi.io | sh -s -- init --apply rommelporras --promptString context=$container"
done

echo ""
echo "=== Done ==="
echo "Enter a container: distrobox enter ${CONTAINERS[*]}"
