// Package broker implements the central message broker for the GOX framework.
// The broker provides both publish/subscribe messaging for event distribution
// and point-to-point pipes for direct agent communication.
//
// Key Features:
// - Publish/Subscribe topics for event distribution across multiple agents
// - Point-to-point pipes for direct agent-to-agent communication
// - Envelope protocol support for message metadata and routing
// - JSON-RPC based protocol for client-broker communication
// - Thread-safe concurrent connection handling
// - Message history and buffering capabilities
//
// The broker serves as the central communication hub that connects all agents
// in the GOX orchestration system, enabling distributed processing workflows.
package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/envelope"
)

// Service represents the central broker service that handles all agent communication.
// It manages both publish/subscribe topics and point-to-point pipes, maintaining
// active connections to all registered agents in the GOX framework.
//
// The service operates as a TCP server listening for JSON-RPC requests from agents.
// Each agent establishes a persistent connection that is used for bidirectional
// communication throughout the agent's lifecycle.
type Service struct {
	// Network configuration
	port     string       // TCP port to listen on (e.g., ":9001")
	protocol string       // Network protocol ("tcp")
	codec    string       // Message encoding format ("json")
	debug    bool         // Enable debug logging
	listener net.Listener // TCP listener for incoming connections

	// Publish/Subscribe topics for event distribution
	// Topics allow multiple agents to receive the same message
	topics    map[string]*Topic // Map of topic name to Topic instance
	topicsMux sync.RWMutex      // Protects topics map from concurrent access

	// Point-to-point pipes for direct agent communication
	// Pipes provide reliable message delivery between specific agents
	pipes    map[string]*Pipe // Map of pipe name to Pipe instance
	pipesMux sync.RWMutex     // Protects pipes map from concurrent access

	// Active agent connections
	// Tracks all currently connected agents for routing and cleanup
	connections map[string]*Connection // Map of connection ID to Connection
	connMux     sync.RWMutex           // Protects connections map from concurrent access
}

// Topic represents a publish/subscribe channel where multiple agents can
// subscribe to receive all messages published to the topic. Topics support
// both simple Message objects and full Envelope protocol messages.
//
// Topics maintain message history for debugging and replay capabilities,
// with automatic cleanup when the buffer exceeds capacity.
type Topic struct {
	Name        string               // Unique topic identifier
	Subscribers []*Connection        // List of agents subscribed to this topic
	Messages    []*Message           // Recent message history (max 100)
	Envelopes   []*envelope.Envelope // Recent envelope history (max 100)
	mux         sync.RWMutex         // Protects topic data from concurrent access
}

// Pipe represents a point-to-point communication channel between two agents.
// Unlike topics, pipes provide direct message delivery with buffering and
// support both simple Message objects and full Envelope protocol messages.
//
// Pipes use Go channels for thread-safe message queuing and can buffer
// up to 100 messages before blocking or rejecting new messages.
type Pipe struct {
	Name      string                  // Unique pipe identifier
	Producer  *Connection             // Agent that sends messages to this pipe
	Consumer  *Connection             // Agent that receives messages from this pipe
	Messages  chan *Message           // Buffered channel for Message objects (capacity 100)
	Envelopes chan *envelope.Envelope // Buffered channel for Envelope objects (capacity 100)
	mux       sync.RWMutex            // Protects pipe metadata from concurrent access
}

// Connection represents an active agent connection to the broker.
// Each connection maintains its own JSON encoder/decoder for efficient
// message serialization and tracks agent metadata for routing.
//
// Connections are used for both control messages (JSON-RPC requests)
// and data delivery (topic publications, pipe messages).
type Connection struct {
	ID       string        // Unique connection identifier (generated)
	Conn     net.Conn      // Underlying TCP connection
	Encoder  *json.Encoder // JSON encoder for sending messages to agent
	Decoder  *json.Decoder // JSON decoder for receiving messages from agent
	AgentID  string        // Agent identifier provided during connection handshake
	LastSeen time.Time     // Timestamp of last received message (for health monitoring)
}

