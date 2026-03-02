# Aurora DX vs Chezmoi Dotfiles — Full Comparison

Generated: 2026-03-02
System: Aurora 43 "Stargazer" (Kinoite/KDE), brew CLI tools via `ujust aurora-cli`

---

## Shell Load Order (Critical Context)

Understanding what runs when is essential for avoiding double-init and conflicts.

### Aurora Zsh load order (no chezmoi):

```
1. /etc/zsh/zshenv         → PATH ($HOME/.bin, $HOME/bin), SYSTEM var, skip_global_compinit
2. /etc/zsh/zprofile        → locale (en_US.UTF-8), fpath dedup
3. /etc/zsh/zshrc           → compinit, history opts, brew shellenv, starship init, source ~/.zsh/**/*.sh
4. ~/.zshrc (stock Aurora)  → compinit (again!), COMPLETE_IN_WORD, source bling.sh
5. bling.sh                 → eza/bat/ugrep aliases, direnv hook, starship init (THIRD time!), zoxide init, mise activate
```

### Aurora Zsh load order (with chezmoi dot_zshrc.tmpl):

```
1. /etc/zsh/zshenv         → PATH, SYSTEM var
2. /etc/zsh/zprofile        → locale, fpath
3. /etc/zsh/zshrc           → compinit, history opts, brew shellenv, STARSHIP INIT (#1), source ~/.zsh/**/*
4. ~/.zshrc (chezmoi)       → Oh-My-Zsh (compinit #2), STARSHIP INIT (#2), atuin init, aliases
                              bling.sh is NOT sourced → no eza/bat/ugrep aliases, no direnv/zoxide/mise
```

### WSL/Distrobox load order (with chezmoi):

```
1. ~/.zshrc (chezmoi)       → Oh-My-Zsh (compinit), starship init, atuin, FZF, aliases
                              No system zshrc, no bling, no brew — chezmoi owns everything
```

---

## 1. Starship

### What Aurora provides out of the box

- **System binary:** `/usr/bin/starship` (v1.24.2) — baked into the OS image
- **Bash init:** `/etc/profile.d/90-aurora-starship.sh` — only for bash, not zsh
- **Zsh init #1:** `/etc/zsh/zshrc` line: `eval "$(starship init zsh)"`
- **Zsh init #2:** `bling.sh` line: `eval "$(starship init "${BLING_SHELL}")"` (runs if bling is sourced)
- **User config:** `~/.config/starship.toml` exists — Aurora's Nerd Font preset with:
  - Nerd Font icons for 30+ languages/tools and 30+ OS symbols
  - `custom.update` module — shows "New deployment staged" when rpm-ostree has a pending update
  - `username` always shown (blue bold), `hostname` always shown (green bold)
  - `character`: `➜` (green) / `✗` (red)

### What our chezmoi starship.toml provides

- **ys-inspired format string** with explicit module ordering and `[@]` / `[ in ]` / `[ on ]` separator text
- Nerd Font icons (subset: git, k8s, aws, docker, node, python, ssh, directory lock)
- `kubernetes` module with `detect_env_vars = ['KUBECONFIG']` — only shows when KUBECONFIG is set
- `time` always shown (HH:MM:SS)
- `cmd_duration` with 2s threshold
- `character`: `❯` (green) / `❯` (red) — different symbol from Aurora default
- **No OS symbols block** — Aurora's has 30+ OS-specific Nerd Font icons
- **No `custom.update` module** — the ostree update indicator is Aurora-specific

### Analysis: Aurora-specific vs universal

| Feature | Aurora-specific? | Notes |
|---|---|---|
| `custom.update` (ostree check) | **Yes** | Only meaningful on rpm-ostree systems. Uses `/usr/bin/rpm-ostree` and `/usr/bin/jq`. |
| OS symbols (Nerd Font) | No, universal | Nice to have everywhere, but large config block. Optional. |
| 30+ language icons | No, universal | Aurora's config has more language icons than ours |
| Explicit format string | No, universal | Our ys-inspired layout is a design choice, works everywhere |
| `kubernetes` with KUBECONFIG detect | No, universal | Better than Aurora's default (always visible) |

### Double-init risk: **YES — currently 2x on Aurora**

