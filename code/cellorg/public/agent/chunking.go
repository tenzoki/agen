package agent

import (
	"fmt"
	"sync"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/envelope"
)

// ChunkCollector accumulates chunked envelopes and reassembles them
// This is used by agents to transparently handle chunked messages
type ChunkCollector struct {
	chunks       map[string][]*envelope.Envelope // chunkID -> chunks
	timestamps   map[string]time.Time           // chunkID -> first chunk time
	mu           sync.Mutex
	timeout      time.Duration // How long to wait for incomplete chunks
	stopCleanup  chan bool     // Signal to stop cleanup goroutine
}

// NewChunkCollector creates a new chunk collector with the specified timeout
func NewChunkCollector(timeout time.Duration) *ChunkCollector {
	if timeout == 0 {
		timeout = 5 * time.Minute // Default timeout
	}

	cc := &ChunkCollector{
		chunks:      make(map[string][]*envelope.Envelope),
		timestamps:  make(map[string]time.Time),
		timeout:     timeout,
		stopCleanup: make(chan bool, 1),
	}

	// Start cleanup goroutine
	go cc.cleanupExpiredChunks()

	return cc
}

// CollectChunk accumulates chunks and returns the complete message when ready
// Returns:
// - merged envelope if complete
// - true if complete, false if still waiting
// - error if invalid chunk
func (cc *ChunkCollector) CollectChunk(env *envelope.Envelope) (*envelope.Envelope, bool, error) {
	chunkID := env.Headers["X-Chunk-ID"]
	if chunkID == "" {
		// Not a chunked message, return immediately
		return env, true, nil
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Initialize chunk storage if first chunk
	if cc.chunks[chunkID] == nil {
		cc.chunks[chunkID] = make([]*envelope.Envelope, 0)
		cc.timestamps[chunkID] = time.Now()
	}

	// Check for duplicate chunks
	chunkIndex := env.Headers["X-Chunk-Index"]
	for _, existing := range cc.chunks[chunkID] {
		if existing.Headers["X-Chunk-Index"] == chunkIndex {
			// Duplicate chunk, ignore
			return nil, false, nil
		}
	}

	// Accumulate chunk
	cc.chunks[chunkID] = append(cc.chunks[chunkID], env)

	// Check if complete
	totalChunks := env.Headers["X-Chunk-Total"]
	if len(cc.chunks[chunkID]) == parseChunkTotal(totalChunks) {
		// All chunks received, merge them
		merged, err := envelope.MergeChunks(cc.chunks[chunkID])
		if err != nil {
			// Clean up failed merge
			delete(cc.chunks, chunkID)
			delete(cc.timestamps, chunkID)
			return nil, false, fmt.Errorf("failed to merge chunks: %w", err)
		}

		// Clean up
		delete(cc.chunks, chunkID)
		delete(cc.timestamps, chunkID)

		return merged, true, nil
	}

	// Still waiting for more chunks
	return nil, false, nil
}

// GetStatus returns the current status of chunk collection
func (cc *ChunkCollector) GetStatus() map[string]ChunkStatus {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	status := make(map[string]ChunkStatus)
	for chunkID, chunks := range cc.chunks {
		totalChunks := 0
		if len(chunks) > 0 {
			totalChunks = parseChunkTotal(chunks[0].Headers["X-Chunk-Total"])
		}

		status[chunkID] = ChunkStatus{
			ChunkID:      chunkID,
			ReceivedCount: len(chunks),
			TotalCount:   totalChunks,
			FirstChunkAt: cc.timestamps[chunkID],
			Age:          time.Since(cc.timestamps[chunkID]),
			Complete:     len(chunks) == totalChunks,
		}
	}

	return status
}

// ChunkStatus represents the status of a chunked message
type ChunkStatus struct {
	ChunkID       string
	ReceivedCount int
	TotalCount    int
	FirstChunkAt  time.Time
	Age           time.Duration
	Complete      bool
}

// cleanupExpiredChunks removes incomplete chunks that have exceeded timeout
func (cc *ChunkCollector) cleanupExpiredChunks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cc.stopCleanup:
			return
		case <-ticker.C:
			cc.mu.Lock()
			now := time.Now()
			expired := make([]string, 0)

			for chunkID, timestamp := range cc.timestamps {
				if now.Sub(timestamp) > cc.timeout {
					expired = append(expired, chunkID)
				}
			}

			// Remove expired chunks
			for _, chunkID := range expired {
				delete(cc.chunks, chunkID)
				delete(cc.timestamps, chunkID)
			}

			cc.mu.Unlock()

			// Log expired chunks (could add logger interface here)
			if len(expired) > 0 {
				// In production, this would log to agent's logger
				_ = expired // Prevent unused variable error
			}
		}
	}
}

// Close stops the cleanup goroutine
func (cc *ChunkCollector) Close() {
	select {
	case cc.stopCleanup <- true:
	default:
		// Channel already has a value or is closed
	}
}

// CountPendingChunks returns the number of incomplete chunk groups
func (cc *ChunkCollector) CountPendingChunks() int {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return len(cc.chunks)
}

// ClearChunk manually removes a chunk group (useful for testing)
func (cc *ChunkCollector) ClearChunk(chunkID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	delete(cc.chunks, chunkID)
	delete(cc.timestamps, chunkID)
}

// parseChunkTotal parses the X-Chunk-Total header value
func parseChunkTotal(totalStr string) int {
	var total int
	fmt.Sscanf(totalStr, "%d", &total)
	return total
}
