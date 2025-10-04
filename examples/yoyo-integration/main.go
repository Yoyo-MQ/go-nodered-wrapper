package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
	nodered "github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

// YoyoWorkflow represents a Yoyo workflow that can be converted to Node-RED
type YoyoWorkflow struct {
	UUID        string                 `json:"uuid"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Steps       []YoyoWorkflowStep     `json:"steps"`
	Triggers    []YoyoWorkflowTrigger  `json:"triggers"`
	Config      map[string]interface{} `json:"config"`
}

type YoyoWorkflowStep struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"` // condition, action
	Order  int                    `json:"order"`
	Config map[string]interface{} `json:"config"`
}

type YoyoWorkflowTrigger struct {
	ID                    string                 `json:"id"`
	Type                  string                 `json:"type"` // DEVICE_TRIGGER, CRON, etc.
	Config                map[string]interface{} `json:"config"`
	WorkflowTriggerableID int64                  `json:"workflow_triggerable_id"`
}

// YoyoConverter converts Yoyo workflows to Node-RED flows
type YoyoConverter struct{}

// ConvertToNodeRedFlow converts a Yoyo workflow to Node-RED format
func (c *YoyoConverter) ConvertToNodeRedFlow(workflow interface{}) (*types.FlowDefinition, error) {
	yoyoWorkflow, ok := workflow.(*YoyoWorkflow)
	if !ok {
		return nil, fmt.Errorf("expected *YoyoWorkflow, got %T", workflow)
	}

	flow := &types.FlowDefinition{
		ID:          yoyoWorkflow.UUID,
		Name:        yoyoWorkflow.Name,
		Description: yoyoWorkflow.Description,
		Version:     "1.0.0",
		Nodes:       []types.Node{},
		Connections: []types.Connection{},
		Metadata: map[string]interface{}{
			"yoyo_workflow": true,
			"enabled":       yoyoWorkflow.Enabled,
		},
	}

	// Convert triggers to injection nodes
	for i, trigger := range yoyoWorkflow.Triggers {
		node := types.Node{
			ID:   fmt.Sprintf("trigger-%d", i),
			Type: "inject",
			Name: fmt.Sprintf("Trigger %s", trigger.Type),
			Position: types.Position{
				X: 100,
				Y: float64(100 + i*100),
			},
			Properties: map[string]interface{}{
				"payload":     "",
				"payloadType": "json",
			},
		}
		flow.Nodes = append(flow.Nodes, node)
	}

	// Convert steps to appropriate Node-RED nodes
	for i, step := range yoyoWorkflow.Steps {
		var node types.Node

		switch step.Type {
		case "condition":
			node = types.Node{
				ID:   fmt.Sprintf("step-%d", i),
				Type: "switch",
				Name: fmt.Sprintf("Condition %d", i),
				Position: types.Position{
					X: 300,
					Y: float64(100 + i*100),
				},
				Properties: map[string]interface{}{
					"property": "payload",
					"rules": []map[string]interface{}{
						{
							"t":  "true",
							"v":  "true",
							"vt": "bool",
						},
					},
				},
			}
		case "action":
			node = types.Node{
				ID:   fmt.Sprintf("step-%d", i),
				Type: "function",
				Name: fmt.Sprintf("Action %d", i),
				Position: types.Position{
					X: 500,
					Y: float64(100 + i*100),
				},
				Properties: map[string]interface{}{
					"func": "msg.payload = msg.payload; return msg;",
				},
			}
		default:
			// Default to function node
			node = types.Node{
				ID:   fmt.Sprintf("step-%d", i),
				Type: "function",
				Name: fmt.Sprintf("Step %d", i),
				Position: types.Position{
					X: 500,
					Y: float64(100 + i*100),
				},
				Properties: map[string]interface{}{
					"func": "return msg;",
				},
			}
		}

		flow.Nodes = append(flow.Nodes, node)
	}

	// Add end node
	endNode := types.Node{
		ID:   "end-1",
		Type: "debug",
		Name: "End",
		Position: types.Position{
			X: 700,
			Y: 100,
		},
		Wires: [][]string{},
	}
	flow.Nodes = append(flow.Nodes, endNode)

	// Create connections
	for i := range yoyoWorkflow.Steps {
		if i == 0 {
			// Connect first trigger to first step
			flow.Connections = append(flow.Connections, types.Connection{
				Source: "trigger-0",
				Target: fmt.Sprintf("step-%d", i),
			})
		} else {
			// Connect step to next step
			flow.Connections = append(flow.Connections, types.Connection{
				Source: fmt.Sprintf("step-%d", i-1),
				Target: fmt.Sprintf("step-%d", i),
			})
		}
	}

	// Connect last step to end
	if len(yoyoWorkflow.Steps) > 0 {
		flow.Connections = append(flow.Connections, types.Connection{
			Source: fmt.Sprintf("step-%d", len(yoyoWorkflow.Steps)-1),
			Target: "end-1",
		})
	}

	return flow, nil
}

