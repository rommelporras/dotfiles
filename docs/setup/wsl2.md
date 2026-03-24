# WSL2 Setup

**Work laptop:** two separate WSL2 Ubuntu instances — one personal, one work-isolated.
**Gaming desktop:** one instance with `personal` context.

| Instance | Context | What it gets |
|---|---|---|
| Ubuntu (personal) | `personal` | Personal tools, homelab access, kubectl, VAULT_ADDR |
| Ubuntu-Work | `work-eam` | Terraform, AWS CLI, kubectl, work aliases |

Both share the same 1Password SSH agent bridge (npiperelay) and the same Atuin server
but with different accounts (`personal` vs `work-eam`).

## 0. Work laptop — creating the second WSL2 instance

If you need a work-isolated WSL2 instance alongside an existing personal one:

```powershell
# In PowerShell on Windows — create a fresh Ubuntu instance named Ubuntu-Work
wsl --install --distribution Ubuntu --name Ubuntu-Work
```

> If `--name` isn't supported on your Windows version, install a second Ubuntu
> version from Microsoft Store (e.g. "Ubuntu 24.04 LTS") — it creates a separate instance.

Then continue with the steps below inside the new instance. Use context `work-eam`
when chezmoi prompts. For the existing personal Ubuntu, use context `personal`.

## 1. Platform Prerequisites

1. Install **1Password for Windows** (desktop app, not Microsoft Store)
2. Settings → Developer → enable **SSH Agent** → choose "Use Key Names"
3. Settings → Developer → enable **Integrate with 1Password CLI** (needed for `op` in WSL)
4. Disable the **OpenSSH Authentication Agent** service
   (`Win+R` → `services.msc` → "OpenSSH Authentication Agent" → Startup type: Disabled)
5. Import SSH keys via 1Password desktop: New Item → SSH Key → Add Private Key → Import
   - Store as "SSH Key" category in **Private** vault (not Secure Note — agent won't serve those)
6. Install the npiperelay bridge (required for WSL to access Windows 1Password agent):
   ```bash
   sudo apt install -y socat
   curl -Lo /tmp/npiperelay.zip "https://github.com/jstarks/npiperelay/releases/latest/download/npiperelay_windows_amd64.zip"
   unzip -o /tmp/npiperelay.zip -d /tmp/npiperelay
   sudo install -m 0755 /tmp/npiperelay/npiperelay.exe /usr/local/bin/npiperelay.exe
   rm -rf /tmp/npiperelay /tmp/npiperelay.zip
   ```
7. Install **1Password CLI** in WSL (connects to Windows 1Password desktop app via socket):
   ```bash
   curl -sS https://downloads.1password.com/linux/keys/1password.asc | \
     sudo gpg --dearmor --output /usr/share/keyrings/1password-archive-keyring.gpg
   echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/1password-archive-keyring.gpg] https://downloads.1password.com/linux/debian/$(dpkg --print-architecture) stable main" | \
     sudo tee /etc/apt/sources.list.d/1password-cli.list
   sudo apt update && sudo apt install -y 1password-cli
   op --version   # verify install
   ```
8. Start the SSH agent bridge manually (needed for git clone before `.zshrc` is applied):
   ```bash
   mkdir -p ~/.1password
   (setsid socat UNIX-LISTEN:$HOME/.1password/agent.sock,fork EXEC:"npiperelay.exe -ei -s //./pipe/openssh-ssh-agent",nofork &) >/dev/null 2>&1
   export SSH_AUTH_SOCK="$HOME/.1password/agent.sock"
   ssh-add -l   # should list your keys — if empty, check 1Password SSH Agent settings
   ```
   After chezmoi apply, the `.zshrc` bridge script handles this automatically on every shell start.

## 2. Install chezmoi and apply dotfiles

```bash
# Cache sudo first — bootstrap installs packages via apt
sudo -v

# Clone claude-config first (SSH clone — enables pushing changes later)
mkdir -p ~/personal
git clone git@github.com:rommelporras/claude-config.git ~/personal/claude-config

# Install chezmoi + clone this repo + run interactive prompts + apply
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
```

> **Note:** The bootstrap auto-clones `claude-config` via HTTPS if not found, but
> cloning it manually first with SSH ensures you can push changes back later.
> The bootstrap creates symlinks from `~/.claude/` to `~/personal/claude-config/`
> for: `CLAUDE.md`, `settings.json`, `rules/`, `hooks/`, `skills/`, `agents/`.

chezmoi will ask (answers vary by context):
- **context** — `personal` for personal use (work laptop or gaming desktop), `work-eam` for work
- **personal email** — your git email (always prompted, even for work contexts on WSL)
- **work email** — only if context is `work-*`
- **work credentials** — only if context is `work-*`
- **homelab credentials** — if context is `personal` (also prompted on Aurora for all contexts)
- **Atuin sync address** — `https://atuin.k8s.rommelporras.com` (or blank to skip)
- **Atuin account** — `personal` or `work-eam` (or `none` to skip)

After install:
```bash
exec zsh
```

> **WSL:** If new terminal windows still open in bash, log out and back in — `chsh`
> requires a new login session.

## 3. Font setup

Install JetBrainsMono Nerd Font manually on Windows:
1. Download from https://www.nerdfonts.com/font-downloads
2. Extract zip, select all `.ttf` files, right-click → Install
3. Windows Terminal → Settings → Profile → Appearance → Font face → `JetBrainsMono Nerd Font`

## 4. Install GitHub CLI

```bash
(type -p wget >/dev/null || sudo apt install wget -y) \
  && sudo mkdir -p -m 755 /etc/apt/keyrings \
  && out=$(mktemp) && wget -nv -O$out https://cli.github.com/packages/githubcli-archive-keyring.gpg \
  && cat $out | sudo tee /etc/apt/keyrings/githubcli-archive-keyring.gpg > /dev/null \
  && sudo chmod go+r /etc/apt/keyrings/githubcli-archive-keyring.gpg \
  && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
  && sudo apt update && sudo apt install gh -y
gh auth login
```

## 5. Claude Code plugins

Bootstrap echoes the exact commands to run. After install:

```bash
claude plugin marketplace add anthropics/claude-plugins-official
claude plugin marketplace add obra/superpowers-marketplace
claude plugin marketplace add nextlevelbuilder/ui-ux-pro-max-skill
claude plugin install context7@claude-plugins-official --scope user
claude plugin install playwright@claude-plugins-official --scope user
claude plugin install superpowers@superpowers-marketplace --scope user
claude plugin install episodic-memory@superpowers-marketplace --scope user
claude plugin install ui-ux-pro-max@ui-ux-pro-max-skill --scope user

claude mcp add --scope user --transport http context7 https://mcp.context7.com/mcp \
  --header "CONTEXT7_API_KEY: $(op read 'op://Kubernetes/Context7/api-key' --no-newline)"
```

## 6. Set up credentials

See [docs/reference/credentials.md](../reference/credentials.md).

## 7. Keeping in sync

When dotfiles are updated on Aurora (or from any machine), pull and apply on WSL2:

```bash
dotup             # alias for: chezmoi update -v && exec zsh
```

Or manually:
```bash
chezmoi update -v   # git pull from GitHub + apply changes
exec zsh            # reload shell if .zshrc changed
```

If the bootstrap script changed (new tools added), re-run it:
```bash
chezmoi state delete-bucket --bucket=scriptState
chezmoi apply
```
