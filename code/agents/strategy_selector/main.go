// Package main provides the strategy selector agent for the GOX framework.
//
// The strategy selector is a Framework-compliant agent that analyzes documents
// and extraction results to determine the optimal chunking strategy. It operates
// as a pure GOX agent using the standard Framework pattern with ProcessMessage interface.
//
// Key Features:
// - Document format analysis and strategy mapping
// - Content-based strategy selection (academic, legal, technical, general)
// - Extraction quality assessment for strategy optimization
// - Framework compliance with ProcessMessage interface
// - Configurable strategy rules and preferences
//
// Strategy Selection Logic:
// - Format-based: PDF/DOCX → document_aware, TXT → general_text
// - Content-based: Keywords analysis for academic, legal, technical content
// - Quality-based: High quality extraction → more sophisticated strategies
// - Size-based: Large documents → section-based, small → paragraph-based
//
// Operation:
// The agent receives StrategySelectionRequest messages containing document metadata
// and extraction results, analyzes them to determine optimal strategy, and returns
// StrategySelectionResponse messages with strategy recommendations.
//
// Called by: GOX orchestrator via Framework message routing
// Calls: Strategy analysis algorithms, content classifiers, format detectors
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// StrategySelector implements the GOX Framework agent pattern for strategy selection.
//
// This agent provides strategy selection services through the standard Framework
// ProcessMessage interface. It embeds DefaultAgentRunner for standard agent
// lifecycle management and focuses solely on strategy selection functionality.
//
// Thread Safety: The Framework handles concurrency and message ordering
type StrategySelector struct {
	agent.DefaultAgentRunner // Embed default implementations for Init/Cleanup
	config                   *SelectorConfig
	strategies               map[string]ProcessingStrategy
	rules                    []SelectionRule
}

// SelectorConfig contains configuration for the strategy selector
type SelectorConfig struct {
	DefaultStrategy      string                       `json:"default_strategy"`
	FormatMappings       map[string]string            `json:"format_mappings"`       // file extension -> strategy
	QualityThresholds    map[string]float64           `json:"quality_thresholds"`    // strategy -> min quality
	WordCountThresholds  map[string]int               `json:"word_count_thresholds"` // strategy -> min word count
	ContentKeywords      map[string][]string          `json:"content_keywords"`      // strategy -> keywords
	CustomStrategies     map[string]ProcessingStrategy `json:"custom_strategies"`
	SelectionRules       []SelectionRule              `json:"selection_rules"`
	EnableContentAnalysis bool                        `json:"enable_content_analysis"`
}

// ProcessingStrategy defines a chunking strategy
type ProcessingStrategy struct {
	Name            string  `json:"name"`
	ChunkingMethod  string  `json:"chunking_method"` // paragraph_based, section_based, boundary_based
	BoundaryType    string  `json:"boundary_type"`   // paragraph, sentence, line, semantic
	PreferredSize   int64   `json:"preferred_size"`
	OverlapSize     int64   `json:"overlap_size"`
	PreserveContext bool    `json:"preserve_context"`
	Priority        int     `json:"priority"`        // Higher = more preferred
	QualityRequired float64 `json:"quality_required"` // Minimum extraction quality
}

