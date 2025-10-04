// Package main provides the context enricher agent for the GOX framework.
//
// The context enricher is a Framework-compliant agent that enriches text chunks with
// contextual metadata and document structure information. It operates as a pure GOX
// agent using the standard Framework pattern with ProcessMessage interface.
//
// Key Features:
// - Document context analysis and semantic tagging
// - Position tracking within document structure
// - Metadata extraction and propagation
// - Framework compliance with ProcessMessage interface
// - Configurable enrichment strategies
//
// Enrichment Types:
// - Positional: Line numbers, paragraph indices, page ranges
// - Semantic: Section titles, content classification, topic detection
// - Structural: Hierarchical context, document organization
// - Relational: Cross-chunk references, continuation markers
//
// Operation:
// The agent receives ContextEnrichmentRequest messages containing chunks and
// document metadata, analyzes them to extract contextual information, and returns
// ContextEnrichmentResponse messages with enriched chunks.
//
// Called by: GOX orchestrator via Framework message routing
// Calls: Context analysis algorithms, metadata extractors, semantic classifiers
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// ContextEnricher implements the GOX Framework agent pattern for context enrichment.
//
// This agent provides context enrichment services through the standard Framework
// ProcessMessage interface. It embeds DefaultAgentRunner for standard agent
// lifecycle management and focuses solely on context enrichment functionality.
//
// Thread Safety: The Framework handles concurrency and message ordering
type ContextEnricher struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
	config                   *EnricherConfig
	patterns                 map[string]*regexp.Regexp
}

// EnricherConfig contains configuration for the context enricher
type EnricherConfig struct {
	EnablePositionalContext bool                   `json:"enable_positional_context"`
	EnableSemanticContext   bool                   `json:"enable_semantic_context"`
	EnableStructuralContext bool                   `json:"enable_structural_context"`
	EnableRelationalContext bool                   `json:"enable_relational_context"`
	SemanticPatterns        map[string]string      `json:"semantic_patterns"`       // pattern name -> regex
	SectionKeywords         map[string][]string    `json:"section_keywords"`        // section type -> keywords
	ContextDepth           int                    `json:"context_depth"`           // Depth of context analysis
	MetadataFields          []string               `json:"metadata_fields"`         // Fields to extract
	CustomEnrichments       map[string]interface{} `json:"custom_enrichments"`
}

// ContextEnrichmentRequest represents a context enrichment request
type ContextEnrichmentRequest struct {
	RequestID        string                 `json:"request_id"`              // Unique identifier for tracking
	Chunks           []ChunkInfo            `json:"chunks"`                  // Chunks to enrich
	DocumentMetadata DocumentMetadata       `json:"document_metadata"`       // Original document metadata
	FullText         string                 `json:"full_text,omitempty"`     // Full document text for context
	Strategy         string                 `json:"strategy,omitempty"`      // Enrichment strategy
	Options          EnrichmentOptions      `json:"options,omitempty"`       // Enrichment options
	Metadata         map[string]interface{} `json:"metadata,omitempty"`      // Additional request metadata
	ReplyTo          string                 `json:"reply_to,omitempty"`      // Optional reply destination
}

// ChunkInfo represents basic chunk information
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

// DocumentMetadata contains document information
type DocumentMetadata struct {
	SourcePath        string    `json:"source_path"`
	Format            string    `json:"format"`            // File extension
	MimeType          string    `json:"mime_type"`
	FileSize          int64     `json:"file_size"`
	ExtractionMethod  string    `json:"extraction_method"`
	ExtractionQuality string    `json:"extraction_quality"` // "good", "fair", "poor"
	WordCount         int       `json:"word_count"`
	PageCount         int       `json:"page_count,omitempty"`
	ProcessedAt       time.Time `json:"processed_at"`
}

