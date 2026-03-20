package query

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rommelporras/dotfiles/dotctl/internal/model"
)

// LokiClient queries the Loki HTTP API.
type LokiClient struct {
	baseURL    string
	httpClient *http.Client
}

type lokiResponse struct {
	Status string   `json:"status"`
	Data   lokiData `json:"data"`
}

type lokiData struct {
	ResultType string       `json:"resultType"`
	Result     []lokiStream `json:"result"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

// NewLokiClient creates a client for the given Loki URL.
func NewLokiClient(baseURL string) *LokiClient {
	return &LokiClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// QueryLatestStates returns the most recent state log per machine.
func (c *LokiClient) QueryLatestStates() ([]model.MachineState, error) {
	logql := `{service_name="dotctl"}`
	since := time.Now().Add(-30 * time.Minute)

	u := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%d&end=%d&limit=100&direction=backward",
		c.baseURL,
		url.QueryEscape(logql),
		since.UnixNano(),
		time.Now().UnixNano(),
	)

	resp, err := c.httpClient.Get(u)
	if err != nil {
		return nil, fmt.Errorf("loki query: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result lokiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var states []model.MachineState

	for _, stream := range result.Data.Result {
		for _, entry := range stream.Values {
			var state model.MachineState
			if err := json.Unmarshal([]byte(entry[1]), &state); err != nil {
				continue
			}
			if seen[state.Hostname] {
				continue
			}
			seen[state.Hostname] = true
			states = append(states, state)
		}
	}

	return states, nil
}
