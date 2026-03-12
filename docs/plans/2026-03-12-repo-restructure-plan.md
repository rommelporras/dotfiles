# Repository Restructure Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restructure the dotfiles monorepo — move Go code under `dotctl/`, clean up `docs/`, rewrite README to ~150 lines, reset CHANGELOG to v1.0.0.

**Architecture:** All Go code moves to `dotctl/` with updated module path `github.com/rommelporras/dotfiles/dotctl`. Root Makefile delegates. Docs split into setup/, reference/, architecture/. Old plan files consolidated and deleted.

**Design doc:** `docs/plans/2026-03-12-repo-restructure-design.md`

---

## Task 1: Move Go Code to dotctl/ Subdirectory

**Files:**
- Create: `dotctl/` (directory)
- Move: `cmd/` → `dotctl/cmd/`
- Move: `internal/` → `dotctl/internal/`
- Move: `deploy/` → `dotctl/deploy/`
- Move: `go.mod` → `dotctl/go.mod`
- Move: `go.sum` → `dotctl/go.sum`
- Move: `Makefile` (root) → `dotctl/Makefile`
- Modify: `dotctl/go.mod` — update module path
- Modify: all 16 Go files — update import paths
- Create: `Makefile` (new root delegator)
- Modify: `.gitignore` — update binary path

**Step 1: Move directories with git mv**

```bash
cd ~/personal/dotfiles
git mv cmd dotctl/cmd
git mv internal dotctl/internal
git mv deploy dotctl/deploy
git mv go.mod dotctl/go.mod
git mv go.sum dotctl/go.sum
git mv Makefile dotctl/Makefile
```

**Step 2: Update module path in dotctl/go.mod**

Change line 1 from:
```
module github.com/rommelporras/dotfiles
```
to:
```
module github.com/rommelporras/dotfiles/dotctl
```

**Step 3: Bulk-update all import paths**

```bash
cd ~/personal/dotfiles/dotctl
find . -name "*.go" -exec sed -i \
  's|github.com/rommelporras/dotfiles/internal|github.com/rommelporras/dotfiles/dotctl/internal|g' {} \;
```

Verify the change took effect (should show 16 matches):
```bash
grep -r "github.com/rommelporras/dotfiles/dotctl/internal" . --include="*.go" | wc -l
```
Expected: 16

Verify no old paths remain:
```bash
grep -r "github.com/rommelporras/dotfiles/internal" . --include="*.go" | wc -l
```
Expected: 0

**Step 4: Verify Go build compiles cleanly**

```bash
cd ~/personal/dotfiles/dotctl
go build ./...
```
Expected: no output (success)

**Step 5: Run full test suite**

```bash
cd ~/personal/dotfiles/dotctl
go test ./... -v 2>&1 | tail -20
```
Expected: all 22 tests pass

**Step 6: Create new root Makefile**

Write `~/personal/dotfiles/Makefile`:
```makefile
.PHONY: build test lint install install-systemd uninstall-systemd clean

build:
	$(MAKE) -C dotctl build

test:
	$(MAKE) -C dotctl test

lint:
	$(MAKE) -C dotctl lint

install:
	$(MAKE) -C dotctl install

install-systemd:
	$(MAKE) -C dotctl install-systemd

uninstall-systemd:
	$(MAKE) -C dotctl uninstall-systemd

clean:
	$(MAKE) -C dotctl clean
```

**Step 7: Update .gitignore**

Change `/dotctl` to `/dotctl/dotctl` (binary is now inside the dotctl/ subdirectory).

**Step 8: Verify root Makefile delegates correctly**

```bash
cd ~/personal/dotfiles
make build
```
Expected: builds successfully, binary at `dotctl/dotctl`

```bash
cd ~/personal/dotfiles
make test 2>&1 | tail -5
```
Expected: all tests pass

**Step 9: Verify dotctl binary runs**

```bash
~/personal/dotfiles/dotctl/dotctl --help
```
Expected: shows usage with status and collect subcommands

**Step 10: Commit**

```bash
cd ~/personal/dotfiles
git add dotctl/ Makefile .gitignore
git commit -m "refactor: move Go code to dotctl/ subdirectory"
```

---

## Task 2: Restructure docs/ — Architecture

