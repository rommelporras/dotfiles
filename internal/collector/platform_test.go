package collector

import "testing"

func TestDetectPlatformFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		procVer string
		osRel   string
		want    string
	}{
		{
			name: "distrobox detected from env",
			env:  map[string]string{"DISTROBOX_ENTER_PATH": "/usr/bin/distrobox-enter"},
			want: "distrobox",
		},
		{
			name:    "wsl detected from /proc/version",
			procVer: "Linux version 5.15.0 (microsoft@microsoft.com)",
			want:    "wsl",
		},
		{
			name:  "aurora detected from os-release",
			osRel: "NAME=\"Aurora\"\nID=aurora\n",
			want:  "aurora",
		},
		{
			name: "unknown when nothing matches",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPlatformWith(tt.env, tt.procVer, tt.osRel)
			if got != tt.want {
				t.Errorf("detectPlatformWith() = %q, want %q", got, tt.want)
			}
		})
	}
}
