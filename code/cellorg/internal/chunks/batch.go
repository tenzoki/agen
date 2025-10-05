package chunks

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/storage"
)

// BatchRegisterChunks registers multiple file splits in a single transaction
// Optimized for parallel chunk processing scenarios
func (ct *ChunkTracker) BatchRegisterChunks(fileSplits []*FileSplitBatch) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.logger.Info("Starting batch chunk registration", "batch_size", len(fileSplits))

	// Prepare all vertices and edges for batch creation
	var vertices []storage.BatchVertex
	var edges []storage.BatchEdge

	for _, split := range fileSplits {
		fileVertexID, chunkVertices, chunkEdges := ct.prepareSplitData(split)

		// Add file vertex
		vertices = append(vertices, storage.BatchVertex{
			ID:         fileVertexID,
			Label:      "file_original",
			Properties: split.FileVertex,
		})

		// Add chunk vertices
		vertices = append(vertices, chunkVertices...)

		// Add all edges for this split
		edges = append(edges, chunkEdges...)
	}

	// Execute batch operations
	if err := ct.storageClient.BatchCreateVertices(vertices); err != nil {
		return fmt.Errorf("failed to batch create vertices: %w", err)
	}

	if err := ct.storageClient.BatchCreateEdges(edges); err != nil {
		return fmt.Errorf("failed to batch create edges: %w", err)
	}

	ct.logger.Info("Batch chunk registration completed", "files", len(fileSplits))
	return nil
}

// BatchGetChunksForProcessing retrieves chunks for multiple files optimized for parallel processing
func (ct *ChunkTracker) BatchGetChunksForProcessing(fileHashes []string) (map[string][]*ChunkInfo, error) {
	// Build batch query for multiple files
	hashList := strings.Join(fileHashes, "', '")
	query := fmt.Sprintf(`
		g.V().has('id', within('file:original:%s'))
		.as('file')
		.out('HAS_CHUNK')
		.as('chunk')
		.select('file', 'chunk')
		.by(valueMap())
		.by(valueMap())
		.order().by(select('chunk').values('chunk_index'), asc)
	`, hashList)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to batch query chunks: %w", err)
	}

	return ct.parseParallelChunkResults(results)
}

// GetChunksForParallelProcessing retrieves chunks with processing metadata
func (ct *ChunkTracker) GetChunksForParallelProcessing(fileHash string) (*ParallelChunkSet, error) {
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.as('file')
		.out('HAS_CHUNK')
		.as('chunk')
		.project('file_info', 'chunk_info', 'dependencies')
		.by(select('file').valueMap())
		.by(select('chunk').valueMap())
		.by(select('chunk').out('DEPENDS_ON').valueMap().fold())
		.order().by(select('chunk_info').select('chunk_index'), asc)
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query parallel chunks: %w", err)
	}

	return ct.parseParallelChunkSet(results)
}

// GetIndependentChunks finds chunks that can be processed asynchronously
func (ct *ChunkTracker) GetIndependentChunks(fileHash string) ([]*ParallelChunkInfo, error) {
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.has('can_process_async', true)
		.has('status', 'created')
		.valueMap()
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, err
	}

	return ct.parseAsyncChunks(results)
}

// GetChunksReadyForProcessing finds chunks whose dependencies are satisfied
func (ct *ChunkTracker) GetChunksReadyForProcessing(fileHash string) ([]*ParallelChunkInfo, error) {
	// Find chunks whose dependencies are satisfied
	query := fmt.Sprintf(`
		g.V().has('id', 'file:original:%s')
		.out('HAS_CHUNK')
		.where(
			out('DEPENDS_ON').has('status', 'completed').count().as('deps_done')
			.select('deps_done').is(
				__.out('DEPENDS_ON').count()
			)
		)
		.has('status', 'created')
		.valueMap()
	`, fileHash)

	results, err := ct.storageClient.GraphQuery(query)
	if err != nil {
		return nil, err
	}

	return ct.parseReadyChunks(results)
}

// DistributeChunksToAgents distributes chunks across multiple processing agents
func (ct *ChunkTracker) DistributeChunksToAgents(fileHash string, agentCount int) (map[string][]*ChunkInfo, error) {
	chunks, err := ct.GetChunksForParallelProcessing(fileHash)
	if err != nil {
		return nil, err
	}

	distribution := make(map[string][]*ChunkInfo)

	// Round-robin distribution of independent chunks
	independentChunks := ct.filterIndependentChunks(chunks.Chunks)
	for i, chunk := range independentChunks {
		agentID := fmt.Sprintf("agent-%d", i%agentCount)
		distribution[agentID] = append(distribution[agentID], chunk.ChunkInfo)
	}

	return distribution, nil
}

