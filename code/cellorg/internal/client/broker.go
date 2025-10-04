// Package client provides client-side broker communication for GOX agents.
// This package enables agents to connect to the central broker and participate
// in the distributed messaging system using both simple messages and full
// envelope protocol support.
//
// Key Features:
// - TCP connection management with automatic reconnection
// - JSON-RPC protocol for broker communication
// - Publish/Subscribe messaging for event distribution
// - Point-to-point pipes for direct agent communication
// - Full envelope protocol support with metadata tracking
// - Concurrent message handling with proper synchronization
// - Request/response correlation and timeout handling
//
// The client handles all the complexity of broker communication, allowing
// agents to focus on their specific processing tasks while providing
// reliable message delivery and routing capabilities.
package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/agen/cellorg/internal/envelope"
)

// BrokerClient manages communication between an agent and the central broker.
// It handles all aspects of broker connectivity including connection management,
// message routing, subscription handling, and protocol translation.
//
// The client supports both simple message and full envelope protocols,
// allowing agents to choose the appropriate level of metadata richness
// for their communication needs.
//
// Thread Safety: All public methods are thread-safe and can be called
// concurrently from multiple goroutines.
type BrokerClient struct {
	// Connection configuration
	address string // Broker TCP address (e.g., "localhost:9001")
	agentID string // Unique agent identifier for this client
	debug   bool   // Enable debug logging

	// Network connection state
	conn    net.Conn      // TCP connection to broker
	encoder *json.Encoder // JSON encoder for sending requests
	decoder *json.Decoder // JSON decoder for receiving responses
	mux     sync.Mutex    // Protects connection state during connect/disconnect

	// Request/response correlation
	reqID int64 // Incrementing request ID counter

	// Message routing for subscriptions
	listeners    map[string]chan *BrokerMessage     // Topic message listeners
	envListeners map[string]chan *envelope.Envelope // Topic envelope listeners
	listenersMux sync.RWMutex                       // Protects listener maps

	// Request/response correlation for JSON-RPC calls
	responseChans map[string]chan *BrokerResponse // Pending request channels
	responseChMux sync.RWMutex                    // Protects response channel map
}

// BrokerRequest represents a JSON-RPC request sent to the broker.
// This follows the JSON-RPC 2.0 specification for standardized
// remote procedure call communication.
type BrokerRequest struct {
	ID     string          `json:"id"`     // Request identifier for response correlation
	Method string          `json:"method"` // Broker method to invoke
	Params json.RawMessage `json:"params"` // Method parameters (method-specific structure)
}

// BrokerResponse represents a JSON-RPC response received from the broker.
// Contains either a successful result or an error, following JSON-RPC 2.0
// specification for standardized error handling.
type BrokerResponse struct {
	ID     string          `json:"id"`               // Request ID for correlation
	Result json.RawMessage `json:"result,omitempty"` // Success result (method-specific)
	Error  *BrokerError    `json:"error,omitempty"`  // Error information if request failed
}

// BrokerError represents an error response from the broker following
// JSON-RPC error conventions. Standard error codes include -32601
// (Method not found), -32602 (Invalid params), -32603 (Internal error).
type BrokerError struct {
	Code    int    `json:"code"`    // JSON-RPC error code
	Message string `json:"message"` // Human-readable error description
}

// BrokerMessage represents a simple message received from the broker.
// This is used for basic agent communication when full envelope protocol
// is not required, providing a lightweight alternative for simple data exchange.
type BrokerMessage struct {
	ID        string                 `json:"id"`        // Unique message identifier
	Type      string                 `json:"type"`      // Message type for handling dispatch
	Target    string                 `json:"target"`    // Routing target (topic or pipe)
	Payload   interface{}            `json:"payload"`   // Message data (any JSON-serializable type)
	Meta      map[string]interface{} `json:"meta"`      // Metadata for message processing
	Timestamp time.Time              `json:"timestamp"` // When message was processed by broker
}

