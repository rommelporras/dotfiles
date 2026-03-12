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