// SelectionRule defines a rule for strategy selection
type SelectionRule struct {
	Name        string            `json:"name"`
	Condition   RuleCondition     `json:"condition"`
	Strategy    string            `json:"strategy"`
	Priority    int               `json:"priority"` // Higher = evaluated first
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// RuleCondition defines conditions for rule matching
type RuleCondition struct {
	FileExtension    []string `json:"file_extension,omitempty"`
	MimeType         []string `json:"mime_type,omitempty"`
	WordCountMin     int      `json:"word_count_min,omitempty"`
	WordCountMax     int      `json:"word_count_max,omitempty"`
	QualityMin       float64  `json:"quality_min,omitempty"`
	ContentKeywords  []string `json:"content_keywords,omitempty"`
	ExtractionMethod []string `json:"extraction_method,omitempty"`
}

// StrategySelectionRequest represents a strategy selection request
type StrategySelectionRequest struct {
	RequestID        string                 `json:"request_id"`              // Unique identifier for tracking
	DocumentPath     string                 `json:"document_path,omitempty"` // Path to original document
	DocumentMetadata DocumentMetadata       `json:"document_metadata"`       // Document metadata
	ExtractionResult ExtractionResult       `json:"extraction_result"`       // Text extraction results
	Preferences      SelectionPreferences   `json:"preferences,omitempty"`   // User preferences
	Metadata         map[string]interface{} `json:"metadata,omitempty"`      // Additional request metadata
	ReplyTo          string                 `json:"reply_to,omitempty"`      // Optional reply destination
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

// ExtractionResult contains text extraction results
type ExtractionResult struct {
	Text             string        `json:"text"`
	Extractor        string        `json:"extractor"`
	Duration         time.Duration `json:"duration"`
	Quality          float64       `json:"quality"`          // 0.0-1.0
	WordCount        int           `json:"word_count"`
	CharacterCount   int           `json:"character_count"`
	Language         string        `json:"language,omitempty"`
	Confidence       float64       `json:"confidence"`       // 0.0-1.0
}

// SelectionPreferences contains user preferences for strategy selection
type SelectionPreferences struct {
	PreferredStrategy string `json:"preferred_strategy,omitempty"`
	ChunkSizeHint     int64  `json:"chunk_size_hint,omitempty"`
	PreserveStructure bool   `json:"preserve_structure,omitempty"`
	AnalysisDepth     string `json:"analysis_depth,omitempty"` // "fast", "balanced", "thorough"
}

// StrategySelectionResponse represents the result of strategy selection
type StrategySelectionResponse struct {
	RequestID          string                 `json:"request_id"`                    // Original request identifier
	Success            bool                   `json:"success"`                       // Selection success status
	SelectedStrategy   ProcessingStrategy     `json:"selected_strategy,omitempty"`   // Recommended strategy
	AlternativeStrategies []ProcessingStrategy `json:"alternative_strategies,omitempty"` // Alternative options
	SelectionReason    string                 `json:"selection_reason,omitempty"`    // Explanation of selection
	AnalysisResults    AnalysisResults        `json:"analysis_results,omitempty"`    // Detailed analysis
	ProcessingTime     time.Duration          `json:"processing_time"`               // Time taken for selection
	Error              string                 `json:"error,omitempty"`               // Error message (on failure)
	Metadata           map[string]interface{} `json:"metadata,omitempty"`            // Additional response metadata
	OriginalRequest    StrategySelectionRequest `json:"original_request,omitempty"`  // Echo of original request
}

// AnalysisResults contains detailed analysis information
type AnalysisResults struct {
	ContentClassification string            `json:"content_classification"` // academic, legal, technical, general
	DetectedKeywords      []string          `json:"detected_keywords"`
	QualityScore          float64           `json:"quality_score"`
	ComplexityScore       float64           `json:"complexity_score"`
	StructureAnalysis     StructureAnalysis `json:"structure_analysis"`
	MatchedRules          []string          `json:"matched_rules"`
}

// StructureAnalysis contains document structure information
type StructureAnalysis struct {
	HasSections     bool    `json:"has_sections"`
	HasHeadings     bool    `json:"has_headings"`
	HasLists        bool    `json:"has_lists"`
	ParagraphCount  int     `json:"paragraph_count"`
	AverageParagraphLength float64 `json:"average_paragraph_length"`
	StructureScore  float64 `json:"structure_score"` // 0.0-1.0
}

// Init initializes the strategy selector agent with strategies and rules.
//
// This method is called once during agent startup after BaseAgent initialization.
// It loads the selection configuration, initializes processing strategies,
// and prepares the agent for strategy selection operations.
//
// Parameters:
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - error: Initialization error or nil on success
//
// Called by: GOX agent framework during startup
// Calls: config loading, strategy initialization methods
func (s *StrategySelector) Init(base *agent.BaseAgent) error {
	// Load configuration from base agent or defaults
	cfg, err := s.loadConfiguration(base)
	if err != nil {
		return fmt.Errorf("failed to load selector configuration: %w", err)
	}
	s.config = cfg

	// Initialize processing strategies
	s.strategies = make(map[string]ProcessingStrategy)

	// Load built-in strategies
	builtinStrategies := getBuiltinProcessingStrategies()
	for name, strategy := range builtinStrategies {
		s.strategies[name] = strategy
	}

	// Load custom strategies from configuration
	for name, strategy := range cfg.CustomStrategies {
		s.strategies[name] = strategy
	}

	// Initialize selection rules
	s.rules = append(getBuiltinSelectionRules(), cfg.SelectionRules...)

	agentID := base.GetConfigString("agent_id", "strategy-selector")

	base.LogInfo("Strategy Selector Agent initialized with %d strategies and %d rules", len(s.strategies), len(s.rules))
	base.LogDebug("Agent ID: %s", agentID)
	base.LogDebug("Default strategy: %s", cfg.DefaultStrategy)
	base.LogDebug("Content analysis enabled: %t", cfg.EnableContentAnalysis)

	// Log registered strategies for debugging
	for name, strategy := range s.strategies {
		base.LogDebug("Registered strategy: %s (method: %s, priority: %d, quality: %.2f)",
			name, strategy.ChunkingMethod, strategy.Priority, strategy.QualityRequired)
	}

	return nil
}

// ProcessMessage performs strategy selection on incoming requests.
//
// This is the core business logic for the strategy selector agent. It receives
// StrategySelectionRequest messages, analyzes document and extraction data,
// and returns StrategySelectionResponse messages with strategy recommendations.
//
// Processing Steps:
// 1. Parse StrategySelectionRequest from message payload
// 2. Validate request and extract analysis data
// 3. Apply selection rules and content analysis
// 4. Select optimal strategy based on criteria
// 5. Create response message with results and analysis
//
// Parameters:
//   - msg: BrokerMessage containing StrategySelectionRequest in payload
//   - base: BaseAgent providing logging and framework integration
//
// Returns:
//   - *client.BrokerMessage: Response message with strategy selection results
//   - error: Always nil (errors are returned in response message)
//
// Called by: GOX agent framework during message processing
// Calls: Analysis methods, selection algorithms, response creation methods
func (s *StrategySelector) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Parse StrategySelectionRequest from message payload
	var request StrategySelectionRequest
	var payload []byte

	// Handle different payload types
	switch p := msg.Payload.(type) {
	case []byte:
		payload = p
	case string:
		payload = []byte(p)
	default:
		return s.createErrorResponse("unknown", "Invalid payload type", base), nil
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		base.LogError("Failed to parse strategy selection request: %v", err)
		return s.createErrorResponse("unknown", "Invalid request format: "+err.Error(), base), nil
	}

	base.LogDebug("Processing strategy selection request %s", request.RequestID)

	// Validate request
	if request.ExtractionResult.Text == "" {
		return s.createErrorResponse(request.RequestID, "Extraction result text is required", base), nil
	}

	startTime := time.Now()

	// Perform document analysis
	analysis, err := s.analyzeDocument(request, base)
	if err != nil {
		return s.createErrorResponse(request.RequestID, err.Error(), base), nil
	}

	// Select optimal strategy
	strategy, alternatives, reason, err := s.selectOptimalStrategy(request, analysis, base)
	if err != nil {
		base.LogError("Strategy selection failed for request %s: %v", request.RequestID, err)
		return s.createErrorResponse(request.RequestID, err.Error(), base), nil
	}

	processingTime := time.Since(startTime)

	base.LogInfo("Successfully selected strategy for request %s: %s (reason: %s, time: %v)",
		request.RequestID, strategy.Name, reason, processingTime)

	return s.createSuccessResponse(request, strategy, alternatives, reason, analysis, processingTime, base), nil
}

// loadConfiguration loads selector configuration from BaseAgent or defaults
func (s *StrategySelector) loadConfiguration(base *agent.BaseAgent) (*SelectorConfig, error) {
	cfg := DefaultSelectorConfig()

	// Override with BaseAgent configuration if available
	if defaultStrategy := base.GetConfigString("default_strategy", ""); defaultStrategy != "" {
		cfg.DefaultStrategy = defaultStrategy
	}

	if enableContentAnalysis := base.GetConfigBool("enable_content_analysis", false); enableContentAnalysis {
		cfg.EnableContentAnalysis = enableContentAnalysis
	}

	base.LogDebug("Using selector configuration with BaseAgent overrides")
	return cfg, nil
}

// analyzeDocument performs comprehensive document analysis
func (s *StrategySelector) analyzeDocument(request StrategySelectionRequest, base *agent.BaseAgent) (AnalysisResults, error) {
	text := request.ExtractionResult.Text

	analysis := AnalysisResults{
		QualityScore: request.ExtractionResult.Quality,
	}

	// Content classification
	if s.config.EnableContentAnalysis {
		analysis.ContentClassification = s.classifyContent(text)
		analysis.DetectedKeywords = s.extractKeywords(text)
		analysis.ComplexityScore = s.calculateComplexityScore(text)
	} else {
		analysis.ContentClassification = "general"
		analysis.ComplexityScore = 0.5
	}

	// Structure analysis
	analysis.StructureAnalysis = s.analyzeStructure(text)

	// Rule matching
	analysis.MatchedRules = s.findMatchingRules(request, analysis)

	base.LogDebug("Document analysis completed for request %s: classification=%s, quality=%.2f, complexity=%.2f",
		request.RequestID, analysis.ContentClassification, analysis.QualityScore, analysis.ComplexityScore)

	return analysis, nil
}

// classifyContent classifies content into categories
func (s *StrategySelector) classifyContent(text string) string {
	textLower := strings.ToLower(text)

	// Academic paper indicators
	academicKeywords := []string{"abstract", "methodology", "results", "conclusion", "references", "hypothesis", "research", "study", "analysis"}
	academicScore := s.countKeywords(textLower, academicKeywords)

	// Legal document indicators
	legalKeywords := []string{"contract", "agreement", "clause", "whereas", "party", "shall", "hereby", "legal", "court", "law"}
	legalScore := s.countKeywords(textLower, legalKeywords)

	// Technical manual indicators
	technicalKeywords := []string{"procedure", "step", "installation", "configuration", "manual", "guide", "setup", "system", "technical"}
	technicalScore := s.countKeywords(textLower, technicalKeywords)

	// Determine classification based on highest score
	maxScore := academicScore
	classification := "academic"

	if legalScore > maxScore {
		maxScore = legalScore
		classification = "legal"
	}

	if technicalScore > maxScore {
		maxScore = technicalScore
		classification = "technical"
	}

	// If no clear classification, default to general
	if maxScore < 3 { // Minimum threshold for classification
		classification = "general"
	}

	return classification
}

// extractKeywords extracts relevant keywords from text
func (s *StrategySelector) extractKeywords(text string) []string {
	// Simple keyword extraction - in production would use more sophisticated NLP
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Count word frequency
	for _, word := range words {
		// Filter out common words and short words
		if len(word) > 4 && !s.isCommonWord(word) {
			wordCount[word]++
		}
	}

	// Extract top keywords
	var keywords []string
	for word, count := range wordCount {
		if count > 2 { // Minimum frequency
			keywords = append(keywords, word)
		}
	}

	// Limit to top 10 keywords
	if len(keywords) > 10 {
		keywords = keywords[:10]
	}

	return keywords
}

// calculateComplexityScore calculates text complexity score
func (s *StrategySelector) calculateComplexityScore(text string) float64 {
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	paragraphs := strings.Split(text, "\n\n")

	if len(sentences) == 0 || len(words) == 0 {
		return 0.0
	}

	// Average words per sentence
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))

	// Average sentences per paragraph
	avgSentencesPerParagraph := float64(len(sentences)) / float64(len(paragraphs))

	// Complexity score based on structure
	complexity := (avgWordsPerSentence + avgSentencesPerParagraph) / 30.0 // Normalize

	if complexity > 1.0 {
		complexity = 1.0
	}

	return complexity
}