// EnrichmentOptions provides fine-grained control over enrichment
type EnrichmentOptions struct {
	IncludePositional bool     `json:"include_positional,omitempty"` // Include positional metadata
	IncludeSemantic   bool     `json:"include_semantic,omitempty"`   // Include semantic classification
	IncludeStructural bool     `json:"include_structural,omitempty"` // Include structural context
	CustomFields      []string `json:"custom_fields,omitempty"`     // Custom metadata fields to extract
	AnalysisDepth     string   `json:"analysis_depth,omitempty"`    // "basic", "standard", "deep"
}

// ContextEnrichmentResponse represents the result of context enrichment
type ContextEnrichmentResponse struct {
	RequestID        string                 `json:"request_id"`                 // Original request identifier
	Success          bool                   `json:"success"`                    // Enrichment success status
	EnrichedChunks   []EnrichedChunkInfo    `json:"enriched_chunks,omitempty"`  // Enriched chunks (on success)
	TotalChunks      int                    `json:"total_chunks"`               // Number of chunks processed
	EnrichmentStats  EnrichmentStats        `json:"enrichment_stats,omitempty"` // Statistics about enrichment
	ProcessingTime   time.Duration          `json:"processing_time"`            // Time taken for enrichment
	Error            string                 `json:"error,omitempty"`            // Error message (on failure)
	Metadata         map[string]interface{} `json:"metadata,omitempty"`         // Additional response metadata
	OriginalRequest  ContextEnrichmentRequest `json:"original_request,omitempty"` // Echo of original request
}

// EnrichedChunkInfo extends ChunkInfo with contextual metadata
type EnrichedChunkInfo struct {
	ChunkInfo
	DocumentContext   DocumentContext    `json:"document_context"`           // Document-level context
	SemanticContext   SemanticContext    `json:"semantic_context"`           // Semantic classification
	StructuralContext StructuralContext  `json:"structural_context"`         // Structural position
	RelationalContext RelationalContext  `json:"relational_context"`         // Relationships to other chunks
	ExtractedMetadata map[string]interface{} `json:"extracted_metadata,omitempty"` // Additional extracted metadata
}

// DocumentContext provides document-level positioning
type DocumentContext struct {
	PageRange       []int  `json:"page_range,omitempty"`    // Pages covered by chunk
	SectionTitle    string `json:"section_title,omitempty"` // Parent section title
	ParagraphIndex  int    `json:"paragraph_index"`         // Paragraph number in document
	LineNumbers     []int  `json:"line_numbers,omitempty"`  // Line numbers covered
	PositionRatio   float64 `json:"position_ratio"`         // Relative position in document (0.0-1.0)
}

// SemanticContext provides semantic classification
type SemanticContext struct {
	ContentType     string   `json:"content_type"`              // abstract, introduction, methodology, etc.
	TopicKeywords   []string `json:"topic_keywords,omitempty"`  // Extracted topic keywords
	Sentiment       string   `json:"sentiment,omitempty"`       // positive, negative, neutral
	Language        string   `json:"language,omitempty"`        // Detected language
	Confidence      float64  `json:"confidence"`                // Classification confidence (0.0-1.0)
	SubjectMatter   string   `json:"subject_matter,omitempty"`  // Domain classification
}

// StructuralContext provides structural positioning
type StructuralContext struct {
	HierarchyLevel  int      `json:"hierarchy_level"`        // Nesting level in document structure
	ParentSections  []string `json:"parent_sections,omitempty"` // Chain of parent sections
	ChildElements   []string `json:"child_elements,omitempty"`  // Child elements (lists, figures, etc.)
	StructureType   string   `json:"structure_type"`         // heading, paragraph, list, table, etc.
	FormattingHints []string `json:"formatting_hints,omitempty"` // bold, italic, code, etc.
}

// RelationalContext provides relationships to other chunks
type RelationalContext struct {
	PrecedingChunks []int               `json:"preceding_chunks,omitempty"` // Related previous chunks
	FollowingChunks []int               `json:"following_chunks,omitempty"` // Related following chunks
	References      []Reference         `json:"references,omitempty"`       // References to other parts
	Continuations   []ContinuationInfo  `json:"continuations,omitempty"`    // Sentence/thought continuations
	Similarity      []SimilarityInfo    `json:"similarity,omitempty"`       // Similar chunks
}

