# WSL2 Setup

**Work laptop:** two separate WSL2 Ubuntu instances — one personal, one work-isolated.
**Gaming desktop:** one instance with `personal` context.

| Instance | Context | What it gets |
|---|---|---|
| Ubuntu (personal) | `personal` | Personal tools, homelab access, kubectl, VAULT_ADDR |
| Ubuntu-Work | `work-eam` | Terraform, AWS CLI, kubectl, work aliases |

Both share the same 1Password SSH agent bridge (npiperelay) and the same Atuin server
but with different accounts (`personal` vs `work-eam`).

## 0. Creating a new WSL2 instance

Each context gets its own isolated WSL2 instance:

```powershell
# In PowerShell on Windows
wsl --install --distribution Ubuntu --name Ubuntu-Work      # work-eam context
wsl --install --distribution Ubuntu --name Ubuntu-Personal  # personal context
```

> If `--name` isn't supported on your Windows version, install additional Ubuntu
> versions from Microsoft Store (e.g. "Ubuntu 24.04 LTS") — each creates a separate instance.

Then continue with the steps below inside the new instance.

## 1. Platform Prerequisites

### Windows side (one time per machine)

1. Install **1Password for Windows** (desktop app, not Microsoft Store)
2. Settings → Developer → enable **SSH Agent** → choose "Use Key Names"
3. Settings → Developer → enable **Integrate with 1Password CLI**
4. Install **1Password CLI on Windows**: `winget install AgileBits.1Password.CLI`
5. Disable the **OpenSSH Authentication Agent** service
   (`Win+R` → `services.msc` → "OpenSSH Authentication Agent" → Startup type: Disabled)
6. Import SSH keys via 1Password desktop: New Item → SSH Key → Add Private Key → Import
   - Store as "SSH Key" category in **Private** vault (not Secure Note — agent won't serve those)

> **1Password CLI on WSL:** The native Linux `op` binary cannot connect to the Windows
> desktop app. The bootstrap deploys a wrapper at `~/.local/bin/op` that forwards calls
> to Windows `op.exe`, which triggers the desktop app biometric popup for authorization.
> No `op account add` or `eval $(op signin)` needed — just keep the desktop app running.

### WSL side (once per instance)

7. Install the npiperelay bridge (required for WSL to access Windows 1Password agent):
   ```bash
   sudo apt install -y socat unzip git curl
   curl -Lo /tmp/npiperelay.zip "https://github.com/jstarks/npiperelay/releases/latest/download/npiperelay_windows_amd64.zip"
   unzip -o /tmp/npiperelay.zip -d /tmp/npiperelay
   sudo install -m 0755 /tmp/npiperelay/npiperelay.exe /usr/local/bin/npiperelay.exe
   rm -rf /tmp/npiperelay /tmp/npiperelay.zip
   ```
8. Start the SSH agent bridge manually (needed for git clone before `.zshrc` is applied):
   ```bash
   mkdir -p ~/.1password
   (setsid socat UNIX-LISTEN:$HOME/.1password/agent.sock,fork EXEC:"npiperelay.exe -ei -s //./pipe/openssh-ssh-agent",nofork &) >/dev/null 2>&1
   export SSH_AUTH_SOCK="$HOME/.1password/agent.sock"
   ssh-add -l   # should list your keys — if empty, check 1Password SSH Agent settings
   ```
   After chezmoi apply, the `.zshrc` bridge script handles this automatically on every shell start.

## 2. Run setup script

Clone the repo first, then one command handles the rest — installs chezmoi, applies
dotfiles, installs Claude plugins, and seeds credentials:

```bash
# Install uv (Python runner — needed for the setup script)
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.local/bin/env

# Clone dotfiles repo
mkdir -p ~/personal
git clone git@github.com:rommelporras/dotfiles.git ~/personal/dotfiles
cd ~/personal/dotfiles

# Run setup — work context
uv run python scripts/wsl_setup.py work-eam --work-email you@company.com

# Or — personal context
uv run python scripts/wsl_setup.py personal --personal-email you@example.com
```

> To skip credential seeding (if 1Password isn't unlocked yet), add `--skip-creds`.
> You can run `setup-creds` later to finish.

The script:
1. Validates WSL prerequisites (1Password, npiperelay, SSH agent)
2. Clones `claude-config` repo to `~/personal/` (dotfiles already cloned above)
3. Installs chezmoi and symlinks source → `~/personal/dotfiles`
4. Writes non-interactive chezmoi config and runs `chezmoi init --apply --no-pager --force`
5. Runs `setup-creds` (Claude plugins, Context7 MCP, Atuin login)

After setup:
```bash
exec zsh
```

> If new terminal windows still open in bash, log out and back in — `chsh`
> requires a new login session.

## 3. Font setup

Install JetBrainsMono Nerd Font manually on Windows:
1. Download from https://www.nerdfonts.com/font-downloads
2. Extract zip, select all `.ttf` files, right-click → Install
3. Windows Terminal → Settings → Profile → Appearance → Font face → `JetBrainsMono Nerd Font`

## 4. Remaining manual steps

```bash
# Enable dotctl auto-collection (pushes metrics to OTel every 10 min)
cd ~/personal/dotfiles && make install-systemd

# GitHub CLI (browser OAuth)
gh auth login
```

For additional credential setup (AWS SSO, kubeconfig), see
[docs/reference/credentials.md](../reference/credentials.md).

If `setup-creds` failed (1Password was locked), re-run it:
```bash
setup-creds
```

## 5. Keeping in sync

When dotfiles are updated on Aurora (or from any machine), pull and apply on WSL2:

```bash
dotup             # alias for: chezmoi update -v --no-pager --force && exec zsh
```

Or manually:
```bash
chezmoi update -v --no-pager --force   # git pull from GitHub + apply changes
exec zsh                               # reload shell if .zshrc changed
```

If the bootstrap script changed (new tools added), re-run it:
```bash
sudo -v   # cache sudo first — chezmoi captures stdin so you can't type passwords during apply
chezmoi state delete-bucket --bucket=scriptState
chezmoi apply --no-pager --force
```

> `--no-pager` prevents chezmoi from opening a diff viewer (blocks on `:` colon).
> `--force` auto-overwrites oh-my-zsh cache files without prompting.

If `.chezmoi.toml.tmpl` changed (new template variables added), re-run init
to pick up the new prompts:
```bash
chezmoi init
chezmoi apply --no-pager --force
```

## What bootstrap installs automatically

These tools are installed by `chezmoi init --apply` — no manual action needed:

| Tool | All contexts | `work-*` only | `personal` only |
|---|---|---|---|
| zsh, starship, fzf, atuin | x | | |
| NVM + Node.js 24, Bun | x | | |
| Claude Code | x | | |
| Go, GitHub CLI, uv, gitleaks | x | | |
| dotctl (built from source) | x | | |
| 1Password CLI (`op.exe` wrapper) | prerequisite (step 1) | | |
| Terraform, AWS CLI | | x | |
| kubectl | | x | x |

**Still manual:**
- 1Password + npiperelay (step 1 — Windows-side setup)
- `gh auth login` (step 4 — browser OAuth)
- Font install (step 3 — Windows-side)
