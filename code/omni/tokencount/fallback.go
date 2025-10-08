package tokencount

// fallbackCounter implements Counter using conservative estimation
// Used when provider is unknown or unsupported
type fallbackCounter struct {
	model        string
	provider     string
	safetyMargin float64
}

func newFallbackCounter(cfg Config) (Counter, error) {
	return &fallbackCounter{
		model:        cfg.Model,
		provider:     cfg.Provider,
		safetyMargin: cfg.SafetyMargin + 0.10, // Extra 10% margin for unknown models
	}, nil
}

// Count uses conservative estimation: text length / 4
// This is more conservative than Anthropic's 3.5 to account for unknown tokenizers
func (f *fallbackCounter) Count(text string) (int, error) {
	charCount := len(text)
	tokens := int(float64(charCount) / 4.0)
	return tokens, nil
}

func (f *fallbackCounter) CountMessages(messages []Message) (int, error) {
	total := 0

	for _, msg := range messages {
		// Conservative role overhead (15 tokens per message)
		total += 15

		// Count message content using conservative heuristic
		charCount := len(msg.Content)
		tokens := int(float64(charCount) / 4.0)
		total += tokens
	}

	// Conservative message formatting overhead (10 tokens)
	total += 10

	return total, nil
}

func (f *fallbackCounter) MaxContextWindow() int {
	// Conservative default: 128K context
	return 128000
}

func (f *fallbackCounter) MaxOutputTokens() int {
	// Conservative default: 4K output
	return 4096
}

func (f *fallbackCounter) ReserveTokens() int {
	return int(float64(f.MaxContextWindow()) * f.safetyMargin)
}

func (f *fallbackCounter) Provider() string {
	return f.provider
}

func (f *fallbackCounter) Model() string {
	return f.model
}
