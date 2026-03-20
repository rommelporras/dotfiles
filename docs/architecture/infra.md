# Homelab Infrastructure for dotctl

## Approach

Reuse existing OTel Collector at `10.10.30.22:4317`. Zero new K8s deployments.
Metrics flow: `dotctl collect` → OTLP gRPC → OTel Collector → Prometheus exporter
(:8889) → ServiceMonitor → Prometheus. Logs: same Collector → Loki via internal OTLP.

## What Was Applied

| Resource | File | Status |
|---|---|---|
| Loki HTTPRoute | `manifests/monitoring/grafana/loki-httproute.yaml` | Applied |
| PrometheusRules | `manifests/monitoring/alerts/dotctl-alerts.yaml` | Applied |
| Grafana Dashboard | `manifests/monitoring/dashboards/dotctl-dashboard-configmap.yaml` | Applied |

## Alerts

Both `severity: warning` → routed to `#apps` (Discord catch-all, not `#status`):

- `DotctlCollectionStale` — no collection in >30 min, or `absent_over_time(35m)`
- `DotctlDriftDetected` — drift > 0 persisting for >1 hour

## Confirmed Config Values

```toml
otel_endpoint   = "10.10.30.22:4317"                        # Cilium L2 VIP
prometheus_url  = "https://prometheus.k8s.rommelporras.com"
loki_url        = "https://loki.k8s.rommelporras.com"
```
