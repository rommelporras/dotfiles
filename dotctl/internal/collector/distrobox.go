package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rommelporras/dotfiles/dotctl/internal/model"
)

const containerTimeout = 30 * time.Second

func parseDistroboxList(output string) []model.ContainerInfo {
	var containers []model.ContainerInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] { // skip header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		rawStatus := strings.TrimSpace(parts[2])
		image := strings.TrimSpace(parts[3])

		status := "stopped"
		if strings.HasPrefix(rawStatus, "Up") {
			status = "running"
		}

		containers = append(containers, model.ContainerInfo{
			Name:   name,
			Status: status,
			Image:  image,
		})
	}
	return containers
}

// ListContainers returns all distrobox containers.
func ListContainers(runner CommandRunner) ([]model.ContainerInfo, error) {
	out, err := runner.Run("distrobox", "list")
	if err != nil {
		return nil, fmt.Errorf("distrobox list: %w", err)
	}
	return parseDistroboxList(out), nil
}

// RunInContainer executes a shell command inside a container with timeout.
func RunInContainer(name, shellCmd string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), containerTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "distrobox", "enter", name, "--", "sh", "-c", shellCmd)
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return "", fmt.Errorf("container %s: command timed out after %s", name, containerTimeout)
	}
	return string(out), err
}