When chezmoi deploys `dot_zshrc.tmpl` on Aurora:
1. `/etc/zsh/zshrc` runs `eval "$(starship init zsh)"` — init #1
2. `dot_zshrc.tmpl` runs `command -v starship && eval "$(starship init zsh)"` — init #2
3. bling.sh is NOT sourced (chezmoi replaces `~/.zshrc`), so no init #3

**Impact:** Double starship init causes duplicate prompt hooks. Starship is somewhat resilient to this (later init overwrites earlier), but it adds ~50ms startup overhead and is unnecessary.

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Deploy `~/.config/starship.toml` | **Yes** — overwrite Aurora default with our ys-inspired config | **Yes** — same config |
| Run `starship init zsh` in `.zshrc` | **No** — system zshrc already does it | **Yes** — only source of init |
| Add `custom.update` module | **Consider** — template-guard for Aurora only | **No** — not applicable |
| chezmoi template guard | `{{ if ne .environment "aurora" }}starship init{{ end }}` | Always init |

---

## 2. Atuin

### What Aurora provides out of the box

- **Binary:** `/home/linuxbrew/.linuxbrew/bin/atuin` (installed via brew cli.Brewfile)
- **Shell integration:** **Disabled by default** in bling.sh (commented out):
  ```sh
  # [ "$(command -v atuin)" ] && eval "$(atuin init "${BLING_SHELL}" ${ATUIN_INIT_FLAGS})"
  ```
- **Config file:** `~/.config/atuin/config.toml` — **does not exist**. No user config, no system config.
- **History sync:** Not configured. No account, no sync address.

### What our chezmoi provides

- **Config:** `dot_config/atuin/config.toml.tmpl` deploys to `~/.config/atuin/config.toml`
  ```toml
  sync_address = "https://atuin.k8s.rommelporras.com"  # (when atuin_sync_address is set)
  auto_sync = true
  search_mode = "fuzzy"
  filter_mode = "global"
  style = "compact"
  inline_height = 20
  show_preview = true
  ```
- **Shell init:** `dot_zshrc.tmpl` line (when `atuin_account != "none"`):
  ```zsh
  command -v atuin &>/dev/null && eval "$(atuin init zsh)"
  ```

### Discrepancy: planned vs actual config

The task description mentions `inline_height=30`, but the actual chezmoi template has `inline_height = 20`. Decide which is correct.

### Double-init risk: **No**

Bling.sh has atuin init **commented out**. Our `.zshrc` is the only source of `atuin init zsh`. No conflict.

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Deploy `~/.config/atuin/config.toml` | **Yes** — nothing exists | **Yes** — same config |
| Run `atuin init zsh` in `.zshrc` | **Yes** — bling doesn't do it | **Yes** — same |
| `sync_address` | Set via template var (homelab K8s) | Same (when atuin_account != none) |
| chezmoi template guard | None needed — no Aurora conflict | None needed |

---

## 3. Zsh Config

### What the system `/etc/zsh/zshrc` already provides on Aurora

This is the most complex area. The system zshrc provides a **full Prezto-inspired setup**:

#### Completion system (DO NOT duplicate on Aurora):
- `autoload -Uz compinit` with 20-hour cache (`$XDG_CACHE_HOME/prezto/zcompdump`)
- Brew completions fpath (`/home/linuxbrew/.linuxbrew/share/zsh/site-functions`)
- Case-insensitive completion matching
- Fuzzy completion with max-errors
- Colored completion lists (LS_COLORS)
- Menu selection for completions
- Group/describe completion formatting
- SSH/SCP/rsync host completion from known_hosts
- Kill process completion with colored PIDs
- History-word completion

#### Shell options (DO NOT duplicate on Aurora):
- `INC_APPEND_HISTORY`, `SHARE_HISTORY`, `APPEND_HISTORY` (**BUT** `HISTFILE` and `SAVEHIST` are never set — see below)
- `COMPLETE_IN_WORD`, `ALWAYS_TO_END`
- `AUTO_MENU`, `AUTO_LIST`, `AUTO_PARAM_SLASH`, `AUTO_CD`
- `EXTENDED_GLOB`, `PATH_DIRS`
- `interactivecomments`

#### Tool init (DO NOT duplicate on Aurora):
- Brew shellenv (`eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"`)
- Brew completions linking
- Starship init (`eval "$(starship init zsh)"`)