// BatchUpdateChunkStatus updates status for multiple chunks in one operation
func (ct *ChunkTracker) BatchUpdateChunkStatus(updates []ChunkStatusUpdate) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	// Group updates by status for efficient batch operations
	statusGroups := make(map[string][]string) // status -> chunk_ids

	for _, update := range updates {
		chunkID := fmt.Sprintf("chunk:%s:%d", update.FileHash, update.ChunkIndex)
		statusGroups[update.NewStatus] = append(statusGroups[update.NewStatus], chunkID)
	}

	// Execute batch updates per status
	for status, chunkIDs := range statusGroups {
		if err := ct.storageClient.BatchUpdateVertexProperties(chunkIDs, map[string]interface{}{
			"status":       status,
			"processed_at": time.Now().Format(time.RFC3339),
		}); err != nil {
			return fmt.Errorf("failed to batch update status %s: %w", status, err)
		}
	}

	ct.logger.Info("Batch status update completed", "total_updates", len(updates))
	return nil
}

// Helper methods for batch operations and parallel processing optimization

// prepareSplitData prepares vertices and edges for batch creation
func (ct *ChunkTracker) prepareSplitData(split *FileSplitBatch) (string, []storage.BatchVertex, []storage.BatchEdge) {
	fileVertexID := fmt.Sprintf("file:original:%s", split.FileInfo.Hash)

	var chunkVertices []storage.BatchVertex
	var edges []storage.BatchEdge
	var previousChunkID string

	for _, chunk := range split.Chunks {
		chunkVertexID := fmt.Sprintf("chunk:%s:%d", split.FileInfo.Hash, chunk.Index)

		// Prepare chunk vertex
		chunkVertex := storage.BatchVertex{
			ID:    chunkVertexID,
			Label: "file_chunk",
			Properties: map[string]interface{}{
				"chunk_index":       chunk.Index,
				"chunk_hash":        chunk.Hash,
				"size":              chunk.Size,
				"start_offset":      chunk.StartOffset,
				"end_offset":        chunk.EndOffset,
				"status":            "created",
				"can_process_async": ct.canProcessAsync(chunk, split.Chunks),
			},
		}
		chunkVertices = append(chunkVertices, chunkVertex)

		// File -> Chunk edge
		edges = append(edges, storage.BatchEdge{
			From:  fileVertexID,
			To:    chunkVertexID,
			Label: "HAS_CHUNK",
		})

		// Ordering edges
		if previousChunkID != "" {
			edges = append(edges,
				storage.BatchEdge{From: previousChunkID, To: chunkVertexID, Label: "NEXT_CHUNK"},
				storage.BatchEdge{From: chunkVertexID, To: previousChunkID, Label: "PREV_CHUNK"},
			)
		}

		previousChunkID = chunkVertexID
	}

	return fileVertexID, chunkVertices, edges
}

// canProcessAsync determines if a chunk can be processed independently
func (ct *ChunkTracker) canProcessAsync(chunk *ChunkInfo, allChunks []*ChunkInfo) bool {
	// Simple heuristic: chunks that don't span across logical boundaries can be async
	// This would be customized based on file type and processing requirements
	return chunk.Size < 1024*1024 // < 1MB chunks can typically be processed async
}

// parseParallelChunkResults processes batch query results for parallel processing
func (ct *ChunkTracker) parseParallelChunkResults(results []interface{}) (map[string][]*ChunkInfo, error) {
	fileChunks := make(map[string][]*ChunkInfo)

	for _, result := range results {
		// Parse the result which contains file and chunk information
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		fileInfo := resultMap["file"].(map[string]interface{})
		chunkInfo := resultMap["chunk"].(map[string]interface{})

		fileHash := fileInfo["hash"].(string)

		chunk := &ChunkInfo{
			Index:       int(chunkInfo["chunk_index"].(float64)),
			Hash:        chunkInfo["chunk_hash"].(string),
			Size:        int64(chunkInfo["size"].(float64)),
			StartOffset: int64(chunkInfo["start_offset"].(float64)),
			EndOffset:   int64(chunkInfo["end_offset"].(float64)),
		}

		fileChunks[fileHash] = append(fileChunks[fileHash], chunk)
	}

	return fileChunks, nil
}

