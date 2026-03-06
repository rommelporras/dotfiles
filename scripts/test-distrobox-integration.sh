#!/bin/bash
set -euo pipefail

# ─── Distrobox Integration Test ─────────────────────────────────────────────
# Full lifecycle test: delete → create → bootstrap → verify → delete
# Tests real container creation and chezmoi deployment, not just templates.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
DISTROBOX_INI="$REPO_DIR/containers/distrobox.ini"
HOST_USER="$(whoami)"

# Normalize /var/home → /home (containers don't have the atomic symlink)
REPO_DIR="${REPO_DIR/#\/var\/home\//\/home\/}"

# Test counters
PASS=0
FAIL=0
ERRORS=()

# CLI flags
KEEP=false
ALL=false
CONTAINERS=()

# ─── CLI parsing ─────────────────────────────────────────────────────────────

usage() {
    echo "Usage: $(basename "$0") [--all] [--keep] [container-name]"
    echo ""
    echo "Integration test for distrobox + chezmoi bootstrap."
    echo "Lifecycle: delete → create → bootstrap → verify → delete."
    echo ""
    echo "Options:"
    echo "  --all   Test all containers (sandbox, personal, personal-fintrack, work-eam)"
    echo "  --keep  Keep container after test (for manual inspection)"
    echo ""
    echo "Default container: personal-fintrack"
    exit 1
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --all)  ALL=true; shift ;;
        --keep) KEEP=true; shift ;;
        -h|--help) usage ;;
        -*) echo "Unknown option: $1"; usage ;;
        *)  CONTAINERS+=("$1"); shift ;;
    esac
done

if [ "$ALL" = true ]; then
    CONTAINERS=(sandbox personal personal-fintrack work-eam)
elif [ ${#CONTAINERS[@]} -eq 0 ]; then
    CONTAINERS=(personal-fintrack)
fi

# ─── Preflight ───────────────────────────────────────────────────────────────

if [ -n "${DISTROBOX_ENTER_PATH:-}" ]; then
    echo "Error: Cannot run inside a distrobox container."
    exit 1
fi

if ! command -v distrobox &>/dev/null; then
    echo "Error: distrobox not found."
    exit 1
fi

if [ ! -f "$DISTROBOX_INI" ]; then
    echo "Error: $DISTROBOX_INI not found."
    exit 1
fi

# ─── Test helpers ────────────────────────────────────────────────────────────

_run_in() {
    local container="$1"; shift
    distrobox enter "$container" -- sh -c "$*" 2>/dev/null
}

assert_exec() {
    local desc="$1" container="$2"; shift 2
    local cmd="$*"
    if _run_in "$container" "$cmd"; then
        echo "  PASS: $desc"
        PASS=$((PASS + 1))
    else
        echo "  FAIL: $desc"
        FAIL=$((FAIL + 1))
        ERRORS+=("[$container] $desc")
    fi
}

assert_file() {
    local desc="$1" container="$2" path="$3"
    assert_exec "$desc" "$container" "test -f \"$path\""
}

assert_no_file() {
    local desc="$1" container="$2" path="$3"
    assert_exec "$desc" "$container" "test ! -f \"$path\""
}

assert_dir() {
    local desc="$1" container="$2" path="$3"
    assert_exec "$desc" "$container" "test -d \"$path\""
}

assert_no_dir() {
    local desc="$1" container="$2" path="$3"
    assert_exec "$desc" "$container" "test ! -d \"$path\""
}

assert_symlink() {
    local desc="$1" container="$2" path="$3"
    assert_exec "$desc" "$container" "test -L \"$path\""
}

assert_contains() {
    local desc="$1" container="$2" file="$3" pattern="$4"
    assert_exec "$desc" "$container" "grep -q '$pattern' \"$file\""
}

assert_not_contains() {
    local desc="$1" container="$2" file="$3" pattern="$4"
    assert_exec "$desc" "$container" "! grep -q '$pattern' \"$file\""
}

assert_executable() {
    local desc="$1" container="$2" path="$3"
    assert_exec "$desc" "$container" "test -x \"$path\""
}

# ─── chezmoi config templates ───────────────────────────────────────────────

chezmoi_config_for() {
    local context="$1"
    local atuin_account="none"
    local has_homelab_creds="false"
    local has_work_creds="false"

    case "$context" in
        personal)
            atuin_account="rommel-personal"
            has_homelab_creds="true"
            ;;
        personal-*)
            atuin_account="rommel-personal"
            ;;
        work-*)
            atuin_account="rommel-eam"
            has_work_creds="true"
            ;;
        sandbox)
            ;;
    esac

    # has_op_cli is derived: true for personal-<project> on distrobox
    local has_op_cli="false"
    if [[ "$context" == personal-* ]]; then
        has_op_cli="true"
    fi

    cat <<TOML
