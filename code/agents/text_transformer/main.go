// Package main provides the text transformer agent for the GOX framework.
//
// The text transformer is a processing agent that performs text transformations
// on messages flowing through the pipeline. It demonstrates the core pattern
// for content processing agents: receive message, transform payload, preserve
// metadata, and forward the result.
//
// Key Features:
// - Text content transformation (uppercase conversion)
// - Metadata preservation and augmentation
// - Processing timestamp tracking
// - Flexible payload handling (string or byte array)
// - Debug logging for message flow analysis
//
// Transformations Applied:
// - Converts all text to uppercase
// - Adds processing metadata and timestamps
// - Calculates and logs character count changes
// - Preserves original message metadata
//
// Configuration (via cells.yaml):
// - ingress: "sub:topic-name" - Topic to subscribe to for messages
// - egress: "pipe:target" or "pub:topic" - Destination for transformed messages
// - transformation: Configuration key for transformation type
// - add_metadata: Enable/disable metadata augmentation
//
// Called by: GOX orchestrator during pipeline processing
// Calls: String transformation functions, metadata handling
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// TextTransformer implements the AgentRunner interface for text processing.
//
// This agent processes text content from incoming messages, applies configured
// transformations, and forwards the results with enhanced metadata. It serves
// as a template for content processing agents that need to modify message
// payloads while preserving the message flow.
//
// Thread Safety: The agent framework handles concurrency and message ordering
type TextTransformer struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
}

// ProcessMessage handles text transformation with metadata preservation.
//
// This is the core business logic for the text transformer. It extracts text
// content from the message payload, applies uppercase transformation, adds
// processing metadata, and returns a new message for downstream processing.
//
// Processing Steps:
// 1. Extract and log incoming message metadata for debugging
// 2. Parse payload content (supports string and byte array formats)
// 3. Apply text transformation (uppercase conversion)
// 4. Add processing annotations and timestamps
// 5. Create new message with preserved and enhanced metadata
// 6. Log transformation results for monitoring
//
// Parameters:
//   - msg: BrokerMessage containing text payload to transform
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Transformed message with enhanced metadata
//   - error: Processing error (always nil for successful transformation)
//
// Called by: GOX agent framework during message processing
// Calls: String transformation functions, timestamp generation, logging
func (t *TextTransformer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Debug logging: Record received message metadata for pipeline analysis
	base.LogDebug("TextTransformer received message %s with meta: %+v", msg.ID, msg.Meta)
	for k, v := range msg.Meta {
		base.LogDebug("  Received Meta[%s] = %+v (type: %T)", k, v, v)
	}

	// Extract text content from payload with flexible type handling
	var text string
	switch payload := msg.Payload.(type) {
	case string:
		text = payload // Direct string payload
	case []byte:
		text = string(payload) // Convert byte array to string
	default:
		// Unsupported payload type - log error and skip processing
		base.LogError("Unsupported payload type: %T", payload)
		return nil, nil // Returning nil skips this message
	}

	// Apply text transformation: uppercase conversion with processing annotations
	transformedText := strings.ToUpper(text)
	transformedText += "\n\n--- PROCESSED BY GOX V3 TEXT TRANSFORMER ---\n"
	transformedText += "Processed at: " + time.Now().Format(time.RFC3339) + "\n"
	transformedText += fmt.Sprintf("Original length: %d characters\n", len(text))

	// Create new message with transformed content
	transformedMsg := &client.BrokerMessage{
		ID:      msg.ID + "_transformed",      // Append suffix to track transformation
		Type:    "transformed_text",           // Update message type
		Payload: transformedText,              // Set transformed content
		Meta:    make(map[string]interface{}), // Initialize metadata map
	}

	// Preserve original metadata while adding transformation information
	for k, v := range msg.Meta {
		transformedMsg.Meta[k] = v // Copy all original metadata
	}
	// Add transformation-specific metadata
	transformedMsg.Meta["transformed_at"] = time.Now()
	transformedMsg.Meta["transformer"] = "text-transformer"
	transformedMsg.Meta["transformation"] = "uppercase"

	// Log transformation results for monitoring and debugging
	base.LogDebug("Text length: %d -> %d characters", len(text), len(transformedText))

	// Debug logging: Record outgoing message metadata
	base.LogDebug("TextTransformer sending message %s with meta: %+v", transformedMsg.ID, transformedMsg.Meta)
	for k, v := range transformedMsg.Meta {
		base.LogDebug("  Sending Meta[%s] = %+v (type: %T)", k, v, v)
	}

	return transformedMsg, nil
}

func main() {
	// ALL boilerplate eliminated - framework handles everything!
	if err := agent.Run(&TextTransformer{}, "text-transformer"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
