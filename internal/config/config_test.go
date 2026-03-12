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
