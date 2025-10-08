package broker

import (
	"fmt"

	"github.com/tenzoki/agen/cellorg/internal/envelope"
	"github.com/tenzoki/agen/omni/tokencount"
)

// ChunkingHelper provides utilities for chunking large envelopes before broker publication
// This is used by the agent framework to automatically handle large payloads
type ChunkingHelper struct {
	counter tokencount.Counter
}

// NewChunkingHelper creates a new chunking helper with the specified token counter
func NewChunkingHelper(counter tokencount.Counter) *ChunkingHelper {
	return &ChunkingHelper{
		counter: counter,
	}
}

// ShouldChunk determines if an envelope needs chunking based on token budget
// Returns true if chunking is recommended, false otherwise
func (h *ChunkingHelper) ShouldChunk(env *envelope.Envelope) (bool, error) {
	if h.counter == nil {
		// No counter available, assume no chunking needed
		return false, nil
	}

	budget, err := envelope.CalculateBudget(env, h.counter)
	if err != nil {
		return false, fmt.Errorf("failed to calculate budget: %w", err)
	}

	return budget.NeedsSplitting, nil
}

// PrepareForPublish checks if envelope needs chunking and returns envelopes ready for publishing
// If chunking is needed, returns multiple chunk envelopes
// If no chunking needed, returns the original envelope
func (h *ChunkingHelper) PrepareForPublish(env *envelope.Envelope) ([]*envelope.Envelope, error) {
	if h.counter == nil {
		// No counter, return original envelope
		return []*envelope.Envelope{env}, nil
	}

	// Calculate budget
	budget, err := envelope.CalculateBudget(env, h.counter)
	if err != nil {
		// On error, fallback to sending without chunking
		return []*envelope.Envelope{env}, nil
	}

	if !budget.NeedsSplitting {
		// No chunking needed
		return []*envelope.Envelope{env}, nil
	}

	// Chunk the envelope
	chunks, err := envelope.ChunkEnvelope(env, budget)
	if err != nil {
		return nil, fmt.Errorf("failed to chunk envelope: %w", err)
	}

	return chunks, nil
}

// ChunkingPublisher wraps broker publish operations with automatic chunking support
type ChunkingPublisher struct {
	helper   *ChunkingHelper
	publishFunc func(*envelope.Envelope) error
}

// NewChunkingPublisher creates a publisher that automatically chunks large envelopes
// publishFunc should be the actual broker publish function (e.g., client.PublishEnvelope)
func NewChunkingPublisher(counter tokencount.Counter, publishFunc func(*envelope.Envelope) error) *ChunkingPublisher {
	return &ChunkingPublisher{
		helper:      NewChunkingHelper(counter),
		publishFunc: publishFunc,
	}
}

// Publish publishes an envelope with automatic chunking if needed
// Large envelopes are automatically split into chunks before publishing
func (cp *ChunkingPublisher) Publish(env *envelope.Envelope) error {
	if cp.publishFunc == nil {
		return fmt.Errorf("publish function not configured")
	}

	// Prepare envelope(s) for publishing
	envelopes, err := cp.helper.PrepareForPublish(env)
	if err != nil {
		return fmt.Errorf("failed to prepare envelope: %w", err)
	}

	// Publish all envelopes (chunks or single envelope)
	for i, env := range envelopes {
		if err := cp.publishFunc(env); err != nil {
			return fmt.Errorf("failed to publish envelope %d: %w", i, err)
		}
	}

	return nil
}

// ProviderConfig maps agent destinations to their AI provider configurations
// This is used by the framework to determine which counter to use for chunking
type ProviderConfig struct {
	providers map[string]tokencount.Counter // destination -> counter
}

// NewProviderConfig creates a new provider configuration registry
func NewProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		providers: make(map[string]tokencount.Counter),
	}
}

// RegisterProvider associates a destination (agent ID or pattern) with a token counter
func (pc *ProviderConfig) RegisterProvider(destination string, counter tokencount.Counter) {
	pc.providers[destination] = counter
}

// GetCounter returns the token counter for a destination, or nil if not found
func (pc *ProviderConfig) GetCounter(destination string) tokencount.Counter {
	return pc.providers[destination]
}

// CreatePublisher creates a chunking publisher for a specific destination
func (pc *ProviderConfig) CreatePublisher(destination string, publishFunc func(*envelope.Envelope) error) *ChunkingPublisher {
	counter := pc.GetCounter(destination)
	return NewChunkingPublisher(counter, publishFunc)
}
