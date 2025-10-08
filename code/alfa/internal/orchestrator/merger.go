package orchestrator

import (
	"context"
	"fmt"
	"strings"

	"github.com/tenzoki/agen/alfa/internal/ai"
)

// ContentType indicates the type of content being merged
type ContentType int

const (
	ContentTypeAuto ContentType = iota
	ContentTypeCode
	ContentTypeText
	ContentTypeAnalysis
)

// ResponseMerger handles merging and deduplication of chunked responses
type ResponseMerger struct {
	llm ai.LLM
}

// NewResponseMerger creates a new response merger
func NewResponseMerger(llm ai.LLM) *ResponseMerger {
	return &ResponseMerger{
		llm: llm,
	}
}

// Merge combines multiple responses intelligently based on content type
func (m *ResponseMerger) Merge(responses []string, contentType ContentType) (string, error) {
	if len(responses) == 0 {
		return "", fmt.Errorf("no responses to merge")
	}

	if len(responses) == 1 {
		return responses[0], nil
	}

	// Detect content type if auto
	if contentType == ContentTypeAuto {
		contentType = m.detectContentType(responses[0])
	}

	switch contentType {
	case ContentTypeCode:
		return m.mergeCode(responses), nil
	case ContentTypeText:
		return m.mergeText(responses), nil
	case ContentTypeAnalysis:
		return m.mergeAnalysis(responses), nil
	default:
		return m.mergeText(responses), nil
	}
}

// Deduplicate removes redundancies using AI
func (m *ResponseMerger) Deduplicate(merged string, contentType ContentType) (string, error) {
	// Detect content type if auto
	if contentType == ContentTypeAuto {
		contentType = m.detectContentType(merged)
	}

	// Build deduplication prompt
	prompt := m.buildDeduplicationPrompt(merged, contentType)

	// Send to LLM
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a text deduplication assistant. Remove redundancies while preserving unique information.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := m.llm.Chat(context.Background(), messages)
	if err != nil {
		// If deduplication fails, return original
		return merged, fmt.Errorf("deduplication failed: %w", err)
	}

	return response.Content, nil
}

// detectContentType infers content type from the text
func (m *ResponseMerger) detectContentType(text string) ContentType {
	// Check for code markers
	if strings.Contains(text, "```") ||
		strings.Contains(text, "func ") ||
		strings.Contains(text, "class ") ||
		strings.Contains(text, "def ") ||
		strings.Contains(text, "package ") ||
		strings.Contains(text, "import ") {
		return ContentTypeCode
	}

	// Check for analysis markers
	if strings.Contains(text, "Issue:") ||
		strings.Contains(text, "Finding:") ||
		strings.Contains(text, "Problem:") ||
		strings.Contains(text, "Analysis:") {
		return ContentTypeAnalysis
	}

	// Default to text
	return ContentTypeText
}

// mergeCode combines code responses
// Strategy:
// - Concatenate chunks
// - Remove duplicate imports
// - Preserve function/class definitions
func (m *ResponseMerger) mergeCode(responses []string) string {
	var merged strings.Builder

	// Track seen imports to deduplicate
	seenImports := make(map[string]bool)
	var imports []string
	var codeBlocks []string

	for _, response := range responses {
		// Extract imports and code separately
		lines := strings.Split(response, "\n")

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)

			// Check if it's an import line
			if strings.HasPrefix(trimmed, "import ") ||
				strings.HasPrefix(trimmed, "from ") ||
				strings.HasPrefix(trimmed, "using ") {

				if !seenImports[trimmed] {
					seenImports[trimmed] = true
					imports = append(imports, line)
				}
			} else {
				codeBlocks = append(codeBlocks, line)
			}
		}
	}

	// Combine: imports first, then code
	if len(imports) > 0 {
		merged.WriteString(strings.Join(imports, "\n"))
		merged.WriteString("\n\n")
	}

	merged.WriteString(strings.Join(codeBlocks, "\n"))

	return merged.String()
}

// mergeText combines text responses
// Strategy:
// - Linear concatenation
// - Remove repetitive introductions/conclusions
// - Smooth transitions
func (m *ResponseMerger) mergeText(responses []string) string {
	var merged strings.Builder

	for i, response := range responses {
		cleaned := m.cleanTextChunk(response, i == 0, i == len(responses)-1)

		if i > 0 {
			merged.WriteString("\n\n")
		}

		merged.WriteString(cleaned)
	}

	return merged.String()
}

// mergeAnalysis combines analysis responses
// Strategy:
// - Group by category
// - Remove duplicate findings
// - Aggregate statistics
func (m *ResponseMerger) mergeAnalysis(responses []string) string {
	var merged strings.Builder

	// Track seen issues to avoid duplicates
	seenIssues := make(map[string]bool)

	for i, response := range responses {
		if i > 0 {
			merged.WriteString("\n\n---\n\n")
		}

		merged.WriteString(fmt.Sprintf("## Analysis Part %d\n\n", i+1))

		// Split into findings
		findings := strings.Split(response, "\n")

		for _, finding := range findings {
			trimmed := strings.TrimSpace(finding)
			if trimmed == "" {
				continue
			}

			// Simple deduplication based on content similarity
			normalized := strings.ToLower(trimmed)
			if !seenIssues[normalized] {
				seenIssues[normalized] = true
				merged.WriteString(finding)
				merged.WriteString("\n")
			}
		}
	}

	return merged.String()
}

// cleanTextChunk removes redundant introductions/conclusions
// For now, we keep this simple and just return the text as-is
// More aggressive cleaning can be done via AI deduplication
func (m *ResponseMerger) cleanTextChunk(text string, isFirst, isLast bool) string {
	// Simple implementation: just return the text
	// The AI deduplication step will handle removing redundancies
	return text
}

// buildDeduplicationPrompt creates the AI prompt for deduplication
func (m *ResponseMerger) buildDeduplicationPrompt(merged string, contentType ContentType) string {
	contentTypeName := "text"
	switch contentType {
	case ContentTypeCode:
		contentTypeName = "code"
	case ContentTypeAnalysis:
		contentTypeName = "analysis"
	}

	return fmt.Sprintf(`You are receiving a merged response from multiple AI calls.
Remove redundancies and duplications while preserving unique information.

Content Type: %s

Merged content:
%s

Output the deduplicated version only. Do not add any explanations or commentary.`, contentTypeName, merged)
}
