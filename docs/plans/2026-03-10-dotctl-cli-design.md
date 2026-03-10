# dotctl — Dotfiles Status CLI

Design document for a CLI tool that provides observability into chezmoi-managed
dotfiles across Aurora DX, Distrobox containers, and (future) WSL2/Windows.

## Problem

Managing dotfiles across 1 Aurora host + 4 Distrobox containers + WSL2 with
context-specific credentials and tools. No single view of what's deployed,
what's drifted, or what tools are installed across environments.

## Decision Record

| Decision | Choice | Rationale |
|---|---|---|
| Scope | Aurora + Distrobox + future WSL2/Windows | Cross-machine is the real value |
| Language | Go | Single binary, cross-compilation, no runtime deps |
| Data shown | Drift, container state, template data, tools, credentials | All were requested |
| Storage | OTel Collector → Prometheus (metrics) + Loki (logs) | Zero new infra, existing pipeline |
| Collection | Periodic (systemd timer) + on-demand CLI | Grafana dashboard + live queries |
| Architecture | Single binary, three modes (status / status --live / collect) | Approach B — graceful degradation when Prometheus is unreachable |

## CLI Interface

```
dotctl status [--live] [--machine <name>]
dotctl collect [--container <name>] [--verbose]
```

### `dotctl status`

Queries Prometheus + Loki to render a terminal dashboard:

- **Machines table** — hostname, platform/context, last seen, drift count
- **Drift details** — per-machine list of modified/added/deleted files
- **Tools grid** — which tools are installed per machine (glab, kubectl, terraform, etc.)
- **Credentials grid** — SSH agent type, setup-creds status, Atuin sync status

Falls back to `--live` mode with a warning if Prometheus is unreachable.

### `dotctl status --live`

Collects fresh data locally (host + running containers) and displays directly.
No Prometheus/Loki dependency. Works anywhere the binary runs.

### `dotctl status --machine <name>`

Filters output to a single machine with expanded detail.

### `dotctl collect`

Gathers status from the local machine + all running Distrobox containers.
Pushes metrics to OTel Collector via OTLP gRPC. Designed for systemd timer.
Quiet output unless `--verbose`.

Container data collected via `distrobox enter <name> -- sh -c "..."` — no
binary installation required inside containers.

## Data Model

### Prometheus (low-cardinality gauges — ~88 time series total)

| Metric | Labels | Purpose |
|---|---|---|
| `dotctl_up` | hostname, platform, context | Machine is reporting |
| `dotctl_drift_total` | hostname | Count of drifted files |
| `dotctl_tool_installed` | hostname, tool | Tool presence (0/1) |
| `dotctl_credential_status` | hostname, credential | Credential state (0/1) |
| `dotctl_container_running` | name | Distrobox container up/down |
| `dotctl_collect_timestamp` | hostname | Unix time of last collection |

### Loki (structured JSON logs — detailed state)

One log entry per machine per collection:

```json
{
  "service_name": "dotctl",
  "hostname": "aurora-dx",
  "platform": "aurora",
  "context": "personal",
  "drift_files": [
    {"path": "~/.zshrc", "status": "M"}
  ],
  "template_data": {
    "atuin_account": "personal",
    "has_homelab_creds": true
  },
  "tools": {
    "glab": "/usr/bin/glab",
    "kubectl": "/usr/local/bin/kubectl"
  },
  "ssh_agent": "1password",
  "setup_creds": "n/a"
}
```

### Storage impact

- Prometheus: ~17KB/day (88 series × 1 sample/15s scrape)
- Loki: <1MB/day (6 machines × 144 entries/day × ~1KB each)
- Both within existing PVCs (50Gi Prometheus, 10Gi Loki) with 90-day auto-retention

### Collection flow

```
dotctl collect
  ├── Detect platform (aurora/distrobox/wsl)
  ├── Collect local machine status
  │     ├── chezmoi status → parse drift
  │     ├── chezmoi data --format json → template variables
  │     ├── command -v <tool> → tool inventory
  │     └── Check SSH_AUTH_SOCK, setup-creds, atuin config
  ├── If aurora: enumerate distrobox containers
  │     └── For each running container:
  │           distrobox enter <name> -- sh -c "<collection commands>"
  └── Push via OTLP gRPC to OTel Collector
```

### Expected tools per context

| Tool | personal | personal-* | work-* | sandbox | aurora |
|---|---|---|---|---|---|
| glab | yes | yes | no | no | yes |
| kubectl | yes | no | yes | no | yes |
| terraform | no | no | yes | no | no |
| aws | no | no | yes | no | no |
| ansible | yes | no | no | no | no |
| op | no | yes | no | no | yes |
| atuin | yes | yes | yes | no | yes |
| bun | no | yes | no | no | no |

## Project Structure

```
dotctl/                          # New repo: github.com/rommelporras/dotctl
├── cmd/dotctl/main.go           # CLI entrypoint
├── internal/
│   ├── collector/
│   │   ├── collector.go         # Orchestrates collection
│   │   ├── chezmoi.go           # chezmoi status/data parsing
│   │   ├── tools.go             # Tool installation probes
│   │   ├── credentials.go       # SSH agent, setup-creds, atuin checks
│   │   └── distrobox.go         # Container enumeration + remote commands
│   ├── push/
│   │   └── otel.go              # OTLP gRPC push (metrics + logs)
│   ├── query/
│   │   ├── prometheus.go        # PromQL via HTTP API
│   │   └── loki.go              # LogQL via HTTP API
│   └── display/
│       └── table.go             # Terminal table rendering
├── go.mod
├── Makefile                     # build, lint, test, install
├── CLAUDE.md
└── README.md
```

### Dependencies

