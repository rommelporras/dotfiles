package collector

import (
	"strings"

	"github.com/rommelporras/dotfiles/internal/model"
)

// CollectLocal gathers state from the local machine.
func CollectLocal(runner CommandRunner, hostname, platform string) (*model.MachineState, error) {
	drift, tmplData, chezmoiErr := CollectChezmoi(runner)

	context := ""
	if tmplData != nil {
		if c, ok := tmplData["context"].(string); ok {
			context = c
		}
	}

	state := &model.MachineState{
		Hostname:     hostname,
		Platform:     platform,
		Context:      context,
		DriftFiles:   drift,
		TemplateData: tmplData,
		Tools:        ProbeTools(),
		SSHAgent:     DetectSSHAgent(),
		SetupCreds:   DetectSetupCreds(),
		AtuinSync:    DetectAtuinSync(),
	}

	if chezmoiErr != nil && drift == nil {
		state.DriftFiles = nil
	}

	return state, nil
}

// CollectAll gathers state from local machine + running distrobox containers.
func CollectAll(runner CommandRunner, hostname, platform string) (*model.CollectResult, error) {
	local, err := CollectLocal(runner, hostname, platform)
	if err != nil {
		return nil, err
	}

	result := &model.CollectResult{
		Machines: []model.MachineState{*local},
	}

	if platform == "aurora" {
		containers, err := ListContainers(runner)
		if err == nil {
			result.Containers = containers
			for _, c := range containers {
				if c.Status != "running" {
					continue
				}
				containerState, err := collectFromContainer(c.Name)
				if err != nil {
					continue
				}
				result.Machines = append(result.Machines, *containerState)
			}
		}
	}

	return result, nil
}

func collectFromContainer(name string) (*model.MachineState, error) {
	statusOut, _ := RunInContainer(name, "chezmoi status 2>/dev/null || $HOME/bin/chezmoi status 2>/dev/null || true")
	dataOut, _ := RunInContainer(name, "chezmoi data --format json 2>/dev/null || $HOME/bin/chezmoi data --format json 2>/dev/null || echo '{}'")

	drift := parseChezmoiStatus(statusOut)
	tmplData, _ := parseChezmoiData(dataOut)

	context := ""
	if tmplData != nil {
		if c, ok := tmplData["context"].(string); ok {
			context = c
		}
	}

	// Probe tools inside container via single command
	toolsScript := ""
	for _, tool := range TrackedTools {
		toolsScript += "printf '%s\\n' \"$(command -v " + tool + " 2>/dev/null)\"; "
	}
	toolsOut, _ := RunInContainer(name, toolsScript)
	tools := make(map[string]string, len(TrackedTools))
	lines := strings.Split(strings.TrimSpace(toolsOut), "\n")
	for i, tool := range TrackedTools {
		if i < len(lines) {
			tools[tool] = strings.TrimSpace(lines[i])
		} else {
			tools[tool] = ""
		}
	}

	sshOut, _ := RunInContainer(name, "printf '%s' \"$SSH_AUTH_SOCK\"")
	sshAgent := detectSSHAgentType(strings.TrimSpace(sshOut))

	setupOut, _ := RunInContainer(name, "test -x $HOME/.local/bin/setup-creds && printf ran || printf n/a")
	setupCreds := strings.TrimSpace(setupOut)

	atuinOut, _ := RunInContainer(name, "grep -q sync_address $HOME/.config/atuin/config.toml 2>/dev/null && printf synced || printf n/a")
	atuinSync := strings.TrimSpace(atuinOut)

	return &model.MachineState{
		Hostname:     name,
		Platform:     "distrobox",
		Context:      context,
		DriftFiles:   drift,
		TemplateData: tmplData,
		Tools:        tools,
		SSHAgent:     sshAgent,
		SetupCreds:   setupCreds,
		AtuinSync:    atuinSync,
	}, nil
}
