# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-03-24

WSL2 automation brought to parity with Distrobox. Credential seeding unified
across all three platforms. Multiple UX fixes for chezmoi apply hangs and
shell noise.

### Added

- **WSL2 one-command setup** (`scripts/wsl_setup.py`) — mirrors the Distrobox
  workflow: validates prerequisites, clones repos, installs chezmoi, writes
  non-interactive config, runs `chezmoi init --apply`, and seeds credentials.
  Supports `--skip-creds`, `--personal-email`, `--work-email` flags
- **WSL `op` wrapper** (`~/.local/bin/op`) — forwards `op` calls to Windows
  `op.exe` for biometric unlock via the 1Password desktop app. Eliminates the
  need for native Linux `op` in WSL
- **WSL credential seeding** (`setup-wsl-creds`) — WSL-specific script for
  Claude plugins, Context7 MCP, Atuin, GitLab, and GitHub CLI. Later unified
  into `setup-creds` (see Refactored)
- **`dotup` alias** — universal one-command shorthand:
  `chezmoi update -v --no-pager --force && exec zsh`. Documented in all three
  platform setup guides
- **Go auto-install on WSL** — downloads official Go tarball (1.26+) since the
  Ubuntu PPA only has 1.24. Includes version comparison to skip if adequate
- **dotctl auto-build** — bootstrap builds dotctl from source on Aurora and WSL
  if Go is in PATH and `~/.local/bin/dotctl` doesn't exist. Skipped on Distrobox
- **`--skip-creds` flag on `distrobox_setup.py`** — skip credential seeding
  when 1Password isn't unlocked, matching `wsl_setup.py` parity
- **dotctl credential collectors** (`dotctl/internal/collector/creds.go`) —
  `DetectSSHAgent()`, `DetectSetupCreds()`, `DetectAtuinSync()` functions.
  These were missing from v0.1.0 due to a `.gitignore` pattern matching the
  original `credentials.go` filename
- **Bootstrap: jq** installed on apt-based platforms (WSL, Distrobox) for
  Claude Code statusline JSON parsing
- **Bootstrap: GitHub CLI (`gh`)** auto-installed on WSL via apt
- **Bootstrap: `uv`** auto-installed on WSL and Distrobox via curl
- **Bootstrap: gitleaks** auto-installed on WSL via GitHub release tarball
  (previously Aurora-only via brew)
- **Go PATH** added to `.zshrc` on WSL (`/usr/local/go/bin`)
- **chezmoi PATH symlink** — bootstrap ensures chezmoi is at `~/.local/bin/`
  regardless of where `get.chezmoi.io` installs it

### Changed

- **`personal` alias is now universal** — `alias personal='cd ~/personal'` is
  no longer gated to personal/personal-* contexts since `~/personal/` exists on
  all environments
- **`--no-pager --force` on `dotup`** — prevents diff pager hangs and oh-my-zsh
  cache overwrite prompts when syncing dotfiles
- **`--no-pager --force` on Distrobox chezmoi apply** — same fix applied to
  `distrobox_lib.py` bootstrap, preventing hangs in non-interactive container
  setup
- **Bootstrap post-install messages simplified** — all non-Distrobox platforms
  now print `setup-creds` instead of listing individual plugin commands
- **Aurora docs overhauled** — added Antigravity IDE install step, "What
  bootstrap installs automatically" summary table, credential seeding via
  `setup-creds`, `--no-pager --force` on re-apply, chezmoi init re-run guidance,
  dotctl auto-build note, `--skip-creds` on distrobox setup
- **Distrobox docs updated** — added note that `dotup` is unnecessary inside
  containers (symlinked source), `--no-pager --force` on re-apply, chezmoi init
  re-run guidance, `--skip-creds` documentation
- **WSL2 docs rewritten** — streamlined from 7 manual sections to 5 automated
  steps, prerequisites split into Windows-side vs WSL-side, added bootstrap
  summary table, added `make install-systemd` step

### Fixed

- **zsh job control noise** on new terminal tabs — background `git pull` for
  claude-config no longer registers in zsh's job table. Fixed by moving `&`
  inside the subshell: `(cmd &)` instead of `(cmd) &`
- **chezmoi apply hangs on WSL** — `--no-pager` prevents the diff viewer from
  blocking on `:` colon, `--force` auto-overwrites oh-my-zsh cache files
- **Atuin re-login on every `setup-creds` run** — login detection changed from
  matching the account name (`{{ .atuin_account }}`) to checking for any
  `^Username:` in `atuin status`. The old check never matched because the
  template variable differed from the server-side username
- **Missing dotctl credential functions** — `credentials.go` was silently
  excluded by `.gitignore`. Renamed to `creds.go` (test file renamed to match)
- **jq missing on WSL/Distrobox** — Claude Code statusline script requires jq
  for JSON parsing; now auto-installed during bootstrap

### Refactored

- **Unified `setup-creds`** — single script auto-detects platform
  (distrobox/WSL/Aurora) and sets the 1Password access method accordingly.
  Eliminates ~130 lines of duplicated logic from `setup-wsl-creds` (now a
  thin backwards-compatibility wrapper: `exec setup-creds "$@"`). Also adds
  GitHub CLI auth detection and Context7 MCP registration to all platforms
  (previously missing from Distrobox)
- **`setup-creds` now works on Aurora** — previously Aurora had no credential
  seeding script. Users had to manually run 8+ CLI commands for plugins, MCP,
  and auth

### Verified Platforms

- **Aurora DX** — `chezmoi apply`, `setup-creds`, `dotup` alias verified
- **Distrobox** — both `personal` and `work-eam` containers deleted, recreated
  from scratch, `setup-creds` run successfully (plugins, Context7, Atuin, GitLab)
- **AI Sandbox** — `fintrack` project created, tools verified (8/8), SSH agent
  connected via 1Password socket, git clone via SSH confirmed, destroyed cleanly
- **WSL2** — `wsl_setup.py` run from clean instance, `work-eam` context
  bootstrapped end-to-end

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