// analyzeStructure analyzes document structure
func (s *StrategySelector) analyzeStructure(text string) StructureAnalysis {
	analysis := StructureAnalysis{}

	// Check for sections and headings
	analysis.HasSections = strings.Contains(text, "\n#") || strings.Contains(text, "Chapter") || strings.Contains(text, "Section")
	analysis.HasHeadings = strings.Contains(text, "\n##") || strings.Contains(text, "\n###")
	analysis.HasLists = strings.Contains(text, "\n-") || strings.Contains(text, "\n*") || strings.Contains(text, "\n1.")

	// Count paragraphs
	paragraphs := strings.Split(text, "\n\n")
	analysis.ParagraphCount = len(paragraphs)

	// Calculate average paragraph length
	totalLength := 0
	for _, para := range paragraphs {
		totalLength += len(strings.TrimSpace(para))
	}

	if len(paragraphs) > 0 {
		analysis.AverageParagraphLength = float64(totalLength) / float64(len(paragraphs))
	}

	// Calculate structure score
	structureScore := 0.0
	if analysis.HasSections {
		structureScore += 0.3
	}
	if analysis.HasHeadings {
		structureScore += 0.3
	}
	if analysis.HasLists {
		structureScore += 0.2
	}
	if analysis.ParagraphCount > 5 {
		structureScore += 0.2
	}

	analysis.StructureScore = structureScore

	return analysis
}

