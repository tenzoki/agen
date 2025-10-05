//go:build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/chunks"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/cellorg/internal/storage"
)

// ParallelChunkProcessingDemo demonstrates the complete chunk processing workflow
type ParallelChunkProcessingDemo struct {
	chunkTracker  *chunks.ChunkTracker
	chunkOps      *chunks.ChunkOperations
	coordinator   *chunks.ParallelProcessingCoordinator
	storageClient *storage.Client
	logger        *slog.Logger
}

// NewParallelChunkProcessingDemo creates a new demo instance
func NewParallelChunkProcessingDemo(brokerAddr string) (*ParallelChunkProcessingDemo, error) {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize broker client
	brokerClient := client.NewBrokerClient(brokerAddr, "parallel-demo", false)
	if err := brokerClient.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to broker: %w", err)
	}

	// Initialize storage client
	storageClient := storage.NewClient("chunk-processing-demo", brokerClient)

	// Initialize chunk tracker
	chunkTracker := chunks.NewChunkTracker(storageClient, logger)

	// Initialize chunk operations
	chunkOps := chunks.NewChunkOperations(chunkTracker)

	// Initialize parallel processing coordinator
	coordinator := chunks.NewParallelProcessingCoordinator(chunkTracker, 4) // 4 parallel workers

	return &ParallelChunkProcessingDemo{
		chunkTracker:  chunkTracker,
		chunkOps:      chunkOps,
		coordinator:   coordinator,
		storageClient: storageClient,
		logger:        logger,
	}, nil
}

// RunDemo executes the complete parallel chunk processing demonstration
func (demo *ParallelChunkProcessingDemo) RunDemo() error {
	demo.logger.Info("Starting Parallel Chunk Processing Demo")

	// Step 1: Create sample files for processing
	demo.logger.Info("Step 1: Creating sample files")
	sampleFiles, err := demo.createSampleFiles()
	if err != nil {
		return fmt.Errorf("failed to create sample files: %w", err)
	}
	defer demo.cleanupSampleFiles(sampleFiles)

	// Step 2: Split files into chunks
	demo.logger.Info("Step 2: Splitting files into chunks")
	fileSplits := make([]*chunks.FileSplitBatch, 0)

	for _, filePath := range sampleFiles {
		fileInfo, chunkList, metadata, err := demo.chunkOps.SplitFileIntoChunks(filePath, 1024, "byte_size")
		if err != nil {
			return fmt.Errorf("failed to split file %s: %w", filePath, err)
		}

		// Store chunk data in storage
		for _, chunk := range chunkList {
			data, err := demo.readChunkFromFile(filePath, chunk)
			if err != nil {
				return fmt.Errorf("failed to read chunk data: %w", err)
			}

			err = demo.chunkOps.StoreChunkData(chunk, data)
			if err != nil {
				return fmt.Errorf("failed to store chunk data: %w", err)
			}
		}

		// Prepare for batch registration
		fileSplit := &chunks.FileSplitBatch{
			FileInfo: fileInfo,
			Chunks:   chunkList,
			Metadata: metadata,
			FileVertex: map[string]interface{}{
				"id":           fmt.Sprintf("file:original:%s", fileInfo.Hash),
				"type":         "file_original",
				"path":         fileInfo.Path,
				"hash":         fileInfo.Hash,
				"size":         fileInfo.Size,
				"mime_type":    fileInfo.MimeType,
				"created_at":   fileInfo.CreatedAt.Format(time.RFC3339),
				"modified_at":  fileInfo.ModifiedAt.Format(time.RFC3339),
				"chunk_count":  len(chunkList),
				"split_method": metadata.SplitMethod,
				"chunk_size":   metadata.ChunkSize,
			},
		}

		fileSplits = append(fileSplits, fileSplit)
		demo.logger.Info("File split prepared", "file", filePath, "chunks", len(chunkList))
	}

	// Step 3: Batch register all file splits
	demo.logger.Info("Step 3: Batch registering file splits")
	err = demo.chunkTracker.BatchRegisterChunks(fileSplits)
	if err != nil {
		return fmt.Errorf("failed to batch register chunks: %w", err)
	}

	// Step 4: Demonstrate parallel processing for each file
	demo.logger.Info("Step 4: Starting parallel chunk processing")
	var wg sync.WaitGroup

	for _, fileSplit := range fileSplits {
		wg.Add(1)
		go func(split *chunks.FileSplitBatch) {
			defer wg.Done()
			demo.processFileInParallel(split.FileInfo.Hash, split.FileInfo.Path)
		}(fileSplit)
	}

	// Step 5: Monitor progress
	demo.logger.Info("Step 5: Monitoring processing progress")
	go demo.monitorProgress(fileSplits)

	// Wait for all processing to complete
	wg.Wait()

	// Step 6: Demonstrate metrics and reporting
	demo.logger.Info("Step 6: Generating final metrics")
	err = demo.generateFinalReport(fileSplits)
	if err != nil {
		return fmt.Errorf("failed to generate final report: %w", err)
	}

	// Step 7: Demonstrate chunk validation and reassembly
	demo.logger.Info("Step 7: Validating and reassembling files")
	err = demo.validateAndReassemble(fileSplits[0]) // Demonstrate with first file
	if err != nil {
		return fmt.Errorf("failed to validate and reassemble: %w", err)
	}

	demo.logger.Info("Parallel Chunk Processing Demo completed successfully!")
	return nil
}

