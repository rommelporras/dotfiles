package collector

import (
	"fmt"
	"testing"
)

type mockLookPath struct {
	found map[string]string
}

func (m *mockLookPath) LookPath(name string) (string, error) {
	if p, ok := m.found[name]; ok {
		return p, nil
	}
	return "", fmt.Errorf("%s: not found", name)
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
