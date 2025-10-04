// Package envelope provides the core message structure for GOX agent communication.
//
// The envelope system wraps all inter-agent messages with rich metadata for
// routing, tracing, quality of service, and debugging. This enables reliable
// message delivery, distributed tracing, and comprehensive observability
// across agent pipelines.
//
// Key Features:
// - Unique message identification and correlation tracking
// - Distributed tracing with hop counting and route history
// - Quality of service controls (priority, persistence, TTL)
// - Extensible headers and properties for custom metadata
// - Request/response pattern support with correlation IDs
//
// Called by: All agents, broker service, client libraries
// Calls: Standard JSON marshaling, UUID generation
package envelope

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Envelope wraps all messages with metadata for routing, tracing, and debugging.
//
// This structure provides the fundamental message format for all communication
// between GOX agents. It includes comprehensive metadata for:
// - Message routing and delivery guarantees
// - Distributed tracing and observability
// - Quality of service controls
// - Request/response correlation
// - Custom application metadata
//
// Thread Safety: Envelope instances are designed to be immutable after creation,
// with mutation methods creating new instances or updating specific fields atomically.
type Envelope struct {
	// Core identification fields for message tracking
	ID            string `json:"id"`                       // Unique message ID (UUID)
	CorrelationID string `json:"correlation_id,omitempty"` // Links request/response pairs

	// Routing information for message delivery
	Source      string `json:"source"`       // Source agent ID (e.g., "file-ingester-001")
	Destination string `json:"destination"`  // Target topic/pipe (e.g., "pub:new-file")
	MessageType string `json:"message_type"` // Semantic message type (e.g., "file_processed")

	// Timing and sequencing for delivery control
	Timestamp time.Time `json:"timestamp"`          // Message creation timestamp
	TTL       int64     `json:"ttl,omitempty"`      // Time-to-live in seconds (0=no expiry)
	Sequence  int64     `json:"sequence,omitempty"` // Message sequence number for ordering

	// Content and metadata payload
	Payload    json.RawMessage        `json:"payload"`              // Actual message data (JSON)
	Headers    map[string]string      `json:"headers,omitempty"`    // String-only headers
	Properties map[string]interface{} `json:"properties,omitempty"` // Typed properties

	// Distributed tracing and debugging information
	TraceID  string   `json:"trace_id,omitempty"`  // Distributed tracing ID (spans multiple messages)
	SpanID   string   `json:"span_id,omitempty"`   // Current span ID within trace
	HopCount int      `json:"hop_count,omitempty"` // Number of agents that processed this message
	Route    []string `json:"route,omitempty"`     // Agent IDs in processing order

	// Quality of service controls
	Priority   int  `json:"priority,omitempty"`   // Message priority (0-9, 9=highest priority)
	Persistent bool `json:"persistent,omitempty"` // True if message should survive broker restart
}

// NewEnvelope creates a new envelope with basic required fields.
//
// This is the primary constructor for creating envelopes in agent communication.
// It automatically generates a unique ID, sets the timestamp, and marshals the
// payload to JSON for transport.
//
// Parameters:
//   - source: Agent ID sending the message (e.g., "text-extractor-001")
//   - destination: Target topic or pipe (e.g., "pub:extracted-text")
//   - messageType: Semantic message type (e.g., "text_extracted")
//   - payload: Message data to be JSON-marshaled
//
// Returns:
//   - *Envelope: Ready-to-send envelope with all required fields populated
//   - error: JSON marshaling error if payload is not serializable
//
// Called by: All agents when sending messages
// Calls: json.Marshal(), uuid.New(), time.Now()
func NewEnvelope(source, destination, messageType string, payload interface{}) (*Envelope, error) {
	// Marshal payload to JSON for transport
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Envelope{
		ID:          uuid.New().String(), // Generate unique message ID
		Source:      source,
		Destination: destination,
		MessageType: messageType,
		Timestamp:   time.Now(),                   // Record creation time
		Payload:     payloadBytes,                 // JSON-encoded payload
		Headers:     make(map[string]string),      // Initialize empty headers map
		Properties:  make(map[string]interface{}), // Initialize empty properties map
		Route:       make([]string, 0),            // Initialize empty route history
	}, nil
}