// parseParallelChunkSet builds a parallel chunk set with dependency information
func (ct *ChunkTracker) parseParallelChunkSet(results []interface{}) (*ParallelChunkSet, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no chunks found")
	}

	var fileInfo *FileInfo
	var chunks []*ParallelChunkInfo
	dependencies := make(map[string][]string)

	for _, result := range results {
		resultMap := result.(map[string]interface{})

		// Extract file info (same for all chunks)
		if fileInfo == nil {
			fileData := resultMap["file_info"].(map[string]interface{})
			fileInfo = &FileInfo{
				Hash: fileData["hash"].(string),
				Path: fileData["path"].(string),
				Size: int64(fileData["size"].(float64)),
			}
		}

		// Extract chunk info
		chunkData := resultMap["chunk_info"].(map[string]interface{})
		deps := resultMap["dependencies"].([]interface{})

		chunk := &ParallelChunkInfo{
			ChunkInfo: &ChunkInfo{
				Index:       int(chunkData["chunk_index"].(float64)),
				Hash:        chunkData["chunk_hash"].(string),
				Size:        int64(chunkData["size"].(float64)),
				StartOffset: int64(chunkData["start_offset"].(float64)),
				EndOffset:   int64(chunkData["end_offset"].(float64)),
			},
			CanProcessAsync: chunkData["can_process_async"].(bool),
			ProcessingMeta:  make(map[string]interface{}),
		}

		// Process dependencies
		chunkID := fmt.Sprintf("chunk:%s:%d", fileInfo.Hash, chunk.Index)
		for _, dep := range deps {
			depMap := dep.(map[string]interface{})
			depID := depMap["id"].(string)
			chunk.Dependencies = append(chunk.Dependencies, depID)
		}
		dependencies[chunkID] = chunk.Dependencies

		chunks = append(chunks, chunk)
	}

	return &ParallelChunkSet{
		FileInfo:     fileInfo,
		Chunks:       chunks,
		TotalChunks:  len(chunks),
		Dependencies: dependencies,
	}, nil
}

// parseAsyncChunks parses chunks that can be processed asynchronously
func (ct *ChunkTracker) parseAsyncChunks(results []interface{}) ([]*ParallelChunkInfo, error) {
	var chunks []*ParallelChunkInfo

	for _, result := range results {
		resultMap := result.(map[string]interface{})

		chunk := &ParallelChunkInfo{
			ChunkInfo: &ChunkInfo{
				Index:       int(resultMap["chunk_index"].(float64)),
				Hash:        resultMap["chunk_hash"].(string),
				Size:        int64(resultMap["size"].(float64)),
				StartOffset: int64(resultMap["start_offset"].(float64)),
				EndOffset:   int64(resultMap["end_offset"].(float64)),
				Status:      resultMap["status"].(string),
			},
			CanProcessAsync: true,
			ProcessingMeta:  make(map[string]interface{}),
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// parseReadyChunks parses chunks that are ready for processing
func (ct *ChunkTracker) parseReadyChunks(results []interface{}) ([]*ParallelChunkInfo, error) {
	var chunks []*ParallelChunkInfo

	for _, result := range results {
		resultMap := result.(map[string]interface{})

		chunk := &ParallelChunkInfo{
			ChunkInfo: &ChunkInfo{
				Index:       int(resultMap["chunk_index"].(float64)),
				Hash:        resultMap["chunk_hash"].(string),
				Size:        int64(resultMap["size"].(float64)),
				StartOffset: int64(resultMap["start_offset"].(float64)),
				EndOffset:   int64(resultMap["end_offset"].(float64)),
				Status:      resultMap["status"].(string),
			},
			ProcessingMeta: make(map[string]interface{}),
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// filterIndependentChunks filters chunks that can be processed independently
func (ct *ChunkTracker) filterIndependentChunks(chunks []*ParallelChunkInfo) []*ParallelChunkInfo {
	var independent []*ParallelChunkInfo

	for _, chunk := range chunks {
		if chunk.CanProcessAsync && len(chunk.Dependencies) == 0 {
			independent = append(independent, chunk)
		}
	}

	return independent
}

// ProcessChunksConcurrently processes multiple chunks concurrently with coordination
func (ct *ChunkTracker) ProcessChunksConcurrently(fileHash string, maxConcurrency int, processorFunc func(*ChunkInfo) error) error {
	chunks, err := ct.GetIndependentChunks(fileHash)
	if err != nil {
		return fmt.Errorf("failed to get independent chunks: %w", err)
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	errChan := make(chan error, len(chunks))

	for _, chunk := range chunks {
		wg.Add(1)
		go func(c *ParallelChunkInfo) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Update status to processing
			if err := ct.UpdateChunkStatus(fileHash, c.Index, "processing", "parallel-processor"); err != nil {
				errChan <- err
				return
			}

			// Process the chunk
			if err := processorFunc(c.ChunkInfo); err != nil {
				// Mark as failed
				ct.UpdateChunkStatus(fileHash, c.Index, "failed", "parallel-processor")
				errChan <- err
				return
			}

			// Mark as completed
			if err := ct.UpdateChunkStatus(fileHash, c.Index, "completed", "parallel-processor"); err != nil {
				errChan <- err
				return
			}
		}(chunk)
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
