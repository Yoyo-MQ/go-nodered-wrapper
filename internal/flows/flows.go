package flows

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/yoyo-mq/go-nodered-wrapper/internal/converters"
	"github.com/yoyo-mq/go-nodered-wrapper/internal/httpclient"
	"github.com/yoyo-mq/go-nodered-wrapper/internal/util"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// Service handles flow operations
type Service struct {
	httpClient *httpclient.Client
	logger     *slog.Logger
}

// NewService creates a new flows service
func NewService(httpClient *httpclient.Client, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		httpClient: httpClient,
		logger:     logger,
	}
}

// DeployFlow deploys a flow to Node-RED
func (s *Service) DeployFlow(ctx context.Context, flow *types.FlowDefinition) (string, error) {
	nodeRedNodes := converters.ConvertFlowToNodeRedFormat(flow)
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

	url := fmt.Sprintf("%s/flow/%s", s.httpClient.BaseURL(), flow.ID)
	resp, err := s.httpClient.DoRequest(ctx, types.MethodPut, url, jsonData)
	if err != nil {
		return "", err
	}
	defer s.httpClient.CloseResponseBody(resp)

	switch resp.StatusCode {
	case http.StatusNotFound:
		s.logger.Debug("Flow not found, creating new flow with POST")
		return s.createFlow(ctx, jsonData)
	case http.StatusBadRequest:
		s.logger.Debug("Flow deployment returned 400, attempting to delete and recreate")
		_ = s.DeleteFlow(ctx, flow.ID)
		return s.createFlow(ctx, jsonData)
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

func (s *Service) createFlow(ctx context.Context, jsonData []byte) (string, error) {
	return s.createFlowWithRetry(ctx, jsonData, false)
}

func (s *Service) createFlowWithRetry(ctx context.Context, jsonData []byte, isRetry bool) (string, error) {
	url := fmt.Sprintf("%s/flow", s.httpClient.BaseURL())
	if s.httpClient.Debug() {
		fmt.Printf("Creating new flow at %s: %s\n", url, string(jsonData))
	}

	resp, err := s.httpClient.DoRequest(ctx, types.MethodPost, url, jsonData)
	if err != nil {
		return "", err
	}
	defer s.httpClient.CloseResponseBody(resp)

	if resp.StatusCode == http.StatusBadRequest && !isRetry {
		if s.httpClient.Debug() {
			s.logger.Debug("Flow creation returned 400, attempting to delete and recreate")
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(jsonData, &payload); err == nil {
			if flowID, ok := payload["id"].(string); ok {
				_ = s.DeleteFlow(ctx, flowID)
				return s.createFlowWithRetry(ctx, jsonData, true)
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
				s.logger.Debug("Flow created", "id", response.ID)
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

// GetFlow retrieves a flow by ID
func (s *Service) GetFlow(ctx context.Context, flowID string) (*types.FlowDefinition, error) {
	url := fmt.Sprintf("%s/flow/%s", s.httpClient.BaseURL(), flowID)
	resp, err := s.httpClient.DoRequest(ctx, types.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer s.httpClient.CloseResponseBody(resp)

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("flow not found: %s", flowID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get flow: status %d", resp.StatusCode)
	}

	var flowResponse struct {
		ID    string                   `json:"id"`
		Label string                   `json:"label"`
		Info  string                   `json:"info"`
		Nodes []map[string]interface{} `json:"nodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&flowResponse); err != nil {
		return nil, fmt.Errorf("failed to decode flow: %w", err)
	}

	var nodes []types.Node
	for _, m := range flowResponse.Nodes {
		nodeObj := types.Node{
			ID:   util.GetString(m, "id"),
			Type: util.GetString(m, "type"),
			Name: util.GetString(m, "name"),
			Position: types.Position{
				X: util.GetFloat64(m, "x"),
				Y: util.GetFloat64(m, "y"),
			},
			Properties: make(map[string]interface{}),
		}
		for k, v := range m {
			if k != "id" && k != "type" && k != "name" && k != "x" && k != "y" && k != "z" && k != "wires" {
				nodeObj.Properties[k] = v
			}
		}
		if wires, ok := m["wires"].([]interface{}); ok {
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

	return &types.FlowDefinition{
		ID:          flowResponse.ID,
		Name:        flowResponse.Label,
		Description: flowResponse.Info,
		Nodes:       nodes,
	}, nil
}

// GetFlows retrieves all flows
func (s *Service) GetFlows(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := s.httpClient.DoRequest(ctx, types.MethodGet, fmt.Sprintf("%s/flows", s.httpClient.BaseURL()), nil)
	if err != nil {
		return nil, err
	}
	defer s.httpClient.CloseResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get flows: status %d", resp.StatusCode)
	}

	var flows []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&flows); err != nil {
		return nil, fmt.Errorf("failed to decode flows: %w", err)
	}

	return flows, nil
}

// DeleteFlow deletes a flow by ID
func (s *Service) DeleteFlow(ctx context.Context, flowID string) error {
	url := fmt.Sprintf("%s/flows/%s", s.httpClient.BaseURL(), flowID)
	resp, err := s.httpClient.DoRequest(ctx, types.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	defer s.httpClient.CloseResponseBody(resp)

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("flow not found: %s", flowID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete flow: status %d", resp.StatusCode)
	}

	return nil
}