// Message represents a simple message object used for basic agent communication.
// Messages are used when full envelope protocol is not required, providing
// a lightweight alternative for simple data exchange between agents.
//
// The Target field is automatically set by the broker based on the routing
// destination (topic or pipe), and Timestamp is set when the message is processed.
type Message struct {
	ID        string                 `json:"id"`        // Unique message identifier
	Type      string                 `json:"type"`      // Message type for handling dispatch
	Target    string                 `json:"target"`    // Routing target (set by broker)
	Payload   interface{}            `json:"payload"`   // Message data (any JSON-serializable type)
	Meta      map[string]interface{} `json:"meta"`      // Metadata for message processing
	Timestamp time.Time              `json:"timestamp"` // When message was processed by broker
}

// BrokerRequest represents a JSON-RPC request from an agent to the broker.
// All agent-broker communication uses this standardized request format
// for method invocation and parameter passing.
//
// Supported methods: connect, publish, publish_envelope, subscribe,
// send_pipe, send_pipe_envelope, receive_pipe
type BrokerRequest struct {
	ID     string          `json:"id"`     // Request identifier for response correlation
	Method string          `json:"method"` // Broker method to invoke
	Params json.RawMessage `json:"params"` // Method parameters (method-specific structure)
}

// BrokerResponse represents a JSON-RPC response from the broker to an agent.
// Responses contain either a successful result or an error, following
// JSON-RPC 2.0 specification for standardized error handling.
//
// The ID field matches the corresponding request for correlation.
type BrokerResponse struct {
	ID     string       `json:"id"`               // Request ID for correlation
	Result interface{}  `json:"result,omitempty"` // Success result (method-specific type)
	Error  *BrokerError `json:"error,omitempty"`  // Error information if request failed
}

// BrokerError represents an error response following JSON-RPC error conventions.
// Standard error codes: -32601 (Method not found), -32602 (Invalid params),
// -32603 (Internal error). Custom codes may be used for broker-specific errors.
type BrokerError struct {
	Code    int    `json:"code"`    // JSON-RPC error code
	Message string `json:"message"` // Human-readable error description
}

// Info contains broker service information for agent discovery and connection.
// This structure is used to advertise broker capabilities and connection
// details to agents that need to establish communication.
type Info struct {
	Protocol string // Network protocol ("tcp")
	Address  string // Broker IP address or hostname
	Port     string // TCP port number
	Codec    string // Message encoding format ("json")
}

// BrokerConfig holds configuration parameters for initializing the broker service.
// This structure is used during broker startup to configure network settings,
// encoding format, and debugging options.
type BrokerConfig struct {
	Port     string // TCP port to listen on (e.g., ":9001")
	Protocol string // Network protocol ("tcp")
	Codec    string // Message encoding ("json")
	Debug    bool   // Enable debug logging
}

// NewService creates a new broker service instance with the provided configuration.
// The config parameter can be either a BrokerConfig struct or a compatible
// struct with Port, Protocol, Codec, and Debug fields.
//
// Default values are used if configuration is not provided or invalid:
// - Port: ":9001"
// - Protocol: "tcp"
// - Codec: "json"
// - Debug: false
//
// Returns a fully initialized Service ready to accept agent connections.
func NewService(config interface{}) *Service {
	// Set default configuration values
	port := ":9001"
	protocol := "tcp"
	codec := "json"
	debug := false

	// Extract configuration from provided interface
	// Support both BrokerConfig and anonymous struct types
	if bc, ok := config.(BrokerConfig); ok {
		port = bc.Port
		protocol = bc.Protocol
		codec = bc.Codec
		debug = bc.Debug
	} else if bc, ok := config.(struct {
		Port, Protocol, Codec string
		Debug                 bool
	}); ok {
		port = bc.Port
		protocol = bc.Protocol
		codec = bc.Codec
		debug = bc.Debug
	}

	// Initialize service with configuration and empty collections
	return &Service{
		port:        port,
		protocol:    protocol,
		codec:       codec,
		debug:       debug,
		topics:      make(map[string]*Topic),      // Initialize empty topics map
		pipes:       make(map[string]*Pipe),       // Initialize empty pipes map
		connections: make(map[string]*Connection), // Initialize empty connections map
	}
}

