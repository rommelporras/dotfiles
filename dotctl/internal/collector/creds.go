package collector

import (
	"os"
	"strings"
)

// detectSSHAgentType classifies the SSH agent from the socket path.
func detectSSHAgentType(sock string) string {
	if sock == "" {
		return "none"
	}
	if strings.Contains(sock, "1password") {
		return "1password"
	}
	return "system"
}

// DetectSSHAgent checks the current SSH_AUTH_SOCK.
func DetectSSHAgent() string {
	return detectSSHAgentType(os.Getenv("SSH_AUTH_SOCK"))
}

// DetectSetupCreds checks if setup-creds has been deployed.
func DetectSetupCreds() string {
	path := os.Getenv("HOME") + "/.local/bin/setup-creds"
	info, err := os.Stat(path)
	if err == nil && info.Mode()&0o111 != 0 {
		return "ran"
	}
	return "n/a"
}

// DetectAtuinSync checks if Atuin is configured for sync.
func DetectAtuinSync() string {
	configPath := os.Getenv("HOME") + "/.config/atuin/config.toml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "n/a"
	}
	if strings.Contains(string(data), "sync_address") {
		return "synced"
	}
	return "n/a"
}
