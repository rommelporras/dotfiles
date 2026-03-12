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

## Keeping in sync

Container chezmoi source is symlinked to the host repo (`~/personal/dotfiles`).
When the host pulls new changes, containers automatically see them — no separate
git pull needed inside containers.

**On the Aurora host:**
```bash
cd ~/personal/dotfiles
git pull
chezmoi apply     # update Aurora host dotfiles
exec zsh
```

**Inside each container (after host pulls):**
```bash
~/bin/chezmoi apply -v    # pick up changes from the now-updated host repo
exec zsh
```

If the bootstrap script changed (new tools added), re-run it inside the container:
```bash
~/bin/chezmoi state delete-bucket --bucket=scriptState
~/bin/chezmoi apply
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

## dotctl collection timer

To enable automatic metric collection from the Aurora host (runs every 10 minutes):
```bash
cd ~/personal/dotfiles
make install-systemd
```