// Start begins the broker service, listening for agent connections on the configured port.
// The service runs indefinitely until the provided context is cancelled, at which point
// it performs graceful shutdown by closing the listener and rejecting new connections.
//
// Each incoming connection is handled in a separate goroutine to support concurrent
// agent communication. The service will continue accepting connections even if
// individual connection handling encounters errors.
//
// Parameters:
//   - ctx: Context for service lifecycle management and graceful shutdown
//
// Returns:
//   - error: Network setup errors or nil on successful shutdown
//
// Called by: GOX main service during startup
func (s *Service) Start(ctx context.Context) error {
	// Create TCP listener on configured port and protocol
	listener, err := net.Listen(s.protocol, s.port)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.port, err)
	}
	s.listener = listener

	if s.debug {
		log.Printf("Broker service listening on %s (%s/%s)", s.port, s.protocol, s.codec)
	}

	// Handle graceful shutdown when context is cancelled
	// This allows the service to stop accepting new connections cleanly
	go func() {
		<-ctx.Done()
		if s.debug {
			log.Printf("Broker service shutting down")
		}
		s.listener.Close()
	}()

	// Main accept loop - handle incoming agent connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if shutdown was requested via context cancellation
			if ctx.Err() != nil {
				return nil // Clean shutdown
			}
			// Log other errors but continue accepting connections
			log.Printf("Broker service accept error: %v", err)
			continue
		}

		// Handle each connection in a separate goroutine for concurrency
		go s.handleConnection(conn)
	}
}

// handleConnection manages a single agent connection throughout its lifecycle.
// This method runs in its own goroutine and handles all JSON-RPC communication
// with the connected agent, including request processing and response delivery.
//
// Connection lifecycle:
// 1. Create unique connection ID and register connection
// 2. Setup JSON encoder/decoder for message serialization
// 3. Enter request processing loop
// 4. Clean up connection on disconnect or error
//
// The connection is automatically removed from the broker's connection registry
// when this method exits, ensuring proper cleanup of resources.
//
// Parameters:
//   - netConn: TCP connection from agent
//
// Called by: Start() method for each accepted connection
func (s *Service) handleConnection(netConn net.Conn) {
	// Ensure connection is properly closed when handler exits
	defer netConn.Close()

	// Generate unique connection identifier using nanosecond timestamp
	connID := fmt.Sprintf("conn_%d", time.Now().UnixNano())

	// Create connection object with JSON codec for message handling
	conn := &Connection{
		ID:       connID,
		Conn:     netConn,
		Encoder:  json.NewEncoder(netConn), // For sending responses to agent
		Decoder:  json.NewDecoder(netConn), // For receiving requests from agent
		LastSeen: time.Now(),               // Track connection health
	}

	// Register connection in broker's connection registry
	s.connMux.Lock()
	s.connections[connID] = conn
	s.connMux.Unlock()

	// Ensure connection is removed from registry when handler exits
	// This cleanup is critical for preventing memory leaks
	defer func() {
		s.connMux.Lock()
		delete(s.connections, connID)
		s.connMux.Unlock()
	}()

	if s.debug {
		log.Printf("Broker: new connection %s", connID)
	}

	// Main request processing loop - handle JSON-RPC requests from agent
	for {
		var req BrokerRequest

		// Decode JSON-RPC request from agent
		if err := conn.Decoder.Decode(&req); err != nil {
			if s.debug {
				log.Printf("Broker: decode error from %s: %v", connID, err)
			}
			return // Exit on decode error (connection likely closed)
		}

		// Update connection health timestamp
		conn.LastSeen = time.Now()

		if s.debug {
			log.Printf("Broker: received %s from %s", req.Method, connID)
		}

		// Process request and generate response
		resp := s.handleRequest(conn, &req)

		// Send JSON-RPC response back to agent
		if err := conn.Encoder.Encode(resp); err != nil {
			if s.debug {
				log.Printf("Broker: encode error to %s: %v", connID, err)
			}
			return // Exit on encode error (connection likely closed)
		}
	}
}

