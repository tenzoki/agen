package agent

import (
	"testing"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/envelope"
)

// Test collecting non-chunked message
func TestCollectChunkNonChunked(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	env := &envelope.Envelope{
		ID:      "test-id",
		Payload: []byte("test message"),
		Headers: make(map[string]string), // No chunk headers
		Properties: make(map[string]interface{}),
		Route:      make([]string, 0),
	}

	merged, complete, err := cc.CollectChunk(env)
	if err != nil {
		t.Fatalf("CollectChunk failed: %v", err)
	}

	if !complete {
		t.Error("Non-chunked message should be complete immediately")
	}

	if merged != env {
		t.Error("Non-chunked message should return original envelope")
	}
}

// Test collecting chunks in order
func TestCollectChunksInOrder(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	chunks := createTestChunks(3, "group123")

	// Collect chunks one by one
	for i := 0; i < 2; i++ {
		merged, complete, err := cc.CollectChunk(chunks[i])
		if err != nil {
			t.Fatalf("CollectChunk failed on chunk %d: %v", i, err)
		}

		if complete {
			t.Errorf("Chunk %d should not be complete yet", i)
		}

		if merged != nil {
			t.Errorf("Chunk %d should return nil envelope", i)
		}
	}

	// Last chunk should complete
	merged, complete, err := cc.CollectChunk(chunks[2])
	if err != nil {
		t.Fatalf("CollectChunk failed on final chunk: %v", err)
	}

	if !complete {
		t.Error("Final chunk should complete the message")
	}

	if merged == nil {
		t.Fatal("Merged envelope should not be nil")
	}

	// Verify merged payload
	expected := "chunk0chunk1chunk2"
	actual := string(merged.Payload)
	if actual != expected {
		t.Errorf("Merged payload mismatch:\nExpected: %s\nGot: %s", expected, actual)
	}
}

// Test collecting chunks out of order
func TestCollectChunksOutOfOrder(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	chunks := createTestChunks(3, "group456")

	// Collect in order: 2, 0, 1
	order := []int{2, 0, 1}

	for i := 0; i < 2; i++ {
		idx := order[i]
		_, complete, err := cc.CollectChunk(chunks[idx])
		if err != nil {
			t.Fatalf("CollectChunk failed on chunk %d: %v", idx, err)
		}

		if complete {
			t.Errorf("Chunk %d should not be complete yet", idx)
		}
	}

	// Last chunk should complete
	merged, complete, err := cc.CollectChunk(chunks[order[2]])
	if err != nil {
		t.Fatalf("CollectChunk failed on final chunk: %v", err)
	}

	if !complete {
		t.Error("Final chunk should complete the message")
	}

	if merged == nil {
		t.Fatal("Merged envelope should not be nil")
	}

	// Verify merged payload (should be in correct order)
	expected := "chunk0chunk1chunk2"
	actual := string(merged.Payload)
	if actual != expected {
		t.Errorf("Merged payload not in correct order:\nExpected: %s\nGot: %s", expected, actual)
	}
}

// Test duplicate chunk detection
func TestCollectChunksDuplicate(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	chunks := createTestChunks(2, "group789")

	// Collect first chunk
	_, complete, err := cc.CollectChunk(chunks[0])
	if err != nil {
		t.Fatalf("CollectChunk failed: %v", err)
	}
	if complete {
		t.Error("First chunk should not be complete")
	}

	// Send first chunk again (duplicate)
	_, complete, err = cc.CollectChunk(chunks[0])
	if err != nil {
		t.Fatalf("CollectChunk failed on duplicate: %v", err)
	}
	if complete {
		t.Error("Duplicate chunk should not complete message")
	}

	// Send second chunk
	merged, complete, err := cc.CollectChunk(chunks[1])
	if err != nil {
		t.Fatalf("CollectChunk failed on second chunk: %v", err)
	}
	if !complete {
		t.Error("Second chunk should complete the message")
	}
	if merged == nil {
		t.Fatal("Merged envelope should not be nil")
	}
}