// Reference represents a reference to another part of the document
type Reference struct {
	Type        string `json:"type"`         // "section", "figure", "table", "page"
	Target      string `json:"target"`       // Referenced element identifier
	Context     string `json:"context"`      // Context of the reference
	ChunkIndex  int    `json:"chunk_index"`  // Index of chunk containing the reference
}

// ContinuationInfo represents text continuations between chunks
type ContinuationInfo struct {
	Type         string `json:"type"`          // "sentence", "paragraph", "thought"
	FromChunk    int    `json:"from_chunk"`    // Source chunk index
	ToChunk      int    `json:"to_chunk"`      // Target chunk index
	Overlap      string `json:"overlap"`       // Overlapping text
	Confidence   float64 `json:"confidence"`   // Continuation confidence
}

// SimilarityInfo represents similarity between chunks
type SimilarityInfo struct {
	ChunkIndex int     `json:"chunk_index"` // Index of similar chunk
	Score      float64 `json:"score"`       // Similarity score (0.0-1.0)
	Reason     string  `json:"reason"`      // Reason for similarity
}

// EnrichmentStats contains statistics about the enrichment process
type EnrichmentStats struct {
	TotalChunks        int                    `json:"total_chunks"`
	EnrichedChunks     int                    `json:"enriched_chunks"`
	SkippedChunks      int                    `json:"skipped_chunks"`
	ContextTypesAdded  []string               `json:"context_types_added"`
	AverageConfidence  float64                `json:"average_confidence"`
	ProcessingDetails  map[string]interface{} `json:"processing_details,omitempty"`
}

// Init initializes the context enricher agent with patterns and configuration.
//
// This method is called once during agent startup after BaseAgent initialization.
// It loads the enrichment configuration, initializes analysis patterns,
// and prepares the agent for context enrichment operations.
//
// Parameters:
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - error: Initialization error or nil on success
//
// Called by: GOX agent framework during startup
// Calls: config loading, pattern compilation methods
func (c *ContextEnricher) Init(base *agent.BaseAgent) error {
	// Load configuration from base agent or defaults
	cfg, err := c.loadConfiguration(base)
	if err != nil {
		return fmt.Errorf("failed to load enricher configuration: %w", err)
	}
	c.config = cfg

	// Compile semantic patterns
	c.patterns = make(map[string]*regexp.Regexp)
	for name, pattern := range cfg.SemanticPatterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			base.LogError("Failed to compile pattern %s: %v", name, err)
			continue
		}
		c.patterns[name] = compiled
	}

	// Add built-in patterns
	builtinPatterns := getBuiltinPatterns()
	for name, pattern := range builtinPatterns {
		if _, exists := c.patterns[name]; !exists {
			compiled, err := regexp.Compile(pattern)
			if err != nil {
				base.LogError("Failed to compile built-in pattern %s: %v", name, err)
				continue
			}
			c.patterns[name] = compiled
		}
	}

	agentID := base.GetConfigString("agent_id", "context-enricher")

	base.LogInfo("Context Enricher Agent initialized with %d patterns", len(c.patterns))
	base.LogDebug("Agent ID: %s", agentID)
	base.LogDebug("Positional context: %t", cfg.EnablePositionalContext)
	base.LogDebug("Semantic context: %t", cfg.EnableSemanticContext)
	base.LogDebug("Structural context: %t", cfg.EnableStructuralContext)
	base.LogDebug("Relational context: %t", cfg.EnableRelationalContext)

	// Log registered patterns for debugging
	for name := range c.patterns {
		base.LogDebug("Registered pattern: %s", name)
	}

	return nil
}