// createSampleFiles creates sample files for the demo
func (demo *ParallelChunkProcessingDemo) createSampleFiles() ([]string, error) {
	tempDir, err := ioutil.TempDir("", "chunk_demo_")
	if err != nil {
		return nil, err
	}

	files := []string{
		filepath.Join(tempDir, "sample_text.txt"),
		filepath.Join(tempDir, "sample_data.json"),
		filepath.Join(tempDir, "sample_code.go"),
	}

	// Create sample text file
	textContent := strings.Repeat("This is a sample text file for chunk processing demonstration. ", 100)
	err = ioutil.WriteFile(files[0], []byte(textContent), 0644)
	if err != nil {
		return nil, err
	}

	// Create sample JSON file
	jsonContent := `{
	"name": "Chunk Processing Demo",
	"version": "1.0.0",
	"description": "Demonstration of parallel chunk processing with the Gox framework",
	"chunks": [` + strings.Repeat(`{"id": "chunk_%d", "size": 1024, "processed": false},`, 50) + `]
}`
	err = ioutil.WriteFile(files[1], []byte(jsonContent), 0644)
	if err != nil {
		return nil, err
	}

	// Create sample Go code file
	goContent := `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Sample Go code for chunk processing demo")

	// Simulate some processing
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond)
		fmt.Printf("Processing item %d\n", i)
	}

	fmt.Println("Processing complete!")
}
`
	goContent = strings.Repeat(goContent, 10) // Make it larger
	err = ioutil.WriteFile(files[2], []byte(goContent), 0644)
	if err != nil {
		return nil, err
	}

	demo.logger.Info("Sample files created", "count", len(files), "directory", tempDir)
	return files, nil
}

// processFileInParallel processes a file's chunks in parallel
func (demo *ParallelChunkProcessingDemo) processFileInParallel(fileHash, filePath string) {
	demo.logger.Info("Starting parallel processing", "file_hash", fileHash)

	// Simulate chunk processing with the coordinator
	processorFunc := func(chunk *chunks.ChunkInfo) ([]byte, error) {
		// Simulate processing time based on chunk size
		processingTime := time.Duration(chunk.Size/100) * time.Millisecond
		time.Sleep(processingTime)

		// Retrieve and "process" the chunk data
		data, err := demo.chunkOps.GetChunkData(chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to get chunk data: %w", err)
		}

		// Simulate processing (e.g., text transformation, compression, etc.)
		processedData := demo.simulateChunkProcessing(data, chunk.Index)

		demo.logger.Debug("Chunk processed",
			"file_hash", fileHash,
			"chunk_index", chunk.Index,
			"original_size", len(data),
			"processed_size", len(processedData),
			"processing_time", processingTime)

		return processedData, nil
	}

	// Process chunks using the coordinator
	err := demo.coordinator.ProcessFileInParallel(fileHash, processorFunc)
	if err != nil {
		demo.logger.Error("Failed to process file in parallel", "file_hash", fileHash, "error", err)
	} else {
		demo.logger.Info("Parallel processing completed", "file_hash", fileHash)
	}
}

