package model

// DriftFile represents a single chezmoi-managed file with drift.
type DriftFile struct {
	Path   string `json:"path"`
	Status string `json:"status"` // M=modified, A=added, D=deleted, R=re-run needed
}

// MachineState holds the collected state of a single machine/container.
type MachineState struct {
	Hostname     string            `json:"hostname"`
	Platform     string            `json:"platform"`     // aurora, distrobox, wsl
	Context      string            `json:"context"`      // personal, work-eam, etc.
	DriftFiles   []DriftFile       `json:"drift_files"`
	TemplateData map[string]any    `json:"template_data"`
	Tools        map[string]string `json:"tools"`        // tool name -> path or ""
	SSHAgent     string            `json:"ssh_agent"`    // 1password, manual, none, n/a
	SetupCreds   string            `json:"setup_creds"`  // ran, skipped, n/a
	AtuinSync    string            `json:"atuin_sync"`   // synced, disabled, n/a
	ClaudeLinks  map[string]string `json:"claude_links"` // item -> ok, wrong, file, missing, n/a
}

// ContainerInfo describes a distrobox container.
type ContainerInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"` // running, stopped, created
	Image  string `json:"image"`
}

// CollectResult is the full output of a collection run.
type CollectResult struct {
	Machines   []MachineState  `json:"machines"`
	Containers []ContainerInfo `json:"containers"`
}
