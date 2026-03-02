# CLAUDE.md

## Project Overview

chezmoi-managed dotfiles repository for bootstrapping consistent dev environments across
WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora), and Distrobox containers. Supports
6 target environments with credential isolation and AI agent sandboxing.

## Repository Structure

```
dotfiles/
├── .chezmoiroot                         # Points chezmoi source to home/
├── home/                                # chezmoi source dir → maps to ~/
│   ├── .chezmoi.toml.tmpl               # Interactive prompts (chezmoi init)
│   ├── .chezmoiignore                   # Per-environment file skipping
│   ├── .chezmoiexternal.toml            # External deps (oh-my-zsh, plugins)
│   ├── dot_zshrc.tmpl                   # Shell config (templated per env)
│   ├── dot_gitconfig.tmpl               # Git config (conditional includes)
│   ├── run_once_before_bootstrap.sh.tmpl # First-run setup (installs tools)
│   ├── private_dot_claude/              # Claude Code global config (~/.claude/)
│   └── dot_config/                      # ~/.config/ files
├── scripts/                             # Setup automation
├── containers/                          # Distrobox + Podman definitions
├── bin/                                 # CLI tools (ai-sandbox)
└── hooks/                               # Git hooks (gitleaks)
```

## chezmoi Conventions

- `dot_` prefix → `.` in target (e.g., `dot_zshrc` → `.zshrc`)
- `private_dot_` prefix → `.` with owner-only permissions (0600 files, 0700 dirs)
- `.tmpl` suffix → Go text/template processing
- `run_once_before_` prefix → script that runs once on first apply, before files
- Template data defined in `.chezmoi.toml.tmpl`, stored locally in `~/.config/chezmoi/chezmoi.toml`

## Environment Matrix

| Environment | Platform | SSH Agent | Key differences |
|---|---|---|---|
| wsl-work | WSL2 Ubuntu | 1Password via npiperelay | Work + personal creds, NVM/Bun |
| wsl-gaming | WSL2 Ubuntu | 1Password via npiperelay | Work + personal creds, NVM/Bun |
| aurora | Aurora DX host | 1Password native socket | Immutable OS, bling.sh, no chsh, Atuin sync |
| distrobox-work | Ubuntu container | Inherited from host | Work AWS/EKS creds |
| distrobox-personal | Ubuntu container | Inherited from host | Homelab kubeconfig, glab |
| distrobox-sandbox | Ubuntu container | Fallback ssh-agent | No creds, no Claude config |

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

`home/private_dot_claude/` deploys to `~/.claude/` — Claude Code's global config.
Non-exact directory (chezmoi won't delete runtime files like `history.jsonl`, `projects/`, etc.).

- `CLAUDE.md.tmpl` — templated, conditional environment section (WSL, Aurora, Distrobox variants)
- `settings.json` — same everywhere (hooks use `$HOME` which resolves correctly)
- `hooks/` — executable_ prefix for chezmoi to set +x permissions
- `agents/`, `skills/` — plain files, no templating needed

## Rules

- **NEVER commit secrets** — no API keys, tokens, passwords, SSH keys, cloud credentials.
  gitleaks pre-commit hook enforces this. If it blocks, fix the issue — don't bypass.
- **Template data is local** — `.chezmoi.toml.tmpl` defines prompts, answers live in
  `~/.config/chezmoi/chezmoi.toml` on each machine. Never hardcode environment-specific values.
- **Test with `chezmoi diff`** before `chezmoi apply` — review what will change.
- **Pin versions** in bootstrap URLs when possible. `.chezmoiexternal.toml` uses
  `master.tar.gz` for oh-my-zsh (intentional — always latest).