// ProcessMessage performs context enrichment on incoming requests.
//
// This is the core business logic for the context enricher agent. It receives
// ContextEnrichmentRequest messages, analyzes chunks and document context,
// and returns ContextEnrichmentResponse messages with enriched chunks.
//
// Processing Steps:
// 1. Parse ContextEnrichmentRequest from message payload
// 2. Validate request and prepare analysis context
// 3. Enrich each chunk with appropriate context types
// 4. Generate relational context between chunks
// 5. Create response message with enriched results
//
// Parameters:
//   - msg: BrokerMessage containing ContextEnrichmentRequest in payload
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Response message with enrichment results
//   - error: Always nil (errors are returned in response message)
//
// Called by: GOX agent framework during message processing
// Calls: Enrichment methods, analysis algorithms, response creation methods
func (c *ContextEnricher) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Parse ContextEnrichmentRequest from message payload
	var request ContextEnrichmentRequest
	var payload []byte

	// Handle different payload types
	switch p := msg.Payload.(type) {
	case []byte:
		payload = p
	case string:
		payload = []byte(p)
	default:
		return c.createErrorResponse("unknown", "Invalid payload type", base), nil
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		base.LogError("Failed to parse context enrichment request: %v", err)
		return c.createErrorResponse("unknown", "Invalid request format: "+err.Error(), base), nil
	}

	base.LogDebug("Processing context enrichment request %s", request.RequestID)

	// Validate request
	if len(request.Chunks) == 0 {
		return c.createErrorResponse(request.RequestID, "Chunks field is required", base), nil
	}

	startTime := time.Now()

	// Perform context enrichment
	enrichedChunks, stats, err := c.enrichChunks(request, base)
	if err != nil {
		base.LogError("Context enrichment failed for request %s: %v", request.RequestID, err)
		return c.createErrorResponse(request.RequestID, err.Error(), base), nil
	}

	processingTime := time.Since(startTime)

	base.LogInfo("Successfully enriched context for request %s (chunks: %d/%d, time: %v)",
		request.RequestID, stats.EnrichedChunks, stats.TotalChunks, processingTime)

	return c.createSuccessResponse(request, enrichedChunks, stats, processingTime, base), nil
}

// loadConfiguration loads enricher configuration from BaseAgent or defaults
func (c *ContextEnricher) loadConfiguration(base *agent.BaseAgent) (*EnricherConfig, error) {
	cfg := DefaultEnricherConfig()

	// Override with BaseAgent configuration if available
	if enablePositional := base.GetConfigBool("enable_positional_context", false); enablePositional {
		cfg.EnablePositionalContext = enablePositional
	}

	if enableSemantic := base.GetConfigBool("enable_semantic_context", false); enableSemantic {
		cfg.EnableSemanticContext = enableSemantic
	}

	if enableStructural := base.GetConfigBool("enable_structural_context", false); enableStructural {
		cfg.EnableStructuralContext = enableStructural
	}

	if enableRelational := base.GetConfigBool("enable_relational_context", false); enableRelational {
		cfg.EnableRelationalContext = enableRelational
	}

	if contextDepth := base.GetConfigInt("context_depth", 0); contextDepth > 0 {
		cfg.ContextDepth = contextDepth
	}

	base.LogDebug("Using enricher configuration with BaseAgent overrides")
	return cfg, nil
}

