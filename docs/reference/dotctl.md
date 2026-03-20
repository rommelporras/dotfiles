# dotctl Reference

## Installation

```bash
cd ~/personal/dotfiles
make install          # copies binary to ~/.local/bin/dotctl
make install-systemd  # installs + enables systemd timer (10-minute collection)
make uninstall-systemd  # disables + removes systemd timer
```

## Commands

### dotctl status

Query Prometheus + Loki and render a terminal dashboard.

```bash
dotctl status                    # query from homelab cluster
dotctl status --live             # collect locally, no cluster needed
dotctl status --machine aurora   # filter to one machine
```

Falls back to `--live` automatically if Prometheus is unreachable.

Dashboard sections: Machines (hostname, platform/context, drift count), Drift Details
(per-file status), Tools Grid (8 tools × machines), Credentials (SSH agent, setup-creds,
Atuin sync), Claude Config (6 symlink statuses).

### dotctl collect

Collect status from local machine + running Distrobox containers, push to OTel Collector.

```bash
dotctl collect                   # silent unless errors
dotctl collect --verbose         # print per-machine status + push results
dotctl collect --container work-eam  # collect from one container only
```

## What It Tracks

### Tools

`glab`, `kubectl`, `terraform`, `aws`, `ansible`, `op`, `atuin`, `bun`

### Claude Config Symlinks

`CLAUDE.md`, `settings.json`, `rules`, `hooks`, `skills`, `agents`

Each symlink is checked in `~/.claude/` and reported as:
- `ok` — symlink pointing to a path containing `claude-config`
- `wrong` — symlink to an unexpected target
- `file` — regular file (not a symlink)
- `missing` — file does not exist

Skipped when chezmoi is not available (e.g., AI sandbox containers).

### Credentials

- **SSH Agent** — `1password` (socket path contains "1password"), `system`, or `none`
- **setup-creds** — `ran` (executable), `present` (exists but not executable), `n/a`
- **Atuin Sync** — `synced` (config has sync_address), `disabled`, `n/a`

## Config

`~/.config/dotctl/config.toml` — optional, all fields have defaults.

```toml
otel_endpoint   = "10.10.30.22:4317"
prometheus_url  = "https://prometheus.k8s.rommelporras.com"
loki_url        = "https://loki.k8s.rommelporras.com"
hostname        = ""    # leave blank for auto-detection
```

## Systemd Timer

```bash
systemctl --user status dotctl-collect.timer   # check timer
systemctl --user list-timers                   # see next run
journalctl --user -u dotctl-collect.service    # view logs
```

Runs every 10 minutes, 2 minutes after boot, with ±30s jitter.
