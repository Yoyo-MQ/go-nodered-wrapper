package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

func main() {
	log.Println("🧪 Testing Node-RED Wrapper Integration...")

	// Create configuration for Node-RED
	config := &types.Config{
		NodeRedURL:    "http://localhost:1880",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		Debug:         true,
	}

	// Create wrapper instance
	log.Println("📦 Creating Node-RED wrapper...")
	nodeRedWrapper, err := wrapper.New(config)
	if err != nil {
		log.Fatalf("❌ Failed to create Node-RED wrapper: %v", err)
	}

	// Test health check first
	log.Println("🏥 Testing Node-RED health check...")
	ctx := context.Background()
	if err := nodeRedWrapper.HealthCheck(ctx); err != nil {
		log.Printf("⚠️  Health check failed: %v", err)
		log.Println("📝 Continuing with flow deployment test anyway...")
	} else {
		log.Println("✅ Node-RED health check passed!")
	}

	// Create a simple test flow
	log.Println("🔧 Creating simple test flow...")
	testFlow := createSimpleTestFlow()

	// Deploy the flow
	log.Println("🚀 Deploying test flow to Node-RED...")
	if err := nodeRedWrapper.DeployFlow(ctx, testFlow); err != nil {
		log.Fatalf("❌ Failed to deploy flow: %v", err)
	}

	log.Printf("✅ Successfully deployed flow '%s' to Node-RED!", testFlow.Name)
	log.Printf("🎯 Flow ID: %s", testFlow.ID)
	log.Println("📋 You can now check the Node-RED UI at http://localhost:1880/admin/ to see the deployed flow!")

	// Try to retrieve the deployed flow
	log.Println("🔍 Retrieving deployed flow...")
	retrievedFlow, err := nodeRedWrapper.GetFlow(ctx, testFlow.ID)
	if err != nil {
		log.Printf("⚠️  Failed to retrieve flow: %v", err)
	} else {
		log.Printf("✅ Successfully retrieved flow: %s", retrievedFlow.Name)
	}

	// Test execution with sample data
	log.Println("▶️  Testing flow execution...")
	sampleInput := map[string]interface{}{
		"temperature": 25.5,
		"humidity":    60.0,
		"device_id":   "sensor-001",
		"timestamp":   time.Now().Unix(),
	}

	result, err := nodeRedWrapper.ExecuteFlow(ctx, testFlow.ID, sampleInput)
	if err != nil {
		log.Printf("⚠️  Flow execution failed: %v", err)
	} else {
		log.Printf("✅ Flow executed successfully!")
		log.Printf("📊 Execution result: %+v", result)
	}

	log.Println("🎉 Node-RED integration test completed!")
	log.Println("👀 Check the Node-RED UI to see your deployed flow:")
	log.Println("   URL: http://localhost:1880/admin/")
	log.Println("   Username: admin")
	log.Println("   Password: password")
}

// createSimpleTestFlow creates a basic test flow for demonstration
func createSimpleTestFlow() *types.FlowDefinition {
	now := time.Now()

	return &types.FlowDefinition{
		ID:          fmt.Sprintf("yoyo-test-flow-%d", now.Unix()),
		Name:        "Yoyo Integration Test Flow",
		Version:     "1.0.0",
		Description: "A simple test flow created by Yoyo Node-RED wrapper",
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata: map[string]interface{}{
			"created_by": "yoyo-wrapper",
			"test_flow":  true,
		},
		Nodes: []types.Node{
			{
				ID:   "inject-1",
				Type: "inject",
				Name: "Data Input",
				Position: types.Position{
					X: 100,
					Y: 100,
				},
				Properties: map[string]interface{}{
					"payload":     "",
					"payloadType": "json",
					"topic":       "test/data",
					"repeat":      "",
					"crontab":     "",
					"once":        false,
				},
				Wires: [][]string{{"function-1"}},
			},
			{
				ID:   "function-1",
				Type: "function",
				Name: "Process Data",
				Position: types.Position{
					X: 300,
					Y: 100,
				},
				Properties: map[string]interface{}{
					"func": `
// Process incoming sensor data
const data = msg.payload;

// Add processing timestamp
data.processed_at = new Date().toISOString();

// Simple temperature alert logic
if (data.temperature && data.temperature > 30) {
    data.alert = "HIGH_TEMPERATURE";
    data.message = "Temperature is above 30°C!";
} else if (data.temperature && data.temperature < 10) {
    data.alert = "LOW_TEMPERATURE"; 
    data.message = "Temperature is below 10°C!";
} else {
    data.alert = "NORMAL";
    data.message = "Temperature is normal";
}

// Log the processing
node.log("Processed data for device: " + (data.device_id || "unknown"));

msg.payload = data;
return msg;
`,
					"outputs": 1,
				},
				Wires: [][]string{{"debug-1"}},
			},
			{
				ID:   "debug-1",
				Type: "debug",
				Name: "Debug Output",
				Position: types.Position{
					X: 500,
					Y: 100,
				},
				Properties: map[string]interface{}{
					"name":      "Debug Output",
					"active":    true,
					"tosidebar": true,
					"console":   false,
					"tostatus":  false,
					"complete":  "payload",
				},
				Wires: [][]string{},
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