**Files:**
- Create: `docs/architecture/dotctl-design.md`
- Create: `docs/architecture/infra.md`
- Delete: `docs/plans/2026-03-10-dotctl-cli-design.md`
- Delete: `docs/plans/2026-03-10-dotctl-cli-plan.md`
- Delete: `docs/plans/2026-03-12-dotctl-homelab-infra-review.md`
- Delete: `docs/plans/2026-03-05-claude-code-plugins-design.md`
- Delete: `docs/plans/2026-03-05-claude-code-plugins-plan.md`
- Delete: `docs/prompts/dotctl-implementation.md`
- Delete: `docs/prompts/homelab-apply-dotctl-infra.md`

**Step 1: Create docs/architecture/dotctl-design.md**

This consolidates the key decisions from the 3 dotctl plan files into a concise design reference. Write:

```markdown
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
```

**Step 2: Create docs/architecture/infra.md**

Write a concise version of the homelab infra decisions:

```markdown
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
```

**Step 3: Delete old plan and prompt files**

```bash
cd ~/personal/dotfiles
git rm docs/plans/2026-03-10-dotctl-cli-design.md
git rm docs/plans/2026-03-10-dotctl-cli-plan.md
git rm docs/plans/2026-03-12-dotctl-homelab-infra-review.md
git rm docs/plans/2026-03-05-claude-code-plugins-design.md
git rm docs/plans/2026-03-05-claude-code-plugins-plan.md
git rm docs/prompts/dotctl-implementation.md
git rm docs/prompts/homelab-apply-dotctl-infra.md
```

Also remove the now-empty prompts directory if empty:
```bash
rmdir docs/prompts 2>/dev/null || true
```

**Step 4: Commit**

```bash
cd ~/personal/dotfiles
git add docs/architecture/ docs/plans/ docs/prompts/
git commit -m "docs: consolidate architecture docs, delete stale plans"
```

---

## Task 3: Restructure docs/ — Reference

**Files:**
- Create: `docs/reference/dotctl.md`
- Create: `docs/reference/environment-model.md`
- Create: `docs/reference/credentials.md`
- Move: `docs/distrobox-scripts.md` → `docs/reference/distrobox-scripts.md`

**Step 1: Create docs/reference/dotctl.md**

```markdown
# dotctl Reference

## Installation

```bash
cd ~/personal/dotfiles/dotctl
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
```

**Step 2: Create docs/reference/environment-model.md**

Extract the environment model table from README and CLAUDE.md:

```markdown
# Environment Model

Templates use two variables: **platform** (auto-detected) and **context** (user-selected at `chezmoi init`).

## Platform Detection

Detected automatically — never prompted:

| Value | Detected when |
|---|---|
| `distrobox` | `$DISTROBOX_ENTER_PATH` is set |
| `wsl` | `/proc/version` contains "microsoft" |
| `aurora` | `/etc/os-release` contains "aurora" |
| `unknown` | none of the above |

## Context Values

Chosen by the user at `chezmoi init`:

| Context | Platform | Use case |
|---|---|---|
| `personal` | aurora | Personal laptop host — launches Distrobox containers |
| `gaming` | wsl | Gaming desktop — personal projects |
| `work-eam` | wsl or distrobox | EAM work projects |
| `work-<name>` | distrobox | Any other work context |
| `personal-<project>` | distrobox | Project-scoped dev (Bun, Playwright, native op) |
| `sandbox` | distrobox | Clean experiment space, no credentials |

## What Each Context Gets

| Feature | personal | personal-\<project\> | work-\<name\> | gaming | sandbox |
|---|---|---|---|---|---|
| 1Password SSH agent | host socket | host socket | host socket | Windows bridge | fallback |
| glab | ✓ | ✓ | — | — | — |
| kubectl | ✓ | — | ✓ | — | — |
| terraform | — | — | ✓ | — | — |
| AWS CLI | — | — | ✓ | — | — |
| ansible | ✓ | — | — | — | — |
| op CLI (native) | — | ✓ | — | — | — |
| bun | — | ✓ | — | — | — |
| atuin | ✓ | ✓ | ✓ | ✓ | — |
| Claude Code | ✓ | ✓ | ✓ | ✓ | — |
| setup-creds | ✓ | ✓ | ✓ | — | — |

