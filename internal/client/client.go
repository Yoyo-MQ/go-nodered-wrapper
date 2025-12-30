package client

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/yoyo-mq/go-nodered-wrapper/internal/auth"
	"github.com/yoyo-mq/go-nodered-wrapper/internal/flows"
	"github.com/yoyo-mq/go-nodered-wrapper/internal/httpclient"
	"github.com/yoyo-mq/go-nodered-wrapper/internal/nodes"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// NodeRedClient is a facade that composes flow, node, and auth services
type NodeRedClient struct {
	flows      *flows.Service
	nodes      *nodes.Service
	auth       *auth.Service
	httpClient *httpclient.Client
	logger     *slog.Logger
}

// NewNodeRedClient creates a new Node-RED client
func NewNodeRedClient(config *types.Config) (*NodeRedClient, error) {
	if config.NodeRedURL == "" {
		return nil, fmt.Errorf("node_red_url is required")
	}

	httpClient := httpclient.NewClient(
		config.NodeRedURL,
		&http.Client{Timeout: config.Timeout},
		config.APIKey,
		config.Debug,
		config.Logger,
	)

	flowsService := flows.NewService(httpClient, config.Logger)
	nodesService := nodes.NewService(httpClient, flowsService, config.Logger)
	authService := auth.NewService(httpClient, config.Logger)

	return &NodeRedClient{
		flows:      flowsService,
		nodes:      nodesService,
		auth:       authService,
		httpClient: httpClient,
		logger:     config.Logger,
	}, nil
}

// DeployFlow deploys a flow to Node-RED
func (c *NodeRedClient) DeployFlow(ctx context.Context, flow *types.FlowDefinition) (string, error) {
	return c.flows.DeployFlow(ctx, flow)
}

// ExecuteFlow executes a flow by finding and triggering its inject node
func (c *NodeRedClient) ExecuteFlow(ctx context.Context, flowID string, input map[string]interface{}) (*types.ExecutionResult, error) {
	return c.nodes.ExecuteFlow(ctx, flowID, input)
}

// TriggerNode triggers an inject node with the given input data
func (c *NodeRedClient) TriggerNode(ctx context.Context, nodeID string, input map[string]interface{}) error {
	return c.nodes.TriggerNode(ctx, nodeID, input)
}

// GetFlow retrieves a flow by ID
func (c *NodeRedClient) GetFlow(ctx context.Context, flowID string) (*types.FlowDefinition, error) {
	return c.flows.GetFlow(ctx, flowID)
}

// GetFlows retrieves all flows
func (c *NodeRedClient) GetFlows(ctx context.Context) ([]map[string]interface{}, error) {
	return c.flows.GetFlows(ctx)
}

// DeleteFlow deletes a flow by ID
func (c *NodeRedClient) DeleteFlow(ctx context.Context, flowID string) error {
	return c.flows.DeleteFlow(ctx, flowID)
}

// HealthCheck checks if Node-RED is healthy
func (c *NodeRedClient) HealthCheck(ctx context.Context) error {
	baseURL := c.httpClient.BaseURL()
	if len(baseURL) > 6 && baseURL[len(baseURL)-6:] == "/admin" {
		baseURL = baseURL[:len(baseURL)-6]
	}

	resp, err := c.httpClient.DoRequest(ctx, "GET", baseURL+"/", nil)
	if err != nil {
		return err
	}
	defer c.httpClient.CloseResponseBody(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf("Node-RED is not healthy: status %d", resp.StatusCode)
	}

	return nil
}

// GetAuthToken retrieves an authentication token
func (c *NodeRedClient) GetAuthToken(ctx context.Context, username, password string) (string, error) {
	return c.auth.GetAuthToken(ctx, username, password)
}
