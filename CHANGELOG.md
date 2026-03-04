# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
