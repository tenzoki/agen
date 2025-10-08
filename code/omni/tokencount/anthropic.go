package tokencount

import "strings"

// anthropicCounter implements Counter for Anthropic Claude models using heuristic
type anthropicCounter struct {
	model        string
	safetyMargin float64
	limits       modelLimits
}

// Anthropic model limits (as of 2025)
var anthropicLimits = map[string]modelLimits{
	// Claude 4.x series (2025)
	"claude-sonnet-4-5-20250929": {
		contextWindow: 200000,
		maxOutput:     64000,
	},
	"claude-opus-4-1-20250805": {
		contextWindow: 200000,
		maxOutput:     32000,
	},
	"claude-sonnet-4-20250514": {
		contextWindow: 200000,
		maxOutput:     64000,
	},
	// Claude 3.5 series (2024)
	"claude-3-5-sonnet-20241022": {
		contextWindow: 200000,
		maxOutput:     8192,
	},
	"claude-3-5-haiku-20241022": {
		contextWindow: 200000,
		maxOutput:     8192,
	},
	// Claude 3 series (legacy)
	"claude-3-opus-20240229": {
		contextWindow: 200000,
		maxOutput:     4096,
	},
	"claude-3-sonnet-20240229": {
		contextWindow: 200000,
		maxOutput:     4096,
	},
	"claude-3-haiku-20240307": {
		contextWindow: 200000,
		maxOutput:     4096,
	},
}

func newAnthropicCounter(cfg Config) (Counter, error) {
	// Get model limits
	limits, ok := anthropicLimits[cfg.Model]
	if !ok {
		// Default limits for unknown Claude models
		limits = modelLimits{
			contextWindow: 200000,
			maxOutput:     4096,
		}
	}

	return &anthropicCounter{
		model:        cfg.Model,
		safetyMargin: cfg.SafetyMargin,
		limits:       limits,
	}, nil
}

// Count uses heuristic: text length / 3.5 (Anthropic recommendation)
func (a *anthropicCounter) Count(text string) (int, error) {
	// Anthropic's heuristic: divide character count by 3.5
	charCount := len(text)
	tokens := int(float64(charCount) / 3.5)
	return tokens, nil
}

func (a *anthropicCounter) CountMessages(messages []Message) (int, error) {
	total := 0

	for _, msg := range messages {
		// Count role overhead (approximately 10 tokens per message for Claude)
		total += 10

		// Count message content using heuristic
		charCount := len(msg.Content)
		tokens := int(float64(charCount) / 3.5)
		total += tokens
	}

	// Add overhead for message formatting (approximately 5 tokens)
	total += 5

	return total, nil
}

func (a *anthropicCounter) MaxContextWindow() int {
	return a.limits.contextWindow
}

func (a *anthropicCounter) MaxOutputTokens() int {
	return a.limits.maxOutput
}

func (a *anthropicCounter) ReserveTokens() int {
	return int(float64(a.limits.contextWindow) * a.safetyMargin)
}

func (a *anthropicCounter) Provider() string {
	return "anthropic"
}

func (a *anthropicCounter) Model() string {
	return a.model
}

// countWords is a utility for more refined estimation (not currently used)
func countWords(text string) int {
	return len(strings.Fields(text))
}
