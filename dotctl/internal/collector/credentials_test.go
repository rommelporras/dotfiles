package collector

import "testing"

func TestDetectSSHAgentType(t *testing.T) {
	tests := []struct {
		sock string
		want string
	}{
		{"/home/user/.1password/agent.sock", "1password"},
		{"/tmp/ssh-XXX/agent.123", "system"},
		{"", "none"},
	}

	for _, tt := range tests {
		got := detectSSHAgentType(tt.sock)
		if got != tt.want {
			t.Errorf("detectSSHAgentType(%q) = %q, want %q", tt.sock, got, tt.want)
		}
	}
}
