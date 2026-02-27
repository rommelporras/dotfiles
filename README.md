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

### 1. Install chezmoi and apply dotfiles

```bash
# Cache sudo credentials first — bootstrap installs packages via apt
sudo -v

# Install chezmoi + clone this repo + run interactive prompts + apply
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
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

The bootstrap script automatically installs: zsh, Starship, fzf, xclip, and
environment-specific tools (NVM, Bun, Terraform, glab, Ansible) based on your answers.

**After install, restart your shell:**

```bash
exec zsh
```

### Font setup (Nerd Font)

The prompt uses [Nerd Font](https://www.nerdfonts.com/) icons. The bootstrap
installs JetBrainsMono Nerd Font automatically on Aurora/Distrobox. On WSL,
fonts render on the Windows side — install manually:

1. Download **JetBrainsMono** from https://www.nerdfonts.com/font-downloads
2. Extract the zip, select all `.ttf` files, right-click → **Install**
3. In Windows Terminal: Settings → Profile → Appearance → Font face → `JetBrainsMono Nerd Font`

### 2. Set up credentials (manual, per machine)

The bootstrap creates the directories — you populate them:

```bash
# SSH keys (copy from 1Password)
chmod 600 ~/.ssh/id_*

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
~/.local/share/chezmoi/scripts/distrobox-setup.sh
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
chezmoi update        # Pull latest from GitHub + apply in one step
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

The `--git` flag mounts `~/.ssh/ai-deploy-key` (not your personal `id_ed25519`).
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

**Recommended:** Ottoson (perceptually uniform, good contrast on dark backgrounds)

| Terminal | How to set |
|---|---|
| Windows Terminal (WSL) | Settings → Color schemes → **Ottoson** |
| Ptyxis (Aurora DX) | Preferences → pick closest match, or import custom palette |

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
│   └── dot_config/            # ~/.config/ files
│       ├── starship.toml
│       ├── atuin/config.toml.tmpl
│       ├── ghostty/config
│       ├── k9s/config.yaml
│       └── git/               # Git identity includes
├── bin/                       # CLI tools (ai-sandbox)
├── containers/                # Containerfile + distrobox.ini
├── scripts/                   # Setup automation (distrobox-setup.sh)
└── hooks/                     # Git hooks (gitleaks pre-commit)
```

### chezmoi naming conventions

| Prefix/suffix | Meaning |
|---|---|
| `dot_` | Becomes `.` in target (`dot_zshrc` → `.zshrc`) |
| `private_dot_` | `.` with 0600 permissions |
| `.tmpl` | Processed as Go text/template |
| `run_once_before_` | Script that runs once, before file creation |

## History Migration

When setting up Atuin on a new machine with existing shell history:

1. Review history for sensitive entries: `less ~/.zsh_history`
2. Clean passwords or tokens that were typed in the shell
3. Import into Atuin: `atuin import zsh`

## License

MIT License. See [LICENSE](LICENSE) for details.
