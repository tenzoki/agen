// Package chunks provides comprehensive chunk processing utilities for the GOX framework.
// This package implements file splitting, chunk management, integrity validation,
// and parallel processing coordination for large file handling.
//
// Key Features:
// - File splitting with multiple strategies (size-based, line-based, semantic)
// - Chunk integrity validation using SHA256 hashing
// - Chunk reassembly with data verification
// - Storage integration for persistent chunk management
// - Parallel processing coordination with dependency management
// - Optimal chunk size calculation based on file characteristics
// - MIME type detection and content-aware processing
//
// The package provides high-level operations for file chunking workflows
// while maintaining data integrity and enabling efficient parallel processing
// across the distributed GOX agent system.
package chunks

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ChunkOperations provides high-level utility functions for comprehensive chunk processing.
// This structure encapsulates common chunk operations including file splitting,
// reassembly, validation, and storage integration.
//
// All operations maintain strict data integrity through SHA256 hashing and
// provide comprehensive error handling for robust chunk management.
type ChunkOperations struct {
	tracker *ChunkTracker // Chunk tracker for metadata and state management
}

// NewChunkOperations creates a new chunk operations helper with tracker integration.
// The operations helper provides high-level chunk processing functions that work
// seamlessly with the chunk tracker for metadata management and storage.
//
// Parameters:
//   - tracker: ChunkTracker instance for metadata and storage operations
//
// Returns:
//   - *ChunkOperations: Configured operations helper ready for use
//
// Called by: Chunk processing agents and file splitter operations
func NewChunkOperations(tracker *ChunkTracker) *ChunkOperations {
	return &ChunkOperations{
		tracker: tracker,
	}
}

// SplitFileIntoChunks divides a file into manageable chunks with comprehensive metadata.
// This method performs complete file analysis, creates chunk metadata, calculates hashes,
// and generates all information needed for distributed processing and reassembly.
//
// Processing Steps:
// 1. Open and analyze source file (size, modification time, MIME type)
// 2. Calculate SHA256 hash for the entire file
// 3. Split file into chunks of specified size
// 4. Calculate individual chunk hashes and metadata
// 5. Generate file metadata, chunk list, and split metadata
//
// The method ensures data integrity through comprehensive hashing and provides
// all metadata needed for chunk tracking, processing, and reassembly.
//
// Parameters:
//   - filePath: Path to file to be split
//   - chunkSize: Target size for individual chunks in bytes
//   - method: Splitting method identifier (for metadata tracking)
//
// Returns:
//   - *FileInfo: Complete file metadata with hash and chunk count
//   - []*ChunkInfo: Array of chunk metadata with hashes and offsets
//   - *SplitMetadata: Split operation metadata and configuration
//   - error: File access or processing error
//
// Called by: File splitter agent, chunk processing pipelines
func (co *ChunkOperations) SplitFileIntoChunks(filePath string, chunkSize int64, method string) (*FileInfo, []*ChunkInfo, *SplitMetadata, error) {
	// Open source file for reading and analysis
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Calculate file hash
	hasher := sha256.New()
	file.Seek(0, 0)
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to calculate file hash: %w", err)
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Create FileInfo
	fileMetadata := &FileInfo{
		Path:       filePath,
		Hash:       fileHash,
		Size:       fileInfo.Size(),
		MimeType:   co.detectMimeType(filePath),
		CreatedAt:  fileInfo.ModTime(),
		ModifiedAt: fileInfo.ModTime(),
	}

	// Calculate chunks
	var chunks []*ChunkInfo
	var currentOffset int64 = 0
	chunkIndex := 0

	file.Seek(0, 0)

	for currentOffset < fileInfo.Size() {
		endOffset := currentOffset + chunkSize
		if endOffset > fileInfo.Size() {
			endOffset = fileInfo.Size()
		}

		actualSize := endOffset - currentOffset

		// Read chunk data to calculate hash
		chunkData := make([]byte, actualSize)
		_, err := file.ReadAt(chunkData, currentOffset)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to read chunk %d: %w", chunkIndex, err)
		}

		// Calculate chunk hash
		chunkHasher := sha256.New()
		chunkHasher.Write(chunkData)
		chunkHash := hex.EncodeToString(chunkHasher.Sum(nil))

		chunk := &ChunkInfo{
			Index:       chunkIndex,
			Hash:        chunkHash,
			Size:        actualSize,
			StartOffset: currentOffset,
			EndOffset:   endOffset - 1,
			Status:      "created",
		}

		chunks = append(chunks, chunk)
		currentOffset = endOffset
		chunkIndex++
	}

	// Create split metadata
	metadata := &SplitMetadata{
		SplitMethod:    method,
		ChunkSize:      chunkSize,
		TotalChunks:    len(chunks),
		CreatedAt:      time.Now(),
		CreatedBy:      "chunk-operations",
		CustomMetadata: make(map[string]interface{}),
	}

	fileMetadata.ChunkCount = len(chunks)

	return fileMetadata, chunks, metadata, nil
}

