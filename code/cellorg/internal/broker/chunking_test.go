package broker

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/envelope"
	"github.com/tenzoki/agen/omni/tokencount"
)

// Test ChunkingHelper ShouldChunk
func TestShouldChunk(t *testing.T) {
	// Create token counter
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	helper := NewChunkingHelper(counter)

	// Test small envelope - should not chunk
	smallEnv := &envelope.Envelope{
		ID:          "test-small",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte("Small message"),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	shouldChunk, err := helper.ShouldChunk(smallEnv)
	if err != nil {
		t.Fatalf("ShouldChunk failed: %v", err)
	}
	if shouldChunk {
		t.Error("Small envelope should not need chunking")
	}

	// Test large envelope - should chunk
	largeText := strings.Repeat("This is a test sentence. ", 50000)
	largeEnv := &envelope.Envelope{
		ID:          "test-large",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte(largeText),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	shouldChunk, err = helper.ShouldChunk(largeEnv)
	if err != nil {
		t.Fatalf("ShouldChunk failed: %v", err)
	}
	if !shouldChunk {
		t.Error("Large envelope should need chunking")
	}
}

// Test ChunkingHelper with nil counter
func TestShouldChunkNilCounter(t *testing.T) {
	helper := NewChunkingHelper(nil)

	env := &envelope.Envelope{
		ID:      "test",
		Payload: []byte("test"),
		Headers: make(map[string]string),
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	shouldChunk, err := helper.ShouldChunk(env)
	if err != nil {
		t.Fatalf("ShouldChunk failed: %v", err)
	}
	if shouldChunk {
		t.Error("Nil counter should return false for ShouldChunk")
	}
}

// Test PrepareForPublish with small envelope
func TestPrepareForPublishSmall(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	helper := NewChunkingHelper(counter)

	env := &envelope.Envelope{
		ID:          "test",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte("Small message"),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	envelopes, err := helper.PrepareForPublish(env)
	if err != nil {
		t.Fatalf("PrepareForPublish failed: %v", err)
	}

	if len(envelopes) != 1 {
		t.Errorf("Expected 1 envelope, got %d", len(envelopes))
	}

	if envelopes[0] != env {
		t.Error("Small envelope should return original")
	}
}

// Test PrepareForPublish with large envelope
func TestPrepareForPublishLarge(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	helper := NewChunkingHelper(counter)

	largeText := strings.Repeat("This is a test sentence. ", 50000)
	env := &envelope.Envelope{
		ID:          "test-large",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte(largeText),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	envelopes, err := helper.PrepareForPublish(env)
	if err != nil {
		t.Fatalf("PrepareForPublish failed: %v", err)
	}

	if len(envelopes) <= 1 {
		t.Errorf("Expected multiple envelopes (chunks), got %d", len(envelopes))
	}

	// Verify all chunks have chunk headers
	chunkID := envelopes[0].Headers["X-Chunk-ID"]
	if chunkID == "" {
		t.Error("First chunk missing X-Chunk-ID header")
	}

	for i, chunk := range envelopes {
		if chunk.Headers["X-Chunk-ID"] != chunkID {
			t.Errorf("Chunk %d has different chunk ID", i)
		}
		if chunk.Headers["X-Chunk-Index"] == "" {
			t.Errorf("Chunk %d missing X-Chunk-Index", i)
		}
	}

	t.Logf("Large envelope split into %d chunks", len(envelopes))
}

// Test ChunkingPublisher with small envelope
func TestChunkingPublisherSmall(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Track published envelopes
	published := make([]*envelope.Envelope, 0)
	publishFunc := func(env *envelope.Envelope) error {
		published = append(published, env)
		return nil
	}

	publisher := NewChunkingPublisher(counter, publishFunc)

	env := &envelope.Envelope{
		ID:          "test",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte("Small message"),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	err = publisher.Publish(env)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if len(published) != 1 {
		t.Errorf("Expected 1 published envelope, got %d", len(published))
	}
}

// Test ChunkingPublisher with large envelope
func TestChunkingPublisherLarge(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Track published envelopes
	published := make([]*envelope.Envelope, 0)
	publishFunc := func(env *envelope.Envelope) error {
		published = append(published, env)
		return nil
	}

	publisher := NewChunkingPublisher(counter, publishFunc)

	largeText := strings.Repeat("This is a test sentence. ", 50000)
	env := &envelope.Envelope{
		ID:          "test-large",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "test",
		Timestamp:   time.Now(),
		Payload:     []byte(largeText),
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	err = publisher.Publish(env)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if len(published) <= 1 {
		t.Errorf("Expected multiple published envelopes (chunks), got %d", len(published))
	}

	t.Logf("Published %d chunks", len(published))
}

// Test ChunkingPublisher with failing publish function
func TestChunkingPublisherError(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	// Publish function that fails
	publishFunc := func(env *envelope.Envelope) error {
		return fmt.Errorf("publish failed")
	}

	publisher := NewChunkingPublisher(counter, publishFunc)

	env := &envelope.Envelope{
		ID:      "test",
		Payload: []byte("test"),
		Headers: make(map[string]string),
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	err = publisher.Publish(env)
	if err == nil {
		t.Error("Expected publish error, got nil")
	}
}

// Test ProviderConfig
func TestProviderConfig(t *testing.T) {
	cfg1 := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter1, _ := tokencount.NewCounter(cfg1)

	cfg2 := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter2, _ := tokencount.NewCounter(cfg2)

	providerConfig := NewProviderConfig()

	// Register providers
	providerConfig.RegisterProvider("agent-ai-1", counter1)
	providerConfig.RegisterProvider("agent-ai-2", counter2)

	// Get counters
	c1 := providerConfig.GetCounter("agent-ai-1")
	if c1 != counter1 {
		t.Error("GetCounter returned wrong counter for agent-ai-1")
	}

	c2 := providerConfig.GetCounter("agent-ai-2")
	if c2 != counter2 {
		t.Error("GetCounter returned wrong counter for agent-ai-2")
	}

	c3 := providerConfig.GetCounter("unknown-agent")
	if c3 != nil {
		t.Error("GetCounter should return nil for unknown agent")
	}
}

// Test ProviderConfig CreatePublisher
func TestProviderConfigCreatePublisher(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, _ := tokencount.NewCounter(cfg)

	providerConfig := NewProviderConfig()
	providerConfig.RegisterProvider("agent-ai", counter)

	published := make([]*envelope.Envelope, 0)
	publishFunc := func(env *envelope.Envelope) error {
		published = append(published, env)
		return nil
	}

	publisher := providerConfig.CreatePublisher("agent-ai", publishFunc)
	if publisher == nil {
		t.Fatal("CreatePublisher returned nil")
	}

	env := &envelope.Envelope{
		ID:      "test",
		Payload: []byte("test message"),
		Headers: make(map[string]string),
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	err := publisher.Publish(env)
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if len(published) != 1 {
		t.Errorf("Expected 1 published envelope, got %d", len(published))
	}
}

// Test PrepareForPublish with JSON array
func TestPrepareForPublishJSONArray(t *testing.T) {
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	helper := NewChunkingHelper(counter)

	// Create large JSON array
	items := make([]map[string]string, 5000)
	for i := 0; i < 5000; i++ {
		items[i] = map[string]string{
			"id":   fmt.Sprintf("%d", i),
			"data": strings.Repeat("x", 100),
		}
	}
	payloadBytes, _ := json.Marshal(items)

	env := &envelope.Envelope{
		ID:          "test-json",
		Source:      "sender",
		Destination: "receiver",
		MessageType: "json_array",
		Timestamp:   time.Now(),
		Payload:     payloadBytes,
		Headers:     make(map[string]string),
		Properties:  make(map[string]interface{}),
		Route:       make([]string, 0),
	}

	envelopes, err := helper.PrepareForPublish(env)
	if err != nil {
		t.Fatalf("PrepareForPublish failed: %v", err)
	}

	if len(envelopes) <= 1 {
		t.Errorf("Expected multiple envelopes for large JSON array, got %d", len(envelopes))
	}

	// Verify each chunk contains valid JSON array
	totalItems := 0
	for i, chunk := range envelopes {
		var arr []map[string]string
		if err := json.Unmarshal(chunk.Payload, &arr); err != nil {
			t.Errorf("Chunk %d contains invalid JSON: %v", i, err)
		}
		totalItems += len(arr)
	}

	if totalItems != 5000 {
		t.Errorf("Expected 5000 total items across chunks, got %d", totalItems)
	}

	t.Logf("Large JSON array split into %d chunks", len(envelopes))
}
