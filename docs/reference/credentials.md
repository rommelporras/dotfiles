# Credential Setup

## Distrobox Containers (automated)

Run inside any non-sandbox container:

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
claude plugin install context7@claude-plugins-official --scope user
claude plugin install playwright@claude-plugins-official --scope user
claude plugin install superpowers@superpowers-marketplace --scope user
claude plugin install episodic-memory@superpowers-marketplace --scope user

# Context7 MCP
claude mcp add --scope user --transport http context7 https://mcp.context7.com/mcp \
  --header "CONTEXT7_API_KEY: $(op read 'op://Kubernetes/Context7/api-key' --no-newline)"

# Homelab kubeconfig
cp homelab.yaml ~/.kube/

# GitLab
glab auth login --hostname gitlab.k8s.rommelporras.com \
  --token "$(op read 'op://Kubernetes/Gitlab/personal-access-token')"

# Atuin
atuin login -u <account> \
  -p "$(op read 'op://Kubernetes/Atuin/<context>-password')" \
  -k "$(op read 'op://Kubernetes/Atuin/encryption-key')"
```

## WSL2 (manual)

Same as Aurora, plus GitHub CLI:

```bash
gh auth login
```

## AI Sandbox — Podman Secrets

```bash
podman secret create anthropic_key <(echo "sk-ant-...")
podman secret create gemini_key <(echo "AI...")

# Deploy key for --git flag
ssh-keygen -t ed25519 -f ~/.ssh/ai-deploy-key -C 'ai-sandbox-deploy'
# Add .pub to GitHub/GitLab as deploy key
```