// simulateChunkProcessing simulates processing of chunk data
func (demo *ParallelChunkProcessingDemo) simulateChunkProcessing(data []byte, chunkIndex int) []byte {
	// Simulate different types of processing based on chunk index
	switch chunkIndex % 3 {
	case 0:
		// Simulate text transformation (uppercase)
		return []byte(strings.ToUpper(string(data)))
	case 1:
		// Simulate compression (simple duplication removal)
		return demo.simulateCompression(data)
	case 2:
		// Simulate encryption (simple XOR)
		return demo.simulateEncryption(data, byte(chunkIndex))
	default:
		return data
	}
}

func (demo *ParallelChunkProcessingDemo) simulateCompression(data []byte) []byte {
	// Simple compression simulation - remove consecutive duplicates
	if len(data) == 0 {
		return data
	}

	compressed := make([]byte, 0, len(data))
	prev := data[0]
	compressed = append(compressed, prev)

	for i := 1; i < len(data); i++ {
		if data[i] != prev {
			compressed = append(compressed, data[i])
			prev = data[i]
		}
	}

	return compressed
}

func (demo *ParallelChunkProcessingDemo) simulateEncryption(data []byte, key byte) []byte {
	encrypted := make([]byte, len(data))
	for i, b := range data {
		encrypted[i] = b ^ key
	}
	return encrypted
}

// monitorProgress monitors processing progress for all files
func (demo *ParallelChunkProcessingDemo) monitorProgress(fileSplits []*chunks.FileSplitBatch) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			demo.logger.Info("=== Processing Progress Report ===")

			allComplete := true
			for _, fileSplit := range fileSplits {
				progress, err := demo.chunkTracker.GetProcessingProgress(fileSplit.FileInfo.Hash)
				if err != nil {
					demo.logger.Error("Failed to get progress", "file_hash", fileSplit.FileInfo.Hash, "error", err)
					continue
				}

				demo.logger.Info("File progress",
					"file", filepath.Base(fileSplit.FileInfo.Path),
					"progress", fmt.Sprintf("%.1f%%", progress.ProgressPercent),
					"completed", progress.CompletedChunks,
					"total", progress.TotalChunks,
					"failed", progress.FailedChunks,
					"throughput", fmt.Sprintf("%.2f chunks/sec", progress.Throughput))

				if progress.ProgressPercent < 100.0 {
					allComplete = false
				}
			}

			if allComplete {
				demo.logger.Info("All files processing completed!")
				return
			}

		case <-time.After(30 * time.Second):
			demo.logger.Info("Monitor timeout reached")
			return
		}
	}
}

// generateFinalReport generates comprehensive metrics report
func (demo *ParallelChunkProcessingDemo) generateFinalReport(fileSplits []*chunks.FileSplitBatch) error {
	demo.logger.Info("=== Final Processing Report ===")

	// Get system-wide metrics
	systemMetrics, err := demo.chunkTracker.GetSystemWideMetrics()
	if err != nil {
		return fmt.Errorf("failed to get system metrics: %w", err)
	}

	demo.logger.Info("System Metrics",
		"total_files", systemMetrics.TotalFiles,
		"total_chunks", systemMetrics.TotalChunks,
		"completed_chunks", systemMetrics.CompletedChunks,
		"failed_chunks", systemMetrics.FailedChunks,
		"overall_progress", fmt.Sprintf("%.1f%%", systemMetrics.OverallProgress),
		"failure_rate", fmt.Sprintf("%.2f%%", systemMetrics.FailureRate),
		"async_utilization", fmt.Sprintf("%.1f%%", systemMetrics.AsyncUtilization))

	// Get individual file metrics
	fileHashes := make([]string, len(fileSplits))
	for i, split := range fileSplits {
		fileHashes[i] = split.FileInfo.Hash
	}

	batchMetrics, err := demo.chunkTracker.GetBatchProcessingMetrics(fileHashes)
	if err != nil {
		return fmt.Errorf("failed to get batch metrics: %w", err)
	}

	for hash, metrics := range batchMetrics {
		// Find corresponding file path
		var filePath string
		for _, split := range fileSplits {
			if split.FileInfo.Hash == hash {
				filePath = filepath.Base(split.FileInfo.Path)
				break
			}
		}

		demo.logger.Info("File Metrics",
			"file", filePath,
			"hash", hash[:8]+"...",
			"total_chunks", metrics.TotalChunks,
			"processed_chunks", metrics.ProcessedChunks,
			"failed_chunks", metrics.FailedChunks,
			"parallel_processors", metrics.ParallelProcessors,
			"efficiency", fmt.Sprintf("%.1f%%", metrics.BatchEfficiency*100),
			"avg_processing_time", metrics.AverageProcessingTime,
			"throughput", fmt.Sprintf("%.2f chunks/sec", metrics.ThroughputPerSecond))
	}

	// Get processor performance
	processorMetrics, err := demo.chunkTracker.GetProcessorPerformance()
	if err != nil {
		return fmt.Errorf("failed to get processor metrics: %w", err)
	}

	demo.logger.Info("Processor Performance Summary")
	for processorID, metrics := range processorMetrics {
		demo.logger.Info("Processor Stats",
			"processor", processorID,
			"completed_chunks", metrics.CompletedChunks,
			"bytes_processed", metrics.TotalBytesProcessed,
			"avg_chunk_size", metrics.AverageChunkSize,
			"throughput", fmt.Sprintf("%.2f bytes/sec", metrics.ThroughputBytesPerSec))
	}

	return nil
}