// NewBrokerClient creates a new broker client instance for agent communication.
// The client is created in a disconnected state and requires calling Connect()
// before it can be used for broker communication.
//
// Parameters:
//   - address: Broker TCP address (e.g., "localhost:9001")
//   - agentID: Unique identifier for this agent
//   - debug: Enable debug logging for troubleshooting
//
// Returns:
//   - *BrokerClient: Ready-to-connect client instance
//
// Called by: Agent initialization code during startup
func NewBrokerClient(address, agentID string, debug bool) *BrokerClient {
	return &BrokerClient{
		address:       address,
		agentID:       agentID,
		debug:         debug,
		listeners:     make(map[string]chan *BrokerMessage),     // Initialize message listeners
		envListeners:  make(map[string]chan *envelope.Envelope), // Initialize envelope listeners
		responseChans: make(map[string]chan *BrokerResponse),    // Initialize response channels
	}
}

// Connect establishes a TCP connection to the broker and performs agent registration.
// This method sets up the complete communication pipeline including connection,
// message listener, and agent handshake with the broker.
//
// Connection process:
// 1. Establish TCP connection to broker
// 2. Create JSON encoder/decoder for message serialization
// 3. Start background message listener goroutine
// 4. Send agent registration request
// 5. Complete handshake and mark agent as connected
//
// The method is idempotent - calling it multiple times on an already
// connected client will return immediately without error.
//
// Returns:
//   - error: Network connection error, registration error, or nil on success
//
// Called by: Agent initialization code during startup
func (c *BrokerClient) Connect() error {
	c.mux.Lock()

	// Check if already connected to avoid duplicate connections
	if c.conn != nil {
		c.mux.Unlock()
		return nil // Already connected
	}

	// Establish TCP connection to broker
	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		c.mux.Unlock()
		return fmt.Errorf("failed to connect to broker at %s: %w", c.address, err)
	}

	// Set up JSON codec for message serialization
	c.conn = conn
	c.encoder = json.NewEncoder(conn)
	c.decoder = json.NewDecoder(conn)

	// Start background message listener for incoming messages
	// This goroutine handles subscription deliveries and response correlation
	go c.messageListener()

	// Release mutex before making JSON-RPC calls to avoid deadlock
	c.mux.Unlock()

	// Allow message listener time to initialize
	time.Sleep(10 * time.Millisecond)

	// Send agent registration request to broker
	params := map[string]interface{}{
		"agent_id": c.agentID,
	}
	if _, err := c.call("connect", params); err != nil {
		// Clean up connection on registration failure
		c.mux.Lock()
		conn.Close()
		c.conn = nil
		c.encoder = nil
		c.decoder = nil
		c.mux.Unlock()
		return fmt.Errorf("failed to register with broker: %w", err)
	}

	if c.debug {
		log.Printf("Connected to broker at %s", c.address)
	}

	return nil
}

// Disconnect closes the connection to the broker and cleans up resources.
// This method performs graceful shutdown of the broker connection, stopping
// the message listener and clearing all connection state.
//
// Cleanup includes:
// - Closing TCP connection (triggers message listener shutdown)
// - Clearing encoder/decoder references
// - Resetting connection state for potential reconnection
//
// The method is idempotent - calling it multiple times or on an already
// disconnected client will return nil without error.
//
// Returns:
//   - error: Connection close error or nil on success
//
// Called by: Agent shutdown code or for connection management
func (c *BrokerClient) Disconnect() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Close connection if it exists
	if c.conn != nil {
		err := c.conn.Close() // This will cause messageListener to exit
		// Clear connection state to allow future reconnection
		c.conn = nil
		c.encoder = nil
		c.decoder = nil
		return err
	}
	return nil
}

