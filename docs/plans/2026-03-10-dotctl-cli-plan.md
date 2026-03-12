# dotctl CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool that collects dotfiles status from Aurora + Distrobox containers, pushes metrics to the existing OTel Collector, and displays a terminal dashboard querying Prometheus + Loki.

**Architecture:** Single Go binary with three modes: `dotctl collect` (gather + push), `dotctl status` (query + display), `dotctl status --live` (gather + display directly). Collection shells out to `chezmoi` and `distrobox` CLIs. Push via OTLP gRPC. Query via Prometheus/Loki HTTP APIs.

**Tech Stack:** Go, cobra (CLI), lipgloss (terminal UI), OTel SDK (push), BurntSushi/toml (config), stdlib net/http (queries)

**Design doc:** `docs/plans/2026-03-10-dotctl-cli-design.md`

**Infra review:** `docs/plans/2026-03-12-dotctl-homelab-infra-review.md`

**Repository:** Monorepo — lives in the dotfiles repo (`~/personal/dotfiles/`).
Go code at repo root (`cmd/`, `internal/`, `go.mod`). chezmoi only sees `home/`
via `.chezmoiroot`, so Go files are invisible to it.

---

## Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/dotctl/main.go`
- Create: `Makefile`
- Create: `CLAUDE.md`
- Create: `.gitignore`

**Step 1: Initialize Go module in the dotfiles repo root**

```bash
cd ~/personal/dotfiles
go mod init github.com/rommelporras/dotfiles
```

**Step 2: Write minimal main.go**

```go
// cmd/dotctl/main.go
package main

import "fmt"

func main() {
	fmt.Println("dotctl")
}
```

**Step 3: Write Makefile**

```makefile
BINARY := dotctl
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build test lint install clean

build:
	go build -o $(BINARY) ./cmd/dotctl/

test:
	go test ./... -v

lint:
	go vet ./...

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)

clean:
	rm -f $(BINARY)
```

**Step 4: Update CLAUDE.md with dotctl section**

Add a `## dotctl` section to the existing `CLAUDE.md` in the repo root:

```markdown
## dotctl (Go CLI)

Go CLI tool living at repo root (`cmd/`, `internal/`, `go.mod`). chezmoi only
sees `home/` via `.chezmoiroot` — Go files are invisible to it.

### Architecture

Single binary, three modes:
- `dotctl collect` — gather status, push to OTel Collector via OTLP gRPC
- `dotctl status` — query Prometheus + Loki, render terminal tables
- `dotctl status --live` — gather + display directly (no cluster dependency)

### Key Packages

- `internal/collector/` — gathers chezmoi status, tool inventory, credentials
- `internal/push/` — OTLP gRPC push to OTel Collector
- `internal/query/` — Prometheus + Loki HTTP API queries
- `internal/display/` — terminal table rendering with lipgloss

### Conventions

- TDD: write failing test first, then implement
- Shell out to `chezmoi` and `distrobox` CLIs — don't reimplement them
- Interfaces for external commands to enable test mocking
- `go vet` must pass before every commit

### Commands

- `make build` — compile dotctl binary
- `make test` — run Go tests
- `make lint` — go vet
- `make install` — copy dotctl to ~/.local/bin/
```

**Step 5: Add dotctl to .gitignore**

Append to the existing `.gitignore`:

```
# dotctl build output
/dotctl
```

**Step 6: Verify build works**

Run: `make build && ./dotctl`
Expected: prints `dotctl`

**Step 7: Commit**

```bash
git add cmd/ go.mod Makefile .gitignore CLAUDE.md
git commit -m "chore: scaffold dotctl Go project in monorepo"
```

---

## Task 2: Domain Types & Config

**Files:**
- Create: `internal/model/model.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write domain types**

```go
// internal/model/model.go
package model

// DriftFile represents a single chezmoi-managed file with drift.
type DriftFile struct {
	Path   string `json:"path"`
	Status string `json:"status"` // M=modified, A=added, D=deleted, R=re-run needed
}

// MachineState holds the collected state of a single machine/container.
type MachineState struct {
	Hostname     string            `json:"hostname"`
	Platform     string            `json:"platform"`     // aurora, distrobox, wsl
	Context      string            `json:"context"`      // personal, work-eam, etc.
	DriftFiles   []DriftFile       `json:"drift_files"`
	TemplateData map[string]any    `json:"template_data"`
	Tools        map[string]string `json:"tools"`        // tool name → path or ""
	SSHAgent     string            `json:"ssh_agent"`    // 1password, manual, none, n/a
	SetupCreds   string            `json:"setup_creds"`  // ran, skipped, n/a
	AtuinSync    string            `json:"atuin_sync"`   // synced, disabled, n/a
}

// ContainerInfo describes a distrobox container.
type ContainerInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"` // running, stopped, created
	Image  string `json:"image"`
}

// CollectResult is the full output of a collection run.
type CollectResult struct {
	Machines   []MachineState  `json:"machines"`
	Containers []ContainerInfo `json:"containers"`
}
```

**Step 2: Write config test**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load defaults: %v", err)
	}
	if cfg.OTelEndpoint != "10.10.30.22:4317" {
		t.Errorf("OTelEndpoint = %q, want default", cfg.OTelEndpoint)
	}
	if cfg.Hostname == "" {
		t.Error("Hostname should be auto-detected")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
otel_endpoint = "localhost:4317"
prometheus_url = "http://prom:9090"
loki_url = "http://loki:3100"
hostname = "test-host"
`), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.OTelEndpoint != "localhost:4317" {
		t.Errorf("OTelEndpoint = %q, want localhost:4317", cfg.OTelEndpoint)
	}
	if cfg.Hostname != "test-host" {
		t.Errorf("Hostname = %q, want test-host", cfg.Hostname)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("Load missing file should not error: %v", err)
	}
	if cfg.OTelEndpoint != "10.10.30.22:4317" {
		t.Errorf("should return defaults when file missing")
	}
}
```

**Step 3: Run tests to verify they fail**

Run: `go test ./internal/config/ -v`
Expected: FAIL (package doesn't exist yet)

**Step 4: Install toml dependency and implement config**

```bash
go get github.com/BurntSushi/toml
```

```go
// internal/config/config.go
package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds dotctl configuration.
type Config struct {
	OTelEndpoint  string `toml:"otel_endpoint"`
	PrometheusURL string `toml:"prometheus_url"`
	LokiURL       string `toml:"loki_url"`
	Hostname      string `toml:"hostname"`
}

