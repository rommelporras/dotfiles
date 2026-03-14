package collector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckClaudeLink(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(dir string) string // returns link path
		expect string
		want   string
	}{
		{
			name: "valid symlink to claude-config",
			setup: func(dir string) string {
				target := filepath.Join(dir, "claude-config", "CLAUDE.md")
				os.MkdirAll(filepath.Join(dir, "claude-config"), 0o755)
				os.WriteFile(target, []byte("test"), 0o644)
				link := filepath.Join(dir, ".claude", "CLAUDE.md")
				os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
				os.Symlink(target, link)
				return link
			},
			expect: "claude-config",
			want:   "ok",
		},
		{
			name: "symlink to wrong target",
			setup: func(dir string) string {
				target := filepath.Join(dir, "other-repo", "CLAUDE.md")
				os.MkdirAll(filepath.Join(dir, "other-repo"), 0o755)
				os.WriteFile(target, []byte("test"), 0o644)
				link := filepath.Join(dir, ".claude", "CLAUDE.md")
				os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
				os.Symlink(target, link)
				return link
			},
			expect: "claude-config",
			want:   "wrong",
		},
		{
			name: "regular file not a symlink",
			setup: func(dir string) string {
				path := filepath.Join(dir, ".claude", "CLAUDE.md")
				os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
				os.WriteFile(path, []byte("test"), 0o644)
				return path
			},
			expect: "claude-config",
			want:   "file",
		},
		{
			name: "missing file",
			setup: func(dir string) string {
				os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
				return filepath.Join(dir, ".claude", "CLAUDE.md")
			},
			expect: "claude-config",
			want:   "missing",
		},
		{
			name: "directory symlink ok",
			setup: func(dir string) string {
				target := filepath.Join(dir, "claude-config", "hooks")
				os.MkdirAll(target, 0o755)
				link := filepath.Join(dir, ".claude", "hooks")
				os.MkdirAll(filepath.Join(dir, ".claude"), 0o755)
				os.Symlink(target, link)
				return link
			},
			expect: "claude-config",
			want:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			linkPath := tt.setup(dir)
			got := checkClaudeLink(linkPath, tt.expect)
			if got != tt.want {
				t.Errorf("checkClaudeLink() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectClaudeLinksWith(t *testing.T) {
	dir := t.TempDir()

	// Create fake claude-config repo
	configDir := filepath.Join(dir, "personal", "claude-config")
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(configDir, 0o755)
	os.MkdirAll(claudeDir, 0o755)

	// Create targets and symlinks for some items
	for _, item := range []string{"CLAUDE.md", "settings.json"} {
		os.WriteFile(filepath.Join(configDir, item), []byte("test"), 0o644)
		os.Symlink(filepath.Join(configDir, item), filepath.Join(claudeDir, item))
	}
	for _, item := range []string{"rules", "hooks"} {
		os.MkdirAll(filepath.Join(configDir, item), 0o755)
		os.Symlink(filepath.Join(configDir, item), filepath.Join(claudeDir, item))
	}
	// skills: regular dir (not symlink)
	os.MkdirAll(filepath.Join(claudeDir, "skills"), 0o755)
	// agents: missing entirely

	links := detectClaudeLinksWith(claudeDir, configDir)

	if links["CLAUDE.md"] != "ok" {
		t.Errorf("CLAUDE.md = %q, want ok", links["CLAUDE.md"])
	}
	if links["settings.json"] != "ok" {
		t.Errorf("settings.json = %q, want ok", links["settings.json"])
	}
	if links["rules"] != "ok" {
		t.Errorf("rules = %q, want ok", links["rules"])
	}
	if links["hooks"] != "ok" {
		t.Errorf("hooks = %q, want ok", links["hooks"])
	}
	if links["skills"] != "file" {
		t.Errorf("skills = %q, want file", links["skills"])
	}
	if links["agents"] != "missing" {
		t.Errorf("agents = %q, want missing", links["agents"])
	}
}

func TestParseClaudeLinkStatus(t *testing.T) {
	tests := []struct {
		output string
		want   string
	}{
		{"ok", "ok"},
		{"wrong", "wrong"},
		{"file", "file"},
		{"missing", "missing"},
		{"", "missing"},
		{"  ok  ", "ok"},
	}

	for _, tt := range tests {
		got := parseClaudeLinkStatus(tt.output)
		if got != tt.want {
			t.Errorf("parseClaudeLinkStatus(%q) = %q, want %q", tt.output, got, tt.want)
		}
	}
}
