#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
DISTROBOX_INI="$REPO_DIR/containers/distrobox.ini"

echo "=== Distrobox Container Setup ==="

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

# Create containers from ini file
echo "Creating Distrobox containers from $DISTROBOX_INI..."
distrobox assemble create --file "$DISTROBOX_INI"

# Bootstrap dotfiles inside each container
for container in work personal sandbox; do
    echo ""
    echo "--- Bootstrapping '$container' container ---"
    echo "chezmoi will prompt for environment-specific settings."
    echo "Use environment: distrobox-$container"
    echo ""
    distrobox enter "$container" -- sh -c 'curl -fsLS get.chezmoi.io | sh -s -- init --apply rommelporras'
done

echo ""
echo "=== All containers ready ==="
echo "Enter a container: distrobox enter work|personal|sandbox"
