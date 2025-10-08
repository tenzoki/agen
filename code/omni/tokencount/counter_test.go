package tokencount

import (
	"testing"
)

// Test OpenAI counter with known samples
func TestOpenAICounter(t *testing.T) {
	cfg := Config{
		Provider:     "openai",
		Model:        "gpt-5",
		SafetyMargin: 0.10,
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create OpenAI counter: %v", err)
	}

	// Test simple text counting
	text := "Hello, world!"
	tokens, err := counter.Count(text)
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}
	if tokens < 2 || tokens > 5 {
		t.Errorf("Expected 2-5 tokens for 'Hello, world!', got %d", tokens)
	}

	// Test longer text
	longText := "The quick brown fox jumps over the lazy dog. " +
		"This is a test of token counting accuracy."
	tokens, err = counter.Count(longText)
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}
	// Rough estimate: ~20-30 tokens
	if tokens < 15 || tokens > 40 {
		t.Errorf("Expected 15-40 tokens for long text, got %d", tokens)
	}

	// Verify model info
	if counter.Provider() != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", counter.Provider())
	}
	if counter.Model() != "gpt-5" {
		t.Errorf("Expected model 'gpt-5', got '%s'", counter.Model())
	}
}

// Test GPT-5 model limits
func TestGPT5Limits(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Verify GPT-5 limits
	if counter.MaxContextWindow() != 272000 {
		t.Errorf("Expected 272000 context window, got %d", counter.MaxContextWindow())
	}
	if counter.MaxOutputTokens() != 128000 {
		t.Errorf("Expected 128000 max output, got %d", counter.MaxOutputTokens())
	}

	// Verify safety margin (10% of 272000 = 27200)
	reserve := counter.ReserveTokens()
	expected := int(float64(272000) * 0.10)
	if reserve != expected {
		t.Errorf("Expected reserve tokens %d, got %d", expected, reserve)
	}
}

// Test message counting with conversation
func TestOpenAIMessageCounting(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "What is 2+2?"},
		{Role: "assistant", Content: "The answer is 4."},
	}

	tokens, err := counter.CountMessages(messages)
	if err != nil {
		t.Errorf("CountMessages failed: %v", err)
	}

	// Expected: ~20-30 tokens (content + overhead)
	if tokens < 15 || tokens > 40 {
		t.Errorf("Expected 15-40 tokens for message sequence, got %d", tokens)
	}

	// Verify overhead is included (4 tokens per message + 3 for formatting)
	minOverhead := 4*3 + 3
	if tokens < minOverhead {
		t.Errorf("Expected at least %d tokens (overhead), got %d", minOverhead, tokens)
	}
}

// Test Anthropic counter with heuristic
func TestAnthropicCounter(t *testing.T) {
	cfg := Config{
		Provider:     "anthropic",
		Model:        "claude-3-5-sonnet-20241022",
		SafetyMargin: 0.10,
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create Anthropic counter: %v", err)
	}

	// Test with known character count
	text := "12345678901234567890" // 20 characters
	tokens, err := counter.Count(text)
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}

	// Expected: 20 / 3.5 = ~5.7 = 5 tokens
	floatResult := float64(20) / 3.5
	expected := int(floatResult)
	if tokens != expected {
		t.Errorf("Expected %d tokens (20 chars / 3.5), got %d", expected, tokens)
	}

	// Verify model info
	if counter.Provider() != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got '%s'", counter.Provider())
	}
	if counter.Model() != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", counter.Model())
	}
}

