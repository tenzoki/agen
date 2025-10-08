package envelope

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/tenzoki/agen/omni/tokencount"
)

// Test basic budget calculation
func TestCalculateBudget(t *testing.T) {
	// Create test envelope with small payload
	payload := map[string]string{
		"text": "This is a test message.",
	}
	payloadBytes, _ := json.Marshal(payload)

	env := &Envelope{
		ID:          "test-id",
		Source:      "test-agent",
		Destination: "target-agent",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     payloadBytes,
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	// Create token counter
	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Calculate budget
	budget, err := CalculateBudget(env, counter)
	if err != nil {
		t.Fatalf("CalculateBudget failed: %v", err)
	}

	// Verify basic fields
	if budget.PayloadTokens <= 0 {
		t.Errorf("Expected positive payload tokens, got %d", budget.PayloadTokens)
	}

	if budget.HeaderTokens <= 0 {
		t.Errorf("Expected positive header tokens, got %d", budget.HeaderTokens)
	}

	if budget.TotalTokens != budget.PayloadTokens+budget.HeaderTokens {
		t.Errorf("Total tokens mismatch: %d != %d + %d",
			budget.TotalTokens, budget.PayloadTokens, budget.HeaderTokens)
	}

	// Small payload should not need splitting
	if budget.NeedsSplitting {
		t.Error("Small payload should not need splitting")
	}

	if budget.SuggestedChunks != 1 {
		t.Errorf("Expected 1 chunk for small payload, got %d", budget.SuggestedChunks)
	}

	t.Logf("Budget: payload=%d, headers=%d, total=%d, available=%d",
		budget.PayloadTokens, budget.HeaderTokens, budget.TotalTokens, budget.AvailableTokens)
}

// Test with large payload requiring splitting
func TestCalculateBudgetLargePayload(t *testing.T) {
	// Create large text payload (simulate large document)
	largeText := ""
	for i := 0; i < 50000; i++ {
		largeText += "This is a line of text in a large document. "
	}

	payload := map[string]string{
		"document": largeText,
	}
	payloadBytes, _ := json.Marshal(payload)

	env := &Envelope{
		ID:          "test-id",
		Source:      "test-agent",
		Destination: "target-agent",
		MessageType: "large_document",
		Timestamp:   time.Now(),
		Payload:     payloadBytes,
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	// Use Anthropic counter (smaller max output)
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-3-opus-20240229",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Calculate budget
	budget, err := CalculateBudget(env, counter)
	if err != nil {
		t.Fatalf("CalculateBudget failed: %v", err)
	}

	// Large payload should trigger splitting
	if !budget.NeedsSplitting {
		t.Error("Large payload should need splitting")
	}

	if budget.SuggestedChunks <= 1 {
		t.Errorf("Expected multiple chunks for large payload, got %d", budget.SuggestedChunks)
	}

	t.Logf("Large payload: %d tokens, needs %d chunks",
		budget.PayloadTokens, budget.SuggestedChunks)
}

// Test metadata token estimation
func TestEstimateMetadataTokens(t *testing.T) {
	tests := []struct {
		name          string
		envelope      *Envelope
		minExpected   int
		maxExpected   int
	}{
		{
			name: "Minimal envelope",
			envelope: &Envelope{
				ID:          "id",
				Source:      "src",
				Destination: "dst",
				MessageType: "test",
				Timestamp:   time.Now(),
				Payload:     []byte("{}"),
				Headers:     make(map[string]string),
				Properties:  make(map[string]interface{}),
				Route:       make([]string, 0),
			},
			minExpected: 180,
			maxExpected: 220,
		},
		{
			name: "Envelope with headers",
			envelope: &Envelope{
				ID:          "id",
				Source:      "src",
				Destination: "dst",
				MessageType: "test",
				Timestamp:   time.Now(),
				Payload:     []byte("{}"),
				Headers: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
				Properties: make(map[string]interface{}),
				Route:      make([]string, 0),
			},
			minExpected: 220, // base + 3*10 headers
			maxExpected: 260,
		},
		{
			name: "Envelope with route",
			envelope: &Envelope{
				ID:          "id",
				Source:      "src",
				Destination: "dst",
				MessageType: "test",
				Timestamp:   time.Now(),
				Payload:     []byte("{}"),
				Headers:     make(map[string]string),
				Properties:  make(map[string]interface{}),
				Route:       []string{"agent1", "agent2", "agent3"},
			},
			minExpected: 220, // base + 3*10 route
			maxExpected: 260,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateMetadataTokens(tt.envelope)

			if tokens < tt.minExpected || tokens > tt.maxExpected {
				t.Errorf("Expected tokens between %d and %d, got %d",
					tt.minExpected, tt.maxExpected, tokens)
			}

			t.Logf("%s: %d tokens", tt.name, tokens)
		})
	}
}

// Test with different AI providers
func TestCalculateBudgetDifferentProviders(t *testing.T) {
	payload := map[string]string{
		"text": "Test message for different providers.",
	}
	payloadBytes, _ := json.Marshal(payload)

	env := &Envelope{
		ID:          "test-id",
		Source:      "test-agent",
		Destination: "target-agent",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     payloadBytes,
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	providers := []struct {
		provider string
		model    string
	}{
		{"openai", "gpt-5"},
		{"anthropic", "claude-sonnet-4-5-20250929"},
		{"unknown", "test-model"},
	}

	for _, p := range providers {
		cfg := tokencount.Config{
			Provider: p.provider,
			Model:    p.model,
		}

		counter, err := tokencount.NewCounter(cfg)
		if err != nil {
			t.Errorf("Failed to create counter for %s/%s: %v", p.provider, p.model, err)
			continue
		}

		budget, err := CalculateBudget(env, counter)
		if err != nil {
			t.Errorf("CalculateBudget failed for %s/%s: %v", p.provider, p.model, err)
			continue
		}

		t.Logf("%s/%s: max_context=%d, max_output=%d, payload=%d tokens",
			p.provider, p.model,
			budget.MaxContextWindow,
			budget.MaxOutputTokens,
			budget.PayloadTokens)
	}
}

// Test edge case: empty payload
func TestCalculateBudgetEmptyPayload(t *testing.T) {
	env := &Envelope{
		ID:          "test-id",
		Source:      "test-agent",
		Destination: "target-agent",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte("{}"),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	budget, err := CalculateBudget(env, counter)
	if err != nil {
		t.Fatalf("CalculateBudget failed: %v", err)
	}

	// Empty payload should still have some tokens (metadata)
	if budget.TotalTokens <= 0 {
		t.Errorf("Expected positive total tokens even for empty payload, got %d", budget.TotalTokens)
	}

	// Should not need splitting
	if budget.NeedsSplitting {
		t.Error("Empty payload should not need splitting")
	}
}