// handleRequest dispatches JSON-RPC requests to appropriate handler methods.
// This is the central routing function that determines how to process each
// incoming request based on the method name.
//
// Supported methods:
//   - "connect": Agent registration and handshake
//   - "publish": Send message to topic subscribers
//   - "publish_envelope": Send envelope to topic subscribers
//   - "subscribe": Subscribe to topic for message delivery
//   - "send_pipe": Send message to point-to-point pipe
//   - "send_pipe_envelope": Send envelope to point-to-point pipe
//   - "receive_pipe": Receive message from point-to-point pipe
//
// Parameters:
//   - conn: Connection that sent the request
//   - req: JSON-RPC request to process
//
// Returns:
//   - BrokerResponse: JSON-RPC response with result or error
//
// Called by: handleConnection() for each incoming request
func (s *Service) handleRequest(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Dispatch request to appropriate handler based on method name
	switch req.Method {
	case "connect":
		return s.handleConnect(conn, req)
	case "publish":
		return s.handlePublish(conn, req)
	case "publish_envelope":
		return s.handlePublishEnvelope(conn, req)
	case "subscribe":
		return s.handleSubscribe(conn, req)
	case "send_pipe":
		return s.handleSendPipe(conn, req)
	case "send_pipe_envelope":
		return s.handleSendPipeEnvelope(conn, req)
	case "receive_pipe":
		return s.handleReceivePipe(conn, req)
	default:
		// Return JSON-RPC "Method not found" error for unknown methods
		return &BrokerResponse{
			ID: req.ID,
			Error: &BrokerError{
				Code:    -32601, // JSON-RPC standard error code
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

// handleConnect processes agent registration requests during the connection handshake.
// This method associates an agent ID with the connection, enabling proper message
// routing and identification throughout the agent's lifecycle.
//
// The agent ID is used for:
//   - Debug logging and connection tracking
//   - Envelope hop tracking for message routing history
//   - Agent identification in topic subscriptions
//
// Parameters:
//   - conn: Connection requesting registration
//   - req: JSON-RPC request with agent_id parameter
//
// Returns:
//   - BrokerResponse: Success confirmation or parameter validation error
//
// Called by: handleRequest() when method is "connect"
func (s *Service) handleConnect(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		AgentID string `json:"agent_id"` // Unique agent identifier
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Associate agent ID with this connection for future message routing
	conn.AgentID = params.AgentID

	if s.debug {
		log.Printf("Broker: agent %s connected on %s", params.AgentID, conn.ID)
	}

	// Confirm successful agent registration
	return &BrokerResponse{
		ID:     req.ID,
		Result: "connected",
	}
}

// handlePublish processes message publication requests to topics.
// This method distributes messages to all subscribers of the specified topic,
// implementing the publish/subscribe messaging pattern for event distribution.
//
// Message processing:
//   - Sets timestamp and target fields automatically
//   - Stores message in topic history for debugging
//   - Distributes to all subscribers except the sender
//   - Maintains message ordering and metadata integrity
//
// Topic management:
//   - Creates topics automatically if they don't exist
//   - Maintains message history with circular buffer (max 100 messages)
//   - Handles concurrent access with proper locking
//
// Parameters:
//   - conn: Connection publishing the message
//   - req: JSON-RPC request with topic and message parameters
//
// Returns:
//   - BrokerResponse: Success confirmation or parameter validation error
//
// Called by: handleRequest() when method is "publish"
func (s *Service) handlePublish(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		Topic   string  `json:"topic"`   // Target topic name
		Message Message `json:"message"` // Message to publish
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Set broker-managed fields for proper message routing
	params.Message.Timestamp = time.Now()                       // Record processing time
	params.Message.Target = fmt.Sprintf("pub:%s", params.Topic) // Set routing target

	// Find or create the target topic
	s.topicsMux.Lock()
	topic, exists := s.topics[params.Topic]
	if !exists {
		// Create new topic with initialized collections
		topic = &Topic{
			Name:        params.Topic,
			Subscribers: make([]*Connection, 0),             // Empty subscriber list
			Messages:    make([]*Message, 0, 100),           // Message history buffer
			Envelopes:   make([]*envelope.Envelope, 0, 100), // Envelope history buffer
		}
		s.topics[params.Topic] = topic
	}
	s.topicsMux.Unlock()

	// Add message to topic and distribute to subscribers
	topic.mux.Lock()

	// Store message in topic history with circular buffer behavior
	topic.Messages = append(topic.Messages, &params.Message)
	if len(topic.Messages) > 100 {
		topic.Messages = topic.Messages[1:] // Remove oldest message
	}

	// Distribute message to all subscribers except the sender
	for _, subscriber := range topic.Subscribers {
		if subscriber.ID != conn.ID { // Prevent message echo to sender
			// Create properly formatted message for subscriber delivery
			// CRITICAL: Preserve all message fields including metadata
			pubMsg := Message{
				ID:        params.Message.ID,                   // Original message ID
				Type:      params.Message.Type,                 // Message type
				Target:    fmt.Sprintf("pub:%s", params.Topic), // Routing info
				Payload:   params.Message.Payload,              // Message data
				Meta:      params.Message.Meta,                 // Critical: preserve metadata!
				Timestamp: params.Message.Timestamp,            // Processing timestamp
			}

			// Send message to subscriber (non-blocking)
			if err := subscriber.Encoder.Encode(pubMsg); err != nil {
				if s.debug {
					log.Printf("Broker: failed to send to subscriber %s: %v", subscriber.ID, err)
				}
				// Continue with other subscribers even if one fails
			}
		}
	}
	topic.mux.Unlock()

	if s.debug {
		log.Printf("Broker: published to topic %s (%d subscribers)", params.Topic, len(topic.Subscribers))
	}

	// Confirm successful message publication
	return &BrokerResponse{
		ID:     req.ID,
		Result: "published",
	}
}

// handleSubscribe processes topic subscription requests from agents.
// Subscribing to a topic allows an agent to receive all messages published
// to that topic, enabling event-driven communication patterns.
//
// Topic management:
//   - Creates new topics automatically if they don't exist
//   - Prevents duplicate subscriptions for the same connection
//   - Maintains subscriber list for message distribution
//
// Parameters:
//   - conn: Connection requesting subscription
//   - req: JSON-RPC request with topic parameter
//
// Returns:
//   - BrokerResponse: Success confirmation or parameter validation error
//
// Called by: handleRequest() when method is "subscribe"
func (s *Service) handleSubscribe(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		Topic string `json:"topic"` // Topic name to subscribe to
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Find or create the requested topic
	s.topicsMux.Lock()
	topic, exists := s.topics[params.Topic]
	if !exists {
		// Create new topic with initialized collections
		topic = &Topic{
			Name:        params.Topic,
			Subscribers: make([]*Connection, 0),             // Empty subscriber list
			Messages:    make([]*Message, 0, 100),           // Message history buffer
			Envelopes:   make([]*envelope.Envelope, 0, 100), // Envelope history buffer
		}
		s.topics[params.Topic] = topic
	}
	s.topicsMux.Unlock()

	// Add connection to topic's subscriber list (avoid duplicates)
	topic.mux.Lock()
	found := false
	for _, sub := range topic.Subscribers {
		if sub.ID == conn.ID {
			found = true
			break
		}
	}
	if !found {
		topic.Subscribers = append(topic.Subscribers, conn)
	}
	topic.mux.Unlock()

	if s.debug {
		log.Printf("Broker: agent %s subscribed to topic %s", conn.AgentID, params.Topic)
	}

	// Confirm successful subscription
	return &BrokerResponse{
		ID:     req.ID,
		Result: "subscribed",
	}
}

// handlePublishEnvelope processes envelope publication requests to topics.
// This method provides full envelope protocol support for message distribution,
// including metadata tracking, hop recording, and routing information.
//
// Envelope processing:
//   - Validates envelope structure and required fields
//   - Records message routing hops for debugging and audit trails
//   - Sets destination field for proper routing
//   - Preserves all envelope metadata and routing information
//
// The envelope protocol provides richer metadata compared to simple messages,
// including sender information, routing history, and processing context.
//
// Parameters:
//   - conn: Connection publishing the envelope
//   - req: JSON-RPC request with topic and envelope parameters
//
// Returns:
//   - BrokerResponse: Success confirmation or validation error
//
// Called by: handleRequest() when method is "publish_envelope"
func (s *Service) handlePublishEnvelope(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		Topic    string             `json:"topic"`    // Target topic name
		Envelope *envelope.Envelope `json:"envelope"` // Envelope to publish
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Validate envelope structure and required fields
	if err := params.Envelope.Validate(); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: fmt.Sprintf("Invalid envelope: %v", err)},
		}
	}

	// Record message routing hop for audit trail
	// This tracks which agents have processed this envelope
	if conn.AgentID != "" {
		params.Envelope.AddHop(conn.AgentID)
	}

	// Set routing destination if not already specified
	if params.Envelope.Destination == "" {
		params.Envelope.Destination = fmt.Sprintf("pub:%s", params.Topic)
	}

	// Find or create the target topic
	s.topicsMux.Lock()
	topic, exists := s.topics[params.Topic]
	if !exists {
		// Create new topic with initialized collections
		topic = &Topic{
			Name:        params.Topic,
			Subscribers: make([]*Connection, 0),             // Empty subscriber list
			Messages:    make([]*Message, 0, 100),           // Message history buffer
			Envelopes:   make([]*envelope.Envelope, 0, 100), // Envelope history buffer
		}
		s.topics[params.Topic] = topic
	}
	s.topicsMux.Unlock()

	// Add envelope to topic and distribute to subscribers
	topic.mux.Lock()

	// Store envelope in topic history with circular buffer behavior
	topic.Envelopes = append(topic.Envelopes, params.Envelope)
	if len(topic.Envelopes) > 100 {
		topic.Envelopes = topic.Envelopes[1:] // Remove oldest envelope
	}

	// Distribute envelope to all subscribers except the sender
	for _, subscriber := range topic.Subscribers {
		if subscriber.ID != conn.ID { // Prevent envelope echo to sender
			// Send complete envelope with all metadata preserved
			if err := subscriber.Encoder.Encode(params.Envelope); err != nil {
				if s.debug {
					log.Printf("Broker: failed to send envelope to subscriber %s: %v", subscriber.ID, err)
				}
				// Continue with other subscribers even if one fails
			}
		}
	}
	topic.mux.Unlock()

	if s.debug {
		log.Printf("Broker: published envelope to topic %s (%d subscribers)", params.Topic, len(topic.Subscribers))
	}

	// Confirm successful envelope publication
	return &BrokerResponse{
		ID:     req.ID,
		Result: "published",
	}
}

// handleSendPipe processes message sending requests to point-to-point pipes.
// Pipes provide reliable message delivery between specific agents with buffering
// and flow control to handle varying processing speeds.
//
// Pipe management:
//   - Creates pipes automatically when first used
//   - Uses buffered channels for non-blocking message delivery
//   - Handles buffer overflow with appropriate error responses
//   - Sets proper routing information for message delivery
//
// Unlike topics, pipes provide one-to-one communication with guaranteed
// delivery order and buffering capabilities.
//
// Parameters:
//   - conn: Connection sending the message
//   - req: JSON-RPC request with pipe and message parameters
//
// Returns:
//   - BrokerResponse: Success confirmation or buffer overflow error
//
// Called by: handleRequest() when method is "send_pipe"
func (s *Service) handleSendPipe(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		Pipe    string  `json:"pipe"`    // Target pipe name
		Message Message `json:"message"` // Message to send
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Set broker-managed fields for proper message routing
	params.Message.Timestamp = time.Now()                       // Record processing time
	params.Message.Target = fmt.Sprintf("pipe:%s", params.Pipe) // Set routing target

	// Find or create the target pipe
	s.pipesMux.Lock()
	pipe, exists := s.pipes[params.Pipe]
	if !exists {
		// Create new pipe with buffered channels
		pipe = &Pipe{
			Name:      params.Pipe,
			Messages:  make(chan *Message, 100),           // Buffered message channel
			Envelopes: make(chan *envelope.Envelope, 100), // Buffered envelope channel
		}
		s.pipes[params.Pipe] = pipe
	}
	s.pipesMux.Unlock()

	// Attempt to send message to pipe with flow control
	select {
	case pipe.Messages <- &params.Message:
		// Message successfully queued in pipe buffer
		if s.debug {
			log.Printf("Broker: sent message to pipe %s", params.Pipe)
		}
		return &BrokerResponse{
			ID:     req.ID,
			Result: "sent",
		}
	default:
		// Pipe buffer is full - cannot accept more messages
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32603, Message: "Pipe buffer full"},
		}
	}
}