// enrichChunks performs context enrichment on all chunks
func (c *ContextEnricher) enrichChunks(request ContextEnrichmentRequest, base *agent.BaseAgent) ([]EnrichedChunkInfo, EnrichmentStats, error) {
	chunks := request.Chunks
	enrichedChunks := make([]EnrichedChunkInfo, len(chunks))

	stats := EnrichmentStats{
		TotalChunks:       len(chunks),
		ContextTypesAdded: []string{},
		ProcessingDetails: make(map[string]interface{}),
	}

	totalConfidence := 0.0
	contextTypesSet := make(map[string]bool)

	// Enrich each chunk
	for i, chunk := range chunks {
		enriched, err := c.enrichSingleChunk(chunk, request, i, base)
		if err != nil {
			base.LogError("Failed to enrich chunk %d: %v", i, err)
			stats.SkippedChunks++
			// Create basic enriched chunk on error
			enriched = EnrichedChunkInfo{
				ChunkInfo: chunk,
				DocumentContext: DocumentContext{
					ParagraphIndex: i,
					PositionRatio: float64(i) / float64(len(chunks)),
				},
				SemanticContext: SemanticContext{
					ContentType: "unknown",
					Confidence: 0.0,
				},
			}
		} else {
			stats.EnrichedChunks++
			totalConfidence += enriched.SemanticContext.Confidence
		}

		enrichedChunks[i] = enriched

		// Track context types added
		if c.config.EnablePositionalContext {
			contextTypesSet["positional"] = true
		}
		if c.config.EnableSemanticContext {
			contextTypesSet["semantic"] = true
		}
		if c.config.EnableStructuralContext {
			contextTypesSet["structural"] = true
		}
	}

	// Add relational context if enabled
	if c.config.EnableRelationalContext {
		c.addRelationalContext(enrichedChunks, request, base)
		contextTypesSet["relational"] = true
	}

	// Finalize stats
	for contextType := range contextTypesSet {
		stats.ContextTypesAdded = append(stats.ContextTypesAdded, contextType)
	}

	if stats.EnrichedChunks > 0 {
		stats.AverageConfidence = totalConfidence / float64(stats.EnrichedChunks)
	}

	base.LogDebug("Context enrichment completed: %d/%d chunks enriched, avg confidence: %.2f",
		stats.EnrichedChunks, stats.TotalChunks, stats.AverageConfidence)

	return enrichedChunks, stats, nil
}

// enrichSingleChunk enriches a single chunk with context
func (c *ContextEnricher) enrichSingleChunk(chunk ChunkInfo, request ContextEnrichmentRequest, index int, base *agent.BaseAgent) (EnrichedChunkInfo, error) {
	enriched := EnrichedChunkInfo{
		ChunkInfo: chunk,
		ExtractedMetadata: make(map[string]interface{}),
	}

	// Add positional context
	if c.config.EnablePositionalContext {
		enriched.DocumentContext = c.buildDocumentContext(chunk, request, index)
	}

	// Add semantic context
	if c.config.EnableSemanticContext {
		enriched.SemanticContext = c.buildSemanticContext(chunk, request)
	}

	// Add structural context
	if c.config.EnableStructuralContext {
		enriched.StructuralContext = c.buildStructuralContext(chunk, request, index)
	}

	// Extract additional metadata
	enriched.ExtractedMetadata = c.extractMetadata(chunk, request)

	return enriched, nil
}

// buildDocumentContext builds document-level positional context
func (c *ContextEnricher) buildDocumentContext(chunk ChunkInfo, request ContextEnrichmentRequest, index int) DocumentContext {
	context := DocumentContext{
		ParagraphIndex: index,
		PositionRatio: float64(index) / float64(len(request.Chunks)),
	}

	// Extract line numbers if full text is available
	if request.FullText != "" {
		context.LineNumbers = c.findLineNumbers(chunk.Text, request.FullText)
	}

	// Try to identify section title
	context.SectionTitle = c.findSectionTitle(chunk.Text, request.FullText, index)

	return context
}

// buildSemanticContext builds semantic classification context
func (c *ContextEnricher) buildSemanticContext(chunk ChunkInfo, request ContextEnrichmentRequest) SemanticContext {
	text := strings.ToLower(chunk.Text)

	context := SemanticContext{
		TopicKeywords: c.extractTopicKeywords(chunk.Text),
		Language: "en", // Default, could be detected
		Confidence: 0.5, // Default confidence
	}

	// Classify content type
	context.ContentType = c.classifyContentType(text)

	// Adjust confidence based on classification certainty
	if context.ContentType != "unknown" {
		context.Confidence = 0.8
	}

	// Determine subject matter
	context.SubjectMatter = c.classifySubjectMatter(text)

	return context
}

