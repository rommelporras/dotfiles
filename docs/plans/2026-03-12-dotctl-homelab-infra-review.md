# dotctl Homelab Infrastructure Review

> **For Claude:** This file was produced by a homelab repo session that reviewed
> the infrastructure section of `docs/plans/2026-03-10-dotctl-cli-design.md`
> against the actual K8s cluster state. Decisions here are final — do NOT
> re-evaluate or second-guess them. Use this as the source of truth for what
> the homelab repo needs.

**Reviewed by:** Homelab repo Claude session (2026-03-12)
**Design doc reviewed:** `docs/plans/2026-03-10-dotctl-cli-design.md` (Infrastructure Section, lines 205-307)
**Cluster context read:** Monitoring.md, Gateway.md, Architecture.md, Networking.md, OTel Collector config, Loki Helm values, existing alerts, existing dashboards

---

## Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Which approach? | **Approach 1: Reuse OTel Collector** | Zero new deployments, pipeline already handles arbitrary OTLP |
| OTel Collector config changes? | **None** | Collector accepts all OTLP. `dotctl_*` prefix separates from `claude_code_*` |
| ServiceMonitor changes? | **None** | Existing ServiceMonitor scrapes `:8889` at 60s. dotctl metrics appear automatically |
| Loki access for CLI queries? | **HTTPRoute** at `loki.k8s.rommelporras.com` | Consistent with Prometheus (already exposed unauthenticated). Network boundary = security boundary |
| Alert routing channel? | **#apps** (catch-all) | dotctl is workstation dotfiles, not cluster infra. Falls to default warning route |
| Alertmanager config changes? | **None** | `Dotctl.*` should NOT be added to the infra regex. Catch-all `discord-apps` handles it |
| Storage concerns? | **None** | +88 series on 50Gi Prometheus PVC, +1MB/day on 10Gi Loki PVC. Negligible |
| Alternative A (dedicated deployment)? | **Rejected** | Duplicates what OTel Collector does, adds maintenance |
| Alternative B (GitHub repo)? | **Rejected** | 52K commits/year, merge conflicts, no alerting, no Grafana |

## Corrections to Design Doc

These are inaccuracies in the design doc that should be noted but do NOT block implementation:

1. **Line 100** says "88 series x 1 sample/15s scrape" — actual ServiceMonitor interval is **60s**, and metrics only update every **10 minutes** (collection interval). Storage estimate is still valid.
2. **Line 239** says "Route to Discord #status" — there is **no #status channel**. Alerts route to **#apps** via the catch-all warning receiver.

## Gotcha: OTel Collector Restarts

The Prometheus exporter in the OTel Collector holds metrics **in memory**. If the collector pod restarts, all dotctl metrics disappear until the next `dotctl collect` push (up to 10 minutes later). The `DotctlCollectionStale` alert expression accounts for this with `absent_over_time()` and a 35-minute window.

---

## Homelab Repo Action Items

These are the exact files to create in `~/personal/homelab/`. No existing files need modification (except docs, which are optional).

### Action Item 1: Loki HTTPRoute

**File:** `manifests/monitoring/grafana/loki-httproute.yaml`

The other monitoring HTTPRoutes (Grafana, Prometheus, Alertmanager) live in this directory. Loki currently has no external route — only cluster-internal access via `loki.monitoring.svc.cluster.local:3100`.

**What to know:**
- Loki Helm chart deploys with `gateway.enabled: false` (we use Gateway API, not the Loki nginx gateway)
- Loki service name: `loki` in namespace `monitoring`, port `3100`
- Loki has `auth_enabled: false` — same security model as Prometheus
- Use `sectionName: https` on `homelab-gateway` in `default` namespace
- Hostname: `loki.k8s.rommelporras.com` (wildcard DNS already resolves `*.k8s.rommelporras.com` to `10.10.30.20`)

**Exact content:**

```yaml
# Loki HTTPRoute — exposes Loki HTTP API for external queries
# Used by dotctl CLI to query LogQL via https://loki.k8s.rommelporras.com
#
# Security: auth_enabled=false, same posture as Prometheus.
# Access restricted to VLAN 30 + Tailscale (network boundary).
#
# Apply:
#   kubectl-homelab apply -f manifests/monitoring/grafana/loki-httproute.yaml
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: loki
  namespace: monitoring
spec:
  parentRefs:
    - name: homelab-gateway
      namespace: default
      sectionName: https
  hostnames:
    - loki.k8s.rommelporras.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: loki
          port: 3100
```

