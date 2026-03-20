package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryLatestStates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "streams",
				"result": [
					{
						"stream": {"service_name": "dotctl", "hostname": "aurora-dx"},
						"values": [
							["1710000000000000000", "{\"hostname\":\"aurora-dx\",\"platform\":\"aurora\",\"context\":\"personal\",\"drift_files\":[{\"path\":\".zshrc\",\"status\":\"M\"}],\"tools\":{\"glab\":\"/usr/bin/glab\"},\"ssh_agent\":\"1password\",\"setup_creds\":\"n/a\",\"atuin_sync\":\"synced\"}"]
						]
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLokiClient(server.URL)
	states, err := client.QueryLatestStates()
	if err != nil {
		t.Fatalf("QueryLatestStates: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("len = %d, want 1", len(states))
	}
	if states[0].Hostname != "aurora-dx" {
		t.Errorf("Hostname = %q", states[0].Hostname)
	}
	if len(states[0].DriftFiles) != 1 {
		t.Errorf("DriftFiles len = %d, want 1", len(states[0].DriftFiles))
	}
	if states[0].SSHAgent != "1password" {
		t.Errorf("SSHAgent = %q", states[0].SSHAgent)
	}
}

func TestQueryLatestStatesDeduplication(t *testing.T) {
	// Two entries for the same hostname — should deduplicate to 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "streams",
				"result": [
					{
						"stream": {"service_name": "dotctl"},
						"values": [
							["1710000000000000001", "{\"hostname\":\"aurora-dx\",\"platform\":\"aurora\",\"context\":\"personal\"}"],
							["1710000000000000000", "{\"hostname\":\"aurora-dx\",\"platform\":\"aurora\",\"context\":\"personal\"}"]
						]
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewLokiClient(server.URL)
	states, err := client.QueryLatestStates()
	if err != nil {
		t.Fatalf("QueryLatestStates: %v", err)
	}
	if len(states) != 1 {
		t.Errorf("expected 1 deduplicated state, got %d", len(states))
	}
}
