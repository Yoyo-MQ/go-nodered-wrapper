package main

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/yoyo-nodered-wrapper/pkg/types"
	nodered "github.com/yourusername/yoyo-nodered-wrapper/pkg/wrapper"
)

func main() {
	// Create configuration
	config := &types.Config{
		NodeRedURL: "http://localhost:1880",
		APIKey:     "your-api-key", // Optional
		Timeout:    30 * time.Second,
		Debug:      true,
	}

	// Create wrapper instance
	wrapper, err := nodered.New(config)
	if err != nil {
		log.Fatal("Failed to create wrapper:", err)
	}

	// Check Node-RED health
	ctx := context.Background()
	if err := wrapper.HealthCheck(ctx); err != nil {
		log.Printf("Node-RED health check failed: %v", err)
		log.Println("Make sure Node-RED is running on http://localhost:1880")
		return
	}

	log.Println("Node-RED is healthy!")

	// Create a simple flow
	flow := &types.FlowDefinition{
		ID:          "example-flow",
		Name:        "Example Flow",
		Description: "A simple example flow",
		Version:     "1.0.0",
		Nodes: []types.Node{
			{
				ID:   "inject-1",
				Type: "inject",
				Name: "Start",
				Position: types.Position{
					X: 100,
					Y: 100,
				},
				Wires: [][]string{{"debug-1"}},
				Properties: map[string]interface{}{
					"payload":     "Hello World",
					"payloadType": "str",
				},
			},
			{
				ID:   "debug-1",
				Type: "debug",
				Name: "Log",
				Position: types.Position{
					X: 300,
					Y: 100,
				},
				Wires: [][]string{},
				Properties: map[string]interface{}{
					"complete": "payload",
				},
			},
		},
		Connections: []types.Connection{
			{
				Source: "inject-1",
				Target: "debug-1",
			},
		},
	}

	// Deploy the flow
	log.Println("Deploying flow...")
	if err := wrapper.DeployFlow(ctx, flow); err != nil {
		log.Fatal("Failed to deploy flow:", err)
	}
	log.Println("Flow deployed successfully!")

	// Execute the flow
	log.Println("Executing flow...")
	input := map[string]interface{}{
		"message":   "Hello from Go!",
		"timestamp": time.Now().Unix(),
	}

	result, err := wrapper.ExecuteFlow(ctx, flow.ID, input)
	if err != nil {
		log.Fatal("Failed to execute flow:", err)
	}

	log.Printf("Execution result: %+v", result)

	// Get the deployed flow
	log.Println("Retrieving flow...")
	retrievedFlow, err := wrapper.GetFlow(ctx, flow.ID)
	if err != nil {
		log.Fatal("Failed to get flow:", err)
	}

	log.Printf("Retrieved flow: %s - %s", retrievedFlow.ID, retrievedFlow.Name)

	// Clean up - delete the flow
	log.Println("Cleaning up...")
	if err := wrapper.DeleteFlow(ctx, flow.ID); err != nil {
		log.Printf("Failed to delete flow: %v", err)
	} else {
		log.Println("Flow deleted successfully!")
	}
}
