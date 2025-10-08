package tokencount

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

// openaiCounter implements Counter for OpenAI models using tiktoken
type openaiCounter struct {
	model        string
	encoding     *tiktoken.Tiktoken
	safetyMargin float64
	limits       modelLimits
}

// modelLimits holds token limits for a specific model
type modelLimits struct {
	contextWindow int
	maxOutput     int
}

// OpenAI model limits (as of 2025)
var openaiLimits = map[string]modelLimits{
	"gpt-5": {
		contextWindow: 272000,
		maxOutput:     128000,
	},
	"gpt-5-mini": {
		contextWindow: 272000,
		maxOutput:     128000,
	},
	"gpt-5-nano": {
		contextWindow: 272000,
		maxOutput:     128000,
	},
	"gpt-5-chat": {
		contextWindow: 272000,
		maxOutput:     128000,
	},
	"gpt-4o": {
		contextWindow: 128000,
		maxOutput:     16384,
	},
	"gpt-4o-mini": {
		contextWindow: 128000,
		maxOutput:     16384,
	},
	"o1": {
		contextWindow: 200000,
		maxOutput:     100000,
	},
	"o1-mini": {
		contextWindow: 128000,
		maxOutput:     65536,
	},
	"o1-preview": {
		contextWindow: 128000,
		maxOutput:     32768,
	},
	"gpt-4": {
		contextWindow: 8192,
		maxOutput:     4096,
	},
	"gpt-4-turbo": {
		contextWindow: 128000,
		maxOutput:     4096,
	},
}

func newOpenAICounter(cfg Config) (Counter, error) {
	// Get encoding for model
	var encodingName string
	switch cfg.Model {
	case "gpt-5", "gpt-5-mini", "gpt-5-nano", "gpt-5-chat",
		"gpt-4o", "gpt-4o-mini",
		"o1", "o1-mini", "o1-preview":
		encodingName = "o200k_base" // New encoding for GPT-4o and later
	case "gpt-4", "gpt-4-turbo", "gpt-3.5-turbo":
		encodingName = "cl100k_base" // GPT-4 encoding
	default:
		encodingName = "o200k_base" // Default to newest
	}

	encoding, err := tiktoken.GetEncoding(encodingName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoding: %w", err)
	}

	// Get model limits
	limits, ok := openaiLimits[cfg.Model]
	if !ok {
		// Default limits for unknown models
		limits = modelLimits{
			contextWindow: 128000,
			maxOutput:     4096,
		}
	}

	return &openaiCounter{
		model:        cfg.Model,
		encoding:     encoding,
		safetyMargin: cfg.SafetyMargin,
		limits:       limits,
	}, nil
}

func (o *openaiCounter) Count(text string) (int, error) {
	tokens := o.encoding.Encode(text, nil, nil)
	return len(tokens), nil
}

func (o *openaiCounter) CountMessages(messages []Message) (int, error) {
	total := 0

	for _, msg := range messages {
		// Count role overhead (approximately 4 tokens per message)
		total += 4

		// Count message content
		tokens := o.encoding.Encode(msg.Content, nil, nil)
		total += len(tokens)
	}

	// Add overhead for message formatting (approximately 3 tokens)
	total += 3

	return total, nil
}

func (o *openaiCounter) MaxContextWindow() int {
	return o.limits.contextWindow
}

func (o *openaiCounter) MaxOutputTokens() int {
	return o.limits.maxOutput
}

func (o *openaiCounter) ReserveTokens() int {
	return int(float64(o.limits.contextWindow) * o.safetyMargin)
}

func (o *openaiCounter) Provider() string {
	return "openai"
}

func (o *openaiCounter) Model() string {
	return o.model
}