## Adding a New Context

**New work context:**
1. Add container to `containers/distrobox.ini`
2. Add job-specific aliases in `home/dot_zshrc.tmpl` under `hasPrefix .context "work-"`
3. Run: `uv run python scripts/distrobox_setup.py work-<name> --work-email you@company.com`

**New personal project:**
1. Add container to `containers/distrobox.ini`
2. Run: `uv run python scripts/distrobox_setup.py personal-<name> --personal-email you@email.com`
```

**Step 3: Create docs/reference/credentials.md**

Extract the credential setup section from README:

```markdown
# Credential Setup

## Distrobox Containers (automated)

Run inside any non-sandbox container:

```bash
setup-creds
```

Handles: Claude Code plugins, Context7 MCP, Atuin login, glab auth, kubeconfig (manual step), AWS (manual step). Pulls secrets from 1Password on the Aurora host via `distrobox-host-exec op`.

Requires: 1Password desktop app unlocked with CLI integration enabled
(Settings → Developer → Integrate with 1Password CLI).

## Aurora DX (manual)

```bash
# SSH public keys
cp id_ed25519.pub ~/.ssh/
chmod 644 ~/.ssh/*.pub
ssh-add -l    # verify 1Password agent works

# Claude Code plugins
claude plugin marketplace add anthropics/claude-plugins-official
claude plugin marketplace add obra/superpowers-marketplace
claude plugin install context7@claude-plugins-official --scope user
claude plugin install superpowers@superpowers-marketplace --scope user
claude plugin install episodic-memory@superpowers-marketplace --scope user

# Context7 MCP
claude mcp add --scope user --transport http context7 https://mcp.context7.com/mcp \
  --header "CONTEXT7_API_KEY: $(op read 'op://Kubernetes/Context7/api-key' --no-newline)"

# Homelab kubeconfig
cp homelab.yaml ~/.kube/

# GitLab
glab auth login --hostname gitlab.k8s.rommelporras.com \
  --token "$(op read 'op://Kubernetes/Gitlab/personal-access-token')"

# Atuin
atuin login -u <account> \
  -p "$(op read 'op://Kubernetes/Atuin/<context>-password')" \
  -k "$(op read 'op://Kubernetes/Atuin/encryption-key')"
```

## WSL2 (manual)

Same as Aurora, plus GitHub CLI:

```bash
gh auth login
```

## AI Sandbox — Podman Secrets

```bash
podman secret create anthropic_key <(echo "sk-ant-...")
podman secret create gemini_key <(echo "AI...")

# Deploy key for --git flag
ssh-keygen -t ed25519 -f ~/.ssh/ai-deploy-key -C 'ai-sandbox-deploy'
# Add .pub to GitHub/GitLab as deploy key
```
```

**Step 4: Move distrobox-scripts.md**

```bash
cd ~/personal/dotfiles
git mv docs/distrobox-scripts.md docs/reference/distrobox-scripts.md
```

**Step 5: Commit**

```bash
cd ~/personal/dotfiles
git add docs/reference/
git commit -m "docs: add reference docs for dotctl, environment model, credentials"
```

---

## Task 4: Restructure docs/ — Setup Guides

**Files:**
- Create: `docs/setup/wsl2.md`
- Create: `docs/setup/aurora.md`
- Create: `docs/setup/distrobox.md`

These are extracted directly from the current README. Content is the same — just moved so README can be concise.

**Step 1: Create docs/setup/wsl2.md**

Copy the "WSL2 (Ubuntu) — platform setup" section from README verbatim, then append the chezmoi init section and credential setup link:

```markdown
# WSL2 Setup

## 1. Platform Prerequisites

[copy the WSL2 platform setup section from README — 1Password, npiperelay steps]

## 2. Install chezmoi and apply dotfiles

```bash
sudo -v
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply rommelporras
```

chezmoi will ask: context, personal email, work email, work credentials, homelab credentials, Atuin URL, Atuin account.

After install:
```bash
exec zsh
```

## 3. Font setup

Install JetBrainsMono Nerd Font manually on Windows:
1. Download from https://www.nerdfonts.com/font-downloads
2. Extract zip, select all `.ttf` files, right-click → Install
3. Windows Terminal → Settings → Profile → Appearance → Font face → `JetBrainsMono Nerd Font`

