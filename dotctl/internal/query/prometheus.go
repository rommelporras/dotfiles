package query

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PrometheusClient queries the Prometheus HTTP API.
type PrometheusClient struct {
	baseURL    string
	httpClient *http.Client
}

// MachineInfo is a summary from Prometheus labels.
type MachineInfo struct {
	Hostname string
	Platform string
	Context  string
}

type promResponse struct {
	Status string   `json:"status"`
	Error  string   `json:"error"`
	Data   promData `json:"data"`
}

type promData struct {
	ResultType string       `json:"resultType"`
	Result     []promResult `json:"result"`
}

type promResult struct {
	Metric map[string]string `json:"metric"`
	Value  [2]any            `json:"value"`
}

// NewPrometheusClient creates a client for the given Prometheus URL.
func NewPrometheusClient(baseURL string) *PrometheusClient {
	return &PrometheusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *PrometheusClient) query(promql string) (*promResponse, error) {
	u := c.baseURL + "/api/v1/query?query=" + url.QueryEscape(promql)
	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("prometheus query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result promResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Status != "success" {
		return nil, fmt.Errorf("prometheus returned status %q: %s", result.Status, result.Error)
	}
	return &result, nil
}

// QueryMachines returns all machines that have reported in.
func (c *PrometheusClient) QueryMachines() ([]MachineInfo, error) {
	resp, err := c.query("dotctl_up")
	if err != nil {
		return nil, err
	}

	var machines []MachineInfo
	for _, r := range resp.Data.Result {
		machines = append(machines, MachineInfo{
			Hostname: r.Metric["hostname"],
			Platform: r.Metric["platform"],
			Context:  r.Metric["context"],
		})
	}
	return machines, nil
}

// QueryDriftTotals returns drift count per hostname.
func (c *PrometheusClient) QueryDriftTotals() (map[string]int, error) {
	resp, err := c.query("dotctl_drift_total")
	if err != nil {
		return nil, err
	}

	drift := make(map[string]int)
	for _, r := range resp.Data.Result {
		hostname := r.Metric["hostname"]
		if valStr, ok := r.Value[1].(string); ok {
			val, _ := strconv.Atoi(valStr)
			drift[hostname] = val
		}
	}
	return drift, nil
}

// QueryToolsInstalled returns tool installation status per hostname.
func (c *PrometheusClient) QueryToolsInstalled() (map[string]map[string]bool, error) {
	resp, err := c.query("dotctl_tool_installed")
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]bool)
	for _, r := range resp.Data.Result {
		hostname := r.Metric["hostname"]
		tool := r.Metric["tool"]
		if result[hostname] == nil {
			result[hostname] = make(map[string]bool)
		}
		if valStr, ok := r.Value[1].(string); ok {
			result[hostname][tool] = valStr == "1"
		}
	}
	return result, nil
}

// QueryCredentials returns credential status per hostname.
func (c *PrometheusClient) QueryCredentials() (map[string]map[string]bool, error) {
	resp, err := c.query("dotctl_credential_status")
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]bool)
	for _, r := range resp.Data.Result {
		hostname := r.Metric["hostname"]
		cred := r.Metric["credential"]
		if result[hostname] == nil {
			result[hostname] = make(map[string]bool)
		}
		if valStr, ok := r.Value[1].(string); ok {
			result[hostname][cred] = valStr == "1"
		}
	}
	return result, nil
}
