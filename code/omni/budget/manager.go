package budget

import (
	"fmt"
	"math"
	"strings"

	"github.com/tenzoki/agen/omni/tokencount"
)

// Manager handles token budget calculations and input splitting
type Manager struct {
	counter    tokencount.Counter
	maxContext int // Model's max context window
	maxOutput  int // Model's max output tokens
}

// Budget represents the token allocation for a request
type Budget struct {
	SystemTokens    int  // Tokens used by system prompt
	ContextTokens   int  // Tokens used by conversation history
	InputTokens     int  // Tokens used by current input
	UsedTokens      int  // Total tokens used (system + context + input)
	AvailableTokens int  // Tokens available for output
	MaxOutputTokens int  // Maximum output tokens for this model
	NeedsSplitting  bool // Whether input needs to be split
	SuggestedChunks int  // Number of chunks if splitting is needed
}

// NewManager creates a new budget manager
func NewManager(counter tokencount.Counter) *Manager {
	return &Manager{
		counter:    counter,
		maxContext: counter.MaxContextWindow(),
		maxOutput:  counter.MaxOutputTokens(),
	}
}

// Calculate returns the current token budget for a request
// Algorithm:
// 1. Count system prompt tokens → S
// 2. Count context tokens (conversation history) → C
// 3. Count input tokens → I
// 4. Calculate used = S + C + I
// 5. Calculate available = maxContext - used - safetyMargin
// 6. If available < desiredOutput: needsSplitting = true
func (m *Manager) Calculate(system, context, input string) (*Budget, error) {
	// Count tokens for each component
	systemTokens, err := m.counter.Count(system)
	if err != nil {
		return nil, fmt.Errorf("failed to count system tokens: %w", err)
	}

	contextTokens, err := m.counter.Count(context)
	if err != nil {
		return nil, fmt.Errorf("failed to count context tokens: %w", err)
	}

	inputTokens, err := m.counter.Count(input)
	if err != nil {
		return nil, fmt.Errorf("failed to count input tokens: %w", err)
	}

	// Calculate total used tokens
	usedTokens := systemTokens + contextTokens + inputTokens

	// Reserve tokens for safety margin
	reserveTokens := m.counter.ReserveTokens()

	// Calculate available tokens for output
	// available = maxContext - used - reserve
	availableTokens := m.maxContext - usedTokens - reserveTokens

	// Determine if we need splitting
	needsSplitting := false
	suggestedChunks := 1

	// If available tokens < desired output, we need splitting
	if availableTokens < m.maxOutput {
		needsSplitting = true

		// Calculate how many chunks we need
		// Each chunk should fit: system + context + chunk_input + desired_output + reserve
		maxInputPerChunk := m.maxContext - systemTokens - contextTokens - m.maxOutput - reserveTokens

		if maxInputPerChunk <= 0 {
			return nil, fmt.Errorf("cannot fit input: system+context alone exceed token limit (need %d tokens, have %d max context)",
				systemTokens+contextTokens+m.maxOutput+reserveTokens, m.maxContext)
		}

		suggestedChunks = int(math.Ceil(float64(inputTokens) / float64(maxInputPerChunk)))
	}

	return &Budget{
		SystemTokens:    systemTokens,
		ContextTokens:   contextTokens,
		InputTokens:     inputTokens,
		UsedTokens:      usedTokens,
		AvailableTokens: availableTokens,
		MaxOutputTokens: m.maxOutput,
		NeedsSplitting:  needsSplitting,
		SuggestedChunks: suggestedChunks,
	}, nil
}

// SplitInput splits input into chunks that fit within budget
// This is a basic text splitting strategy - can be enhanced with:
// - Code-aware splitting (AST boundaries)
// - Document structure awareness
// - Context overlap for continuity
func (m *Manager) SplitInput(input string, budget *Budget) ([]string, error) {
	if !budget.NeedsSplitting {
		return []string{input}, nil
	}

	if budget.SuggestedChunks <= 0 {
		return nil, fmt.Errorf("invalid suggested chunks: %d", budget.SuggestedChunks)
	}

	// Calculate target tokens per chunk
	targetTokensPerChunk := int(math.Ceil(float64(budget.InputTokens) / float64(budget.SuggestedChunks)))

	// Split by paragraphs first (better for readability)
	paragraphs := strings.Split(input, "\n\n")

	var chunks []string
	var currentChunk strings.Builder
	currentTokens := 0

	for _, para := range paragraphs {
		paraTokens, err := m.counter.Count(para)
		if err != nil {
			return nil, fmt.Errorf("failed to count paragraph tokens: %w", err)
		}

		// If single paragraph exceeds target, split it
		if paraTokens > targetTokensPerChunk {
			// Flush current chunk if not empty
			if currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				currentTokens = 0
			}

			// Split large paragraph by sentences
			sentences := splitBySentences(para)
			for _, sentence := range sentences {
				sentTokens, err := m.counter.Count(sentence)
				if err != nil {
					return nil, fmt.Errorf("failed to count sentence tokens: %w", err)
				}

				if currentTokens+sentTokens > targetTokensPerChunk && currentChunk.Len() > 0 {
					chunks = append(chunks, currentChunk.String())
					currentChunk.Reset()
					currentTokens = 0
				}

				if currentChunk.Len() > 0 {
					currentChunk.WriteString(" ")
				}
				currentChunk.WriteString(sentence)
				currentTokens += sentTokens
			}
		} else {
			// Check if adding this paragraph exceeds target
			if currentTokens+paraTokens > targetTokensPerChunk && currentChunk.Len() > 0 {
				chunks = append(chunks, currentChunk.String())
				currentChunk.Reset()
				currentTokens = 0
			}

			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n\n")
			}
			currentChunk.WriteString(para)
			currentTokens += paraTokens
		}
	}

	// Add remaining chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	// If we ended up with no chunks (edge case), return original input
	if len(chunks) == 0 {
		return []string{input}, nil
	}

	return chunks, nil
}

// splitBySentences splits text into sentences
// Simple implementation - can be enhanced with better sentence detection
func splitBySentences(text string) []string {
	// Split by common sentence terminators
	text = strings.ReplaceAll(text, ". ", ".\n")
	text = strings.ReplaceAll(text, "! ", "!\n")
	text = strings.ReplaceAll(text, "? ", "?\n")

	sentences := strings.Split(text, "\n")

	// Filter empty sentences
	var result []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}

	return result
}
