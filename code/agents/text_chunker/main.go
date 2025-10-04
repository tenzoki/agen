// Package main provides the text chunker agent for the GOX framework.
//
// The text chunker is a Framework-compliant agent that splits text into chunks using
// various strategies. It operates as a pure GOX agent using the standard Framework
// pattern with ProcessMessage interface.
//
// Key Features:
// - Multiple chunking strategies (paragraph-based, section-based, boundary-based)
// - Configurable chunk sizes with overlap support
// - Strategy-aware processing with boundary preservation
// - Framework compliance with ProcessMessage interface
// - Metadata generation for each chunk
//
// Chunking Strategies:
// - paragraph_based: Splits at paragraph boundaries with target size
// - section_based: Splits at section markers (headers, chapters)
// - boundary_based: Splits at specific boundaries (sentence, line, etc.)
// - size_based: Fixed-size chunks regardless of content boundaries
//
// Operation:
// The agent receives TextChunkingRequest messages containing text and strategy,
// processes them through appropriate chunking algorithms, and returns
// TextChunkingResponse messages with chunk information and metadata.
//
// Called by: GOX orchestrator via Framework message routing
// Calls: Chunking algorithms, hash calculators, boundary detectors
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// TextChunker implements the GOX Framework agent pattern for text chunking.
//
// This agent provides text chunking services through the standard Framework
// ProcessMessage interface. It embeds DefaultAgentRunner for standard agent
// lifecycle management and focuses solely on text chunking functionality.
//
// Thread Safety: The Framework handles concurrency and message ordering
type TextChunker struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
	strategies               map[string]ChunkingStrategy
	config                   *ChunkerConfig
}

// ChunkerConfig contains configuration for the text chunker
type ChunkerConfig struct {
	DefaultChunkSize int64                       `json:"default_chunk_size"`
	MaxChunkSize     int64                       `json:"max_chunk_size"`
	MinChunkSize     int64                       `json:"min_chunk_size"`
	DefaultOverlap   int64                       `json:"default_overlap"`
	TempDir          string                      `json:"temp_dir"`
	Strategies       map[string]ChunkingStrategy `json:"strategies"`
}

// ChunkingStrategy defines how text should be chunked
type ChunkingStrategy struct {
	Name            string `json:"name"`
	ChunkingMethod  string `json:"chunking_method"` // paragraph_based, section_based, boundary_based, size_based
	BoundaryType    string `json:"boundary_type"`   // paragraph, sentence, line, semantic, section
	PreferredSize   int64  `json:"preferred_size"`
	OverlapSize     int64  `json:"overlap_size"`
	PreserveContext bool   `json:"preserve_context"`
}

// TextChunkingRequest represents a text chunking request
type TextChunkingRequest struct {
	RequestID    string                 `json:"request_id"`              // Unique identifier for tracking
	Text         string                 `json:"text"`                    // Text to be chunked
	Strategy     string                 `json:"strategy,omitempty"`      // Chunking strategy name
	ChunkSize    int64                  `json:"chunk_size,omitempty"`    // Override default chunk size
	OverlapSize  int64                  `json:"overlap_size,omitempty"`  // Override default overlap
	Options      ChunkingOptions        `json:"options,omitempty"`       // Additional chunking options
	Metadata     map[string]interface{} `json:"metadata,omitempty"`      // Additional request metadata
	ReplyTo      string                 `json:"reply_to,omitempty"`      // Optional reply destination
}

// ChunkingOptions provides fine-grained control over chunking behavior
type ChunkingOptions struct {
	PreserveLineBreaks bool     `json:"preserve_line_breaks,omitempty"` // Preserve line breaks in chunks
	EnableHashing      bool     `json:"enable_hashing,omitempty"`       // Generate hash for each chunk
	CustomBoundaries   []string `json:"custom_boundaries,omitempty"`    // Custom boundary markers
	MinWordCount       int      `json:"min_word_count,omitempty"`       // Minimum words per chunk
	MaxWordCount       int      `json:"max_word_count,omitempty"`       // Maximum words per chunk
}

