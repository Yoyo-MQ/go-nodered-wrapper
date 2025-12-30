package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// Client provides HTTP operations for Node-RED API
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	debug      bool
	logger     *slog.Logger
}

// NewClient creates a new HTTP client
func NewClient(baseURL string, httpClient *http.Client, apiKey string, debug bool, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		apiKey:     apiKey,
		debug:      debug,
		logger:     logger,
	}
}

// DoRequest performs an HTTP request
func (c *Client) DoRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	return resp, nil
}

// CloseResponseBody closes the response body, logging errors if debug is enabled
func (c *Client) CloseResponseBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil && c.debug {
		fmt.Printf("Warning: failed to close response body: %v\n", err)
	}
}

// BaseURL returns the base URL
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Debug returns whether debug mode is enabled
func (c *Client) Debug() bool {
	return c.debug
}
