# dotctl Architecture

## What It Is

Go CLI tool in the dotfiles monorepo (`dotctl/`) that collects dotfiles status
from Aurora + running Distrobox containers, pushes metrics and logs to the
homelab OTel Collector, and renders a terminal dashboard.

## Three Modes

| Command | What it does |
|---|---|
| `dotctl collect` | Collect from local + containers, push via OTLP gRPC |
| `dotctl status` | Query Prometheus + Loki, render dashboard |
| `dotctl status --live` | Collect locally, render directly (no cluster needed) |

## Package Layout

```
dotctl/
├── cmd/dotctl/main.go     — cobra CLI wiring only
├── internal/
│   ├── model/             — domain types (MachineState, DriftFile, ContainerInfo)
│   ├── config/            — loads ~/.config/dotctl/config.toml, fills defaults
│   ├── collector/         — chezmoi status, tool probe, credentials, distrobox
│   ├── push/              — OTLP gRPC push (metrics + logs)
│   ├── query/             — Prometheus + Loki HTTP API clients
│   └── display/           — lipgloss terminal table rendering
└── deploy/                — systemd service + timer units
```

## Collection Strategy

- **Local machine**: shells out to `chezmoi status`, `chezmoi data --format json`,
  `which <tool>`, reads `$SSH_AUTH_SOCK`, checks `~/.local/bin/setup-creds`,
  reads `~/.config/atuin/config.toml`
- **Distrobox containers**: `distrobox enter <name> -- sh -c "..."` with 30s timeout.
  Runs chezmoi, tools probe, and credential checks inside each running container.
  No binary needed inside containers.

## Metrics (Prometheus via OTel Collector)

All gauges, pushed via OTLP gRPC to `10.10.30.22:4317`:

| Metric | Labels |
|---|---|
| `dotctl_up` | hostname, platform, context |
| `dotctl_collect_timestamp` | hostname |
| `dotctl_drift_total` | hostname |
| `dotctl_tool_installed` | hostname, tool |
| `dotctl_credential_status` | hostname, credential |
| `dotctl_container_running` | name |

## Logs (Loki via OTel Collector)

One structured JSON entry per machine per collection. Full `MachineState`
serialized as the log body. Query: `{service_name="dotctl"}`.

## Config File

`~/.config/dotctl/config.toml` — missing file returns defaults silently.

```toml
otel_endpoint   = "10.10.30.22:4317"
prometheus_url  = "https://prometheus.k8s.rommelporras.com"
loki_url        = "https://loki.k8s.rommelporras.com"
hostname        = "aurora-dx"   # optional, auto-detected if absent
```

## Key Design Decisions

- **Monorepo**: dotctl lives in the dotfiles repo because it is tightly coupled to
  chezmoi, distrobox, and the specific environment model.
- **Shell out, don't reimplement**: collection uses `chezmoi` and `distrobox` CLIs
  via `CommandRunner` interface (mockable in tests).
- **OTel Collector reuse**: zero new K8s infrastructure — existing Collector at
  `10.10.30.22:4317` accepts all OTLP. Metrics appear on `:8889`, scraped by
  existing ServiceMonitor every 60s.
- **In-memory metrics**: OTel Collector holds metrics in RAM. Lost on restart,
  recovered within 10 minutes. `DotctlCollectionStale` alert uses
  `absent_over_time(35m)` to account for this.
- **Display alignment**: `fmt.Sprintf("%-Ns", styledString)` breaks on ANSI codes
  because it counts escape bytes as width. Fixed with `lipgloss.Width()` helper.