// findMatchingRules finds rules that match the request
func (s *StrategySelector) findMatchingRules(request StrategySelectionRequest, analysis AnalysisResults) []string {
	var matchedRules []string

	for _, rule := range s.rules {
		if s.ruleMatches(rule, request, analysis) {
			matchedRules = append(matchedRules, rule.Name)
		}
	}

	return matchedRules
}

// ruleMatches checks if a rule matches the given request and analysis
func (s *StrategySelector) ruleMatches(rule SelectionRule, request StrategySelectionRequest, analysis AnalysisResults) bool {
	condition := rule.Condition

	// Check file extension
	if len(condition.FileExtension) > 0 {
		ext := strings.ToLower(filepath.Ext(request.DocumentPath))
		if !s.contains(condition.FileExtension, ext) {
			return false
		}
	}

	// Check MIME type
	if len(condition.MimeType) > 0 {
		if !s.contains(condition.MimeType, request.DocumentMetadata.MimeType) {
			return false
		}
	}

	// Check word count range
	wordCount := request.ExtractionResult.WordCount
	if condition.WordCountMin > 0 && wordCount < condition.WordCountMin {
		return false
	}
	if condition.WordCountMax > 0 && wordCount > condition.WordCountMax {
		return false
	}

	// Check quality minimum
	if condition.QualityMin > 0 && analysis.QualityScore < condition.QualityMin {
		return false
	}

	// Check content keywords
	if len(condition.ContentKeywords) > 0 {
		if !s.hasAnyKeywords(analysis.DetectedKeywords, condition.ContentKeywords) {
			return false
		}
	}

	// Check extraction method
	if len(condition.ExtractionMethod) > 0 {
		if !s.contains(condition.ExtractionMethod, request.DocumentMetadata.ExtractionMethod) {
			return false
		}
	}

	return true
}

