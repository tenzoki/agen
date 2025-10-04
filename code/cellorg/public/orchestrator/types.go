// Package orchestrator provides a public API for embedding Gox in other Go applications.
//
// This package allows applications like Alfa to run Gox cells in-process, managing
// agent lifecycles, VFS isolation per project, and event-driven communication between
// the host application and Gox agents.
//
// Key Features:
// - Embedded orchestrator: Run Gox cells in the same process
// - Event bridge: Subscribe to Gox topics via Go channels
// - VFS per project: Automatic file pattern isolation
// - Cell-based integration: Work with complete functional units, not individual agents
//
// Example usage:
//
//	orch := orchestrator.NewEmbedded(orchestrator.Config{
//	    ConfigPath: "/etc/gox",
//	})
//
//	// Start a cell for a project
//	orch.StartCell("rag:knowledge-backend", orchestrator.CellOptions{
//	    ProjectID: "project-a",
//	    VFSRoot:   "/Users/kai/workspace/project-a",
//	})
//
//	// Subscribe to events
//	events := orch.Subscribe("project-a:index-updated")
//	for event := range events {
//	    log.Printf("Index updated: %v", event.Data)
//	}
//
//	// Query RAG synchronously
//	result, _ := orch.PublishAndWait(
//	    "project-a:rag-queries",
//	    "project-a:rag-results",
//	    queryData,
//	    5*time.Second,
//	)
package orchestrator

import (
	"fmt"
	"time"
)

// Config defines configuration for the embedded orchestrator
type Config struct {
	// ConfigPath is the directory containing gox.yaml, pool.yaml, cells.yaml
	ConfigPath string

	// DefaultDataRoot is the default VFS root (can be overridden per cell)
	DefaultDataRoot string

	// Debug enables debug logging
	Debug bool

	// SupportPort is the support service port (default: 9000)
	SupportPort string

	// BrokerPort is the broker service port (default: 9001)
	BrokerPort string
}

// CellOptions defines options for starting a cell
type CellOptions struct {
	// ProjectID is the unique identifier for this project/cell instance
	ProjectID string

	// VFSRoot is the root directory for VFS (all file: patterns relative to this)
	VFSRoot string

	// Environment contains environment variable overrides for this cell
	Environment map[string]string

	// Config contains cell configuration overrides
	Config map[string]interface{}

	// ReadOnly indicates if VFS should be read-only
	ReadOnly bool
}

// Event represents an event from Gox cells
type Event struct {
	// Topic is the broker topic this event was published to
	Topic string

	// ProjectID is the project this event relates to (extracted from topic)
	ProjectID string

	// Data is the event payload
	Data map[string]interface{}

	// Timestamp is when the event was created
	Timestamp time.Time

	// Source is the agent ID that generated this event
	Source string

	// TraceID for distributed tracing
	TraceID string
}

// CellInfo provides information about a running cell
type CellInfo struct {
	// CellID is the cell identifier (e.g., "rag:knowledge-backend")
	CellID string

	// ProjectID is the associated project
	ProjectID string

	// VFSRoot is the VFS root path
	VFSRoot string

	// Status is the cell status
	Status CellStatus

	// Agents lists the agents in this cell
	Agents []AgentInfo

	// StartedAt is when the cell started
	StartedAt time.Time
}

// CellStatus represents the status of a cell
type CellStatus string

const (
	// CellStatusStarting indicates cell is starting up
	CellStatusStarting CellStatus = "starting"

	// CellStatusRunning indicates cell is running
	CellStatusRunning CellStatus = "running"

	// CellStatusStopping indicates cell is shutting down
	CellStatusStopping CellStatus = "stopping"

	// CellStatusStopped indicates cell is stopped
	CellStatusStopped CellStatus = "stopped"

	// CellStatusError indicates cell encountered an error
	CellStatusError CellStatus = "error"
)

// AgentInfo provides information about an agent in a cell
type AgentInfo struct {
	// ID is the agent instance ID (e.g., "rag-agent-001")
	ID string

	// Type is the agent type (e.g., "rag-agent")
	Type string

	// Status is the agent status
	Status string

	// Ingress is the agent's ingress pattern
	Ingress string

	// Egress is the agent's egress pattern
	Egress string
}

// QueryRequest represents a synchronous query request
type QueryRequest struct {
	// RequestID uniquely identifies this request
	RequestID string

	// Topic to publish the request to
	Topic string

	// Data is the request payload
	Data interface{}

	// Timeout for waiting for response
	Timeout time.Duration
}

// QueryResponse represents the response to a QueryRequest
type QueryResponse struct {
	// RequestID matches the request
	RequestID string

	// Data is the response payload
	Data map[string]interface{}

	// Error contains error message if request failed
	Error string

	// Duration is how long the request took
	Duration time.Duration
}

// ApplyDefaults applies default values to Config
func ApplyDefaults(cfg Config) Config {
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = "./config"
	}
	if cfg.DefaultDataRoot == "" {
		cfg.DefaultDataRoot = "/var/lib/gox"
	}
	if cfg.SupportPort == "" {
		cfg.SupportPort = ":9000"
	}
	if cfg.BrokerPort == "" {
		cfg.BrokerPort = ":9001"
	}
	return cfg
}

// ValidateCellOptions validates CellOptions
func ValidateCellOptions(opts CellOptions) error {
	if opts.ProjectID == "" {
		return fmt.Errorf("ProjectID is required")
	}
	return nil
}