// handleSendPipeEnvelope processes envelope sending requests to point-to-point pipes.
// This method provides full envelope protocol support for pipe communication,
// including metadata tracking, hop recording, and routing information.
//
// Envelope processing for pipes:
//   - Validates envelope structure and required fields
//   - Records message routing hops for debugging and audit trails
//   - Sets destination field for proper routing
//   - Uses buffered channels for reliable delivery
//
// The envelope protocol provides richer metadata for pipe communication,
// useful for complex agent workflows that require detailed routing information.
//
// Parameters:
//   - conn: Connection sending the envelope
//   - req: JSON-RPC request with pipe and envelope parameters
//
// Returns:
//   - BrokerResponse: Success confirmation, validation error, or buffer overflow error
//
// Called by: handleRequest() when method is "send_pipe_envelope"
func (s *Service) handleSendPipeEnvelope(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		Pipe     string             `json:"pipe"`     // Target pipe name
		Envelope *envelope.Envelope `json:"envelope"` // Envelope to send
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Validate envelope structure and required fields
	if err := params.Envelope.Validate(); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: fmt.Sprintf("Invalid envelope: %v", err)},
		}
	}

	// Record message routing hop for audit trail
	// This tracks which agents have processed this envelope
	if conn.AgentID != "" {
		params.Envelope.AddHop(conn.AgentID)
	}

	// Set routing destination if not already specified
	if params.Envelope.Destination == "" {
		params.Envelope.Destination = fmt.Sprintf("pipe:%s", params.Pipe)
	}

	// Find or create the target pipe
	s.pipesMux.Lock()
	pipe, exists := s.pipes[params.Pipe]
	if !exists {
		// Create new pipe with buffered channels
		pipe = &Pipe{
			Name:      params.Pipe,
			Messages:  make(chan *Message, 100),           // Buffered message channel
			Envelopes: make(chan *envelope.Envelope, 100), // Buffered envelope channel
		}
		s.pipes[params.Pipe] = pipe
	}
	s.pipesMux.Unlock()

	// Attempt to send envelope to pipe with flow control
	select {
	case pipe.Envelopes <- params.Envelope:
		// Envelope successfully queued in pipe buffer
		if s.debug {
			log.Printf("Broker: sent envelope to pipe %s", params.Pipe)
		}
		return &BrokerResponse{
			ID:     req.ID,
			Result: "sent",
		}
	default:
		// Pipe buffer is full - cannot accept more envelopes
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32603, Message: "Pipe buffer full"},
		}
	}
}

