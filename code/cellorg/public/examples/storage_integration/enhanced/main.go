package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
	"github.com/tenzoki/agen/cellorg/internal/storage"
)

// EnhancedFileProcessor demonstrates how to use the Godast Storage Agent
// for file processing with persistence, deduplication, and full-text search
type EnhancedFileProcessor struct {
	agent.DefaultAgentRunner
	storageClient *storage.Client
	processed     map[string]bool // Local cache for demonstration
}

// ProcessMessage handles file processing with storage integration
func (p *EnhancedFileProcessor) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("EnhancedFileProcessor received message %s", msg.ID)

	// Extract file path from payload
	var filePath string
	switch payload := msg.Payload.(type) {
	case string:
		filePath = payload
	case []byte:
		filePath = string(payload)
	default:
		base.LogError("Unsupported payload type: %T", payload)
		return nil, nil
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		base.LogError("Failed to read file %s: %v", filePath, err)
		return nil, err
	}

	base.LogInfo("Processing file: %s (size: %d bytes)", filePath, len(content))

	// Step 1: Check if file already processed using content hash
	fileHash := p.calculateFileHash(content)
	processedKey := fmt.Sprintf("processed:file:%s", fileHash)

	exists, err := p.storageClient.KVExists(processedKey)
	if err != nil {
		base.LogError("Failed to check if file processed: %v", err)
	} else if exists {
		base.LogInfo("File %s already processed (hash: %s), skipping", filePath, fileHash)
		return nil, nil
	}

	// Step 2: Store file content in file store for persistence
	fileMetadata := map[string]interface{}{
		"original_path":   filePath,
		"processed_at":    time.Now().Format(time.RFC3339),
		"processor_id":    base.ID,
		"file_size":       len(content),
		"content_type":    p.detectContentType(filePath),
		"processing_hash": fileHash,
	}

	storageHash, err := p.storageClient.StoreFile(content, fileMetadata)
	if err != nil {
		base.LogError("Failed to store file content: %v", err)
		return nil, err
	}

	base.LogInfo("File content stored with hash: %s", storageHash)

	// Step 3: Process the content (example: extract text and create summary)
	processedContent := p.processFileContent(string(content), filePath)

	// Step 4: Store processing metadata and relationships
	processingRecord := map[string]interface{}{
		"file_path":      filePath,
		"file_hash":      fileHash,
		"storage_hash":   storageHash,
		"processed_at":   time.Now().Format(time.RFC3339),
		"processor_id":   base.ID,
		"content_length": len(processedContent),
		"original_size":  len(content),
		"status":         "completed",
	}

	err = p.storageClient.KVSet(processedKey, processingRecord)
	if err != nil {
		base.LogError("Failed to store processing record: %v", err)
	}

	// Step 5: Create graph relationships for analytics
	err = p.createGraphRelationships(base, filePath, fileHash, storageHash)
	if err != nil {
		base.LogError("Failed to create graph relationships: %v", err)
	}

	// Step 6: Index content for full-text search
	err = p.indexForSearch(base, filePath, processedContent, fileMetadata)
	if err != nil {
		base.LogError("Failed to index content for search: %v", err)
	}

	// Step 7: Store enhanced statistics
	err = p.updateProcessingStats(base, filePath, len(content), len(processedContent))
	if err != nil {
		base.LogError("Failed to update processing stats: %v", err)
	}

	// Create response message with enhanced metadata
	response := &client.BrokerMessage{
		ID:      msg.ID + "_enhanced",
		Target:  base.GetEgress(),
		Type:    "enhanced_file_processed",
		Payload: processedContent,
		Meta: map[string]interface{}{
			"original_file":   filePath,
			"file_hash":       fileHash,
			"storage_hash":    storageHash,
			"processed_at":    time.Now(),
			"processor_id":    base.ID,
			"content_length":  len(processedContent),
			"original_size":   len(content),
			"deduplication":   exists,
			"storage_enabled": true,
			"searchable":      true,
		},
	}

	base.LogInfo("Enhanced processing completed for %s", filePath)
	return response, nil
}

// Init initializes the enhanced file processor with storage integration
func (p *EnhancedFileProcessor) Init(base *agent.BaseAgent) error {
	// Initialize storage client
	// In a real implementation, this would get the broker client from the base agent
	// For this example, we'll create a placeholder
	p.storageClient = &storage.Client{} // This would be properly initialized

	p.processed = make(map[string]bool)

	base.LogInfo("Enhanced File Processor initialized with storage integration")
	return nil
}

// Helper methods

func (p *EnhancedFileProcessor) calculateFileHash(content []byte) string {
	// Simple hash calculation for demonstration
	return fmt.Sprintf("hash_%d_%d", len(content), time.Now().UnixNano()%1000000)
}

func (p *EnhancedFileProcessor) detectContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}