**Verify after apply:**

```bash
kubectl-homelab get httproute loki -n monitoring
curl -s https://loki.k8s.rommelporras.com/ready
curl -s https://loki.k8s.rommelporras.com/loki/api/v1/labels | head
```

### Action Item 2: PrometheusRules for dotctl alerts

**File:** `manifests/monitoring/alerts/dotctl-alerts.yaml`

**What to know:**
- Labels `release: prometheus` and `app.kubernetes.io/part-of: kube-prometheus-stack` are required for Prometheus Operator discovery
- Both alerts are `severity: warning` — routes to `discord-apps` catch-all
- `DotctlCollectionStale` uses `absent_over_time()` with a 35m window to handle OTel Collector restarts (metrics are in-memory, lost on restart, next push is up to 10 minutes later)
- `DotctlDriftDetected` fires after 1 hour of continuous drift > 0

**Exact content:**

```yaml
# PrometheusRule for dotctl Dotfiles Status Alerts
# Alerts on stale collection and persistent drift
#
# Metrics come through OTel Collector Prometheus exporter (:8889):
#   dotctl_collect_timestamp - Unix timestamp of last collection (gauge)
#   dotctl_drift_total - Count of drifted files per machine (gauge)
#
# Routing: severity=warning -> discord-apps (catch-all)
#
# Apply:
#   kubectl-homelab apply -f manifests/monitoring/alerts/dotctl-alerts.yaml
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: dotctl-alerts
  namespace: monitoring
  labels:
    release: prometheus
    app.kubernetes.io/part-of: kube-prometheus-stack
spec:
  groups:
    - name: dotctl
      rules:
        # No collection from any machine in >30 minutes
        # Uses absent_over_time with 35m window to avoid false positives
        # after OTel Collector restarts (metrics are in-memory only)
        - alert: DotctlCollectionStale
          expr: |
            (time() - dotctl_collect_timestamp) > 1800
            or
            absent_over_time(dotctl_collect_timestamp[35m])
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "dotctl collection stale{{ if $labels.hostname }} on {{ $labels.hostname }}{{ end }}"
            description: "No dotctl collection received in over 30 minutes. The systemd timer on the reporting machine may have stopped."

        # Drift detected and persisting for >1 hour
        - alert: DotctlDriftDetected
          expr: dotctl_drift_total > 0
          for: 1h
          labels:
            severity: warning
          annotations:
            summary: "Dotfile drift on {{ $labels.hostname }}"
            description: "{{ $labels.hostname }} has {{ $value }} drifted files for over 1 hour. Run `dotctl status --machine {{ $labels.hostname }}` for details."
```

### Action Item 3: Grafana Dashboard ConfigMap

**File:** `manifests/monitoring/dashboards/dotctl-dashboard-configmap.yaml`

**What to know:**
- Label `grafana_dashboard: "1"` for sidecar auto-provisioning
- Annotation `grafana_folder: "Homelab"` to place in the Homelab folder
- Timezone: `Asia/Manila`
- Tags: `["dotctl", "homelab"]`
- dotctl is NOT a K8s workload, so the standard Pod Status / Network / Resource rows don't apply
- Dashboard rows: Machine Status, Drift, Tools & Credentials, Collection Health
- Uses both Prometheus (metrics) and Loki (drift file details) datasources
- Datasource UIDs: `prometheus` for Prometheus, `loki` for Loki (matching existing datasource configs)

**Metrics referenced (all are gauges pushed via OTel):**

| Metric | Labels | Panel |
|--------|--------|-------|
| `dotctl_up` | hostname, platform, context | Machine Status grid |
| `dotctl_collect_timestamp` | hostname | Collection Age stat |
| `dotctl_drift_total` | hostname | Drift Count stat, Drift Over Time timeseries |
| `dotctl_tool_installed` | hostname, tool | Tool Inventory table |
| `dotctl_credential_status` | hostname, credential | Credential Status table |
| `dotctl_container_running` | name | Distrobox Containers stat |

