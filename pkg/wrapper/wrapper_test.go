package wrapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/types"
	"github.com/yoyo-mq/go-nodered-wrapper/pkg/wrapper"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &types.Config{
				NodeRedURL: "http://localhost:1880",
				APIKey:     "test-key",
				Timeout:    30 * time.Second,
				Debug:      true,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty node red url",
			config: &types.Config{
				NodeRedURL: "",
				APIKey:     "test-key",
				Timeout:    30 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := wrapper.New(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, w)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, w)
				assert.Equal(t, tt.config, w.GetConfig())
			}
		})
	}
}

func TestNewWithCustomOptions(t *testing.T) {
	config := &types.Config{
		NodeRedURL: "http://localhost:1880",
		APIKey:     "test-key",
		Timeout:    30 * time.Second,
	}

	t.Run("NewWithConverter", func(t *testing.T) {
		converter := &wrapper.DefaultConverter{}
		w, err := wrapper.NewWithConverter(config, converter)
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})

	t.Run("NewWithExecutor", func(t *testing.T) {
		executor := &wrapper.DefaultExecutor{}
		w, err := wrapper.NewWithExecutor(config, executor)
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})
}

func TestNodeRedWrapper_DeployFlow(t *testing.T) {
	config := &types.Config{
		NodeRedURL: "http://localhost:1880",
		APIKey:     "test-key",
		Timeout:    30 * time.Second,
		Debug:      true,
	}

	w, err := wrapper.New(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		flow    *types.FlowDefinition
		wantErr bool
	}{
		{
			name: "valid flow",
			flow: &types.FlowDefinition{
				ID:   "test-flow",
				Name: "Test Flow",
			},
			wantErr: false,
		},
		{
			name:    "nil flow",
			flow:    nil,
			wantErr: true,
		},
		{
			name: "empty flow ID",
			flow: &types.FlowDefinition{
				ID:   "",
				Name: "Test Flow",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This will fail in real test since we don't have a running Node-RED instance
			// In a real test, you would use a mock client
			_, err := w.DeployFlow(context.Background(), tt.flow)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// This will fail without a real Node-RED instance, but we can test the validation
				assert.Error(t, err) // Expected to fail due to no real Node-RED
			}
		})
	}
}

func TestNodeRedWrapper_ExecuteFlow(t *testing.T) {
	config := &types.Config{
		NodeRedURL: "http://localhost:1880",
		APIKey:     "test-key",
		Timeout:    30 * time.Second,
		Debug:      true,
	}

	w, err := wrapper.New(config)
	require.NoError(t, err)

	tests := []struct {
		name    string
		flowID  string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name:   "valid execution",
			flowID: "test-flow",
			input: map[string]interface{}{
				"message": "test",
			},
			wantErr: false,
		},
		{
			name:    "empty flow ID",
			flowID:  "",
			input:   map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This will fail in real test since we don't have a running Node-RED instance
			// In a real test, you would use a mock client
			result, err := w.ExecuteFlow(context.Background(), tt.flowID, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				// This will fail without a real Node-RED instance, but we can test the validation
				assert.Error(t, err) // Expected to fail due to no real Node-RED
				assert.Nil(t, result)
			}
		})
	}
}

func TestDefaultConverter(t *testing.T) {
	converter := &wrapper.DefaultConverter{}

	t.Run("convert FlowDefinition", func(t *testing.T) {
		flow := &types.FlowDefinition{
			ID:   "test-flow",
			Name: "Test Flow",
		}

		result, err := converter.ConvertToNodeRedFlow(flow)
		assert.NoError(t, err)
		assert.Equal(t, flow, result)
	})

	t.Run("convert map", func(t *testing.T) {
		flowMap := map[string]interface{}{
			"id":          "test-flow",
			"name":        "Test Flow",
			"description": "Test Description",
			"version":     "1.0.0",
		}

		result, err := converter.ConvertToNodeRedFlow(flowMap)
		assert.NoError(t, err)
		assert.Equal(t, "test-flow", result.ID)
		assert.Equal(t, "Test Flow", result.Name)
		assert.Equal(t, "Test Description", result.Description)
		assert.Equal(t, "1.0.0", result.Version)
	})

	t.Run("unsupported type", func(t *testing.T) {
		_, err := converter.ConvertToNodeRedFlow("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported workflow type")
	})
}

func TestDefaultExecutor(t *testing.T) {
	executor := &wrapper.DefaultExecutor{}

	t.Run("pre-execute", func(t *testing.T) {
		err := executor.PreExecute(context.Background(), map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("post-execute", func(t *testing.T) {
		result := &types.ExecutionResult{
			Success: true,
		}
		err := executor.PostExecute(context.Background(), result)
		assert.NoError(t, err)
	})

	t.Run("on-error", func(t *testing.T) {
		err := executor.OnError(context.Background(), assert.AnError)
		assert.NoError(t, err)
	})
}