// Load reads config from path, falling back to defaults for missing file or fields.
func Load(path string) (*Config, error) {
	cfg := &Config{
		OTelEndpoint:  "10.10.30.22:4317",
		PrometheusURL: "https://prometheus.k8s.rommelporras.com",
		LokiURL:       "https://loki.k8s.rommelporras.com",
	}

	if path != "" {
		_, err := toml.DecodeFile(path, cfg)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	if cfg.Hostname == "" {
		name, err := os.Hostname()
		if err != nil {
			cfg.Hostname = "unknown"
		} else {
			cfg.Hostname = name
		}
	}

	return cfg, nil
}

// DefaultPath returns ~/.config/dotctl/config.toml.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.config/dotctl/config.toml"
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: PASS (3 tests)

**Step 6: Commit**

```bash
git add internal/model/ internal/config/ go.mod go.sum
git commit -m "feat: add domain types and config loading"
```

---

## Task 3: Platform Detection

**Files:**
- Create: `internal/collector/platform.go`
- Create: `internal/collector/platform_test.go`

**Step 1: Write platform detection tests**

```go
// internal/collector/platform_test.go
package collector

import "testing"

func TestDetectPlatformFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		procVer  string
		osRel    string
		want     string
	}{
		{
			name: "distrobox detected from env",
			env:  map[string]string{"DISTROBOX_ENTER_PATH": "/usr/bin/distrobox-enter"},
			want: "distrobox",
		},
		{
			name:    "wsl detected from /proc/version",
			procVer: "Linux version 5.15.0 (microsoft@microsoft.com)",
			want:    "wsl",
		},
		{
			name:  "aurora detected from os-release",
			osRel: "NAME=\"Aurora\"\nID=aurora\n",
			want:  "aurora",
		},
		{
			name: "unknown when nothing matches",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPlatformWith(tt.env, tt.procVer, tt.osRel)
			if got != tt.want {
				t.Errorf("detectPlatformWith() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/collector/ -v -run TestDetectPlatform`
Expected: FAIL

**Step 3: Implement platform detection**

```go
// internal/collector/platform.go
package collector

import (
	"os"
	"strings"
)

// DetectPlatform returns "aurora", "distrobox", "wsl", or "unknown".
func DetectPlatform() string {
	env := map[string]string{
		"DISTROBOX_ENTER_PATH": os.Getenv("DISTROBOX_ENTER_PATH"),
	}

	procVer := ""
	if b, err := os.ReadFile("/proc/version"); err == nil {
		procVer = string(b)
	}

	osRel := ""
	if b, err := os.ReadFile("/etc/os-release"); err == nil {
		osRel = string(b)
	}

	return detectPlatformWith(env, procVer, osRel)
}

func detectPlatformWith(env map[string]string, procVersion, osRelease string) string {
	if env["DISTROBOX_ENTER_PATH"] != "" {
		return "distrobox"
	}
	if strings.Contains(strings.ToLower(procVersion), "microsoft") {
		return "wsl"
	}
	if strings.Contains(strings.ToLower(osRelease), "aurora") {
		return "aurora"
	}
	return "unknown"
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/collector/ -v -run TestDetectPlatform`
Expected: PASS (4 subtests)

**Step 5: Commit**

```bash
git add internal/collector/platform.go internal/collector/platform_test.go
git commit -m "feat: add platform detection (aurora/distrobox/wsl)"
```

---

## Task 4: Chezmoi Collector

**Files:**
- Create: `internal/collector/chezmoi.go`
- Create: `internal/collector/chezmoi_test.go`

**Step 1: Write chezmoi parsing tests**

```go
// internal/collector/chezmoi_test.go
package collector

import (
	"testing"

	"github.com/rommelporras/dotfiles/internal/model"
)

func TestParseChezmoiStatus(t *testing.T) {
	output := ` M .zshrc
 A .config/new-file.toml
 R bootstrap.sh
 M .local/bin/setup-creds
`
	got := parseChezmoiStatus(output)

	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}

	want := []model.DriftFile{
		{Path: ".zshrc", Status: "M"},
		{Path: ".config/new-file.toml", Status: "A"},
		{Path: "bootstrap.sh", Status: "R"},
		{Path: ".local/bin/setup-creds", Status: "M"},
	}

	for i, w := range want {
		if got[i].Path != w.Path || got[i].Status != w.Status {
			t.Errorf("[%d] got %+v, want %+v", i, got[i], w)
		}
	}
}

func TestParseChezmoiStatusEmpty(t *testing.T) {
	got := parseChezmoiStatus("")
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestParseChezmoiData(t *testing.T) {
	jsonOutput := `{
		"context": "personal",
		"platform": "aurora",
		"atuin_account": "personal",
		"has_homelab_creds": true,
		"has_work_creds": false
	}`

	data, err := parseChezmoiData(jsonOutput)
	if err != nil {
		t.Fatalf("parseChezmoiData: %v", err)
	}
	if data["context"] != "personal" {
		t.Errorf("context = %v, want personal", data["context"])
	}
	if data["has_homelab_creds"] != true {
		t.Errorf("has_homelab_creds = %v, want true", data["has_homelab_creds"])
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/collector/ -v -run TestParseChezmoi`
Expected: FAIL

**Step 3: Implement chezmoi parsing**

```go
// internal/collector/chezmoi.go
package collector

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/rommelporras/dotfiles/internal/model"
)

// CommandRunner abstracts exec.Command for testing.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}

// ExecRunner shells out to real commands.
type ExecRunner struct{}

func (r *ExecRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	return string(out), err
}

func parseChezmoiStatus(output string) []model.DriftFile {
	var files []model.DriftFile
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 3 {
			continue
		}
		status := string(line[0])
		path := strings.TrimSpace(line[1:])
		if path == "" {
			continue
		}
		files = append(files, model.DriftFile{Path: path, Status: status})
	}
	return files
}

func parseChezmoiData(jsonOutput string) (map[string]any, error) {
	// chezmoi data --format json wraps everything under top-level keys.
	// We try to extract the [data] section first, fall back to flat parse.
	var full map[string]any
	if err := json.Unmarshal([]byte(jsonOutput), &full); err != nil {
		return nil, err
	}
	return full, nil
}

// CollectChezmoi gathers drift and template data from chezmoi CLI.
func CollectChezmoi(runner CommandRunner) ([]model.DriftFile, map[string]any, error) {
	statusOut, err := runner.Run("chezmoi", "status")
	if err != nil {
		return nil, nil, err
	}
	drift := parseChezmoiStatus(statusOut)

	dataOut, err := runner.Run("chezmoi", "data", "--format", "json")
	if err != nil {
		return drift, nil, err
	}
	data, err := parseChezmoiData(dataOut)
	if err != nil {
		return drift, nil, err
	}

	return drift, data, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/collector/ -v -run TestParseChezmoi`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/collector/chezmoi.go internal/collector/chezmoi_test.go
git commit -m "feat: add chezmoi status and data parsing"
```

---

## Task 5: Tool Probes & Credential Checks

**Files:**
- Create: `internal/collector/tools.go`
- Create: `internal/collector/tools_test.go`
- Create: `internal/collector/credentials.go`
- Create: `internal/collector/credentials_test.go`

**Step 1: Write tool probe tests**

```go
// internal/collector/tools_test.go
package collector

import "testing"

type mockLookPath struct {
	found map[string]string // tool → path
}

func (m *mockLookPath) LookPath(name string) (string, error) {
	if p, ok := m.found[name]; ok {
		return p, nil
	}
	return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
}

func TestProbeTools(t *testing.T) {
	mock := &mockLookPath{found: map[string]string{
		"glab":    "/usr/bin/glab",
		"kubectl": "/usr/local/bin/kubectl",
	}}

	got := probeToolsWith(mock)

	if got["glab"] != "/usr/bin/glab" {
		t.Errorf("glab = %q", got["glab"])
	}
	if got["kubectl"] != "/usr/local/bin/kubectl" {
		t.Errorf("kubectl = %q", got["kubectl"])
	}
	if got["terraform"] != "" {
		t.Errorf("terraform should be empty, got %q", got["terraform"])
	}
}
```

**Step 2: Write credential check tests**

```go
// internal/collector/credentials_test.go
package collector

import "testing"

func TestDetectSSHAgentType(t *testing.T) {
	tests := []struct {
		sock string
		want string
	}{
		{"/home/user/.1password/agent.sock", "1password"},
		{"/tmp/ssh-XXX/agent.123", "system"},
		{"", "none"},
	}

	for _, tt := range tests {
		got := detectSSHAgentType(tt.sock)
		if got != tt.want {
			t.Errorf("detectSSHAgentType(%q) = %q, want %q", tt.sock, got, tt.want)
		}
	}
}
```

**Step 3: Run tests to verify they fail**

Run: `go test ./internal/collector/ -v -run "TestProbeTools|TestDetectSSH"`
Expected: FAIL

**Step 4: Implement tool probes**

```go
// internal/collector/tools.go
package collector

import "os/exec"

// TrackedTools is the list of tools dotctl checks for.
var TrackedTools = []string{
	"glab", "kubectl", "terraform", "aws", "ansible",
	"op", "atuin", "bun",
}

// PathLookup abstracts exec.LookPath for testing.
type PathLookup interface {
	LookPath(name string) (string, error)
}

type realLookup struct{}

func (r *realLookup) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// ProbeTools checks which tracked tools are installed.
func ProbeTools() map[string]string {
	return probeToolsWith(&realLookup{})
}

func probeToolsWith(lookup PathLookup) map[string]string {
	result := make(map[string]string, len(TrackedTools))
	for _, tool := range TrackedTools {
		path, err := lookup.LookPath(tool)
		if err != nil {
			result[tool] = ""
		} else {
			result[tool] = path
		}
	}
	return result
}
```

**Step 5: Implement credential checks**

```go
// internal/collector/credentials.go
package collector

import (
	"os"
	"strings"
)

func detectSSHAgentType(sshAuthSock string) string {
	if sshAuthSock == "" {
		return "none"
	}
	if strings.Contains(sshAuthSock, "1password") {
		return "1password"
	}
	return "system"
}

// DetectSSHAgent reads SSH_AUTH_SOCK and classifies the agent type.
func DetectSSHAgent() string {
	return detectSSHAgentType(os.Getenv("SSH_AUTH_SOCK"))
}

// DetectSetupCreds checks if setup-creds is present and executable.
func DetectSetupCreds() string {
	home, _ := os.UserHomeDir()
	path := home + "/.local/bin/setup-creds"
	info, err := os.Stat(path)
	if err != nil {
		return "n/a"
	}
	if info.Mode()&0o111 != 0 {
		return "ran"
	}
	return "present"
}

// DetectAtuinSync checks if atuin is configured with a sync address.
func DetectAtuinSync() string {
	home, _ := os.UserHomeDir()
	configPath := home + "/.config/atuin/config.toml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "n/a"
	}
	if strings.Contains(string(data), "sync_address") {
		return "synced"
	}
	return "disabled"
}
```

**Step 6: Run tests to verify they pass**

Run: `go test ./internal/collector/ -v -run "TestProbeTools|TestDetectSSH"`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/collector/tools.go internal/collector/tools_test.go \
       internal/collector/credentials.go internal/collector/credentials_test.go
git commit -m "feat: add tool probes and credential detection"
```

