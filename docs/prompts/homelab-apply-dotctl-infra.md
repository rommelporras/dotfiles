# Prompt: Apply dotctl Infrastructure on Homelab K8s

> Run this in a Claude Code session inside the homelab repo (`~/personal/homelab/`).

## Context

The dotfiles repo has a new tool called `dotctl` — a Go CLI that collects
dotfiles status from Aurora + Distrobox containers and pushes metrics to the
homelab observability stack via the existing OTel Collector.

A homelab infrastructure review was completed and saved to the dotfiles repo at:
`~/personal/dotfiles/docs/plans/2026-03-12-dotctl-homelab-infra-review.md`

That review contains 4 action items with exact YAML manifests to apply.

## Task

Read the infrastructure review document at:
`~/personal/dotfiles/docs/plans/2026-03-12-dotctl-homelab-infra-review.md`

Then apply the action items **in this order** (matches the dependency order
from the review):

### 1. Loki HTTPRoute (Action Item 1)

Create `manifests/monitoring/grafana/loki-httproute.yaml` with the exact
content from the review. This exposes Loki at `loki.k8s.rommelporras.com`
so dotctl CLI can query LogQL.

Do NOT apply to the cluster yet — just create the file and commit.

### 2. PrometheusRules (Action Item 2)

Create `manifests/monitoring/alerts/dotctl-alerts.yaml` with the exact
content from the review. Two alerts: `DotctlCollectionStale` and
`DotctlDriftDetected`.

Do NOT apply to the cluster yet — just create the file and commit.

### 3. Grafana Dashboard (Action Item 3)

Create `manifests/monitoring/dashboards/dotctl-dashboard-configmap.yaml`.
The review provides the structure (4 rows, 8 panels) and implementation
notes. Use `claude-dashboard-configmap.yaml` in the same directory as the
reference for ConfigMap format, datasource UIDs, and panel conventions.

Follow the exact panel specifications from the review:
- Machine Status stat (dotctl_up, value mappings 0=DOWN/red, 1=UP/green)
- Drift Count stat + Drift Over Time timeseries (dotctl_drift_total)
- Drifted Files logs panel (Loki, service_name="dotctl")
- Tool Inventory table (dotctl_tool_installed, instant query)
- Credential Status table (dotctl_credential_status)
- Collection Age stat (time() - dotctl_collect_timestamp, thresholds 900s/1800s)
- Distrobox Containers stat (dotctl_container_running)

Timezone: Asia/Manila. Tags: ["dotctl", "homelab"]. UID: dotctl-status.
All panels and row headers must have description fields.

Do NOT apply to the cluster yet — just create the file and commit.

### 4. Documentation Updates (Action Item 4, optional)

Update these docs if they exist:
- `docs/context/Gateway.md` — add Loki to Exposed Services table
- `docs/context/Networking.md` — add Loki to Service URLs table
- `docs/context/Monitoring.md` — add to Access, Dashboards, and Alert Rules tables

## Important

- Do NOT apply any manifests to the cluster (`kubectl apply`). Only create
  files and commit. I will apply them manually after dotctl is pushing data.
- Follow existing patterns in the repo for YAML formatting, comments, and
  label conventions.
- Use conventional commits (e.g., `infra: add dotctl Grafana dashboard`).
- The review document is the source of truth for all YAML content and
  implementation decisions. Do not deviate from it.
