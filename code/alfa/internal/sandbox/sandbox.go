package sandbox

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Sandbox defines the interface for sandboxed code execution
type Sandbox interface {
	// Execute runs a command in the sandbox
	Execute(ctx context.Context, cmd ExecuteRequest) (*ExecuteResult, error)

	// IsAvailable checks if sandbox is available (Docker installed)
	IsAvailable() bool

	// Cleanup removes containers and resources
	Cleanup() error
}

// ExecuteRequest defines parameters for sandboxed execution
type ExecuteRequest struct {
	Command    string            // Command to execute
	WorkDir    string            // Working directory (host path to mount)
	Env        map[string]string // Environment variables
	Timeout    time.Duration     // Execution timeout
	CPULimit   float64           // CPU limit (e.g., 1.0 = 1 core)
	MemoryMB   int64             // Memory limit in MB
	Image      string            // Docker image to use
	NetworkOff bool              // Disable network access
}

// ExecuteResult contains the output and status of execution
type ExecuteResult struct {
	Stdout   string        // Standard output
	Stderr   string        // Standard error
	ExitCode int           // Exit code
	Duration time.Duration // Execution time
	Error    error         // Execution error (timeout, etc.)
}

// Config holds sandbox configuration
type Config struct {
	DefaultImage    string        // Default Docker image
	DefaultTimeout  time.Duration // Default timeout
	DefaultCPULimit float64       // Default CPU limit
	DefaultMemoryMB int64         // Default memory limit MB
	NetworkOff      bool          // Disable network by default
	AutoCleanup     bool          // Auto-remove containers after execution
}

// DefaultConfig returns sensible defaults for sandbox
func DefaultConfig() Config {
	return Config{
		DefaultImage:    "golang:1.24-alpine",
		DefaultTimeout:  30 * time.Second,
		DefaultCPULimit: 1.0,
		DefaultMemoryMB: 512,
		NetworkOff:      true,
		AutoCleanup:     true,
	}
}

// Error represents a sandbox error
type Error struct {
	Operation string // Operation that failed
	Message   string // Error message
	Err       error  // Underlying error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("sandbox %s: %s (%v)", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("sandbox %s: %s", e.Operation, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// OutputWriter allows streaming output during execution
type OutputWriter interface {
	io.Writer
}
