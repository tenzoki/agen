package tokencount

// Counter estimates tokens for different providers
// This interface abstracts token counting across OpenAI, Anthropic, and other providers
type Counter interface {
	// Count estimates tokens in a single text string
	Count(text string) (int, error)

	// CountMessages estimates tokens for a conversation
	CountMessages(messages []Message) (int, error)

	// MaxContextWindow returns the maximum context window for this model
	MaxContextWindow() int

	// MaxOutputTokens returns the maximum output tokens for this model
	MaxOutputTokens() int

	// ReserveTokens returns the safety margin (typically 5-10% of max)
	ReserveTokens() int

	// Provider returns the provider name (e.g., "openai", "anthropic")
	Provider() string

	// Model returns the model name (e.g., "gpt-5", "claude-3-5-sonnet")
	Model() string
}

// Message represents a chat message for token counting
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// Config holds configuration for token counter creation
type Config struct {
	Provider     string  // "openai", "anthropic", "auto"
	Model        string  // Model identifier
	SafetyMargin float64 // Percentage of tokens to reserve (0.05 = 5%)
	CacheDir     string  // Directory for tokenizer cache (optional)
}

// NewCounter creates a token counter for the specified provider and model
func NewCounter(cfg Config) (Counter, error) {
	// Set defaults
	if cfg.SafetyMargin == 0 {
		cfg.SafetyMargin = 0.10 // 10% default safety margin
	}

	switch cfg.Provider {
	case "openai":
		return newOpenAICounter(cfg)
	case "anthropic":
		return newAnthropicCounter(cfg)
	default:
		return newFallbackCounter(cfg)
	}
}
