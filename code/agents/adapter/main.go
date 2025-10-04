// Package main provides the adapter agent for the GOX framework.
//
// The adapter agent performs schema mapping and data transformation services
// for other agents in the pipeline. It acts as a data conversion hub that can
// transform between different formats (JSON, CSV, XML, etc.) and apply various
// text transformations.
//
// Key Features:
// - Schema mapping between different data formats
// - Text transformations (uppercase, lowercase, trim)
// - JSON pretty-printing and compacting
// - CSV to JSON conversion and vice versa
// - Base64 encoding/decoding
// - Extensible adapter function registry
//
// Operation:
// The agent receives AdapterRequest messages containing source format,
// target format, and data to transform. It applies the appropriate
// transformation function and returns an AdapterResponse with the
// converted data or error information.
//
// Configuration (via cells.yaml):
// - ingress: "sub:transform-requests" - Topic for transformation requests
// - egress: "pub:transform-responses" - Topic for transformation responses
// - supported_formats: Array of supported format combinations
// - max_request_size: Maximum size for transformation requests
//
// Called by: Other agents requiring data format conversion
// Calls: Internal transformation functions, JSON/CSV processors
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// Adapter implements schema mapping and data transformation services.
//
// This agent provides a central hub for data format conversions between
// different agents in the pipeline. It maintains a registry of transformation
// functions that can be applied based on source and target format specifications.
//
// Thread Safety: The agent framework handles concurrency and message ordering
type Adapter struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
	adapters                 map[string]AdapterFunc
}

// AdapterFunc defines the signature for transformation functions
type AdapterFunc func(input []byte) ([]byte, error)

// AdapterRequest represents a transformation request from another agent
type AdapterRequest struct {
	RequestID    string `json:"request_id"`    // Unique identifier for tracking
	SourceFormat string `json:"source_format"` // Source data format (e.g., "csv", "json")
	TargetFormat string `json:"target_format"` // Target data format (e.g., "json", "csv")
	Data         []byte `json:"data"`          // Raw data to transform
	ReplyTo      string `json:"reply_to,omitempty"` // Optional reply destination
}

// AdapterResponse represents the result of a transformation operation
type AdapterResponse struct {
	RequestID       string `json:"request_id"`                 // Original request identifier
	Success         bool   `json:"success"`                    // Transformation success status
	TransformedData []byte `json:"transformed_data,omitempty"` // Converted data (on success)
	Error           string `json:"error,omitempty"`            // Error message (on failure)
}

// Init initializes the adapter agent with all available transformation functions.
//
// This method is called once during agent startup after BaseAgent initialization.
// It sets up the adapter function registry and registers all built-in
// transformation capabilities for schema mapping and data conversion.
//
// Parameters:
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - error: Initialization error (always nil for this implementation)
//
// Called by: GOX agent framework during startup
// Calls: registerAdapters() to populate transformation registry
func (a *Adapter) Init(base *agent.BaseAgent) error {
	a.adapters = make(map[string]AdapterFunc)
	a.registerAdapters()
	base.LogInfo("Adapter initialized with %d transformation functions", len(a.adapters))
	return nil
}

