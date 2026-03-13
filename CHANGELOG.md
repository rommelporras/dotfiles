# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-03-12

### Added

#### Core

- chezmoi-managed dotfiles for WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora),
  and Distrobox containers with `.chezmoiroot` pointing to `home/`
- Two-variable environment model: `platform` (auto-detected: `wsl`, `aurora`,
  `distrobox`) + `context` (user-selected: `personal`, `personal-<project>`,
  `work-<name>`, `gaming`, `sandbox`)
- Interactive chezmoi init prompts with context-aware conditional logic
  (`.chezmoi.toml.tmpl`) — minimal prompts per environment combination
- Templated `.zshrc` with per-platform PATH setup, SSH agent configuration,
  NVM/Bun paths, and context-specific aliases
- Templated `.gitconfig` with conditional includes for personal and work emails
- External dependency management (`.chezmoiexternal.toml`) for oh-my-zsh and
  zsh-autosuggestions
- Starship prompt configuration with Nerd Font symbols and Aurora deployment
  indicator
- Atuin shell history sync (self-hosted) with per-context account separation
- Ghostty and k9s configuration

#### Bootstrap

- `run_once_before_bootstrap.sh.tmpl` — first-run setup script with platform
  detection, tool installation, and post-install instructions
- Platform-specific tool installation: brew (Aurora), apt-get (WSL/Distrobox)
- Context-specific tools: glab, ansible, kubectl (personal); op CLI, bun,
  Playwright (personal-project); terraform, AWS CLI, kubectl (work);
  minimal (sandbox)
- NVM v0.40.1 with platform-correct paths (`~/.nvm` on WSL,
  `~/.config/nvm` on Distrobox)
- JetBrainsMono Nerd Font v3.3.0 for Distrobox containers
- Claude Code install for non-Aurora, non-sandbox environments
- Gitleaks pre-commit hook auto-install via symlink

#### Distrobox Automation

- Python-based container lifecycle scripts (`scripts/distrobox_setup.py`,
  `scripts/distrobox_lib.py`) invoked via `uv run`
- Four container definitions in `containers/distrobox.ini`: `work-eam`,
  `personal`, `personal-fintrack`, `sandbox`
- Symlink-based chezmoi source (host repo linked into container — uncommitted
  changes apply immediately)
- Context-aware non-interactive mode with `--personal-email`/`--work-email` flags
- Path normalization for Aurora's `/var/home` symlink
- IDE forwarding aliases (`code`, `antigravity`, `agy`) via `distrobox-host-exec`

#### Credential Seeding

- `setup-creds` script for Distrobox containers — seeds credentials from
  1Password via `distrobox-host-exec op` (host app, not container CLI)
- Claude Code plugin installation (official + superpowers marketplaces) with
  no 1Password dependency — runs first to avoid blocking on auth failures
- Atuin login with backup/restore error handling (protects against re-encryption
  on failed login)
- GitLab PAT authentication for self-hosted instance
- Context7 MCP server registration with graceful 1Password fallback
- Manual step guidance for AWS SSO login and kubeconfig copy

#### Claude Code Integration

- Claude config managed by separate `claude-config` repo, cloned and symlinked
  by bootstrap
- Auto-sync on shell start (`git pull --ff-only` in background, non-distrobox)
- `update-claude` script for manual sync with distrobox-aware behavior
- Plugin marketplace declarations in `settings.json` (`extraKnownMarketplaces`)

#### AI Sandbox

- Podman-based AI sandbox container (`containers/Containerfile.ai-sandbox`) with
  Ubuntu 24.04, Node.js 24, Python 3, uv, Claude Code
- `ai-sandbox` CLI (`bin/ai-sandbox`) with tiered credential access:
  `--git` for push access, `--no-network` for maximum containment

#### dotctl CLI

- Go CLI (`dotctl/`) for monitoring dotfiles status across machines
- Three modes: `dotctl collect` (push to OTel Collector via OTLP gRPC),
  `dotctl status` (query Prometheus + Loki), `dotctl status --live` (local)
- Collectors for chezmoi state, tool inventory, and credential status
- Terminal dashboard with lipgloss table rendering
- Systemd service + timer units for periodic collection (`dotctl/deploy/`)
- 24 Go unit tests across collector, push, query, and display packages

#### Testing

- Integration test suite (`scripts/test_distrobox_integration.py`) with full
  lifecycle: delete, create, bootstrap, verify, delete
- 85 assertions across 4 container types (sandbox: 16, personal: 22,
  personal-fintrack: 24, work-eam: 23)
- Signal handling for cleanup on interrupt
- Context-aware verification with negative assertions for sandbox isolation

#### Documentation

- Per-platform setup guides (`docs/setup/aurora.md`, `wsl2.md`, `distrobox.md`)
- Reference documentation: environment model, dotctl CLI, credentials,
  distrobox scripts
- Architecture documentation: dotctl design, infrastructure review
- Concise README with quick-start links and environment matrix
- Windows git setup script for 1Password SSH agent (`scripts/windows-git-setup.ps1`)

### Known Issues

- Bootstrap `chsh` check compares `$SHELL` at runtime, causing repeated sudo
  prompts on re-apply (doesn't update until next login)
- glab version detection parses redirect URL — fragile if GitLab changes format
- Starship install has no version pinning
- `ubuntu:24.04` image tag in `distrobox.ini` is mutable (no pinned digest)
- Distrobox bootstrap subprocess doesn't capture stderr on failure
- Work-eam aliases contain hardcoded username in paths
- OTel endpoint IP is hardcoded in zshrc template
- `make setup-hooks` not documented in setup guides
