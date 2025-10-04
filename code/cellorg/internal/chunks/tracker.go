package chunks

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/agen/cellorg/internal/storage"
)

// StorageInterface defines the storage operations needed by ChunkTracker
type StorageInterface interface {
	CreateVertex(label string, properties map[string]interface{}) (string, error)
	CreateEdge(from, to, label string) error
	GraphQuery(query string) ([]interface{}, error)
	UpdateVertexProperties(vertexID string, updates map[string]interface{}) error
	BatchCreateVertices(vertices []storage.BatchVertex) error
	BatchCreateEdges(edges []storage.BatchEdge) error
	BatchUpdateVertexProperties(vertexIDs []string, updates map[string]interface{}) error
	ParallelGraphQuery(queries []string) ([][]interface{}, error)
	RetrieveFile(fileID string) ([]byte, error)
	StoreFile(data []byte, metadata map[string]interface{}) (string, error)
}

// ChunkTracker manages file chunk relationships in the graph store
type ChunkTracker struct {
	storageClient StorageInterface
	logger        *slog.Logger
	mu            sync.RWMutex
}

// FileInfo represents metadata about an original file
type FileInfo struct {
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
	ChunkCount int       `json:"chunk_count"`
}

// ChunkInfo represents a file chunk
type ChunkInfo struct {
	Index       int    `json:"index"`
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	StartOffset int64  `json:"start_offset"`
	EndOffset   int64  `json:"end_offset"`
	Status      string `json:"status"` // created, processing, completed, failed
}

// SplitMetadata contains information about the splitting operation
type SplitMetadata struct {
	SplitMethod    string                 `json:"split_method"` // byte_size, logical_boundary, etc.
	ChunkSize      int64                  `json:"chunk_size"`
	TotalChunks    int                    `json:"total_chunks"`
	CreatedAt      time.Time              `json:"created_at"`
	CreatedBy      string                 `json:"created_by"`
	CustomMetadata map[string]interface{} `json:"custom_metadata"`
}

// Batch operation types for parallel processing optimization
type FileSplitBatch struct {
	FileInfo   *FileInfo
	Chunks     []*ChunkInfo
	Metadata   *SplitMetadata
	FileVertex map[string]interface{}
}

type ParallelChunkSet struct {
	FileInfo     *FileInfo
	Chunks       []*ParallelChunkInfo
	TotalChunks  int
	Dependencies map[string][]string // chunk_id -> dependencies
}

type ParallelChunkInfo struct {
	*ChunkInfo
	Dependencies    []string               // Other chunks this depends on
	ProcessingMeta  map[string]interface{} // Metadata for parallel processing
	CanProcessAsync bool                   // Whether this chunk can be processed independently
}

type ChunkProcessingMetrics struct {
	TotalChunks           int
	ProcessedChunks       int
	FailedChunks          int
	ParallelProcessors    int
	AverageProcessingTime time.Duration
	ThroughputPerSecond   float64
	BatchEfficiency       float64
}

type ChunkStatusUpdate struct {
	FileHash   string
	ChunkIndex int
	NewStatus  string
	AgentID    string
}

type ProcessingProgress struct {
	FileHash        string         `json:"file_hash"`
	TotalChunks     int            `json:"total_chunks"`
	StatusCounts    map[string]int `json:"status_counts"`
	CompletedChunks int            `json:"completed_chunks"`
	FailedChunks    int            `json:"failed_chunks"`
	ProgressPercent float64        `json:"progress_percent"`
	Throughput      float64        `json:"throughput"`
	EstimatedTime   time.Duration  `json:"estimated_time"`
}

// NewChunkTracker creates a new chunk tracker instance
func NewChunkTracker(storageClient StorageInterface, logger *slog.Logger) *ChunkTracker {
	if logger == nil {
		logger = slog.Default()
	}

	return &ChunkTracker{
		storageClient: storageClient,
		logger:        logger,
	}
}

