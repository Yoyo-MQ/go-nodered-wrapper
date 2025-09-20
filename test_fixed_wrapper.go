package main

import (
	"context"
	"log"
	"time"

	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

func main() {
	log.Println("🧪 Testing FIXED Node-RED Wrapper (Direct Test)...")

	// Create configuration (use localhost since we're running outside container)
	config := &types.Config{
		NodeRedURL:    "http://localhost:1880/admin",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		Debug:         true,
	}

	// Create wrapper
	nodeRedWrapper, err := wrapper.New(config)
	if err != nil {
		log.Fatalf("❌ Failed to create wrapper: %v", err)
	}

	ctx := context.Background()

	// Test authentication
	log.Println("🔐 Testing authentication...")
	if err := nodeRedWrapper.Authenticate(ctx, "admin", "password"); err != nil {
		log.Fatalf("❌ Authentication failed: %v", err)
	}
	log.Println("✅ Authentication successful!")

	// Create simple test flow
	testFlow := &types.FlowDefinition{
		ID:          "wrapper-direct-test",
		Name:        "Direct Wrapper Test",
		Description: "Testing the fixed wrapper directly",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Nodes: []types.Node{
			{
				ID:       "wrapper-direct-test-inject-1",
				Type:     "inject",
				Name:     "Direct Test Input",
				Position: types.Position{X: 100, Y: 100},
				Properties: map[string]interface{}{
					"payload":     `{"test": "direct wrapper"}`,
					"payloadType": "json",
					"topic":       "direct/test",
				},
				Wires: [][]string{{"wrapper-direct-test-debug-1"}},
			},
			{
				ID:       "wrapper-direct-test-debug-1",
				Type:     "debug",
				Name:     "Direct Test Output",
				Position: types.Position{X: 300, Y: 100},
				Properties: map[string]interface{}{
					"complete": "payload",
					"active":   true,
				},
				Wires: [][]string{},
			},
		},
	}

	// Test deployment
	log.Println("🚀 Testing flow deployment...")
	if err := nodeRedWrapper.DeployFlow(ctx, testFlow); err != nil {
		log.Fatalf("❌ Deployment failed: %v", err)
	}
	log.Println("✅ Flow deployed successfully!")

	// Test execution
	log.Println("▶️  Testing flow execution...")
	result, err := nodeRedWrapper.ExecuteFlow(ctx, testFlow.ID, map[string]interface{}{
		"test_data": "direct wrapper execution",
	})
	if err != nil {
		log.Printf("⚠️  Execution failed: %v", err)
	} else {
		log.Printf("✅ Execution result: %+v", result)
	}

	log.Println("🎉 Direct wrapper test completed!")
}