#### User script sourcing:
- Sources all `~/.zsh/**/*.sh` files (if `~/.zsh/` exists)

### What our chezmoi `.zshrc` provides that CONFLICTS on Aurora

| Our `.zshrc` | System zshrc | Conflict? |
|---|---|---|
| Oh-My-Zsh `compinit` | System compinit (Prezto-style) | **YES** — double compinit, conflicting completion styles |
| `starship init zsh` | System `starship init zsh` | **YES** — double init |
| `alias ll="ls -ltra"` | Bling: `alias ll='eza -l --icons=auto'` | **YES** — ours loses eza features (but bling isn't sourced with chezmoi) |
| SSH agent auto-start | System SSH agent via systemd | **PARTIAL** — `SSH_AUTH_SOCK=/run/user/1000/ssh-agent.socket` already set by systemd |
| `PATH="$HOME/.local/bin:$PATH"` | System zshenv: `$HOME/bin:/usr/local/bin:$PATH` and `$PATH:$HOME/.bin` | **MINOR** — adds `.local/bin`, non-conflicting |

### What our `.zshrc` provides that the system DOES NOT

| Feature | Status on Aurora |
|---|---|
| Oh-My-Zsh plugins (30+) | Not available — system uses raw Prezto-style setup |
| zsh-autosuggestions | Not installed system-wide |
| Atuin init | Not provided (bling has it commented out) |
| FZF integration | Not installed system-wide |
| Work/personal aliases | Not applicable to system config |
| OTel environment variables | Not applicable to system config |

### What the system provides that chezmoi SHOULD NOT touch

- `/etc/zsh/zshenv` — PATH setup, `skip_global_compinit=1`, `noglobalrcs`
- `/etc/zsh/zprofile` — locale, fpath dedup
- `/etc/zsh/zshrc` — the entire file is immutable (rpm-ostree)
- `ZSHCONFIG` variable and `~/.zsh/` sourcing mechanism

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Oh-My-Zsh | **Deploy** — adds 30+ plugins not in system. Accept double-compinit overhead. | **Deploy** — sole completion system |
| History options | **Skip** — system zshrc already sets them | **Need to add** — no system zshrc |
| `starship init` | **Skip** — system zshrc does it | **Do it** — only source |
| SSH agent block | **Skip** — systemd already provides `SSH_AUTH_SOCK` | **Do it** — no systemd SSH agent in WSL |
| Source bling.sh | **Consider** — gets eza/bat/ugrep aliases, direnv, zoxide, mise | **Skip** — bling doesn't exist |
| Brew shellenv | **Skip** — system zshrc does it | **Skip** — no brew |
| chezmoi template guard | Wrap starship init, SSH agent, history opts in `{{ if ne .environment "aurora" }}` | Always run |

---

## 4. Bling Aliases & Tool Inits

### Complete inventory of what bling.sh provides

#### Aliases:
| Alias | Command | Requires |
|---|---|---|
| `ll` | `eza -l --icons=auto --group-directories-first` | eza |
| `l.` | `eza -d .*` | eza |
| `ls` | `eza` | eza |
| `l1` | `eza -1` | eza |
| `grep` | `ug` | ugrep |
| `egrep` | `ug -E` | ugrep |
| `fgrep` | `ug -F` | ugrep |
| `xzgrep` | `ug -z` | ugrep |
| `xzegrep` | `ug -zE` | ugrep |
| `xzfgrep` | `ug -zF` | ugrep |
| `cat` | `bat --style=plain --pager=never` | bat |

#### Tool inits:
| Tool | Init | Notes |
|---|---|---|
| direnv | `eval "$(direnv hook zsh)"` | Runs before bash-preexec to avoid PROMPT_COMMAND conflicts |
| atuin | **Commented out / disabled** | User must enable manually |
| starship | `eval "$(starship init zsh)"` | Active |
| zoxide | `eval "$(zoxide init zsh)"` | Active |
| mise | `eval "$(mise activate zsh)"` | Active, can be disabled with `MISE_ZSH_AUTO_ACTIVATE=0` |

#### Guard mechanism:
- `BLING_SOURCED=1` flag prevents double-sourcing

### What chezmoi needs on WSL/Distrobox (bling doesn't exist there)

chezmoi must provide ALL of the following on non-Aurora environments:

```zsh
# Aliases (only if tools are installed — WSL may not have eza/bat/ugrep)
command -v eza &>/dev/null && alias ll='eza -l --icons=auto --group-directories-first'
command -v eza &>/dev/null && alias ls='eza' && alias l.='eza -d .*' && alias l1='eza -1'
command -v bat &>/dev/null && alias cat='bat --style=plain --pager=never'

# Tool inits
command -v direnv &>/dev/null && eval "$(direnv hook zsh)"
command -v zoxide &>/dev/null && eval "$(zoxide init zsh)"
command -v mise &>/dev/null && eval "$(mise activate zsh)"
```

### Current gap in chezmoi

Our `dot_zshrc.tmpl` currently has `alias ll="ls -ltra"` which:
- On Aurora: overrides system/bling eza alias with plain `ls` (bad)
- On WSL/Distrobox: is the only `ll` alias (ok, but could be better with eza if installed)

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Source bling.sh in `.zshrc` | **Yes** — add `test -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh` | **No** — doesn't exist |
| Provide eza/bat/ugrep aliases | **No** — bling handles it | **Yes** — template-guarded |
| Provide direnv/zoxide/mise init | **No** — bling handles it | **Yes** — template-guarded |
| Remove `alias ll="ls -ltra"` | **Yes** — conflicts with bling's eza alias | Replace with eza-aware version |
| chezmoi template guard | `{{ if eq .environment "aurora" }}source bling{{ else }}provide aliases+inits{{ end }}` | Non-aurora block |

---

## 5. Git Config

### What exists on Aurora now

- `~/.gitconfig` — **does not exist** (chezmoi not yet applied)
- `~/.config/git/` — **does not exist**
- No system-wide git config beyond Fedora defaults

### What our chezmoi provides

- `dot_gitconfig.tmpl` → `~/.gitconfig`:
  - User name: `Rommel Porras` (hardcoded)
  - Conditional includes: `gitdir:~/personal/` → `~/.config/git/personal.gitconfig`, `gitdir:~/eam/` → `~/.config/git/work.gitconfig`
  - Custom alias `cc` for conventional commits
- `dot_config/git/personal.gitconfig.tmpl` — sets `user.email` from template var
- `dot_config/git/work.gitconfig.tmpl` — sets `user.email` from template var
- `dot_config/git/ignore` — global gitignore

### Planned features not yet implemented

- **1Password SSH commit signing** — mentioned in task context but NOT in current templates
  - Would need: `gpg.format = ssh`, `user.signingkey`, `gpg.ssh.program = /opt/1Password/op-ssh-sign`, `commit.gpgsign = true`
  - Template-guard: only on machines with 1Password (currently none — see section 9)

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Deploy `~/.gitconfig` | **Yes** — nothing exists | **Yes** — same |
| Deploy `~/.config/git/*.gitconfig` | **Yes** | **Yes** |
| Add 1Password SSH signing | **Later** — 1Password not installed yet | **Later** — needs 1Password |
| chezmoi template guard | Signing config: `{{ if eq .environment "aurora" }}` (or a `has_1password` var) | Per-environment |

---

## 6. Shell Integrations — Source Map

### Where each tool init comes from

| Tool | `/etc/profile.d/` | `/etc/zsh/zshrc` | `bling.sh` | chezmoi `.zshrc` | Status |
|---|---|---|---|---|---|
| **brew** (bash) | `brew.sh` — interactive bash only, PATH appended | — | — | — | Covered for bash |
| **brew** (zsh) | — | Full `brew shellenv` + completions | — | — | Covered by system |
| **starship** (bash) | `90-aurora-starship.sh` | — | `starship init bash` | — | Double init for bash (profile.d + bling) |
| **starship** (zsh) | — | `starship init zsh` | `starship init zsh` | `starship init zsh` | **TRIPLE init risk** if all 3 active |
| **direnv** | — | — | `direnv hook $SHELL` | — | Only from bling |
| **zoxide** | — | — | `zoxide init $SHELL` | — | Only from bling |
| **mise** | — | — | `mise activate zsh` | — | Only from bling |
| **atuin** | — | — | **Disabled** (commented out) | `atuin init zsh` | Only from chezmoi |
| **eza aliases** | — | — | `ll`, `l.`, `ls`, `l1` | — | Only from bling |
| **bat alias** | — | — | `cat='bat ...'` | — | Only from bling |
| **ugrep aliases** | — | — | `grep`, `egrep`, etc. | — | Only from bling |
| **SSH agent** | — | — | — | SSH_AUTH_SOCK check | systemd provides on Aurora |
| **FZF** | — | — | — | `source ~/.fzf.zsh` | Only from chezmoi |
| **NVM/Bun** | — | — | — | WSL-only block | Only from chezmoi |

### What's MISSING that chezmoi must add

| Missing integration | Where to add | Notes |
|---|---|---|
| **HISTFILE + SAVEHIST** | chezmoi `.zshrc` (ALL environments) | **Critical bug:** System zshrc sets history *options* but never sets `HISTFILE` or `SAVEHIST`. Result: zsh history is memory-only and lost on terminal close. Needs `HISTFILE=~/.zsh_history` and `SAVEHIST=50000` (or similar). |
| **Atuin init** | chezmoi `.zshrc` | Done — already in template |
| **SSH_AUTH_SOCK for 1Password** | chezmoi `.zshrc` | `export SSH_AUTH_SOCK=~/.1Password/agent.sock` — not yet implemented, 1Password not installed |
| **direnv/zoxide/mise on WSL** | chezmoi `.zshrc` | Only needed if these tools are installed on WSL. Currently not in WSL bootstrap. |
| **eza/bat aliases on WSL** | chezmoi `.zshrc` | Same — only if tools installed |
| **FZF** | chezmoi `.zshrc` | Already there (`~/.fzf.zsh` source) |
| **Source bling on Aurora** | chezmoi `.zshrc` | Critical — without it, Aurora loses eza/bat/direnv/zoxide/mise |

---

## 7. Profile.d Scripts

### `/etc/profile.d/brew.sh`

```bash
if [[ -d /home/linuxbrew/.linuxbrew && $- == *i* ]] ; then
  eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv | grep -Ev '\bPATH=')"
  HOMEBREW_PREFIX="${HOMEBREW_PREFIX:-/home/linuxbrew/.linuxbrew}"
  export PATH="${PATH}:${HOMEBREW_PREFIX}/bin:${HOMEBREW_PREFIX}/sbin"
fi
```

**Key behavior:** Brew is **appended** to PATH (not prepended) for bash. This is intentional — system binaries take priority over brew to prevent brew from overriding things like dbus.

**Conflict with chezmoi?** No — this only runs for bash. For zsh, `/etc/zsh/zshrc` handles brew (also appending via fpath). chezmoi doesn't need to touch brew init on Aurora.

### `/etc/profile.d/90-aurora-starship.sh`

```bash
command -v starship >/dev/null 2>&1 || return 0
if [ "$(basename "$(readlink /proc/$$/exe)")" = "bash" ]; then
  eval "$(starship init bash)"
fi
```

**Key behavior:** Bash-only starship init. Does NOT run for zsh (zsh gets it from `/etc/zsh/zshrc`).

**Conflict with chezmoi?** No — chezmoi only deploys `.zshrc`, not `.bashrc`. No bash-level conflict.

### Other profile.d scripts of note

- `ublue-fastfetch.sh` — aliases `neofetch`/`fastfetch` to `ublue-fastfetch`. No conflict.
- `ublue-motd.sh` — terminal MOTD. No conflict.
- Standard Fedora scripts (`colorgrep.sh`, `colorls.sh`, `lang.sh`, `less.sh`, `vim.sh`). No conflicts.

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Touch profile.d scripts | **Never** — immutable OS layer | **N/A** — don't exist |
| Duplicate brew init | **No** — system handles it | **No** — no brew |
| Duplicate starship bash init | **No** — profile.d handles it | **N/A** — we don't manage .bashrc |

---

## 8. Bat / Eza / Zoxide Configs

### Bat

- `~/.config/bat/` — **does not exist** on Aurora
- No system-wide bat config
- Bling provides `alias cat='bat --style=plain --pager=never'` — no config file needed
- chezmoi does NOT manage any bat config

**Recommendation:** No action needed. The bat alias from bling (on Aurora) or from chezmoi's alias block (on WSL/Distrobox) is sufficient. No custom theme or config file needed unless user wants one.

### Eza

- No `~/.config/eza/` or eza config anywhere
- Eza is configured purely through bling aliases
- chezmoi does NOT manage any eza config

**Recommendation:** No action needed. Aliases are sufficient.

### Zoxide

- No `~/.config/zoxide/` or custom config
- Zoxide is initialized via `eval "$(zoxide init zsh)"` in bling.sh
- No config file needed — zoxide uses its own database at `~/.local/share/zoxide/db.zo`
- chezmoi does NOT manage any zoxide config

**Recommendation:** No action needed for config. Just ensure `zoxide init zsh` runs (via bling on Aurora, via chezmoi on WSL/Distrobox).

---

## 9. 1Password SSH Agent

### Current state on Aurora

- `~/.1Password/` — **does not exist**
- `~/.config/1Password/ssh/agent.toml` — **does not exist**
- `SSH_AUTH_SOCK=/run/user/1000/ssh-agent.socket` — systemd's default ssh-agent
- 1Password desktop app — **not installed** (neither Flatpak nor native)

### Planned chezmoi integration

When 1Password is installed:
1. The 1Password desktop app creates `~/.1Password/agent.sock`
2. chezmoi should set `export SSH_AUTH_SOCK=~/.1Password/agent.sock` in `.zshrc`
3. chezmoi should deploy `~/.config/1Password/ssh/agent.toml` to whitelist SSH keys
4. Git config should add SSH commit signing (`gpg.format = ssh`, `gpg.ssh.program`)

### Recommendations

| Action | Aurora | WSL/Distrobox |
|---|---|---|
| Deploy SSH_AUTH_SOCK export | **Later** — install 1Password first | **Later** — needs 1Password |
| Deploy agent.toml | **Later** | **Later** |
| Add git signing config | **Later** | **Later** |
| Template guard | Add `has_1password` chezmoi var, or detect `~/.1Password/agent.sock` at init time | Same |
| Remove SSH agent auto-start | On Aurora when 1Password is active — systemd/1P handles it | When 1Password is active |

---

## Summary: Per-Tool Action Matrix

### Legend
- **Deploy** = chezmoi should manage this file/config
- **Skip** = chezmoi should NOT touch this (Aurora provides it)
- **Guard** = wrap in template conditional `{{ if ... }}`
- **Add** = new config/feature to create in chezmoi

| Tool/Config | Aurora | WSL | Distrobox | Template Guard |
|---|---|---|---|---|
| **starship.toml** | Deploy (replace Aurora default) | Deploy | Deploy | None — same file everywhere |
| **starship init** | Skip (system zshrc does it) | Deploy | Deploy | `{{ if ne .environment "aurora" }}` |
| **custom.update module** | Add to starship.toml | Skip | Skip | `{{ if eq .environment "aurora" }}` (if templated) |
| **atuin config.toml** | Deploy | Deploy | Deploy | Existing `atuin_account` guard |
| **atuin init** | Deploy | Deploy | Deploy | Existing `atuin_account` guard |
| **Oh-My-Zsh** | Deploy | Deploy | Deploy | None — universal |
| **compinit** | Skip (system does it, OMZ does it too — accept double) | Via OMZ | Via OMZ | None |
| **HISTFILE/SAVEHIST vars** | **Add** (system sets options but NOT these vars — no history persists!) | **Add** | **Add** | None — needed everywhere |
| **History options** | Skip (system sets them) | Add explicitly | Add explicitly | `{{ if ne .environment "aurora" }}` |
| **Brew shellenv** | Skip (system zshrc) | Skip (no brew) | Skip (no brew) | N/A |
| **Source bling.sh** | **Add** | Skip (doesn't exist) | Skip (doesn't exist) | `{{ if eq .environment "aurora" }}` |
| **eza/bat/ugrep aliases** | Skip (bling provides) | **Add** | **Add** | `{{ if ne .environment "aurora" }}` |
| **direnv hook** | Skip (bling provides) | **Add** (if installed) | **Add** (if installed) | `{{ if ne .environment "aurora" }}` |
| **zoxide init** | Skip (bling provides) | **Add** (if installed) | **Add** (if installed) | `{{ if ne .environment "aurora" }}` |
| **mise activate** | Skip (bling provides) | **Add** (if installed) | **Add** (if installed) | `{{ if ne .environment "aurora" }}` |
| **SSH agent start** | Skip (systemd provides) | Deploy | Deploy | `{{ if ne .environment "aurora" }}` |
| **FZF source** | Deploy | Deploy | Deploy | None — guarded by `[ -f ]` |
| **gitconfig** | Deploy | Deploy | Deploy | Existing conditional includes |
| **git signing (1P)** | Later | Later | Later | New `has_1password` var |
| **1Password SSH_AUTH_SOCK** | Later | Later | Later | New `has_1password` var |
| **bat/eza/zoxide config files** | Skip (none needed) | Skip | Skip | N/A |
| **`alias ll`** | **Remove** (conflicts with bling) | Replace with eza-aware | Replace with eza-aware | Guard or make eza-aware universally |

---

## Recommended `.zshrc` Template Structure

```zsh
# ─── PATH ──────────────────────────────────────────────────────────
export PATH="$HOME/.local/bin:$PATH"

# ─── Oh My Zsh ────────────────────────────────────────────────────
export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME=""
plugins=( ... )
source $ZSH/oh-my-zsh.sh

{{ if eq .environment "aurora" -}}
# ─── Aurora: source bling for eza/bat/direnv/zoxide/mise ──────────
# System /etc/zsh/zshrc already provides: compinit, history, brew, starship
test -f /usr/share/ublue-os/bling/bling.sh && source /usr/share/ublue-os/bling/bling.sh
{{ else -}}
# ─── Non-Aurora: provide tool inits that bling would give ─────────
command -v starship &>/dev/null && eval "$(starship init zsh)"
command -v direnv &>/dev/null && eval "$(direnv hook zsh)"
command -v zoxide &>/dev/null && eval "$(zoxide init zsh)"
command -v mise &>/dev/null && eval "$(mise activate zsh)"

# Aliases (eza-aware)
if command -v eza &>/dev/null; then
    alias ll='eza -l --icons=auto --group-directories-first'
    alias ls='eza'
    alias l.='eza -d .*'
    alias l1='eza -1'
else
    alias ll='ls -ltra'
fi
command -v bat &>/dev/null && alias cat='bat --style=plain --pager=never'

# History (Aurora system zshrc provides these; WSL/Distrobox need them)
setopt INC_APPEND_HISTORY SHARE_HISTORY APPEND_HISTORY
setopt COMPLETE_IN_WORD ALWAYS_TO_END AUTO_MENU AUTO_LIST
setopt AUTO_PARAM_SLASH EXTENDED_GLOB AUTO_CD
unsetopt MENU_COMPLETE FLOW_CONTROL

# SSH agent
if [ -z "$SSH_AUTH_SOCK" ]; then
    eval "$(ssh-agent -s)" > /dev/null
    ssh-add ~/.ssh/id_ed25519 2>/dev/null
fi
{{ end -}}

# ─── Universal (all environments) ─────────────────────────────────
{{ if ne .atuin_account "none" -}}
command -v atuin &>/dev/null && eval "$(atuin init zsh)"
{{ end -}}
[ -f ~/.fzf.zsh ] && source ~/.fzf.zsh

# Aliases
alias gd="git diff"
alias gcmsg="git commit -m"
# ... (work/personal/WSL blocks unchanged)
```

---

## Open Questions

1. **Should we template `starship.toml` to add `custom.update` on Aurora?** Or keep a single universal file and accept that the ostree indicator only works on Aurora (it gracefully no-ops elsewhere since `rpm-ostree` won't exist)?

2. **`inline_height`**: Task says 30, template says 20. Which is correct?

3. **Should bling.sh sourcing go before or after Oh-My-Zsh?** After is safer — bling's `BLING_SOURCED` guard prevents issues, and Oh-My-Zsh's plugins need to load first.

4. **Do we need `ugrep` aliases on WSL/Distrobox?** ugrep is not in the WSL bootstrap script. If not installed, aliases are harmless (guarded by `command -v`), but they're also useless.

5. **`ZSH_THEME="spaceship"` in system zshrc** — The system sets this but starship overrides it. With Oh-My-Zsh's `ZSH_THEME=""`, this is effectively ignored. No action needed.

6. **Double compinit on Aurora** — System zshrc runs `compinit` with Prezto-style cache, then Oh-My-Zsh runs `compinit` again. This adds ~100ms startup time. Acceptable? Or should we set `skip_global_compinit=1` and `DISABLE_COMPFIX=true`? (Note: system zshenv already sets `skip_global_compinit=1`.)
