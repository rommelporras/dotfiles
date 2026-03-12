package display

import (
	"strings"
	"testing"

	"github.com/rommelporras/dotfiles/dotctl/internal/model"
)

func TestRenderMachinesTable(t *testing.T) {
	machines := []model.MachineState{
		{
			Hostname:   "aurora-dx",
			Platform:   "aurora",
			Context:    "personal",
			DriftFiles: []model.DriftFile{{Path: ".zshrc", Status: "M"}, {Path: ".gitconfig", Status: "M"}},
		},
		{
			Hostname: "work-eam",
			Platform: "distrobox",
			Context:  "work-eam",
		},
	}

	output := RenderMachinesTable(machines)

	if !strings.Contains(output, "aurora-dx") {
		t.Error("output should contain aurora-dx")
	}
	if !strings.Contains(output, "work-eam") {
		t.Error("output should contain work-eam")
	}
	// aurora-dx has 2 drift files
	if !strings.Contains(output, "2") {
		t.Error("output should show drift count 2")
	}
}

func TestRenderDriftDetails(t *testing.T) {
	machines := []model.MachineState{
		{
			Hostname: "aurora-dx",
			DriftFiles: []model.DriftFile{
				{Path: ".zshrc", Status: "M"},
				{Path: ".gitconfig", Status: "A"},
			},
		},
		{
			Hostname:   "work-eam",
			DriftFiles: nil, // no drift
		},
	}

	output := RenderDriftDetails(machines)

	if !strings.Contains(output, ".zshrc") {
		t.Error("output should contain .zshrc")
	}
	if !strings.Contains(output, ".gitconfig") {
		t.Error("output should contain .gitconfig")
	}
	if strings.Contains(output, "work-eam") {
		t.Error("output should NOT contain work-eam (no drift)")
	}
}

func TestRenderToolsGrid(t *testing.T) {
	machines := []model.MachineState{
		{
			Hostname: "aurora-dx",
			Tools: map[string]string{
				"glab": "/usr/bin/glab", "kubectl": "/usr/local/bin/kubectl",
				"terraform": "", "aws": "", "ansible": "", "op": "", "atuin": "", "bun": "",
			},
		},
	}

	output := RenderToolsGrid(machines)

	if !strings.Contains(output, "aurora-dx") {
		t.Error("output should contain aurora-dx")
	}
	if !strings.Contains(output, "glab") {
		t.Error("output should contain glab header")
	}
}

func TestRenderCredentials(t *testing.T) {
	machines := []model.MachineState{
		{Hostname: "aurora-dx", SSHAgent: "1password", SetupCreds: "n/a", AtuinSync: "synced"},
		{Hostname: "work-eam", SSHAgent: "1password", SetupCreds: "ran", AtuinSync: "synced"},
	}

	output := RenderCredentials(machines)

	if !strings.Contains(output, "aurora-dx") {
		t.Error("output should contain aurora-dx")
	}
	if !strings.Contains(output, "1password") {
		t.Error("output should contain 1password")
	}
}
