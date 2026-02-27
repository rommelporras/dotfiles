# dotfiles

chezmoi-managed dotfiles for consistent dev environments across WSL2 Ubuntu,
Aurora DX, and Distrobox containers.

## Quick Start

### New machine (WSL2 or Aurora DX host)

```bash
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
```

chezmoi will prompt for your environment type and preferences. Restart your shell after.

### Aurora DX: Set up Distrobox containers

```bash
~/.local/share/chezmoi/scripts/distrobox-setup.sh
```

### Aurora DX: Build AI sandbox

```bash
ai-sandbox --build
```

## Credential Setup (Per Machine)

After bootstrap, configure credentials manually:

| Step | Command |
|---|---|
| SSH keys | Copy from 1Password → `~/.ssh/`, then `chmod 600 ~/.ssh/id_*` |
| AI deploy key | `ssh-keygen -t ed25519 -f ~/.ssh/ai-deploy-key -C 'ai-sandbox-deploy'` then add `.pub` to GitHub/GitLab as deploy key |
| AWS | `aws sso login --profile <name>` |
| EKS kubeconfig | `aws eks update-kubeconfig --name <cluster> --region <region>` (writes to `~/.kube/config`) |
| Homelab kubeconfig | Copy `homelab.yaml` → `~/.kube/` |
| GitLab | `glab auth login --hostname gitlab.k8s.rommelporras.com` |
| GitHub | `gh auth login` |
| AI sandbox API keys | `podman secret create anthropic_key <(echo "sk-ant-...")` |
| Atuin | `atuin login -u <account-name>` |

## AI Sandbox Usage

Run AI coding agents in an isolated Podman container:

```bash
# Code only — no git push, no credentials
ai-sandbox claude -- --dangerously-skip-permissions

# Code + git push via deploy key (for vibe-coding projects like Fintrack)
ai-sandbox --git claude -- --dangerously-skip-permissions

# Maximum containment — no network, can't exfiltrate anything
ai-sandbox --no-network gemini

# Any AI tool works (generic command)
ai-sandbox --git aider
ai-sandbox --git antigravity

# Debugging
ai-sandbox --shell
```

The sandbox mounts the context group (~/personal/ or ~/eam/) based on your
current directory. No access to host HOME, personal SSH keys, or cloud credentials.
The --git flag mounts a dedicated deploy key (not your personal id_ed25519).

### When to use sandbox vs Distrobox directly

| Scenario | Where to run | Why |
|---|---|---|
| Homelab ops (kubectl, helm) | `distrobox-personal` | Needs cluster creds |
| Vibe-coding (trusted AI, needs push) | `ai-sandbox --git` | Contained, can push |
| Trying untrusted AI tool | `ai-sandbox --no-network` | Maximum containment |
| Work Terraform/EKS | `distrobox-work` | Needs AWS + EKS creds |

## History Migration (from existing WSL2)

Before syncing Atuin to a new machine:

1. Review current history: `less ~/.zsh_history`
2. Clean sensitive entries (passwords, tokens typed in shell)
3. Import into Atuin: `atuin import zsh`
4. Or copy directly: `cp ~/.zsh_history ~/.distrobox/<container>/.zsh_history`

## Updating Dotfiles

```bash
chezmoi update        # Pull latest from GitHub + apply
chezmoi diff          # Preview changes before applying
chezmoi apply -v      # Apply with verbose output
```

## Environments

| Value | Machine | Description |
|---|---|---|
| `wsl-work` | Work laptop (Intel Ultra 7, 32GB) | WSL2 Ubuntu, work + personal |
| `wsl-gaming` | Gaming desktop (Ryzen 5800X, 32GB) | WSL2 Ubuntu, work + personal |
| `aurora` | Personal laptop (i5-8th gen, 16GB) | Aurora DX host, launches containers |
| `distrobox-work` | Aurora container | Work projects, work credentials |
| `distrobox-personal` | Aurora container | Personal projects, homelab credentials |
| `distrobox-sandbox` | Aurora container | Clean experiment space |