[data]
  platform = "distrobox"
  context = "$context"
  personal_email = "test@example.com"
  work_email = "test@company.com"
  has_work_creds = $has_work_creds
  has_homelab_creds = $has_homelab_creds
  has_op_cli = $has_op_cli
  atuin_sync_address = ""
  atuin_account = "$atuin_account"
TOML
}

# ─── Per-container assertions ────────────────────────────────────────────────

verify_common() {
    local c="$1"
    echo "  --- Common assertions ---"
    assert_file     "chezmoi binary exists"        "$c" "\$HOME/bin/chezmoi"
    assert_file     ".zshrc exists"                 "$c" "\$HOME/.zshrc"
    assert_exec     ".zshrc is not empty"           "$c" "test -s \$HOME/.zshrc"
    assert_file     ".gitconfig exists"             "$c" "\$HOME/.gitconfig"
    assert_symlink  "chezmoi source is symlink"     "$c" "\$HOME/.local/share/chezmoi"
    assert_exec     "NVM installed"                 "$c" "test -f \$HOME/.nvm/nvm.sh || test -f \$HOME/.config/nvm/nvm.sh"
    assert_exec     "NVM functional"                "$c" "export NVM_DIR=\${NVM_DIR:-\$HOME/.config/nvm} && . \$NVM_DIR/nvm.sh && nvm --version"
    assert_exec     "fzf installed"                 "$c" "test -d \$HOME/.fzf"
}

verify_personal() {
    local c="personal"
    echo "  --- personal-specific assertions ---"
    assert_executable "setup-creds is executable"   "$c" "\$HOME/.local/bin/setup-creds"
    assert_exec       "glab installed"              "$c" "command -v glab"
    assert_file       "atuin installed"             "$c" "\$HOME/.atuin/bin/atuin"
    assert_contains   "SSH_AUTH_SOCK = 1Password"   "$c" "\$HOME/.zshrc" "1password/agent.sock"
    assert_contains   "invoicetron alias"           "$c" "\$HOME/.zshrc" "invoicetron"
    assert_contains   "kubectl-homelab alias"       "$c" "\$HOME/.zshrc" "kubectl-homelab"
    assert_contains   "OTEL env vars"               "$c" "\$HOME/.zshrc" "OTEL_METRICS_EXPORTER"
    assert_exec       "ansible installed"           "$c" "command -v ansible"
    assert_not_contains "no OP_BIOMETRIC in .zshrc" "$c" "\$HOME/.zshrc" "OP_BIOMETRIC_UNLOCK_ENABLED"
}