---

## Task 6: Distrobox Integration

**Files:**
- Create: `internal/collector/distrobox.go`
- Create: `internal/collector/distrobox_test.go`

**Step 1: Write distrobox list parsing test**

```go
// internal/collector/distrobox_test.go
package collector

import (
	"testing"

	"github.com/rommelporras/dotfiles/internal/model"
)

func TestParseDistroboxList(t *testing.T) {
	output := `ID           | NAME                 | STATUS             | IMAGE
abc123       | work-eam             | Up 2 hours         | ubuntu:24.04
def456       | personal             | Up 2 hours         | ubuntu:24.04
ghi789       | sandbox              | Created            | ubuntu:24.04
`
	got := parseDistroboxList(output)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}

	want := []model.ContainerInfo{
		{Name: "work-eam", Status: "running", Image: "ubuntu:24.04"},
		{Name: "personal", Status: "running", Image: "ubuntu:24.04"},
		{Name: "sandbox", Status: "stopped", Image: "ubuntu:24.04"},
	}

	for i, w := range want {
		if got[i].Name != w.Name {
			t.Errorf("[%d] Name = %q, want %q", i, got[i].Name, w.Name)
		}
		if got[i].Status != w.Status {
			t.Errorf("[%d] Status = %q, want %q", i, got[i].Status, w.Status)
		}
	}
}

func TestParseDistroboxListEmpty(t *testing.T) {
	got := parseDistroboxList("ID           | NAME                 | STATUS             | IMAGE\n")
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/collector/ -v -run TestParseDistrobox`
Expected: FAIL

**Step 3: Implement distrobox parsing**

```go
// internal/collector/distrobox.go
package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rommelporras/dotfiles/internal/model"
)

const containerTimeout = 30 * time.Second

func parseDistroboxList(output string) []model.ContainerInfo {
	var containers []model.ContainerInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		rawStatus := strings.TrimSpace(parts[2])
		image := strings.TrimSpace(parts[3])

		status := "stopped"
		if strings.HasPrefix(rawStatus, "Up") {
			status = "running"
		}

		containers = append(containers, model.ContainerInfo{
			Name:   name,
			Status: status,
			Image:  image,
		})
	}
	return containers
}

// ListContainers returns all distrobox containers.
func ListContainers(runner CommandRunner) ([]model.ContainerInfo, error) {
	out, err := runner.Run("distrobox", "list")
	if err != nil {
		return nil, fmt.Errorf("distrobox list: %w", err)
	}
	return parseDistroboxList(out), nil
}

// RunInContainer executes a shell command inside a container with timeout.
func RunInContainer(name, shellCmd string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), containerTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "distrobox", "enter", name, "--", "sh", "-c", shellCmd)
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("container %s: command timed out after %s", name, containerTimeout)
	}
	return string(out), err
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/collector/ -v -run TestParseDistrobox`
Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add internal/collector/distrobox.go internal/collector/distrobox_test.go
git commit -m "feat: add distrobox container listing and remote execution"
```

---

## Task 7: Collector Orchestrator

**Files:**
- Create: `internal/collector/collector.go`
- Create: `internal/collector/collector_test.go`

**Step 1: Write orchestrator test with mocks**

```go
// internal/collector/collector_test.go
package collector