**Loki query:** `{service_name="dotctl"} | json | drift_files != "[]"` for the Drifted Files log panel.

**Dashboard structure (4 rows, 8 panels):**

```
Row 1: Machine Status
  - [stat] Machine UP/DOWN per hostname+platform (full width)

Row 2: Drift
  - [stat] Drift Count per hostname (1/3 width)
  - [timeseries] Drift Over Time per hostname (2/3 width)
  - [logs] Drifted Files from Loki (1/3 width, below stat)

Row 3: Tools & Credentials
  - [table] Tool Inventory — hostname x tool matrix (half width)
  - [table] Credential Status — hostname x credential matrix (half width)

Row 4: Collection Health
  - [stat] Collection Age per hostname with thresholds: green <15m, yellow <30m, red >30m (half width)
  - [stat] Distrobox Containers running/stopped (half width)
```

**The dashboard JSON is large.** When implementing, follow the pattern in `manifests/monitoring/dashboards/claude-dashboard-configmap.yaml` (same OTel-sourced metrics, same datasource UIDs). Key implementation notes:

- Machine Status stat: `dotctl_up` with value mappings (0=DOWN/red, 1=UP/green), `legendFormat: "{{ hostname }} ({{ platform }})"`
- Collection Age stat: `time() - dotctl_collect_timestamp`, unit `s`, thresholds at 900s/1800s
- Tool Inventory table: `dotctl_tool_installed` with `format: "table"`, `instant: true`, value mappings (0="-"/red, 1="OK"/green). Hide Time, __name__, job, instance columns via organize transform
- Credential Status table: same pattern as Tool Inventory but with `dotctl_credential_status`
- Drift Over Time timeseries: `dotctl_drift_total`, line chart, `legendFormat: "{{ hostname }}"`
- Drifted Files logs panel: Loki datasource, `{service_name="dotctl"} | json | drift_files != "[]"`, sort descending
- Distrobox Containers stat: `dotctl_container_running`, value mappings (0=STOPPED/gray, 1=RUNNING/green)
- All panels must have a `description` field (renders as tooltip on hover)
- All row headers must have a `description` field
- `"timezone": "Asia/Manila"`, `"uid": "dotctl-status"`, `"title": "Dotfiles Status"`

### Action Item 4: Documentation Updates (optional)

These are low-priority and can be done in a docs commit after the infra commit:

1. **`docs/context/Gateway.md`** — add Loki to the Exposed Services table:
   `| Loki | https://loki.k8s.rommelporras.com | loki | monitoring | https |`

2. **`docs/context/Networking.md`** — add Loki to the Service URLs table:
   `| Loki | https://loki.k8s.rommelporras.com | base |`

3. **`docs/context/Monitoring.md`** — add to Access table, Grafana Dashboards table, and Alert Rules table

---

## dotctl Config Values (confirmed correct)

For `~/.config/dotctl/config.toml` on Aurora:

```toml
otel_endpoint = "10.10.30.22:4317"      # OTel Collector VIP (Cilium L2 LoadBalancer)
prometheus_url = "https://prometheus.k8s.rommelporras.com"  # Existing HTTPRoute
loki_url = "https://loki.k8s.rommelporras.com"             # Requires Action Item 1
hostname = "aurora-dx"
```

---

## Dependency Order

```
dotctl collect (push) works immediately
  └── OTel Collector at 10.10.30.22:4317 already accepts OTLP
  └── Metrics appear on :8889, scraped by existing ServiceMonitor
  └── Logs go to Loki via internal OTLP endpoint

dotctl status (query) requires Action Item 1 first
  └── Prometheus queries work now (HTTPRoute exists)
  └── Loki queries need loki-httproute.yaml applied first

Grafana dashboard requires Action Item 3
  └── But also needs dotctl to be pushing data, otherwise panels are empty

Alerts require Action Item 2
  └── But will fire immediately if no dotctl data exists (DotctlCollectionStale)
  └── Apply alerts AFTER dotctl collect is running to avoid noise
```

**Recommended sequence:**
1. Build dotctl and get `dotctl collect` pushing data
2. Apply Loki HTTPRoute (Action Item 1)
3. Apply dashboard (Action Item 3)
4. Verify data appears in Grafana
5. Apply alerts (Action Item 2)
