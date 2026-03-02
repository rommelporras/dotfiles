# dotfiles

chezmoi-managed dotfiles for consistent dev environments across WSL2 Ubuntu,
Aurora DX, and Distrobox containers.

## What This Sets Up

- **Zsh** with Oh My Zsh (31 plugins) and zsh-autosuggestions
- **Starship** cross-shell prompt
- **Atuin** shell history sync (optional, self-hosted)
- **FZF** fuzzy finder
- **Git** with conditional identity (personal vs work, auto-switches by directory)
- **Conventional commit** helper (`git cc feat "message"`)
- **Claude Code** global config (CLAUDE.md, settings, hooks, skills, agents)
- **AI sandbox** — Podman container for running AI agents with tiered credential access

## Supported Environments

| Environment | Platform | Description |
|---|---|---|
| `wsl-work` | WSL2 Ubuntu | Work laptop — work + personal projects |
| `wsl-gaming` | WSL2 Ubuntu | Gaming desktop — work + personal projects |
| `aurora` | Aurora DX host | Personal laptop — launches Distrobox containers |
| `distrobox-work` | Ubuntu container | Work projects with work credentials |
| `distrobox-personal` | Ubuntu container | Personal projects with homelab credentials |
| `distrobox-sandbox` | Ubuntu container | Clean experiment space, no credentials |

## Quick Start

### Prerequisites

- `curl` and `git` installed
- `sudo` access (bootstrap installs packages)
- A **Nerd Font** in your terminal (for Starship icons)
- **1Password** desktop app installed with SSH Agent enabled (see platform setup below)

#### WSL2 (Ubuntu) — platform setup

1. Install **1Password for Windows** (desktop app, not Microsoft Store)
2. Settings → Developer → enable **SSH Agent** → choose "Use Key Names"
3. Settings → Developer → disable **OpenSSH Authentication Agent** service
   (`Win+R` → `services.msc` → "OpenSSH Authentication Agent" → Startup type: Disabled)