// selectOptimalStrategy selects the best strategy based on analysis
func (s *StrategySelector) selectOptimalStrategy(request StrategySelectionRequest, analysis AnalysisResults, base *agent.BaseAgent) (ProcessingStrategy, []ProcessingStrategy, string, error) {
	// Check user preferences first
	if request.Preferences.PreferredStrategy != "" {
		if strategy, exists := s.strategies[request.Preferences.PreferredStrategy]; exists {
			return strategy, []ProcessingStrategy{}, "user preference", nil
		}
	}

	// Apply rule-based selection
	for _, ruleName := range analysis.MatchedRules {
		for _, rule := range s.rules {
			if rule.Name == ruleName {
				if strategy, exists := s.strategies[rule.Strategy]; exists {
					alternatives := s.getAlternativeStrategies(strategy, 2)
					return strategy, alternatives, fmt.Sprintf("rule: %s", rule.Name), nil
				}
			}
		}
	}

	// Content-based selection
	var strategyName string
	var reason string

	switch analysis.ContentClassification {
	case "academic":
		strategyName = "academic_paper"
		reason = "academic content detected"
	case "legal":
		strategyName = "legal_document"
		reason = "legal content detected"
	case "technical":
		strategyName = "technical_manual"
		reason = "technical content detected"
	default:
		// Quality and structure-based selection
		if analysis.QualityScore > 0.8 && analysis.StructureAnalysis.StructureScore > 0.6 {
			strategyName = "document_aware"
			reason = "high quality with good structure"
		} else {
			strategyName = s.config.DefaultStrategy
			reason = "default strategy"
		}
	}

	strategy, exists := s.strategies[strategyName]
	if !exists {
		strategy = s.strategies[s.config.DefaultStrategy]
		reason = "fallback to default"
	}

	alternatives := s.getAlternativeStrategies(strategy, 2)
	return strategy, alternatives, reason, nil
}

