package push

import (
	"testing"

	"github.com/rommelporras/dotfiles/dotctl/internal/model"
)

func TestBuildMetrics(t *testing.T) {
	state := &model.MachineState{
		Hostname: "aurora-dx",
		Platform: "aurora",
		Context:  "personal",
		DriftFiles: []model.DriftFile{
			{Path: ".zshrc", Status: "M"},
			{Path: ".gitconfig", Status: "M"},
		},
		Tools: map[string]string{
			"glab":      "/usr/bin/glab",
			"kubectl":   "/usr/local/bin/kubectl",
			"terraform": "",
		},
		SSHAgent:   "1password",
		SetupCreds: "n/a",
		AtuinSync:  "synced",
		ClaudeLinks: map[string]string{
			"CLAUDE.md":     "ok",
			"settings.json": "ok",
			"rules":         "ok",
			"hooks":         "ok",
			"skills":        "ok",
			"agents":        "wrong",
		},
	}

	ms := BuildMetrics(state)

	if ms.DriftTotal != 2 {
		t.Errorf("DriftTotal = %d, want 2", ms.DriftTotal)
	}
	if ms.Up != 1 {
		t.Errorf("Up = %d, want 1", ms.Up)
	}
	if ms.ToolsInstalled["glab"] != 1 {
		t.Errorf("glab should be 1, got %d", ms.ToolsInstalled["glab"])
	}
	if ms.ToolsInstalled["terraform"] != 0 {
		t.Errorf("terraform should be 0, got %d", ms.ToolsInstalled["terraform"])
	}
	if ms.Credentials["ssh_agent"] != 1 {
		t.Errorf("ssh_agent should be 1, got %d", ms.Credentials["ssh_agent"])
	}
	if ms.Credentials["setup_creds"] != 0 {
		t.Errorf("setup_creds should be 0 (n/a), got %d", ms.Credentials["setup_creds"])
	}
	if ms.Credentials["atuin_sync"] != 1 {
		t.Errorf("atuin_sync should be 1, got %d", ms.Credentials["atuin_sync"])
	}
	if ms.ClaudeLinks["CLAUDE.md"] != 1 {
		t.Errorf("claude CLAUDE.md should be 1, got %d", ms.ClaudeLinks["CLAUDE.md"])
	}
	if ms.ClaudeLinks["agents"] != 0 {
		t.Errorf("claude agents should be 0 (wrong), got %d", ms.ClaudeLinks["agents"])
	}
	if ms.Hostname != "aurora-dx" {
		t.Errorf("Hostname = %q", ms.Hostname)
	}
}

func TestBuildMetricsBoolConversions(t *testing.T) {
	tests := []struct {
		sshAgent   string
		setupCreds string
		atuinSync  string
		wantSSH    int64
		wantSetup  int64
		wantAtuin  int64
	}{
		{"1password", "ran", "synced", 1, 1, 1},
		{"system", "present", "disabled", 1, 0, 0},
		{"none", "n/a", "n/a", 0, 0, 0},
	}

	for _, tt := range tests {
		state := &model.MachineState{
			Hostname:   "host",
			SSHAgent:   tt.sshAgent,
			SetupCreds: tt.setupCreds,
			AtuinSync:  tt.atuinSync,
			Tools:      map[string]string{},
		}
		ms := BuildMetrics(state)
		if ms.Credentials["ssh_agent"] != tt.wantSSH {
			t.Errorf("[%s] ssh_agent = %d, want %d", tt.sshAgent, ms.Credentials["ssh_agent"], tt.wantSSH)
		}
		if ms.Credentials["setup_creds"] != tt.wantSetup {
			t.Errorf("[%s] setup_creds = %d, want %d", tt.setupCreds, ms.Credentials["setup_creds"], tt.wantSetup)
		}
		if ms.Credentials["atuin_sync"] != tt.wantAtuin {
			t.Errorf("[%s] atuin_sync = %d, want %d", tt.atuinSync, ms.Credentials["atuin_sync"], tt.wantAtuin)
		}
	}
}
