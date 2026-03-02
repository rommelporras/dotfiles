# Aurora DX: Chezmoi Apply Guide

Generated: 2026-03-02
Target machine: Aurora 43 "Stargazer" (Kinoite/KDE) — 0xwsh

---

## Pre-Apply State

### What exists on Aurora now

- **Shell**: zsh is installed (`/usr/bin/zsh`) but login shell is `/bin/bash`
- **~/.zshrc**: Stock Aurora default — `compinit` + `source bling.sh` (38 lines)
- **~/.config/starship.toml**: Partial config — nerd font symbols, missing all language modules
- **~/.gitconfig**: Does not exist
- **~/.config/git/**: Does not exist
- **~/.config/atuin/**: Does not exist (atuin installed via brew, not configured)
- **~/.config/ghostty/**: Does not exist
- **~/.oh-my-zsh/**: Does not exist
- **~/.claude/**: Exists with runtime data, settings.json is minimal (`skipDangerousModePermissionPrompt` only)
- **Chezmoi**: Installed via brew (`v2.69.4`), never initialized on this machine

### Tools already installed (no action needed)

| Tool | Path | Source |
|---|---|---|
| zsh | /usr/bin/zsh | System RPM |
| starship | /usr/bin/starship | System RPM |
| fzf | /usr/bin/fzf (v0.67.0) | System RPM |
| atuin | /home/linuxbrew/.linuxbrew/bin/atuin | Homebrew |
| eza | /home/linuxbrew/.linuxbrew/bin/eza | Homebrew |
| bat | /home/linuxbrew/.linuxbrew/bin/bat | Homebrew |
| zoxide | /home/linuxbrew/.linuxbrew/bin/zoxide | Homebrew |
| direnv | /home/linuxbrew/.linuxbrew/bin/direnv | Homebrew |
| mise | /home/linuxbrew/.linuxbrew/bin/mise | Homebrew |
| chezmoi | /home/linuxbrew/.linuxbrew/bin/chezmoi | Homebrew |

---

## What Chezmoi Will Do

### Bootstrap script (`run_once_before_bootstrap.sh`)

On Aurora, most tool installs are guarded and will skip:

| Action | Will run? | Reason |
|---|---|---|
| Install zsh | Skip | Already at `/usr/bin/zsh` |
| Set zsh as login shell (`chsh`) | **YES** | Current shell is `/bin/bash` |
| Install starship | Skip | Already at `/usr/bin/starship` |
| Install JetBrainsMono Nerd Font | **YES** | Not currently installed |
| Install atuin | Skip | atuin_account = "none" for aurora |
| Install FZF via git | Skip | Aurora guard added (system FZF exists) |
| Install xclip | Skip | Aurora guard exists |
| WSL-only tools (nvm, bun) | Skip | Not WSL |
| Work tools (terraform) | Skip | Not work environment |
| Personal tools (glab, ansible) | Skip | Not personal environment |
| Create ~/.ssh with 700 perms | **YES** | Standard setup |

### File changes

| Target path | Action | Source |
|---|---|---|
| `~/.zshrc` | **REPLACE** | `dot_zshrc.tmpl` (Aurora-aware: bling.sh, history, starship guard) |
| `~/.config/starship.toml` | **REPLACE** | `dot_config/starship.toml` (full ys layout, 30+ modules, rpm-ostree indicator) |
| `~/.gitconfig` | **CREATE** | `dot_gitconfig.tmpl` (conditional identity, conventional commit alias) |
| `~/.config/git/ignore` | **CREATE** | Global gitignore |
| `~/.config/git/personal.gitconfig` | **CREATE** | Personal email for ~/personal/ projects |
| `~/.config/git/work.gitconfig` | **CREATE** | Work email (inactive on aurora — no ~/eam/) |
| `~/.config/atuin/config.toml` | **CREATE** | Atuin config (sync disabled, fuzzy search, compact style) |
| `~/.config/ghostty/config` | **CREATE** | Ghostty terminal (JetBrains Mono, catppuccin-mocha theme) |
| `~/.config/k9s/config.yaml` | **CREATE** | K9s with icons enabled |
| `~/.oh-my-zsh/` | **CREATE** | Downloaded from GitHub (oh-my-zsh + zsh-autosuggestions) |
| `~/.claude/CLAUDE.md` | **CREATE** | Global Claude Code instructions |
| `~/.claude/settings.json` | **REPLACE** | Full config with deny permissions, hooks, plugins |
| `~/.claude/hooks/` | **CREATE** | 4 security hooks (bash-write-protect, notify, protect-sensitive, scan-secrets) |
| `~/.claude/agents/code-reviewer.md` | **CREATE** | Code review agent |
| `~/.claude/skills/` | **CREATE** | commit, explain-code, push skills |

### What gets LOST from current Aurora config

These Aurora defaults are replaced, but their functionality is preserved:

| Lost default | Preserved by |
|---|---|
| Stock `~/.zshrc` bling.sh sourcing | Chezmoi template sources bling.sh on Aurora |
| Aurora starship.toml nerd font symbols | Chezmoi starship.toml has all the same symbols + more |
| Aurora starship.toml `custom.update` module | Chezmoi starship.toml includes the same module |
| `~/.claude/settings.json` minimal config | Replaced with full security config (superset) |

---

## Fixed Issues (this commit)

These critical Aurora-specific bugs were fixed before this guide was created:

### 1. Added HISTFILE/SAVEHIST (all environments)
```zsh
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
```
**Before**: Zsh history lost on every shell exit (HISTFILE was empty).

### 2. Guarded starship init on Aurora
```
{{ if ne .environment "aurora" }}
command -v starship &>/dev/null && eval "$(starship init zsh)"
{{ end }}
```
**Before**: Triple starship init (system `/etc/zsh/zshrc` + bling.sh + chezmoi `.zshrc`).
**After**: System and bling handle it; chezmoi skips.

### 3. Added bling.sh sourcing on Aurora
```
{{ if eq .environment "aurora" }}
test -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh
{{ end }}
```
**Before**: Replacing `.zshrc` killed eza/bat/direnv/zoxide/mise/ugrep integrations.
**After**: bling.sh sourced after Oh-My-Zsh, preserving all tool integrations.

### 4. Fixed FZF path for Aurora
```
{{ if eq .environment "aurora" }}
[ -f /usr/share/fzf/shell/key-bindings.zsh ] && source /usr/share/fzf/shell/key-bindings.zsh
{{ else }}
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh
{{ end }}
```
**Before**: `~/.fzf.zsh` doesn't exist on Aurora; FZF keybindings silently broken.
**After**: Uses system FZF shell integration at `/usr/share/fzf/shell/`.

Also guarded bootstrap FZF git clone to skip on Aurora (system RPM FZF already present).

### 5. Guarded `ll` alias on Aurora
```
{{ if ne .environment "aurora" }}
alias ll="ls -ltra"
{{ end }}
```
**Before**: Hardcoded `ls -ltra` overwrites bling.sh's `eza -l --icons=auto --group-directories-first`.
**After**: Aurora uses eza from bling; other environments keep the ls alias.

### 6. Enabled Claude Code config deployment on Aurora
Removed `.claude/` from `.chezmoiignore` for aurora environment. Claude Code is actively used on the Aurora host, so security hooks, skills, and agents should be deployed.

---

## Shell Load Order After Apply (Aurora)

```
1. /etc/zsh/zshenv          → sets skip_global_compinit=1
2. /etc/zsh/zprofile         → (login shells only)
3. ~/.zprofile               → (login shells only, not managed by chezmoi)
4. /etc/zsh/zshrc            → Prezto-cached compinit, brew shellenv, starship init
5. ~/.zshrc (chezmoi)        → PATH, HISTFILE, Oh-My-Zsh (38 plugins), atuin (disabled),
                                FZF system keybindings, SSH agent guard,
                                bling.sh (eza, bat, direnv, zoxide, mise, ugrep, starship*),
                                aliases
```

\*bling.sh has a `BLING_SOURCED` guard, so starship init from bling is a no-op since `/etc/zsh/zshrc` already ran it.

**Note**: Oh-My-Zsh will run its own `compinit` after the system Prezto-cached one. This adds ~100ms startup. Acceptable trade-off for 38 plugins. Can optimize later with `DISABLE_COMPFIX=true` if needed.

---

## Apply Commands

Run in a terminal on the Aurora machine:

```bash
# Step 1: Initialize chezmoi with the local repo
chezmoi init ~/personal/dotfiles

# Prompts (answer these):
#   environment  → aurora
#   personal_email → <your personal email>
#   work_email → <your work email>
#   has_work_creds → false
#   has_homelab_creds → false
#   atuin_sync_address → (leave empty)
#   atuin_account → none

# Step 2: Preview all changes (review carefully)
chezmoi diff

# Step 3: Apply
chezmoi apply -v

# Step 4: Restart shell
exec zsh

# Step 5: Verify
echo $HISTFILE           # Should be ~/.zsh_history
echo $SAVEHIST           # Should be 50000
type ll                  # Should be eza alias from bling
type z                   # Should be zoxide function
starship --version       # Should work
fc-list | grep -i "JetBrains Mono"  # Should show Nerd Font
ls ~/.oh-my-zsh          # Should exist
ls ~/.claude/hooks/      # Should show 4 hook scripts
cat ~/.claude/settings.json | head -5  # Should show deny permissions
```

---

## Post-Apply Tasks

```bash
# Set zsh as login shell (if bootstrap didn't prompt for password)
chsh -s /bin/zsh

# Set Ghostty font (if using Ghostty terminal)
# The config is deployed but Ghostty must be configured to use it

# Verify git identity
cd ~/personal/dotfiles && git config user.email
# Should show your personal email
```

---

## Rollback

If anything goes wrong:

```bash
# Restore original Aurora zshrc
cp ~/personal/dotfiles/docs/aurora-audit/defaults/dot_zshrc.aurora-default ~/.zshrc

# Restore original starship config
cp ~/personal/dotfiles/docs/aurora-audit/defaults/starship.toml.aurora-default ~/.config/starship.toml

# Remove chezmoi state to start fresh
rm -rf ~/.local/share/chezmoi ~/.config/chezmoi

# Restart shell
exec zsh
```

---

## Known Remaining Items (not blockers)

| Item | Severity | Notes |
|---|---|---|
| Double compinit (Prezto + OMZ) | Low | ~100ms startup overhead, acceptable |
| Dual starship binary (RPM + brew) | Low | Same version, just wastes disk — not a chezmoi issue |
| SSH agent block redundant on Aurora | Low | Guard works correctly, just unnecessary code |
| Private IP 10.10.30.22 hardcoded in zshrc | Low | Only renders in personal/homelab context (not aurora) |
| Login shell change needs password | Low | `chsh` will prompt for password on Aurora |