// Helper methods
func (s *StrategySelector) countKeywords(text string, keywords []string) int {
	count := 0
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			count++
		}
	}
	return count
}

func (s *StrategySelector) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "that": true, "have": true, "for": true,
		"not": true, "with": true, "you": true, "this": true, "but": true,
		"his": true, "from": true, "they": true, "she": true, "her": true,
		"been": true, "than": true, "its": true, "were": true, "said": true,
	}
	return commonWords[word]
}

func (s *StrategySelector) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *StrategySelector) hasAnyKeywords(detected, required []string) bool {
	for _, req := range required {
		for _, det := range detected {
			if strings.Contains(det, req) {
				return true
			}
		}
	}
	return false
}

func (s *StrategySelector) getAlternativeStrategies(selected ProcessingStrategy, count int) []ProcessingStrategy {
	var alternatives []ProcessingStrategy
	for _, strategy := range s.strategies {
		if strategy.Name != selected.Name && len(alternatives) < count {
			alternatives = append(alternatives, strategy)
		}
	}
	return alternatives
}

// createSuccessResponse creates a successful strategy selection response
func (s *StrategySelector) createSuccessResponse(request StrategySelectionRequest, strategy ProcessingStrategy, alternatives []ProcessingStrategy, reason string, analysis AnalysisResults, processingTime time.Duration, base *agent.BaseAgent) *client.BrokerMessage {
	response := StrategySelectionResponse{
		RequestID:             request.RequestID,
		Success:               true,
		SelectedStrategy:      strategy,
		AlternativeStrategies: alternatives,
		SelectionReason:       reason,
		AnalysisResults:       analysis,
		ProcessingTime:        processingTime,
		OriginalRequest:       request,
	}

	responseBytes, _ := json.Marshal(response)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("strategy_selection_success_%s", request.RequestID),
		Target:  base.GetEgress(),
		Type:    "strategy_selection_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"selection_success":    true,
			"original_request":     request.RequestID,
			"selected_strategy":    strategy.Name,
			"selection_reason":     reason,
			"processing_time_ms":   processingTime.Milliseconds(),
			"content_classification": analysis.ContentClassification,
			"quality_score":        analysis.QualityScore,
		},
	}
}

// createErrorResponse creates an error response message for failed selection
func (s *StrategySelector) createErrorResponse(requestID, errorMsg string, base *agent.BaseAgent) *client.BrokerMessage {
	response := StrategySelectionResponse{
		RequestID: requestID,
		Success:   false,
		Error:     errorMsg,
	}
	responseBytes, _ := json.Marshal(response)

	base.LogError("Strategy selection error for request %s: %s", requestID, errorMsg)

	return &client.BrokerMessage{
		ID:      fmt.Sprintf("strategy_selection_error_%s", requestID),
		Target:  base.GetEgress(),
		Type:    "strategy_selection_response",
		Payload: responseBytes,
		Meta: map[string]interface{}{
			"selection_success": false,
			"original_request":  requestID,
			"error_message":     errorMsg,
		},
	}
}

