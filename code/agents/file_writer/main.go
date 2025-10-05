// Package main provides the file writer agent for the GOX framework.
//
// The file writer is a terminal agent that persists processed data to the
// filesystem. It serves as the output endpoint for processing pipelines,
// converting message payloads into files with configurable naming patterns
// and directory structures.
//
// Key Features:
// - Automatic file writing via framework's FileEgressHandler
// - Configurable output paths with template variables
// - Directory creation with proper permissions
// - Metadata preservation in output files
// - Multiple output formats (text, JSON, binary)
// - Template-based filename generation
//
// Operation:
// The agent leverages the framework's built-in FileEgressHandler which
// handles the actual file writing operations. This agent simply signals
// that messages should be written by returning nil, triggering the
// framework's egress handler.
//
// Configuration (via cells.yaml):
// - ingress: "sub:topic-name" or "pipe:source" - Message source
// - egress: "file:path/template_{{.var}}.ext" - Output file pattern
// - create_directories: Enable/disable automatic directory creation
// - output_format: File format (txt, json, binary)
// - preserve_metadata: Include metadata in output files
//
// Template Variables Available:
// - {{.timestamp}}: Current timestamp in various formats
// - {{.filename}}: Original filename (if available)
// - {{.id}}: Message ID
// - Custom variables from message metadata
//
// Called by: GOX orchestrator as pipeline terminus
// Calls: Framework FileEgressHandler, file system operations
package main

import (
	"fmt"
	"os"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// FileWriter implements the AgentRunner interface for file output operations.
//
// This agent acts as a pipeline terminus that triggers file writing through
// the framework's FileEgressHandler. It provides a lightweight interface
// for converting processed messages into persistent file outputs with
// configurable naming and formatting.
//
// Thread Safety: The agent framework handles concurrency and file locking
type FileWriter struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
}

// ProcessMessage triggers file writing via the framework's FileEgressHandler.
//
// This method serves as a passthrough that signals the framework to write
// the message payload to a file. By returning nil, it indicates that the
// original message should be processed by the egress handler, which performs
// the actual file writing based on the egress configuration.
//
// Processing Flow:
// 1. Log incoming message metadata for debugging
// 2. Return nil to trigger framework's FileEgressHandler
// 3. FileEgressHandler writes payload to configured file path
// 4. Framework handles directory creation, templating, and error handling
//
// Parameters:
//   - msg: BrokerMessage containing payload to write to file
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Always nil to trigger egress handler
//   - error: Always nil for this passthrough operation
//
// Called by: GOX agent framework during message processing
// Calls: base.LogDebug() for operation logging, framework egress handler
func (f *FileWriter) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// For file_writer, the actual file writing is handled by the FileEgressHandler
	// This agent doesn't transform the message, just signals that it should be written
	// by returning nil (which triggers the framework to call the egress handler)

	base.LogDebug("FileWriter processing message %s with meta: %+v", msg.ID, msg.Meta)
	for k, v := range msg.Meta {
		base.LogDebug("  FileWriter Meta[%s] = %+v (type: %T)", k, v, v)
	}

	// Return nil to indicate the egress handler should process the original message
	// This triggers the framework's FileEgressHandler to write the payload to file
	return nil, nil
}

// main is the entry point for the file writer agent.
//
// Initializes and starts the agent using the GOX framework's agent runner.
// The framework handles all boilerplate including broker connection,
// message routing, lifecycle management, and the FileEgressHandler setup.
//
// The agent ID "file-writer" is used for:
// - Agent registration with the support service
// - Logging and debugging identification
// - Message routing and correlation
//
// Called by: Operating system process execution
// Calls: agent.Run() with FileWriter implementation
func main() {
	// Framework handles all boilerplate including file writing, broker connection,
	// message routing, and lifecycle management
	if err := agent.Run(&FileWriter{}, "file-writer"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