| Package | Purpose |
|---|---|
| `go.opentelemetry.io/otel` | OTLP SDK for metrics + logs |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/spf13/cobra` | CLI subcommands |
| stdlib `net/http` | Prometheus/Loki HTTP API queries |

### Config

`~/.config/dotctl/config.toml`:

```toml
otel_endpoint = "10.10.30.22:4317"
prometheus_url = "https://prometheus.k8s.rommelporras.com"
loki_url = "https://loki.k8s.rommelporras.com"
hostname = "aurora-dx"
```

## Systemd Integration

User-level systemd timer on Aurora (no root):

```
~/.config/systemd/user/dotctl-collect.service
~/.config/systemd/user/dotctl-collect.timer    # every 10 minutes
```

Binary installed to `~/.local/bin/dotctl`. `make install` handles binary + systemd units.

Future WSL2: same binary, cron or Windows Task Scheduler. OTel endpoint
reachable via Tailscale if not on LAN.

## Error Handling

| Scenario | Behavior |
|---|---|
| OTel Collector unreachable | Warning, exit non-zero. Timer retries next interval. |
| Prometheus unreachable | Auto-fallback to `--live` mode with warning |
| Container stopped | Reported as `stopped`. No collection attempted. Not an error. |
| chezmoi not installed | Skip drift/template data, still report tools/creds. Warning. |
| `distrobox enter` timeout | 30s per container. Skip with warning, continue. |
| Config file missing | Auto-detect hostname/platform. Only endpoints required for non-live. |

---

## Infrastructure Section

**For homelab repo review** — this section describes the infra changes needed
on the K8s cluster and presents alternatives for the homelab Claude session to
evaluate.

### Preferred approach: Reuse existing OTel Collector pipeline

dotctl pushes OTLP metrics + logs to the existing OTel Collector at
`10.10.30.22:4317`. The collector already routes metrics → Prometheus and
logs → Loki. No new deployments, services, or storage.

**Required changes:**

1. **OTel Collector config** — currently the collector accepts all OTLP data.
   No filtering changes needed. dotctl metrics arrive with `service_name=dotctl`
   and are distinguishable from Claude Code telemetry by metric prefix (`dotctl_*`).

2. **ServiceMonitor** — the existing OTel Collector ServiceMonitor already
   scrapes the Prometheus exporter endpoint (`:8889/metrics`). dotctl metrics
   will appear there automatically. No changes needed.

3. **Grafana dashboard** — new ConfigMap in `manifests/monitoring/dashboards/`
   with label `grafana_dashboard: "1"`. Auto-provisioned by sidecar. Dashboard
   panels:
   - Machine status grid (up/down/drifted)
   - Drift count over time per machine
   - Tool inventory matrix
   - Credential status
   - Collection health (last successful collect timestamp)

4. **PrometheusRules (optional)** — alert definitions:
   - `DotctlCollectionStale` — no collection from a machine in >30 minutes
   - `DotctlDriftDetected` — `dotctl_drift_total > 0` for >1 hour
   - Route to Discord #status (warning severity)

5. **Loki access** — dotctl CLI queries Loki via HTTP API at
   `https://loki.k8s.rommelporras.com`. Currently Loki is internal-only.
   Options:
   - (a) Add an HTTPRoute for Loki (like Grafana/Prometheus already have)
   - (b) Query via Grafana's Loki proxy API (`/api/datasources/proxy/`)
   - (c) Use Tailscale to reach Loki's ClusterIP directly
   - Homelab session should decide based on security posture.

**Storage impact on cluster:**
- Prometheus: +88 time series (~17KB/day). Current usage is well within 50Gi PVC.
- Loki: +~1MB/day structured logs. Current usage is well within 10Gi PVC.
- No PVC resizing or retention changes needed.

### Alternative A: Dedicated dotctl deployment on K8s

Deploy a lightweight Go HTTP server in the `monitoring` namespace that:
- Receives state pushes from dotctl agents via REST API
- Stores state in a PVC (JSON files or SQLite)
- Exposes a `/metrics` endpoint for Prometheus to scrape
- Logs state to stdout for Alloy → Loki collection

**Components:**
- Deployment (1 replica, ~64Mi memory)
- Service (ClusterIP)
- HTTPRoute (for external access from Aurora/WSL2)
- PVC (1Gi, Longhorn)
- ServiceMonitor

**Pros:**
- Decoupled from OTel Collector — dotctl has its own endpoint
- Could add REST API features later (query state directly without Prometheus)
- Stdout/stderr naturally collected by Alloy

**Cons:**
- New deployment to maintain, monitor, and upgrade
- Duplicates what OTel Collector already does (receive data, route to Prometheus/Loki)
- Adds a custom API that needs versioning and error handling
- PVC lifecycle management (backup, retention)

### Alternative B: GitHub private repo as state store

dotctl pushes state as JSON commits to a private GitHub repo. CLI reads
the latest commit.

**Pros:**
- Zero K8s infrastructure
- Git history = free versioning
- Works when cluster is down

**Cons:**
- ~52K commits/year (every 10 min) — git history bloat, needs periodic gc
- Merge conflicts when multiple machines push concurrently
- GitHub token management on every machine
- No Grafana integration without a custom exporter
- No alerting capability
- More maintenance than Prometheus/Loki despite appearing simpler

### Recommendation for homelab review

The preferred approach (reuse OTel Collector) adds zero new deployments and
~88 time series to Prometheus. The only infrastructure work is:
1. A Grafana dashboard ConfigMap
2. Optional PrometheusRules for staleness/drift alerts
3. Possibly an HTTPRoute for Loki (if not using Grafana proxy)

Alternatives A and B are presented for completeness but add infrastructure
and maintenance burden that the preferred approach avoids.
