package collector

import (
	"testing"
)

func TestParseChezmoiStatus(t *testing.T) {
	output := ` M .zshrc
 A .config/new-file.toml
 R bootstrap.sh
 M .local/bin/setup-creds
`
	got := parseChezmoiStatus(output)

	if len(got) != 4 {
		t.Fatalf("len = %d, want 4", len(got))
	}

	want := []struct {
		path   string
		status string
	}{
		{".zshrc", "M"},
		{".config/new-file.toml", "A"},
		{"bootstrap.sh", "R"},
		{".local/bin/setup-creds", "M"},
	}

	for i, w := range want {
		if got[i].Path != w.path || got[i].Status != w.status {
			t.Errorf("[%d] got {%q, %q}, want {%q, %q}", i, got[i].Path, got[i].Status, w.path, w.status)
		}
	}
}

func TestParseChezmoiStatusEmpty(t *testing.T) {
	got := parseChezmoiStatus("")
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestParseChezmoiData(t *testing.T) {
	jsonOutput := `{
		"context": "personal",
		"platform": "aurora",
		"atuin_account": "personal",
		"has_homelab_creds": true,
		"has_work_creds": false
	}`

	data, err := parseChezmoiData(jsonOutput)
	if err != nil {
		t.Fatalf("parseChezmoiData: %v", err)
	}
	if data["context"] != "personal" {
		t.Errorf("context = %v, want personal", data["context"])
	}
	if data["has_homelab_creds"] != true {
		t.Errorf("has_homelab_creds = %v, want true", data["has_homelab_creds"])
	}
}
