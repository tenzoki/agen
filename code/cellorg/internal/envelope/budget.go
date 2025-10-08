package envelope

import (
	"fmt"

	"github.com/tenzoki/agen/omni/tokencount"
)

// EnvelopeBudget represents token budget analysis for an envelope
type EnvelopeBudget struct {
	// Token counts
	PayloadTokens   int // Tokens in the payload
	HeaderTokens    int // Estimated tokens for headers/metadata
	TotalTokens     int // Total tokens (payload + headers)

	// Budget analysis
	NeedsSplitting  bool // Whether the envelope needs chunking
	SuggestedChunks int  // Number of chunks recommended

	// Model limits (from counter)
	MaxContextWindow int // Model's max context window
	MaxOutputTokens  int // Model's max output tokens
	AvailableTokens  int // Tokens available after this envelope
}

// CalculateBudget estimates token usage for an envelope
// This function determines if the envelope payload fits within the target model's limits
func CalculateBudget(env *Envelope, counter tokencount.Counter) (*EnvelopeBudget, error) {
	// Count payload tokens
	payloadStr := string(env.Payload)
	payloadTokens, err := counter.Count(payloadStr)
	if err != nil {
		return nil, fmt.Errorf("failed to count payload tokens: %w", err)
	}

	// Estimate metadata tokens (conservative estimate)
	headerTokens := estimateMetadataTokens(env)

	// Total tokens for this envelope
	totalTokens := payloadTokens + headerTokens

	// Get model limits
	maxContext := counter.MaxContextWindow()
	maxOutput := counter.MaxOutputTokens()

	// Reserve space for output and safety margin
	reserveTokens := counter.ReserveTokens()
	requiredSpace := maxOutput + reserveTokens

	// Check if splitting is needed
	needsSplitting := totalTokens > (maxContext - requiredSpace)
	suggestedChunks := 1

	if needsSplitting {
		// Calculate how many chunks we need
		maxPayloadPerChunk := maxContext - headerTokens - requiredSpace
		if maxPayloadPerChunk <= 0 {
			return nil, fmt.Errorf("cannot fit payload: headers alone exceed available space")
		}

		// Round up division
		suggestedChunks = (payloadTokens + maxPayloadPerChunk - 1) / maxPayloadPerChunk
		if suggestedChunks < 2 {
			suggestedChunks = 2 // Minimum 2 chunks if splitting
		}
	}

	return &EnvelopeBudget{
		PayloadTokens:    payloadTokens,
		HeaderTokens:     headerTokens,
		TotalTokens:      totalTokens,
		NeedsSplitting:   needsSplitting,
		SuggestedChunks:  suggestedChunks,
		MaxContextWindow: maxContext,
		MaxOutputTokens:  maxOutput,
		AvailableTokens:  maxContext - totalTokens - requiredSpace,
	}, nil
}

// estimateMetadataTokens provides a conservative estimate of tokens used by envelope metadata
// This includes routing info, headers, tracing data, etc.
func estimateMetadataTokens(env *Envelope) int {
	// Conservative estimate based on typical envelope structure:
	// - ID, CorrelationID, TraceID, SpanID: ~150 tokens (UUIDs + field names)
	// - Source, Destination, MessageType: ~30 tokens
	// - Timestamp, TTL, Sequence, HopCount: ~20 tokens
	// - Headers (string map): ~10 tokens per header
	// - Properties (interface map): ~15 tokens per property
	// - Route array: ~10 tokens per hop

	baseTokens := 200 // Base envelope structure

	// Add header tokens
	headerTokens := len(env.Headers) * 10

	// Add properties tokens
	propertyTokens := len(env.Properties) * 15

	// Add route tokens
	routeTokens := len(env.Route) * 10

	return baseTokens + headerTokens + propertyTokens + routeTokens
}