import (
	"testing"
)

type mockRunner struct {
	outputs map[string]string // "cmd arg1 arg2" → output
}

func (m *mockRunner) Run(name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	if out, ok := m.outputs[key]; ok {
		return out, nil
	}
	return "", fmt.Errorf("mock: no output for %q", key)
}

func TestCollectLocal(t *testing.T) {
	runner := &mockRunner{outputs: map[string]string{
		"chezmoi status":              " M .zshrc\n",
		"chezmoi data --format json":  `{"context":"personal","platform":"aurora"}`,
	}}

	state, err := CollectLocal(runner, "test-host", "aurora")
	if err != nil {
		t.Fatalf("CollectLocal: %v", err)
	}
	if state.Hostname != "test-host" {
		t.Errorf("Hostname = %q", state.Hostname)
	}
	if state.Platform != "aurora" {
		t.Errorf("Platform = %q", state.Platform)
	}
	if len(state.DriftFiles) != 1 {
		t.Errorf("DriftFiles len = %d, want 1", len(state.DriftFiles))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/collector/ -v -run TestCollectLocal`
Expected: FAIL

**Step 3: Implement collector orchestrator**

```go
// internal/collector/collector.go
package collector

import (
	"github.com/rommelporras/dotfiles/internal/model"
)

// CollectLocal gathers state from the local machine.
func CollectLocal(runner CommandRunner, hostname, platform string) (*model.MachineState, error) {
	drift, tmplData, chezmoiErr := CollectChezmoi(runner)

	context := ""
	if tmplData != nil {
		if c, ok := tmplData["context"].(string); ok {
			context = c
		}
	}

	state := &model.MachineState{
		Hostname:     hostname,
		Platform:     platform,
		Context:      context,
		DriftFiles:   drift,
		TemplateData: tmplData,
		Tools:        ProbeTools(),
		SSHAgent:     DetectSSHAgent(),
		SetupCreds:   DetectSetupCreds(),
		AtuinSync:    DetectAtuinSync(),
	}

	if chezmoiErr != nil && drift == nil {
		// chezmoi not available at all — still return partial state
		state.DriftFiles = nil
	}

	return state, nil
}

// CollectAll gathers state from local machine + running distrobox containers.
func CollectAll(runner CommandRunner, hostname, platform string) (*model.CollectResult, error) {
	local, err := CollectLocal(runner, hostname, platform)
	if err != nil {
		return nil, err
	}

	result := &model.CollectResult{
		Machines: []model.MachineState{*local},
	}

	// Only enumerate containers on aurora host
	if platform == "aurora" {
		containers, err := ListContainers(runner)
		if err == nil {
			result.Containers = containers
			for _, c := range containers {
				if c.Status != "running" {
					continue
				}
				containerState, err := collectFromContainer(c.Name)
				if err != nil {
					continue // skip failed containers
				}
				result.Machines = append(result.Machines, *containerState)
			}
		}
	}

	return result, nil
}

func collectFromContainer(name string) (*model.MachineState, error) {
	// Collect chezmoi status from inside container
	statusOut, err := RunInContainer(name, "chezmoi status 2>/dev/null || $HOME/bin/chezmoi status 2>/dev/null || true")
	if err != nil {
		return nil, err
	}

	dataOut, err := RunInContainer(name, "chezmoi data --format json 2>/dev/null || $HOME/bin/chezmoi data --format json 2>/dev/null || echo '{}'")
	if err != nil {
		return nil, err
	}

	drift := parseChezmoiStatus(statusOut)
	tmplData, _ := parseChezmoiData(dataOut)

	context := ""
	if tmplData != nil {
		if c, ok := tmplData["context"].(string); ok {
			context = c
		}
	}

	// Probe tools inside container
	toolsScript := ""
	for _, tool := range TrackedTools {
		toolsScript += "command -v " + tool + " 2>/dev/null || echo ''; "
	}
	toolsOut, _ := RunInContainer(name, toolsScript)
	tools := make(map[string]string, len(TrackedTools))
	lines := strings.Split(strings.TrimSpace(toolsOut), "\n")
	for i, tool := range TrackedTools {
		if i < len(lines) {
			tools[tool] = strings.TrimSpace(lines[i])
		} else {
			tools[tool] = ""
		}
	}

	// SSH agent type
	sshOut, _ := RunInContainer(name, "echo $SSH_AUTH_SOCK")
	sshAgent := detectSSHAgentType(strings.TrimSpace(sshOut))

	// setup-creds
	setupOut, _ := RunInContainer(name, "test -x $HOME/.local/bin/setup-creds && echo ran || echo n/a")
	setupCreds := strings.TrimSpace(setupOut)

	// atuin
	atuinOut, _ := RunInContainer(name, "grep -q sync_address $HOME/.config/atuin/config.toml 2>/dev/null && echo synced || echo n/a")
	atuinSync := strings.TrimSpace(atuinOut)

	return &model.MachineState{
		Hostname:     name,
		Platform:     "distrobox",
		Context:      context,
		DriftFiles:   drift,
		TemplateData: tmplData,
		Tools:        tools,
		SSHAgent:     sshAgent,
		SetupCreds:   setupCreds,
		AtuinSync:    atuinSync,
	}, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/collector/ -v -run TestCollectLocal`
Expected: PASS

**Step 5: Run all collector tests**

Run: `go test ./internal/collector/ -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/collector/collector.go internal/collector/collector_test.go
git commit -m "feat: add collector orchestrator for local + container state"
```

---

## Task 8: OTLP Push

**Files:**
- Create: `internal/push/otel.go`
- Create: `internal/push/otel_test.go`

**Step 1: Install OTel dependencies**

```bash
go get go.opentelemetry.io/otel
go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
go get go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc
go get go.opentelemetry.io/otel/sdk/metric
go get go.opentelemetry.io/otel/sdk/log
```

**Step 2: Write push test (verifies serialization, not network)**

```go
// internal/push/otel_test.go
package push

import (
	"testing"

	"github.com/rommelporras/dotfiles/internal/model"
)

func TestBuildMetricAttributes(t *testing.T) {
	state := &model.MachineState{
		Hostname: "aurora-dx",
		Platform: "aurora",
		Context:  "personal",
		DriftFiles: []model.DriftFile{
			{Path: ".zshrc", Status: "M"},
			{Path: ".gitconfig", Status: "M"},
		},
		Tools: map[string]string{
			"glab":    "/usr/bin/glab",
			"kubectl": "/usr/local/bin/kubectl",
			"terraform": "",
		},
		SSHAgent:   "1password",
		SetupCreds: "n/a",
		AtuinSync:  "synced",
	}

	metrics := BuildMetrics(state)

	if metrics.DriftTotal != 2 {
		t.Errorf("DriftTotal = %d, want 2", metrics.DriftTotal)
	}
	if metrics.Up != 1 {
		t.Errorf("Up = %d, want 1", metrics.Up)
	}
	if metrics.ToolsInstalled["glab"] != 1 {
		t.Errorf("glab should be 1")
	}
	if metrics.ToolsInstalled["terraform"] != 0 {
		t.Errorf("terraform should be 0")
	}
	if metrics.Credentials["ssh_agent"] != 1 {
		t.Errorf("ssh_agent should be 1")
	}
}
```

**Step 3: Implement OTLP push**

```go
// internal/push/otel.go
package push

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"

	"github.com/rommelporras/dotfiles/internal/model"
)

// MetricSet is the intermediate representation before pushing.
type MetricSet struct {
	Hostname       string
	Platform       string
	Context        string
	Up             int64
	DriftTotal     int64
	ToolsInstalled map[string]int64
	Credentials    map[string]int64
	Timestamp      int64
}

// BuildMetrics converts a MachineState into pushable metrics.
func BuildMetrics(state *model.MachineState) *MetricSet {
	tools := make(map[string]int64, len(state.Tools))
	for name, path := range state.Tools {
		if path != "" {
			tools[name] = 1
		} else {
			tools[name] = 0
		}
	}

	creds := map[string]int64{
		"ssh_agent":   boolToInt(state.SSHAgent != "none" && state.SSHAgent != "n/a"),
		"setup_creds": boolToInt(state.SetupCreds == "ran"),
		"atuin_sync":  boolToInt(state.AtuinSync == "synced"),
	}

	return &MetricSet{
		Hostname:       state.Hostname,
		Platform:       state.Platform,
		Context:        state.Context,
		Up:             1,
		DriftTotal:     int64(len(state.DriftFiles)),
		ToolsInstalled: tools,
		Credentials:    creds,
		Timestamp:      time.Now().Unix(),
	}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// Push sends metrics and a log entry for a MachineState to the OTel Collector.
func Push(ctx context.Context, endpoint string, state *model.MachineState) error {
	// Metrics via OTLP gRPC
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("create metric exporter: %w", err)
	}
	defer exporter.Shutdown(ctx)

	res, _ := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			attribute.String("service.name", "dotctl"),
		),
	)

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(res),
	)
	defer provider.Shutdown(ctx)

	meter := provider.Meter("dotctl")
	ms := BuildMetrics(state)

	// Register gauges
	upGauge, _ := meter.Int64Gauge("dotctl_up")
	driftGauge, _ := meter.Int64Gauge("dotctl_drift_total")
	toolGauge, _ := meter.Int64Gauge("dotctl_tool_installed")
	credGauge, _ := meter.Int64Gauge("dotctl_credential_status")
	tsGauge, _ := meter.Int64Gauge("dotctl_collect_timestamp")

	hostAttr := attribute.String("hostname", ms.Hostname)
	platAttr := attribute.String("platform", ms.Platform)
	ctxAttr := attribute.String("context", ms.Context)

	upGauge.Record(ctx, ms.Up, otelmetric.WithAttributes(hostAttr, platAttr, ctxAttr))
	driftGauge.Record(ctx, ms.DriftTotal, otelmetric.WithAttributes(hostAttr))
	tsGauge.Record(ctx, ms.Timestamp, otelmetric.WithAttributes(hostAttr))

	for tool, val := range ms.ToolsInstalled {
		toolGauge.Record(ctx, val, otelmetric.WithAttributes(hostAttr, attribute.String("tool", tool)))
	}
	for cred, val := range ms.Credentials {
		credGauge.Record(ctx, val, otelmetric.WithAttributes(hostAttr, attribute.String("credential", cred)))
	}

	// Force flush metrics
	var rm metricdata.ResourceMetrics
	reader.Collect(ctx, &rm)
	if err := exporter.Export(ctx, &rm); err != nil {
		return fmt.Errorf("export metrics: %w", err)
	}

	return nil
}