// call executes a JSON-RPC method call to the broker with request/response correlation.
// This is the core communication method that handles request serialization,
// response correlation, timeout handling, and error processing.
//
// Request/response flow:
// 1. Generate unique request ID for correlation
// 2. Marshal parameters to JSON
// 3. Create response channel for this specific request
// 4. Send JSON-RPC request to broker
// 5. Wait for correlated response or timeout
// 6. Clean up response channel and return result
//
// The method uses Go channels for response correlation, allowing multiple
// concurrent requests without blocking each other.
//
// Parameters:
//   - method: Broker method name (e.g., "publish", "subscribe")
//   - params: Method parameters (will be JSON-marshaled)
//
// Returns:
//   - json.RawMessage: Raw JSON response data for method-specific parsing
//   - error: Connection error, marshaling error, broker error, or timeout
//
// Called by: All public broker communication methods
func (c *BrokerClient) call(method string, params interface{}) (json.RawMessage, error) {
	// Verify connection is established
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to broker")
	}

	// Generate unique request ID for response correlation
	c.reqID++
	reqID := fmt.Sprintf("req_%d", c.reqID)

	// Marshal parameters to JSON if provided
	var paramsBytes json.RawMessage
	if params != nil {
		var err error
		paramsBytes, err = json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	// Create JSON-RPC request structure
	req := BrokerRequest{
		ID:     reqID,
		Method: method,
		Params: paramsBytes,
	}

	// Create dedicated response channel for this request
	// Buffer size of 1 ensures non-blocking response delivery
	respChan := make(chan *BrokerResponse, 1)
	c.responseChMux.Lock()
	c.responseChans[reqID] = respChan
	c.responseChMux.Unlock()

	// Send JSON-RPC request to broker
	if err := c.encoder.Encode(req); err != nil {
		// Clean up response channel on send failure
		c.responseChMux.Lock()
		delete(c.responseChans, reqID)
		c.responseChMux.Unlock()
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response with timeout handling
	select {
	case resp := <-respChan:
		// Clean up response channel after receiving response
		c.responseChMux.Lock()
		delete(c.responseChans, reqID)
		c.responseChMux.Unlock()

		// Handle closed channel (connection lost during request)
		if resp == nil {
			return nil, fmt.Errorf("response channel closed")
		}

		// Check for broker-level errors
		if resp.Error != nil {
			return nil, fmt.Errorf("broker error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
		}

		return resp.Result, nil
	case <-time.After(30 * time.Second):
		// Clean up response channel on timeout
		c.responseChMux.Lock()
		delete(c.responseChans, reqID)
		c.responseChMux.Unlock()
		return nil, fmt.Errorf("request timeout")
	}
}

// messageListener runs in a background goroutine to handle incoming messages from the broker.
// This method is the core of the client's message routing system, responsible for:
// - Receiving all messages from the broker connection
// - Determining message types (responses, envelopes, regular messages)
// - Routing messages to appropriate handlers (response correlation, subscriptions)
// - Managing concurrent access to routing data structures
//
// Message type detection:
// - JSON-RPC responses: Have ID field with result or error
// - Envelopes: Have source, destination, and message_type fields
// - Regular messages: Have type and target fields
//
// The listener runs until the connection is closed, at which point it exits
// gracefully. Any panics are caught and logged for debugging.
//
// Called by: Connect() method as a background goroutine
func (c *BrokerClient) messageListener() {
	// Catch and log any panics to prevent client crashes
	defer func() {
		if r := recover(); r != nil {
			if c.debug {
				log.Printf("Broker message listener panic: %v", r)
			}
		}
	}()

	for {
		// Note: decoder is set once during Connect() and only cleared during Disconnect()
		// Since messageListener is only running between Connect() and Disconnect(),
		// it's safe to read decoder without mutex
		decoder := c.decoder

		if decoder == nil {
			return // Connection closed
		}

		// Read raw JSON to determine message type
		var rawMsg json.RawMessage
		if err := decoder.Decode(&rawMsg); err != nil {
			if c.debug {
				log.Printf("Broker message decode error: %v", err)
			}
			return
		}

		// Try to determine message type: envelope, regular message, or response
		var msgType struct {
			// Response fields
			ID     string          `json:"id"`
			Result json.RawMessage `json:"result,omitempty"`
			Error  *BrokerError    `json:"error,omitempty"`
			// Envelope fields
			Source      string `json:"source"`
			Destination string `json:"destination"`
			MessageType string `json:"message_type"`
			// Regular message fields
			Type   string `json:"type"`
			Target string `json:"target"`
		}

		if err := json.Unmarshal(rawMsg, &msgType); err != nil {
			if c.debug {
				log.Printf("Failed to parse message type: %v", err)
			}
			continue
		}

		// Check if this is a response message (has result or error field)
		if msgType.ID != "" && (msgType.Result != nil || msgType.Error != nil) {
			// This is a response message - route it to the waiting call
			var resp BrokerResponse
			if err := json.Unmarshal(rawMsg, &resp); err != nil {
				if c.debug {
					log.Printf("Failed to decode response: %v", err)
				}
				continue
			}

			c.responseChMux.RLock()
			if responseChan, exists := c.responseChans[resp.ID]; exists {
				select {
				case responseChan <- &resp:
					// Response delivered
				default:
					if c.debug {
						log.Printf("Warning: response channel full for request %s", resp.ID)
					}
				}
			}
			c.responseChMux.RUnlock()
			continue
		} else if msgType.Source != "" && msgType.Destination != "" && msgType.MessageType != "" {
			// This is an envelope
			var env envelope.Envelope
			if err := json.Unmarshal(rawMsg, &env); err != nil {
				if c.debug {
					log.Printf("Failed to decode envelope: %v", err)
				}
				continue
			}

			if c.debug {
				log.Printf("Received envelope: %s -> %s (%s)", env.Source, env.Destination, env.MessageType)
			}

			// Route envelope to appropriate listener
			c.listenersMux.RLock()
			if listener, exists := c.envListeners[env.Destination]; exists {
				select {
				case listener <- &env:
					// Envelope delivered
				default:
					if c.debug {
						log.Printf("Warning: envelope listener channel full for target %s", env.Destination)
					}
				}
			}
			c.listenersMux.RUnlock()
		} else if msgType.Type != "" && msgType.Target != "" {
			// This is a regular message
			var msg BrokerMessage
			if err := json.Unmarshal(rawMsg, &msg); err != nil {
				if c.debug {
					log.Printf("Failed to decode regular message: %v", err)
				}
				continue
			}

			if c.debug {
				log.Printf("Received message: ID=%s, Target=%s, Type=%s, Meta=%+v", msg.ID, msg.Target, msg.Type, msg.Meta)
			}

			// Route message to appropriate listener
			c.listenersMux.RLock()
			if listener, exists := c.listeners[msg.Target]; exists {
				select {
				case listener <- &msg:
					// Message delivered
				default:
					if c.debug {
						log.Printf("Warning: listener channel full for target %s", msg.Target)
					}
				}
			}
			c.listenersMux.RUnlock()
		} else {
			if c.debug {
				log.Printf("Unknown message format received: %s", string(rawMsg))
			}
		}
	}
}

func (c *BrokerClient) Publish(topic string, message BrokerMessage) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	params := map[string]interface{}{
		"topic":   topic,
		"message": message,
	}

	_, err := c.call("publish", params)
	return err
}

// Subscribe registers for message delivery on a specific topic.
// Creates a buffered channel for receiving messages published to the topic,
// enabling event-driven communication patterns between agents.
//
// Subscription process:
// 1. Send subscription request to broker
// 2. Create buffered message channel (100 message capacity)
// 3. Register channel with message router for delivery
// 4. Return channel for agent message consumption
//
// The returned channel will receive all messages published to the topic
// by other agents. Messages are delivered asynchronously via the background
// message listener goroutine.
//
// Parameters:
//   - topic: Topic name to subscribe to (e.g., "new-files")
//
// Returns:
//   - <-chan *BrokerMessage: Read-only channel for receiving messages
//   - error: Subscription error or broker communication error
//
// Called by: Agents that need to receive event notifications
func (c *BrokerClient) Subscribe(topic string) (<-chan *BrokerMessage, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Send subscription request to broker
	params := map[string]interface{}{
		"topic": topic,
	}

	if _, err := c.call("subscribe", params); err != nil {
		return nil, err
	}

	// Create buffered message channel for this subscription
	// Buffer size of 100 provides reasonable backpressure handling
	target := fmt.Sprintf("pub:%s", topic)
	msgChan := make(chan *BrokerMessage, 100)

	// Register channel with message router for automatic delivery
	c.listenersMux.Lock()
	c.listeners[target] = msgChan
	c.listenersMux.Unlock()

	if c.debug {
		log.Printf("Subscribed to topic: %s", topic)
	}

	return msgChan, nil
}

// Envelope-based publish method
func (c *BrokerClient) PublishEnvelope(topic string, env *envelope.Envelope) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	params := map[string]interface{}{
		"topic":    topic,
		"envelope": env,
	}

	_, err := c.call("publish_envelope", params)
	return err
}