// TextChunkingResponse represents the result of a text chunking operation
type TextChunkingResponse struct {
	RequestID       string                 `json:"request_id"`                 // Original request identifier
	Success         bool                   `json:"success"`                    // Chunking success status
	Chunks          []ChunkInfo            `json:"chunks,omitempty"`           // Generated chunks (on success)
	TotalChunks     int                    `json:"total_chunks"`               // Number of chunks created
	StrategyUsed    string                 `json:"strategy_used,omitempty"`    // Strategy that was applied
	ChunkSizeAvg    int64                  `json:"chunk_size_avg"`             // Average chunk size
	ProcessingTime  time.Duration          `json:"processing_time"`            // Time taken for chunking
	Error           string                 `json:"error,omitempty"`            // Error message (on failure)
	Metadata        map[string]interface{} `json:"metadata,omitempty"`         // Additional response metadata
	OriginalRequest TextChunkingRequest    `json:"original_request,omitempty"` // Echo of original request
}

// ChunkInfo represents information about a single chunk
type ChunkInfo struct {
	Index       int    `json:"index"`        // Chunk index (0-based)
	Text        string `json:"text"`         // Chunk text content
	Hash        string `json:"hash"`         // SHA256 hash of chunk content
	Size        int64  `json:"size"`         // Size in bytes
	WordCount   int    `json:"word_count"`   // Number of words in chunk
	StartOffset int64  `json:"start_offset"` // Start position in original text
	EndOffset   int64  `json:"end_offset"`   // End position in original text
	Status      string `json:"status"`       // Chunk status (created, processed, etc.)
}

// Init initializes the text chunker agent with strategies and configuration.
//
// This method is called once during agent startup after BaseAgent initialization.
// It loads the chunking configuration, initializes chunking strategies,
// and prepares the agent for text chunking operations.
//
// Parameters:
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - error: Initialization error or nil on success
//
// Called by: GOX agent framework during startup
// Calls: config loading, strategy initialization methods
func (t *TextChunker) Init(base *agent.BaseAgent) error {
	// Load configuration from base agent or defaults
	cfg, err := t.loadConfiguration(base)
	if err != nil {
		return fmt.Errorf("failed to load chunker configuration: %w", err)
	}
	t.config = cfg

	// Initialize chunking strategies
	t.strategies = make(map[string]ChunkingStrategy)

	// Load built-in strategies
	builtinStrategies := getBuiltinStrategies()
	for name, strategy := range builtinStrategies {
		t.strategies[name] = strategy
	}

	// Load custom strategies from configuration
	for name, strategy := range cfg.Strategies {
		t.strategies[name] = strategy
	}

	agentID := base.GetConfigString("agent_id", "text-chunker")

	base.LogInfo("Text Chunker Agent initialized with %d strategies", len(t.strategies))
	base.LogDebug("Agent ID: %s", agentID)
	base.LogDebug("Default chunk size: %d bytes", cfg.DefaultChunkSize)
	base.LogDebug("Default overlap: %d bytes", cfg.DefaultOverlap)

	// Log registered strategies for debugging
	for name, strategy := range t.strategies {
		base.LogDebug("Registered strategy: %s (method: %s, boundary: %s, size: %d)",
			name, strategy.ChunkingMethod, strategy.BoundaryType, strategy.PreferredSize)
	}

	return nil
}

// ProcessMessage performs text chunking operations on incoming requests.
//
// This is the core business logic for the text chunker agent. It receives
// TextChunkingRequest messages, applies appropriate chunking strategies,
// and returns TextChunkingResponse messages with chunk information.
//
// Processing Steps:
// 1. Parse TextChunkingRequest from message payload
// 2. Validate request and select chunking strategy
// 3. Apply chunking algorithm based on strategy
// 4. Generate chunk metadata and hashes
// 5. Create response message with results
//
// Parameters:
//   - msg: BrokerMessage containing TextChunkingRequest in payload
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Response message with chunking results
//   - error: Always nil (errors are returned in response message)
//
// Called by: GOX agent framework during message processing
// Calls: Chunking methods, response creation methods
func (t *TextChunker) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Parse TextChunkingRequest from message payload
	var request TextChunkingRequest
	var payload []byte

	// Handle different payload types
	switch p := msg.Payload.(type) {
	case []byte:
		payload = p
	case string:
		payload = []byte(p)
	default:
		return t.createErrorResponse("unknown", "Invalid payload type", base), nil
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		base.LogError("Failed to parse text chunking request: %v", err)
		return t.createErrorResponse("unknown", "Invalid request format: "+err.Error(), base), nil
	}

	base.LogDebug("Processing text chunking request %s", request.RequestID)

	// Validate request
	if request.Text == "" {
		return t.createErrorResponse(request.RequestID, "Text field is required", base), nil
	}

	startTime := time.Now()

	// Select chunking strategy
	strategy, err := t.selectStrategy(request, base)
	if err != nil {
		return t.createErrorResponse(request.RequestID, err.Error(), base), nil
	}

	// Apply chunking strategy
	chunks, err := t.chunkText(request, strategy, base)
	if err != nil {
		base.LogError("Text chunking failed for request %s: %v", request.RequestID, err)
		return t.createErrorResponse(request.RequestID, err.Error(), base), nil
	}

	processingTime := time.Since(startTime)

	base.LogInfo("Successfully chunked text for request %s using %s (chunks: %d, time: %v)",
		request.RequestID, strategy.Name, len(chunks), processingTime)

	return t.createSuccessResponse(request, chunks, strategy, processingTime, base), nil
}

