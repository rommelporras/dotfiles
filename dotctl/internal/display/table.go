package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/rommelporras/dotfiles/dotctl/internal/collector"
	"github.com/rommelporras/dotfiles/dotctl/internal/model"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	passStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	failStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle  = lipgloss.NewStyle().Bold(true).Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12"))
)

// col pads s to width visible characters, accounting for ANSI escape codes.
func col(s string, width int) string {
	visible := lipgloss.Width(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

// RenderAll renders the complete dashboard output.
func RenderAll(machines []model.MachineState, containers []model.ContainerInfo) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("dotctl — dotfiles status"))
	sb.WriteString("\n\n")
	sb.WriteString(RenderMachinesTable(machines))
	sb.WriteString(RenderDriftDetails(machines))
	sb.WriteString("\n")
	sb.WriteString(RenderToolsGrid(machines))
	sb.WriteString("\n")
	sb.WriteString(RenderCredentials(machines))
	return sb.String()
}

// RenderMachinesTable renders the machine summary table.
func RenderMachinesTable(machines []model.MachineState) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Machines"))
	sb.WriteString("\n")

	for _, m := range machines {
		drift := len(m.DriftFiles)
		driftStr := passStyle.Render("clean")
		if drift > 0 {
			driftStr = failStyle.Render(fmt.Sprintf("%d files", drift))
		}
		label := fmt.Sprintf("%s/%s", m.Platform, m.Context)
		sb.WriteString(" " + col(m.Hostname, 20) + " " + col(dimStyle.Render(label), 25) + " " + driftStr + "\n")
	}
	return sb.String()
}

// RenderDriftDetails shows which files have drifted per machine.
func RenderDriftDetails(machines []model.MachineState) string {
	var sb strings.Builder
	for _, m := range machines {
		if len(m.DriftFiles) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n %s\n", headerStyle.Render("Drift: "+m.Hostname)))
		for _, f := range m.DriftFiles {
			sb.WriteString(fmt.Sprintf("   %s %s\n", failStyle.Render(f.Status), f.Path))
		}
	}
	return sb.String()
}

// RenderToolsGrid renders the tool installation matrix.
func RenderToolsGrid(machines []model.MachineState) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Tools"))
	sb.WriteString("\n")

	// Header row
	sb.WriteString(" " + col("", 16))
	for _, tool := range collector.TrackedTools {
		sb.WriteString(col(tool, 11))
	}
	sb.WriteString("\n")

	// Data rows
	for _, m := range machines {
		sb.WriteString(" " + col(m.Hostname, 16))
		for _, tool := range collector.TrackedTools {
			if path := m.Tools[tool]; path != "" {
				sb.WriteString(col(passStyle.Render("yes"), 11))
			} else {
				sb.WriteString(col(dimStyle.Render("—"), 11))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// RenderCredentials renders the credential status table.
func RenderCredentials(machines []model.MachineState) string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Credentials"))
	sb.WriteString("\n")
	sb.WriteString(" " + col("", 16) + " " + col("SSH Agent", 14) + " " + col("setup-creds", 14) + " " + col("Atuin Sync", 14) + "\n")

	for _, m := range machines {
		sshStr := styleValue(m.SSHAgent, "1password", "system")
		setupStr := styleValue(m.SetupCreds, "ran", "")
		atuinStr := styleValue(m.AtuinSync, "synced", "")

		sb.WriteString(" " + col(m.Hostname, 16) + " " + col(sshStr, 14) + " " + col(setupStr, 14) + " " + col(atuinStr, 14) + "\n")
	}
	return sb.String()
}

func styleValue(val, good, warn string) string {
	switch val {
	case good:
		return passStyle.Render(val)
	case warn:
		if warn != "" {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(val)
		}
		return dimStyle.Render(val)
	default:
		return dimStyle.Render(val)
	}
}
