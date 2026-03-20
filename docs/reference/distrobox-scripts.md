# Distrobox Scripts Reference

Python scripts for creating, bootstrapping, and testing Distrobox containers.

## Prerequisites

Run these on the **Aurora host** (not inside a container):

1. **uv** (Python project manager) — install via brew if not already present:
   ```bash
   brew install uv
   ```

2. **Distrobox** — installed by `ujust devmode` on Aurora DX.

3. **1Password unlocked** (for credential seeding):
   Open the 1Password desktop app and unlock it. Ensure CLI integration is enabled:
   Settings → Developer → **Integrate with 1Password CLI**.

4. **Working directory** — all commands assume you're in the repo root:
   ```bash
   cd ~/personal/dotfiles
   ```

## Setup Script

```bash
uv run python scripts/distrobox_setup.py [container] [--personal-email EMAIL] [--work-email EMAIL]
```

Creates a Distrobox container, bootstraps chezmoi inside it, and runs credential seeding.

### Arguments

| Argument | Required | Description |
|---|---|---|
| `container` | No | Container name to set up. Without this, creates all default containers (work-eam, personal). |
| `--personal-email` | No | Personal git email. Required for non-interactive personal/personal-\*. |
| `--work-email` | No | Work git email. Required for non-interactive work-\* containers. |

### Config modes

The script has two config modes. Each context only needs its relevant email:

**Non-interactive (recommended):** Provide the email flag relevant to the context.
All other config values are derived automatically from the container name.

| Context | Required flag | Optional flag |
|---|---|---|
| personal, personal-\* | `--personal-email` | `--work-email` |
| work-\* | `--work-email` | `--personal-email` |

```bash
# Personal — only needs personal email
uv run python scripts/distrobox_setup.py personal \
  --personal-email git@rommelporras.com

# Work — only needs work email
uv run python scripts/distrobox_setup.py work-eam \
  --work-email work@company.com
```

**Interactive (fallback):** When the required email flag is omitted, the script pre-seeds
only `platform=distrobox` and `context=<name>`. chezmoi then prompts for the remaining
values (email, Atuin, credentials).

```bash
# Interactive — chezmoi prompts for remaining values
uv run python scripts/distrobox_setup.py personal
```

### Config derivation by context

All values except email are automatically derived from the container name:

| Value | personal | personal-\<project\> | work-\<name\> |
|---|---|---|---|
| `has_homelab_creds` | true | false | false |
| `has_work_creds` | false | false | true |
| `atuin_sync_address` | self-hosted URL | self-hosted URL | self-hosted URL |
| `atuin_account` | personal | personal | \<context name\> |

### Examples

```bash
# Set up personal container (only personal email needed)
uv run python scripts/distrobox_setup.py personal \
  --personal-email git@rommelporras.com

# Set up work container (only work email needed)
uv run python scripts/distrobox_setup.py work-eam \
  --work-email work@company.com

# Set up all default containers (provide both for mixed contexts)
uv run python scripts/distrobox_setup.py \
  --personal-email git@rommelporras.com \
  --work-email work@company.com
```

### What the script does

1. Creates the container from `containers/distrobox.ini` (single container or all)
2. Installs chezmoi inside the container (`~/bin/chezmoi`)
3. Symlinks `~/.local/share/chezmoi` → host repo (uncommitted changes apply immediately)
4. Writes chezmoi config to `~/.config/chezmoi/chezmoi.toml`
5. Runs `chezmoi init --apply`
6. Runs `setup-creds` to seed plugins, MCP, and credentials from 1Password

### Verifying it worked

After the script completes, enter the container and check:

```bash
distrobox enter personal
```

Inside the container:

```bash
# Core files deployed
test -f ~/.zshrc && echo "OK: .zshrc" || echo "FAIL: .zshrc"
test -f ~/.gitconfig && echo "OK: .gitconfig" || echo "FAIL: .gitconfig"
test -L ~/.local/share/chezmoi && echo "OK: chezmoi symlink" || echo "FAIL: chezmoi symlink"

# Tools installed (varies by context)
~/bin/chezmoi --version
command -v glab && echo "OK: glab" || echo "FAIL: glab"               # personal, personal-*
command -v ansible && echo "OK: ansible" || echo "FAIL: ansible"       # personal only
command -v op && echo "OK: op CLI" || echo "FAIL: op CLI"              # personal-* only
test -f ~/.atuin/bin/atuin && echo "OK: atuin" || echo "FAIL: atuin"

# Credential seeding ran
test -x ~/.local/bin/setup-creds && echo "OK: setup-creds" || echo "FAIL: setup-creds"

# Check .zshrc has expected content
grep -q "1password/agent.sock" ~/.zshrc && echo "OK: 1Password SSH" || echo "N/A: no 1Password SSH"
```

Or run the automated integration test to verify everything:

```bash
# From the host (not inside a container)
uv run python scripts/test_distrobox_integration.py personal
```

### Valid container names

Must match: `work-<name>`, `personal`, or `personal-<project>`.
The container must also be defined in `containers/distrobox.ini`.

---

## Integration Test Script

```bash
uv run python scripts/test_distrobox_integration.py [--all] [--keep] [container ...]
```

Full lifecycle test: delete → create → bootstrap → verify → delete.

### Arguments

| Argument | Description |
|---|---|
| `container` | Container(s) to test. Default: `personal`. |
| `--all` | Test all containers: personal, work-eam. |
| `--keep` | Keep container after test for manual inspection (skip teardown). |

### Examples

```bash
# Test default container (personal)
uv run python scripts/test_distrobox_integration.py

# Test all containers (~49 assertions across 2 containers)
uv run python scripts/test_distrobox_integration.py --all

# Test specific container, keep it for inspection
uv run python scripts/test_distrobox_integration.py --keep personal

# Test multiple specific containers
uv run python scripts/test_distrobox_integration.py personal work-eam
```

### Assertions per container

| Container | Assertions | Key checks |
|---|---|---|
| personal | ~24 | glab, ansible, kubectl, atuin sync, 1Password SSH, homelab aliases, IDE forwarding |
| work-eam | ~25 | terraform, aws CLI, kubectl, EAM aliases, atuin sync, 1Password SSH, IDE forwarding |

The test uses `full_config_for()` with test email defaults — fully non-interactive,
no 1Password dependency.

---

## Shared Library

`scripts/distrobox_lib.py` contains all shared functions used by both scripts.
Not intended to be run directly.

### Key functions

| Function | Description |
|---|---|
| `parse_distrobox_ini(path)` | Parse `containers/distrobox.ini` into a dict of container configs |
| `container_create(name, image, home, packages)` | Create a single container |
| `container_assemble(ini_path)` | Create all containers from ini file |
| `bootstrap_chezmoi(name, repo, config)` | Install chezmoi, symlink source, configure, and apply (15-min timeout) |
| `full_config_for(context, email, work_email)` | Generate complete non-interactive chezmoi config |
| `partial_config(context)` | Generate minimal config (interactive — prompts for remaining values) |
| `run_setup_creds(container)` | Run `setup-creds` inside a container |
| `validate_container_name(name)` | Check name matches allowed pattern |
