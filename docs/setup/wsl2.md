# WSL2 Setup

**Work laptop uses two separate WSL2 Ubuntu instances** — one personal, one work-isolated.
Gaming desktop uses one instance (gaming context).

| Instance | Context | What it gets |
|---|---|---|
| Ubuntu (personal) | `personal` | Personal tools, homelab access, VAULT_ADDR |
| Ubuntu-Work | `work-eam` | Terraform, AWS CLI, kubectl, work aliases, OTel |

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

## 2. Install chezmoi and apply dotfiles

```bash
# Cache sudo first — bootstrap installs packages via apt
sudo -v

# Install chezmoi + clone this repo + run interactive prompts + apply
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
```

chezmoi will ask (answers vary by context):
- **context** — `gaming` for personal gaming desktop, `work-eam` for work laptop
- **personal email** — your git email (always prompted, even for work contexts on WSL)
- **work email** — only if context is `work-*`
- **work credentials** — only if context is `work-*`
- **homelab credentials** — only if context is `gaming` or `personal`
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

## 4. Set up credentials

See [docs/reference/credentials.md](../reference/credentials.md).

## 5. Keeping in sync

When dotfiles are updated on Aurora (or from any machine), pull and apply on WSL2:

```bash
chezmoi update    # git pull from GitHub + apply changes
exec zsh          # reload shell if .zshrc changed
```

`chezmoi update` is equivalent to:
```bash
cd ~/.local/share/chezmoi && git pull
chezmoi apply
```

If the bootstrap script changed (new tools added), re-run it:
```bash
chezmoi state delete-bucket --bucket=scriptState
chezmoi apply
```
