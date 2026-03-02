# Chezmoi Dotfiles vs Aurora Live System — Verification Report

Generated: 2026-03-02
System: Aurora 43 "Stargazer" (Kinoite/KDE)

---

## Files Backed Up

### User configs (will be overwritten by `chezmoi apply`)

| Backup file | Original path | Will chezmoi touch it? |
|---|---|---|
| `dot_zshrc.aurora-default` | `~/.zshrc` | **Yes** — replaced by `dot_zshrc.tmpl` |
| `.config/starship.toml.aurora-default` | `~/.config/starship.toml` | **Yes** — replaced by `dot_config/starship.toml` |
| `dot_bashrc.aurora-default` | `~/.bashrc` | **No** — chezmoi doesn't manage .bashrc |
| `dot_zprofile.aurora-default` | `~/.zprofile` | **No** — chezmoi doesn't manage .zprofile |
| `.claude/settings.json.aurora-default` | `~/.claude/settings.json` | **No** — `.chezmoiignore` skips `.claude/` on aurora |

### System configs (immutable, reference only)

| Backup file | Original path |
|---|---|
| `etc-zsh-zshrc` | `/etc/zsh/zshrc` |
| `etc-zsh-zshenv` | `/etc/zsh/zshenv` |
| `etc-zsh-zprofile` | `/etc/zsh/zprofile` |
| `bling.sh` | `/usr/share/ublue-os/bling/bling.sh` |
| `profile.d-brew.sh` | `/etc/profile.d/brew.sh` |
| `profile.d-90-aurora-starship.sh` | `/etc/profile.d/90-aurora-starship.sh` |

### Files chezmoi will CREATE (no backup needed — nothing exists)

| Target path | chezmoi source |
|---|---|
| `~/.gitconfig` | `dot_gitconfig.tmpl` |
| `~/.config/atuin/config.toml` | `dot_config/atuin/config.toml.tmpl` |
| `~/.config/ghostty/config` | `dot_config/ghostty/config` |
| `~/.config/git/ignore` | `dot_config/git/ignore` |
| `~/.config/git/personal.gitconfig` | `dot_config/git/personal.gitconfig.tmpl` |
| `~/.config/git/work.gitconfig` | `dot_config/git/work.gitconfig.tmpl` |
| `~/.config/k9s/config.yaml` | `dot_config/k9s/config.yaml` |
| `~/.oh-my-zsh/` | `.chezmoiexternal.toml` (GitHub archive) |

---

## Accuracy Issues Found

### CRITICAL — Will break or misbehave on Aurora

#### 1. No bling.sh sourcing — eza/bat/direnv/zoxide/mise all lost

**Problem:** The stock `~/.zshrc` sources bling.sh at the bottom:
```zsh
test -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh
```
Chezmoi's `dot_zshrc.tmpl` replaces `~/.zshrc` entirely and does NOT source bling.sh. This means on Aurora after `chezmoi apply`:
- No `eza` aliases (`ll`, `ls`, `l.`, `l1`) — falls back to basic `ls`
- No `bat` alias (`cat`) — falls back to real cat
- No `ugrep` aliases (`grep`) — falls back to system grep
- No `direnv hook` — direnv won't activate in directories
- No `zoxide init` — `z` command won't work
- No `mise activate` — mise-managed tool versions won't activate

**Fix needed:** Add Aurora-only bling.sh sourcing block to `dot_zshrc.tmpl`.

#### 2. No HISTFILE or SAVEHIST — zsh history not persisted

**Problem:** Neither the system `/etc/zsh/zshrc` nor chezmoi's `dot_zshrc.tmpl` sets `HISTFILE` or `SAVEHIST`. The system sets history *options* (`INC_APPEND_HISTORY`, etc.) but without these variables, zsh holds 1000 lines in memory and throws them away on shell exit.

**Current state on this machine:**
```
HISTFILE=       (empty)
HISTSIZE=1000   (default)
SAVEHIST=       (empty)
```

**Fix needed:** Add to `dot_zshrc.tmpl` (ALL environments):
```zsh
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
```