// buildStructuralContext builds structural positioning context
func (c *ContextEnricher) buildStructuralContext(chunk ChunkInfo, request ContextEnrichmentRequest, index int) StructuralContext {
	text := chunk.Text

	context := StructuralContext{
		HierarchyLevel: 0,
		StructureType: "paragraph", // Default
	}

	// Detect structure type
	context.StructureType = c.detectStructureType(text)

	// Find parent sections
	context.ParentSections = c.findParentSections(text, request.FullText, index)

	// Detect formatting hints
	context.FormattingHints = c.detectFormattingHints(text)

	return context
}

// addRelationalContext adds relationships between chunks
func (c *ContextEnricher) addRelationalContext(chunks []EnrichedChunkInfo, request ContextEnrichmentRequest, base *agent.BaseAgent) {
	for i := range chunks {
		chunks[i].RelationalContext = c.buildRelationalContext(i, chunks, request)
	}
}

// buildRelationalContext builds relationships for a specific chunk
func (c *ContextEnricher) buildRelationalContext(index int, chunks []EnrichedChunkInfo, request ContextEnrichmentRequest) RelationalContext {
	context := RelationalContext{}

	// Find continuations
	context.Continuations = c.findContinuations(index, chunks)

	// Find references
	context.References = c.findReferences(chunks[index].Text, index)

	// Find similar chunks
	context.Similarity = c.findSimilarChunks(index, chunks)

	return context
}

// Helper methods for content analysis
func (c *ContextEnricher) findLineNumbers(chunkText, fullText string) []int {
	// Simple implementation - in production would be more sophisticated
	lines := strings.Split(fullText, "\n")
	chunkLines := strings.Split(chunkText, "\n")

	var lineNumbers []int
	for i, line := range lines {
		for _, chunkLine := range chunkLines {
			if strings.TrimSpace(line) == strings.TrimSpace(chunkLine) {
				lineNumbers = append(lineNumbers, i+1)
				break
			}
		}
	}

	return lineNumbers
}

func (c *ContextEnricher) findSectionTitle(chunkText, fullText string, index int) string {
	// Look for headers before this chunk
	if fullText == "" {
		return ""
	}

	// Find position of chunk in full text
	chunkPos := strings.Index(fullText, chunkText)
	if chunkPos == -1 {
		return ""
	}

	// Look backward for section headers
	beforeText := fullText[:chunkPos]
	lines := strings.Split(beforeText, "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "Chapter") || strings.HasPrefix(line, "Section") {
			return line
		}
	}

	return ""
}

func (c *ContextEnricher) extractTopicKeywords(text string) []string {
	// Simple keyword extraction - in production would use NLP
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	for _, word := range words {
		if len(word) > 4 && !c.isCommonWord(word) {
			wordCount[word]++
		}
	}

	var keywords []string
	for word, count := range wordCount {
		if count > 1 {
			keywords = append(keywords, word)
		}
	}

	if len(keywords) > 5 {
		keywords = keywords[:5]
	}

	return keywords
}

func (c *ContextEnricher) classifyContentType(text string) string {
	// Pattern-based classification
	for name, pattern := range c.patterns {
		if pattern.MatchString(text) {
			return name
		}
	}

	// Keyword-based classification
	if strings.Contains(text, "abstract") {
		return "abstract"
	}
	if strings.Contains(text, "introduction") {
		return "introduction"
	}
	if strings.Contains(text, "methodology") || strings.Contains(text, "method") {
		return "methodology"
	}
	if strings.Contains(text, "results") {
		return "results"
	}
	if strings.Contains(text, "conclusion") {
		return "conclusion"
	}
	if strings.Contains(text, "references") || strings.Contains(text, "bibliography") {
		return "references"
	}

	return "content"
}

func (c *ContextEnricher) classifySubjectMatter(text string) string {
	// Simple subject matter classification
	if strings.Contains(text, "research") || strings.Contains(text, "study") {
		return "academic"
	}
	if strings.Contains(text, "contract") || strings.Contains(text, "legal") {
		return "legal"
	}
	if strings.Contains(text, "technical") || strings.Contains(text, "system") {
		return "technical"
	}

	return "general"
}

