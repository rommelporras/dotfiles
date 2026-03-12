package query

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryMachines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{
						"metric": {"hostname": "aurora-dx", "platform": "aurora", "context": "personal"},
						"value": [1710000000, "1"]
					},
					{
						"metric": {"hostname": "work-eam", "platform": "distrobox", "context": "work-eam"},
						"value": [1710000000, "1"]
					}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewPrometheusClient(server.URL)
	machines, err := client.QueryMachines()
	if err != nil {
		t.Fatalf("QueryMachines: %v", err)
	}
	if len(machines) != 2 {
		t.Fatalf("len = %d, want 2", len(machines))
	}
	if machines[0].Hostname != "aurora-dx" {
		t.Errorf("machines[0].Hostname = %q", machines[0].Hostname)
	}
	if machines[1].Platform != "distrobox" {
		t.Errorf("machines[1].Platform = %q", machines[1].Platform)
	}
}

func TestQueryDriftTotals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"status": "success",
			"data": {
				"resultType": "vector",
				"result": [
					{"metric": {"hostname": "aurora-dx"}, "value": [1710000000, "3"]},
					{"metric": {"hostname": "work-eam"}, "value": [1710000000, "0"]}
				]
			}
		}`))
	}))
	defer server.Close()

	client := NewPrometheusClient(server.URL)
	drift, err := client.QueryDriftTotals()
	if err != nil {
		t.Fatalf("QueryDriftTotals: %v", err)
	}
	if drift["aurora-dx"] != 3 {
		t.Errorf("aurora-dx drift = %d, want 3", drift["aurora-dx"])
	}
	if drift["work-eam"] != 0 {
		t.Errorf("work-eam drift = %d, want 0", drift["work-eam"])
	}
}

func TestQueryPrometheusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "error", "error": "query error"}`))
	}))
	defer server.Close()

	client := NewPrometheusClient(server.URL)
	_, err := client.QueryMachines()
	if err == nil {
		t.Error("expected error for status=error response")
	}
}
