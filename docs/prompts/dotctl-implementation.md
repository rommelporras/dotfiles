# Prompt: Implement dotctl CLI

> Run this in a Claude Code session inside the dotfiles repo (`~/personal/dotfiles/`).

## Context

dotctl is a Go CLI tool for observing chezmoi-managed dotfiles across
Aurora DX, Distrobox containers, and (future) WSL2/Windows. It collects
drift status, tool inventory, and credential state, then pushes metrics to
the existing homelab OTel Collector (Prometheus + Loki). It also renders a
terminal dashboard.

This is a monorepo — Go code lives at the dotfiles repo root (`cmd/`,
`internal/`, `go.mod`). chezmoi only sees `home/` via `.chezmoiroot`, so
Go files are invisible to it.

## Documents to Read First

Read these three files before starting. They are the source of truth:

1. **Design doc:** `docs/plans/2026-03-10-dotctl-cli-design.md`
   - Architecture decisions, CLI interface, data model, metrics, error handling

2. **Implementation plan:** `docs/plans/2026-03-10-dotctl-cli-plan.md`
   - 15 TDD tasks with exact code, file paths, and test commands
   - Follow this task-by-task. Each task has steps: write test, verify fail,
     implement, verify pass, commit.

3. **Homelab infra review:** `docs/plans/2026-03-12-dotctl-homelab-infra-review.md`
   - Confirmed OTel Collector approach. Key details:
   - ServiceMonitor scrapes at 60s (not 15s)
   - Alerts route to Discord #apps (no #status channel)
   - OTel Collector metrics are in-memory — lost on restart
   - Loki HTTPRoute at loki.k8s.rommelporras.com (being created separately)

## Task

Execute the implementation plan (`docs/plans/2026-03-10-dotctl-cli-plan.md`)
task by task, starting from Task 1.

Key rules:
- **TDD** — write the failing test first, verify it fails, then implement
- **Conventional commits** after each task (e.g., `feat:`, `test:`, `chore:`)
- **Go module path** is `github.com/rommelporras/dotfiles` (monorepo)
- **No AI attribution** in commits or code
- **go vet** must pass before each commit
- Shell out to `chezmoi` and `distrobox` CLIs — don't reimplement them
- Use interfaces for external commands so tests can mock them
- Use `uv` for any Python tooling, `bun` for any JS (per project conventions)

## Key Config Values

```toml
otel_endpoint = "10.10.30.22:4317"
prometheus_url = "https://prometheus.k8s.rommelporras.com"
loki_url = "https://loki.k8s.rommelporras.com"
hostname = "aurora-dx"
```

## When Done

After completing all 15 tasks:
1. Run `make test` and `make lint` — verify all pass
2. Run `./dotctl status --live` — verify it displays output
3. Run `./dotctl collect --verbose` — verify it attempts to push metrics
4. Do NOT push to remote — I will review and reset git history first
