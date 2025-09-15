package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/yourusername/yoyo-nodered-wrapper/pkg/types"
	nodered "github.com/yourusername/yoyo-nodered-wrapper/pkg/wrapper"
)

func main() {
	var (
		nodeRedURL = flag.String("url", "http://localhost:1880", "Node-RED URL")
		apiKey     = flag.String("key", "", "Node-RED API key")
		timeout    = flag.Duration("timeout", 30*time.Second, "Request timeout")
		debug      = flag.Bool("debug", false, "Enable debug mode")
		action     = flag.String("action", "health", "Action to perform: health, deploy, execute, get, delete")
		flowID     = flag.String("flow-id", "example-flow", "Flow ID for operations")
	)
	flag.Parse()

	// Create configuration
	config := &types.Config{
		NodeRedURL: *nodeRedURL,
		APIKey:     *apiKey,
		Timeout:    *timeout,
		Debug:      *debug,
	}

	// Create wrapper instance
	wrapper, err := nodered.New(config)
	if err != nil {
		log.Fatal("Failed to create wrapper:", err)
	}

	ctx := context.Background()

	switch *action {
	case "health":
		if err := wrapper.HealthCheck(ctx); err != nil {
			log.Fatal("Health check failed:", err)
		}
		fmt.Println("✅ Node-RED is healthy!")

	case "deploy":
		flow := createExampleFlow(*flowID)
		if err := wrapper.DeployFlow(ctx, flow); err != nil {
			log.Fatal("Failed to deploy flow:", err)
		}
		fmt.Printf("✅ Flow '%s' deployed successfully!\n", *flowID)

	case "execute":
		input := map[string]interface{}{
			"message":   "Hello from CLI!",
			"timestamp": time.Now().Unix(),
		}
		result, err := wrapper.ExecuteFlow(ctx, *flowID, input)
		if err != nil {
			log.Fatal("Failed to execute flow:", err)
		}
		fmt.Printf("✅ Flow executed successfully! Result: %+v\n", result)

	case "get":
		flow, err := wrapper.GetFlow(ctx, *flowID)
		if err != nil {
			log.Fatal("Failed to get flow:", err)
		}
		fmt.Printf("✅ Retrieved flow: %s - %s\n", flow.ID, flow.Name)
		fmt.Printf("   Description: %s\n", flow.Description)
		fmt.Printf("   Nodes: %d\n", len(flow.Nodes))
		fmt.Printf("   Connections: %d\n", len(flow.Connections))

	case "delete":
		if err := wrapper.DeleteFlow(ctx, *flowID); err != nil {
			log.Fatal("Failed to delete flow:", err)
		}
		fmt.Printf("✅ Flow '%s' deleted successfully!\n", *flowID)

	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Available actions: health, deploy, execute, get, delete")
		os.Exit(1)
	}
}

func createExampleFlow(flowID string) *types.FlowDefinition {
	return &types.FlowDefinition{
		ID:          flowID,
		Name:        "CLI Example Flow",
		Description: "A simple flow created from the CLI",
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
				Wires: [][]string{{"function-1"}},
				Properties: map[string]interface{}{
					"payload":     "",
					"payloadType": "json",
				},
			},
			{
				ID:   "function-1",
				Type: "function",
				Name: "Process",
				Position: types.Position{
					X: 300,
					Y: 100,
				},
				Wires: [][]string{{"debug-1"}},
				Properties: map[string]interface{}{
					"func": `
msg.payload = {
    message: "Processed: " + msg.payload.message,
    timestamp: msg.payload.timestamp,
    processed_at: new Date().toISOString()
};
return msg;`,
				},
			},
			{
				ID:   "debug-1",
				Type: "debug",
				Name: "Log",
				Position: types.Position{
					X: 500,
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
				Target: "function-1",
			},
			{
				Source: "function-1",
				Target: "debug-1",
			},
		},
	}
}