// PushLog sends the full MachineState as a structured JSON log to the OTel Collector.
// This goes to Loki for detailed queries.
func PushLog(ctx context.Context, endpoint string, state *model.MachineState) error {
	// Serialize state to JSON for Loki
	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	_ = payload // TODO: implement OTLP log export
	// Log export uses otlploggrpc — implementation deferred to
	// integration testing phase since it requires OTel log SDK
	// which has a different API surface.
	return nil
}
```

Note: The OTLP log push has a placeholder. The OTel Go log SDK API is
newer and may need adjustments during integration testing. The metric
push is the priority — log push can be completed during Task 12.

**Step 4: Run tests**

Run: `go test ./internal/push/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/push/ go.mod go.sum
git commit -m "feat: add OTLP metric push to OTel Collector"
```

---

## Task 9: Prometheus Query Client

**Files:**
- Create: `internal/query/prometheus.go`
- Create: `internal/query/prometheus_test.go`

**Step 1: Write Prometheus query test with httptest**

```go
// internal/query/prometheus_test.go
package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryMachines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		query := r.URL.Query().Get("query")
		if query == "" {
			t.Error("empty query")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {"hostname": "aurora-dx", "platform": "aurora", "context": "personal"},
						"value": [1710000000, "1"]
					},
					{
						"metric": {"hostname": "work-eam", "platform": "distrobox", "context": "work-eam"},
						"value": [1710000000, "1"]
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewPrometheusClient(server.URL)
	machines, err := client.QueryMachines()
	if err != nil {
		t.Fatalf("QueryMachines: %v", err)
	}
	if len(machines) != 2 {
		t.Fatalf("len = %d, want 2", len(machines))
	}
	if machines[0].Hostname != "aurora-dx" {
		t.Errorf("machines[0].Hostname = %q", machines[0].Hostname)
	}
}

func TestQueryDriftTotal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{"metric": {"hostname": "aurora-dx"}, "value": [1710000000, "3"]}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewPrometheusClient(server.URL)
	drift, err := client.QueryDriftTotals()
	if err != nil {
		t.Fatalf("QueryDriftTotals: %v", err)
	}
	if drift["aurora-dx"] != 3 {
		t.Errorf("aurora-dx drift = %d, want 3", drift["aurora-dx"])
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/query/ -v`
Expected: FAIL

**Step 3: Implement Prometheus client**

```go
// internal/query/prometheus.go
package query

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PrometheusClient queries the Prometheus HTTP API.
type PrometheusClient struct {
	baseURL    string
	httpClient *http.Client
}

// MachineInfo is a summary from Prometheus labels.
type MachineInfo struct {
	Hostname string
	Platform string
	Context  string
}

type promResponse struct {
	Status string   `json:"status"`
	Data   promData `json:"data"`
}

type promData struct {
	ResultType string       `json:"resultType"`
	Result     []promResult `json:"result"`
}

type promResult struct {
	Metric map[string]string `json:"metric"`
	Value  [2]any            `json:"value"` // [timestamp, "value"]
}

// NewPrometheusClient creates a client for the given Prometheus URL.
func NewPrometheusClient(baseURL string) *PrometheusClient {
	return &PrometheusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *PrometheusClient) query(promql string) (*promResponse, error) {
	u := c.baseURL + "/api/v1/query?query=" + url.QueryEscape(promql)
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("prometheus query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result promResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus returned status: %s", result.Status)
	}
	return &result, nil
}

