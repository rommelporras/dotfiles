# CLAUDE.md

## Project Overview

chezmoi-managed dotfiles repository for bootstrapping consistent dev environments across
WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora), and Distrobox containers. Uses a
two-variable model (`platform` + `context`) for scalable environment targeting with
credential isolation and AI agent sandboxing via Podman.

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
│   │   ├── executable_setup-creds.tmpl  # Credential + plugin seeding for distrobox
│   │   └── executable_update-claude     # Manual claude-config sync (git pull)
│   └── dot_config/                      # ~/.config/ files (atuin, ghostty, git, k9s, starship.toml)
├── dotctl/                              # Go CLI (module: github.com/rommelporras/dotfiles/dotctl)
│   ├── cmd/dotctl/main.go               # cobra CLI wiring
│   ├── internal/                        # collector, push, query, display, model, config
│   ├── deploy/                          # systemd service + timer units
│   ├── go.mod / go.sum
│   └── Makefile                         # build, test, lint, install, install-systemd
├── scripts/                             # Setup automation (Python, invoke via uv run)
│   ├── distrobox_setup.py               # Container creation + chezmoi bootstrap
│   ├── distrobox_cleanup.py             # Container removal + home wipe + Ptyxis cleanup
│   ├── distrobox_lib.py                 # Shared library for distrobox scripts
│   ├── test_distrobox_integration.py    # E2E test: delete → create → bootstrap → verify → delete
│   ├── test_sandbox_integration.py      # E2E test: tools, shell, security, persistence
│   └── windows-git-setup.ps1            # Windows git setup for WSL
├── docs/
│   ├── setup/                           # Per-platform setup guides (wsl2.md, aurora.md, distrobox.md)
│   ├── reference/                       # CLI ref, environment model, credentials, distrobox-scripts
│   └── architecture/                    # Design decisions (dotctl-design.md, infra.md)
├── containers/                          # Distrobox + Podman definitions
│   ├── distrobox.ini                    # Container definitions (work-eam, personal — 2 containers)
│   └── Containerfile.sandbox-base       # AI sandbox base image
├── bin/                                 # CLI tools (ai-sandbox)
├── Makefile                             # Root delegator → $(MAKE) -C dotctl <target>
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
- **`context`** (user-selected): `personal`, `personal-<project>`, `work-eam`, `work-<name>` — controls aliases, credentials, tools

| Platform | Context | SSH Agent | Key differences |
|---|---|---|---|
| wsl | personal | 1Password via npiperelay | NVM/Bun, personal creds, homelab access |
| wsl | work-eam | 1Password via npiperelay | NVM/Bun, work + personal creds |
| aurora | personal | 1Password native socket | Immutable OS, no chsh, Atuin sync |
| distrobox | work-eam | 1Password via absolute host path | AWS CLI, kubectl, Terraform, work creds |
| distrobox | personal | 1Password via absolute host path | kubectl, homelab kubeconfig, glab, Ansible |
| distrobox | personal-\<project\> | 1Password via absolute host path | Native op CLI, Bun, Playwright, glab, no homelab |

**Adding contexts:** For work: add container to `distrobox.ini`, add job-specific aliases
in `dot_zshrc.tmpl`, run `distrobox_setup.py work-acme`. Shared work tools apply via
`hasPrefix .context "work-"`. For personal projects: add container to `distrobox.ini`,
run `distrobox_setup.py personal-<project>`. Shared personal tools (glab, Bun, Playwright,
native `op` CLI) apply via `hasPrefix .context "personal-"`.

### Distrobox chezmoi workflow

`scripts/distrobox_setup.py` bootstraps containers with chezmoi (see
[docs/reference/distrobox-scripts.md](docs/reference/distrobox-scripts.md) for full reference):
1. Installs chezmoi inside the container (`~/bin/chezmoi`)
2. Symlinks `~/.local/share/chezmoi` → host repo (uncommitted changes apply immediately)
3. Writes chezmoi config (non-interactive when `--personal-email`/`--work-email` provided)
4. Runs `chezmoi init --apply`
5. Runs `setup-creds` to seed plugins, MCP, and credentials from 1Password

Inside containers, `$HOME` is `~/.distrobox/<context>/` (NOT the host home). Paths to
host resources (e.g. 1Password socket) must use absolute paths like `/home/<user>/...`.

To update dotfiles inside a container: `~/bin/chezmoi apply -v && exec zsh`

### IDE and browser forwarding

Distrobox containers alias `code`, `antigravity`, and `agy` to forward
to the Aurora host via `distrobox-host-exec`. `BROWSER` is set to
`distrobox-host-exec xdg-open` for OAuth flows. No IDEs needed in containers.

### Credential and plugin seeding

`setup-creds` (deployed to `~/.local/bin/`) uses `distrobox-host-exec op` to pull
secrets from 1Password on the host. Handles:
1. Claude Code plugin marketplace registration and plugin installation
2. Context7 MCP server registration (requires API key from 1Password)
3. Atuin login and sync (with backup/restore error handling)
4. GitLab auth (personal context)
5. Manual step instructions for kubeconfig and AWS

Plugins run before credentials so 1Password failures don't block plugin setup.

## AI Sandbox (Podman)

Isolated, project-scoped Podman containers for AI-assisted development. Each project
gets a persistent home at `~/.sandbox/<project>/` with no host filesystem access.

### Image

