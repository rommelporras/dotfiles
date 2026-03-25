# Credential Setup

## Distrobox Containers (automated)

Run inside any distrobox container:

```bash
setup-creds
```

Handles: Claude Code plugins, Context7 MCP, Atuin login, glab auth, kubeconfig (manual step), AWS (manual step). Pulls secrets from 1Password on the Aurora host via `distrobox-host-exec op`.

Requires: 1Password desktop app unlocked with CLI integration enabled
(Settings → Developer → Integrate with 1Password CLI).

## Aurora DX (manual)

```bash
# SSH public keys
cp id_ed25519.pub ~/.ssh/
chmod 644 ~/.ssh/*.pub
ssh-add -l    # verify 1Password agent works

# Claude Code plugins (marketplaces first, then install)
claude plugin marketplace add anthropics/claude-plugins-official
claude plugin marketplace add obra/superpowers-marketplace
claude plugin marketplace add nextlevelbuilder/ui-ux-pro-max-skill
claude plugin install context7@claude-plugins-official --scope user
claude plugin install playwright@claude-plugins-official --scope user
claude plugin install superpowers@superpowers-marketplace --scope user
claude plugin install episodic-memory@superpowers-marketplace --scope user
claude plugin install ui-ux-pro-max@ui-ux-pro-max-skill --scope user

# Context7 MCP
claude mcp add --scope user --transport http context7 https://mcp.context7.com/mcp \
  --header "CONTEXT7_API_KEY: $(op read 'op://Kubernetes/Context7/api-key' --no-newline)"

# Homelab kubeconfigs — automated by setup-creds
# (pulls from 1Password "Kubeconfig" item in Kubernetes vault)

# GitLab
glab auth login --hostname gitlab.k8s.rommelporras.com \
  --token "$(op read 'op://Kubernetes/Gitlab/personal-access-token')"

# Atuin
atuin login -u <account> \
  -p "$(op read 'op://Kubernetes/Atuin/<context>-password')" \
  -k "$(op read 'op://Kubernetes/Atuin/encryption-key')"

# GitHub CLI
gh auth login
```

## WSL2 (manual)

Same commands as Aurora — `op` CLI works in WSL via socket bridge to the Windows
1Password desktop app. Requires 1Password CLI installed in WSL and "Integrate with
1Password CLI" enabled in the Windows 1Password app (see [WSL2 setup](../setup/wsl2.md)
step 1 for install instructions).

Additionally, set up GitHub CLI:

```bash
gh auth login
```

## AI Sandbox — Podman Secrets

```bash
op read 'op://Kubernetes/Anthropic/api-key' | tr -d '\n' | podman secret create anthropic_key -
op read 'op://Kubernetes/Gemini/api-key' | tr -d '\n' | podman secret create gemini_key -
op read 'op://Kubernetes/Context7/api-key' | tr -d '\n' | podman secret create context7_key -

# Atuin auto-login in sandbox containers
op read 'op://Kubernetes/Atuin/personal-password' | tr -d '\n' | podman secret create atuin_password -
op read 'op://Kubernetes/Atuin/encryption-key' | tr -d '\n' | podman secret create atuin_key -

# Deploy key for --git flag
ssh-keygen -t ed25519 -f ~/.ssh/ai-deploy-key -C 'ai-sandbox-deploy'
# Add .pub to GitHub/GitLab as deploy key
```

`atuin_password` and `atuin_key` are injected as `ATUIN_PASSWORD` and `ATUIN_KEY` env
vars at container start. On the first shell login, the sandbox `.zshrc` automatically
calls `atuin login -u personal` using these values, then unsets the env vars.
The `tr -d '\n'` is required — podman secrets must not have a trailing newline.
