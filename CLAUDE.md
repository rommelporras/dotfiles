# CLAUDE.md

## Project Overview

chezmoi-managed dotfiles repository for bootstrapping consistent dev environments across
WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora), and Distrobox containers. Uses a
two-variable model (`platform` + `context`) for scalable environment targeting with
credential isolation and AI agent sandboxing.

## Branching

Trunk-based development. All commits go directly to `main`. No feature branches.

## Repository Structure

```
dotfiles/
├── .chezmoiroot                         # Points chezmoi source to home/
├── home/                                # chezmoi source dir → maps to ~/
│   ├── .chezmoi.toml.tmpl               # Interactive prompts (chezmoi init)
│   ├── .chezmoiignore                   # Per-environment file skipping
│   ├── .chezmoiexternal.toml            # External deps (oh-my-zsh, zsh-autosuggestions)
│   ├── dot_zshrc.tmpl                   # Shell config (templated per env)
│   ├── dot_gitconfig.tmpl               # Git config (conditional includes)
│   ├── run_once_before_bootstrap.sh.tmpl # First-run setup (installs tools)
│   ├── dot_local/bin/                   # User scripts (~/.local/bin/)
│   │   └── executable_setup-creds.tmpl  # Credential + plugin seeding for distrobox
│   └── dot_config/                      # ~/.config/ files
├── scripts/                             # Setup automation (Python, invoke via uv run)
│   ├── distrobox_setup.py               # Container creation + chezmoi bootstrap
│   ├── distrobox_lib.py                 # Shared library for distrobox scripts
│   ├── test_distrobox_integration.py    # E2E test: delete → create → bootstrap → verify → delete
│   └── windows-git-setup.ps1            # Windows git setup for WSL
├── docs/                                # Reference documentation
│   └── distrobox-scripts.md             # Distrobox scripts parameter reference
├── containers/                          # Distrobox + Podman definitions
│   ├── distrobox.ini                    # Container definitions (work-eam, personal, personal-fintrack, sandbox)
│   └── Containerfile.ai-sandbox         # AI sandbox container (Ubuntu 24.04, Claude Code, Node.js, Python, uv)
├── bin/                                 # CLI tools (ai-sandbox)
└── hooks/                               # Git hooks (gitleaks)
```

## chezmoi Conventions

- `dot_` prefix → `.` in target (e.g., `dot_zshrc` → `.zshrc`)
- `private_dot_` prefix → `.` with owner-only permissions (0600 files, 0700 dirs)
- `.tmpl` suffix → Go text/template processing
- `run_once_before_` prefix → script that runs once on first apply, before files
- Template data defined in `.chezmoi.toml.tmpl`, stored locally in `~/.config/chezmoi/chezmoi.toml`

## Environment Model

Templates use two variables:

- **`platform`** (auto-detected): `wsl`, `aurora`, `distrobox` — controls SSH agent, package manager, system paths
- **`context`** (user-selected): `personal`, `personal-<project>`, `work-eam`, `work-<name>`, `gaming`, `sandbox` — controls aliases, credentials, tools

| Platform | Context | SSH Agent | Key differences |
|---|---|---|---|
| wsl | work-eam | 1Password via npiperelay | NVM/Bun, work + personal creds |
| wsl | gaming | 1Password via npiperelay | NVM/Bun, personal creds |
| aurora | personal | 1Password native socket | Immutable OS, no chsh, Atuin sync |
| distrobox | work-eam | 1Password via absolute host path | AWS CLI, kubectl, Terraform, work creds |
| distrobox | personal | 1Password via absolute host path | kubectl, homelab kubeconfig, glab, Ansible |
| distrobox | personal-\<project\> | No 1Password SSH (manual keys) | Native op CLI, Bun, Playwright, glab, no homelab |
| distrobox | sandbox | Fallback ssh-agent | No creds, no Claude config |

**Adding contexts:** For work: add container to `distrobox.ini`, add job-specific aliases
in `dot_zshrc.tmpl`, run `distrobox_setup.py work-acme`. Shared work tools apply via
`hasPrefix .context "work-"`. For personal projects: add container to `distrobox.ini`,
run `distrobox_setup.py personal-<project>`. Shared personal tools (glab, Bun, Playwright,
native `op` CLI) apply via `hasPrefix .context "personal-"`.

### Distrobox chezmoi workflow