func (p *EnhancedFileProcessor) processFileContent(content, filePath string) string {
	// Example processing: create summary and add metadata
	lines := strings.Split(content, "\n")
	wordCount := len(strings.Fields(content))

	summary := fmt.Sprintf(`=== ENHANCED PROCESSING REPORT ===
File: %s
Processed at: %s
Original lines: %d
Word count: %d
Content preview: %s

=== PROCESSED CONTENT ===
%s

=== PROCESSING METADATA ===
- Content type: %s
- Storage integration: ENABLED
- Search indexing: ENABLED
- Deduplication: ENABLED
- Graph relationships: ENABLED
`,
		filePath,
		time.Now().Format(time.RFC3339),
		len(lines),
		wordCount,
		p.getContentPreview(content, 100),
		strings.ToUpper(content), // Example transformation
		p.detectContentType(filePath),
	)

	return summary
}

func (p *EnhancedFileProcessor) getContentPreview(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

func (p *EnhancedFileProcessor) createGraphRelationships(base *agent.BaseAgent, filePath, fileHash, storageHash string) error {
	// Create vertex for file
	fileVertex := map[string]interface{}{
		"path":         filePath,
		"hash":         fileHash,
		"storage_hash": storageHash,
		"type":         "file",
		"processed_at": time.Now().Format(time.RFC3339),
	}

	_, err := p.storageClient.CreateVertex("file:"+fileHash, fileVertex)
	if err != nil {
		return fmt.Errorf("failed to create file vertex: %w", err)
	}

	// Create vertex for processor
	processorVertex := map[string]interface{}{
		"id":   base.ID,
		"type": "processor",
		"name": "enhanced_file_processor",
	}

	_, err = p.storageClient.CreateVertex("processor:"+base.ID, processorVertex)
	if err != nil {
		return fmt.Errorf("failed to create processor vertex: %w", err)
	}

	// Create processing relationship
	err = p.storageClient.CreateEdge("processor:"+base.ID, "file:"+fileHash, "processed")
	if err != nil {
		return fmt.Errorf("failed to create processing edge: %w", err)
	}

	return nil
}

func (p *EnhancedFileProcessor) indexForSearch(base *agent.BaseAgent, filePath, content string, metadata map[string]interface{}) error {
	// Create searchable document
	searchableContent := fmt.Sprintf(`
File: %s
Content: %s
Metadata: %s
`, filePath, content, p.metadataToString(metadata))

	// Index with metadata
	searchMetadata := map[string]interface{}{
		"file_path":    filePath,
		"indexed_at":   time.Now().Format(time.RFC3339),
		"processor_id": base.ID,
		"searchable":   true,
	}

	err := p.storageClient.IndexContent("file:"+filePath, searchableContent, searchMetadata)
	if err != nil {
		return fmt.Errorf("failed to index content: %w", err)
	}

	base.LogDebug("Content indexed for search: %s", filePath)
	return nil
}

func (p *EnhancedFileProcessor) updateProcessingStats(base *agent.BaseAgent, filePath string, originalSize, processedSize int) error {
	// Update global statistics
	statsKey := "stats:processing:global"

	// Get current stats
	currentStats, err := p.storageClient.KVGet(statsKey)
	if err != nil {
		// Initialize stats if not exists
		currentStats = map[string]interface{}{
			"total_files":         0,
			"total_bytes":         0,
			"total_processed":     0,
			"average_compression": 0.0,
			"last_updated":        time.Now().Format(time.RFC3339),
		}
	}

	stats := currentStats.(map[string]interface{})

	// Update statistics
	totalFiles := int(stats["total_files"].(float64)) + 1
	totalBytes := int(stats["total_bytes"].(float64)) + originalSize
	totalProcessed := int(stats["total_processed"].(float64)) + processedSize
	avgCompression := float64(totalProcessed) / float64(totalBytes)

	updatedStats := map[string]interface{}{
		"total_files":         totalFiles,
		"total_bytes":         totalBytes,
		"total_processed":     totalProcessed,
		"average_compression": avgCompression,
		"last_updated":        time.Now().Format(time.RFC3339),
		"last_file":           filePath,
	}

	err = p.storageClient.KVSet(statsKey, updatedStats)
	if err != nil {
		return fmt.Errorf("failed to update stats: %w", err)
	}

	// Update daily statistics
	today := time.Now().Format("2006-01-02")
	dailyStatsKey := fmt.Sprintf("stats:daily:%s", today)

	dailyStats, err := p.storageClient.KVGet(dailyStatsKey)
	if err != nil {
		dailyStats = map[string]interface{}{
			"date":        today,
			"files_count": 0,
			"bytes_total": 0,
		}
	}

	daily := dailyStats.(map[string]interface{})
	daily["files_count"] = int(daily["files_count"].(float64)) + 1
	daily["bytes_total"] = int(daily["bytes_total"].(float64)) + originalSize
	daily["last_updated"] = time.Now().Format(time.RFC3339)

	err = p.storageClient.KVSet(dailyStatsKey, daily)
	if err != nil {
		return fmt.Errorf("failed to update daily stats: %w", err)
	}

	base.LogDebug("Processing stats updated")
	return nil
}

func (p *EnhancedFileProcessor) metadataToString(metadata map[string]interface{}) string {
	jsonBytes, _ := json.Marshal(metadata)
	return string(jsonBytes)
}

func main() {
	agent.Run(&EnhancedFileProcessor{}, "enhanced-file-processor")
}
