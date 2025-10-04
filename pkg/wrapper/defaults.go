package wrapper

import (
	"context"
	"fmt"

	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// DefaultConverter provides a basic implementation of WorkflowConverter
type DefaultConverter struct{}

// ConvertToNodeRedFlow converts a workflow to Node-RED format
func (c *DefaultConverter) ConvertToNodeRedFlow(workflow interface{}) (*types.FlowDefinition, error) {
	// Try to convert from FlowDefinition directly
	if flow, ok := workflow.(*types.FlowDefinition); ok {
		return flow, nil
	}

	// Try to convert from map
	if flowMap, ok := workflow.(map[string]interface{}); ok {
		flow := &types.FlowDefinition{}

		if id, ok := flowMap["id"].(string); ok {
			flow.ID = id
		}
		if name, ok := flowMap["name"].(string); ok {
			flow.Name = name
		}
		if desc, ok := flowMap["description"].(string); ok {
			flow.Description = desc
		}
		if version, ok := flowMap["version"].(string); ok {
			flow.Version = version
		}

		return flow, nil
	}

	return nil, fmt.Errorf("unsupported workflow type: %T", workflow)
}

// ConvertFromNodeRedFlow converts a Node-RED flow to a workflow
func (c *DefaultConverter) ConvertFromNodeRedFlow(flow *types.FlowDefinition) (interface{}, error) {
	if flow == nil {
		return nil, fmt.Errorf("flow is required")
	}

	return flow, nil
}

// DefaultExecutor provides a basic implementation of ExecutionHandler
type DefaultExecutor struct{}

// PreExecute is called before flow execution
func (e *DefaultExecutor) PreExecute(ctx context.Context, input map[string]interface{}) error {
	// Default implementation does nothing
	return nil
}

// PostExecute is called after flow execution
func (e *DefaultExecutor) PostExecute(ctx context.Context, result *types.ExecutionResult) error {
	// Default implementation does nothing
	return nil
}

// OnError is called when execution fails
func (e *DefaultExecutor) OnError(ctx context.Context, err error) error {
	// Default implementation does nothing
	return nil
}