// getBuiltinProcessingStrategies returns built-in processing strategies
func getBuiltinProcessingStrategies() map[string]ProcessingStrategy {
	return map[string]ProcessingStrategy{
		"document_aware": {
			Name:            "document_aware",
			ChunkingMethod:  "paragraph_based",
			BoundaryType:    "paragraph",
			PreferredSize:   2048,
			OverlapSize:     256,
			PreserveContext: true,
			Priority:        90,
			QualityRequired: 0.7,
		},
		"academic_paper": {
			Name:            "academic_paper",
			ChunkingMethod:  "section_based",
			BoundaryType:    "semantic",
			PreferredSize:   1024,
			OverlapSize:     128,
			PreserveContext: true,
			Priority:        95,
			QualityRequired: 0.8,
		},
		"legal_document": {
			Name:            "legal_document",
			ChunkingMethod:  "section_based",
			BoundaryType:    "paragraph",
			PreferredSize:   1024,
			OverlapSize:     512,
			PreserveContext: true,
			Priority:        95,
			QualityRequired: 0.8,
		},
		"technical_manual": {
			Name:            "technical_manual",
			ChunkingMethod:  "boundary_based",
			BoundaryType:    "line",
			PreferredSize:   1536,
			OverlapSize:     256,
			PreserveContext: true,
			Priority:        85,
			QualityRequired: 0.6,
		},
		"general_text": {
			Name:            "general_text",
			ChunkingMethod:  "boundary_based",
			BoundaryType:    "sentence",
			PreferredSize:   2048,
			OverlapSize:     256,
			PreserveContext: false,
			Priority:        70,
			QualityRequired: 0.0,
		},
	}
}

// getBuiltinSelectionRules returns built-in selection rules
func getBuiltinSelectionRules() []SelectionRule {
	return []SelectionRule{
		{
			Name: "high_quality_pdf",
			Condition: RuleCondition{
				FileExtension: []string{".pdf"},
				QualityMin:    0.8,
			},
			Strategy: "document_aware",
			Priority: 100,
		},
		{
			Name: "academic_content",
			Condition: RuleCondition{
				ContentKeywords: []string{"abstract", "methodology", "results"},
				WordCountMin:    1000,
			},
			Strategy: "academic_paper",
			Priority: 95,
		},
		{
			Name: "legal_content",
			Condition: RuleCondition{
				ContentKeywords: []string{"contract", "agreement", "clause"},
			},
			Strategy: "legal_document",
			Priority: 95,
		},
		{
			Name: "short_text",
			Condition: RuleCondition{
				WordCountMax: 500,
			},
			Strategy: "general_text",
			Priority: 80,
		},
	}
}

// DefaultSelectorConfig returns default configuration
func DefaultSelectorConfig() *SelectorConfig {
	return &SelectorConfig{
		DefaultStrategy: "general_text",
		FormatMappings: map[string]string{
			".pdf":  "document_aware",
			".docx": "document_aware",
			".txt":  "general_text",
		},
		QualityThresholds: map[string]float64{
			"academic_paper":   0.8,
			"legal_document":   0.8,
			"document_aware":   0.7,
			"technical_manual": 0.6,
			"general_text":     0.0,
		},
		WordCountThresholds: map[string]int{
			"academic_paper": 1000,
			"legal_document": 500,
		},
		ContentKeywords: map[string][]string{
			"academic_paper":   {"abstract", "methodology", "results", "conclusion"},
			"legal_document":   {"contract", "agreement", "clause", "party"},
			"technical_manual": {"procedure", "step", "installation", "guide"},
		},
		CustomStrategies:      make(map[string]ProcessingStrategy),
		SelectionRules:        []SelectionRule{},
		EnableContentAnalysis: true,
	}
}

// main is the entry point for the strategy selector agent.
//
// Initializes and starts the agent using the GOX framework's agent runner.
// The framework handles all boilerplate including broker connection, message
// routing, lifecycle management, and strategy selection request processing.
//
// The agent ID "strategy-selector" is used for:
// - Agent registration with the support service
// - Logging and debugging identification
// - Message routing and correlation
//
// Called by: Operating system process execution
// Calls: agent.Run() with StrategySelector implementation
func main() {
	// Framework handles all boilerplate including broker connection, message
	// routing, strategy selection processing, and lifecycle management
	if err := agent.Run(&StrategySelector{}, "strategy-selector"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}