4. Import SSH keys via 1Password desktop: New Item → SSH Key → Add Private Key → Import
   - Store as "SSH Key" category in **Private** vault (not Secure Note — agent won't serve those)
5. Install the npiperelay bridge (required for WSL to access Windows 1Password agent):
   ```bash
   sudo apt install -y socat
   curl -Lo /tmp/npiperelay.zip "https://github.com/jstarks/npiperelay/releases/latest/download/npiperelay_windows_amd64.zip"
   unzip -o /tmp/npiperelay.zip -d /tmp/npiperelay
   sudo install -m 0755 /tmp/npiperelay/npiperelay.exe /usr/local/bin/npiperelay.exe
   rm -rf /tmp/npiperelay /tmp/npiperelay.zip
   ```
6. After chezmoi apply, the `.zshrc` bridge script connects WSL to the Windows agent
   via socket at `~/.1password/agent.sock`

#### Aurora DX — platform setup

1. Enable developer mode (installs Docker, Podman, Distrobox, dev tooling):
   ```bash
   ujust devmode
   ```
   Follow the prompts, then reboot when finished.

2. Add your user to developer groups (docker, etc.):
   ```bash
   ujust dx-group
   ```
   Log out and back in for group changes to take effect.

3. Install brew CLI tools:
   ```bash
   ujust aurora-cli
   SHELL=zsh ujust aurora-cli
   ```
   Close and reopen terminal.

4. Switch default shell to zsh (Aurora ships bash as default):
   ```bash
   brew install zsh
   ```
   Then in Ptyxis: edit the **first** profile (not a new one) →
   "Use Custom Command" → `/home/linuxbrew/.linuxbrew/bin/zsh`
   (Do NOT use `chsh` — atomic systems don't have it, and changing login shell risks login loops)

5. Install **1Password** via rpm-ostree (NOT Flatpak — Flatpak SSH agent is broken by sandbox):
   ```bash
   cat << 'EOF' | sudo tee /etc/yum.repos.d/1password.repo
   [1password]
   name=1Password Stable Channel
   baseurl=https://downloads.1password.com/linux/rpm/stable/$basearch
   enabled=1
   gpgcheck=1
   repo_gpgcheck=1
   gpgkey=https://downloads.1password.com/linux/keys/1password.asc
   EOF
   rpm-ostree install 1password 1password-cli
   systemctl reboot
   ```

6. Open 1Password → sign in → Settings → Developer → enable **SSH Agent**
7. Settings → Security → enable **Unlock using system authentication** (uses Aurora user password or fingerprint)
8. SSH agent socket is at `~/.1Password/agent.sock` (capital P).
   After chezmoi apply, `.zshrc` sets `SSH_AUTH_SOCK` automatically.

9. Install Claude Code:
   ```bash
   brew install claude-code
   ```

> **Note:** Since chezmoi is installed via brew on Aurora, if you've already cloned
> this repo (e.g. to `~/personal/dotfiles`), use the local path approach below
> instead of the `curl | sh` installer.

### 1. Install chezmoi and apply dotfiles

```bash
# WSL: cache sudo first — bootstrap installs packages via apt
sudo -v

# Install chezmoi + clone this repo + run interactive prompts + apply
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
```

#### Already have chezmoi installed?

If chezmoi is already installed (e.g. via `brew` on Aurora) and the repo is already
cloned somewhere, pass the local path directly:

```bash
chezmoi init --apply ~/personal/dotfiles
```

chezmoi will ask you:

| Prompt | What to enter |
|---|---|
| Environment | One of: `wsl-work`, `wsl-gaming`, `aurora`, `distrobox-work`, `distrobox-personal`, `distrobox-sandbox` |
| Personal git email | Your personal email for git commits |
| Work git email | Your work email (leave blank to skip) |
| Work credentials? | `true` if this machine has AWS/EKS access |
| Homelab credentials? | `true` if this machine has homelab kubeconfig |
| Atuin sync server URL | Your self-hosted Atuin URL (leave blank if not set up yet) |
| Atuin account | Your Atuin username, or `none` to skip |

Answers are saved locally to `~/.config/chezmoi/chezmoi.toml` and never committed.

The bootstrap script automatically installs: zsh, Starship, JetBrainsMono Nerd Font,
and environment-specific tools (FZF, xclip, NVM, Bun, Terraform, glab, Ansible)
based on your environment. On Aurora, most tools are pre-installed via brew/RPM and skipped.

**After install, restart your shell:**

```bash
exec zsh
```

> **WSL:** If new terminal windows still open in bash, log out and back in — `chsh`
> requires a new login session.
> **Aurora:** Shell is set via Ptyxis custom command (see platform setup above), not `chsh`.

### Font setup (Nerd Font)

The prompt uses [Nerd Font](https://www.nerdfonts.com/) icons. The bootstrap
installs JetBrainsMono Nerd Font automatically on Aurora/Distrobox. On WSL,
fonts render on the Windows side — install manually:

1. Download **JetBrainsMono** from https://www.nerdfonts.com/font-downloads
2. Extract the zip, select all `.ttf` files, right-click → **Install**
3. In Windows Terminal: Settings → Profile → Appearance → Font face → `JetBrainsMono Nerd Font`

### 2. Set up credentials (manual, per machine)

The bootstrap creates `~/.ssh/` — you populate it with public keys only.
Private keys stay in 1Password; the SSH agent serves them.

```bash
# SSH public keys (needed for IdentityFile matching — private keys stay in 1Password)
# Copy .pub files from another machine, or export from 1Password
cp id_ed25519.pub ~/.ssh/
cp proxmox.pub ~/.ssh/      # if applicable
chmod 644 ~/.ssh/*.pub

# Verify 1Password SSH agent works
ssh-add -l                   # Should list your 1Password SSH keys

# GitHub CLI
gh auth login

# AWS (work environments only)
aws sso login --profile <name>

# EKS kubeconfig (work environments only)
aws eks update-kubeconfig --name <cluster> --region <region>

# Homelab kubeconfig (personal environments only)
cp homelab.yaml ~/.kube/

# GitLab (personal environments only)
glab auth login --hostname gitlab.k8s.rommelporras.com

# Atuin (if configured)
atuin login -u <account-name>
```

### 3. Aurora DX only: set up Distrobox containers

```bash
~/personal/dotfiles/scripts/distrobox-setup.sh
```

This creates three containers (`work`, `personal`, `sandbox`) with separate home
directories and bootstraps chezmoi inside each one. You'll answer the prompts
once per container.

Enter a container:

```bash
distrobox enter work      # or: personal, sandbox
```

### 4. Aurora DX only: build AI sandbox

```bash
ai-sandbox --build
```

## Day-to-Day Usage

### Updating dotfiles

```bash
chezmoi diff          # Preview what would change
chezmoi apply         # Apply changes
chezmoi update        # Pull latest from remote + apply in one step
```

### Editing a managed file

```bash
chezmoi edit ~/.zshrc       # Opens the template in your editor
chezmoi apply               # Apply the change
```

### Adding a new file to chezmoi

```bash
chezmoi add ~/.config/some/config.toml
```

### Re-running bootstrap

The bootstrap is a `run_once_` script — chezmoi skips it on subsequent applies.
To force a re-run:

```bash
chezmoi state delete-bucket --bucket=scriptState
sudo -v && chezmoi apply
```

### Git conventional commits

```bash
git cc feat "add user auth"            # feat: add user auth
git cc fix -s api "handle null body"   # fix(api): handle null body
git cc docs "update README"            # docs: update README
```

## AI Sandbox

Run AI coding agents in an isolated Podman container. No access to host HOME,
personal SSH keys, or cloud credentials.

```bash
# Code only — no git push, no credentials
ai-sandbox claude -- --dangerously-skip-permissions

# Code + git push via dedicated deploy key
ai-sandbox --git claude -- --dangerously-skip-permissions

# Maximum containment — no network at all
ai-sandbox --no-network gemini

# Any AI CLI tool works
ai-sandbox --git aider
ai-sandbox --git antigravity

# Debug the sandbox
ai-sandbox --shell
```

### Credential tiers

| Flag | Network | Git push | Credentials |
|---|---|---|---|
| (default) | yes | no | API keys only (via Podman secrets) |
| `--git` | yes | yes | API keys + deploy key (read-only) |
| `--no-network` | no | no | API keys only |

The `--git` flag mounts `~/.ssh/ai-deploy-key` (not your 1Password-managed keys).
Create it once:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/ai-deploy-key -C 'ai-sandbox-deploy'
# Add the .pub to your GitHub/GitLab repos as a deploy key
```

Store API keys as Podman secrets:

```bash
podman secret create anthropic_key <(echo "sk-ant-...")
podman secret create gemini_key <(echo "AI...")
```

### When to use AI sandbox vs Distrobox

| Scenario | Where | Why |
|---|---|---|
| Daily dev with cluster access | `distrobox-personal` | Needs kubeconfig, SSH keys |
| Vibe-coding with git push | `ai-sandbox --git` | Contained, can push via deploy key |
| Trying untrusted AI tool | `ai-sandbox --no-network` | No network, can't exfiltrate |
| Work Terraform/EKS | `distrobox-work` | Needs AWS + EKS credentials |
| Quick experiment | `distrobox-sandbox` | Clean space, no credentials |

## Terminal Setup

The Starship prompt uses ANSI color names — their actual appearance depends on
your terminal's color scheme. Use the same scheme everywhere for a consistent look.

**Recommended:** Ottosson (perceptually uniform, good contrast on dark backgrounds)

| Terminal | How to set |
|---|---|
| Windows Terminal (WSL) | Settings → Color schemes → **Ottosson** |
| Ptyxis (Aurora DX) | Copy palette TOML to `~/.local/share/org.gnome.Ptyxis/palettes/` (see hex values below) |

<details>
<summary>Ottosson palette (hex values for manual import)</summary>

Based on Björn Ottosson's [Oklab](https://bottosson.github.io/posts/oklab/) color space
for perceptually uniform hue and chroma.

```json
{
    "name": "Ottosson",
    "background": "#000000",
    "foreground": "#bebebe",
    "cursorColor": "#ffffff",
    "selectionBackground": "#92a4fd",
    "black": "#000000",
    "red": "#be2c21",
    "green": "#3fae3a",
    "yellow": "#be9a4a",
    "blue": "#204dbe",
    "purple": "#bb54be",
    "cyan": "#00a7b2",
    "white": "#bebebe",
    "brightBlack": "#808080",
    "brightRed": "#ff3e30",
    "brightGreen": "#58ea51",
    "brightYellow": "#ffc944",
    "brightBlue": "#2f6aff",
    "brightPurple": "#fc74ff",
    "brightCyan": "#00e1f0",
    "brightWhite": "#ffffff"
}
```

Use this JSON directly in Windows Terminal `settings.json` under `schemes`,
or convert the hex values to your terminal's config format (Ptyxis, Konsole, Ghostty, etc.).

</details>

## Repository Structure

```
dotfiles/
├── .chezmoiroot               # Points chezmoi source to home/
├── home/                      # chezmoi source dir (maps to ~/)
│   ├── .chezmoi.toml.tmpl     # Interactive prompts (chezmoi init)
│   ├── .chezmoiexternal.toml  # External deps (oh-my-zsh, plugins)
│   ├── .chezmoiignore         # Per-environment file skipping
│   ├── dot_zshrc.tmpl         # Shell config
│   ├── dot_gitconfig.tmpl     # Git config (conditional includes)
│   ├── run_once_before_bootstrap.sh.tmpl  # First-run setup script
│   ├── private_dot_claude/    # Claude Code global config (~/.claude/)
│   │   ├── CLAUDE.md.tmpl     # Global instructions (templated per env)
│   │   ├── settings.json      # Permissions, hooks, plugins
│   │   ├── hooks/             # Security hooks (secret scan, write protection)
│   │   ├── agents/            # Custom agents (code-reviewer)
│   │   └── skills/            # Skills (/commit, /push, /explain-code)
│   └── dot_config/            # ~/.config/ files
│       ├── starship.toml
│       ├── atuin/config.toml.tmpl
│       ├── ghostty/config
│       ├── k9s/config.yaml
│       └── git/               # Git identity includes + global gitignore
├── bin/                       # CLI tools (ai-sandbox)
├── containers/                # Containerfile.ai-sandbox + distrobox.ini
├── scripts/                   # Setup automation (distrobox-setup.sh)
└── hooks/                     # Git hooks (gitleaks pre-commit)
```

A gitleaks pre-commit hook scans staged changes for secrets before every commit.
If gitleaks is not installed, the hook prints a warning and allows the commit.

### chezmoi naming conventions

| Prefix/suffix | Meaning |
|---|---|
| `dot_` | Becomes `.` in target (`dot_zshrc` → `.zshrc`) |
| `private_dot_` | `.` with owner-only permissions (0600 files, 0700 dirs) |
| `.tmpl` | Processed as Go text/template |
| `run_once_before_` | Script that runs once, before file creation |

## Claude Code Config

Claude Code's global configuration (`~/.claude/`) is managed by chezmoi. This
replaces the old `claude-config` repo that used symlinks.

### What's included

| File | Purpose |
|---|---|
| `CLAUDE.md` | Global instructions (templated per environment) |
| `settings.json` | Permission deny rules, hooks, enabled plugins |
| `hooks/` | Security hooks: secret scanning, write protection, destructive command blocking |
| `agents/code-reviewer.md` | Code review agent with per-project memory |
| `skills/commit/` | `/commit` — conventional commit workflow |
| `skills/push/` | `/push` — push to all configured remotes |
| `skills/explain-code/` | `/explain-code` — structured code explanations |

### Environment differences

`CLAUDE.md` is templated — the "Personal Environment" section adapts:

| Environment | Section content |
|---|---|
| `wsl-work`, `wsl-gaming` | WSL2 specifics (Windows Chrome, no `op`, GitLab primary) |
| `distrobox-personal` | Distrobox specifics (no `op`, GitLab primary) |
| `distrobox-work` | Work specifics (GitHub for work, GitLab for personal) |
| `aurora` | Aurora DX specifics (ostree, 1Password SSH, no `op`, GitLab primary) |
| `distrobox-sandbox` | Skipped entirely (excluded in `.chezmoiignore`) |

### Runtime files

Claude Code generates many runtime files (`history.jsonl`, `projects/`, `cache/`,
etc.). These are excluded in `.chezmoiignore` so chezmoi doesn't touch them.

## History Migration

When setting up Atuin on a new machine with existing shell history:

1. Review history for sensitive entries: `less ~/.zsh_history`
2. Clean passwords or tokens that were typed in the shell
3. Import into Atuin: `atuin import zsh`

## License

MIT License. See [LICENSE](LICENSE) for details.
