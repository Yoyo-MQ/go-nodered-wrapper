package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
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

// DeployFlow deploys or updates a flow in Node-RED
func (c *NodeRedClient) DeployFlow(ctx context.Context, flow *types.FlowDefinition) error {
	// Node-RED expects a flat array of nodes, not a FlowDefinition object
	// Convert FlowDefinition to Node-RED format
	nodeRedNodes := c.convertFlowToNodeRedFormat(flow)

	// For /flow endpoint, send the tab node and its children as nodes array
	payload := map[string]interface{}{
		"id":    flow.ID,
		"label": flow.Name,
		"nodes": nodeRedNodes[1:], // Skip the tab node itself, just send the child nodes
	}
	if flow.Description != "" {
		payload["info"] = flow.Description
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal flow: %w", err)
	}

	// Try to update first (PUT /flow/:id) - this works for existing flows
	url := fmt.Sprintf("%s/flow/%s", c.baseURL, flow.ID)
	method := "PUT"

	if c.debug {
		fmt.Printf("Deploying flow to %s using %s: %s\n", url, method, string(jsonData))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
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

	// If flow doesn't exist (404), try creating it with POST
	if resp.StatusCode == http.StatusNotFound {
		if c.debug {
			fmt.Println("Flow not found, creating new flow with POST")
		}
		return c.createFlow(ctx, flow, payload, jsonData)
	}

	// /flow endpoint returns 200 for success
	if resp.StatusCode != http.StatusOK {
		// Try to read error body
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to deploy flow: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// createFlow creates a new flow using POST /flow
func (c *NodeRedClient) createFlow(ctx context.Context, flow *types.FlowDefinition, payload map[string]interface{}, jsonData []byte) error {
	url := fmt.Sprintf("%s/flow", c.baseURL)

	if c.debug {
		fmt.Printf("Creating new flow at %s: %s\n", url, string(jsonData))
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
		return fmt.Errorf("failed to create flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create flow: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// convertFlowToNodeRedFormat converts a FlowDefinition to Node-RED's expected format
func (c *NodeRedClient) convertFlowToNodeRedFormat(flow *types.FlowDefinition) []map[string]interface{} {
	var nodeRedNodes []map[string]interface{}

	// Add a tab (flow container) node
	tabNode := map[string]interface{}{
		"id":    flow.ID,
		"type":  "tab",
		"label": flow.Name,
	}
	if flow.Description != "" {
		tabNode["info"] = flow.Description
	}
	nodeRedNodes = append(nodeRedNodes, tabNode)

	// Convert each node
	for _, node := range flow.Nodes {
		nodeRedNode := map[string]interface{}{
			"id":    node.ID,
			"type":  node.Type,
			"name":  node.Name,
			"x":     node.Position.X,
			"y":     node.Position.Y,
			"z":     flow.ID, // Link node to the tab/flow
			"wires": node.Wires,
		}

		// Add all properties from the node
		for key, value := range node.Properties {
			nodeRedNode[key] = value
		}

		nodeRedNodes = append(nodeRedNodes, nodeRedNode)
	}

	return nodeRedNodes
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

// GetAuthToken authenticates with Node-RED and returns an access token
func (c *NodeRedClient) GetAuthToken(ctx context.Context, username, password string) (string, error) {
	url := fmt.Sprintf("%s/auth/token", c.baseURL)

	// Prepare authentication request payload
	authPayload := map[string]interface{}{
		"client_id":  "node-red-admin",
		"grant_type": "password",
		"scope":      "*",
		"username":   username,
		"password":   password,
	}

	jsonData, err := json.Marshal(authPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response to extract the access token
	var authResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResponse.AccessToken == "" {
		return "", fmt.Errorf("no access token in auth response")
	}

	// Update the client's API key
	c.apiKey = authResponse.AccessToken

	return authResponse.AccessToken, nil
}