// handleReceivePipe processes message receiving requests from point-to-point pipes.
// This method blocks until a message or envelope is available, or until a timeout
// occurs, providing reliable message consumption for agent workflows.
//
// Receive behavior:
//   - Blocks until message/envelope is available or timeout occurs
//   - Handles both simple messages and full envelope protocol
//   - Creates pipes automatically if they don't exist
//   - Supports configurable timeout with reasonable defaults
//
// The method uses Go's select statement to handle multiple channel operations
// concurrently, ensuring proper timeout handling and message ordering.
//
// Timeout handling:
//   - Default timeout: 5 seconds (5000ms)
//   - Configurable via timeout_ms parameter
//   - Returns error response on timeout
//
// Parameters:
//   - conn: Connection requesting message reception
//   - req: JSON-RPC request with pipe and optional timeout parameters
//
// Returns:
//   - BrokerResponse: Message/envelope result, parameter error, or timeout error
//
// Called by: handleRequest() when method is "receive_pipe"
func (s *Service) handleReceivePipe(conn *Connection, req *BrokerRequest) *BrokerResponse {
	// Define expected parameter structure for type-safe unmarshaling
	var params struct {
		Pipe    string `json:"pipe"`                 // Source pipe name
		Timeout int    `json:"timeout_ms,omitempty"` // Optional timeout in milliseconds
	}

	// Parse and validate request parameters
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Find or create the source pipe
	s.pipesMux.Lock()
	pipe, exists := s.pipes[params.Pipe]
	if !exists {
		// Create new pipe with buffered channels
		// This allows senders to queue messages even if no receiver is waiting
		pipe = &Pipe{
			Name:      params.Pipe,
			Messages:  make(chan *Message, 100),           // Buffered message channel
			Envelopes: make(chan *envelope.Envelope, 100), // Buffered envelope channel
		}
		s.pipes[params.Pipe] = pipe
	}
	s.pipesMux.Unlock()

	// Configure timeout with reasonable default
	timeout := 5000 // Default: 5 seconds
	if params.Timeout > 0 {
		timeout = params.Timeout
	}

	// Wait for message, envelope, or timeout using select statement
	// This provides concurrent handling of multiple channel operations
	select {
	case msg := <-pipe.Messages:
		// Message received successfully
		return &BrokerResponse{
			ID:     req.ID,
			Result: msg,
		}
	case env := <-pipe.Envelopes:
		// Envelope received successfully
		return &BrokerResponse{
			ID:     req.ID,
			Result: env,
		}
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		// Timeout occurred - no message available within specified time
		return &BrokerResponse{
			ID:    req.ID,
			Error: &BrokerError{Code: -32603, Message: "Timeout waiting for message"},
		}
	}
}
