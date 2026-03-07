# dotfiles

chezmoi-managed dotfiles for consistent dev environments across WSL2 Ubuntu,
Aurora DX, and Distrobox containers.

## What This Sets Up

- **Zsh** with Oh My Zsh (32 plugins) and zsh-autosuggestions
- **Starship** cross-shell prompt
- **Atuin** shell history sync (optional, self-hosted)
- **FZF** fuzzy finder
- **Git** with conditional identity (personal vs work, auto-switches by directory)
- **Conventional commit** helper (`git cc feat "message"`)
- **Claude Code** config via [claude-config](https://github.com/rommelporras/claude-config) repo (CLAUDE.md, settings, hooks, skills, agents)
- **AI sandbox** — Podman container for running AI agents with tiered credential access

## Supported Environments

Templates use two variables: **platform** (auto-detected) and **context** (user-selected).

| Platform | Context | Description |
|---|---|---|
| `wsl` | `work-eam` | Work laptop — EAM work projects |
| `wsl` | `gaming` | Gaming desktop — personal projects |
| `aurora` | `personal` | Personal laptop — launches Distrobox containers |
| `distrobox` | `work-eam` | Work projects with work credentials |
| `distrobox` | `personal` | Personal projects with homelab credentials |
| `distrobox` | `personal-<project>` | Project-scoped dev (native `op` CLI, Bun, Playwright, no homelab) |
| `distrobox` | `sandbox` | Clean experiment space, no credentials |

**Adding contexts:** For work: add a container to `containers/distrobox.ini`, add job-specific
aliases in `dot_zshrc.tmpl`, run the [setup script](docs/distrobox-scripts.md) with
`work-acme`. Shared work tools (AWS CLI, kubectl, Terraform, work email) apply automatically
via `hasPrefix .context "work-"`. For personal projects: add a container to `distrobox.ini`,
run the setup script with `personal-<project>`. Gets glab, Bun, Playwright, native 1Password
CLI (biometric unlock), and OTel telemetry — but no homelab kubeconfig or Ansible.

## Quick Start

Pick your platform below, then continue to [Install chezmoi](#1-install-chezmoi-and-apply-dotfiles).

---

### WSL2 (Ubuntu) — platform setup

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

Continue to [Install chezmoi](#1-install-chezmoi-and-apply-dotfiles).

---

### Aurora DX — platform setup

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

10. Clone this repo (chezmoi is already installed via brew):
    ```bash
    mkdir -p ~/personal
    git clone git@github.com:rommelporras/dotfiles.git ~/personal/dotfiles
    ```

Continue to [Install chezmoi](#1-install-chezmoi-and-apply-dotfiles) — use the
"Already have chezmoi installed?" path.

---

### 1. Install chezmoi and apply dotfiles

**WSL (fresh install):**

```bash
# Cache sudo first — bootstrap installs packages via apt
sudo -v

# Install chezmoi + clone this repo + run interactive prompts + apply
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
```

**Aurora DX (chezmoi already installed via brew):**

```bash
chezmoi init --apply ~/personal/dotfiles
```

chezmoi will ask you:

| Prompt | What to enter |
|---|---|
| Context | `personal`, `personal-<project>`, `work-eam`, `work-<name>`, `gaming`, or `sandbox` |
| Personal git email | Your personal email (skipped for sandbox and non-WSL work contexts) |
| Work git email | Your work email (only for `work-*` contexts) |
| Work credentials? | `true` if this machine has AWS/EKS access (only for `work-*` contexts) |
| Homelab credentials? | `true` if this machine has homelab kubeconfig (personal/gaming/aurora) |
| Atuin sync server URL | Your self-hosted Atuin URL (leave blank if not set up yet) |
| Atuin account | Your Atuin username, or `none` to skip |

Platform (`wsl`, `aurora`, `distrobox`) is auto-detected — you'll never be prompted for it.
Answers are saved locally to `~/.config/chezmoi/chezmoi.toml` and never committed.

The bootstrap script automatically installs: zsh, Starship, Claude Code (native installer),
and environment-specific tools (FZF, xclip, NVM, Bun, AWS CLI, kubectl, Terraform, glab,
Ansible) based on your environment. On Aurora, most tools are pre-installed via brew/RPM
and skipped.

**After install, restart your shell:**

```bash
exec zsh
```

> **WSL:** If new terminal windows still open in bash, log out and back in — `chsh`
> requires a new login session.
> **Aurora:** Shell is set via Ptyxis custom command (see platform setup above), not `chsh`.

### Font setup (Nerd Font)

The prompt uses [Nerd Font](https://www.nerdfonts.com/) icons.

- **Aurora:** `ujust devmode` installs JetBrainsMono Nerd Font via brew cask — no action needed.
- **Distrobox:** The bootstrap installs JetBrainsMono Nerd Font automatically.
- **WSL:** Fonts render on the Windows side — install manually:
  1. Download **JetBrainsMono** from https://www.nerdfonts.com/font-downloads
  2. Extract the zip, select all `.ttf` files, right-click → **Install**
  3. In Windows Terminal: Settings → Profile → Appearance → Font face → `JetBrainsMono Nerd Font`

Verify icons render correctly:

```bash
echo -e "\uf418 git  \ue718 node  \ue73c python  \uf308 docker"
```

You should see icons next to each label, not boxes or blanks.

### 2. Set up credentials (per machine)

The bootstrap creates `~/.ssh/` — you populate it with public keys only.
Private keys stay in 1Password; the SSH agent serves them.

**Distrobox containers** have an automated credential-seeding script that pulls
secrets from 1Password on the Aurora host via `distrobox-host-exec op`:

```bash
setup-creds    # Run inside any non-sandbox container
```

This handles Claude Code plugin/marketplace installation, Context7 MCP registration,
Atuin login, glab auth, and prints manual steps for kubeconfig/AWS. The script runs
automatically during container setup — you only need to run it manually if you
skipped it or need to re-authenticate.

**WSL and Aurora host** — set up credentials manually:

```bash
# SSH public keys (needed for IdentityFile matching — private keys stay in 1Password)
cp id_ed25519.pub ~/.ssh/
cp proxmox.pub ~/.ssh/      # if applicable
chmod 644 ~/.ssh/*.pub

# Verify 1Password SSH agent works
ssh-add -l                   # Should list your 1Password SSH keys

# GitHub CLI
gh auth login

# Claude Code plugins (run once)
claude plugin marketplace add anthropics/claude-plugins-official
claude plugin marketplace add obra/superpowers-marketplace
claude plugin install context7@claude-plugins-official --scope user
claude plugin install superpowers@superpowers-marketplace --scope user
claude plugin install episodic-memory@superpowers-marketplace --scope user

# Context7 MCP (needs API key from 1Password)
claude mcp add --scope user --transport http context7 https://mcp.context7.com/mcp \
  --header "CONTEXT7_API_KEY: $(op read 'op://Kubernetes/Context7/api-key' --no-newline)"

# AWS (work environments only)
aws sso login --profile <name>

# EKS kubeconfig (work environments only)
aws eks update-kubeconfig --name <cluster> --region <region>

# Homelab kubeconfig (personal environments only)
cp homelab.yaml ~/.kube/

# GitLab (personal environments only)
glab auth login --hostname gitlab.k8s.rommelporras.com \
  --token "$(op read 'op://Kubernetes/Gitlab/personal-access-token')"

# Atuin (if configured)
atuin login -u <account-name> \
  -p "$(op read 'op://Kubernetes/Atuin/<context>-password')" \
  -k "$(op read 'op://Kubernetes/Atuin/encryption-key')"
```

### 3. Aurora DX only: set up Distrobox containers

Ensure 1Password desktop app is unlocked and CLI integration is enabled
(Settings → Developer → **Integrate with 1Password CLI**). This lets `op` commands
authenticate via biometric/system password instead of manual signin.

Then create containers:

```bash
cd ~/personal/dotfiles

# Personal container — only needs personal email
uv run python scripts/distrobox_setup.py personal \
  --personal-email git@rommelporras.com

# Work container — only needs work email
uv run python scripts/distrobox_setup.py work-eam \
  --work-email work@company.com

# All default containers (provide both for mixed contexts)
uv run python scripts/distrobox_setup.py \
  --personal-email git@rommelporras.com \
  --work-email work@company.com
```

Each context only requires its relevant email flag — the rest is derived from the
container name. Without the flag, chezmoi prompts interactively. Sandbox is always
non-interactive (no flags needed).

See [docs/distrobox-scripts.md](docs/distrobox-scripts.md) for full parameter reference,
config derivation table, and verification steps.

Container home directories persist at `~/.distrobox/<name>/` on the host — removing
and recreating a container preserves your data.

```bash
distrobox enter personal                # Enter a container
~/bin/chezmoi apply -v && exec zsh      # Update dotfiles inside a container
```

**IDE forwarding:** `code` and `agy` (Antigravity) commands inside non-sandbox
containers are forwarded to the Aurora host via `distrobox-host-exec`.

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
| Daily dev with cluster access | `distrobox enter personal` | Needs kubeconfig, SSH keys |
| Project-scoped dev (e.g., fintrack) | `distrobox enter personal-fintrack` | Native `op`, Bun, Playwright, no homelab |
| Vibe-coding with git push | `ai-sandbox --git` | Contained, can push via deploy key |
| Trying untrusted AI tool | `ai-sandbox --no-network` | No network, can't exfiltrate |
| Work Terraform/EKS | `distrobox enter work-eam` | Needs AWS + EKS credentials |
| Quick experiment | `distrobox enter sandbox` | Clean space, no credentials |

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
│   ├── .chezmoiexternal.toml  # External deps (oh-my-zsh, zsh-autosuggestions)
│   ├── .chezmoiignore         # Per-environment file skipping
│   ├── dot_zshrc.tmpl         # Shell config
│   ├── dot_gitconfig.tmpl     # Git config (conditional includes)
│   ├── run_once_before_bootstrap.sh.tmpl  # First-run setup script
│   ├── dot_local/bin/         # User scripts (~/.local/bin/)
│   │   └── setup-creds       # Credential + plugin seeding for Distrobox
│   └── dot_config/            # ~/.config/ files
│       ├── starship.toml
│       ├── atuin/config.toml.tmpl
│       ├── ghostty/config
│       ├── k9s/config.yaml
│       └── git/               # Git identity includes + global gitignore
├── bin/                       # CLI tools (ai-sandbox)
├── containers/                # Containerfile.ai-sandbox + distrobox.ini
├── scripts/                   # Setup + testing automation
│   ├── distrobox_setup.py     # Container creation + chezmoi bootstrap
│   ├── distrobox_lib.py       # Shared library for distrobox scripts
│   ├── test_distrobox_integration.py  # E2E test (delete → create → bootstrap → verify → delete)
│   └── windows-git-setup.ps1  # Windows git setup for WSL
├── docs/                      # Reference documentation
│   └── distrobox-scripts.md   # Distrobox scripts parameter reference
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

Claude Code's global configuration (`~/.claude/`) is managed by a separate
[claude-config](https://github.com/rommelporras/claude-config) repo — cloned to
`~/personal/claude-config` and symlinked into `~/.claude/` by the bootstrap script.

On distrobox containers, symlinks point to the host's clone via absolute paths
(`/home/<user>/personal/claude-config/`). The `.claude/` directory is blanket-ignored
in `.chezmoiignore` so chezmoi never touches it.

### What's in claude-config

| File | Purpose |
|---|---|
| `CLAUDE.md` | Universal global instructions (same across all environments) |
| `settings.json` | Permission deny rules, hooks, enabled plugins, marketplace declarations |
| `hooks/` | Security hooks: secret scanning, write protection, destructive command blocking |
| `agents/` | Custom agents (code-reviewer) |
| `skills/` | Skills (/commit, /push, /explain-code) |

### Plugin and MCP setup

Plugins and MCP servers require CLI commands (not just config files):

- **Distrobox:** `setup-creds` handles marketplace registration, plugin installation, and
  Context7 MCP setup automatically via `distrobox-host-exec op` for the API key.
- **Aurora/WSL:** Run the `claude plugin` and `claude mcp` commands manually after bootstrap
  (see [credential setup](#2-set-up-credentials-per-machine) for the exact commands).

## Testing

Integration tests verify the full distrobox lifecycle: delete → create → bootstrap →
verify → delete. 77 assertions across 4 container types.

```bash
cd ~/personal/dotfiles

uv run python scripts/test_distrobox_integration.py --all     # All containers
uv run python scripts/test_distrobox_integration.py personal  # Single container
uv run python scripts/test_distrobox_integration.py --keep personal  # Keep for inspection
```

Tests pre-seed all chezmoi config (no interactive prompts, no 1Password dependency).
See [docs/distrobox-scripts.md](docs/distrobox-scripts.md) for assertion counts per
container and full parameter reference.

## History Migration

When setting up Atuin on a new machine with existing shell history:

1. Review history for sensitive entries: `less ~/.zsh_history`
2. Clean passwords or tokens that were typed in the shell
3. Import into Atuin: `atuin import zsh`

## License

MIT License. See [LICENSE](LICENSE) for details.
