package agent

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/broker"
	"github.com/tenzoki/agen/cellorg/internal/envelope"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/omni/tokencount"
)

// Mock broker client for testing
type mockBrokerClient struct {
	published   []*envelope.Envelope
	subscribers map[string]chan *envelope.Envelope
}

func newMockBrokerClient() *mockBrokerClient {
	return &mockBrokerClient{
		published:   make([]*envelope.Envelope, 0),
		subscribers: make(map[string]chan *envelope.Envelope),
	}
}

func (m *mockBrokerClient) PublishEnvelope(topic string, env *envelope.Envelope) error {
	m.published = append(m.published, env)

	// If there's a subscriber, send to them
	if ch, exists := m.subscribers[topic]; exists {
		ch <- env
	}

	return nil
}

func (m *mockBrokerClient) SubscribeEnvelopes(topic string) (<-chan *envelope.Envelope, error) {
	ch := make(chan *envelope.Envelope, 10)
	m.subscribers[topic] = ch
	return ch, nil
}

// Test basic envelope publish without chunking
func TestEnvelopeFrameworkPublishSmall(t *testing.T) {
	// Create framework with nil client for testing
	ef := &EnvelopeFramework{
		client:         (*client.BrokerClient)(nil),
		chunkCollector: NewChunkCollector(5 * time.Minute),
		providerConfig: broker.NewProviderConfig(),
		subscriptions:  make(map[string]<-chan *envelope.Envelope),
	}
	defer ef.Close()

	// For testing, we'll manually test the chunking logic
	// by using the broker chunking utilities directly
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	ef.RegisterProvider("test-agent", counter)

	// Test that provider was registered
	registeredCounter := ef.providerConfig.GetCounter("test-agent")
	if registeredCounter == nil {
		t.Error("Provider was not registered")
	}
}