## 4. Set up credentials

See [docs/reference/credentials.md](../reference/credentials.md).
```

**Step 2: Create docs/setup/aurora.md**

Copy the "Aurora DX — platform setup" section from README verbatim, then append:

```markdown
# Aurora DX Setup

## 1. Platform Prerequisites (follow in order)

[copy the Aurora DX platform setup section from README — ujust devmode, brew, 1Password, etc.]

## 2. Install chezmoi and apply dotfiles

```bash
# Clone the repo first (chezmoi is already installed via brew)
mkdir -p ~/personal
git clone git@github.com:rommelporras/dotfiles.git ~/personal/dotfiles

chezmoi init --apply ~/personal/dotfiles
```

After install:
```bash
exec zsh
```

## 3. Build dotctl

```bash
cd ~/personal/dotfiles
make install          # builds and copies to ~/.local/bin/
make install-systemd  # enables 10-minute collection timer
```

## 4. Set up Distrobox containers

Ensure 1Password desktop app is unlocked and CLI integration is enabled.

```bash
cd ~/personal/dotfiles
uv run python scripts/distrobox_setup.py personal \
  --personal-email git@rommelporras.com

uv run python scripts/distrobox_setup.py work-eam \
  --work-email work@company.com
```

See [docs/reference/distrobox-scripts.md](../reference/distrobox-scripts.md) for full reference.

## 5. Set up credentials

See [docs/reference/credentials.md](../reference/credentials.md).
```

**Step 3: Create docs/setup/distrobox.md**

```markdown
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

## Day-to-day

```bash
distrobox enter personal        # enter a container
~/bin/chezmoi apply -v          # update dotfiles inside container
exec zsh                        # reload shell
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
```

**Step 4: Commit**

```bash
cd ~/personal/dotfiles
git add docs/setup/
git commit -m "docs: add per-platform setup guides"
```

---

## Task 5: Rewrite README.md

**Files:**
- Modify: `README.md` — full rewrite, ~150 lines

**Step 1: Rewrite README.md**

Replace the entire file with:

```markdown
# dotfiles

chezmoi-managed dotfiles for consistent dev environments across WSL2, Aurora DX, and Distrobox containers. Includes `dotctl` — a CLI for monitoring dotfiles status across all machines.

## Quick Start

Choose your platform:

- **WSL2** — [docs/setup/wsl2.md](docs/setup/wsl2.md)
- **Aurora DX** — [docs/setup/aurora.md](docs/setup/aurora.md)
- **Distrobox containers** — [docs/setup/distrobox.md](docs/setup/distrobox.md)

All platforms run `chezmoi init --apply` — it detects the platform automatically and asks for context.

## dotctl

Monitor dotfiles status across all machines.

```bash
# Install
cd ~/personal/dotfiles
make install          # ~/.local/bin/dotctl
make install-systemd  # auto-collect every 10 minutes

# Commands
dotctl status           # query Prometheus + Loki dashboard
dotctl status --live    # collect locally, no cluster needed
dotctl collect --verbose  # push metrics to OTel Collector
```

See [docs/reference/dotctl.md](docs/reference/dotctl.md) for full reference.

## Day-to-Day

```bash
chezmoi diff        # preview what would change
chezmoi apply       # apply changes
chezmoi update      # pull latest + apply
chezmoi edit ~/.zshrc && chezmoi apply  # edit a managed file
```

## AI Sandbox

Run AI agents in an isolated Podman container with no access to host credentials.

```bash
ai-sandbox claude -- --dangerously-skip-permissions  # code only
ai-sandbox --git claude -- --dangerously-skip-permissions  # code + git push
ai-sandbox --no-network gemini  # maximum containment
```

## Repository Layout

```
dotfiles/
├── dotctl/          — Go CLI (build, collect, status)
├── home/            — chezmoi source → maps to ~/
├── scripts/         — Distrobox setup + integration tests (Python)
├── containers/      — distrobox.ini + Containerfile.ai-sandbox
├── bin/             — ai-sandbox CLI
├── deploy/          — (inside dotctl/) systemd units
└── docs/
    ├── setup/       — per-platform setup guides
    ├── reference/   — CLI reference, environment model, credentials
    └── architecture/ — design decisions
