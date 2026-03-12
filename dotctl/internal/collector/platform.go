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