// NewReplyEnvelope creates a reply envelope for request/response patterns.
//
// Constructs a response envelope that links back to the original request through
// correlation ID and maintains distributed tracing context. The reply is
// automatically routed back to the original sender.
//
// Parameters:
//   - originalEnvelope: The request envelope being replied to
//   - source: Agent ID sending the reply (e.g., "text-extractor-001")
//   - payload: Reply data to be JSON-marshaled
//
// Returns:
//   - *Envelope: Reply envelope with correlation ID and tracing context
//   - error: JSON marshaling error if payload is not serializable
//
// Called by: Agents implementing request/response patterns
// Calls: NewEnvelope(), uuid.New()
func NewReplyEnvelope(originalEnvelope *Envelope, source string, payload interface{}) (*Envelope, error) {
	// Create reply envelope with original sender as destination
	replyEnvelope, err := NewEnvelope(source, originalEnvelope.Source, "reply", payload)
	if err != nil {
		return nil, err
	}

	// Link reply to original request for correlation
	replyEnvelope.CorrelationID = originalEnvelope.ID
	// Maintain distributed tracing context
	replyEnvelope.TraceID = originalEnvelope.TraceID
	replyEnvelope.SpanID = uuid.New().String() // New span for reply processing

	return replyEnvelope, nil
}

// AddHop records that this message was processed by an agent.
//
// Updates the message's route history and hop count for distributed tracing
// and debugging. This enables tracking the complete processing path through
// the agent pipeline.
//
// Parameters:
//   - agentID: ID of the agent that processed this message
//
// Called by: Agent framework during message processing
// Calls: None (modifies envelope in-place)
func (e *Envelope) AddHop(agentID string) {
	e.HopCount++                       // Increment total hop count
	e.Route = append(e.Route, agentID) // Add agent to processing history
}

// SetHeader sets a custom header
func (e *Envelope) SetHeader(key, value string) {
	if e.Headers == nil {
		e.Headers = make(map[string]string)
	}
	e.Headers[key] = value
}

// GetHeader retrieves a custom header
func (e *Envelope) GetHeader(key string) (string, bool) {
	if e.Headers == nil {
		return "", false
	}
	value, exists := e.Headers[key]
	return value, exists
}

// SetProperty sets a custom property
func (e *Envelope) SetProperty(key string, value interface{}) {
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}
	e.Properties[key] = value
}

// GetProperty retrieves a custom property
func (e *Envelope) GetProperty(key string) (interface{}, bool) {
	if e.Properties == nil {
		return nil, false
	}
	value, exists := e.Properties[key]
	return value, exists
}

// UnmarshalPayload unmarshals the payload into the provided struct
func (e *Envelope) UnmarshalPayload(v interface{}) error {
	return json.Unmarshal(e.Payload, v)
}

// IsExpired checks if the message has exceeded its TTL
func (e *Envelope) IsExpired() bool {
	if e.TTL <= 0 {
		return false
	}
	return time.Now().Unix() > e.Timestamp.Unix()+e.TTL
}

// Clone creates a deep copy of the envelope
func (e *Envelope) Clone() *Envelope {
	clone := *e

	// Deep copy maps
	if e.Headers != nil {
		clone.Headers = make(map[string]string)
		for k, v := range e.Headers {
			clone.Headers[k] = v
		}
	}

	if e.Properties != nil {
		clone.Properties = make(map[string]interface{})
		for k, v := range e.Properties {
			clone.Properties[k] = v
		}
	}

	// Deep copy route
	if e.Route != nil {
		clone.Route = make([]string, len(e.Route))
		copy(clone.Route, e.Route)
	}

	// Copy payload
	if e.Payload != nil {
		clone.Payload = make(json.RawMessage, len(e.Payload))
		copy(clone.Payload, e.Payload)
	}

	return &clone
}

// ToJSON serializes the envelope to JSON
func (e *Envelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes an envelope from JSON
func FromJSON(data []byte) (*Envelope, error) {
	var envelope Envelope
	err := json.Unmarshal(data, &envelope)
	return &envelope, err
}

// MessageSize returns the approximate size of the envelope in bytes
func (e *Envelope) MessageSize() int {
	data, err := e.ToJSON()
	if err != nil {
		return 0
	}
	return len(data)
}

// Validate checks if the envelope has all required fields
func (e *Envelope) Validate() error {
	if e.ID == "" {
		return &ValidationError{Field: "id", Message: "envelope ID is required"}
	}
	if e.Source == "" {
		return &ValidationError{Field: "source", Message: "source agent ID is required"}
	}
	if e.Destination == "" {
		return &ValidationError{Field: "destination", Message: "destination is required"}
	}
	if e.MessageType == "" {
		return &ValidationError{Field: "message_type", Message: "message type is required"}
	}
	if e.Payload == nil {
		return &ValidationError{Field: "payload", Message: "payload is required"}
	}
	return nil
}

// ValidationError represents an envelope validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