verify_personal_fintrack() {
    local c="personal-fintrack"
    echo "  --- personal-fintrack-specific assertions ---"
    assert_executable "setup-creds is executable"   "$c" "\$HOME/.local/bin/setup-creds"
    assert_exec       "glab installed"              "$c" "command -v glab"
    assert_file       "atuin installed"             "$c" "\$HOME/.atuin/bin/atuin"
    assert_not_contains "no 1Password SSH socket"   "$c" "\$HOME/.zshrc" "1password/agent.sock"
    assert_no_dir     ".kube/ does not exist"       "$c" "\$HOME/.kube"
    assert_exec       "op CLI installed"            "$c" "command -v op"
    assert_exec       "bun installed"               "$c" "\$HOME/.bun/bin/bun --version"
    assert_contains   "OP_BIOMETRIC in .zshrc"      "$c" "\$HOME/.zshrc" "OP_BIOMETRIC_UNLOCK_ENABLED"
    assert_contains   "OTEL env vars"               "$c" "\$HOME/.zshrc" "OTEL_METRICS_EXPORTER"
    assert_not_contains "no invoicetron alias"      "$c" "\$HOME/.zshrc" "invoicetron"
    assert_not_contains "no kubectl-homelab alias"  "$c" "\$HOME/.zshrc" "kubectl-homelab"
    assert_contains   "BUN_INSTALL in .zshrc"       "$c" "\$HOME/.zshrc" "BUN_INSTALL"
}

verify_work_eam() {
    local c="work-eam"
    echo "  --- work-eam-specific assertions ---"
    assert_executable "setup-creds is executable"   "$c" "\$HOME/.local/bin/setup-creds"
    assert_contains   "SSH_AUTH_SOCK = 1Password"   "$c" "\$HOME/.zshrc" "1password/agent.sock"
    assert_contains   "terraform alias tfi"         "$c" "\$HOME/.zshrc" "tfi"
    assert_contains   "EAM alias"                   "$c" "\$HOME/.zshrc" "eam-sre"
    assert_not_contains "no invoicetron alias"      "$c" "\$HOME/.zshrc" "invoicetron"
    assert_not_contains "no OP_BIOMETRIC"           "$c" "\$HOME/.zshrc" "OP_BIOMETRIC_UNLOCK_ENABLED"
}

verify_sandbox() {
    local c="sandbox"
    echo "  --- sandbox-specific assertions ---"
    assert_no_file      "no setup-creds"            "$c" "\$HOME/.local/bin/setup-creds"
    assert_contains     "SSH_AUTH_SOCK unset"        "$c" "\$HOME/.zshrc" "unset SSH_AUTH_SOCK"
    assert_not_contains "no 1Password socket"       "$c" "\$HOME/.zshrc" "1password/agent.sock"
    assert_not_contains "no OTEL env vars"          "$c" "\$HOME/.zshrc" "OTEL_METRICS_EXPORTER"
    assert_not_contains "no invoicetron alias"      "$c" "\$HOME/.zshrc" "invoicetron"
}

# ─── Test a single container ────────────────────────────────────────────────

