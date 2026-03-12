# Distrobox Setup

Distrobox containers are set up from the Aurora host. Each container gets its own
home directory at `~/.distrobox/<name>/` — persists across container recreation.

## Create containers

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

# Sandbox (no flags needed)
uv run python scripts/distrobox_setup.py sandbox
```

See [docs/reference/distrobox-scripts.md](../reference/distrobox-scripts.md) for full parameter reference.

## Day-to-day

```bash
distrobox enter personal        # enter a container
~/bin/chezmoi apply -v          # update dotfiles inside container
exec zsh                        # reload shell
```

## IDE forwarding

`code` and `agy` (Antigravity) inside non-sandbox containers are forwarded
to the Aurora host via `distrobox-host-exec`. No IDE installation needed in containers.

## Credentials

Run inside any non-sandbox container after first bootstrap:
```bash
setup-creds
```

See [docs/reference/credentials.md](../reference/credentials.md).
