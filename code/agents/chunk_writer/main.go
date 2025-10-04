// Package main provides the chunk writer agent for the GOX framework.
//
// The chunk writer is a Framework-compliant agent that saves enriched text chunks
// to various output formats and storage systems. It operates as a pure GOX agent
// using the standard Framework pattern with ProcessMessage interface.
//
// Key Features:
// - Multiple output formats (JSON, plain text, markdown, CSV)
// - Flexible storage destinations (files, directories, structured storage)
// - Metadata preservation and propagation
// - Framework compliance with ProcessMessage interface
// - Configurable naming and organization schemes
//
// Output Formats:
// - json: Structured JSON with full metadata
// - text: Plain text content only
// - markdown: Markdown-formatted with headers and metadata
// - csv: Tabular format for analysis
// - xml: Structured XML format
//
// Operation:
// The agent receives ChunkWriteRequest messages containing enriched chunks and
// output specifications, processes them through appropriate writers, and returns
// ChunkWriteResponse messages with file paths and storage metadata.
//
// Called by: GOX orchestrator via Framework message routing
// Calls: File writers, formatters, storage systems, path generators
package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// ChunkWriter implements the GOX Framework agent pattern for chunk writing.
//
// This agent provides chunk writing services through the standard Framework
// ProcessMessage interface. It embeds DefaultAgentRunner for standard agent
// lifecycle management and focuses solely on chunk writing functionality.
//
// Thread Safety: The Framework handles concurrency and message ordering
type ChunkWriter struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
	config                   *WriterConfig
	formatters               map[string]OutputFormatter
}

// WriterConfig contains configuration for the chunk writer
type WriterConfig struct {
	DefaultOutputFormat   string                 `json:"default_output_format"`   // json, text, markdown, csv, xml
	OutputDirectory       string                 `json:"output_directory"`        // Base output directory
	CreateDirectories     bool                   `json:"create_directories"`      // Auto-create directories
	PreserveMetadata      bool                   `json:"preserve_metadata"`       // Include metadata in output
	NamingScheme          string                 `json:"naming_scheme"`           // chunk_XXXX, hash, timestamp, custom
	FileExtensions        map[string]string      `json:"file_extensions"`         // format -> extension
	FormatterOptions      map[string]interface{} `json:"formatter_options"`       // format-specific options
	CompressionEnabled    bool                   `json:"compression_enabled"`     // Enable compression
	BackupEnabled         bool                   `json:"backup_enabled"`          // Enable backup of existing files
	MaxFileSize           int64                  `json:"max_file_size"`           // Max size per file
	OrganizationScheme    string                 `json:"organization_scheme"`     // flat, by_type, by_date, hierarchical
}

// OutputFormatter interface for different output formats
type OutputFormatter interface {
	Format(chunk EnrichedChunkInfo, metadata ChunkMetadata) ([]byte, error)
	Extension() string
	MimeType() string
}

// ChunkWriteRequest represents a chunk writing request
type ChunkWriteRequest struct {
	RequestID       string                 `json:"request_id"`              // Unique identifier for tracking
	EnrichedChunks  []EnrichedChunkInfo    `json:"enriched_chunks"`         // Enriched chunks to write
	OutputOptions   OutputOptions          `json:"output_options"`          // Output configuration
	NamingOptions   NamingOptions          `json:"naming_options,omitempty"` // File naming configuration
	Metadata        map[string]interface{} `json:"metadata,omitempty"`      // Additional request metadata
	ReplyTo         string                 `json:"reply_to,omitempty"`      // Optional reply destination
}

// EnrichedChunkInfo represents an enriched chunk with full context
type EnrichedChunkInfo struct {
	// Basic chunk information
	Index       int    `json:"index"`
	Text        string `json:"text"`
	Hash        string `json:"hash"`
	Size        int64  `json:"size"`
	WordCount   int    `json:"word_count"`
	StartOffset int64  `json:"start_offset"`
	EndOffset   int64  `json:"end_offset"`
	Status      string `json:"status"`

	// Enriched context
	DocumentContext   DocumentContext    `json:"document_context"`
	SemanticContext   SemanticContext    `json:"semantic_context"`
	StructuralContext StructuralContext  `json:"structural_context"`
	RelationalContext RelationalContext  `json:"relational_context"`
	ExtractedMetadata map[string]interface{} `json:"extracted_metadata,omitempty"`
}

