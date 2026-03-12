package collector

import (
	"testing"
)

func TestParseDistroboxList(t *testing.T) {
	output := `ID           | NAME                 | STATUS             | IMAGE
abc123       | work-eam             | Up 2 hours         | ubuntu:24.04
def456       | personal             | Up 2 hours         | ubuntu:24.04
ghi789       | sandbox              | Created            | ubuntu:24.04
`
	got := parseDistroboxList(output)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}

	want := []struct {
		name   string
		status string
	}{
		{"work-eam", "running"},
		{"personal", "running"},
		{"sandbox", "stopped"},
	}

	for i, w := range want {
		if got[i].Name != w.name {
			t.Errorf("[%d] Name = %q, want %q", i, got[i].Name, w.name)
		}
		if got[i].Status != w.status {
			t.Errorf("[%d] Status = %q, want %q", i, got[i].Status, w.status)
		}
	}
}

func TestParseDistroboxListEmpty(t *testing.T) {
	got := parseDistroboxList("ID           | NAME                 | STATUS             | IMAGE\n")
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}