// RegisterFileSplit registers a file split operation in the graph store
func (ct *ChunkTracker) RegisterFileSplit(originalFile *FileInfo, chunks []*ChunkInfo, metadata *SplitMetadata) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.logger.Info("Registering file split", "file_hash", originalFile.Hash, "chunk_count", len(chunks))

	// Create file vertex
	fileVertex := map[string]interface{}{
		"id":           fmt.Sprintf("file:original:%s", originalFile.Hash),
		"type":         "file_original",
		"path":         originalFile.Path,
		"hash":         originalFile.Hash,
		"size":         originalFile.Size,
		"mime_type":    originalFile.MimeType,
		"created_at":   originalFile.CreatedAt.Format(time.RFC3339),
		"modified_at":  originalFile.ModifiedAt.Format(time.RFC3339),
		"chunk_count":  len(chunks),
		"split_method": metadata.SplitMethod,
		"chunk_size":   metadata.ChunkSize,
	}

	fileVertexID, err := ct.storageClient.CreateVertex("file_original", fileVertex)
	if err != nil {
		return fmt.Errorf("failed to create file vertex: %w", err)
	}

	// Create metadata vertex
	metadataVertex := map[string]interface{}{
		"id":              fmt.Sprintf("metadata:%s", originalFile.Hash),
		"type":            "split_metadata",
		"split_method":    metadata.SplitMethod,
		"chunk_size":      metadata.ChunkSize,
		"total_chunks":    metadata.TotalChunks,
		"created_at":      metadata.CreatedAt.Format(time.RFC3339),
		"created_by":      metadata.CreatedBy,
		"custom_metadata": metadata.CustomMetadata,
	}

	metadataVertexID, err := ct.storageClient.CreateVertex("split_metadata", metadataVertex)
	if err != nil {
		return fmt.Errorf("failed to create metadata vertex: %w", err)
	}

	// Link file to metadata
	err = ct.storageClient.CreateEdge(fileVertexID, metadataVertexID, "HAS_METADATA")
	if err != nil {
		return fmt.Errorf("failed to link metadata: %w", err)
	}

	// Create chunk vertices and edges
	var previousChunkID string
	for i, chunk := range chunks {
		chunkVertex := map[string]interface{}{
			"id":           fmt.Sprintf("chunk:%s:%d", originalFile.Hash, chunk.Index),
			"type":         "file_chunk",
			"chunk_index":  chunk.Index,
			"chunk_hash":   chunk.Hash,
			"size":         chunk.Size,
			"start_offset": chunk.StartOffset,
			"end_offset":   chunk.EndOffset,
			"status":       "created",
		}

		chunkVertexID, err := ct.storageClient.CreateVertex("file_chunk", chunkVertex)
		if err != nil {
			return fmt.Errorf("failed to create chunk vertex %d: %w", i, err)
		}

		// Link file to chunk
		err = ct.storageClient.CreateEdge(fileVertexID, chunkVertexID, "HAS_CHUNK")
		if err != nil {
			return fmt.Errorf("failed to link chunk %d: %w", i, err)
		}

		// Create bi-directional ordering edges
		if previousChunkID != "" {
			// Forward link: previous -> current
			err = ct.storageClient.CreateEdge(previousChunkID, chunkVertexID, "NEXT_CHUNK")
			if err != nil {
				return fmt.Errorf("failed to create forward ordering edge %d: %w", i, err)
			}

			// Backward link: current -> previous
			err = ct.storageClient.CreateEdge(chunkVertexID, previousChunkID, "PREV_CHUNK")
			if err != nil {
				return fmt.Errorf("failed to create backward ordering edge %d: %w", i, err)
			}
		}

		previousChunkID = chunkVertexID
	}

	ct.logger.Info("File split registered successfully", "file_hash", originalFile.Hash, "chunk_count", len(chunks))
	return nil
}

// GetFileChunksInOrder retrieves all chunks for a file in correct order
func (ct *ChunkTracker) GetFileChunksInOrder(fileHash string) ([]*ChunkInfo, error) {
	// Graph query to get chunks in order
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.order().by('chunk_index', asc)
		.valueMap()
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}

	chunks := make([]*ChunkInfo, len(results))
	for i, result := range results {
		resultMap := result.(map[string]interface{})

		chunk := &ChunkInfo{
			Index:       int(resultMap["chunk_index"].(float64)),
			Hash:        resultMap["chunk_hash"].(string),
			Size:        int64(resultMap["size"].(float64)),
			StartOffset: int64(resultMap["start_offset"].(float64)),
			EndOffset:   int64(resultMap["end_offset"].(float64)),
			Status:      resultMap["status"].(string),
		}

		chunks[i] = chunk
	}

	return chunks, nil
}