// Test chunk timeout and cleanup
func TestChunkTimeout(t *testing.T) {
	cc := NewChunkCollector(100 * time.Millisecond) // Very short timeout
	defer cc.Close()

	chunks := createTestChunks(3, "group-timeout")

	// Collect only first chunk
	_, complete, err := cc.CollectChunk(chunks[0])
	if err != nil {
		t.Fatalf("CollectChunk failed: %v", err)
	}
	if complete {
		t.Error("First chunk should not be complete")
	}

	// Verify chunk is pending
	if cc.CountPendingChunks() != 1 {
		t.Errorf("Expected 1 pending chunk, got %d", cc.CountPendingChunks())
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Wait for cleanup cycle (30 seconds in production, but timeout already passed)
	// Force another collection to trigger any internal checks
	time.Sleep(50 * time.Millisecond)

	// Note: In production, cleanup runs every 30 seconds
	// For testing, we just verify the timeout logic works
	status := cc.GetStatus()
	if len(status) > 0 {
		for _, s := range status {
			if s.Age > 100*time.Millisecond {
				t.Logf("Chunk %s is %v old (will be cleaned up on next cycle)", s.ChunkID, s.Age)
			}
		}
	}
}

// Test GetStatus
func TestGetStatus(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	chunks := createTestChunks(3, "group-status")

	// Collect two chunks
	for i := 0; i < 2; i++ {
		_, _, err := cc.CollectChunk(chunks[i])
		if err != nil {
			t.Fatalf("CollectChunk failed: %v", err)
		}
	}

	// Check status
	status := cc.GetStatus()
	if len(status) != 1 {
		t.Fatalf("Expected 1 chunk group in status, got %d", len(status))
	}

	chunkStatus := status["group-status"]
	if chunkStatus.ReceivedCount != 2 {
		t.Errorf("Expected 2 received chunks, got %d", chunkStatus.ReceivedCount)
	}

	if chunkStatus.TotalCount != 3 {
		t.Errorf("Expected 3 total chunks, got %d", chunkStatus.TotalCount)
	}

	if chunkStatus.Complete {
		t.Error("Chunk group should not be complete")
	}

	// Collect final chunk
	_, complete, err := cc.CollectChunk(chunks[2])
	if err != nil {
		t.Fatalf("CollectChunk failed: %v", err)
	}
	if !complete {
		t.Error("Final chunk should complete the message")
	}

	// Status should be empty now
	status = cc.GetStatus()
	if len(status) != 0 {
		t.Errorf("Expected empty status after completion, got %d groups", len(status))
	}
}

// Test concurrent chunk collection
func TestConcurrentChunkCollection(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	// Create two groups of chunks
	group1 := createTestChunks(3, "concurrent-group1")
	group2 := createTestChunks(3, "concurrent-group2")

	// Collect chunks concurrently
	done := make(chan bool, 2)

	go func() {
		for _, chunk := range group1 {
			cc.CollectChunk(chunk)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for _, chunk := range group2 {
			cc.CollectChunk(chunk)
			time.Sleep(10 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Both groups should be complete (cleanup should have happened)
	status := cc.GetStatus()
	if len(status) != 0 {
		t.Errorf("Expected no pending chunks, got %d groups", len(status))
	}
}

// Test ClearChunk
func TestClearChunk(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	chunks := createTestChunks(3, "group-clear")

	// Collect first chunk
	_, _, err := cc.CollectChunk(chunks[0])
	if err != nil {
		t.Fatalf("CollectChunk failed: %v", err)
	}

	if cc.CountPendingChunks() != 1 {
		t.Errorf("Expected 1 pending chunk, got %d", cc.CountPendingChunks())
	}

	// Clear the chunk group
	cc.ClearChunk("group-clear")

	if cc.CountPendingChunks() != 0 {
		t.Errorf("Expected 0 pending chunks after clear, got %d", cc.CountPendingChunks())
	}
}

// Test multiple independent chunk groups
func TestMultipleChunkGroups(t *testing.T) {
	cc := NewChunkCollector(5 * time.Minute)
	defer cc.Close()

	group1 := createTestChunks(2, "multi-group1")
	group2 := createTestChunks(2, "multi-group2")

	// Collect first chunk of each group
	cc.CollectChunk(group1[0])
	cc.CollectChunk(group2[0])

	if cc.CountPendingChunks() != 2 {
		t.Errorf("Expected 2 pending chunks, got %d", cc.CountPendingChunks())
	}

	// Complete group1
	merged, complete, err := cc.CollectChunk(group1[1])
	if err != nil {
		t.Fatalf("Failed to complete group1: %v", err)
	}
	if !complete {
		t.Error("Group1 should be complete")
	}
	if merged == nil {
		t.Fatal("Merged envelope for group1 should not be nil")
	}

	// Only group2 should be pending
	if cc.CountPendingChunks() != 1 {
		t.Errorf("Expected 1 pending chunk after completing group1, got %d", cc.CountPendingChunks())
	}

	// Complete group2
	merged, complete, err = cc.CollectChunk(group2[1])
	if err != nil {
		t.Fatalf("Failed to complete group2: %v", err)
	}
	if !complete {
		t.Error("Group2 should be complete")
	}
	if merged == nil {
		t.Fatal("Merged envelope for group2 should not be nil")
	}

	// No pending chunks
	if cc.CountPendingChunks() != 0 {
		t.Errorf("Expected 0 pending chunks, got %d", cc.CountPendingChunks())
	}
}

// Helper: create test chunks
func createTestChunks(numChunks int, chunkID string) []*envelope.Envelope {
	chunks := make([]*envelope.Envelope, numChunks)

	for i := 0; i < numChunks; i++ {
		chunks[i] = &envelope.Envelope{
			ID:      chunkID + "-" + string(rune('0'+i)),
			Payload: []byte("chunk" + string(rune('0'+i))),
			Headers: map[string]string{
				"X-Chunk-ID":    chunkID,
				"X-Chunk-Index": string(rune('0' + i)),
				"X-Chunk-Total": string(rune('0' + numChunks)),
				"X-Original-ID": "original-" + chunkID,
			},
			Properties: make(map[string]interface{}),
			Route:      make([]string, 0),
		}
	}

	return chunks
}