`Containerfile.sandbox-base`: Ubuntu 24.04, Node.js 24, Python 3, uv, Bun, podman,
podman-compose, host-spawn, Starship, fzf, oh-my-zsh, Atuin, Claude Code, Playwright,
JetBrainsMono Nerd Font. Built with `ai-sandbox <project> --build`.

### Host integration (mounted into container by ai-sandbox)

| Mount | Host path | Container path | Purpose |
|---|---|---|---|
| Persistent home | `~/.sandbox/<project>/` | `/home/developer` | Project files, tool config |
| 1Password SSH | `~/.1password/agent.sock` | `/run/1password/agent.sock` | Git clone/push |
| D-Bus session bus | `/run/user/1000/bus` | `/run/user/1000/bus` | host-spawn IDE forwarding |
| Claude config | `~/personal/claude-config` | `/run/claude-config` (read-only) | CLAUDE.md, rules, hooks, skills, agents |

`settings.json` is **copied** (not symlinked) because `claude plugin install` writes to it.

### Credential injection

Podman secrets (created once on host, always available):
- `atuin_password` / `atuin_key` → Atuin auto-login
- `anthropic_key` → `ANTHROPIC_API_KEY`
- `gemini_key` → `GEMINI_API_KEY`
- `context7_key` → `CONTEXT7_API_KEY` + Context7 MCP registration

Create with: `op read "op://..." | tr -d '\n' | podman secret create <name> -`

### First-run automation (.zshrc)

On first interactive shell, gated by marker files:
- Atuin login (`~/.local/share/atuin/key` marker)
- Claude marketplace add (3) + plugin install (5) (`~/.claude/.plugins-installed` marker)
- Context7 MCP registration (`~/.claude/.mcp-configured` marker)

### IDE/browser forwarding

Uses `host-spawn` via Flatpak D-Bus portal (not nsenter — rootless podman blocks
namespace access). Container paths translated to host paths via `SANDBOX_HOST_HOME` env var.

- `code .`, `antigravity`, `agy` → wrapper functions in .zshrc
- `BROWSER` → `~/.local/bin/host-open` (calls `host-spawn xdg-open`)

### Sandbox commands

```bash
ai-sandbox fintrack --shell              # interactive shell
ai-sandbox fintrack --build              # build/rebuild base image
ai-sandbox fintrack --destroy            # remove container + home + Ptyxis profile
ai-sandbox fintrack --services up        # start compose services
ai-sandbox fintrack claude -- --dangerously-skip-permissions  # run tool
ai-sandbox --list                        # list all projects
```

## Common Commands

```bash
# Preview changes without applying
chezmoi diff

# Apply changes
chezmoi apply -v

# Re-run bootstrap (won't re-run — it's run_once_)
chezmoi state delete-bucket --bucket=scriptState
chezmoi apply

# Bootstrap a distrobox container
uv run python scripts/distrobox_setup.py personal --personal-email you@example.com
uv run python scripts/distrobox_setup.py work-eam --work-email you@work.com

# Remove a distrobox container
uv run python scripts/distrobox_cleanup.py personal --wipe-home

# Run integration tests
uv run python scripts/test_distrobox_integration.py --all        # 59 assertions
uv run python scripts/test_sandbox_integration.py --skip-build   # 51 assertions
make test                                                         # 29 Go unit tests

# Install gitleaks pre-commit hook
make setup-hooks
```

## Claude Code Config

Claude Code global config (`~/.claude/`) is managed by a separate repo:
[claude-config](https://github.com/rommelporras/claude-config) — cloned to
`~/personal/claude-config` and symlinked into `~/.claude/`.

The bootstrap script handles cloning and symlinking. On distrobox containers,
symlinks point to the host's clone via absolute paths. The `.claude/` directory
is blanket-ignored in `.chezmoiignore` so chezmoi never touches it.

For AI sandbox: claude-config is mounted read-only at `/run/claude-config/`,
with symlinks created during home seeding and `settings.json` copied writable.

Plugin and MCP setup is handled by `setup-creds` (all platforms — auto-detects
distrobox/WSL/Aurora), first-run .zshrc automation (sandbox), or manual CLI commands.

## dotctl (Go CLI)

Go CLI tool living at `dotctl/` (module: `github.com/rommelporras/dotfiles/dotctl`).
chezmoi only sees `home/` via `.chezmoiroot` — Go files are invisible to it.

### Architecture

Single binary, three modes:
- `dotctl collect` — gather status, push to OTel Collector via OTLP gRPC
- `dotctl status` — query Prometheus + Loki, render terminal tables
- `dotctl status --live` — gather + display directly (no cluster dependency)

### Commands (run from repo root — root Makefile delegates to dotctl/)

- `make build` — compile dotctl binary (output: `dotctl/dotctl`)
- `make test` — run Go tests (29 tests)
- `make lint` — go vet
- `make install` — copy dotctl to ~/.local/bin/
- `make install-systemd` — install + enable systemd timer
- `make setup-hooks` — install gitleaks pre-commit hook (run once after cloning)

## Rules

- **NEVER commit secrets** — no API keys, tokens, passwords, SSH keys, cloud credentials.
  gitleaks pre-commit hook enforces this. If it blocks, fix the issue — don't bypass.
- **Template data is local** — `.chezmoi.toml.tmpl` defines prompts, answers live in
  `~/.config/chezmoi/chezmoi.toml` on each machine. Never hardcode environment-specific values.
- **Test with `chezmoi diff`** before `chezmoi apply` — review what will change.
- **Pin versions** in bootstrap URLs when possible. `.chezmoiexternal.toml` uses
  `master.tar.gz` for oh-my-zsh (intentional — always latest).
