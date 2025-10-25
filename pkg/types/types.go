package types

import (
	"time"
)

// FlowDefinition represents a Node-RED flow
type FlowDefinition struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Label       string                 `json:"label,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Description string                 `json:"description,omitempty"`
	Info        string                 `json:"info,omitempty"`
	Disabled    bool                   `json:"disabled"`
	Nodes       []Node                 `json:"nodes"`
	Connections []Connection           `json:"connections,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty"`
}

// Node represents a Node-RED node
type Node struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Position   Position               `json:"position"`
	Properties map[string]interface{} `json:"properties"`
	Wires      [][]string             `json:"wires"`
}

// Connection represents a connection between nodes
type Connection struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	SourcePort int    `json:"source_port,omitempty"`
	TargetPort int    `json:"target_port,omitempty"`
}

// Position represents node position in the flow editor
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ExecutionResult represents the result of a flow execution
type ExecutionResult struct {
	ExecutionID string                 `json:"execution_id"`
	Success     bool                   `json:"success"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Logs        []LogEntry             `json:"logs,omitempty"`
}

// LogEntry represents a log entry from flow execution
type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	NodeID  string    `json:"node_id,omitempty"`
}

// Config holds configuration for the Node-RED wrapper
type Config struct {
	NodeRedURL    string        `yaml:"node_red_url" json:"node_red_url"`
	APIKey        string        `yaml:"api_key" json:"api_key"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts" json:"retry_attempts"`
	Debug         bool          `yaml:"debug" json:"debug"`
}

// ExecutionOptions holds options for flow execution
type ExecutionOptions struct {
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	Async       bool          `yaml:"async" json:"async"`
	RetryPolicy *RetryPolicy  `yaml:"retry_policy" json:"retry_policy"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries    int           `yaml:"max_retries" json:"max_retries"`
	InitialDelay  time.Duration `yaml:"initial_delay" json:"initial_delay"`
	MaxDelay      time.Duration `yaml:"max_delay" json:"max_delay"`
	BackoffFactor float64       `yaml:"backoff_factor" json:"backoff_factor"`
}

// CircuitBreaker defines circuit breaker behavior
type CircuitBreaker struct {
	MaxFailures  int           `yaml:"max_failures" json:"max_failures"`
	Timeout      time.Duration `yaml:"timeout" json:"timeout"`
	ResetTimeout time.Duration `yaml:"reset_timeout" json:"reset_timeout"`
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status    string            `json:"status"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}