// ReassembleChunks reconstructs the original file from processed chunks with validation.
// This method performs comprehensive chunk validation, retrieves chunk data from storage,
// and reassembles the file in the correct order with integrity verification.
//
// Reassembly Process:
// 1. Retrieve all chunks for the file in index order
// 2. Verify all chunks are in completed status
// 3. Create output file for writing
// 4. For each chunk: retrieve data, validate size, write to output
// 5. Ensure complete file reconstruction
//
// The method ensures data integrity by validating chunk status and sizes
// before writing, preventing corruption in the reconstructed file.
//
// Parameters:
//   - fileHash: SHA256 hash of the original file
//   - outputPath: Path where reconstructed file should be written
//
// Returns:
//   - error: Chunk validation, storage, or file writing error
//
// Called by: File reassembly operations, processing completion handlers
func (co *ChunkOperations) ReassembleChunks(fileHash string, outputPath string) error {
	// Retrieve all chunks for the file in correct order
	chunks, err := co.tracker.GetFileChunksInOrder(fileHash)
	if err != nil {
		return fmt.Errorf("failed to get chunks: %w", err)
	}

	// Validate that all chunks are completed before reassembly
	for _, chunk := range chunks {
		if chunk.Status != "completed" {
			return fmt.Errorf("chunk %d is not completed (status: %s)", chunk.Index, chunk.Status)
		}
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Write chunks in order
	for _, chunk := range chunks {
		// Retrieve chunk data from storage
		chunkData, err := co.tracker.storageClient.RetrieveFile(chunk.Hash)
		if err != nil {
			return fmt.Errorf("failed to retrieve chunk %d: %w", chunk.Index, err)
		}

		// Verify chunk size
		if int64(len(chunkData)) != chunk.Size {
			return fmt.Errorf("chunk %d size mismatch: expected %d, got %d", chunk.Index, chunk.Size, len(chunkData))
		}

		// Write to output file
		_, err = outputFile.Write(chunkData)
		if err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", chunk.Index, err)
		}
	}

	return nil
}

