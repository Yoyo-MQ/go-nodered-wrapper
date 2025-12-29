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

type NodeRedClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	debug      bool
}

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

func (c *NodeRedClient) DeployFlow(ctx context.Context, flow *types.FlowDefinition) (string, error) {
	nodeRedNodes := c.convertFlowToNodeRedFormat(flow)
	payload := map[string]interface{}{
		"id":    flow.ID,
		"label": flow.Name,
		"nodes": nodeRedNodes[1:],
	}
	if flow.Description != "" {
		payload["info"] = flow.Description
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal flow: %w", err)
	}

	url := fmt.Sprintf("%s/flow/%s", c.baseURL, flow.ID)
	resp, err := c.doRequest(ctx, "PUT", url, jsonData)
	if err != nil {
		return "", err
	}
	defer c.closeResponseBody(resp)

	switch resp.StatusCode {
	case http.StatusNotFound:
		if c.debug {
			fmt.Println("Flow not found, creating new flow with POST")
		}
		return c.createFlow(ctx, jsonData)
	case http.StatusBadRequest:
		if c.debug {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Flow deployment returned 400, attempting to delete and recreate: %s\n", string(body))
		}
		_ = c.DeleteFlow(ctx, flow.ID)
		return c.createFlow(ctx, jsonData)
	case http.StatusOK:
		return flow.ID, nil
	default:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to deploy flow: status %d, failed to read error body: %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("failed to deploy flow: status %d, body: %s", resp.StatusCode, string(body))
	}
}

func (c *NodeRedClient) createFlow(ctx context.Context, jsonData []byte) (string, error) {
	return c.createFlowWithRetry(ctx, jsonData, false)
}

func (c *NodeRedClient) createFlowWithRetry(ctx context.Context, jsonData []byte, isRetry bool) (string, error) {
	url := fmt.Sprintf("%s/flow", c.baseURL)
	if c.debug {
		fmt.Printf("Creating new flow at %s: %s\n", url, string(jsonData))
	}

	resp, err := c.doRequest(ctx, "POST", url, jsonData)
	if err != nil {
		return "", err
	}
	defer c.closeResponseBody(resp)

	if resp.StatusCode == http.StatusBadRequest && !isRetry {
		if c.debug {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Flow creation returned 400, attempting to delete and recreate: %s\n", string(body))
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(jsonData, &payload); err == nil {
			if flowID, ok := payload["id"].(string); ok {
				_ = c.DeleteFlow(ctx, flowID)
				return c.createFlowWithRetry(ctx, jsonData, true)
			}
		}
		return "", fmt.Errorf("failed to create flow: status %d", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		if len(body) > 0 {
			var response struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(body, &response); err == nil && response.ID != "" {
				if c.debug {
					fmt.Printf("Flow created with ID: %s\n", response.ID)
				}
				return response.ID, nil
			}
		}
		return "", nil
	}

	if resp.StatusCode == http.StatusNoContent {
		return "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to create flow: status %d, failed to read error body: %w", resp.StatusCode, err)
	}
	return "", fmt.Errorf("failed to create flow: status %d, body: %s", resp.StatusCode, string(body))
}

func (c *NodeRedClient) ExecuteFlow(ctx context.Context, flowID string, input map[string]interface{}) (*types.ExecutionResult, error) {
	flow, err := c.GetFlow(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	for _, node := range flow.Nodes {
		if node.Type == "inject" {
			if err := c.TriggerNode(ctx, node.ID, input); err != nil {
				return nil, fmt.Errorf("failed to trigger inject node: %w", err)
			}
			return &types.ExecutionResult{
				Success: true,
				Output:  input,
			}, nil
		}
	}

	return nil, fmt.Errorf("no inject node found in flow %s", flowID)
}

func (c *NodeRedClient) TriggerNode(ctx context.Context, nodeID string, input map[string]interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	url := fmt.Sprintf("%s/inject/%s", c.baseURL, nodeID)
	if c.debug {
		fmt.Printf("Triggering node %s: %s\n", nodeID, string(jsonData))
	}

	resp, err := c.doRequest(ctx, "POST", url, jsonData)
	if err != nil {
		return err
	}
	defer c.closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to trigger node: status %d, failed to read error body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("failed to trigger node: status %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *NodeRedClient) GetFlow(ctx context.Context, flowID string) (*types.FlowDefinition, error) {
	flows, err := c.GetFlows(ctx)
	if err != nil {
		return nil, err
	}

	for _, flowNode := range flows {
		nodeID, _ := flowNode["id"].(string)
		nodeType, _ := flowNode["type"].(string)
		if nodeID == flowID && nodeType == "tab" {
			nodes := c.collectNodesForFlow(flows, flowID)
			return &types.FlowDefinition{
				ID:          flowID,
				Name:        getString(flowNode, "label"),
				Description: getString(flowNode, "info"),
				Nodes:       nodes,
			}, nil
		}
	}

	return nil, fmt.Errorf("flow not found: %s", flowID)
}

func (c *NodeRedClient) GetFlows(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("%s/flows", c.baseURL), nil)
	if err != nil {
		return nil, err
	}
	defer c.closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get flows: status %d", resp.StatusCode)
	}

	var flows []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flows); err != nil {
		return nil, fmt.Errorf("failed to decode flows: %w", err)
	}

	return flows, nil
}