// Test Anthropic model limits
func TestAnthropicLimits(t *testing.T) {
	cfg := Config{
		Provider: "anthropic",
		Model:    "claude-3-5-sonnet-20241022",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Verify Claude limits
	if counter.MaxContextWindow() != 200000 {
		t.Errorf("Expected 200000 context window, got %d", counter.MaxContextWindow())
	}
	if counter.MaxOutputTokens() != 8192 {
		t.Errorf("Expected 8192 max output, got %d", counter.MaxOutputTokens())
	}
}

// Test Anthropic message counting
func TestAnthropicMessageCounting(t *testing.T) {
	cfg := Config{
		Provider: "anthropic",
		Model:    "claude-3-opus-20240229",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	messages := []Message{
		{Role: "system", Content: "You are helpful."}, // 17 chars
		{Role: "user", Content: "Hi"},                 // 2 chars
	}

	tokens, err := counter.CountMessages(messages)
	if err != nil {
		t.Errorf("CountMessages failed: %v", err)
	}

	// Expected: (17+2)/3.5 = ~5 + overhead (10*2 + 5 = 25) = ~30
	floatContent := float64(17+2) / 3.5
	expectedContent := int(floatContent)
	expectedOverhead := 10*2 + 5
	expectedTotal := expectedContent + expectedOverhead

	if tokens < expectedTotal-5 || tokens > expectedTotal+5 {
		t.Errorf("Expected ~%d tokens, got %d", expectedTotal, tokens)
	}
}

// Test fallback counter
func TestFallbackCounter(t *testing.T) {
	cfg := Config{
		Provider:     "unknown",
		Model:        "mystery-model",
		SafetyMargin: 0.10,
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create fallback counter: %v", err)
	}

	// Test with known character count
	text := "12345678901234567890" // 20 characters
	tokens, err := counter.Count(text)
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}

	// Expected: 20 / 4 = 5 tokens (conservative)
	expected := int(float64(20) / 4.0)
	if tokens != expected {
		t.Errorf("Expected %d tokens (20 chars / 4), got %d", expected, tokens)
	}

	// Verify conservative defaults
	if counter.MaxContextWindow() != 128000 {
		t.Errorf("Expected 128000 context window, got %d", counter.MaxContextWindow())
	}
	if counter.MaxOutputTokens() != 4096 {
		t.Errorf("Expected 4096 max output, got %d", counter.MaxOutputTokens())
	}

	// Fallback has extra 10% safety margin (20% total)
	reserve := counter.ReserveTokens()
	expected = int(float64(128000) * 0.20) // 0.10 + 0.10
	if reserve != expected {
		t.Errorf("Expected reserve tokens %d, got %d", expected, reserve)
	}
}

// Test fallback message counting
func TestFallbackMessageCounting(t *testing.T) {
	cfg := Config{
		Provider: "custom",
		Model:    "custom-model",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	messages := []Message{
		{Role: "user", Content: "Test"}, // 4 chars
	}

	tokens, err := counter.CountMessages(messages)
	if err != nil {
		t.Errorf("CountMessages failed: %v", err)
	}

	// Expected: 4/4 = 1 + overhead (15 + 10) = 26
	expectedContent := int(float64(4) / 4.0)
	expectedOverhead := 15 + 10
	expectedTotal := expectedContent + expectedOverhead

	if tokens != expectedTotal {
		t.Errorf("Expected %d tokens, got %d", expectedTotal, tokens)
	}
}

// Test default safety margin
func TestDefaultSafetyMargin(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		Model:    "gpt-5",
		// SafetyMargin not set - should default to 0.10
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	reserve := counter.ReserveTokens()
	expected := int(float64(272000) * 0.10)
	if reserve != expected {
		t.Errorf("Expected default 10%% safety margin (%d tokens), got %d", expected, reserve)
	}
}

// Test custom safety margin
func TestCustomSafetyMargin(t *testing.T) {
	cfg := Config{
		Provider:     "openai",
		Model:        "gpt-5",
		SafetyMargin: 0.15, // 15%
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	reserve := counter.ReserveTokens()
	expected := int(float64(272000) * 0.15)
	if reserve != expected {
		t.Errorf("Expected 15%% safety margin (%d tokens), got %d", expected, reserve)
	}
}

// Test o1 model recognition
func TestO1ModelLimits(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		Model:    "o1",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	if counter.MaxContextWindow() != 200000 {
		t.Errorf("Expected 200000 context window for o1, got %d", counter.MaxContextWindow())
	}
	if counter.MaxOutputTokens() != 100000 {
		t.Errorf("Expected 100000 max output for o1, got %d", counter.MaxOutputTokens())
	}
}

// Test unknown OpenAI model falls back to defaults
func TestUnknownOpenAIModel(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		Model:    "gpt-99",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Should use default OpenAI limits
	if counter.MaxContextWindow() != 128000 {
		t.Errorf("Expected default 128000 context window, got %d", counter.MaxContextWindow())
	}
	if counter.MaxOutputTokens() != 4096 {
		t.Errorf("Expected default 4096 max output, got %d", counter.MaxOutputTokens())
	}
}

// Test empty text
func TestEmptyText(t *testing.T) {
	cfg := Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	tokens, err := counter.Count("")
	if err != nil {
		t.Errorf("Count failed for empty text: %v", err)
	}
	if tokens != 0 {
		t.Errorf("Expected 0 tokens for empty text, got %d", tokens)
	}
}

// Test empty message list
func TestEmptyMessages(t *testing.T) {
	cfg := Config{
		Provider: "anthropic",
		Model:    "claude-3-opus-20240229",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	messages := []Message{}
	tokens, err := counter.CountMessages(messages)
	if err != nil {
		t.Errorf("CountMessages failed for empty list: %v", err)
	}

	// Should only have formatting overhead (5 tokens)
	if tokens != 5 {
		t.Errorf("Expected 5 tokens (formatting overhead only), got %d", tokens)
	}
}

// Benchmark OpenAI counting
func BenchmarkOpenAICount(b *testing.B) {
	cfg := Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		b.Fatalf("Failed to create counter: %v", err)
	}

	text := "The quick brown fox jumps over the lazy dog."
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = counter.Count(text)
	}
}

// Benchmark Anthropic counting
func BenchmarkAnthropicCount(b *testing.B) {
	cfg := Config{
		Provider: "anthropic",
		Model:    "claude-3-5-sonnet-20241022",
	}

	counter, err := NewCounter(cfg)
	if err != nil {
		b.Fatalf("Failed to create counter: %v", err)
	}

	text := "The quick brown fox jumps over the lazy dog."
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = counter.Count(text)
	}
}
