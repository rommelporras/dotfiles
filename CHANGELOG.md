# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Non-interactive mode for distrobox setup — each context only needs its relevant email flag
- All config values (Atuin, credentials, op CLI) derived automatically from context name
- Sandbox containers always skip interactive prompts (no email flags needed)
- AWS CLI v2 auto-install for work-* distrobox containers
- kubectl auto-install for work-* and personal distrobox containers
- `apt-get upgrade` early in bootstrap to patch base packages on apt-based systems
- `docs/distrobox-scripts.md` — full parameter reference for setup and test scripts

### Changed

- Migrate distrobox scripts from shell to Python (`distrobox_setup.py`, `test_distrobox_integration.py`, `distrobox_lib.py`)
- Remove shell shims (`distrobox-setup.sh`, `test-distrobox-integration.sh`) — invoke via `uv run python`
- `full_config_for()` accepts optional email parameters (backward-compatible defaults for tests)
- `bootstrap_chezmoi()` has 15-minute timeout to prevent hangs
- `run_setup_creds()` returns exit code instead of raising on failure
- `_command_exists()` uses `shutil.which()` instead of broken shell subprocess
- Single-container setup creates only the requested container (not all via `distrobox assemble`)
- Update Atuin account names from legacy `rommel-*` to current naming convention
- `setup-creds` always uses `distrobox-host-exec op` (personal-* native op lacks desktop app integration for initial setup)
- Integration tests expanded to 77 assertions (from 64) — added kubectl, AWS CLI, Atuin config verification

### Fixed

- `$SHELL` showing `bash` inside distrobox containers (host's `$SHELL` leaked via distrobox; `.zshrc` now sets it correctly)
- Bootstrap `ln -sf` → `ln -sfn` for claude-config symlinks (prevents recursive symlinks when target is existing directory)
- `setup-creds` crash when Context7 MCP server already configured (`claude mcp add` non-zero exit with `set -e`)
- `setup-creds` Atuin error message showing wrong 1Password field for `personal-*` contexts
- `chezmoi.toml.tmpl` Atuin prompt hint updated from `rommel-personal/rommel-eam` to `personal/work-eam`

## [v1.4.0] - 2026-03-06

### Added

- `personal-<project>` distrobox pattern for project-scoped containers (e.g. `personal-fintrack`)
- Native 1Password CLI (`op`) with biometric unlock for personal-<project> containers
- Bun and Playwright (chromium) auto-install for personal-<project> containers
- glab CLI and GitLab auth for personal-<project> containers
- `has_op_cli` derived template variable in chezmoi config
- OTel telemetry and personal aliases shared across personal and personal-<project> contexts
- E2E integration test script (`scripts/test-distrobox-integration.sh`) with 64 assertions across 4 container types
- Testing section in README with usage examples

### Fixed

- Missing `mkdir -p ~/.local/share` in `distrobox-setup.sh` that broke symlink creation on fresh containers
- glab-cli `.chezmoiignore` condition now allows glab config for personal-<project> contexts
- `.kube/` directory correctly excluded for personal-<project> (no homelab access)

### Changed

- Migrate Claude Code config from chezmoi (`private_dot_claude/`) to separate [claude-config](https://github.com/rommelporras/claude-config) repo with symlinks
- Bootstrap clones claude-config repo and symlinks CLAUDE.md, settings.json, hooks, skills, agents into `~/.claude/`
- `setup-creds` installs plugins and Context7 MCP before credentials (1Password failures don't block plugin setup)
- Atuin login has backup/restore error handling to prevent state corruption on server failure
- Sandbox SSH_AUTH_SOCK forcefully unset (overrides distrobox host passthrough)
- Homelab aliases (invoicetron, kubectl-homelab) restricted to personal/gaming contexts only
- AI sandbox Containerfile installs uv and Claude Code as developer user
- Split glab and Ansible install conditions (glab shared, Ansible personal-only)
- Updated CLAUDE.md and README.md with personal-<project> documentation

## [v1.3.0] - 2026-03-05

### Added

- Node.js LTS install and Context7 MCP setup instructions in bootstrap
- Distrobox automation, Antigravity aliases, and credential seeding (`setup-creds`)
- SSH agent fix, IDE forwarding, and credential seeding for distrobox containers
- Claude Code project plugin settings (`.claude/settings.json`)
- Claude Code plugin dependencies design and implementation plan docs

### Fixed

- SSH socket path, conditional prompts, glab install, and distrobox script issues

### Changed

- Replace flat `environment` variable with `platform` + `context` two-variable model
- Add prompt hints and auto-set distrobox environment in chezmoi config

## [v1.2.0] - 2026-03-03

### Added

- Windows git setup script for 1Password SSH agent integration (`scripts/windows-git-setup.ps1`)

## [v1.1.1] - 2026-03-03

### Fixed

- Only install Nerd Font manually in Distrobox containers (avoids conflict with brew cask on immutable Fedora)

### Changed

- Restructure Aurora DX setup: devmode before 1Password (rebase resets layered packages)
- Add git clone step and devmode rebase warning to Aurora setup
- Separate WSL and Aurora chezmoi install paths in README
- Update font setup docs to reflect per-platform install method

## [v1.1.0] - 2026-03-02

### Added

- 1Password SSH agent setup and Aurora platform instructions in README
- Aurora DX audit docs — comparison, verification, and default config backups
- Nerd Font symbols and language module icons to Starship config
- `ujust devmode` and `ujust dx-group` steps to Aurora setup guide
- chezmoi init with local path docs for pre-installed setups

### Fixed

- 6 critical Aurora DX compatibility issues
- chezmoi init to use local path instead of symlink
- Distrobox script path in README

### Changed

- Clean up Aurora audit docs and update project docs
- Update environment matrix in README

## [v1.0.0] - 2026-02-28

### Added

- chezmoi dotfiles scaffold with `.chezmoiroot` pointing to `home/` source dir
- Templated `.zshrc` with 31 Oh My Zsh plugins and environment conditionals (WSL, Aurora, Distrobox)
- Templated `.gitconfig` with conditional includes (personal/work identity by directory)
- Conventional commit helper (`git cc feat "message"`)
- Starship cross-shell prompt with Nerd Font icons and Kubernetes awareness
- Atuin shell history sync (optional, self-hosted)
- K9s and Ghostty terminal configs
- Bootstrap script with platform detection and environment-specific tool installation
- gitleaks pre-commit hook for secret scanning
- Podman AI sandbox with tiered credential injection (`--git`, `--no-network`)
- Distrobox container definitions and setup script (work, personal, sandbox)
- Claude Code global config via chezmoi (`~/.claude/` — CLAUDE.md, settings, hooks, skills, agents)
- Security hooks: secret scanning, write protection, destructive command blocking
- Skills: `/commit`, `/push`, `/explain-code`
- Code reviewer agent with per-project memory
- Generic `/release` command (fallback for projects without their own)
- Support for 6 target environments (wsl-work, wsl-gaming, aurora, distrobox-work, distrobox-personal, distrobox-sandbox)
