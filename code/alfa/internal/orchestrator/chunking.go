package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/tenzoki/agen/alfa/internal/ai"
	"github.com/tenzoki/agen/omni/budget"
	"github.com/tenzoki/agen/omni/tokencount"
)

// ChunkProcessor handles large requests by splitting them into manageable chunks
type ChunkProcessor struct {
	llm    ai.LLM
	budget *budget.Manager
	merger *ResponseMerger
}

// ChunkConfig holds configuration for chunk processing
type ChunkConfig struct {
	Enabled         bool
	OverlapRatio    float64 // Overlap between chunks for context continuity
	AutoDeduplicate bool
	ChunkStrategy   string // "auto", "code", "text"
}

// NewChunkProcessor creates a new chunk processor
func NewChunkProcessor(llm ai.LLM, counter tokencount.Counter) *ChunkProcessor {
	budgetMgr := budget.NewManager(counter)
	merger := NewResponseMerger(llm)

	return &ChunkProcessor{
		llm:    llm,
		budget: budgetMgr,
		merger: merger,
	}
}

// ProcessWithChunking handles large requests intelligently
// Algorithm:
// 1. Calculate token budget
// 2. If splitting needed: split input, process each chunk, merge responses
// 3. Else: process as normal
func (cp *ChunkProcessor) ProcessWithChunking(
	ctx context.Context,
	system, context, input string,
) (string, error) {
	// Calculate budget
	budgetInfo, err := cp.budget.Calculate(system, context, input)
	if err != nil {
		return "", fmt.Errorf("failed to calculate budget: %w", err)
	}

	// If no splitting needed, process normally
	if !budgetInfo.NeedsSplitting {
		return cp.processSingle(ctx, system, context, input)
	}

	// Split input into chunks
	chunks, err := cp.budget.SplitInput(input, budgetInfo)
	if err != nil {
		return "", fmt.Errorf("failed to split input: %w", err)
	}

	// Process each chunk
	responses := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		// Add chunk context to system prompt
		chunkSystem := fmt.Sprintf("%s\n\n[Processing chunk %d of %d]", system, i+1, len(chunks))

		response, err := cp.processSingle(ctx, chunkSystem, context, chunk)
		if err != nil {
			return "", fmt.Errorf("failed to process chunk %d: %w", i+1, err)
		}

		responses = append(responses, response)
	}

	// Merge responses
	merged, err := cp.merger.Merge(responses, ContentTypeAuto)
	if err != nil {
		return "", fmt.Errorf("failed to merge responses: %w", err)
	}

	// Deduplicate if needed
	deduplicated, err := cp.merger.Deduplicate(merged, ContentTypeAuto)
	if err != nil {
		return "", fmt.Errorf("failed to deduplicate response: %w", err)
	}

	return deduplicated, nil
}

// processSingle handles a single (non-chunked) request
func (cp *ChunkProcessor) processSingle(ctx context.Context, system, context, input string) (string, error) {
	// Build messages
	messages := []ai.Message{}

	if system != "" {
		messages = append(messages, ai.Message{
			Role:    "system",
			Content: system,
		})
	}

	if context != "" {
		messages = append(messages, ai.Message{
			Role:    "user",
			Content: context,
		})
	}

	messages = append(messages, ai.Message{
		Role:    "user",
		Content: input,
	})

	// Send to LLM
	response, err := cp.llm.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("LLM chat failed: %w", err)
	}

	return response.Content, nil
}

// GetBudgetInfo returns budget information for a request (useful for debugging)
func (cp *ChunkProcessor) GetBudgetInfo(system, context, input string) (*budget.Budget, error) {
	return cp.budget.Calculate(system, context, input)
}

// AddOverlap adds context overlap between chunks for better continuity
// This is useful when splitting text or code that requires context from previous chunks
func (cp *ChunkProcessor) AddOverlap(chunks []string, overlapRatio float64) []string {
	if len(chunks) <= 1 || overlapRatio <= 0 {
		return chunks
	}

	overlappedChunks := make([]string, len(chunks))

	for i, chunk := range chunks {
		if i == 0 {
			// First chunk: no prefix overlap
			overlappedChunks[i] = chunk
			continue
		}

		// Add overlap from previous chunk
		prevChunk := chunks[i-1]
		overlapSize := int(float64(len(prevChunk)) * overlapRatio)

		if overlapSize > len(prevChunk) {
			overlapSize = len(prevChunk)
		}

		// Take last N characters from previous chunk
		overlap := prevChunk[len(prevChunk)-overlapSize:]

		// Find a good break point (word boundary)
		if idx := strings.LastIndex(overlap, " "); idx > 0 {
			overlap = overlap[idx+1:]
		}

		overlappedChunks[i] = overlap + "\n...\n" + chunk
	}

	return overlappedChunks
}
