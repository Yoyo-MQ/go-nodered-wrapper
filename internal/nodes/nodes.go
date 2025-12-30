package nodes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/yoyo-mq/go-nodered-wrapper/internal/httpclient"
	"github.com/yoyo-mq/go-nodered-wrapper/internal/util"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// Service handles node operations
type Service struct {
	httpClient *httpclient.Client
	flows      FlowGetter
	logger     *slog.Logger
}

// FlowGetter interface for getting flows (to avoid circular dependency)
type FlowGetter interface {
	GetFlow(ctx context.Context, flowID string) (*types.FlowDefinition, error)
}

// NewService creates a new nodes service
func NewService(httpClient *httpclient.Client, flows FlowGetter, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		httpClient: httpClient,
		flows:      flows,
		logger:     logger,
	}
}

// TriggerNode triggers an inject node with the given input data
func (s *Service) TriggerNode(ctx context.Context, nodeID string, input map[string]interface{}) error {
	// Build inject properties from input
	props, err := buildInjectProps(input)
	if err != nil {
		return err
	}

	// Format request body with __user_inject_props__
	requestBody := map[string]interface{}{
		types.PropUserInject: props,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/inject/%s", s.httpClient.BaseURL(), nodeID)
	if s.httpClient.Debug() {
		s.logger.Debug("Triggering node", "nodeID", nodeID, "data", string(jsonData))
	}

	resp, err := s.httpClient.DoRequest(ctx, types.MethodPost, url, jsonData)
	if err != nil {
		return err
	}
	defer s.httpClient.CloseResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to trigger node: status %d, failed to read error body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("failed to trigger node: status %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ExecuteFlow executes a flow by finding and triggering its inject node
func (s *Service) ExecuteFlow(ctx context.Context, flowID string, input map[string]interface{}) (*types.ExecutionResult, error) {
	flow, err := s.flows.GetFlow(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	for _, node := range flow.Nodes {
		if node.Type == types.NodeTypeInject {
			if err := s.TriggerNode(ctx, node.ID, input); err != nil {
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

// buildInjectProps converts input data into Node-RED __user_inject_props__ format
func buildInjectProps(input map[string]interface{}) ([]map[string]interface{}, error) {
	props := []map[string]interface{}{}
	hasPayload := false

	// Handle payload - set as msg.payload with json type
	if payload, hasPayloadInput := input["payload"]; hasPayloadInput {
		payloadValue, err := util.MarshalPayload(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}

		props = append(props, map[string]interface{}{
			"p":  types.PropPayload,
			"v":  payloadValue,
			"vt": types.DataTypeJSON,
		})
		hasPayload = true
	}

	// Handle context - set as msg.context with json type
	if inputContext, hasContext := input["context"]; hasContext {
		contextJSON, err := json.Marshal(inputContext)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal context: %w", err)
		}
		props = append(props, map[string]interface{}{
			"p":  types.PropContext,
			"v":  string(contextJSON),
			"vt": types.DataTypeJSON,
		})
	}

	// Handle other top-level properties
	for key, value := range input {
		if key == types.PropPayload || key == types.PropContext {
			continue // Already handled above
		}
		valueJSON, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal property %s: %w", key, err)
		}
		props = append(props, map[string]interface{}{
			"p":  key,
			"v":  string(valueJSON),
			"vt": types.DataTypeJSON,
		})
	}

	// Always ensure we have at least a payload property
	if !hasPayload {
		props = append(props, map[string]interface{}{
			"p":  types.PropPayload,
			"v":  "{}",
			"vt": types.DataTypeJSON,
		})
	}

	return props, nil
}