test_container() {
    local container="$1"
    local test_pass_before=$PASS
    local test_fail_before=$FAIL

    echo ""
    echo "================================================================"
    echo "Testing: $container"
    echo "================================================================"

    # Phase 1: Clean slate
    echo ""
    echo "Phase 1: Clean slate (removing $container if exists)..."
    distrobox stop -Y "$container" 2>/dev/null || true
    distrobox rm --force "$container" 2>/dev/null || true
    echo "  Done."

    # Phase 2: Create container (single container only, not full assemble)
    echo ""
    echo "Phase 2: Creating container..."
    # Parse ini for this container's settings
    local image home packages
    image=$(awk "/^\[$container\]/{found=1;next} /^\[/{found=0} found && /^image=/{print \$0}" "$DISTROBOX_INI" | cut -d= -f2)
    home=$(awk "/^\[$container\]/{found=1;next} /^\[/{found=0} found && /^home=/{print \$0}" "$DISTROBOX_INI" | cut -d= -f2)
    packages=$(awk "/^\[$container\]/{found=1;next} /^\[/{found=0} found && /^additional_packages=/{print \$0}" "$DISTROBOX_INI" | cut -d= -f2- | tr -d '"')

    if [ -z "$image" ]; then
        echo "  ERROR: Container '$container' not found in $DISTROBOX_INI"
        FAIL=$((FAIL + 1))
        ERRORS+=("[$container] Container not defined in distrobox.ini")
        return
    fi

    # Expand ~ in home path
    home="${home/#\~/$HOME}"

    distrobox create --name "$container" --image "$image" --home "$home" \
        --additional-packages "$packages" --yes 2>&1
    echo "  Done."

    # Phase 3: Bootstrap chezmoi (non-interactive)
    echo ""
    echo "Phase 3: Bootstrapping chezmoi (non-interactive)..."
    local config
    config="$(chezmoi_config_for "$container")"

    distrobox enter "$container" -- sh -c "
        cd \"\$HOME\"

        # Install chezmoi
        if [ ! -x \"\$HOME/bin/chezmoi\" ]; then
            curl -fsLS get.chezmoi.io | sh
        fi

        # Link source to host repo
        mkdir -p \"\$HOME/.local/share\"
        rm -rf \"\$HOME/.local/share/chezmoi\"
        ln -s '$REPO_DIR' \"\$HOME/.local/share/chezmoi\"

        # Pre-seed FULL config (no interactive prompts)
        mkdir -p \"\$HOME/.config/chezmoi\"
        cat > \"\$HOME/.config/chezmoi/chezmoi.toml\" <<'TOML'
$config
TOML

        # Clear run_once state so bootstrap re-runs on fresh container
        \"\$HOME/bin/chezmoi\" state delete-bucket --bucket=scriptState || true

        # Apply
        \"\$HOME/bin/chezmoi\" init --apply
    "
    echo "  Done."

    # Phase 4: Verify deployed state
    echo ""
    echo "Phase 4: Verifying deployed state..."
    verify_common "$container"

    case "$container" in
        personal)          verify_personal ;;
        personal-fintrack) verify_personal_fintrack ;;
        personal-*)        verify_personal_fintrack ;;  # fallback for other personal-<project>
        work-eam)          verify_work_eam ;;
        work-*)            verify_work_eam ;;  # fallback for other work-<name>
        sandbox)           verify_sandbox ;;
    esac

    # Phase 5: Teardown
    echo ""
    if [ "$KEEP" = true ]; then
        echo "Phase 5: Skipped (--keep flag). Container '$container' is still running."
    else
        echo "Phase 5: Teardown..."
        distrobox stop -Y "$container" 2>/dev/null || true
        distrobox rm --force "$container" 2>/dev/null || true
        echo "  Done."
    fi

    local p=$((PASS - test_pass_before))
    local f=$((FAIL - test_fail_before))
    echo ""
    echo "  Container '$container': $p passed, $f failed"
}

# ─── Trap for cleanup on unexpected exit ─────────────────────────────────────

cleanup() {
    if [ "$KEEP" = true ]; then
        return
    fi
    for c in "${CONTAINERS[@]}"; do
        distrobox stop -Y "$c" 2>/dev/null || true
        distrobox rm --force "$c" 2>/dev/null || true
    done
}

# Only trap on error/interrupt — normal exit handles cleanup per-container
trap cleanup INT TERM

# ─── Main ────────────────────────────────────────────────────────────────────

echo "=== Distrobox Integration Test ==="
echo "Containers: ${CONTAINERS[*]}"
echo "Keep after test: $KEEP"

for container in "${CONTAINERS[@]}"; do
    test_container "$container"
done

# ─── Summary ─────────────────────────────────────────────────────────────────

echo ""
echo "================================================================"
echo "SUMMARY"
echo "================================================================"
echo "Total: $((PASS + FAIL)) assertions, $PASS passed, $FAIL failed"

if [ ${#ERRORS[@]} -gt 0 ]; then
    echo ""
    echo "Failures:"
    for err in "${ERRORS[@]}"; do
        echo "  - $err"
    done
fi

echo ""
if [ "$FAIL" -eq 0 ]; then
    echo "ALL TESTS PASSED"
    exit 0
else
    echo "TESTS FAILED"
    exit 1
fi
