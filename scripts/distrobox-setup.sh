#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
DISTROBOX_INI="$REPO_DIR/containers/distrobox.ini"
HOST_USER="$(whoami)"

# Normalize /var/home → /home (containers don't have the atomic symlink)
REPO_DIR="${REPO_DIR/#\/var\/home\//\/home\/}"

usage() {
    echo "Usage: $(basename "$0") [work-eam|work-<name>|personal|gaming|sandbox]"
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
        work-*|personal|gaming|sandbox) CONTAINERS=("$1") ;;
        -h|--help) usage ;;
        *) echo "Error: unknown container '$1'. Choose: work-<name>, personal, gaming, sandbox"; exit 1 ;;
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

    # Install chezmoi, link source to host repo, configure, and apply.
    # All in one sh -c to avoid OCI binary resolution issues with home mounts.
    distrobox enter "$container" -- sh -c "
        cd \"\$HOME\"
        if [ ! -x \"\$HOME/bin/chezmoi\" ]; then
            echo 'Installing chezmoi...'
            curl -fsLS get.chezmoi.io | sh
        fi
        # Link chezmoi source to host repo (live edits without cloning)
        rm -rf \"\$HOME/.local/share/chezmoi\"
        ln -s '$REPO_DIR' \"\$HOME/.local/share/chezmoi\"
        echo 'Linked chezmoi source → $REPO_DIR'
        # Pre-seed platform and context so chezmoi skips those prompts
        mkdir -p \"\$HOME/.config/chezmoi\"
        cat > \"\$HOME/.config/chezmoi/chezmoi.toml\" <<TOML
[data]
  platform = \"distrobox\"
  context = \"$container\"
TOML
        echo ''
        echo 'Configuring chezmoi for $container — answer the prompts below:'
        \"\$HOME/bin/chezmoi\" init --apply
    "

    # Seed credentials from 1Password for non-sandbox containers
    if [ "$container" != "sandbox" ]; then
        echo ""
        echo "Seeding credentials for '$container' (requires 1Password unlock on host)..."
        distrobox enter "$container" -- sh -c "\"\$HOME/.local/bin/setup-creds\""
    fi

    echo ""
    echo "--- '$container' bootstrap complete ---"
done

echo ""
echo "=== Done ==="
echo "Enter a container: distrobox enter ${CONTAINERS[*]}"