// UpdateChunkStatus updates the processing status of a chunk
func (ct *ChunkTracker) UpdateChunkStatus(fileHash string, chunkIndex int, status string, processingAgent string) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	chunkID := fmt.Sprintf("chunk:%s:%d", fileHash, chunkIndex)

	// Get current status for logging
	currentStatusQuery := fmt.Sprintf(`
		g.V().has('id', '%s').values('status')
	`, chunkID)

	statusResults, err := ct.storageClient.GraphQuery(currentStatusQuery)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	var currentStatus string
	if len(statusResults) > 0 {
		currentStatus = statusResults[0].(string)
	}

	// Create status change event
	eventVertex := map[string]interface{}{
		"id":         fmt.Sprintf("event:%s", uuid.New().String()),
		"type":       "status_change",
		"chunk_id":   chunkID,
		"from_state": currentStatus,
		"to_state":   status,
		"timestamp":  time.Now().Format(time.RFC3339),
		"agent_id":   processingAgent,
	}

	eventVertexID, err := ct.storageClient.CreateVertex("chunk_event", eventVertex)
	if err != nil {
		return fmt.Errorf("failed to create event vertex: %w", err)
	}

	// Link event to chunk
	err = ct.storageClient.CreateEdge(chunkID, eventVertexID, "HAS_EVENT")
	if err != nil {
		return fmt.Errorf("failed to link event: %w", err)
	}

	// Update chunk properties
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": time.Now().Format(time.RFC3339),
		"processed_by": processingAgent,
	}

	err = ct.storageClient.UpdateVertexProperties(chunkID, updates)
	if err != nil {
		return fmt.Errorf("failed to update chunk status: %w", err)
	}

	ct.logger.Debug("Chunk status updated",
		"file_hash", fileHash,
		"chunk_index", chunkIndex,
		"from_status", currentStatus,
		"to_status", status,
		"agent", processingAgent)

	return nil
}

// GetNextUnprocessedChunk finds the next chunk ready for processing
func (ct *ChunkTracker) GetNextUnprocessedChunk(fileHash string) (*ChunkInfo, error) {
	// Find the next chunk with status "created" in order
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.has('status', 'created')
		.order().by('chunk_index', asc)
		.limit(1)
		.valueMap()
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query next chunk: %w", err)
	}

	if len(results) == 0 {
		return nil, nil // No unprocessed chunks
	}

	resultMap := results[0].(map[string]interface{})

	chunk := &ChunkInfo{
		Index:       int(resultMap["chunk_index"].(float64)),
		Hash:        resultMap["chunk_hash"].(string),
		Size:        int64(resultMap["size"].(float64)),
		StartOffset: int64(resultMap["start_offset"].(float64)),
		EndOffset:   int64(resultMap["end_offset"].(float64)),
		Status:      resultMap["status"].(string),
	}

	return chunk, nil
}

// GetProcessingProgress returns the current processing progress for a file
func (ct *ChunkTracker) GetProcessingProgress(fileHash string) (*ProcessingProgress, error) {
	// Query to get status counts
	statusQuery := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.groupCount().by('status')
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query status counts: %w", err)
	}

	statusCounts := make(map[string]int)
	if len(results) > 0 {
		for status, count := range results[0].(map[string]interface{}) {
			statusCounts[status] = int(count.(float64))
		}
	}

	totalChunks := 0
	for _, count := range statusCounts {
		totalChunks += count
	}

	progress := &ProcessingProgress{
		FileHash:        fileHash,
		TotalChunks:     totalChunks,
		StatusCounts:    statusCounts,
		CompletedChunks: statusCounts["completed"],
		FailedChunks:    statusCounts["failed"],
	}

	if totalChunks > 0 {
		progress.ProgressPercent = float64(progress.CompletedChunks) / float64(totalChunks) * 100
	}

	// Calculate throughput if there are processed chunks
	if processed := statusCounts["completed"]; processed > 0 {
		progress.Throughput = ct.calculateThroughput(fileHash)
	}

	return progress, nil
}

// calculateThroughput calculates processing throughput based on event history
func (ct *ChunkTracker) calculateThroughput(fileHash string) float64 {
	// Query recent completion events
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.out('HAS_EVENT')
		.has('to_state', 'completed')
		.order().by('timestamp', desc)
		.limit(10)
		.values('timestamp')
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil || len(results) < 2 {
		return 0.0
	}

	// Calculate time difference between first and last events
	timestamps := make([]time.Time, len(results))
	for i, result := range results {
		timestamp, err := time.Parse(time.RFC3339, result.(string))
		if err != nil {
			return 0.0
		}
		timestamps[i] = timestamp
	}

	duration := timestamps[0].Sub(timestamps[len(timestamps)-1])
	if duration.Seconds() == 0 {
		return 0.0
	}

	return float64(len(results)) / duration.Seconds()
}