// ConvertFromNodeRedFlow converts a Node-RED flow back to Yoyo format
func (c *YoyoConverter) ConvertFromNodeRedFlow(flow *types.FlowDefinition) (interface{}, error) {
	if flow == nil {
		return nil, fmt.Errorf("flow is required")
	}

	yoyoWorkflow := &YoyoWorkflow{
		UUID:        flow.ID,
		Name:        flow.Name,
		Description: flow.Description,
		Enabled:     true,
		Steps:       []YoyoWorkflowStep{},
		Triggers:    []YoyoWorkflowTrigger{},
		Config:      flow.Metadata,
	}

	// Convert nodes back to steps
	for _, node := range flow.Nodes {
		switch node.Type {
		case "inject":
			// This is a trigger
			trigger := YoyoWorkflowTrigger{
				ID:   node.ID,
				Type: "DEVICE_TRIGGER", // Default type
				Config: map[string]interface{}{
					"payload": node.Properties["payload"],
				},
			}
			yoyoWorkflow.Triggers = append(yoyoWorkflow.Triggers, trigger)
		case "switch", "function":
			// This is a step
			stepType := "action"
			if node.Type == "switch" {
				stepType = "condition"
			}

			step := YoyoWorkflowStep{
				ID:     node.ID,
				Type:   stepType,
				Order:  len(yoyoWorkflow.Steps),
				Config: node.Properties,
			}
			yoyoWorkflow.Steps = append(yoyoWorkflow.Steps, step)
		}
	}

	return yoyoWorkflow, nil
}

func main() {
	// Create configuration
	config := &types.Config{
		NodeRedURL: "http://localhost:1880",
		APIKey:     "your-api-key",
		Timeout:    30 * time.Second,
		Debug:      true,
	}

	// Create wrapper with custom converter
	converter := &YoyoConverter{}
	wrapper, err := nodered.NewWithConverter(config, converter)
	if err != nil {
		log.Fatal("Failed to create wrapper:", err)
	}

	// Create a Yoyo workflow
	yoyoWorkflow := &YoyoWorkflow{
		UUID:        "yoyo-workflow-1",
		Name:        "Temperature Control Workflow",
		Description: "Controls AC based on temperature readings",
		Enabled:     true,
		Steps: []YoyoWorkflowStep{
			{
				ID:    "step-1",
				Type:  "condition",
				Order: 1,
				Config: map[string]interface{}{
					"expression": "temperature > 29",
				},
			},
			{
				ID:    "step-2",
				Type:  "action",
				Order: 2,
				Config: map[string]interface{}{
					"action": "turn_on_ac",
					"payload": map[string]interface{}{
						"state": "on",
					},
				},
			},
		},
		Triggers: []YoyoWorkflowTrigger{
			{
				ID:   "trigger-1",
				Type: "DEVICE_TRIGGER",
				Config: map[string]interface{}{
					"device_id": "sensor-1",
				},
				WorkflowTriggerableID: 1,
			},
		},
	}

	ctx := context.Background()

	// Deploy the Yoyo workflow (it will be converted to Node-RED)
	log.Println("Deploying Yoyo workflow...")
	if err := wrapper.DeployWorkflow(ctx, yoyoWorkflow); err != nil {
		log.Fatal("Failed to deploy workflow:", err)
	}
	log.Println("Workflow deployed successfully!")

	// Execute the workflow with device data
	log.Println("Executing workflow with device data...")
	deviceData := map[string]interface{}{
		"temperature": 30.5,
		"unit":        "celsius",
		"device_id":   "sensor-1",
		"timestamp":   time.Now().Unix(),
	}

	result, err := wrapper.ExecuteWorkflow(ctx, yoyoWorkflow, deviceData)
	if err != nil {
		log.Fatal("Failed to execute workflow:", err)
	}

	log.Printf("Execution result: %+v", result)

	// Get the deployed flow to verify
	log.Println("Retrieving deployed flow...")
	flow, err := wrapper.GetFlow(ctx, yoyoWorkflow.UUID)
	if err != nil {
		log.Fatal("Failed to get flow:", err)
	}

	log.Printf("Deployed flow: %s - %s", flow.ID, flow.Name)
	log.Printf("Number of nodes: %d", len(flow.Nodes))
	log.Printf("Number of connections: %d", len(flow.Connections))

	// Clean up
	log.Println("Cleaning up...")
	if err := wrapper.DeleteFlow(ctx, yoyoWorkflow.UUID); err != nil {
		log.Printf("Failed to delete flow: %v", err)
	} else {
		log.Println("Flow deleted successfully!")
	}
}