// DocumentContext provides document-level positioning
type DocumentContext struct {
	PageRange       []int   `json:"page_range,omitempty"`
	SectionTitle    string  `json:"section_title,omitempty"`
	ParagraphIndex  int     `json:"paragraph_index"`
	LineNumbers     []int   `json:"line_numbers,omitempty"`
	PositionRatio   float64 `json:"position_ratio"`
}

// SemanticContext provides semantic classification
type SemanticContext struct {
	ContentType     string   `json:"content_type"`
	TopicKeywords   []string `json:"topic_keywords,omitempty"`
	Sentiment       string   `json:"sentiment,omitempty"`
	Language        string   `json:"language,omitempty"`
	Confidence      float64  `json:"confidence"`
	SubjectMatter   string   `json:"subject_matter,omitempty"`
}

// StructuralContext provides structural positioning
type StructuralContext struct {
	HierarchyLevel  int      `json:"hierarchy_level"`
	ParentSections  []string `json:"parent_sections,omitempty"`
	ChildElements   []string `json:"child_elements,omitempty"`
	StructureType   string   `json:"structure_type"`
	FormattingHints []string `json:"formatting_hints,omitempty"`
}

// RelationalContext provides relationships to other chunks
type RelationalContext struct {
	PrecedingChunks []int            `json:"preceding_chunks,omitempty"`
	FollowingChunks []int            `json:"following_chunks,omitempty"`
	References      []Reference      `json:"references,omitempty"`
	Continuations   []Continuation   `json:"continuations,omitempty"`
	Similarities    []Similarity     `json:"similarities,omitempty"`
}

// Reference represents a reference to another part
type Reference struct {
	Type       string `json:"type"`
	Target     string `json:"target"`
	Context    string `json:"context"`
	ChunkIndex int    `json:"chunk_index"`
}

// Continuation represents text continuations
type Continuation struct {
	Type       string  `json:"type"`
	FromChunk  int     `json:"from_chunk"`
	ToChunk    int     `json:"to_chunk"`
	Overlap    string  `json:"overlap"`
	Confidence float64 `json:"confidence"`
}

// Similarity represents chunk similarities
type Similarity struct {
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
}

// OutputOptions contains output configuration
type OutputOptions struct {
	Format              string                 `json:"format"`                       // json, text, markdown, csv, xml
	OutputDirectory     string                 `json:"output_directory"`             // Target directory
	IncludeMetadata     bool                   `json:"include_metadata"`             // Include metadata in output
	IncludeContext      bool                   `json:"include_context"`              // Include enriched context
	FilePerChunk        bool                   `json:"file_per_chunk"`               // One file per chunk vs combined
	CompressionFormat   string                 `json:"compression_format,omitempty"` // gzip, zip, none
	CustomTemplate      string                 `json:"custom_template,omitempty"`    // Custom format template
	FormatOptions       map[string]interface{} `json:"format_options,omitempty"`     // Format-specific options
}

// NamingOptions contains file naming configuration
type NamingOptions struct {
	Scheme       string            `json:"scheme"`        // chunk_XXXX, hash, timestamp, custom
	Prefix       string            `json:"prefix"`        // File prefix
	Suffix       string            `json:"suffix"`        // File suffix
	PadLength    int               `json:"pad_length"`    // Zero-padding length for numbers
	Variables    map[string]string `json:"variables"`     // Template variables
	Organization string            `json:"organization"`  // flat, by_type, by_date, hierarchical
}

// ChunkMetadata contains metadata for each chunk write operation
type ChunkMetadata struct {
	RequestID     string                 `json:"request_id"`
	ChunkIndex    int                    `json:"chunk_index"`
	OriginalFile  string                 `json:"original_file,omitempty"`
	ProcessedAt   time.Time              `json:"processed_at"`
	WriterVersion string                 `json:"writer_version"`
	CustomData    map[string]interface{} `json:"custom_data,omitempty"`
}

