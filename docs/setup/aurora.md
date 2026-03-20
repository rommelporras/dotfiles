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
   `.zshrc` sets `SSH_AUTH_SOCK` automatically — but only after chezmoi applies it.
   For the next step, export it manually in the current session:
   ```bash
   export SSH_AUTH_SOCK="$HOME/.1password/agent.sock"
   ssh-add -l   # should list your keys — if empty, check 1Password SSH Agent settings
   ```

9. Install CLI tools needed for dotfiles and homelab:
   ```bash
   brew install go uv kubectl gh glab
   brew install --cask claude-code
   ```
   - `go` — required to build `dotctl` (`make install` in step 3)
   - `uv` — Python project manager, required for distrobox setup scripts
   - `kubectl` — Kubernetes CLI for homelab cluster access
   - `gh` — GitHub CLI for PRs and repo operations
   - `glab` — GitLab CLI for self-hosted GitLab operations
   - `claude-code` — Claude Code cask (installs `claude` binary)

   > The bootstrap also auto-installs `node@24` and `gitleaks` via brew if not already present.

## 2. Install chezmoi and apply dotfiles

```bash
# Clone both repos (chezmoi installed via brew — bundled with ujust aurora-cli)
mkdir -p ~/personal
git clone git@github.com:rommelporras/dotfiles.git ~/personal/dotfiles
git clone git@github.com:rommelporras/claude-config.git ~/personal/claude-config

chezmoi init --apply ~/personal/dotfiles
```

> **Note:** The bootstrap auto-clones `claude-config` via HTTPS if not found, but
> cloning it manually first with SSH ensures you can push changes back later.
> The bootstrap creates symlinks from `~/.claude/` to `~/personal/claude-config/`
> for: `CLAUDE.md`, `settings.json`, `rules/`, `hooks/`, `skills/`, `agents/`.

chezmoi will prompt for:
- **context** — `personal` for Aurora host
- **personal email** — your git email
- **has homelab creds** — `y` if you use kubectl/Vault
- **Atuin sync address** — `https://atuin.k8s.rommelporras.com` (or blank to skip)
- **Atuin account** — `personal` (or `none` to skip)

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

## 5. AI Sandbox (optional)

For vibe-coded AI projects, use `ai-sandbox` instead of distrobox. Build the base
image once, then create project sandboxes:

```bash
ai-sandbox fintrack --build   # build sandbox-base image (first time only)
ai-sandbox fintrack --shell   # create project + drop into shell
```

See [docs/reference/credentials.md](../reference/credentials.md) for podman secret
setup (Atuin, Anthropic, Context7 API keys).

## 6. Set up credentials

See [docs/reference/credentials.md](../reference/credentials.md).

## 7. Keeping in sync

After pushing changes from any machine:

```bash
cd ~/personal/dotfiles
git pull
chezmoi apply     # apply to Aurora host
exec zsh          # reload shell if .zshrc changed
```

If the bootstrap script changed (new tools added), re-run it:
```bash
chezmoi state delete-bucket --bucket=scriptState
chezmoi apply
```

Then apply inside each Distrobox container — see [docs/setup/distrobox.md](distrobox.md).