func (c *NodeRedClient) DeleteFlow(ctx context.Context, flowID string) error {
	url := fmt.Sprintf("%s/flows/%s", c.baseURL, flowID)
	resp, err := c.doRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	defer c.closeResponseBody(resp)

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("flow not found: %s", flowID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete flow: status %d", resp.StatusCode)
	}

	return nil
}

func (c *NodeRedClient) HealthCheck(ctx context.Context) error {
	baseURL := c.baseURL
	if len(baseURL) > 6 && baseURL[len(baseURL)-6:] == "/admin" {
		baseURL = baseURL[:len(baseURL)-6]
	}

	resp, err := c.doRequest(ctx, "GET", baseURL+"/", nil)
	if err != nil {
		return err
	}
	defer c.closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("Node-RED is not healthy: status %d", resp.StatusCode)
	}

	return nil
}

func (c *NodeRedClient) GetAuthToken(ctx context.Context, username, password string) (string, error) {
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

	url := fmt.Sprintf("%s/auth/token", c.baseURL)
	resp, err := c.doRequest(ctx, "POST", url, jsonData)
	if err != nil {
		return "", err
	}
	defer c.closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("authentication failed: status %d, failed to read error body: %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("authentication failed: status %d, body: %s", resp.StatusCode, string(body))
	}

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

	c.apiKey = authResponse.AccessToken
	return authResponse.AccessToken, nil
}

func (c *NodeRedClient) convertFlowToNodeRedFormat(flow *types.FlowDefinition) []map[string]interface{} {
	var nodeRedNodes []map[string]interface{}

	tabNode := map[string]interface{}{
		"id":    flow.ID,
		"type":  "tab",
		"label": flow.Name,
	}
	if flow.Description != "" {
		tabNode["info"] = flow.Description
	}
	nodeRedNodes = append(nodeRedNodes, tabNode)

	for _, node := range flow.Nodes {
		nodeRedNode := map[string]interface{}{
			"id":    node.ID,
			"type":  node.Type,
			"name":  node.Name,
			"x":     node.Position.X,
			"y":     node.Position.Y,
			"z":     flow.ID,
			"wires": node.Wires,
		}

		for key, value := range node.Properties {
			nodeRedNode[key] = value
		}

		nodeRedNodes = append(nodeRedNodes, nodeRedNode)
	}

	return nodeRedNodes
}

func (c *NodeRedClient) collectNodesForFlow(flows []map[string]interface{}, flowID string) []types.Node {
	var nodes []types.Node

	for _, node := range flows {
		z, _ := node["z"].(string)
		if z != flowID {
			continue
		}

		nodeObj := types.Node{
			ID:   getString(node, "id"),
			Type: getString(node, "type"),
			Name: getString(node, "name"),
			Position: types.Position{
				X: getFloat64(node, "x"),
				Y: getFloat64(node, "y"),
			},
			Properties: make(map[string]interface{}),
		}

		for k, v := range node {
			if k != "id" && k != "type" && k != "name" && k != "x" && k != "y" && k != "z" && k != "wires" {
				nodeObj.Properties[k] = v
			}
		}

		if wires, ok := node["wires"].([]interface{}); ok {
			nodeObj.Wires = make([][]string, len(wires))
			for i, wire := range wires {
				if wireArr, ok := wire.([]interface{}); ok {
					nodeObj.Wires[i] = make([]string, len(wireArr))
					for j, w := range wireArr {
						if wStr, ok := w.(string); ok {
							nodeObj.Wires[i][j] = wStr
						}
					}
				}
			}
		}

		nodes = append(nodes, nodeObj)
	}

	return nodes
}

func (c *NodeRedClient) doRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

func (c *NodeRedClient) closeResponseBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil && c.debug {
		fmt.Printf("Warning: failed to close response body: %v\n", err)
	}
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	return 0
}