// Test chunk collection
func TestEnvelopeFrameworkChunkCollection(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

	// Create test chunks
	chunks := []*envelope.Envelope{
		{
			ID:      "chunk1",
			Payload: []byte("part1"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group123",
				"X-Chunk-Index": "0",
				"X-Chunk-Total": "2",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
		{
			ID:      "chunk2",
			Payload: []byte("part2"),
			Headers: map[string]string{
				"X-Chunk-ID":    "group123",
				"X-Chunk-Index": "1",
				"X-Chunk-Total": "2",
				"X-Original-ID": "original",
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		},
	}

	// Process first chunk
	merged1, complete1, err1 := ef.chunkCollector.CollectChunk(chunks[0])
	if err1 != nil {
		t.Fatalf("CollectChunk failed: %v", err1)
	}
	if complete1 {
		t.Error("First chunk should not be complete")
	}
	if merged1 != nil {
		t.Error("First chunk should return nil")
	}

	// Process second chunk
	merged2, complete2, err2 := ef.chunkCollector.CollectChunk(chunks[1])
	if err2 != nil {
		t.Fatalf("CollectChunk failed: %v", err2)
	}
	if !complete2 {
		t.Error("Second chunk should complete the message")
	}
	if merged2 == nil {
		t.Fatal("Merged envelope should not be nil")
	}

	// Verify merged payload
	expected := "part1part2"
	actual := string(merged2.Payload)
	if actual != expected {
		t.Errorf("Merged payload mismatch:\nExpected: %s\nGot: %s", expected, actual)
	}
}

// Test chunk status
func TestEnvelopeFrameworkChunkStatus(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

	// Initially no pending chunks
	if ef.CountPendingChunks() != 0 {
		t.Errorf("Expected 0 pending chunks, got %d", ef.CountPendingChunks())
	}

	// Add incomplete chunk
	chunk := &envelope.Envelope{
		ID:      "chunk1",
		Payload: []byte("part1"),
		Headers: map[string]string{
			"X-Chunk-ID":    "group123",
			"X-Chunk-Index": "0",
			"X-Chunk-Total": "3",
		},
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	ef.chunkCollector.CollectChunk(chunk)

	// Should have 1 pending chunk
	if ef.CountPendingChunks() != 1 {
		t.Errorf("Expected 1 pending chunk, got %d", ef.CountPendingChunks())
	}

	// Get status
	status := ef.GetChunkStatus()
	if len(status) != 1 {
		t.Errorf("Expected 1 chunk status, got %d", len(status))
	}

	chunkStatus := status["group123"]
	if chunkStatus.ReceivedCount != 1 {
		t.Errorf("Expected 1 received chunk, got %d", chunkStatus.ReceivedCount)
	}

	if chunkStatus.TotalCount != 3 {
		t.Errorf("Expected 3 total chunks, got %d", chunkStatus.TotalCount)
	}

	if chunkStatus.Complete {
		t.Error("Chunk group should not be complete")
	}
}

// Test default counter
func TestEnvelopeFrameworkDefaultCounter(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

	cfg := tokencount.Config{
		Provider: "openai",
		Model:    "gpt-5",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	ef.SetDefaultCounter(counter)

	if ef.defaultCounter == nil {
		t.Error("Default counter was not set")
	}
}

// Test processEnvelopes with chunked input
func TestProcessEnvelopesChunked(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

	input := make(chan *envelope.Envelope, 10)
	output := make(chan *envelope.Envelope, 10)

	// Start processor
	go ef.processEnvelopes(input, output)

	// Send chunks
	input <- &envelope.Envelope{
		ID:      "chunk1",
		Payload: []byte("part1"),
		Headers: map[string]string{
			"X-Chunk-ID":    "test",
			"X-Chunk-Index": "0",
			"X-Chunk-Total": "2",
			"X-Original-ID": "orig",
		},
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	input <- &envelope.Envelope{
		ID:      "chunk2",
		Payload: []byte("part2"),
		Headers: map[string]string{
			"X-Chunk-ID":    "test",
			"X-Chunk-Index": "1",
			"X-Chunk-Total": "2",
			"X-Original-ID": "orig",
		},
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	// Close input to signal end
	close(input)

	// Wait for merged output
	merged := <-output

	if merged == nil {
		t.Fatal("Expected merged envelope, got nil")
	}

	expected := "part1part2"
	actual := string(merged.Payload)
	if actual != expected {
		t.Errorf("Merged payload mismatch:\nExpected: %s\nGot: %s", expected, actual)
	}

	// Output should be closed now
	select {
	case _, ok := <-output:
		if ok {
			t.Error("Output channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		// OK
	}
}

// Test processEnvelopes with non-chunked input
func TestProcessEnvelopesNonChunked(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

	input := make(chan *envelope.Envelope, 10)
	output := make(chan *envelope.Envelope, 10)

	// Start processor
	go ef.processEnvelopes(input, output)

	// Send non-chunked envelope
	original := &envelope.Envelope{
		ID:         "test",
		Payload:    []byte("test message"),
		Headers:    make(map[string]string),
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}
	input <- original

	// Close input
	close(input)

	// Wait for output
	received := <-output

	if received == nil {
		t.Fatal("Expected envelope, got nil")
	}

	if string(received.Payload) != "test message" {
		t.Errorf("Payload mismatch: %s", string(received.Payload))
	}
}

// Test provider registration
func TestProviderRegistration(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

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

	// Register providers
	ef.RegisterProvider("agent-ai-1", counter1)
	ef.RegisterProvider("agent-ai-2", counter2)

	// Verify registration
	c1 := ef.providerConfig.GetCounter("agent-ai-1")
	if c1 != counter1 {
		t.Error("Counter for agent-ai-1 not registered correctly")
	}

	c2 := ef.providerConfig.GetCounter("agent-ai-2")
	if c2 != counter2 {
		t.Error("Counter for agent-ai-2 not registered correctly")
	}

	c3 := ef.providerConfig.GetCounter("unknown")
	if c3 != nil {
		t.Error("Unknown agent should return nil counter")
	}
}

// Test integration: chunking large JSON array through framework
func TestEnvelopeFrameworkIntegrationJSONArray(t *testing.T) {
	ef := NewEnvelopeFramework(nil)
	defer ef.Close()

	// Create large JSON array
	items := make([]map[string]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = map[string]string{
			"id":   string(rune('A' + (i % 26))),
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

	// Setup counter
	cfg := tokencount.Config{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-5-20250929",
	}
	counter, err := tokencount.NewCounter(cfg)
	if err != nil {
		t.Fatalf("Failed to create counter: %v", err)
	}

	ef.SetDefaultCounter(counter)

	// Verify default counter was set
	if ef.defaultCounter == nil {
		t.Fatal("Default counter not set")
	}

	// Test envelope would be chunked using helper
	helper := broker.NewChunkingHelper(counter)
	shouldChunk, err := helper.ShouldChunk(env)
	if err != nil {
		t.Fatalf("ShouldChunk failed: %v", err)
	}

	t.Logf("Large JSON array should chunk: %v", shouldChunk)
}
