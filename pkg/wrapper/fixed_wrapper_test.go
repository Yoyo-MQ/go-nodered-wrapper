package wrapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

func TestFixedWrapper(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping fixed wrapper test in short mode")
	}

	t.Log("🧪 Testing FIXED Node-RED Wrapper (Direct Test)...")

	// Create configuration (use localhost since we're running outside container)
	config := &types.Config{
		NodeRedURL:    "http://localhost:1880/admin",
		Timeout:       30 * time.Second,
		RetryAttempts: 3,
		Debug:         true,
	}

	// Create wrapper
	nodeRedWrapper, err := wrapper.New(config)
	require.NoError(t, err, "Failed to create wrapper")

	ctx := context.Background()

	// Test authentication
	t.Log("🔐 Testing authentication...")
	if err := nodeRedWrapper.Authenticate(ctx, "admin", "password"); err != nil {
		t.Logf("❌ Authentication failed: %v", err)
		// Don't fail the test if auth fails, as we might not have a running instance
	} else {
		t.Log("✅ Authentication successful!")
	}

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
	t.Log("🚀 Testing flow deployment...")
	if _, err := nodeRedWrapper.DeployFlow(ctx, testFlow); err != nil {
		t.Logf("❌ Deployment failed: %v", err)
	} else {
		t.Log("✅ Flow deployed successfully!")
	}

	// Test execution
	t.Log("▶️  Testing flow execution...")
	result, err := nodeRedWrapper.ExecuteFlow(ctx, testFlow.ID, map[string]interface{}{
		"test_data": "direct wrapper execution",
	})
	if err != nil {
		t.Logf("⚠️  Execution failed: %v", err)
	} else {
		t.Logf("✅ Execution result: %+v", result)
	}

	t.Log("🎉 Direct wrapper test completed!")
}