// QueryMachines returns all machines that have reported in.
func (c *PrometheusClient) QueryMachines() ([]MachineInfo, error) {
	resp, err := c.query("dotctl_up")
	if err != nil {
		return nil, err
	}

	var machines []MachineInfo
	for _, r := range resp.Data.Result {
		machines = append(machines, MachineInfo{
			Hostname: r.Metric["hostname"],
			Platform: r.Metric["platform"],
			Context:  r.Metric["context"],
		})
	}
	return machines, nil
}

// QueryDriftTotals returns drift count per hostname.
func (c *PrometheusClient) QueryDriftTotals() (map[string]int, error) {
	resp, err := c.query("dotctl_drift_total")
	if err != nil {
		return nil, err
	}

	drift := make(map[string]int)
	for _, r := range resp.Data.Result {
		hostname := r.Metric["hostname"]
		if valStr, ok := r.Value[1].(string); ok {
			val, _ := strconv.Atoi(valStr)
			drift[hostname] = val
		}
	}
	return drift, nil
}

// QueryToolsInstalled returns tool installation status per hostname.
func (c *PrometheusClient) QueryToolsInstalled() (map[string]map[string]bool, error) {
	resp, err := c.query("dotctl_tool_installed")
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]bool)
	for _, r := range resp.Data.Result {
		hostname := r.Metric["hostname"]
		tool := r.Metric["tool"]
		if result[hostname] == nil {
			result[hostname] = make(map[string]bool)
		}
		if valStr, ok := r.Value[1].(string); ok {
			result[hostname][tool] = valStr == "1"
		}
	}
	return result, nil
}

// QueryCredentials returns credential status per hostname.
func (c *PrometheusClient) QueryCredentials() (map[string]map[string]bool, error) {
	resp, err := c.query("dotctl_credential_status")
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]bool)
	for _, r := range resp.Data.Result {
		hostname := r.Metric["hostname"]
		cred := r.Metric["credential"]
		if result[hostname] == nil {
			result[hostname] = make(map[string]bool)
		}
		if valStr, ok := r.Value[1].(string); ok {
			result[hostname][cred] = valStr == "1"
		}
	}
	return result, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/query/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/query/prometheus.go internal/query/prometheus_test.go
