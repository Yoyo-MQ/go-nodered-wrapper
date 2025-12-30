package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/yoyo-mq/go-nodered-wrapper/internal/httpclient"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
)

// Service handles authentication operations
type Service struct {
	httpClient *httpclient.Client
	logger     *slog.Logger
}

// NewService creates a new auth service
func NewService(httpClient *httpclient.Client, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		httpClient: httpClient,
		logger:     logger,
	}
}

// GetAuthToken retrieves an authentication token
func (s *Service) GetAuthToken(ctx context.Context, username, password string) (string, error) {
	authPayload := map[string]interface{}{
		"client_id":  "node-red-admin",
		"grant_type": "password",
		"scope":      "*",
		"username":   username,
		"password":   password,
	}

	jsonData, err := json.Marshal(authPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	url := fmt.Sprintf("%s/auth/token", s.httpClient.BaseURL())
	resp, err := s.httpClient.DoRequest(ctx, types.MethodPost, url, jsonData)
	if err != nil {
		return "", err
	}
	defer s.httpClient.CloseResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("authentication failed: status %d, failed to read error body: %w", resp.StatusCode, err)
		}
		return "", fmt.Errorf("authentication failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var authResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	if authResponse.AccessToken == "" {
		return "", fmt.Errorf("no access token in auth response")
	}

	return authResponse.AccessToken, nil
}