// loadConfiguration loads chunker configuration from BaseAgent or defaults
func (t *TextChunker) loadConfiguration(base *agent.BaseAgent) (*ChunkerConfig, error) {
	cfg := DefaultChunkerConfig()

	// Override with BaseAgent configuration if available
	if chunkSize := base.GetConfigInt("default_chunk_size", 0); chunkSize > 0 {
		cfg.DefaultChunkSize = int64(chunkSize)
	}

	if maxSize := base.GetConfigInt("max_chunk_size", 0); maxSize > 0 {
		cfg.MaxChunkSize = int64(maxSize)
	}

	if minSize := base.GetConfigInt("min_chunk_size", 0); minSize > 0 {
		cfg.MinChunkSize = int64(minSize)
	}

	if overlap := base.GetConfigInt("default_overlap", 0); overlap >= 0 {
		cfg.DefaultOverlap = int64(overlap)
	}

	if tempDir := base.GetConfigString("temp_dir", ""); tempDir != "" {
		cfg.TempDir = tempDir
	}

	base.LogDebug("Using chunker configuration with BaseAgent overrides")
	return cfg, nil
}

// selectStrategy selects the appropriate chunking strategy
func (t *TextChunker) selectStrategy(request TextChunkingRequest, base *agent.BaseAgent) (ChunkingStrategy, error) {
	// Use requested strategy if specified and valid
	if request.Strategy != "" {
		if strategy, exists := t.strategies[request.Strategy]; exists {
			// Apply size overrides from request
			strategy = t.applyRequestOverrides(strategy, request)
			return strategy, nil
		}
		return ChunkingStrategy{}, fmt.Errorf("unknown chunking strategy: %s", request.Strategy)
	}

	// Use default strategy
	defaultStrategy := t.strategies["general_text"]
	defaultStrategy = t.applyRequestOverrides(defaultStrategy, request)

	base.LogDebug("Using default strategy 'general_text' for request %s", request.RequestID)
	return defaultStrategy, nil
}

// applyRequestOverrides applies request-specific overrides to strategy
func (t *TextChunker) applyRequestOverrides(strategy ChunkingStrategy, request TextChunkingRequest) ChunkingStrategy {
	if request.ChunkSize > 0 {
		strategy.PreferredSize = request.ChunkSize
	}
	if request.OverlapSize >= 0 {
		strategy.OverlapSize = request.OverlapSize
	}
	return strategy
}

// chunkText performs the actual text chunking based on strategy
func (t *TextChunker) chunkText(request TextChunkingRequest, strategy ChunkingStrategy, base *agent.BaseAgent) ([]ChunkInfo, error) {
	text := request.Text

	// Apply strategy-specific chunking
	var textChunks []string

	switch strategy.ChunkingMethod {
	case "paragraph_based":
		textChunks = t.chunkByParagraphs(text, strategy.PreferredSize, strategy.OverlapSize)
	case "section_based":
		textChunks = t.chunkBySections(text, strategy.PreferredSize, strategy.OverlapSize)
	case "boundary_based":
		textChunks = t.chunkByBoundaries(text, strategy.BoundaryType, strategy.PreferredSize, strategy.OverlapSize)
	case "size_based":
		textChunks = t.chunkBySize(text, strategy.PreferredSize, strategy.OverlapSize)
	default:
		return nil, fmt.Errorf("unsupported chunking method: %s", strategy.ChunkingMethod)
	}

	// Convert to ChunkInfo with metadata
	chunks := make([]ChunkInfo, len(textChunks))
	for i, chunkText := range textChunks {
		var chunkHash string
		if request.Options.EnableHashing {
			chunkHash = t.calculateHash(chunkText)
		}

		startOffset, endOffset := t.findOffsets(text, chunkText, i, textChunks)

		chunks[i] = ChunkInfo{
			Index:       i,
			Text:        chunkText,
			Hash:        chunkHash,
			Size:        int64(len(chunkText)),
			WordCount:   len(strings.Fields(chunkText)),
			StartOffset: startOffset,
			EndOffset:   endOffset,
			Status:      "created",
		}
	}

	base.LogDebug("Text chunking completed: strategy=%s, total_chunks=%d, avg_size=%d",
		strategy.Name, len(chunks), t.calculateAverageChunkSize(chunks))

	return chunks, nil
}

