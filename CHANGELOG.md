# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-19

Initial release. Aurora DX, Distrobox, and AI Sandbox verified from scratch.
WSL2 templates are functional but untested from a clean machine.

### Core

- chezmoi-managed dotfiles for WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora),
  and Distrobox containers with `.chezmoiroot` pointing to `home/`
- Two-variable environment model: `platform` (auto-detected: `wsl`, `aurora`,
  `distrobox`) + `context` (user-selected: `personal`, `personal-<project>`,
  `work-<name>`)
- Interactive chezmoi init prompts with context-aware conditional logic
  (`.chezmoi.toml.tmpl`) — minimal prompts per environment combination
- Templated `.zshrc` with per-platform PATH setup, SSH agent configuration,
  NVM/Bun paths, IDE/browser forwarding, and context-specific aliases
- Templated `.gitconfig` with conditional includes for personal and work emails
- External dependency management (`.chezmoiexternal.toml`) for oh-my-zsh and
  zsh-autosuggestions
- Starship prompt configuration with Nerd Font symbols and Aurora deployment
  indicator
- Atuin shell history sync (self-hosted) with per-context account separation
- Ghostty and k9s configuration

### Bootstrap

- `run_once_before_bootstrap.sh.tmpl` — first-run setup script with platform
  detection, tool installation, and post-install instructions
- Platform-specific tool installation: brew (Aurora), apt-get (WSL/Distrobox)
- Context-specific tools: glab, ansible, kubectl (personal); op CLI, bun,
  Playwright (personal-project); terraform, AWS CLI, kubectl (work)
- NVM v0.40.1 with platform-correct paths (`~/.nvm` on WSL,
  `~/.config/nvm` on Distrobox)
- JetBrainsMono Nerd Font v3.3.0 for Distrobox containers
- Claude Code install for non-Aurora environments
- Gitleaks pre-commit hook auto-install via symlink

### Distrobox Automation

- Python-based container lifecycle scripts (`distrobox_setup.py`,
  `distrobox_lib.py`, `distrobox_cleanup.py`) invoked via `uv run`
- Two container definitions in `containers/distrobox.ini`: `work-eam`, `personal`
- Symlink-based chezmoi source (host repo linked into container — uncommitted
  changes apply immediately)
- Context-aware non-interactive mode with `--personal-email`/`--work-email` flags
- Path normalization for Aurora's `/var/home` symlink
- IDE forwarding aliases (`code`, `antigravity`, `agy`) via `distrobox-host-exec`
- Browser forwarding (`BROWSER=distrobox-host-exec xdg-open`) for OAuth flows
- Automated Ptyxis terminal profile creation and removal via dconf
- `distrobox_cleanup.py` for targeted container removal with optional home wipe

### Credential Seeding

- `setup-creds` script for Distrobox containers — seeds credentials from
  1Password via `distrobox-host-exec op` (host app, not container CLI)
- Claude Code plugin installation (3 marketplaces + 5 plugins) with
  no 1Password dependency — runs first to avoid blocking on auth failures
- Atuin login with backup/restore error handling (protects against re-encryption
  on failed login)
- GitLab PAT authentication for self-hosted instance
- Context7 MCP server registration with graceful 1Password fallback
- Manual step guidance for AWS SSO login and kubeconfig copy

### Claude Code Integration

- Claude config managed by separate `claude-config` repo, cloned and symlinked
  by bootstrap
- Auto-sync on shell start (`git pull --ff-only` in background, non-distrobox)
- Plugin marketplace declarations in `settings.json` (`extraKnownMarketplaces`)
- 3 marketplaces: claude-plugins-official, superpowers-marketplace, ui-ux-pro-max-skill
- 5 plugins: context7, playwright, superpowers, episodic-memory, ui-ux-pro-max

### AI Sandbox

