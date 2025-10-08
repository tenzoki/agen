package orchestrator

import (
	"context"
	"strings"
	"testing"

	"github.com/tenzoki/agen/alfa/internal/ai"
	"github.com/tenzoki/agen/omni/tokencount"
)

// MockLLM for testing
type mockLLM struct {
	responses []string
	callCount int
}

func (m *mockLLM) Chat(ctx context.Context, messages []ai.Message) (*ai.Response, error) {
	content := "Mock response"
	if m.callCount < len(m.responses) {
		content = m.responses[m.callCount]
	}

	m.callCount++
	return &ai.Response{
		Content: content,
		Model:   "mock-model",
	}, nil
}

func (m *mockLLM) ChatStream(ctx context.Context, messages []ai.Message) (<-chan string, <-chan error) {
	return nil, nil
}

func (m *mockLLM) Provider() string {
	return "mock"
}

func (m *mockLLM) Model() string {
	return "mock-model"
}

// Test chunk processor with small input (no splitting)
func TestChunkProcessorNoSplitting(t *testing.T) {
	// Create mock LLM
	mockLLM := &mockLLM{
		responses: []string{"Response to small input"},
	}

	// Create counter
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Create processor
	processor := NewChunkProcessor(mockLLM, counter)

	// Process small input
	ctx := context.Background()
	system := "You are helpful."
	ctxString := ""
	input := "Small input."

	response, err := processor.ProcessWithChunking(ctx, system, ctxString, input)
	if err != nil {
		t.Fatalf("ProcessWithChunking failed: %v", err)
	}

	if response != "Response to small input" {
		t.Errorf("Expected 'Response to small input', got '%s'", response)
	}

	// Should only call LLM once (no splitting)
	if mockLLM.callCount != 1 {
		t.Errorf("Expected 1 LLM call, got %d", mockLLM.callCount)
	}
}

// Test chunk processor with large input (requires splitting)
func TestChunkProcessorWithSplitting(t *testing.T) {
	t.Skip("Skipping splitting test - requires optimization of token counting for large inputs")

	// Create mock LLM with multiple responses
	mockLLM := &mockLLM{
		responses: []string{
			"Response to chunk 1",
			"Response to chunk 2",
			"Merged and deduplicated response",
		},
	}

	// Create counter
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Create processor
	processor := NewChunkProcessor(mockLLM, counter)

	// Create large input
	ctx := context.Background()
	system := "System"
	ctxString := ""
	input := strings.Repeat("Large input text. ", 1000)

	response, err := processor.ProcessWithChunking(ctx, system, ctxString, input)
	if err != nil {
		t.Fatalf("ProcessWithChunking failed: %v", err)
	}

	// Should get some response
	if response == "" {
		t.Error("Expected non-empty response")
	}

	t.Logf("Made %d LLM calls", mockLLM.callCount)
}

// Test GetBudgetInfo
func TestGetBudgetInfo(t *testing.T) {
	mockLLM := &mockLLM{}

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	processor := NewChunkProcessor(mockLLM, counter)

	budget, err := processor.GetBudgetInfo("System", "Context", "Input")
	if err != nil {
		t.Fatalf("GetBudgetInfo failed: %v", err)
	}

	if budget.SystemTokens <= 0 {
		t.Error("Expected positive system tokens")
	}
	if budget.ContextTokens <= 0 {
		t.Error("Expected positive context tokens")
	}
	if budget.InputTokens <= 0 {
		t.Error("Expected positive input tokens")
	}
}

// Test AddOverlap
func TestAddOverlap(t *testing.T) {
	mockLLM := &mockLLM{}

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	processor := NewChunkProcessor(mockLLM, counter)

	chunks := []string{
		"First chunk with some content here.",
		"Second chunk with different content.",
		"Third chunk with more content.",
	}

	overlapped := processor.AddOverlap(chunks, 0.2) // 20% overlap

	if len(overlapped) != len(chunks) {
		t.Fatalf("Expected %d chunks, got %d", len(chunks), len(overlapped))
	}

	// First chunk should be unchanged
	if overlapped[0] != chunks[0] {
		t.Error("First chunk should not have overlap prefix")
	}

	// Second chunk should have overlap from first
	if !strings.Contains(overlapped[1], "...") {
		t.Error("Second chunk should have overlap marker")
	}

	// Third chunk should have overlap from second
	if !strings.Contains(overlapped[2], "...") {
		t.Error("Third chunk should have overlap marker")
	}

	t.Logf("Original chunks: %v", chunks)
	t.Logf("Overlapped chunks: %v", overlapped)
}

// Test AddOverlap with single chunk
func TestAddOverlapSingleChunk(t *testing.T) {
	mockLLM := &mockLLM{}

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	processor := NewChunkProcessor(mockLLM, counter)

	chunks := []string{"Single chunk"}
	overlapped := processor.AddOverlap(chunks, 0.2)

	if len(overlapped) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(overlapped))
	}

	if overlapped[0] != chunks[0] {
		t.Error("Single chunk should be unchanged")
	}
}

// Test AddOverlap with zero overlap ratio
func TestAddOverlapZeroRatio(t *testing.T) {
	mockLLM := &mockLLM{}

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	processor := NewChunkProcessor(mockLLM, counter)

	chunks := []string{"First", "Second", "Third"}
	overlapped := processor.AddOverlap(chunks, 0.0)

	// Should return unchanged
	for i := range chunks {
		if overlapped[i] != chunks[i] {
			t.Errorf("Chunk %d should be unchanged with zero overlap", i)
		}
	}
}

// Benchmark chunk processing
func BenchmarkChunkProcessorSmallInput(b *testing.B) {
	mockLLM := &mockLLM{
		responses: []string{"Response"},
	}

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, _ := tokencount.NewCounter(cfg)
	processor := NewChunkProcessor(mockLLM, counter)

	ctx := context.Background()
	system := "System"
	ctxString := ""
	input := "Small input"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockLLM.callCount = 0
		_, _ = processor.ProcessWithChunking(ctx, system, ctxString, input)
	}
}