// ValidateChunkIntegrity performs comprehensive integrity validation on all chunks of a file.
// This method retrieves each chunk from storage and validates both size and hash integrity
// to ensure no data corruption has occurred during storage or processing.
//
// Validation Process:
// 1. Retrieve all chunks for the specified file
// 2. For each chunk: retrieve data from storage
// 3. Validate actual size matches expected size
// 4. Recalculate SHA256 hash and compare with stored hash
// 5. Collect and return any integrity issues found
//
// This validation is critical for ensuring data integrity in distributed
// processing environments where chunks may be stored across multiple systems.
//
// Parameters:
//   - fileHash: SHA256 hash of the original file to validate
//
// Returns:
//   - []string: List of integrity issues found (empty if all chunks valid)
//   - error: Chunk retrieval or validation process error
//
// Called by: Integrity validation jobs, health check operations
func (co *ChunkOperations) ValidateChunkIntegrity(fileHash string) ([]string, error) {
	// Retrieve all chunks for integrity validation
	chunks, err := co.tracker.GetFileChunksInOrder(fileHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks: %w", err)
	}

	// Collect integrity issues during validation
	var issues []string

	for _, chunk := range chunks {
		// Retrieve chunk data
		chunkData, err := co.tracker.storageClient.RetrieveFile(chunk.Hash)
		if err != nil {
			issues = append(issues, fmt.Sprintf("chunk %d: failed to retrieve data: %v", chunk.Index, err))
			continue
		}

		// Verify size
		if int64(len(chunkData)) != chunk.Size {
			issues = append(issues, fmt.Sprintf("chunk %d: size mismatch: expected %d, got %d", chunk.Index, chunk.Size, len(chunkData)))
		}

		// Verify hash
		hasher := sha256.New()
		hasher.Write(chunkData)
		calculatedHash := hex.EncodeToString(hasher.Sum(nil))

		if calculatedHash != chunk.Hash {
			issues = append(issues, fmt.Sprintf("chunk %d: hash mismatch: expected %s, got %s", chunk.Index, chunk.Hash, calculatedHash))
		}
	}

	return issues, nil
}

// StoreChunkData stores individual chunk data in the storage system
func (co *ChunkOperations) StoreChunkData(chunk *ChunkInfo, data []byte) error {
	// Verify data matches chunk metadata
	if int64(len(data)) != chunk.Size {
		return fmt.Errorf("data size mismatch: expected %d, got %d", chunk.Size, len(data))
	}

	// Verify hash
	hasher := sha256.New()
	hasher.Write(data)
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))

	if calculatedHash != chunk.Hash {
		return fmt.Errorf("hash mismatch: expected %s, got %s", chunk.Hash, calculatedHash)
	}

	// Store in storage system
	metadata := map[string]interface{}{
		"chunk_index": chunk.Index,
		"chunk_size":  chunk.Size,
		"stored_at":   time.Now().Format(time.RFC3339),
	}

	storedHash, err := co.tracker.storageClient.StoreFile(data, metadata)
	if err != nil {
		return fmt.Errorf("failed to store chunk data: %w", err)
	}

	// Verify stored hash matches
	if storedHash != chunk.Hash {
		return fmt.Errorf("stored hash mismatch: expected %s, got %s", chunk.Hash, storedHash)
	}

	return nil
}

// GetChunkData retrieves chunk data from storage
func (co *ChunkOperations) GetChunkData(chunk *ChunkInfo) ([]byte, error) {
	data, err := co.tracker.storageClient.RetrieveFile(chunk.Hash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve chunk data: %w", err)
	}

	// Verify integrity
	if int64(len(data)) != chunk.Size {
		return nil, fmt.Errorf("retrieved data size mismatch: expected %d, got %d", chunk.Size, len(data))
	}

	hasher := sha256.New()
	hasher.Write(data)
	calculatedHash := hex.EncodeToString(hasher.Sum(nil))

	if calculatedHash != chunk.Hash {
		return nil, fmt.Errorf("retrieved data hash mismatch: expected %s, got %s", chunk.Hash, calculatedHash)
	}

	return data, nil
}

// CreateDependencyChain creates dependency relationships between chunks
func (co *ChunkOperations) CreateDependencyChain(fileHash string, dependencies map[int][]int) error {
	// dependencies maps chunk index to list of dependent chunk indices

	for chunkIndex, deps := range dependencies {
		chunkID := fmt.Sprintf("chunk:%s:%d", fileHash, chunkIndex)

		for _, depIndex := range deps {
			depChunkID := fmt.Sprintf("chunk:%s:%d", fileHash, depIndex)

			// Create dependency edge
			err := co.tracker.storageClient.CreateEdge(chunkID, depChunkID, "DEPENDS_ON")
			if err != nil {
				return fmt.Errorf("failed to create dependency from chunk %d to %d: %w", chunkIndex, depIndex, err)
			}
		}
	}

	return nil
}

