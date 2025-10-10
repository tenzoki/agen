package ai

import (
	"context"
	"time"
)

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", or "system"
	Content string `json:"content"` // The message content
}

// Response represents the LLM's response
type Response struct {
	Content      string        // The generated text
	Model        string        // Model used for generation
	StopReason   string        // Why generation stopped (e.g., "end_turn", "max_tokens")
	Usage        Usage         // Token usage stats
	FinishTime   time.Time     // When response completed
	ResponseTime time.Duration // How long the request took
}

// Usage tracks token consumption
type Usage struct {
	InputTokens  int // Tokens in the prompt
	OutputTokens int // Tokens in the response
	TotalTokens  int // Total tokens used
}

// Config holds configuration for an LLM client
type Config struct {
	APIKey      string        `json:"api_key"`      // API authentication key
	Model       string        `json:"model"`        // Model identifier (e.g., "claude-3-5-sonnet-20241022")
	MaxTokens   int           `json:"max_tokens"`   // Maximum tokens to generate
	Temperature float64       `json:"temperature"`  // Sampling temperature (0.0-1.0)
	Timeout     time.Duration `json:"timeout"`      // Request timeout
	RetryCount  int           `json:"retry_count"`  // Number of retries on failure
	RetryDelay  time.Duration `json:"retry_delay"`  // Delay between retries
}

// LLM defines the interface for interacting with language models
type LLM interface {
	// Chat sends a conversation and returns the model's response
	Chat(ctx context.Context, messages []Message) (*Response, error)

	// ChatStream sends a conversation and streams the response
	// The returned channel will emit response chunks as they arrive
	ChatStream(ctx context.Context, messages []Message) (<-chan string, <-chan error)

	// Model returns the model identifier being used
	Model() string

	// Provider returns the provider name (e.g., "anthropic", "openai")
	Provider() string
}

// Error types for AI layer
type Error struct {
	Provider string // Which provider caused the error
	Code     string // Error code from provider
	Message  string // Human-readable error message
	Retry    bool   // Whether the error is retryable
}

func (e *Error) Error() string {
	return e.Provider + ": " + e.Message
}