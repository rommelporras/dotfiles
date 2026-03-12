# Aurora DX Setup

## 1. Platform Prerequisites (follow in order)

> **Important:** Follow these steps in order. `ujust devmode` rebases to a new OS
> image, which resets any rpm-ostree layered packages. Install 1Password **after**
> devmode, not before.

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
8. SSH agent socket is at `~/.1password/agent.sock` (lowercase p).
   After chezmoi apply, `.zshrc` sets `SSH_AUTH_SOCK` automatically.

9. Install Claude Code and uv (Python project manager — needed for distrobox scripts):
   ```bash
   brew install claude-code uv
   ```

## 2. Install chezmoi and apply dotfiles

```bash
# Clone the repo first (chezmoi is already installed via brew)
mkdir -p ~/personal
git clone git@github.com:rommelporras/dotfiles.git ~/personal/dotfiles

chezmoi init --apply ~/personal/dotfiles
```

After install:
```bash
exec zsh
```

## 3. Build dotctl

```bash
cd ~/personal/dotfiles
make install          # builds and copies to ~/.local/bin/
make install-systemd  # enables 10-minute collection timer
```

## 4. Set up Distrobox containers

Ensure 1Password desktop app is unlocked and CLI integration is enabled
(Settings → Developer → **Integrate with 1Password CLI**).

```bash
cd ~/personal/dotfiles

# Personal container
uv run python scripts/distrobox_setup.py personal \
  --personal-email git@rommelporras.com

# Work container
uv run python scripts/distrobox_setup.py work-eam \
  --work-email work@company.com

# All default containers
uv run python scripts/distrobox_setup.py \
  --personal-email git@rommelporras.com \
  --work-email work@company.com
```

See [docs/reference/distrobox-scripts.md](../reference/distrobox-scripts.md) for full reference.

## 5. Set up credentials

See [docs/reference/credentials.md](../reference/credentials.md).
