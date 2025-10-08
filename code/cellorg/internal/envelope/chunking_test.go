package envelope

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tenzoki/agen/omni/tokencount"
)

// Test basic chunking of large text payload
func TestChunkEnvelopeTextPayload(t *testing.T) {
	// Create large text payload
	largeText := strings.Repeat("This is a test sentence. ", 10000)

	env := &Envelope{
		ID:          "test-id",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "text_document",
		Timestamp:   time.Now(),
		Payload:     []byte(largeText),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	// Calculate budget
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	budget, err := CalculateBudget(env, counter)
	if err != nil {
		t.Fatalf("Failed to calculate budget: %v", err)
	}

	if !budget.NeedsSplitting {
		t.Skip("Payload not large enough to trigger splitting")
	}

	// Chunk envelope
	chunks, err := ChunkEnvelope(env, budget)
	if err != nil {
		t.Fatalf("ChunkEnvelope failed: %v", err)
	}

	// Verify chunks
	if len(chunks) <= 1 {
		t.Errorf("Expected multiple chunks, got %d", len(chunks))
	}

	if len(chunks) != budget.SuggestedChunks {
		t.Logf("Warning: got %d chunks, suggested %d", len(chunks), budget.SuggestedChunks)
	}

	// Verify chunk metadata
	chunkID := chunks[0].Headers["X-Chunk-ID"]
	if chunkID == "" {
		t.Error("First chunk missing X-Chunk-ID")
	}

	for i, chunk := range chunks {
		if chunk.Headers["X-Chunk-ID"] != chunkID {
			t.Errorf("Chunk %d has different chunk ID", i)
		}

		if chunk.Headers["X-Chunk-Index"] != string(rune('0'+i)) {
			t.Errorf("Chunk %d has wrong index: %s", i, chunk.Headers["X-Chunk-Index"])
		}

		if chunk.Headers["X-Original-ID"] != env.ID {
			t.Errorf("Chunk %d missing original ID", i)
		}

		// Verify chunk is smaller than original
		if len(chunk.Payload) >= len(env.Payload) {
			t.Errorf("Chunk %d not smaller than original", i)
		}
	}

	t.Logf("Split %d token payload into %d chunks", budget.PayloadTokens, len(chunks))
}

// Test chunking JSON array payload
func TestChunkEnvelopeJSONArray(t *testing.T) {
	// Create JSON array with many items
	items := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = map[string]string{
			"id":   string(rune('A' + (i % 26))),
			"data": strings.Repeat("x", 100),
		}
	}

	payloadBytes, _ := json.Marshal(items)

	env := &Envelope{
		ID:          "test-id",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "json_array",
		Timestamp:   time.Now(),
		Payload:     payloadBytes,
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	// Force chunking with small context
	budget := &EnvelopeBudget{
		PayloadTokens:    50000,
		HeaderTokens:     200,
		TotalTokens:      50200,
		NeedsSplitting:   true,
		SuggestedChunks:  4,
		MaxContextWindow: 200000,
		MaxOutputTokens:  64000,
	}

	chunks, err := ChunkEnvelope(env, budget)
	if err != nil {
		t.Fatalf("ChunkEnvelope failed: %v", err)
	}

	if len(chunks) != 4 {
		t.Errorf("Expected 4 chunks, got %d", len(chunks))
	}

	// Verify each chunk contains valid JSON array
	totalItems := 0
	for i, chunk := range chunks {
		var arr []map[string]string
		if err := json.Unmarshal(chunk.Payload, &arr); err != nil {
			t.Errorf("Chunk %d contains invalid JSON: %v", i, err)
		}
		totalItems += len(arr)
	}

	if totalItems != 1000 {
		t.Errorf("Expected 1000 total items, got %d", totalItems)
	}

	t.Logf("Split 1000 items into %d chunks", len(chunks))
}

// Test small payload that doesn't need splitting
func TestChunkEnvelopeNoSplitting(t *testing.T) {
	env := &Envelope{
		ID:          "test-id",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "small_message",
		Timestamp:   time.Now(),
		Payload:     []byte("This is a small message."),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	budget := &EnvelopeBudget{
		PayloadTokens:    10,
		HeaderTokens:     200,
		TotalTokens:      210,
		NeedsSplitting:   false,
		SuggestedChunks:  1,
		MaxContextWindow: 200000,
		MaxOutputTokens:  64000,
	}

	chunks, err := ChunkEnvelope(env, budget)
	if err != nil {
		t.Fatalf("ChunkEnvelope failed: %v", err)
	}

	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk (no splitting), got %d", len(chunks))
	}

	if chunks[0] != env {
		t.Error("Single chunk should be original envelope")
	}
}

// Test merging text chunks
func TestMergeChunksText(t *testing.T) {
	originalText := "This is the first part. This is the second part. This is the third part."

	// Create chunks manually
	chunks := []*Envelope{
		{
			ID:          "chunk1",
			Source:      "sender",
			Destination: "receiver",
			MessageType: "test",
			Timestamp:   time.Now(),
			Payload:     []byte("This is the first part. "),
			Headers: map[string]string{
				"X-Chunk-ID":    "group123",
				"X-Chunk-Index": "0",
				"X-Chunk-Total": "3",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			ID:          "chunk2",
			Source:      "sender",
			Destination: "receiver",
			MessageType: "test",
			Timestamp:   time.Now(),
			Payload:     []byte("This is the second part. "),
			Headers: map[string]string{
				"X-Chunk-ID":    "group123",
				"X-Chunk-Index": "1",
				"X-Chunk-Total": "3",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			ID:          "chunk3",
			Source:      "sender",
			Destination: "receiver",
			MessageType: "test",
			Timestamp:   time.Now(),
			Payload:     []byte("This is the third part."),
			Headers: map[string]string{
				"X-Chunk-ID":    "group123",
				"X-Chunk-Index": "2",
				"X-Chunk-Total": "3",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
	}

	merged, err := MergeChunks(chunks)
	if err != nil {
		t.Fatalf("MergeChunks failed: %v", err)
	}

	mergedText := string(merged.Payload)
	if mergedText != originalText {
		t.Errorf("Merged text doesn't match:\nExpected: %s\nGot: %s",
			originalText, mergedText)
	}

	// Verify chunk headers removed
	if merged.Headers["X-Chunk-ID"] != "" {
		t.Error("Merged envelope still has X-Chunk-ID header")
	}

	// Verify original ID restored
	if merged.ID != "original" {
		t.Errorf("Expected ID 'original', got '%s'", merged.ID)
	}
}

// Test merging JSON array chunks
func TestMergeChunksJSONArray(t *testing.T) {
	// Create JSON array chunks
	chunks := []*Envelope{
		{
			ID:          "chunk1",
			Source:      "sender",
			Destination: "receiver",
			MessageType: "test",
			Timestamp:   time.Now(),
			Payload:     []byte(`[{"id":1},{"id":2}]`),
			Headers: map[string]string{
				"X-Chunk-ID":    "group456",
				"X-Chunk-Index": "0",
				"X-Chunk-Total": "2",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			ID:          "chunk2",
			Source:      "sender",
			Destination: "receiver",
			MessageType: "test",
			Timestamp:   time.Now(),
			Payload:     []byte(`[{"id":3},{"id":4}]`),
			Headers: map[string]string{
				"X-Chunk-ID":    "group456",
				"X-Chunk-Index": "1",
				"X-Chunk-Total": "2",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
	}

	merged, err := MergeChunks(chunks)
	if err != nil {
		t.Fatalf("MergeChunks failed: %v", err)
	}

	// Verify merged JSON
	var arr []map[string]int
	if err := json.Unmarshal(merged.Payload, &arr); err != nil {
		t.Fatalf("Merged payload is not valid JSON: %v", err)
	}

	if len(arr) != 4 {
		t.Errorf("Expected 4 items in merged array, got %d", len(arr))
	}

	// Verify order preserved
	for i, item := range arr {
		if item["id"] != i+1 {
			t.Errorf("Item %d has wrong id: %d", i, item["id"])
		}
	}
}

// Test merging with out-of-order chunks
func TestMergeChunksOutOfOrder(t *testing.T) {
	// Create chunks in wrong order
	chunks := []*Envelope{
		{
			ID:      "chunk2",
			Payload: []byte("second "),
			Headers: map[string]string{
				"X-Chunk-ID":    "group789",
				"X-Chunk-Index": "1",
				"X-Chunk-Total": "3",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			ID:      "chunk3",
			Payload: []byte("third"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group789",
				"X-Chunk-Index": "2",
				"X-Chunk-Total": "3",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			ID:      "chunk1",
			Payload: []byte("first "),
			Headers: map[string]string{
				"X-Chunk-ID":    "group789",
				"X-Chunk-Index": "0",
				"X-Chunk-Total": "3",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
	}

	merged, err := MergeChunks(chunks)
	if err != nil {
		t.Fatalf("MergeChunks failed: %v", err)
	}

	expected := "first second third"
	actual := string(merged.Payload)
	if actual != expected {
		t.Errorf("Chunks not merged in correct order:\nExpected: %s\nGot: %s",
			expected, actual)
	}
}

// Test round-trip: chunk then merge
func TestChunkAndMergeRoundTrip(t *testing.T) {
	originalText := strings.Repeat("This is a test sentence. ", 500)

	env := &Envelope{
		ID:          "original-id",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte(originalText),
		Headers:     map[string]string{"Custom": "Header"},
		Properties:  map[string]interface{}{"key": "value"},
		Route:       []string{"hop1", "hop2"},
	}

	// Force chunking
	budget := &EnvelopeBudget{
		NeedsSplitting:  true,
		SuggestedChunks: 5,
	}

	// Chunk
	chunks, err := ChunkEnvelope(env, budget)
	if err != nil {
		t.Fatalf("ChunkEnvelope failed: %v", err)
	}

	t.Logf("Created %d chunks", len(chunks))

	// Merge
	merged, err := MergeChunks(chunks)
	if err != nil {
		t.Fatalf("MergeChunks failed: %v", err)
	}

	// Verify payload preserved
	if string(merged.Payload) != originalText {
		t.Error("Merged payload doesn't match original")
	}

	// Verify metadata preserved
	if merged.ID != env.ID {
		t.Errorf("ID changed: %s -> %s", env.ID, merged.ID)
	}

	if merged.Headers["Custom"] != "Header" {
		t.Error("Custom header not preserved")
	}

	if merged.Properties["key"] != "value" {
		t.Error("Property not preserved")
	}

	if len(merged.Route) != 2 {
		t.Errorf("Route not preserved: %v", merged.Route)
	}
}

// Test error case: missing chunks
func TestMergeChunksMissingChunk(t *testing.T) {
	chunks := []*Envelope{
		{
			Payload: []byte("first"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group",
				"X-Chunk-Index": "0",
				"X-Chunk-Total": "3",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			Payload: []byte("third"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group",
				"X-Chunk-Index": "2",
				"X-Chunk-Total": "3",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		// Missing chunk 1
	}

	_, err := MergeChunks(chunks)
	if err == nil {
		t.Error("Expected error for missing chunks, got nil")
	}
}

// Test error case: mismatched chunk IDs
func TestMergeChunksMismatchedIDs(t *testing.T) {
	chunks := []*Envelope{
		{
			Payload: []byte("first"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group1",
				"X-Chunk-Index": "0",
				"X-Chunk-Total": "2",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			Payload: []byte("second"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group2", // Different ID
				"X-Chunk-Index": "1",
				"X-Chunk-Total": "2",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
	}

	_, err := MergeChunks(chunks)
	if err == nil {
		t.Error("Expected error for mismatched chunk IDs, got nil")
	}
}