- Podman-based project sandboxes with persistent homes at `~/.sandbox/<project>/`
- `Containerfile.sandbox-base`: Ubuntu 24.04, Node.js 24, Python 3, uv, Bun,
  podman, podman-compose, host-spawn, Starship, fzf, oh-my-zsh, Atuin,
  Claude Code, Playwright, JetBrainsMono Nerd Font
- `ai-sandbox` CLI (`bin/ai-sandbox`) with project lifecycle management:
  `--shell`, `--build`, `--destroy`, `--services`, `--list`
- Tiered credential access: default (code only), `--git` (deploy key),
  `--no-network` (maximum containment)
- 1Password SSH agent mounted from host for git clone/push
- Host-spawn IDE forwarding (`code`, `antigravity`, `agy`) with automatic
  path translation (container `/home/developer` → host `~/.sandbox/<project>/`)
- Browser forwarding via `host-open` script (`BROWSER` env var)
- D-Bus session bus mount for host-spawn Flatpak portal communication
- Claude config linked on first sandbox creation: read-only symlinks for
  CLAUDE.md, rules, hooks, skills, agents; writable copy of settings.json
- First-run automation in `.zshrc`:
  - Atuin auto-login via podman secrets (`atuin_password`, `atuin_key`)
  - Claude marketplace registration + 5 plugin install
  - Context7 MCP registration (when `context7_key` secret exists)
- Podman secrets for credential injection: `anthropic_key`, `gemini_key`,
  `context7_key`, `atuin_password`, `atuin_key`
- Rootless podman-in-podman (`fuse-overlayfs`, `slirp4netns`, `uidmap`,
  `--security-opt label=disable`, `--device /dev/fuse`)
- Compose service auto-start and network joining
- Automated Ptyxis terminal profile creation per project

### dotctl CLI

- Go CLI (`dotctl/`) for monitoring dotfiles status across machines
- Three modes: `dotctl collect` (push to OTel Collector via OTLP gRPC),
  `dotctl status` (query Prometheus + Loki), `dotctl status --live` (local)
- Collectors for chezmoi state, tool inventory, credential status,
  Claude config symlinks, container status
- Terminal dashboard with lipgloss table rendering
- Systemd service + timer units for periodic collection (`dotctl/deploy/`)
- 29 Go unit tests across collector, push, query, config, and display packages

### Testing

- Distrobox integration tests (`test_distrobox_integration.py`): 59 assertions
  across 2 containers (personal: 29, work-eam: 30) — full lifecycle:
  delete, create, bootstrap, verify, delete
- Sandbox integration tests (`test_sandbox_integration.py`): 51 assertions
  covering tools, shell experience, aliases, host integration, security
  isolation, persistence
- Signal handling for cleanup on interrupt

### Documentation

- Per-platform setup guides (`docs/setup/aurora.md`, `wsl2.md`, `distrobox.md`)
- Reference documentation: environment model, dotctl CLI, credentials,
  distrobox scripts
- Architecture documentation: dotctl design, infrastructure review

### Verified Platforms

- **Aurora DX** — full nuke + re-bootstrap from scratch
- **Distrobox** — 59/59 assertions pass across 2 containers (personal, work-eam)
- **AI Sandbox** — 51/51 assertions pass, host integration verified
  (SSH, IDE forwarding, browser, Atuin, Claude plugins)
- **WSL2** — templates present but untested from a clean machine

### Known Issues

- Bootstrap `chsh` check (WSL/Distrobox only) compares `$SHELL` at runtime,
  causing repeated sudo prompts on re-apply
- Starship install has no version pinning
- `ubuntu:24.04` image tag in `distrobox.ini` is mutable (no pinned digest)
- OTel endpoint IP (`10.10.30.22`) is hardcoded in zshrc template
- `ai-sandbox` alias in sandbox `.zshrc` is non-functional (dotfiles repo
  not mounted inside container)
- `verify_personal_project()` in distrobox tests is dead code (no
  personal-\<project\> container in test matrix)
- setup-creds internals untested (only file existence checked)