git commit -m "feat: add Prometheus HTTP API query client"
```

---

## Task 10: Loki Query Client

**Files:**
- Create: `internal/query/loki.go`
- Create: `internal/query/loki_test.go`

**Step 1: Write Loki query test with httptest**

```go
// internal/query/loki_test.go
package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryLatestState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "streams",
				"result": [
					{
						"stream": {"service_name": "dotctl", "hostname": "aurora-dx"},
						"values": [
							["1710000000000000000", "{\"hostname\":\"aurora-dx\",\"drift_files\":[{\"path\":\".zshrc\",\"status\":\"M\"}],\"tools\":{\"glab\":\"/usr/bin/glab\"},\"ssh_agent\":\"1password\"}"]
						]
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLokiClient(server.URL)
	states, err := client.QueryLatestStates()
	if err != nil {
		t.Fatalf("QueryLatestStates: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("len = %d, want 1", len(states))
	}
	if states[0].Hostname != "aurora-dx" {
		t.Errorf("Hostname = %q", states[0].Hostname)
	}
	if len(states[0].DriftFiles) != 1 {
		t.Errorf("DriftFiles len = %d", len(states[0].DriftFiles))
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/query/ -v -run TestQueryLatest`
Expected: FAIL

**Step 3: Implement Loki client**

```go
// internal/query/loki.go
package query

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rommelporras/dotfiles/internal/model"
)

// LokiClient queries the Loki HTTP API.
type LokiClient struct {
	baseURL    string
	httpClient *http.Client
}

type lokiResponse struct {
	Status string   `json:"status"`
	Data   lokiData `json:"data"`
}

type lokiData struct {
	ResultType string       `json:"resultType"`
	Result     []lokiStream `json:"result"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [[nanosecond_ts, log_line], ...]
}

// NewLokiClient creates a client for the given Loki URL.
func NewLokiClient(baseURL string) *LokiClient {
	return &LokiClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueryLatestStates returns the most recent state log per machine.
func (c *LokiClient) QueryLatestStates() ([]model.MachineState, error) {
	// Query last 30 minutes of dotctl logs, one per hostname
	logql := `{service_name="dotctl"}`
	since := time.Now().Add(-30 * time.Minute)

	u := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%d&end=%d&limit=100&direction=backward",
		c.baseURL,
		url.QueryEscape(logql),
		since.UnixNano(),
		time.Now().UnixNano(),
	)

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("loki query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result lokiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Deduplicate: keep only the latest entry per hostname
	seen := make(map[string]bool)
	var states []model.MachineState

	for _, stream := range result.Data.Result {
		for _, entry := range stream.Values {
			var state model.MachineState
			if err := json.Unmarshal([]byte(entry[1]), &state); err != nil {
				continue
			}
			if seen[state.Hostname] {
				continue
			}
			seen[state.Hostname] = true
			states = append(states, state)
		}
	}

	return states, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/query/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/query/loki.go internal/query/loki_test.go
git commit -m "feat: add Loki HTTP API query client"
```

---

## Task 11: Terminal Display

**Files:**
- Create: `internal/display/table.go`
- Create: `internal/display/table_test.go`

**Step 1: Install lipgloss**

```bash
go get github.com/charmbracelet/lipgloss
```

**Step 2: Write display test**

```go
// internal/display/table_test.go
package display

import (
	"strings"
	"testing"

	"github.com/rommelporras/dotfiles/internal/model"
)

func TestRenderMachinesTable(t *testing.T) {
	machines := []model.MachineState{
		{Hostname: "aurora-dx", Platform: "aurora", Context: "personal",
			DriftFiles: []model.DriftFile{{Path: ".zshrc", Status: "M"}}},
		{Hostname: "work-eam", Platform: "distrobox", Context: "work-eam"},
	}

	output := RenderMachinesTable(machines)

	if !strings.Contains(output, "aurora-dx") {
		t.Error("output should contain aurora-dx")
	}
	if !strings.Contains(output, "work-eam") {
		t.Error("output should contain work-eam")
	}
	if !strings.Contains(output, "1") { // drift count
		t.Error("output should show drift count 1")
	}
}

func TestRenderToolsGrid(t *testing.T) {
	machines := []model.MachineState{
		{Hostname: "aurora-dx", Tools: map[string]string{
			"glab": "/usr/bin/glab", "kubectl": "", "terraform": "",
		}},
	}

	output := RenderToolsGrid(machines)

	if !strings.Contains(output, "aurora-dx") {
		t.Error("output should contain aurora-dx")
	}
}
```

**Step 3: Implement display rendering**

```go
// internal/display/table.go
package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/rommelporras/dotfiles/internal/collector"
	"github.com/rommelporras/dotfiles/internal/model"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	passStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12"))
)

// RenderAll renders the complete dashboard output.
func RenderAll(machines []model.MachineState, containers []model.ContainerInfo) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("dotctl — dotfiles status"))
	sb.WriteString("\n\n")
	sb.WriteString(RenderMachinesTable(machines))
	sb.WriteString("\n")
	sb.WriteString(RenderDriftDetails(machines))
	sb.WriteString(RenderToolsGrid(machines))
	sb.WriteString("\n")
	sb.WriteString(RenderCredentials(machines))
	return sb.String()
}

// RenderMachinesTable renders the machine summary table.
func RenderMachinesTable(machines []model.MachineState) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Machines"))
	sb.WriteString("\n")

	for _, m := range machines {
		drift := len(m.DriftFiles)
		driftStr := passStyle.Render("clean")
		if drift > 0 {
			driftStr = failStyle.Render(fmt.Sprintf("%d files", drift))
		}

		label := fmt.Sprintf("%s/%s", m.Platform, m.Context)
		sb.WriteString(fmt.Sprintf(" %-20s %-25s %s\n", m.Hostname, dimStyle.Render(label), driftStr))
	}
	return sb.String()
}

// RenderDriftDetails shows which files have drifted per machine.
func RenderDriftDetails(machines []model.MachineState) string {
	var sb strings.Builder
	for _, m := range machines {
		if len(m.DriftFiles) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n %s\n", headerStyle.Render("Drift: "+m.Hostname)))
		for _, f := range m.DriftFiles {
			sb.WriteString(fmt.Sprintf("   %s %s\n", failStyle.Render(f.Status), f.Path))
		}
	}
	return sb.String()
}

// RenderToolsGrid renders the tool installation matrix.
func RenderToolsGrid(machines []model.MachineState) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Tools"))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(fmt.Sprintf(" %-16s", ""))
	for _, tool := range collector.TrackedTools {
		sb.WriteString(fmt.Sprintf("%-10s", tool))
	}
	sb.WriteString("\n")

	// Data rows
	for _, m := range machines {
		sb.WriteString(fmt.Sprintf(" %-16s", m.Hostname))
		for _, tool := range collector.TrackedTools {
			if path := m.Tools[tool]; path != "" {
				sb.WriteString(fmt.Sprintf("%-10s", passStyle.Render("yes")))
			} else {
				sb.WriteString(fmt.Sprintf("%-10s", dimStyle.Render("—")))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// RenderCredentials renders the credential status table.
func RenderCredentials(machines []model.MachineState) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Credentials"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(" %-16s %-14s %-14s %-14s\n", "", "SSH Agent", "setup-creds", "Atuin Sync"))

	for _, m := range machines {
		sshStyle := dimStyle
		if m.SSHAgent == "1password" {
			sshStyle = passStyle
		}
		setupStyle := dimStyle
		if m.SetupCreds == "ran" {
			setupStyle = passStyle
		}
		atuinStyle := dimStyle
		if m.AtuinSync == "synced" {
			atuinStyle = passStyle
		}

		sb.WriteString(fmt.Sprintf(" %-16s %-14s %-14s %-14s\n",
			m.Hostname,
			sshStyle.Render(m.SSHAgent),
			setupStyle.Render(m.SetupCreds),
			atuinStyle.Render(m.AtuinSync),
		))
	}
	return sb.String()
}
```

**Step 4: Run tests**

Run: `go test ./internal/display/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/display/ go.mod go.sum
git commit -m "feat: add terminal dashboard rendering with lipgloss"
```

---

## Task 12: CLI Wiring with Cobra

**Files:**
- Modify: `cmd/dotctl/main.go`

**Step 1: Install cobra**

```bash
go get github.com/spf13/cobra
```

**Step 2: Implement CLI commands**

```go
// cmd/dotctl/main.go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rommelporras/dotfiles/internal/collector"
	"github.com/rommelporras/dotfiles/internal/config"
	"github.com/rommelporras/dotfiles/internal/display"
	"github.com/rommelporras/dotfiles/internal/push"
	"github.com/rommelporras/dotfiles/internal/query"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dotctl",
		Short: "Dotfiles status dashboard",
	}

	// dotctl status
	var live bool
	var machine string
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show dotfiles status across all machines",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load(config.DefaultPath())

			if live {
				return runStatusLive(cfg, machine)
			}
			return runStatusRemote(cfg, machine)
		},
	}
	statusCmd.Flags().BoolVar(&live, "live", false, "Collect fresh data locally instead of querying Prometheus")
	statusCmd.Flags().StringVar(&machine, "machine", "", "Filter to a single machine")
	rootCmd.AddCommand(statusCmd)

	// dotctl collect
	var container string
	var verbose bool
	collectCmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect status and push to OTel Collector",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load(config.DefaultPath())
			return runCollect(cfg, container, verbose)
		},
	}
	collectCmd.Flags().StringVar(&container, "container", "", "Collect from a single container only")
	collectCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	rootCmd.AddCommand(collectCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runStatusLive(cfg *config.Config, filterMachine string) error {
	runner := &collector.ExecRunner{}
	platform := collector.DetectPlatform()
	result, err := collector.CollectAll(runner, cfg.Hostname, platform)
	if err != nil {
		return err
	}

	machines := result.Machines
	if filterMachine != "" {
		machines = filterByHostname(machines, filterMachine)
	}

	fmt.Println(display.RenderAll(machines, result.Containers))
	return nil
}

func runStatusRemote(cfg *config.Config, filterMachine string) error {
	// Try Prometheus first
	promClient := query.NewPrometheusClient(cfg.PrometheusURL)
	promMachines, err := promClient.QueryMachines()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Prometheus unreachable (%v), falling back to --live\n", err)
		return runStatusLive(cfg, filterMachine)
	}

	driftTotals, _ := promClient.QueryDriftTotals()
	toolsMap, _ := promClient.QueryToolsInstalled()
	credsMap, _ := promClient.QueryCredentials()

	// Try Loki for detailed state
	lokiClient := query.NewLokiClient(cfg.LokiURL)
	lokiStates, _ := lokiClient.QueryLatestStates()

	// Build MachineState from combined sources
	lokiByHost := make(map[string]*model.MachineState)
	for i := range lokiStates {
		lokiByHost[lokiStates[i].Hostname] = &lokiStates[i]
	}

	var machines []model.MachineState
	for _, pm := range promMachines {
		ms := model.MachineState{
			Hostname: pm.Hostname,
			Platform: pm.Platform,
			Context:  pm.Context,
		}

		// Drift details from Loki, count from Prometheus
		if ls, ok := lokiByHost[pm.Hostname]; ok {
			ms.DriftFiles = ls.DriftFiles
			ms.TemplateData = ls.TemplateData
			ms.SSHAgent = ls.SSHAgent
			ms.SetupCreds = ls.SetupCreds
			ms.AtuinSync = ls.AtuinSync
		} else {
			// Prometheus only — count without details
			if count, ok := driftTotals[pm.Hostname]; ok && count > 0 {
				ms.DriftFiles = make([]model.DriftFile, count) // placeholder
			}
		}

		// Tools from Prometheus
		ms.Tools = make(map[string]string)
		if tools, ok := toolsMap[pm.Hostname]; ok {
			for tool, installed := range tools {
				if installed {
					ms.Tools[tool] = "installed"
				} else {
					ms.Tools[tool] = ""
				}
			}
		}

		// Credentials from Prometheus
		if creds, ok := credsMap[pm.Hostname]; ok {
			if creds["ssh_agent"] {
				ms.SSHAgent = "active"
			}
			if creds["setup_creds"] {
				ms.SetupCreds = "ran"
			}
			if creds["atuin_sync"] {
				ms.AtuinSync = "synced"
			}
		}

		machines = append(machines, ms)
	}

	if filterMachine != "" {
		machines = filterByHostname(machines, filterMachine)
	}

	fmt.Println(display.RenderAll(machines, nil))
	return nil
}

func runCollect(cfg *config.Config, container string, verbose bool) error {
	runner := &collector.ExecRunner{}
	platform := collector.DetectPlatform()

	result, err := collector.CollectAll(runner, cfg.Hostname, platform)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, ms := range result.Machines {
		if container != "" && ms.Hostname != container {
			continue
		}
		if verbose {
			fmt.Printf("Pushing metrics for %s...\n", ms.Hostname)
		}
		if err := push.Push(ctx, cfg.OTelEndpoint, &ms); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to push %s: %v\n", ms.Hostname, err)
		}
	}

	if verbose {
		fmt.Printf("Collection complete: %d machines\n", len(result.Machines))
	}
	return nil
}

func filterByHostname(machines []model.MachineState, name string) []model.MachineState {
	var filtered []model.MachineState
	for _, m := range machines {
		if m.Hostname == name {
			filtered = append(filtered, m)
		}
	}
	return filtered
}
```

Note: Add `"github.com/rommelporras/dotfiles/internal/model"` to imports.
The exact imports will be adjusted during implementation based on the
final API of each package.

**Step 3: Build and test CLI**

Run: `make build && ./dotctl --help`
Expected: Shows usage with `status` and `collect` subcommands

Run: `./dotctl status --help`
Expected: Shows `--live` and `--machine` flags

Run: `./dotctl collect --help`
Expected: Shows `--container` and `--verbose` flags

**Step 4: Commit**

```bash
git add cmd/dotctl/main.go go.mod go.sum
git commit -m "feat: wire CLI commands with cobra (status + collect)"
```

---

## Task 13: Systemd Units & Make Install

**Files:**
- Create: `deploy/dotctl-collect.service`
- Create: `deploy/dotctl-collect.timer`
- Modify: `Makefile`

**Step 1: Write systemd service unit**

```ini
# deploy/dotctl-collect.service
[Unit]
Description=dotctl dotfiles status collector
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
ExecStart=%h/.local/bin/dotctl collect
Environment=HOME=%h

[Install]
WantedBy=default.target
```

**Step 2: Write systemd timer unit**

```ini
# deploy/dotctl-collect.timer
[Unit]
Description=Run dotctl collect every 10 minutes

[Timer]
OnBootSec=2min
OnUnitActiveSec=10min
RandomizedDelaySec=30

[Install]
WantedBy=timers.target
```

**Step 3: Update Makefile with install-systemd target**

Add to existing Makefile:

```makefile
SYSTEMD_DIR := $(HOME)/.config/systemd/user

install-systemd: install
	mkdir -p $(SYSTEMD_DIR)
	cp deploy/dotctl-collect.service $(SYSTEMD_DIR)/
	cp deploy/dotctl-collect.timer $(SYSTEMD_DIR)/
	systemctl --user daemon-reload
	systemctl --user enable --now dotctl-collect.timer
	@echo "Timer installed. Check: systemctl --user status dotctl-collect.timer"

uninstall-systemd:
	systemctl --user disable --now dotctl-collect.timer || true
	rm -f $(SYSTEMD_DIR)/dotctl-collect.service $(SYSTEMD_DIR)/dotctl-collect.timer
	systemctl --user daemon-reload
```

**Step 4: Commit**

```bash
git add deploy/ Makefile
git commit -m "feat: add systemd timer for periodic collection"
```

---

## Task 14: Integration Test & Polish

**Files:**
- Modify: `cmd/dotctl/main.go` (fix compilation issues)
- Run full test suite

**Step 1: Fix any compilation issues**

Run: `make build`
Fix any import errors, type mismatches, or missing references.

**Step 2: Run full test suite**

Run: `make test`
Expected: All tests pass

**Step 3: Run lint**

Run: `make lint`
Expected: No issues

**Step 4: Test --live mode manually on Aurora**

Run: `./dotctl status --live`
Expected: Shows machine table with Aurora host status, tools, credentials.

**Step 5: Test collect (will fail on push if OTel Collector unreachable — that's OK)**

Run: `./dotctl collect --verbose`
Expected: Collects local data, attempts push, reports result.

**Step 6: Commit final polish**

```bash
git add -A
git commit -m "chore: fix compilation issues and verify integration"
```

---

## Task 15: Complete OTLP Log Push (deferred from Task 8)

**Files:**
- Modify: `internal/push/otel.go`

Implement the `PushLog` function using `otlploggrpc`. This requires the
OTel Go log SDK (`go.opentelemetry.io/otel/sdk/log`). Consult Context7
or the OTel Go docs for the current API surface, as the log SDK graduated
to stable more recently than the metric SDK.

**Step 1: Research current OTel Go log SDK API**

Check Context7 for `go.opentelemetry.io/otel` documentation.

**Step 2: Implement PushLog**

Serialize `MachineState` as a JSON string in the log body. Set resource
attributes: `service.name=dotctl`. Set log attributes: `hostname`, `platform`,
`context`.

**Step 3: Test with real OTel Collector**

Run: `./dotctl collect --verbose`
Verify logs appear in Loki:
```logql
{service_name="dotctl"}
```

**Step 4: Commit**

```bash
git add internal/push/otel.go go.mod go.sum
git commit -m "feat: add OTLP log push to Loki via OTel Collector"
```

---

## Future Work (not in this plan)

These items are documented for reference but not implemented now:

- **Grafana dashboard** — ConfigMap for auto-provisioning (homelab repo, after infra review)
- **PrometheusRules** — alert definitions for stale collection and drift (homelab repo)
- **Loki HTTPRoute** — expose Loki externally for CLI queries (homelab repo, security review)
- **WSL2 support** — same binary, cron trigger, Tailscale for OTel endpoint access
- **Windows support** — cross-compile with `GOOS=windows`, Task Scheduler instead of systemd
- **Container running gauge** — `dotctl_container_running` metric from `distrobox list` output