// ProcessMessage performs schema mapping and data transformation operations.
//
// This is the core business logic for the adapter agent. It receives transformation
// requests, applies the appropriate conversion function based on source and target
// formats, and returns the transformed data or error information.
//
// Processing Steps:
// 1. Parse AdapterRequest from message payload
// 2. Create transformation key from source and target formats
// 3. Look up appropriate transformation function
// 4. Apply transformation and handle errors
// 5. Create response message with results
//
// Parameters:
//   - msg: BrokerMessage containing AdapterRequest in payload
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Response message with transformation results
//   - error: Always nil (errors are returned in response message)
//
// Called by: GOX agent framework during message processing
// Calls: Transformation functions, response creation methods
func (a *Adapter) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Parse AdapterRequest from message payload
	var request AdapterRequest
	var payload []byte

	// Handle different payload types
	switch p := msg.Payload.(type) {
	case []byte:
		payload = p
	case string:
		payload = []byte(p)
	default:
		return a.createErrorResponse("unknown", "Invalid payload type", base), nil
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		base.LogError("Failed to parse adapter request: %v", err)
		return a.createErrorResponse("unknown", "Invalid request format: "+err.Error(), base), nil
	}

	base.LogDebug("Processing adapter request %s: %s -> %s",
		request.RequestID, request.SourceFormat, request.TargetFormat)

	// Create transformation key for schema mapping
	transformKey := fmt.Sprintf("%s-to-%s", request.SourceFormat, request.TargetFormat)

	// Look up and apply transformation function
	if adapterFunc, exists := a.adapters[transformKey]; exists {
		transformedData, err := adapterFunc(request.Data)
		if err != nil {
			base.LogError("Transformation failed for %s: %v", transformKey, err)
			return a.createErrorResponse(request.RequestID, err.Error(), base), nil
		}
		base.LogInfo("Successfully transformed data using %s", transformKey)
		return a.createSuccessResponse(request.RequestID, transformedData, base), nil
	}

	// Fallback: try direct format name lookup
	if adapterFunc, exists := a.adapters[request.TargetFormat]; exists {
		transformedData, err := adapterFunc(request.Data)
		if err != nil {
			base.LogError("Transformation failed for %s: %v", request.TargetFormat, err)
			return a.createErrorResponse(request.RequestID, err.Error(), base), nil
		}
		base.LogInfo("Successfully transformed data using %s", request.TargetFormat)
		return a.createSuccessResponse(request.RequestID, transformedData, base), nil
	}

	// No suitable transformation found
	errorMsg := fmt.Sprintf("No adapter found for %s (available: %v)",
		transformKey, a.getAvailableAdapters())
	base.LogError("Transformation request failed: %s", errorMsg)
	return a.createErrorResponse(request.RequestID, errorMsg, base), nil
}

// registerAdapters registers all available schema mapping and transformation functions.
//
// This method populates the adapter registry with built-in transformation
// capabilities including text processing, JSON formatting, format conversion,
// and encoding operations. Additional adapters can be registered here.
//
// Called by: Init() during agent startup
// Calls: Individual transformation function implementations
func (a *Adapter) registerAdapters() {
	// Text transformations
	a.adapters["text-to-upper"] = func(input []byte) ([]byte, error) {
		return []byte(strings.ToUpper(string(input))), nil
	}
	a.adapters["text-to-lower"] = func(input []byte) ([]byte, error) {
		return []byte(strings.ToLower(string(input))), nil
	}
	a.adapters["text-to-trim"] = func(input []byte) ([]byte, error) {
		return []byte(strings.TrimSpace(string(input))), nil
	}

	// JSON schema mapping and formatting
	a.adapters["json-to-pretty"] = func(input []byte) ([]byte, error) {
		var obj interface{}
		if err := json.Unmarshal(input, &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return json.MarshalIndent(obj, "", "  ")
	}
	a.adapters["json-to-compact"] = func(input []byte) ([]byte, error) {
		var obj interface{}
		if err := json.Unmarshal(input, &obj); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return json.Marshal(obj)
	}

	// Cross-format conversions
	a.adapters["csv-to-json"] = a.csvToJSON
	a.adapters["json-to-csv"] = a.jsonToCSV

	// Base64 encoding (simplified implementation)
	a.adapters["data-to-base64"] = func(input []byte) ([]byte, error) {
		// Simplified base64-like encoding for demo
		encoded := fmt.Sprintf("base64:%x", input)
		return []byte(encoded), nil
	}
	a.adapters["base64-to-data"] = func(input []byte) ([]byte, error) {
		// Simplified base64-like decoding for demo
		inputStr := string(input)
		if !strings.HasPrefix(inputStr, "base64:") {
			return nil, fmt.Errorf("invalid base64 format")
		}
		hexStr := inputStr[7:] // Remove "base64:" prefix
		decoded := make([]byte, len(hexStr)/2)
		for i := 0; i < len(decoded); i++ {
			fmt.Sscanf(hexStr[i*2:i*2+2], "%02x", &decoded[i])
		}
		return decoded, nil
	}
}

// getAvailableAdapters returns a list of all registered transformation functions
func (a *Adapter) getAvailableAdapters() []string {
	adapters := make([]string, 0, len(a.adapters))
	for name := range a.adapters {
		adapters = append(adapters, name)
	}
	return adapters
}

// createSuccessResponse creates a successful transformation response message
func (a *Adapter) createSuccessResponse(requestID string, data []byte, base *agent.BaseAgent) *client.BrokerMessage {
	response := AdapterResponse{
		RequestID:       requestID,
		Success:         true,
		TransformedData: data,
	}
	responseBytes, _ := json.Marshal(response)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("adapter_success_%s", requestID),
		Type:    "adapter_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"adapter_success":    true,
			"original_request":   requestID,
			"transformation_at":  "adapter-agent",
			"data_size":         len(data),
		},
	}
}

