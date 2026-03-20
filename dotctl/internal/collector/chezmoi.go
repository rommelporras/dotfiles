package collector

import (
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/rommelporras/dotfiles/dotctl/internal/model"
)

// CommandRunner abstracts exec.Command for testing.
type CommandRunner interface {
	Run(name string, args ...string) (string, error)
}

// ExecRunner shells out to real commands.
type ExecRunner struct{}

func (r *ExecRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	return string(out), err
}

func parseChezmoiStatus(output string) []model.DriftFile {
	var files []model.DriftFile
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 3 {
			continue
		}
		status := string(line[0])
		path := strings.TrimSpace(line[1:])
		if path == "" {
			continue
		}
		files = append(files, model.DriftFile{Path: path, Status: status})
	}
	return files
}

func parseChezmoiData(jsonOutput string) (map[string]any, error) {
	var full map[string]any
	if err := json.Unmarshal([]byte(jsonOutput), &full); err != nil {
		return nil, err
	}
	return full, nil
}

// CollectChezmoi gathers drift and template data from chezmoi CLI.
func CollectChezmoi(runner CommandRunner) ([]model.DriftFile, map[string]any, error) {
	statusOut, err := runner.Run("chezmoi", "status")
	if err != nil {
		return nil, nil, err
	}
	drift := parseChezmoiStatus(statusOut)

	dataOut, err := runner.Run("chezmoi", "data", "--format", "json")
	if err != nil {
		return drift, nil, err
	}
	data, err := parseChezmoiData(dataOut)
	if err != nil {
		return drift, nil, err
	}

	return drift, data, nil
}