// ChunkWriteResponse represents the result of chunk writing operations
type ChunkWriteResponse struct {
	RequestID       string                 `json:"request_id"`                 // Original request identifier
	Success         bool                   `json:"success"`                    // Writing success status
	WrittenFiles    []WrittenFileInfo      `json:"written_files,omitempty"`    // Information about written files
	TotalChunks     int                    `json:"total_chunks"`               // Number of chunks processed
	WrittenChunks   int                    `json:"written_chunks"`             // Number of chunks successfully written
	SkippedChunks   int                    `json:"skipped_chunks"`             // Number of chunks skipped
	OutputDirectory string                 `json:"output_directory,omitempty"` // Base output directory
	ProcessingTime  time.Duration          `json:"processing_time"`            // Time taken for writing
	Error           string                 `json:"error,omitempty"`            // Error message (on failure)
	Metadata        map[string]interface{} `json:"metadata,omitempty"`         // Additional response metadata
	OriginalRequest ChunkWriteRequest      `json:"original_request,omitempty"` // Echo of original request
}

// WrittenFileInfo contains information about a written file
type WrittenFileInfo struct {
	ChunkIndex   int       `json:"chunk_index"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	Format       string    `json:"format"`
	Hash         string    `json:"hash"`
	CreatedAt    time.Time `json:"created_at"`
	RelativePath string    `json:"relative_path,omitempty"`
}

// Init initializes the chunk writer agent with formatters and configuration.
//
// This method is called once during agent startup after BaseAgent initialization.
// It loads the writer configuration, initializes output formatters,
// and prepares the agent for chunk writing operations.
//
// Parameters:
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - error: Initialization error or nil on success
//
// Called by: GOX agent framework during startup
// Calls: config loading, formatter initialization methods
func (w *ChunkWriter) Init(base *agent.BaseAgent) error {
	// Load configuration from base agent or defaults
	cfg, err := w.loadConfiguration(base)
	if err != nil {
		return fmt.Errorf("failed to load writer configuration: %w", err)
	}
	w.config = cfg

	// Initialize output formatters
	w.formatters = make(map[string]OutputFormatter)

	// Register built-in formatters
	w.formatters["json"] = &JSONFormatter{}
	w.formatters["text"] = &TextFormatter{}
	w.formatters["markdown"] = &MarkdownFormatter{}
	w.formatters["csv"] = &CSVFormatter{}
	w.formatters["xml"] = &XMLFormatter{}

	// Create output directory if configured
	if cfg.OutputDirectory != "" && cfg.CreateDirectories {
		if err := os.MkdirAll(cfg.OutputDirectory, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	agentID := base.GetConfigString("agent_id", "chunk-writer")

	base.LogInfo("Chunk Writer Agent initialized with %d formatters", len(w.formatters))
	base.LogDebug("Agent ID: %s", agentID)
	base.LogDebug("Default output format: %s", cfg.DefaultOutputFormat)
	base.LogDebug("Output directory: %s", cfg.OutputDirectory)
	base.LogDebug("Create directories: %t", cfg.CreateDirectories)
	base.LogDebug("Preserve metadata: %t", cfg.PreserveMetadata)

	// Log registered formatters for debugging
	for format := range w.formatters {
		base.LogDebug("Registered formatter: %s", format)
	}

	return nil
}

// ProcessMessage performs chunk writing operations on incoming requests.
//
// This is the core business logic for the chunk writer agent. It receives
// ChunkWriteRequest messages, processes enriched chunks through appropriate
// formatters, and returns ChunkWriteResponse messages with file information.
//
// Processing Steps:
// 1. Parse ChunkWriteRequest from message payload
// 2. Validate request and prepare output configuration
// 3. Process each chunk through selected formatter
// 4. Write formatted content to appropriate destinations
// 5. Create response message with file information
//
// Parameters:
//   - msg: BrokerMessage containing ChunkWriteRequest in payload
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Response message with writing results
//   - error: Always nil (errors are returned in response message)
//
// Called by: GOX agent framework during message processing
// Calls: Writing methods, formatter methods, response creation methods
func (w *ChunkWriter) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Parse ChunkWriteRequest from message payload
	var request ChunkWriteRequest
	var payload []byte

	// Handle different payload types
	switch p := msg.Payload.(type) {
	case []byte:
		payload = p
	case string:
		payload = []byte(p)
	default:
		return w.createErrorResponse("unknown", "Invalid payload type", base), nil
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		base.LogError("Failed to parse chunk write request: %v", err)
		return w.createErrorResponse("unknown", "Invalid request format: "+err.Error(), base), nil
	}

	base.LogDebug("Processing chunk write request %s", request.RequestID)

	// Validate request
	if len(request.EnrichedChunks) == 0 {
		return w.createErrorResponse(request.RequestID, "EnrichedChunks field is required", base), nil
	}

	startTime := time.Now()

	// Process chunk writing
	writtenFiles, stats, err := w.writeChunks(request, base)
	if err != nil {
		base.LogError("Chunk writing failed for request %s: %v", request.RequestID, err)
		return w.createErrorResponse(request.RequestID, err.Error(), base), nil
	}

	processingTime := time.Since(startTime)

	base.LogInfo("Successfully wrote chunks for request %s (files: %d, chunks: %d/%d, time: %v)",
		request.RequestID, len(writtenFiles), stats["written"], stats["total"], processingTime)

	return w.createSuccessResponse(request, writtenFiles, stats, processingTime, base), nil
}

// loadConfiguration loads writer configuration from BaseAgent or defaults
func (w *ChunkWriter) loadConfiguration(base *agent.BaseAgent) (*WriterConfig, error) {
	cfg := DefaultWriterConfig()

	// Override with BaseAgent configuration if available
	if outputFormat := base.GetConfigString("default_output_format", ""); outputFormat != "" {
		cfg.DefaultOutputFormat = outputFormat
	}

	if outputDir := base.GetConfigString("output_directory", ""); outputDir != "" {
		cfg.OutputDirectory = outputDir
	}

	if createDirs := base.GetConfigBool("create_directories", false); createDirs {
		cfg.CreateDirectories = createDirs
	}

	if preserveMetadata := base.GetConfigBool("preserve_metadata", false); preserveMetadata {
		cfg.PreserveMetadata = preserveMetadata
	}

	if namingScheme := base.GetConfigString("naming_scheme", ""); namingScheme != "" {
		cfg.NamingScheme = namingScheme
	}

	if maxFileSize := base.GetConfigInt("max_file_size", 0); maxFileSize > 0 {
		cfg.MaxFileSize = int64(maxFileSize)
	}

	base.LogDebug("Using writer configuration with BaseAgent overrides")
	return cfg, nil
}

// writeChunks processes and writes all chunks in the request
func (w *ChunkWriter) writeChunks(request ChunkWriteRequest, base *agent.BaseAgent) ([]WrittenFileInfo, map[string]int, error) {
	chunks := request.EnrichedChunks
	outputOptions := request.OutputOptions

	// Apply defaults from config
	if outputOptions.Format == "" {
		outputOptions.Format = w.config.DefaultOutputFormat
	}
	if outputOptions.OutputDirectory == "" {
		outputOptions.OutputDirectory = w.config.OutputDirectory
	}

	// Get formatter
	formatter, exists := w.formatters[outputOptions.Format]
	if !exists {
		return nil, nil, fmt.Errorf("unsupported output format: %s", outputOptions.Format)
	}

	// Prepare output directory
	outputDir, err := w.prepareOutputDirectory(outputOptions, request, base)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare output directory: %w", err)
	}

	var writtenFiles []WrittenFileInfo
	stats := map[string]int{
		"total":   len(chunks),
		"written": 0,
		"skipped": 0,
	}

	// Process each chunk
	for i, chunk := range chunks {
		if outputOptions.FilePerChunk {
			// Write each chunk to a separate file
			writtenFile, err := w.writeSingleChunk(chunk, formatter, outputDir, outputOptions, request, i, base)
			if err != nil {
				base.LogError("Failed to write chunk %d: %v", i, err)
				stats["skipped"]++
				continue
			}
			writtenFiles = append(writtenFiles, writtenFile)
			stats["written"]++
		} else {
			// Combined file writing would be handled here
			// For now, we'll treat it as file per chunk
			writtenFile, err := w.writeSingleChunk(chunk, formatter, outputDir, outputOptions, request, i, base)
			if err != nil {
				base.LogError("Failed to write chunk %d: %v", i, err)
				stats["skipped"]++
				continue
			}
			writtenFiles = append(writtenFiles, writtenFile)
			stats["written"]++
		}
	}

	base.LogDebug("Chunk writing completed: %d written, %d skipped", stats["written"], stats["skipped"])
	return writtenFiles, stats, nil
}

// writeSingleChunk writes a single chunk to a file
func (w *ChunkWriter) writeSingleChunk(chunk EnrichedChunkInfo, formatter OutputFormatter, outputDir string, options OutputOptions, request ChunkWriteRequest, index int, base *agent.BaseAgent) (WrittenFileInfo, error) {
	// Generate filename
	filename := w.generateFilename(chunk, options, request.NamingOptions, index, formatter.Extension())
	filePath := filepath.Join(outputDir, filename)

	// Create subdirectories if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return WrittenFileInfo{}, fmt.Errorf("failed to create subdirectory: %w", err)
	}

	// Prepare chunk metadata
	metadata := ChunkMetadata{
		RequestID:     request.RequestID,
		ChunkIndex:    index,
		ProcessedAt:   time.Now(),
		WriterVersion: "1.0.0",
		CustomData:    make(map[string]interface{}),
	}

	// Format chunk content
	content, err := formatter.Format(chunk, metadata)
	if err != nil {
		return WrittenFileInfo{}, fmt.Errorf("failed to format chunk: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return WrittenFileInfo{}, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return WrittenFileInfo{}, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create written file info
	writtenFile := WrittenFileInfo{
		ChunkIndex:   index,
		FilePath:     filePath,
		FileSize:     fileInfo.Size(),
		Format:       options.Format,
		Hash:         chunk.Hash,
		CreatedAt:    time.Now(),
		RelativePath: filename,
	}

	base.LogDebug("Wrote chunk %d to file: %s (size: %d bytes)", index, filename, fileInfo.Size())
	return writtenFile, nil
}

// prepareOutputDirectory prepares and validates the output directory
func (w *ChunkWriter) prepareOutputDirectory(options OutputOptions, request ChunkWriteRequest, base *agent.BaseAgent) (string, error) {
	outputDir := options.OutputDirectory

	// Apply organization scheme
	switch request.NamingOptions.Organization {
	case "by_date":
		outputDir = filepath.Join(outputDir, time.Now().Format("2006-01-02"))
	case "by_type":
		outputDir = filepath.Join(outputDir, options.Format)
	case "hierarchical":
		outputDir = filepath.Join(outputDir, request.RequestID)
	}

	// Create directory if it doesn't exist
	if w.config.CreateDirectories {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return "", err
		}
	}

	base.LogDebug("Prepared output directory: %s", outputDir)
	return outputDir, nil
}

// generateFilename generates a filename based on naming options
func (w *ChunkWriter) generateFilename(chunk EnrichedChunkInfo, options OutputOptions, naming NamingOptions, index int, extension string) string {
	var filename string

	switch naming.Scheme {
	case "hash":
		filename = chunk.Hash
	case "timestamp":
		filename = fmt.Sprintf("%d", time.Now().Unix())
	case "custom":
		// Custom template processing would go here
		filename = w.processTemplate(naming.Variables, index)
	default: // "chunk_XXXX"
		padLength := naming.PadLength
		if padLength == 0 {
			padLength = 4
		}
		filename = fmt.Sprintf("chunk_%0*d", padLength, index)
	}

	// Add prefix and suffix
	if naming.Prefix != "" {
		filename = naming.Prefix + "_" + filename
	}
	if naming.Suffix != "" {
		filename = filename + "_" + naming.Suffix
	}

	// Add extension
	if !strings.HasSuffix(filename, extension) {
		filename += extension
	}

	return filename
}

// processTemplate processes custom filename templates
func (w *ChunkWriter) processTemplate(variables map[string]string, index int) string {
	// Simple template processing - in production would use a proper template engine
	template := "chunk_{{index}}"
	if tmpl, exists := variables["template"]; exists {
		template = tmpl
	}

	// Replace variables
	result := strings.ReplaceAll(template, "{{index}}", fmt.Sprintf("%d", index))
	result = strings.ReplaceAll(result, "{{timestamp}}", fmt.Sprintf("%d", time.Now().Unix()))

	return result
}

// createSuccessResponse creates a successful writing response
func (w *ChunkWriter) createSuccessResponse(request ChunkWriteRequest, writtenFiles []WrittenFileInfo, stats map[string]int, processingTime time.Duration, base *agent.BaseAgent) *client.BrokerMessage {
	response := ChunkWriteResponse{
		RequestID:       request.RequestID,
		Success:         true,
		WrittenFiles:    writtenFiles,
		TotalChunks:     stats["total"],
		WrittenChunks:   stats["written"],
		SkippedChunks:   stats["skipped"],
		OutputDirectory: request.OutputOptions.OutputDirectory,
		ProcessingTime:  processingTime,
		OriginalRequest: request,
	}

	responseBytes, _ := json.Marshal(response)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("chunk_write_success_%s", request.RequestID),
		Target:  base.GetEgress(),
		Type:    "chunk_write_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"write_success":      true,
			"original_request":   request.RequestID,
			"total_chunks":       stats["total"],
			"written_chunks":     stats["written"],
			"skipped_chunks":     stats["skipped"],
			"files_written":      len(writtenFiles),
			"processing_time_ms": processingTime.Milliseconds(),
			"output_format":      request.OutputOptions.Format,
		},
	}
}

// createErrorResponse creates an error response message for failed writing
func (w *ChunkWriter) createErrorResponse(requestID, errorMsg string, base *agent.BaseAgent) *client.BrokerMessage {
	response := ChunkWriteResponse{
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
	responseBytes, _ := json.Marshal(response)

	base.LogError("Chunk writing error for request %s: %s", requestID, errorMsg)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("chunk_write_error_%s", requestID),
		Target:  base.GetEgress(),
		Type:    "chunk_write_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"write_success":    false,
			"original_request": requestID,
			"error_message":    errorMsg,
		},
	}
}

// Output formatter implementations

// JSONFormatter formats chunks as JSON
type JSONFormatter struct{}

func (f *JSONFormatter) Format(chunk EnrichedChunkInfo, metadata ChunkMetadata) ([]byte, error) {
	output := map[string]interface{}{
		"chunk":    chunk,
		"metadata": metadata,
	}
	return json.MarshalIndent(output, "", "  ")
}

func (f *JSONFormatter) Extension() string { return ".json" }
func (f *JSONFormatter) MimeType() string  { return "application/json" }

// TextFormatter formats chunks as plain text
type TextFormatter struct{}

func (f *TextFormatter) Format(chunk EnrichedChunkInfo, metadata ChunkMetadata) ([]byte, error) {
	content := fmt.Sprintf("Chunk %d\n", chunk.Index)
	content += fmt.Sprintf("Hash: %s\n", chunk.Hash)
	content += fmt.Sprintf("Size: %d bytes\n", chunk.Size)
	content += fmt.Sprintf("Word Count: %d\n", chunk.WordCount)
	content += fmt.Sprintf("Content Type: %s\n", chunk.SemanticContext.ContentType)
	content += fmt.Sprintf("Position: %.2f%%\n", chunk.DocumentContext.PositionRatio*100)
	content += "\n" + chunk.Text + "\n"
	return []byte(content), nil
}

func (f *TextFormatter) Extension() string { return ".txt" }
func (f *TextFormatter) MimeType() string  { return "text/plain" }

// MarkdownFormatter formats chunks as Markdown
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(chunk EnrichedChunkInfo, metadata ChunkMetadata) ([]byte, error) {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# Chunk %d\n\n", chunk.Index))
	content.WriteString("## Metadata\n\n")
	content.WriteString(fmt.Sprintf("- **Hash**: %s\n", chunk.Hash))
	content.WriteString(fmt.Sprintf("- **Size**: %d bytes\n", chunk.Size))
	content.WriteString(fmt.Sprintf("- **Word Count**: %d\n", chunk.WordCount))
	content.WriteString(fmt.Sprintf("- **Content Type**: %s\n", chunk.SemanticContext.ContentType))
	content.WriteString(fmt.Sprintf("- **Position**: %.2f%%\n", chunk.DocumentContext.PositionRatio*100))

	if chunk.SemanticContext.TopicKeywords != nil && len(chunk.SemanticContext.TopicKeywords) > 0 {
		content.WriteString(fmt.Sprintf("- **Keywords**: %s\n", strings.Join(chunk.SemanticContext.TopicKeywords, ", ")))
	}

	content.WriteString("\n## Content\n\n")
	content.WriteString(chunk.Text)
	content.WriteString("\n")

	return []byte(content.String()), nil
}

func (f *MarkdownFormatter) Extension() string { return ".md" }
func (f *MarkdownFormatter) MimeType() string  { return "text/markdown" }

// CSVFormatter formats chunks as CSV
type CSVFormatter struct{}

func (f *CSVFormatter) Format(chunk EnrichedChunkInfo, metadata ChunkMetadata) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header (would typically be done once per file)
	header := []string{"Index", "Hash", "Size", "WordCount", "ContentType", "Position", "Text"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write chunk data
	record := []string{
		fmt.Sprintf("%d", chunk.Index),
		chunk.Hash,
		fmt.Sprintf("%d", chunk.Size),
		fmt.Sprintf("%d", chunk.WordCount),
		chunk.SemanticContext.ContentType,
		fmt.Sprintf("%.4f", chunk.DocumentContext.PositionRatio),
		strings.ReplaceAll(chunk.Text, "\n", "\\n"), // Escape newlines
	}

	if err := writer.Write(record); err != nil {
		return nil, err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

func (f *CSVFormatter) Extension() string { return ".csv" }
func (f *CSVFormatter) MimeType() string  { return "text/csv" }

// XMLFormatter formats chunks as XML
type XMLFormatter struct{}

type XMLChunk struct {
	XMLName           xml.Name `xml:"chunk"`
	Index             int      `xml:"index,attr"`
	Hash              string   `xml:"hash,attr"`
	Size              int64    `xml:"size,attr"`
	WordCount         int      `xml:"wordCount,attr"`
	ContentType       string   `xml:"contentType,attr"`
	Position          float64  `xml:"position,attr"`
	Text              string   `xml:"text"`
}

func (f *XMLFormatter) Format(chunk EnrichedChunkInfo, metadata ChunkMetadata) ([]byte, error) {
	xmlChunk := XMLChunk{
		Index:       chunk.Index,
		Hash:        chunk.Hash,
		Size:        chunk.Size,
		WordCount:   chunk.WordCount,
		ContentType: chunk.SemanticContext.ContentType,
		Position:    chunk.DocumentContext.PositionRatio,
		Text:        chunk.Text,
	}

	output, err := xml.MarshalIndent(xmlChunk, "", "  ")
	if err != nil {
		return nil, err
	}

	// Add XML header
	result := []byte(xml.Header + string(output))
	return result, nil
}

func (f *XMLFormatter) Extension() string { return ".xml" }
func (f *XMLFormatter) MimeType() string  { return "application/xml" }

// DefaultWriterConfig returns default configuration
func DefaultWriterConfig() *WriterConfig {
	return &WriterConfig{
		DefaultOutputFormat: "json",
		OutputDirectory:     "/tmp/gox-chunk-writer",
		CreateDirectories:   true,
		PreserveMetadata:    true,
		NamingScheme:        "chunk_XXXX",
		FileExtensions: map[string]string{
			"json":     ".json",
			"text":     ".txt",
			"markdown": ".md",
			"csv":      ".csv",
			"xml":      ".xml",
		},
		FormatterOptions:   make(map[string]interface{}),
		CompressionEnabled: false,
		BackupEnabled:      false,
		MaxFileSize:        10485760, // 10MB
		OrganizationScheme: "flat",
	}
}

// main is the entry point for the chunk writer agent.
//
// Initializes and starts the agent using the GOX framework's agent runner.
// The framework handles all boilerplate including broker connection, message
// routing, lifecycle management, and chunk writing request processing.
//
// The agent ID "chunk-writer" is used for:
// - Agent registration with the support service
// - Logging and debugging identification
// - Message routing and correlation
//
// Called by: Operating system process execution
// Calls: agent.Run() with ChunkWriter implementation
func main() {
	// Framework handles all boilerplate including broker connection, message
	// routing, chunk writing processing, and lifecycle management
	if err := agent.Run(&ChunkWriter{}, "chunk-writer"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}