`scripts/distrobox_setup.py` bootstraps containers with chezmoi (see
[docs/distrobox-scripts.md](docs/distrobox-scripts.md) for full reference):
1. Installs chezmoi inside the container (`~/bin/chezmoi`)
2. Symlinks `~/.local/share/chezmoi` → host repo (uncommitted changes apply immediately)
3. Writes chezmoi config (non-interactive when `--personal-email`/`--work-email` provided)
4. Runs `chezmoi init --apply`
5. Runs `setup-creds` to seed plugins, MCP, and credentials from 1Password (non-sandbox only)

Inside containers, `$HOME` is `~/.distrobox/<context>/` (NOT the host home). Paths to
host resources (e.g. 1Password socket) must use absolute paths like `/home/<user>/...`.

To update dotfiles inside a container: `~/bin/chezmoi apply -v && exec zsh`

### IDE forwarding

Non-sandbox distrobox containers alias `code`, `antigravity`, and `agy` to forward
to the Aurora host via `distrobox-host-exec`. No need to install IDEs in containers.

### Credential and plugin seeding

`setup-creds` (deployed to `~/.local/bin/`) uses `distrobox-host-exec op` to pull
secrets from 1Password on the host. Handles:
1. Claude Code plugin marketplace registration and plugin installation
2. Context7 MCP server registration (requires API key from 1Password)
3. Atuin login and sync (with backup/restore error handling)
4. GitLab auth (personal context)
5. Manual step instructions for kubeconfig and AWS

Plugins run before credentials so 1Password failures don't block plugin setup.
Skipped entirely in sandbox (excluded via `.chezmoiignore`).

## Common Commands

```bash
# Preview changes without applying
chezmoi diff

# Apply changes
chezmoi apply -v

# Edit a managed file (opens in source dir)
chezmoi edit ~/.zshrc

# Re-run bootstrap (won't re-run — it's run_once_)
chezmoi state delete-bucket --bucket=scriptState
chezmoi apply

# Add a new file to chezmoi management
chezmoi add ~/.config/some/config.toml

# Update external dependencies (oh-my-zsh, plugins)
chezmoi update
```

## Claude Code Config

Claude Code global config (`~/.claude/`) is managed by a separate repo:
[claude-config](https://github.com/rommelporras/claude-config) — cloned to
`~/personal/claude-config` and symlinked into `~/.claude/`.

The bootstrap script handles cloning and symlinking. On distrobox containers,
symlinks point to the host's clone via absolute paths. The `.claude/` directory
is blanket-ignored in `.chezmoiignore` so chezmoi never touches it.

Plugin and MCP setup is handled by `setup-creds` (distrobox) or manual CLI
commands (Aurora/WSL) — see bootstrap post-install instructions.

## dotctl (Go CLI)

Go CLI tool living at repo root (`cmd/`, `internal/`, `go.mod`). chezmoi only
sees `home/` via `.chezmoiroot` — Go files are invisible to it.

### Architecture

Single binary, three modes:
- `dotctl collect` — gather status, push to OTel Collector via OTLP gRPC
- `dotctl status` — query Prometheus + Loki, render terminal tables
- `dotctl status --live` — gather + display directly (no cluster dependency)

### Key Packages

- `internal/collector/` — gathers chezmoi status, tool inventory, credentials
- `internal/push/` — OTLP gRPC push to OTel Collector
- `internal/query/` — Prometheus + Loki HTTP API queries
- `internal/display/` — terminal table rendering with lipgloss

### Conventions

- TDD: write failing test first, then implement
- Shell out to `chezmoi` and `distrobox` CLIs — don't reimplement them
- Interfaces for external commands to enable test mocking
- `go vet` must pass before every commit

### Commands

- `make build` — compile dotctl binary
- `make test` — run Go tests
- `make lint` — go vet
- `make install` — copy dotctl to ~/.local/bin/

## Rules

- **NEVER commit secrets** — no API keys, tokens, passwords, SSH keys, cloud credentials.
  gitleaks pre-commit hook enforces this. If it blocks, fix the issue — don't bypass.
- **Template data is local** — `.chezmoi.toml.tmpl` defines prompts, answers live in
  `~/.config/chezmoi/chezmoi.toml` on each machine. Never hardcode environment-specific values.
- **Test with `chezmoi diff`** before `chezmoi apply` — review what will change.
- **Pin versions** in bootstrap URLs when possible. `.chezmoiexternal.toml` uses
  `master.tar.gz` for oh-my-zsh (intentional — always latest).
