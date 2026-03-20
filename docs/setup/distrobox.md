# Distrobox Setup

Distrobox containers are set up from the Aurora host. Each container gets its own
home directory at `~/.distrobox/<name>/` — persists across container recreation.

## Prerequisites

Complete [Aurora DX setup](aurora.md) first. You need:
- **Distrobox** — installed by `ujust devmode`
- **uv** — installed via `brew install uv`
- **1Password unlocked** with CLI integration enabled
  (Settings → Developer → Integrate with 1Password CLI)

## Create containers

```bash
cd ~/personal/dotfiles

# Personal container
uv run python scripts/distrobox_setup.py personal \
  --personal-email git@rommelporras.com

# Work container
uv run python scripts/distrobox_setup.py work-eam \
  --work-email work@company.com

# All default containers (work-eam, personal)
uv run python scripts/distrobox_setup.py \
  --personal-email git@rommelporras.com \
  --work-email work@company.com
```

The setup script installs chezmoi, symlinks the host repo, runs `chezmoi init --apply`,
and runs `setup-creds` automatically. Tools installed per context:

| Context | Key tools |
|---|---|
| `personal` | glab, ansible, kubectl, Atuin |
| `personal-<project>` | glab, native op CLI, bun, Playwright, Atuin |
| `work-<name>` | terraform, AWS CLI, kubectl, Atuin |

> For vibe-coded AI projects, use `ai-sandbox` instead — see `bin/ai-sandbox --help`.

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

## IDE and browser forwarding

`code`, `antigravity`, and `agy` inside containers are forwarded to the Aurora
host via `distrobox-host-exec`. No IDE installation needed in containers.

`BROWSER` is set to `distrobox-host-exec xdg-open` — OAuth flows and link-opening
work transparently from inside containers.

## Ptyxis profiles

Terminal profiles are created automatically during container setup. After setup,
close all Ptyxis windows and reopen — the new profiles will appear in the tab menu.

Profiles are also removed automatically during cleanup (`distrobox_cleanup.py`).

## Credentials

`setup-creds` runs automatically during container setup. To re-run it (e.g. after
1Password was locked during initial setup):
```bash
setup-creds
```

See [docs/reference/credentials.md](../reference/credentials.md).

## Removing containers

```bash
cd ~/personal/dotfiles

# Remove a single container (keeps home directory for re-creation)
uv run python scripts/distrobox_cleanup.py personal

# Remove and wipe home directory
uv run python scripts/distrobox_cleanup.py personal --wipe-home

# Remove all containers defined in distrobox.ini
uv run python scripts/distrobox_cleanup.py --all

# Remove all containers and wipe all home directories
uv run python scripts/distrobox_cleanup.py --all --wipe-home
```

## dotctl collection timer

To enable automatic metric collection from the Aurora host (runs every 10 minutes):
```bash
cd ~/personal/dotfiles
make install-systemd
```
