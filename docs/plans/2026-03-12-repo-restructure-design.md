# Design: Repository Restructure

**Date:** 2026-03-12
**Status:** Approved

---

## Goal

Restructure the dotfiles monorepo for long-term usability and maintainability:
- Move all Go code under `dotctl/` subdirectory
- Consolidate and clean up `docs/`
- Rewrite README to ~150 lines (depth lives in docs/)
- Reset CHANGELOG to v1.0.0
- Update CLAUDE.md for new structure

---

## Decisions

| Question | Decision | Rationale |
|---|---|---|
| Repo name | Keep `dotfiles` | Primary content is chezmoi dotfiles; dotctl is a tool within |
| Go module path | `github.com/rommelporras/dotfiles/dotctl` | Honest — matches repo location; no mismatch technical debt |
| Go code location | `dotctl/` subdirectory | Isolates all Go from chezmoi/scripts; root stays clean |
| Future tools | `go.work` pattern | Each tool gets own `go.mod` under its subfolder; workspace ties them |
| docs structure | setup/ + reference/ + architecture/ | Purpose-oriented: do a thing / look up a thing / understand why |

---

## Repository Structure

```
dotfiles/
├── dotctl/                          # Go CLI — all Go code isolated here
│   ├── cmd/dotctl/main.go
│   ├── internal/
│   │   ├── collector/
│   │   ├── config/
│   │   ├── display/
│   │   ├── model/
│   │   ├── push/
│   │   └── query/
│   ├── deploy/                      # systemd units
│   ├── go.mod                       # module github.com/rommelporras/dotfiles/dotctl
│   ├── go.sum
│   └── Makefile                     # dotctl-specific: build/test/install/lint/install-systemd
├── home/                            # chezmoi source (unchanged)
├── scripts/                         # Python distrobox automation (unchanged)
├── containers/                      # distrobox.ini, Containerfile.ai-sandbox
├── bin/                             # ai-sandbox CLI
├── docs/
│   ├── setup/
│   │   ├── wsl2.md                  # WSL2 platform setup
│   │   ├── aurora.md                # Aurora DX setup
│   │   └── distrobox.md             # Distrobox containers setup
│   ├── reference/
│   │   ├── dotctl.md                # dotctl CLI commands reference
│   │   ├── distrobox-scripts.md     # Python scripts params (moved)
│   │   ├── environment-model.md     # platform+context matrix
│   │   └── credentials.md          # Per-platform credential setup
│   └── architecture/
│       ├── dotctl-design.md         # Consolidated dotctl design decisions
│       └── infra.md                 # Consolidated homelab infra review
├── Makefile                         # Root: thin delegator to dotctl/ + uv
├── README.md                        # Rewritten: ~150 lines, links to docs/
├── CHANGELOG.md                     # Reset to v1.0.0
├── CLAUDE.md                        # Updated for new structure
├── pyproject.toml
└── hooks/
```

---

## docs/ Philosophy

- **`docs/setup/`** — one file per platform. Open the file for your machine, follow top to bottom. No cross-referencing required.
- **`docs/reference/`** — look up specific things. Structured like man pages. `dotctl.md` covers all flags and examples. `credentials.md` replaces the credential block in the current README.
- **`docs/architecture/`** — why decisions were made. Consolidates the 5 existing plan files (`2026-03-10-dotctl-cli-design.md`, `2026-03-10-dotctl-cli-plan.md`, `2026-03-12-dotctl-homelab-infra-review.md`, `dotctl-implementation.md`, `homelab-apply-dotctl-infra.md`) into 2 focused docs. Old plan files deleted.

---

## Root Makefile

```makefile
.PHONY: build test lint install install-systemd

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
```

---

## dotctl/Makefile

Same as current Makefile but paths adjust for the new location:

```makefile
BINARY := dotctl
INSTALL_DIR := $(HOME)/.local/bin
SYSTEMD_DIR := $(HOME)/.config/systemd/user

build:
	go build -o $(BINARY) ./cmd/dotctl/

test:
	go test ./... -v

lint:
	go vet ./...

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

install-systemd: install
	mkdir -p $(SYSTEMD_DIR)
	cp deploy/dotctl-collect.service $(SYSTEMD_DIR)/
	cp deploy/dotctl-collect.timer $(SYSTEMD_DIR)/
	systemctl --user daemon-reload
	systemctl --user enable --now dotctl-collect.timer

uninstall-systemd:
	systemctl --user disable --now dotctl-collect.timer || true
	rm -f $(SYSTEMD_DIR)/dotctl-collect.service $(SYSTEMD_DIR)/dotctl-collect.timer
	systemctl --user daemon-reload

clean:
	rm -f $(BINARY)
```

---

## README Structure (~150 lines)

1. **Header** — what this repo is (3 sentences)
2. **Quick Start** — platform sections (WSL2 / Aurora / Distrobox), commands only
3. **dotctl** — install + 3 commands with example output screenshot/block
4. **Day-to-day** — chezmoi diff/apply/update
5. **Repository layout** — tree, one line per entry
6. **Links** — pointers to docs/ for depth

---

## CHANGELOG Reset

```markdown
# Changelog

## [1.0.0] - 2026-03-12

### Added
- chezmoi-managed dotfiles for WSL2, Aurora DX, and Distrobox containers
- Two-variable environment model (platform + context)
- Distrobox container lifecycle automation (distrobox_setup.py)
- AI sandbox (Podman, tiered credential access)
- dotctl Go CLI: collect dotfiles status, push to OTel Collector, query dashboard
```

---

## Files Deleted

- `docs/plans/2026-03-10-dotctl-cli-design.md` → consolidated into `docs/architecture/dotctl-design.md`
- `docs/plans/2026-03-10-dotctl-cli-plan.md` → implementation complete, no longer needed
- `docs/plans/2026-03-12-dotctl-homelab-infra-review.md` → consolidated into `docs/architecture/infra.md`
- `docs/plans/2026-03-05-claude-code-plugins-design.md` → stale, complete
- `docs/plans/2026-03-05-claude-code-plugins-plan.md` → stale, complete
- `docs/prompts/dotctl-implementation.md` → implementation complete
- `docs/prompts/homelab-apply-dotctl-infra.md` → applied, no longer needed
- `docs/distrobox-scripts.md` → moved to `docs/reference/distrobox-scripts.md`
- `cmd/`, `internal/`, `deploy/`, `go.mod`, `go.sum`, `Makefile` (root) → moved under `dotctl/`

---

## Go Module Path Change

Current: `module github.com/rommelporras/dotfiles`
New: `module github.com/rommelporras/dotfiles/dotctl`

All internal imports update accordingly (e.g., `github.com/rommelporras/dotfiles/internal/model` → `github.com/rommelporras/dotfiles/dotctl/internal/model`).