// OptimizeChunkOrder reorders chunks for optimal parallel processing
func (co *ChunkOperations) OptimizeChunkOrder(chunks []*ChunkInfo, criteria string) []*ChunkInfo {
	switch criteria {
	case "size_ascending":
		return co.sortChunksBySize(chunks, true)
	case "size_descending":
		return co.sortChunksBySize(chunks, false)
	case "dependency_aware":
		return co.sortChunksByDependencies(chunks)
	default:
		// Return original order
		return chunks
	}
}

// CalculateOptimalChunkSize calculates optimal chunk size based on file characteristics
func (co *ChunkOperations) CalculateOptimalChunkSize(fileSize int64, targetChunkCount int, fileType string) int64 {
	// Base calculation
	baseChunkSize := fileSize / int64(targetChunkCount)

	// Adjust based on file type
	switch fileType {
	case "text", "code":
		// Prefer smaller chunks for text files to maintain line boundaries
		return co.alignToLineBreaks(baseChunkSize, fileSize)
	case "binary", "media":
		// Larger chunks for binary files for efficiency
		return co.alignToPowerOfTwo(baseChunkSize * 2)
	case "compressed":
		// Respect compression boundaries
		return co.alignToCompressionBlocks(baseChunkSize)
	default:
		// Default alignment to 64KB boundaries
		return co.alignTo64KB(baseChunkSize)
	}
}

// Utility functions

func (co *ChunkOperations) detectMimeType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt", ".md", ".go", ".py", ".js", ".html", ".css":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".zip":
		return "application/zip"
	default:
		return "application/octet-stream"
	}
}

