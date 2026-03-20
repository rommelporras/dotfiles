package collector

import (
	"os"
	"path/filepath"
	"strings"
)

// TrackedClaudeLinks lists the items that should be symlinked from ~/.claude/ to claude-config.
var TrackedClaudeLinks = []string{"CLAUDE.md", "settings.json", "rules", "hooks", "skills", "agents"}

// checkClaudeLink checks a single symlink and returns its status.
// expectSubstr is the substring the target path must contain (e.g. "claude-config").
func checkClaudeLink(linkPath, expectSubstr string) string {
	info, err := os.Lstat(linkPath)
	if err != nil {
		return "missing"
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return "file"
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		return "wrong"
	}

	if strings.Contains(target, expectSubstr) {
		return "ok"
	}
	return "wrong"
}

// detectClaudeLinksWith checks all tracked symlinks using explicit paths (for testing).
func detectClaudeLinksWith(claudeDir, configDir string) map[string]string {
	links := make(map[string]string, len(TrackedClaudeLinks))
	for _, item := range TrackedClaudeLinks {
		linkPath := filepath.Join(claudeDir, item)
		links[item] = checkClaudeLink(linkPath, "claude-config")
	}
	return links
}

// DetectClaudeLinks checks claude-config symlinks for the local machine.
func DetectClaudeLinks() map[string]string {
	home, _ := os.UserHomeDir()
	claudeDir := filepath.Join(home, ".claude")
	configDir := filepath.Join(home, "personal", "claude-config")
	return detectClaudeLinksWith(claudeDir, configDir)
}

// parseClaudeLinkStatus normalizes shell output from container checks.
func parseClaudeLinkStatus(output string) string {
	s := strings.TrimSpace(output)
	switch s {
	case "ok", "wrong", "file", "missing":
		return s
	default:
		return "missing"
	}
}