func (c *ContextEnricher) detectStructureType(text string) string {
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "#") {
		return "heading"
	}
	if strings.HasPrefix(text, "-") || strings.HasPrefix(text, "*") || strings.HasPrefix(text, "1.") {
		return "list"
	}
	if strings.Contains(text, "|") && strings.Count(text, "|") > 2 {
		return "table"
	}
	if strings.Contains(text, "```") || strings.Contains(text, "    ") {
		return "code"
	}

	return "paragraph"
}

func (c *ContextEnricher) findParentSections(chunkText, fullText string, index int) []string {
	// Simple implementation - would be more sophisticated in production
	return []string{}
}

func (c *ContextEnricher) detectFormattingHints(text string) []string {
	var hints []string

	if strings.Contains(text, "**") || strings.Contains(text, "__") {
		hints = append(hints, "bold")
	}
	if strings.Contains(text, "*") || strings.Contains(text, "_") {
		hints = append(hints, "italic")
	}
	if strings.Contains(text, "`") {
		hints = append(hints, "code")
	}

	return hints
}

func (c *ContextEnricher) findContinuations(index int, chunks []EnrichedChunkInfo) []ContinuationInfo {
	// Simple sentence continuation detection
	var continuations []ContinuationInfo

	if index > 0 {
		currentText := chunks[index].Text

		// Check if current chunk starts mid-sentence
		if !strings.HasPrefix(strings.TrimSpace(currentText), strings.ToUpper(string(currentText[0]))) {
			continuations = append(continuations, ContinuationInfo{
				Type:       "sentence",
				FromChunk:  index - 1,
				ToChunk:    index,
				Confidence: 0.7,
			})
		}
	}

	return continuations
}

func (c *ContextEnricher) findReferences(text string, index int) []Reference {
	// Simple reference detection
	var references []Reference

	// Look for common reference patterns
	patterns := map[string]string{
		"section": `(?i)section\s+(\d+)`,
		"figure":  `(?i)figure\s+(\d+)`,
		"table":   `(?i)table\s+(\d+)`,
		"page":    `(?i)page\s+(\d+)`,
	}

	for refType, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)

		for _, match := range matches {
			if len(match) > 1 {
				references = append(references, Reference{
					Type:       refType,
					Target:     match[1],
					Context:    match[0],
					ChunkIndex: index,
				})
			}
		}
	}

	return references
}

func (c *ContextEnricher) findSimilarChunks(index int, chunks []EnrichedChunkInfo) []SimilarityInfo {
	// Simple similarity based on common keywords
	var similarities []SimilarityInfo

	currentKeywords := chunks[index].SemanticContext.TopicKeywords

	for i, chunk := range chunks {
		if i == index {
			continue
		}

		otherKeywords := chunk.SemanticContext.TopicKeywords
		commonCount := c.countCommonKeywords(currentKeywords, otherKeywords)

		if commonCount > 0 {
			totalKeywords := len(currentKeywords) + len(otherKeywords)
			if totalKeywords > 0 {
				score := float64(commonCount*2) / float64(totalKeywords)
				if score > 0.3 {
					similarities = append(similarities, SimilarityInfo{
						ChunkIndex: i,
						Score:      score,
						Reason:     "shared keywords",
					})
				}
			}
		}
	}

	return similarities
}

func (c *ContextEnricher) extractMetadata(chunk ChunkInfo, request ContextEnrichmentRequest) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Extract basic statistics
	metadata["word_density"] = float64(chunk.WordCount) / float64(chunk.Size)
	metadata["avg_word_length"] = c.calculateAverageWordLength(chunk.Text)
	metadata["sentence_count"] = len(strings.Split(chunk.Text, "."))

	return metadata
}

// Utility methods
func (c *ContextEnricher) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "that": true, "have": true, "for": true,
		"not": true, "with": true, "you": true, "this": true, "but": true,
	}
	return commonWords[word]
}

func (c *ContextEnricher) countCommonKeywords(keywords1, keywords2 []string) int {
	common := 0
	for _, k1 := range keywords1 {
		for _, k2 := range keywords2 {
			if k1 == k2 {
				common++
				break
			}
		}
	}
	return common
}

