package main

import (
	"context"
	"log"
	"time"

	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
	nodered "github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

func main() {
	// Create configuration for Yoyo integration
	config := &types.Config{
		NodeRedURL: "http://localhost:1880",
		APIKey:     "your-node-red-api-key",
		Timeout:    30 * time.Second,
		Debug:      true,
	}

	// Create wrapper instance
	wrapper, err := nodered.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Deploy a workflow with debug nodes
	flow := &types.FlowDefinition{
		ID:   "yoyo-workflow-with-debug",
		Name: "Yoyo Workflow with Debug",
		Nodes: []types.Node{
			{
				ID:   "debug-node-1",
				Type: "yoyo-debug",
				Name: "Debug Node",
				Properties: map[string]interface{}{
					"workflowId":      "yoyo-workflow-with-debug",
					"nodeId":          "debug-node-1",
					"yoyoEndpoint":    "http://app:8080",
					"apiKey":          "your-yoyo-api-key",
					"level":           "info",
					"outputToSidebar": true,
				},
			},
		},
	}

	// Deploy the workflow
	err = wrapper.DeployFlow(context.Background(), flow)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Workflow deployed successfully with debug capabilities!")
}
