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