#### 3. Starship double-init on Aurora

**Problem:** System `/etc/zsh/zshrc` already runs `eval "$(starship init zsh)"`. Chezmoi's `dot_zshrc.tmpl` does it again. Results in duplicate prompt hooks and ~50ms extra startup time.

**Fix needed:** Guard starship init with `{{ if ne .environment "aurora" }}`.

#### 4. FZF integration mismatch on Aurora

**Problem:** FZF is system-installed at `/usr/bin/fzf` (RPM package `fzf-0.67.0`). Shell integration files are at `/usr/share/fzf/shell/key-bindings.zsh`. However:
- Chezmoi's `.zshrc` sources `~/.fzf.zsh` — this file does NOT exist (created only by git-installed FZF)
- Bootstrap script checks `[ ! -d "$HOME/.fzf" ]` and would clone FZF from git — redundant on Aurora
- The `[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh` line silently no-ops (not harmful, but FZF keybindings won't work)

**System FZF shell files:**
```
/usr/share/fzf/shell/key-bindings.bash
/usr/share/fzf/shell/key-bindings.fish
/usr/share/fzf/shell/key-bindings.zsh
```
Note: no `completion.zsh` — only key-bindings.

**Fix options:**
- A: Template-guard FZF sourcing to use system path on Aurora, `~/.fzf.zsh` elsewhere
- B: Let bootstrap install git FZF on Aurora too (wasteful but consistent)

#### 5. Bootstrap uses `apt-get` — fails on Aurora

**Problem:** The bootstrap script uses `sudo apt-get` for several installs (zsh, xclip, terraform, ansible, glab). On Aurora (immutable Fedora with rpm-ostree), apt-get doesn't exist.

**Actual impact on Aurora:** Low, because:
- zsh → `/usr/bin/zsh` exists → `command -v zsh` succeeds → skipped ✓
- starship → `/usr/bin/starship` exists → skipped ✓
- xclip → `$IS_AURORA = true` → skipped ✓
- terraform/ansible/glab → only for distrobox/WSL environments → skipped ✓
- FZF → see issue #4 above — would install redundantly via git (not apt-get) ✓

So the apt-get calls are guarded, but it's worth adding explicit Aurora skips for clarity.

---

### MODERATE — Suboptimal behavior

#### 6. `alias ll="ls -ltra"` conflicts with eza

**Problem:** Chezmoi defines `alias ll="ls -ltra"`. On Aurora, `eza` is installed via brew. If bling.sh were sourced, bling would set `ll='eza -l --icons=auto --group-directories-first'` (better). But even with bling.sh fix (#1), the order matters — if chezmoi's alias comes after bling, it overwrites the eza version.

**Fix needed:** Remove hardcoded `ll` alias. Let bling handle it on Aurora, and provide eza-aware fallback on other environments.

#### 7. SSH agent block redundant on Aurora

**Problem:** Chezmoi's `.zshrc` starts ssh-agent if `$SSH_AUTH_SOCK` is unset. On Aurora, systemd already provides `SSH_AUTH_SOCK=/run/user/1000/ssh-agent.socket`. The guard works (block won't trigger), but it's unnecessary code on Aurora.

**Actual impact:** None — the `-z "$SSH_AUTH_SOCK"` check correctly skips it. Low priority fix.

#### 8. Dual starship installs

**Problem:** Starship exists at both:
- `/usr/bin/starship` — system RPM (v1.24.2)
- `/home/linuxbrew/.linuxbrew/bin/starship` — brew (v1.24.2)

Same version, different sources. Brew version is from `cli.Brewfile`. Not harmful but wastes disk.

**Fix:** Not a chezmoi issue — this is an Aurora/brew overlap.

#### 9. Oh-My-Zsh double compinit on Aurora

**Problem:** System `/etc/zsh/zshrc` runs `compinit` with Prezto-style 20-hour cache. Then Oh-My-Zsh runs `compinit` again with its own system. Double compinit adds ~100ms to shell startup.

**Mitigation:** Oh-My-Zsh respects `DISABLE_COMPFIX=true` and the system zshenv already sets `skip_global_compinit=1`. This might reduce the impact. Test with `zsh -x` to profile.

#### 10. Nerd Fonts not installed — k9s icons will break

**Problem:** `k9s/config.yaml` has `noIcons: false` (icons enabled). K9s icons require a Nerd Font. JetBrainsMono Nerd Font is NOT currently installed on this machine:
- Not in `~/.local/share/fonts/`
- Not via brew fonts.Brewfile (not yet run with `ujust aurora-fonts`)
- Not found by `fc-list`

Bootstrap script would install it from GitHub releases on non-WSL systems, so this resolves after bootstrap runs. But if only `chezmoi apply` is run without bootstrap, icons will render as boxes.

---

### LOW — Minor/cosmetic

#### 11. Starship config loses Aurora-specific `custom.update` module

**Problem:** Aurora's default `~/.config/starship.toml` includes an rpm-ostree update indicator:
```toml
[custom.update]
when = "/usr/bin/rpm-ostree status --json | /usr/bin/jq -e \".deployments[0].booted == false\" > /dev/null;"
command = 'echo "New deployment staged"'
```
Chezmoi's `starship.toml` replaces this entirely with a different layout. The update indicator is lost.

**Impact:** Cosmetic — user won't see "New deployment staged" in prompt on Aurora. The module gracefully no-ops on non-ostree systems.

**Fix option:** Add the `custom.update` module to chezmoi's starship.toml (it's harmless on non-Aurora systems since `rpm-ostree` won't exist).

#### 12. Login shell is still bash

**Problem:** `getent passwd` shows login shell is `/bin/bash`, not `/bin/zsh`. The bootstrap script would fix this with `chsh`, but on Aurora, `chsh` may require special handling (immutable OS).

**Verify:** `chsh -s /bin/zsh` should work on Aurora since it modifies `/etc/passwd` which is writable (it's in the persistent layer, not the ostree layer).

#### 13. `.chezmoiignore` skips `.claude/` on aurora

**Observation:** The ignore file skips all of `.claude/` on aurora environment. This means Claude Code on Aurora won't get:
- Custom hooks (notify.sh, protect-sensitive.sh, scan-secrets.sh, bash-write-protect.sh)
- Custom skills (commit, explain-code, push)
- Custom agents (code-reviewer)
- Global CLAUDE.md rules
- Full settings.json (deny permissions, hooks config)

The existing `~/.claude/settings.json` has only `skipDangerousModePermissionPrompt: true` — none of the security hooks or deny rules.

**Is this intentional?** If Claude Code is only used inside Distrobox containers on Aurora, this makes sense. If Claude Code is also used on the Aurora host, the security hooks should be deployed.

---

## Tool Availability Matrix (Live vs Expected)

| Tool | Expected source | Actually installed? | Path | Version |
|---|---|---|---|---|
| zsh | System RPM | ✅ | `/usr/bin/zsh` | 5.9 |
| starship | System RPM + brew | ✅ (both) | `/usr/bin/starship`, `/home/linuxbrew/.linuxbrew/bin/starship` | 1.24.2 |
| brew | System service | ✅ | `/home/linuxbrew/.linuxbrew/bin/brew` | 5.0.15 |
| atuin | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/atuin` | — |
| bat | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/bat` | — |
| eza | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/eza` | — |
| zoxide | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/zoxide` | — |
| direnv | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/direnv` | — |
| mise | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/mise` | — |
| chezmoi | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/chezmoi` | — |
| ugrep | brew (cli.Brewfile) | ✅ | `/home/linuxbrew/.linuxbrew/bin/ug` | — |
| fzf | System RPM | ✅ | `/usr/bin/fzf` | 0.67.0 |
| JetBrainsMono Nerd Font | bootstrap / ujust aurora-fonts | ❌ Not installed | — | — |
| 1Password | Manual install | ❌ Not installed | — | — |
| Oh-My-Zsh | .chezmoiexternal.toml | ❌ Not installed (chezmoi not yet applied) | — | — |