// Subscribe to envelopes on a topic
func (c *BrokerClient) SubscribeEnvelopes(topic string) (<-chan *envelope.Envelope, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	params := map[string]interface{}{
		"topic": topic,
	}

	if _, err := c.call("subscribe", params); err != nil {
		return nil, err
	}

	// Create envelope channel for this subscription
	target := fmt.Sprintf("sub:%s", topic)
	envChan := make(chan *envelope.Envelope, 100)

	c.listenersMux.Lock()
	c.envListeners[target] = envChan
	c.listenersMux.Unlock()

	if c.debug {
		log.Printf("Subscribed to topic for envelopes: %s", topic)
	}

	return envChan, nil
}

// Send envelope via pipe
func (c *BrokerClient) SendPipeEnvelope(pipeName string, env *envelope.Envelope) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	params := map[string]interface{}{
		"pipe":     pipeName,
		"envelope": env,
	}

	_, err := c.call("send_pipe_envelope", params)
	return err
}

// Receive message or envelope from pipe
func (c *BrokerClient) ReceivePipe(pipeName string, timeoutMs int) (interface{}, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	params := map[string]interface{}{
		"pipe": pipeName,
	}
	if timeoutMs > 0 {
		params["timeout_ms"] = timeoutMs
	}

	result, err := c.call("receive_pipe", params)
	if err != nil {
		return nil, err
	}

	// Try to unmarshal as envelope first, then as regular message
	var env envelope.Envelope
	if err := json.Unmarshal(result, &env); err == nil && env.Source != "" {
		return &env, nil
	}

	var msg BrokerMessage
	if err := json.Unmarshal(result, &msg); err == nil {
		return &msg, nil
	}

	return result, nil
}

// Convenience method to create and publish an envelope
func (c *BrokerClient) PublishMessage(topic, messageType string, payload interface{}) error {
	env, err := envelope.NewEnvelope(c.agentID, fmt.Sprintf("pub:%s", topic), messageType, payload)
	if err != nil {
		return fmt.Errorf("failed to create envelope: %w", err)
	}

	return c.PublishEnvelope(topic, env)
}

// Pipe methods using the proper pipe functionality
func (c *BrokerClient) ConnectPipe(pipeName, role string) error {
	if c.debug {
		log.Printf("ConnectPipe: %s as %s", pipeName, role)
	}
	// Pipes are created automatically when first used
	return nil
}

func (c *BrokerClient) SendPipe(pipeName string, message BrokerMessage) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	params := map[string]interface{}{
		"pipe":    pipeName,
		"message": message,
	}

	_, err := c.call("send_pipe", params)
	return err
}
