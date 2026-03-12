package collector

import (
	"fmt"
	"strings"
	"testing"
)

type mockRunner struct {
	outputs map[string]string // "cmd arg1 arg2" -> output
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
		"chezmoi status":             " M .zshrc\n",
		"chezmoi data --format json": `{"context":"personal","platform":"aurora"}`,
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
	if state.Context != "personal" {
		t.Errorf("Context = %q, want personal", state.Context)
	}
	if len(state.DriftFiles) != 1 {
		t.Errorf("DriftFiles len = %d, want 1", len(state.DriftFiles))
	}
	if state.DriftFiles[0].Path != ".zshrc" {
		t.Errorf("DriftFiles[0].Path = %q", state.DriftFiles[0].Path)
	}
}

func TestCollectLocalChezmoiUnavailable(t *testing.T) {
	runner := &mockRunner{outputs: map[string]string{}}

	state, err := CollectLocal(runner, "test-host", "distrobox")
	if err != nil {
		t.Fatalf("CollectLocal should not error on missing chezmoi: %v", err)
	}
	if state.Hostname != "test-host" {
		t.Errorf("Hostname = %q", state.Hostname)
	}
	if state.DriftFiles != nil {
		t.Errorf("DriftFiles should be nil when chezmoi unavailable, got %v", state.DriftFiles)
	}
}
