# CLAUDE.md

## Project Overview

chezmoi-managed dotfiles repository for bootstrapping consistent dev environments across
WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora), and Distrobox containers. Supports
6 target environments with credential isolation and AI agent sandboxing.

**Design doc:** See `~/personal/homelab/docs/plans/2026-02-28-aurora-dx-migration-design.md`

## Repository Structure

```
dotfiles/
├── .chezmoi.toml.tmpl          # Interactive prompts (chezmoi init)
├── .chezmoiignore              # Per-environment file skipping
├── .chezmoiexternal.toml       # External deps (oh-my-zsh, plugins)
├── home/                       # chezmoi source dir → maps to ~/
│   ├── dot_zshrc.tmpl          # Shell config (templated)
│   ├── dot_gitconfig.tmpl      # Git config (conditional includes)
│   └── dot_config/             # ~/.config/ files
├── scripts/                    # Setup automation
├── containers/                 # Distrobox + Podman definitions
├── bin/                        # CLI tools (ai-sandbox)
└── hooks/                      # Git hooks (gitleaks)
```

## chezmoi Conventions

- `dot_` prefix → `.` in target (e.g., `dot_zshrc` → `.zshrc`)
- `private_dot_` prefix → `.` with 0600 permissions
- `.tmpl` suffix → Go text/template processing
- `run_once_before_` prefix → script that runs once on first apply, before files
- Template data defined in `.chezmoi.toml.tmpl`, stored locally in `~/.config/chezmoi/chezmoi.toml`

## Environment Matrix

| Environment | Platform | Work creds | Homelab creds | Atuin account |
|---|---|---|---|---|
| wsl-work | WSL2 Ubuntu | yes | yes | rommel-personal |
| wsl-gaming | WSL2 Ubuntu | yes | yes | rommel-personal |
| aurora | Aurora DX host | no | no | none |
| distrobox-work | Ubuntu container | yes | no | rommel-work |
| distrobox-personal | Ubuntu container | no | yes | rommel-personal |
| distrobox-sandbox | Ubuntu container | no | no | none |

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

## Rules

- **NEVER commit secrets** — no API keys, tokens, passwords, SSH keys, cloud credentials.
  gitleaks pre-commit hook enforces this. If it blocks, fix the issue — don't bypass.
- **Template data is local** — `.chezmoi.toml.tmpl` defines prompts, answers live in
  `~/.config/chezmoi/chezmoi.toml` on each machine. Never hardcode environment-specific values.
- **Test with `chezmoi diff`** before `chezmoi apply` — review what will change.
- **Pin versions** in `.chezmoiexternal.toml` URLs when possible.
