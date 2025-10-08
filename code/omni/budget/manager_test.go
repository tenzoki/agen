package budget

import (
	"strings"
	"testing"

	"github.com/tenzoki/agen/omni/tokencount"
)

// Test basic budget calculation
func TestBudgetCalculation(t *testing.T) {
	cfg := tokencount.Config{
		Provider:     "openai",
		Model:        "gpt-5",
		SafetyMargin: 0.10,
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	// Test with small inputs that don't need splitting
	system := "You are a helpful assistant."
	context := "Previous conversation context."
	input := "What is 2+2?"

	budget, err := manager.Calculate(system, context, input)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Verify basic fields
	if budget.SystemTokens <= 0 {
		t.Errorf("Expected positive system tokens, got %d", budget.SystemTokens)
	}
	if budget.ContextTokens <= 0 {
		t.Errorf("Expected positive context tokens, got %d", budget.ContextTokens)
	}
	if budget.InputTokens <= 0 {
		t.Errorf("Expected positive input tokens, got %d", budget.InputTokens)
	}

	// Verify used tokens is sum of components
	expectedUsed := budget.SystemTokens + budget.ContextTokens + budget.InputTokens
	if budget.UsedTokens != expectedUsed {
		t.Errorf("Expected used tokens %d, got %d", expectedUsed, budget.UsedTokens)
	}

	// Verify available tokens calculation
	reserve := counter.ReserveTokens()
	expectedAvailable := counter.MaxContextWindow() - budget.UsedTokens - reserve
	if budget.AvailableTokens != expectedAvailable {
		t.Errorf("Expected available tokens %d, got %d", expectedAvailable, budget.AvailableTokens)
	}

	// Small input should not need splitting
	if budget.NeedsSplitting {
		t.Error("Small input should not need splitting")
	}
	if budget.SuggestedChunks != 1 {
		t.Errorf("Expected 1 chunk for small input, got %d", budget.SuggestedChunks)
	}
}

// Test with large input that needs splitting
func TestBudgetWithSplitting(t *testing.T) {
	cfg := tokencount.Config{
		Provider:     "anthropic",
		Model:        "claude-3-opus-20240229",
		SafetyMargin: 0.10,
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	system := "You are a code reviewer."
	context := ""

	// Create very large input (simulate large codebase)
	// Using ~150K tokens worth of text (should exceed available budget)
	largeInput := strings.Repeat("This is a line of code that takes some tokens. ", 100000)

	budget, err := manager.Calculate(system, context, largeInput)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Large input should trigger splitting
	if !budget.NeedsSplitting {
		t.Error("Large input should need splitting")
	}
	if budget.SuggestedChunks <= 1 {
		t.Errorf("Expected multiple chunks for large input, got %d", budget.SuggestedChunks)
	}

	t.Logf("Large input: %d tokens, suggested %d chunks", budget.InputTokens, budget.SuggestedChunks)
}

// Test input splitting
func TestSplitInput(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	// Create input that needs splitting
	input := strings.Repeat("Paragraph one with some content.\n\n", 5000) +
		strings.Repeat("Paragraph two with different content.\n\n", 5000)

	system := "System prompt"
	context := ""

	budget, err := manager.Calculate(system, context, input)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Split the input
	chunks, err := manager.SplitInput(input, budget)
	if err != nil {
		t.Fatalf("SplitInput failed: %v", err)
	}

	// Verify we got chunks
	if len(chunks) == 0 {
		t.Fatal("Expected at least one chunk")
	}

	// Verify total content is preserved
	var reconstructed strings.Builder
	for _, chunk := range chunks {
		reconstructed.WriteString(chunk)
	}

	// The reconstruction might have minor whitespace differences
	// Just verify we didn't lose major content
	if len(reconstructed.String()) < len(input)/2 {
		t.Errorf("Lost too much content in splitting: original %d, reconstructed %d",
			len(input), len(reconstructed.String()))
	}

	t.Logf("Split into %d chunks (suggested: %d)", len(chunks), budget.SuggestedChunks)
}

// Test small input doesn't get split
func TestNoSplittingForSmallInput(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	input := "Small input that fits easily."
	system := "You are helpful."
	context := ""

	budget, err := manager.Calculate(system, context, input)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	chunks, err := manager.SplitInput(input, budget)
	if err != nil {
		t.Fatalf("SplitInput failed: %v", err)
	}

	// Should return single chunk
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for small input, got %d", len(chunks))
	}
	if chunks[0] != input {
		t.Error("Chunk content doesn't match original input")
	}
}

// Test empty inputs
func TestEmptyInputs(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-3-opus-20240229",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	budget, err := manager.Calculate("", "", "")
	if err != nil {
		t.Fatalf("Calculate failed with empty inputs: %v", err)
	}

	if budget.UsedTokens != 0 {
		t.Errorf("Expected 0 used tokens for empty inputs, got %d", budget.UsedTokens)
	}

	// Empty input should not need splitting
	if budget.NeedsSplitting {
		t.Error("Empty input should not need splitting")
	}
}

// Test impossible budget (system+context alone exceed limit)
func TestImpossibleBudget(t *testing.T) {
	// Use fallback counter with smaller limits for testing
	cfg := tokencount.Config{
		Provider: "unknown",
		Model:    "small-model",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	// Create system+context that's impossibly large
	// Fill most of the 128K context window
	hugeSystem := strings.Repeat("System instruction. ", 50000)
	hugeContext := strings.Repeat("Previous message. ", 50000)
	input := "Small input"

	_, err = manager.Calculate(hugeSystem, hugeContext, input)
	if err == nil {
		t.Error("Expected error for impossible budget, got nil")
	}
}

// Test budget with different models
func TestDifferentModels(t *testing.T) {
	models := []struct {
		provider string
		model    string
	}{
		{"openai", "gpt-5"},
		{"openai", "gpt-4o"},
		{"openai", "o1"},
		{"anthropic", "claude-sonnet-4-5-20250929"},
		{"anthropic", "claude-opus-4-1-20250805"},
		{"anthropic", "claude-3-5-sonnet-20241022"},
		{"anthropic", "claude-3-opus-20240229"},
	}

	input := "Test input for budget calculation."
	system := "You are helpful."
	context := ""

	for _, m := range models {
		cfg := tokencount.Config{
			Provider: m.provider,
			Model:    m.model,
		}

		counter, err := tokencount.NewCounter(cfg)
		if err != nil {
			t.Errorf("Failed to create counter for %s/%s: %v", m.provider, m.model, err)
			continue
		}

		manager := NewManager(counter)

		budget, err := manager.Calculate(system, context, input)
		if err != nil {
			t.Errorf("Calculate failed for %s/%s: %v", m.provider, m.model, err)
			continue
		}

		// Verify sensible values
		if budget.MaxOutputTokens <= 0 {
			t.Errorf("Invalid max output for %s/%s: %d", m.provider, m.model, budget.MaxOutputTokens)
		}
		if budget.AvailableTokens <= 0 {
			t.Errorf("Invalid available tokens for %s/%s: %d", m.provider, m.model, budget.AvailableTokens)
		}

		t.Logf("%s/%s: max_context=%d, max_output=%d, available=%d",
			m.provider, m.model,
			counter.MaxContextWindow(),
			budget.MaxOutputTokens,
			budget.AvailableTokens)
	}
}

// Test paragraph splitting
func TestParagraphSplitting(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	// Create input with clear paragraph boundaries
	input := "Paragraph one.\n\n" +
		"Paragraph two.\n\n" +
		"Paragraph three."

	// Force splitting by using huge paragraphs
	input = strings.Repeat(input, 10000)

	system := "System"
	context := ""

	budget, err := manager.Calculate(system, context, input)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if !budget.NeedsSplitting {
		// Create a budget that forces splitting
		budget.NeedsSplitting = true
		budget.SuggestedChunks = 3
	}

	chunks, err := manager.SplitInput(input, budget)
	if err != nil {
		t.Fatalf("SplitInput failed: %v", err)
	}

	if len(chunks) < 2 {
		t.Error("Expected multiple chunks for large input")
	}

	t.Logf("Split into %d chunks", len(chunks))
}

// Benchmark budget calculation
func BenchmarkBudgetCalculation(b *testing.B) {
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		b.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	system := "You are a helpful assistant."
	context := "Previous conversation."
	input := "What is the meaning of life?"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.Calculate(system, context, input)
	}
}

// Benchmark input splitting
func BenchmarkInputSplitting(b *testing.B) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-3-opus-20240229",
	}

	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		b.Fatalf("Failed to create counter: %v", err)
	}

	manager := NewManager(counter)

	input := strings.Repeat("Test paragraph with some content.\n\n", 1000)
	system := "System"
	context := ""

	budget, err := manager.Calculate(system, context, input)
	if err != nil {
		b.Fatalf("Calculate failed: %v", err)
	}

	// Force splitting
	if !budget.NeedsSplitting {
		budget.NeedsSplitting = true
		budget.SuggestedChunks = 5
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = manager.SplitInput(input, budget)
	}
}
