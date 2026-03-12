package collector

import "os/exec"

// TrackedTools is the list of tools dotctl checks for.
var TrackedTools = []string{
	"glab", "kubectl", "terraform", "aws", "ansible",
	"op", "atuin", "bun",
}

// PathLookup abstracts exec.LookPath for testing.
type PathLookup interface {
	LookPath(name string) (string, error)
}

type realLookup struct{}

func (r *realLookup) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// ProbeTools checks which tracked tools are installed.
func ProbeTools() map[string]string {
	return probeToolsWith(&realLookup{})
}

func probeToolsWith(lookup PathLookup) map[string]string {
	result := make(map[string]string, len(TrackedTools))
	for _, tool := range TrackedTools {
		path, err := lookup.LookPath(tool)
		if err != nil {
			result[tool] = ""
		} else {
			result[tool] = path
		}
	}
	return result
}