```

## Environment Model

Two variables: **platform** (auto-detected) and **context** (chosen at `chezmoi init`).

| Platform | Context | Description |
|---|---|---|
| `aurora` | `personal` | Personal laptop host |
| `wsl` | `gaming` | Gaming desktop |
| `wsl` | `work-eam` | Work laptop |
| `distrobox` | `personal` | Personal dev container |
| `distrobox` | `personal-<project>` | Project-scoped (Bun, Playwright, native op) |
| `distrobox` | `work-<name>` | Work dev container |
| `distrobox` | `sandbox` | Clean experiment space, no credentials |

See [docs/reference/environment-model.md](docs/reference/environment-model.md) for full matrix.

## Testing

```bash
# dotctl unit tests
make test

# Distrobox integration tests (Aurora only)
uv run python scripts/test_distrobox_integration.py --all
uv run python scripts/test_distrobox_integration.py personal
```

## License

MIT. See [LICENSE](LICENSE).
```

**Step 2: Verify README renders correctly**

Count lines:
```bash
wc -l ~/personal/dotfiles/README.md
```
Expected: under 150

**Step 3: Commit**

```bash
cd ~/personal/dotfiles
git add README.md
git commit -m "docs: rewrite README to concise overview with links to docs/"
```

---

## Task 6: Reset CHANGELOG and Update CLAUDE.md

**Files:**
- Modify: `CHANGELOG.md` — reset to v1.0.0
- Modify: `CLAUDE.md` — update structure references

**Step 1: Rewrite CHANGELOG.md**

Replace entire file with:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-03-12

### Added

- chezmoi-managed dotfiles for WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora),
  and Distrobox containers
- Two-variable environment model: platform (auto-detected) + context (user-selected)
- Distrobox container lifecycle automation (`scripts/distrobox_setup.py`)
- AI sandbox — Podman container for AI agents with tiered credential access (`bin/ai-sandbox`)
- `dotctl` Go CLI — collect dotfiles status, push metrics/logs to OTel Collector,
  query Prometheus + Loki terminal dashboard
- Systemd timer for periodic collection (`dotctl/deploy/`)
- setup-creds — automated credential seeding for Distrobox containers (Claude Code
  plugins, Context7 MCP, Atuin, glab, kubeconfig)
- 84 integration test assertions across 4 container types
```

**Step 2: Update CLAUDE.md — repository structure section**

Find the `## Repository Structure` section and update the tree to reflect the new layout:
- `cmd/`, `internal/`, `deploy/` now live under `dotctl/`
- `docs/` now has `setup/`, `reference/`, `architecture/` subdirs
- Note the new Makefile delegation pattern

Also update the `## dotctl (Go CLI)` section paths (e.g., `make build` now works from repo root via delegation).

**Step 3: Commit**

```bash
cd ~/personal/dotfiles
git add CHANGELOG.md CLAUDE.md
git commit -m "docs: reset CHANGELOG to v1.0.0, update CLAUDE.md structure"
```

---

## Task 7: Final Verification

**Step 1: Full test suite from repo root**

```bash
cd ~/personal/dotfiles
make test 2>&1 | tail -10
```
Expected: all 22 tests pass

**Step 2: Build and run dotctl**

```bash
cd ~/personal/dotfiles
make build
./dotctl/dotctl status --live
```
Expected: renders dashboard with Aurora host data

**Step 3: Verify docs/ structure**

```bash
find ~/personal/dotfiles/docs -type f | sort
```
Expected output:
```
docs/architecture/dotctl-design.md
docs/architecture/infra.md
docs/plans/2026-03-12-repo-restructure-design.md
docs/plans/2026-03-12-repo-restructure-plan.md
docs/reference/credentials.md
docs/reference/distrobox-scripts.md
docs/reference/dotctl.md
docs/reference/environment-model.md
docs/setup/aurora.md
docs/setup/distrobox.md
docs/setup/wsl2.md
```

**Step 4: Verify no stale plan files remain**

```bash
ls ~/personal/dotfiles/docs/plans/
```
Expected: only `2026-03-12-repo-restructure-design.md` and `2026-03-12-repo-restructure-plan.md`

**Step 5: Push**

```bash
cd ~/personal/dotfiles
git push
```
