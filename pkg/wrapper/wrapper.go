package wrapper

import (
	"context"
	"fmt"
	"time"

	"github.com/yoyo-mq/go-nodered-wrapper/internal/client"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// NodeRedWrapper provides a high-level interface for managing Node-RED workflows
type NodeRedWrapper struct {
	client    *client.NodeRedClient
	converter WorkflowConverter
	executor  ExecutionHandler
	config    *types.Config
}

// WorkflowConverter interface for converting workflows to Node-RED format
type WorkflowConverter interface {
	ConvertToNodeRedFlow(workflow interface{}) (*types.FlowDefinition, error)
	ConvertFromNodeRedFlow(flow *types.FlowDefinition) (interface{}, error)
}

// ExecutionHandler interface for handling flow execution
type ExecutionHandler interface {
	PreExecute(ctx context.Context, input map[string]interface{}) error
	PostExecute(ctx context.Context, result *types.ExecutionResult) error
	OnError(ctx context.Context, err error) error
}

// New creates a new Node-RED wrapper instance
func New(config *types.Config) (*NodeRedWrapper, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	client, err := client.NewNodeRedClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Node-RED client: %w", err)
	}

	return &NodeRedWrapper{
		client:    client,
		converter: &DefaultConverter{},
		executor:  &DefaultExecutor{},
		config:    config,
	}, nil
}

// NewWithConverter creates a new wrapper with a custom converter
func NewWithConverter(config *types.Config, converter WorkflowConverter) (*NodeRedWrapper, error) {
	wrapper, err := New(config)
	if err != nil {
		return nil, err
	}

	if converter != nil {
		wrapper.converter = converter
	}

	return wrapper, nil
}

// NewWithExecutor creates a new wrapper with a custom executor
func NewWithExecutor(config *types.Config, executor ExecutionHandler) (*NodeRedWrapper, error) {
	wrapper, err := New(config)
	if err != nil {
		return nil, err
	}

	if executor != nil {
		wrapper.executor = executor
	}

	return wrapper, nil
}

// DeployFlow deploys a workflow to Node-RED
func (w *NodeRedWrapper) DeployFlow(ctx context.Context, flow *types.FlowDefinition) error {
	if flow == nil {
		return fmt.Errorf("flow is required")
	}

	if flow.ID == "" {
		return fmt.Errorf("flow ID is required")
	}

	return w.client.DeployFlow(ctx, flow)
}

// DeployWorkflow deploys a workflow using the converter
func (w *NodeRedWrapper) DeployWorkflow(ctx context.Context, workflow interface{}) error {
	flow, err := w.converter.ConvertToNodeRedFlow(workflow)
	if err != nil {
		return fmt.Errorf("failed to convert workflow: %w", err)
	}

	return w.DeployFlow(ctx, flow)
}

// ExecuteFlow triggers a workflow execution
func (w *NodeRedWrapper) ExecuteFlow(ctx context.Context, flowID string, input map[string]interface{}) (*types.ExecutionResult, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow ID is required")
	}

	// Pre-execution hook
	if err := w.executor.PreExecute(ctx, input); err != nil {
		return nil, fmt.Errorf("pre-execution failed: %w", err)
	}

	startTime := time.Now()
	result, err := w.client.ExecuteFlow(ctx, flowID, input)
	if err != nil {
		// Error hook
		if execErr := w.executor.OnError(ctx, err); execErr != nil {
			return nil, fmt.Errorf("execution failed and error handler failed: %w (original error: %v)", execErr, err)
		}
		return nil, err
	}

	// Calculate duration
	result.Duration = time.Since(startTime)

	// Post-execution hook
	if err := w.executor.PostExecute(ctx, result); err != nil {
		return nil, fmt.Errorf("post-execution failed: %w", err)
	}

	return result, nil
}

// ExecuteWorkflow executes a workflow using the converter
func (w *NodeRedWrapper) ExecuteWorkflow(ctx context.Context, workflow interface{}, input map[string]interface{}) (*types.ExecutionResult, error) {
	flow, err := w.converter.ConvertToNodeRedFlow(workflow)
	if err != nil {
		return nil, fmt.Errorf("failed to convert workflow: %w", err)
	}

	return w.ExecuteFlow(ctx, flow.ID, input)
}

// GetFlow retrieves a deployed flow
func (w *NodeRedWrapper) GetFlow(ctx context.Context, flowID string) (*types.FlowDefinition, error) {
	if flowID == "" {
		return nil, fmt.Errorf("flow ID is required")
	}

	return w.client.GetFlow(ctx, flowID)
}

// DeleteFlow removes a flow from Node-RED
func (w *NodeRedWrapper) DeleteFlow(ctx context.Context, flowID string) error {
	if flowID == "" {
		return fmt.Errorf("flow ID is required")
	}

	return w.client.DeleteFlow(ctx, flowID)
}

// HealthCheck checks if Node-RED is healthy
func (w *NodeRedWrapper) HealthCheck(ctx context.Context) error {
	return w.client.HealthCheck(ctx)
}

// GetConfig returns the current configuration
func (w *NodeRedWrapper) GetConfig() *types.Config {
	return w.config
}

// SetConverter sets a custom converter
func (w *NodeRedWrapper) SetConverter(converter WorkflowConverter) {
	if converter != nil {
		w.converter = converter
	}
}

// SetExecutor sets a custom executor
func (w *NodeRedWrapper) SetExecutor(executor ExecutionHandler) {
	if executor != nil {
		w.executor = executor
	}
}
