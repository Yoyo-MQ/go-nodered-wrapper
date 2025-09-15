package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourusername/yoyo-nodered-wrapper/pkg/types"
)

// NodeRedClient handles communication with Node-RED
type NodeRedClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	debug      bool
}

// NewNodeRedClient creates a new Node-RED client
func NewNodeRedClient(config *types.Config) (*NodeRedClient, error) {
	if config.NodeRedURL == "" {
		return nil, fmt.Errorf("node_red_url is required")
	}

	return &NodeRedClient{
		baseURL: config.NodeRedURL,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey: config.APIKey,
		debug:  config.Debug,
	}, nil
}

// DeployFlow deploys a flow to Node-RED
func (c *NodeRedClient) DeployFlow(ctx context.Context, flow *types.FlowDefinition) error {
	url := fmt.Sprintf("%s/flows", c.baseURL)

	payload := map[string]interface{}{
		"flows": []*types.FlowDefinition{flow},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal flow: %w", err)
	}

	if c.debug {
		fmt.Printf("Deploying flow to %s: %s\n", url, string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to deploy flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deploy flow: status %d", resp.StatusCode)
	}

	return nil
}

// ExecuteFlow triggers a flow execution
func (c *NodeRedClient) ExecuteFlow(ctx context.Context, flowID string, input map[string]interface{}) (*types.ExecutionResult, error) {
	url := fmt.Sprintf("%s/flows/%s/execute", c.baseURL, flowID)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	if c.debug {
		fmt.Printf("Executing flow %s: %s\n", flowID, string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute flow: %w", err)
	}
	defer resp.Body.Close()

	var result types.ExecutionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetFlow retrieves a deployed flow
func (c *NodeRedClient) GetFlow(ctx context.Context, flowID string) (*types.FlowDefinition, error) {
	url := fmt.Sprintf("%s/flows/%s", c.baseURL, flowID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get flow: status %d", resp.StatusCode)
	}

	var flow types.FlowDefinition
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return nil, fmt.Errorf("failed to decode flow: %w", err)
	}

	return &flow, nil
}

// DeleteFlow removes a flow from Node-RED
func (c *NodeRedClient) DeleteFlow(ctx context.Context, flowID string) error {
	url := fmt.Sprintf("%s/flows/%s", c.baseURL, flowID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("flow not found: %s", flowID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete flow: status %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck checks if Node-RED is healthy
func (c *NodeRedClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Node-RED is not healthy: status %d", resp.StatusCode)
	}

	return nil
}
