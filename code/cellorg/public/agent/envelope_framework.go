package agent

import (
	"fmt"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/broker"
	"github.com/tenzoki/agen/cellorg/internal/envelope"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/omni/tokencount"
)

// EnvelopeFramework provides envelope-based communication with automatic chunking support
// This is used by agents that need to handle large payloads and want transparent chunking
type EnvelopeFramework struct {
	client          *client.BrokerClient
	chunkCollector  *ChunkCollector
	defaultCounter  tokencount.Counter
	providerConfig  *broker.ProviderConfig
	subscriptions   map[string]<-chan *envelope.Envelope // topic -> channel
}

// NewEnvelopeFramework creates a new envelope framework with chunking support
func NewEnvelopeFramework(brokerClient *client.BrokerClient) *EnvelopeFramework {
	return &EnvelopeFramework{
		client:         brokerClient,
		chunkCollector: NewChunkCollector(5 * time.Minute), // Default 5 minute timeout
		providerConfig: broker.NewProviderConfig(),
		subscriptions:  make(map[string]<-chan *envelope.Envelope),
	}
}

// RegisterProvider associates a destination with a token counter for automatic chunking
// This enables the framework to automatically chunk outgoing messages to AI-powered agents
func (ef *EnvelopeFramework) RegisterProvider(destination string, counter tokencount.Counter) {
	ef.providerConfig.RegisterProvider(destination, counter)
}

// SetDefaultCounter sets a default counter for all destinations that don't have specific providers
func (ef *EnvelopeFramework) SetDefaultCounter(counter tokencount.Counter) {
	ef.defaultCounter = counter
}

// Subscribe subscribes to a topic and returns a channel for receiving complete envelopes
// Chunked envelopes are automatically reassembled before being delivered
func (ef *EnvelopeFramework) Subscribe(topic string) (<-chan *envelope.Envelope, error) {
	// Check if already subscribed
	if _, exists := ef.subscriptions[topic]; exists {
		return nil, fmt.Errorf("already subscribed to topic: %s", topic)
	}

	// Subscribe to topic via client
	rawChan, err := ef.client.SubscribeEnvelopes(topic)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	// Create processed channel
	processedChan := make(chan *envelope.Envelope, 10)

	// Start processing goroutine that handles chunking
	go ef.processEnvelopes(rawChan, processedChan)

	ef.subscriptions[topic] = processedChan
	return processedChan, nil
}

// processEnvelopes handles incoming envelopes and reassembles chunks
func (ef *EnvelopeFramework) processEnvelopes(input <-chan *envelope.Envelope, output chan<- *envelope.Envelope) {
	defer close(output)

	for env := range input {
		// Collect chunks or pass through non-chunked envelopes
		merged, complete, err := ef.chunkCollector.CollectChunk(env)
		if err != nil {
			// Log error but continue processing
			continue
		}

		if complete && merged != nil {
			// Send complete envelope to output
			output <- merged
		}
		// If not complete, just continue waiting for more chunks
	}
}

// Publish publishes an envelope with automatic chunking if needed
// The framework determines if chunking is needed based on registered providers
func (ef *EnvelopeFramework) Publish(topic string, env *envelope.Envelope) error {
	// Determine if we should use chunking based on destination
	counter := ef.providerConfig.GetCounter(env.Destination)
	if counter == nil && ef.defaultCounter != nil {
		counter = ef.defaultCounter
	}

	if counter == nil {
		// No chunking support, publish directly
		return ef.client.PublishEnvelope(topic, env)
	}

	// Create chunking publisher
	publishFunc := func(e *envelope.Envelope) error {
		return ef.client.PublishEnvelope(topic, e)
	}

	publisher := broker.NewChunkingPublisher(counter, publishFunc)
	return publisher.Publish(env)
}

// PublishTo publishes an envelope to a specific destination with automatic chunking
func (ef *EnvelopeFramework) PublishTo(topic, destination string, env *envelope.Envelope) error {
	// Set destination
	env.Destination = destination

	return ef.Publish(topic, env)
}

// Close stops the envelope framework and cleans up resources
func (ef *EnvelopeFramework) Close() {
	if ef.chunkCollector != nil {
		ef.chunkCollector.Close()
	}
}

// GetChunkStatus returns the current status of pending chunks
// This is useful for monitoring and debugging
func (ef *EnvelopeFramework) GetChunkStatus() map[string]ChunkStatus {
	if ef.chunkCollector == nil {
		return make(map[string]ChunkStatus)
	}
	return ef.chunkCollector.GetStatus()
}

// CountPendingChunks returns the number of incomplete chunk groups
func (ef *EnvelopeFramework) CountPendingChunks() int {
	if ef.chunkCollector == nil {
		return 0
	}
	return ef.chunkCollector.CountPendingChunks()
}