// validateAndReassemble demonstrates chunk validation and file reassembly
func (demo *ParallelChunkProcessingDemo) validateAndReassemble(fileSplit *chunks.FileSplitBatch) error {
	fileHash := fileSplit.FileInfo.Hash
	originalPath := fileSplit.FileInfo.Path

	demo.logger.Info("Validating chunk integrity", "file_hash", fileHash)

	// Validate chunk integrity
	issues, err := demo.chunkOps.ValidateChunkIntegrity(fileHash)
	if err != nil {
		return fmt.Errorf("failed to validate chunks: %w", err)
	}

	if len(issues) > 0 {
		demo.logger.Warn("Chunk integrity issues found", "issues", issues)
	} else {
		demo.logger.Info("All chunks validated successfully")
	}

	// Reassemble file
	outputPath := originalPath + ".reassembled"
	demo.logger.Info("Reassembling file", "output_path", outputPath)

	err = demo.chunkOps.ReassembleChunks(fileHash, outputPath)
	if err != nil {
		return fmt.Errorf("failed to reassemble file: %w", err)
	}

	// Verify reassembled file
	originalData, err := ioutil.ReadFile(originalPath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	reassembledData, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("failed to read reassembled file: %w", err)
	}

	if len(originalData) != len(reassembledData) {
		demo.logger.Warn("File size mismatch after reassembly",
			"original_size", len(originalData),
			"reassembled_size", len(reassembledData))
	} else {
		demo.logger.Info("File reassembled successfully", "size", len(reassembledData))
	}

	// Clean up reassembled file
	os.Remove(outputPath)

	return nil
}

// readChunkFromFile reads chunk data from the original file
func (demo *ParallelChunkProcessingDemo) readChunkFromFile(filePath string, chunk *chunks.ChunkInfo) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]byte, chunk.Size)
	_, err = file.ReadAt(data, chunk.StartOffset)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// cleanupSampleFiles removes sample files after demo
func (demo *ParallelChunkProcessingDemo) cleanupSampleFiles(files []string) {
	if len(files) > 0 {
		tempDir := filepath.Dir(files[0])
		os.RemoveAll(tempDir)
		demo.logger.Info("Sample files cleaned up", "directory", tempDir)
	}
}

// main function to run the demo
func main() {
	// Check for broker address argument
	brokerAddr := "localhost:8080"
	if len(os.Args) > 1 {
		brokerAddr = os.Args[1]
	}

	fmt.Printf("Starting Parallel Chunk Processing Demo\n")
	fmt.Printf("Broker Address: %s\n", brokerAddr)
	fmt.Printf("========================================\n\n")

	// Create and run demo
	demo, err := NewParallelChunkProcessingDemo(brokerAddr)
	if err != nil {
		log.Fatalf("Failed to create demo: %v", err)
	}

	err = demo.RunDemo()
	if err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("Demo completed successfully!\n")
}