// Chunking method implementations
func (t *TextChunker) chunkByParagraphs(text string, targetSize int64, overlap int64) []string {
	paragraphs := strings.Split(text, "\n\n")
	return t.combineToTargetSize(paragraphs, targetSize, overlap)
}

func (t *TextChunker) chunkBySections(text string, targetSize int64, overlap int64) []string {
	// Look for common section markers
	sectionMarkers := []string{"\n# ", "\n## ", "\n### ", "\nChapter ", "\nSection "}

	var sections []string
	remaining := text

	for _, marker := range sectionMarkers {
		if strings.Contains(remaining, marker) {
			parts := strings.Split(remaining, marker)
			sections = append(sections, parts[0])
			for i := 1; i < len(parts); i++ {
				sections = append(sections, marker+parts[i])
			}
			break
		}
	}

	if len(sections) == 0 {
		// Fallback to paragraph-based
		return t.chunkByParagraphs(text, targetSize, overlap)
	}

	return t.combineToTargetSize(sections, targetSize, overlap)
}

func (t *TextChunker) chunkByBoundaries(text string, boundaryType string, targetSize int64, overlap int64) []string {
	var boundaries []string

	switch boundaryType {
	case "sentence":
		boundaries = strings.Split(text, ". ")
		// Re-add periods
		for i := 0; i < len(boundaries)-1; i++ {
			boundaries[i] += "."
		}
	case "line":
		boundaries = strings.Split(text, "\n")
	case "paragraph":
		return t.chunkByParagraphs(text, targetSize, overlap)
	default:
		return t.chunkBySize(text, targetSize, overlap)
	}

	return t.combineToTargetSize(boundaries, targetSize, overlap)
}

func (t *TextChunker) chunkBySize(text string, targetSize int64, overlap int64) []string {
	var chunks []string
	textLen := int64(len(text))

	for start := int64(0); start < textLen; {
		end := start + targetSize
		if end > textLen {
			end = textLen
		}

		chunk := text[start:end]
		chunks = append(chunks, chunk)

		start = end - overlap
		if start <= 0 {
			start = end
		}
	}

	return chunks
}

func (t *TextChunker) combineToTargetSize(pieces []string, targetSize int64, overlap int64) []string {
	var chunks []string
	var currentChunk strings.Builder

	for _, piece := range pieces {
		pieceLen := int64(len(piece))
		currentLen := int64(currentChunk.Len())

		if currentLen+pieceLen > targetSize && currentLen > 0 {
			// Start new chunk
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()

			// Add overlap from previous chunk if requested
			if overlap > 0 && len(chunks) > 0 {
				prevChunk := chunks[len(chunks)-1]
				if int64(len(prevChunk)) > overlap {
					overlapText := prevChunk[len(prevChunk)-int(overlap):]
					currentChunk.WriteString(overlapText)
				}
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(" ")
		}
		currentChunk.WriteString(piece)
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// Helper methods
func (t *TextChunker) calculateHash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))[:16] // Shortened hash
}

func (t *TextChunker) findOffsets(fullText, chunk string, index int, allChunks []string) (int64, int64) {
	if index == 0 {
		return 0, int64(len(chunk))
	}

	// Calculate approximate position based on previous chunks
	start := int64(0)
	for i := 0; i < index; i++ {
		start += int64(len(allChunks[i]))
	}

	return start, start + int64(len(chunk))
}

