# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