func (c *ContextEnricher) calculateAverageWordLength(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0.0
	}

	totalLength := 0
	for _, word := range words {
		totalLength += len(word)
	}

	return float64(totalLength) / float64(len(words))
}

// createSuccessResponse creates a successful enrichment response
func (c *ContextEnricher) createSuccessResponse(request ContextEnrichmentRequest, chunks []EnrichedChunkInfo, stats EnrichmentStats, processingTime time.Duration, base *agent.BaseAgent) *client.BrokerMessage {
	response := ContextEnrichmentResponse{
		RequestID:       request.RequestID,
		Success:         true,
		EnrichedChunks:  chunks,
		TotalChunks:     len(chunks),
		EnrichmentStats: stats,
		ProcessingTime:  processingTime,
		OriginalRequest: request,
	}

	responseBytes, _ := json.Marshal(response)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("context_enrichment_success_%s", request.RequestID),
		Target:  base.GetEgress(),
		Type:    "context_enrichment_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"enrichment_success":   true,
			"original_request":     request.RequestID,
			"total_chunks":         len(chunks),
			"enriched_chunks":      stats.EnrichedChunks,
			"processing_time_ms":   processingTime.Milliseconds(),
			"average_confidence":   stats.AverageConfidence,
			"context_types_added":  stats.ContextTypesAdded,
		},
	}
}

// createErrorResponse creates an error response message for failed enrichment
func (c *ContextEnricher) createErrorResponse(requestID, errorMsg string, base *agent.BaseAgent) *client.BrokerMessage {
	response := ContextEnrichmentResponse{
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
	responseBytes, _ := json.Marshal(response)

	base.LogError("Context enrichment error for request %s: %s", requestID, errorMsg)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("context_enrichment_error_%s", requestID),
		Target:  base.GetEgress(),
		Type:    "context_enrichment_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"enrichment_success": false,
			"original_request":   requestID,
			"error_message":      errorMsg,
		},
	}
}

// getBuiltinPatterns returns built-in semantic patterns
func getBuiltinPatterns() map[string]string {
	return map[string]string{
		"abstract":     `(?i)abstract\s*[:\-]`,
		"introduction": `(?i)introduction\s*[:\-]`,
		"methodology":  `(?i)(methodology|methods?)\s*[:\-]`,
		"results":      `(?i)results?\s*[:\-]`,
		"conclusion":   `(?i)conclusion\s*[:\-]`,
		"references":   `(?i)(references?|bibliography)\s*[:\-]?`,
	}
}

// DefaultEnricherConfig returns default configuration
func DefaultEnricherConfig() *EnricherConfig {
	return &EnricherConfig{
		EnablePositionalContext: true,
		EnableSemanticContext:   true,
		EnableStructuralContext: true,
		EnableRelationalContext: false, // More expensive, disabled by default
		SemanticPatterns:        make(map[string]string),
		SectionKeywords: map[string][]string{
			"academic": {"abstract", "introduction", "methodology", "results", "conclusion"},
			"legal":    {"whereas", "hereby", "clause", "party", "agreement"},
			"technical": {"procedure", "step", "installation", "configuration"},
		},
		ContextDepth:      3,
		MetadataFields:    []string{"word_density", "avg_word_length", "sentence_count"},
		CustomEnrichments: make(map[string]interface{}),
	}
}

// main is the entry point for the context enricher agent.
//
// Initializes and starts the agent using the GOX framework's agent runner.
// The framework handles all boilerplate including broker connection, message
// routing, lifecycle management, and context enrichment request processing.
//
// The agent ID "context-enricher" is used for:
// - Agent registration with the support service
// - Logging and debugging identification
// - Message routing and correlation
//
// Called by: Operating system process execution
// Calls: agent.Run() with ContextEnricher implementation
func main() {
	// Framework handles all boilerplate including broker connection, message
	// routing, context enrichment processing, and lifecycle management
	if err := agent.Run(&ContextEnricher{}, "context-enricher"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}