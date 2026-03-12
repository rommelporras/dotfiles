# dotctl Reference

## Installation

```bash
cd ~/personal/dotfiles
make install          # copies binary to ~/.local/bin/dotctl
make install-systemd  # installs + enables systemd timer (10-minute collection)
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

### dotctl collect

Collect status from local machine + running Distrobox containers, push to OTel Collector.

```bash
dotctl collect                   # silent unless errors
dotctl collect --verbose         # print per-machine status + push results
dotctl collect --container work-eam  # collect from one container only
```

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

## Tracked Tools

`glab`, `kubectl`, `terraform`, `aws`, `ansible`, `op`, `atuin`, `bun`