func (t *TextChunker) calculateAverageChunkSize(chunks []ChunkInfo) int64 {
	if len(chunks) == 0 {
		return 0
	}

	total := int64(0)
	for _, chunk := range chunks {
		total += chunk.Size
	}
	return total / int64(len(chunks))
}

// createSuccessResponse creates a successful chunking response message
func (t *TextChunker) createSuccessResponse(request TextChunkingRequest, chunks []ChunkInfo, strategy ChunkingStrategy, processingTime time.Duration, base *agent.BaseAgent) *client.BrokerMessage {
	response := TextChunkingResponse{
		RequestID:       request.RequestID,
		Success:         true,
		Chunks:          chunks,
		TotalChunks:     len(chunks),
		StrategyUsed:    strategy.Name,
		ChunkSizeAvg:    t.calculateAverageChunkSize(chunks),
		ProcessingTime:  processingTime,
		OriginalRequest: request,
	}

	responseBytes, _ := json.Marshal(response)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("text_chunking_success_%s", request.RequestID),
		Target:  base.GetEgress(),
		Type:    "text_chunking_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"chunking_success":    true,
			"original_request":    request.RequestID,
			"strategy_used":       strategy.Name,
			"total_chunks":        len(chunks),
			"processing_time_ms":  processingTime.Milliseconds(),
			"avg_chunk_size":      t.calculateAverageChunkSize(chunks),
		},
	}
}

// createErrorResponse creates an error response message for failed chunking
func (t *TextChunker) createErrorResponse(requestID, errorMsg string, base *agent.BaseAgent) *client.BrokerMessage {
	response := TextChunkingResponse{
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
	responseBytes, _ := json.Marshal(response)

	base.LogError("Text chunking error for request %s: %s", requestID, errorMsg)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("text_chunking_error_%s", requestID),
		Target:  base.GetEgress(),
		Type:    "text_chunking_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"chunking_success": false,
			"original_request": requestID,
			"error_message":    errorMsg,
		},
	}
}

// getBuiltinStrategies returns the built-in chunking strategies
func getBuiltinStrategies() map[string]ChunkingStrategy {
	return map[string]ChunkingStrategy{
		"document_aware": {
			Name:            "document_aware",
			ChunkingMethod:  "paragraph_based",
			BoundaryType:    "paragraph",
			PreferredSize:   2048,
			OverlapSize:     256,
			PreserveContext: true,
		},
		"academic_paper": {
			Name:            "academic_paper",
			ChunkingMethod:  "section_based",
			BoundaryType:    "semantic",
			PreferredSize:   1024,
			OverlapSize:     128,
			PreserveContext: true,
		},
		"legal_document": {
			Name:            "legal_document",
			ChunkingMethod:  "section_based",
			BoundaryType:    "paragraph",
			PreferredSize:   1024,
			OverlapSize:     512,
			PreserveContext: true,
		},
		"technical_manual": {
			Name:            "technical_manual",
			ChunkingMethod:  "boundary_based",
			BoundaryType:    "line",
			PreferredSize:   1536,
			OverlapSize:     256,
			PreserveContext: true,
		},
		"general_text": {
			Name:            "general_text",
			ChunkingMethod:  "boundary_based",
			BoundaryType:    "sentence",
			PreferredSize:   2048,
			OverlapSize:     256,
			PreserveContext: false,
		},
	}
}

// DefaultChunkerConfig returns default configuration
func DefaultChunkerConfig() *ChunkerConfig {
	return &ChunkerConfig{
		DefaultChunkSize: 2048,
		MaxChunkSize:     10485760, // 10MB
		MinChunkSize:     256,
		DefaultOverlap:   256,
		TempDir:          "/tmp/gox-text-chunker",
		Strategies:       make(map[string]ChunkingStrategy),
	}
}

// main is the entry point for the text chunker agent.
//
// Initializes and starts the agent using the GOX framework's agent runner.
// The framework handles all boilerplate including broker connection, message
// routing, lifecycle management, and text chunking request processing.
//
// The agent ID "text-chunker" is used for:
// - Agent registration with the support service
// - Logging and debugging identification
// - Message routing and correlation
//
// Called by: Operating system process execution
// Calls: agent.Run() with TextChunker implementation
func main() {
	// Framework handles all boilerplate including broker connection, message
	// routing, text chunking processing, and lifecycle management
	if err := agent.Run(&TextChunker{}, "text-chunker"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}