// createErrorResponse creates an error response message for failed transformations
func (a *Adapter) createErrorResponse(requestID, errorMsg string, base *agent.BaseAgent) *client.BrokerMessage {
	response := AdapterResponse{
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
	responseBytes, _ := json.Marshal(response)

	base.LogError("Adapter error for request %s: %s", requestID, errorMsg)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("adapter_error_%s", requestID),
		Type:    "adapter_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"adapter_success":   false,
			"original_request":  requestID,
			"error_message":     errorMsg,
		},
	}
}

// csvToJSON performs simplified CSV to JSON conversion
func (a *Adapter) csvToJSON(input []byte) ([]byte, error) {
	// Simplified CSV to JSON conversion for demonstration
	lines := strings.Split(string(input), "\n")
	if len(lines) == 0 {
		return []byte("[]"), nil
	}

	// Use first line as headers
	headers := strings.Split(lines[0], ",")
	var records []map[string]string

	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		values := strings.Split(lines[i], ",")
		record := make(map[string]string)

		for j, header := range headers {
			if j < len(values) {
				record[strings.TrimSpace(header)] = strings.TrimSpace(values[j])
			}
		}
		records = append(records, record)
	}

	return json.Marshal(records)
}

// jsonToCSV performs simplified JSON to CSV conversion
func (a *Adapter) jsonToCSV(input []byte) ([]byte, error) {
	// Simplified JSON to CSV conversion for demonstration
	var records []map[string]interface{}
	if err := json.Unmarshal(input, &records); err != nil {
		return nil, fmt.Errorf("invalid JSON array: %w", err)
	}

	if len(records) == 0 {
		return []byte(""), nil
	}

	// Extract headers from first record
	var headers []string
	for key := range records[0] {
		headers = append(headers, key)
	}

	// Build CSV
	var result strings.Builder
	result.WriteString(strings.Join(headers, ",") + "\n")

	for _, record := range records {
		var values []string
		for _, header := range headers {
			if val, exists := record[header]; exists {
				values = append(values, fmt.Sprintf("%v", val))
			} else {
				values = append(values, "")
			}
		}
		result.WriteString(strings.Join(values, ",") + "\n")
	}

	return []byte(result.String()), nil
}

// main is the entry point for the adapter agent.
//
// Initializes and starts the agent using the GOX framework's agent runner.
// The framework handles all boilerplate including broker connection, message
// routing, lifecycle management, and transformation request processing.
//
// The agent ID "adapter" is used for:
// - Agent registration with the support service
// - Logging and debugging identification
// - Message routing and correlation
//
// Called by: Operating system process execution
// Calls: agent.Run() with Adapter implementation
func main() {
	// Framework handles all boilerplate including broker connection, message
	// routing, transformation processing, and lifecycle management
	if err := agent.Run(&Adapter{}, "adapter"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}