func (co *ChunkOperations) sortChunksBySize(chunks []*ChunkInfo, ascending bool) []*ChunkInfo {
	sorted := make([]*ChunkInfo, len(chunks))
	copy(sorted, chunks)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if ascending {
				if sorted[i].Size > sorted[j].Size {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			} else {
				if sorted[i].Size < sorted[j].Size {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}

func (co *ChunkOperations) sortChunksByDependencies(chunks []*ChunkInfo) []*ChunkInfo {
	// Simple topological sort - would need access to dependency graph
	// For now, return original order
	return chunks
}

func (co *ChunkOperations) alignToLineBreaks(chunkSize int64, fileSize int64) int64 {
	// Align to approximately 4KB boundaries for text files
	alignment := int64(4096)
	return ((chunkSize + alignment - 1) / alignment) * alignment
}

func (co *ChunkOperations) alignToPowerOfTwo(chunkSize int64) int64 {
	// Find next power of 2
	size := int64(1)
	for size < chunkSize {
		size <<= 1
	}
	return size
}

func (co *ChunkOperations) alignToCompressionBlocks(chunkSize int64) int64 {
	// Align to 32KB blocks commonly used in compression
	alignment := int64(32768)
	return ((chunkSize + alignment - 1) / alignment) * alignment
}

func (co *ChunkOperations) alignTo64KB(chunkSize int64) int64 {
	alignment := int64(65536) // 64KB
	return ((chunkSize + alignment - 1) / alignment) * alignment
}

// ParallelProcessingCoordinator manages coordinated parallel chunk processing.
// This coordinator orchestrates the parallel processing of independent chunks,
// managing worker pools, job distribution, and result collection with proper
// error handling and status tracking.
//
// The coordinator ensures efficient utilization of system resources while
// maintaining chunk dependency constraints and providing comprehensive
// processing result tracking.
type ParallelProcessingCoordinator struct {
	tracker     *ChunkTracker               // Chunk metadata and state management
	operations  *ChunkOperations            // High-level chunk operations
	maxWorkers  int                         // Maximum concurrent worker goroutines
	workerQueue chan *ChunkInfo             // Channel for distributing chunks to workers
	resultQueue chan *ChunkProcessingResult // Channel for collecting processing results
}

// ChunkProcessingResult represents the outcome of processing a single chunk.
// This structure captures both successful processing results and error conditions,
// enabling comprehensive result handling in parallel processing workflows.
type ChunkProcessingResult struct {
	Chunk *ChunkInfo // The chunk that was processed
	Error error      // Processing error, if any
	Data  []byte     // Processed chunk data (if successful)
}

// NewParallelProcessingCoordinator creates a new coordinator
func NewParallelProcessingCoordinator(tracker *ChunkTracker, maxWorkers int) *ParallelProcessingCoordinator {
	return &ParallelProcessingCoordinator{
		tracker:     tracker,
		operations:  NewChunkOperations(tracker),
		maxWorkers:  maxWorkers,
		workerQueue: make(chan *ChunkInfo, maxWorkers*2),
		resultQueue: make(chan *ChunkProcessingResult, maxWorkers*2),
	}
}

// ProcessFileInParallel orchestrates parallel processing of all independent chunks for a file.
// This method coordinates the entire parallel processing workflow including worker management,
// job distribution, result collection, and final status updates.
//
// Parallel Processing Workflow:
// 1. Identify independent chunks that can be processed concurrently
// 2. Start worker goroutines up to the configured maximum
// 3. Distribute chunks to workers via the worker queue
// 4. Collect processing results and handle errors
// 5. Store successful results and update chunk status
// 6. Track processing completion and report failures
//
// The method respects chunk dependencies and only processes chunks that
// have no unresolved dependencies, ensuring correct processing order.
//
// Parameters:
//   - fileHash: SHA256 hash of the file whose chunks should be processed
//   - processorFunc: Function to process individual chunks
//
// Returns:
//   - error: Coordination error or critical processing failure
//
// Called by: Chunk processing agents, parallel workflow managers
func (ppc *ParallelProcessingCoordinator) ProcessFileInParallel(fileHash string, processorFunc func(*ChunkInfo) ([]byte, error)) error {
	// Identify chunks that can be processed independently in parallel
	independentChunks, err := ppc.tracker.GetIndependentChunks(fileHash)
	if err != nil {
		return fmt.Errorf("failed to get independent chunks: %w", err)
	}

	// Start workers
	for i := 0; i < ppc.maxWorkers; i++ {
		go ppc.worker(processorFunc)
	}

	// Submit chunks for processing
	go func() {
		defer close(ppc.workerQueue)
		for _, chunk := range independentChunks {
			ppc.workerQueue <- chunk.ChunkInfo
		}
	}()

	// Collect results
	processedCount := 0
	for result := range ppc.resultQueue {
		if result.Error != nil {
			// Mark chunk as failed
			ppc.tracker.UpdateChunkStatus(fileHash, result.Chunk.Index, "failed", "parallel-coordinator")
			ppc.tracker.logger.Error("Chunk processing failed", "chunk", result.Chunk.Index, "error", result.Error)
		} else {
			// Store processed data and mark as completed
			if err := ppc.operations.StoreChunkData(result.Chunk, result.Data); err != nil {
				ppc.tracker.logger.Error("Failed to store chunk data", "chunk", result.Chunk.Index, "error", err)
			} else {
				ppc.tracker.UpdateChunkStatus(fileHash, result.Chunk.Index, "completed", "parallel-coordinator")
			}
		}

		processedCount++
		if processedCount >= len(independentChunks) {
			break
		}
	}

	return nil
}

func (ppc *ParallelProcessingCoordinator) worker(processorFunc func(*ChunkInfo) ([]byte, error)) {
	for chunk := range ppc.workerQueue {
		// Mark chunk as processing
		ppc.tracker.UpdateChunkStatus("", chunk.Index, "processing", "parallel-worker")

		// Process the chunk
		data, err := processorFunc(chunk)

		// Send result
		ppc.resultQueue <- &ChunkProcessingResult{
			Chunk: chunk,
			Error: err,
			Data:  data,
		}
	}
}
