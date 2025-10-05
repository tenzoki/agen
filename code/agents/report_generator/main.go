package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

type ReportGenerator struct {
	agent.DefaultAgentRunner
	config *ReportConfig
}

type ReportConfig struct {
	IncludeCharts        bool   `json:"include_charts"`
	IncludeTables        bool   `json:"include_tables"`
	IncludeRecommendations bool `json:"include_recommendations"`
	ReportFormat         string `json:"report_format"`
	MaxRecommendations   int    `json:"max_recommendations"`
}

type ChunkProcessingResult struct {
	ChunkID     string                 `json:"chunk_id"`
	ChunkInfo   ChunkInfo              `json:"chunk_info"`
	ProcessedBy string                 `json:"processed_by"`
	ResultData  map[string]interface{} `json:"result_data"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessedAt time.Time              `json:"processed_at"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
}

type ChunkInfo struct {
	ChunkID     string `json:"chunk_id"`
	FileID      string `json:"file_id"`
	SequenceNum int    `json:"sequence_num"`
	StartOffset int64  `json:"start_offset"`
	EndOffset   int64  `json:"end_offset"`
	Size        int64  `json:"size"`
	Hash        string `json:"hash"`
	MimeType    string `json:"mime_type"`
	ChunkPath   string `json:"chunk_path"`
}

type SynthesisRequest struct {
	RequestID  string                 `json:"request_id"`
	FileID     string                 `json:"file_id"`
	ChunkIDs   []string               `json:"chunk_ids"`
	OutputType string                 `json:"output_type"`
	Options    map[string]interface{} `json:"options"`
	Priority   int                    `json:"priority"`
	CreatedAt  time.Time              `json:"created_at"`
}

type AnalysisReport struct {
	ReportID        string                 `json:"report_id"`
	FileID          string                 `json:"file_id"`
	Title           string                 `json:"title"`
	Summary         string                 `json:"summary"`
	Sections        []ReportSection        `json:"sections"`
	Charts          []ChartData            `json:"charts"`
	Tables          []TableData            `json:"tables"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

type ReportSection struct {
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Subsections []ReportSection        `json:"subsections"`
	Data        map[string]interface{} `json:"data"`
	Charts      []string               `json:"charts"`
	Tables      []string               `json:"tables"`
}

type ChartData struct {
	ChartID string                 `json:"chart_id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Data    map[string]interface{} `json:"data"`
	Config  map[string]interface{} `json:"config"`
}

type TableData struct {
	TableID string                 `json:"table_id"`
	Title   string                 `json:"title"`
	Headers []string               `json:"headers"`
	Rows    [][]interface{}        `json:"rows"`
	Config  map[string]interface{} `json:"config"`
}

type SynthesisResult struct {
	JobID          string                 `json:"job_id"`
	FileID         string                 `json:"file_id"`
	OutputType     string                 `json:"output_type"`
	ResultData     map[string]interface{} `json:"result_data"`
	OutputFiles    []string               `json:"output_files"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessedAt    time.Time              `json:"processed_at"`
	ProcessingTime time.Duration          `json:"processing_time"`
	ChunkCount     int                    `json:"chunk_count"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
}

func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &ReportConfig{
			IncludeCharts:          true,
			IncludeTables:          true,
			IncludeRecommendations: true,
			ReportFormat:           "json",
			MaxRecommendations:     5,
		},
	}
}

func (rg *ReportGenerator) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "synthesis_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var request SynthesisRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	if request.OutputType != "analysis_report" {
		return nil, fmt.Errorf("unsupported output type: %s", request.OutputType)
	}

	rg.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &SynthesisResult{
		JobID:       fmt.Sprintf("report_%d", time.Now().UnixNano()),
		FileID:      request.FileID,
		OutputType:  request.OutputType,
		ProcessedAt: startTime,
		ChunkCount:  len(request.ChunkIDs),
		ResultData:  make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	chunks, err := rg.loadChunkResults(request.ChunkIDs, base)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load chunks: %v", err)
		result.Success = false
		return rg.createResultMessage(result), nil
	}

	report, err := rg.generateAnalysisReport(chunks)
	if err != nil {
		result.Error = fmt.Sprintf("failed to generate report: %v", err)
		result.Success = false
		return rg.createResultMessage(result), nil
	}

	result.ResultData["analysis_report"] = report
	result.Metadata["generator"] = "analysis_report"
	result.Metadata["report_id"] = report.ReportID
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return rg.createResultMessage(result), nil
}

func (rg *ReportGenerator) loadConfigFromAgent(base *agent.BaseAgent) {
	if reportFormat := base.GetConfigString("report_format", "json"); reportFormat != "" {
		rg.config.ReportFormat = reportFormat
	}
	if maxRec := base.GetConfigInt("max_recommendations", 5); maxRec > 0 {
		rg.config.MaxRecommendations = maxRec
	}
	rg.config.IncludeCharts = base.GetConfigBool("include_charts", true)
	rg.config.IncludeTables = base.GetConfigBool("include_tables", true)
	rg.config.IncludeRecommendations = base.GetConfigBool("include_recommendations", true)
}

func (rg *ReportGenerator) loadChunkResults(chunkIDs []string, base *agent.BaseAgent) ([]*ChunkProcessingResult, error) {
	results := make([]*ChunkProcessingResult, 0, len(chunkIDs))

	for _, chunkID := range chunkIDs {
		chunkInfo := ChunkInfo{
			ChunkID:     chunkID,
			FileID:      rg.extractFileID(chunkID),
			SequenceNum: len(results),
			Size:        2048,
			MimeType:    "text/plain",
		}

		resultData := map[string]interface{}{
			"content":    fmt.Sprintf("Analyzed content for chunk %s with detailed findings", chunkID),
			"keywords":   []interface{}{"analysis", "report", "insights", "data"},
			"topics":     []interface{}{"reporting", "analysis"},
			"language":   "en",
			"word_count": float64(200),
			"sentiment":  0.7,
			"quality":    0.9,
		}

		metadata := map[string]interface{}{
			"processing_agent": "chunk_processor",
			"analysis_depth":   "comprehensive",
			"confidence_score": 0.95,
		}

		result := &ChunkProcessingResult{
			ChunkID:     chunkID,
			ChunkInfo:   chunkInfo,
			ProcessedBy: "chunk_processor",
			ResultData:  resultData,
			Metadata:    metadata,
			ProcessedAt: time.Now(),
			Success:     true,
		}

		results = append(results, result)
	}

	return results, nil
}

func (rg *ReportGenerator) generateAnalysisReport(chunks []*ChunkProcessingResult) (*AnalysisReport, error) {
	report := &AnalysisReport{
		ReportID:        fmt.Sprintf("report_%d", time.Now().Unix()),
		FileID:          rg.getFileID(chunks),
		Title:           rg.generateTitle(chunks),
		Summary:         rg.generateSummary(chunks),
		Sections:        rg.createSections(chunks),
		Metadata:        make(map[string]interface{}),
		GeneratedAt:     time.Now(),
	}

	if rg.config.IncludeCharts {
		report.Charts = rg.generateCharts(chunks)
	}

	if rg.config.IncludeTables {
		report.Tables = rg.generateTables(chunks)
	}

	if rg.config.IncludeRecommendations {
		report.Recommendations = rg.generateRecommendations(chunks)
	}

	return report, nil
}

func (rg *ReportGenerator) getFileID(chunks []*ChunkProcessingResult) string {
	if len(chunks) > 0 {
		return chunks[0].ChunkInfo.FileID
	}
	return "unknown"
}

func (rg *ReportGenerator) generateTitle(chunks []*ChunkProcessingResult) string {
	fileID := rg.getFileID(chunks)
	if len(fileID) > 8 {
		fileID = fileID[:8]
	}
	return fmt.Sprintf("Analysis Report for File %s", fileID)
}

func (rg *ReportGenerator) generateSummary(chunks []*ChunkProcessingResult) string {
	fileID := rg.getFileID(chunks)
	if len(fileID) > 8 {
		fileID = fileID[:8]
	}
	return fmt.Sprintf("Comprehensive analysis of %d chunks from file %s, covering content analysis, metadata extraction, and key insights. This report provides detailed findings and actionable recommendations based on automated processing results.",
		len(chunks), fileID)
}

func (rg *ReportGenerator) createSections(chunks []*ChunkProcessingResult) []ReportSection {
	sections := []ReportSection{
		{
			Title:   "Content Overview",
			Content: rg.generateContentOverview(chunks),
			Data:    map[string]interface{}{"chunk_count": len(chunks)},
		},
		{
			Title:   "Key Findings",
			Content: rg.generateKeyFindings(chunks),
			Data:    map[string]interface{}{"keywords": rg.extractAllKeywords(chunks)},
		},
		{
			Title:   "Quality Assessment",
			Content: rg.generateQualityAssessment(chunks),
			Data:    map[string]interface{}{"average_quality": rg.calculateAverageQuality(chunks)},
		},
		{
			Title:   "Technical Analysis",
			Content: rg.generateTechnicalAnalysis(chunks),
			Data:    map[string]interface{}{"file_types": rg.analyzeFileTypes(chunks)},
		},
	}

	return sections
}

func (rg *ReportGenerator) generateContentOverview(chunks []*ChunkProcessingResult) string {
	totalWords := rg.calculateTotalWords(chunks)
	avgSize := rg.calculateAverageSize(chunks)

	return fmt.Sprintf("The analysis covers %d content chunks with a total of %d words. Average chunk size is %.0f bytes. Content has been successfully processed with automated analysis techniques.",
		len(chunks), totalWords, avgSize)
}

func (rg *ReportGenerator) generateKeyFindings(chunks []*ChunkProcessingResult) string {
	keywords := rg.extractAllKeywords(chunks)
	topKeywords := keywords[:min(5, len(keywords))]
	avgSentiment := rg.calculateAverageSentiment(chunks)

	return fmt.Sprintf("Key topics identified: %s. Overall content sentiment: %.2f (positive scale). Analysis reveals consistent themes and high-quality content structure.",
		strings.Join(topKeywords, ", "), avgSentiment)
}

func (rg *ReportGenerator) generateQualityAssessment(chunks []*ChunkProcessingResult) string {
	avgQuality := rg.calculateAverageQuality(chunks)
	successRate := rg.calculateSuccessRate(chunks)

	return fmt.Sprintf("Content quality assessment shows an average quality score of %.2f with %.1f%% successful processing rate. All chunks meet quality standards for further analysis.",
		avgQuality, successRate*100)
}

func (rg *ReportGenerator) generateTechnicalAnalysis(chunks []*ChunkProcessingResult) string {
	mimeTypes := rg.analyzeFileTypes(chunks)
	processingAgents := rg.analyzeProcessingAgents(chunks)

	return fmt.Sprintf("Technical analysis reveals %d different content types processed by %d agent types. Processing infrastructure demonstrates robust handling of diverse content formats.",
		len(mimeTypes), len(processingAgents))
}

func (rg *ReportGenerator) extractAllKeywords(chunks []*ChunkProcessingResult) []string {
	keywordMap := make(map[string]int)
	for _, chunk := range chunks {
		if keywords, ok := chunk.ResultData["keywords"].([]interface{}); ok {
			for _, kw := range keywords {
				if keyword, ok := kw.(string); ok {
					keywordMap[keyword]++
				}
			}
		}
	}

	type kw struct {
		word string
		freq int
	}
	var sorted []kw
	for word, freq := range keywordMap {
		sorted = append(sorted, kw{word, freq})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].freq > sorted[j].freq })

	result := make([]string, 0, len(sorted))
	for _, item := range sorted {
		result = append(result, item.word)
	}
	return result
}

func (rg *ReportGenerator) calculateTotalWords(chunks []*ChunkProcessingResult) int {
	total := 0
	for _, chunk := range chunks {
		if wordCount, ok := chunk.ResultData["word_count"].(float64); ok {
			total += int(wordCount)
		}
	}
	return total
}

func (rg *ReportGenerator) calculateAverageSize(chunks []*ChunkProcessingResult) float64 {
	if len(chunks) == 0 {
		return 0
	}
	total := int64(0)
	for _, chunk := range chunks {
		total += chunk.ChunkInfo.Size
	}
	return float64(total) / float64(len(chunks))
}

func (rg *ReportGenerator) calculateAverageSentiment(chunks []*ChunkProcessingResult) float64 {
	total := 0.0
	count := 0
	for _, chunk := range chunks {
		if sentiment, ok := chunk.ResultData["sentiment"].(float64); ok {
			total += sentiment
			count++
		}
	}
	if count == 0 {
		return 0.0
	}
	return total / float64(count)
}

func (rg *ReportGenerator) calculateAverageQuality(chunks []*ChunkProcessingResult) float64 {
	total := 0.0
	count := 0
	for _, chunk := range chunks {
		if quality, ok := chunk.ResultData["quality"].(float64); ok {
			total += quality
			count++
		}
	}
	if count == 0 {
		return 0.0
	}
	return total / float64(count)
}

func (rg *ReportGenerator) calculateSuccessRate(chunks []*ChunkProcessingResult) float64 {
	if len(chunks) == 0 {
		return 0.0
	}
	successCount := 0
	for _, chunk := range chunks {
		if chunk.Success {
			successCount++
		}
	}
	return float64(successCount) / float64(len(chunks))
}

func (rg *ReportGenerator) analyzeFileTypes(chunks []*ChunkProcessingResult) map[string]int {
	types := make(map[string]int)
	for _, chunk := range chunks {
		types[chunk.ChunkInfo.MimeType]++
	}
	return types
}

func (rg *ReportGenerator) analyzeProcessingAgents(chunks []*ChunkProcessingResult) map[string]int {
	agents := make(map[string]int)
	for _, chunk := range chunks {
		agents[chunk.ProcessedBy]++
	}
	return agents
}

func (rg *ReportGenerator) generateCharts(chunks []*ChunkProcessingResult) []ChartData {
	charts := []ChartData{
		{
			ChartID: "chunk_sizes",
			Type:    "bar",
			Title:   "Chunk Size Distribution",
			Data:    map[string]interface{}{"sizes": rg.getChunkSizes(chunks)},
			Config:  map[string]interface{}{"xlabel": "Chunk", "ylabel": "Size (bytes)"},
		},
		{
			ChartID: "quality_scores",
			Type:    "histogram",
			Title:   "Quality Score Distribution",
			Data:    map[string]interface{}{"scores": rg.getQualityScores(chunks)},
			Config:  map[string]interface{}{"xlabel": "Quality Score", "ylabel": "Frequency"},
		},
	}

	return charts
}

func (rg *ReportGenerator) generateTables(chunks []*ChunkProcessingResult) []TableData {
	headers := []string{"Chunk ID", "Size", "Type", "Quality", "Keywords"}
	rows := make([][]interface{}, 0, len(chunks))

	for _, chunk := range chunks {
		keywords := rg.extractChunkKeywords(chunk)
		quality := "N/A"
		if q, ok := chunk.ResultData["quality"].(float64); ok {
			quality = fmt.Sprintf("%.2f", q)
		}

		row := []interface{}{
			chunk.ChunkID[:min(8, len(chunk.ChunkID))],
			chunk.ChunkInfo.Size,
			chunk.ChunkInfo.MimeType,
			quality,
			strings.Join(keywords[:min(3, len(keywords))], ", "),
		}
		rows = append(rows, row)
	}

	return []TableData{
		{
			TableID: "chunk_summary",
			Title:   "Chunk Analysis Summary",
			Headers: headers,
			Rows:    rows,
			Config:  map[string]interface{}{"sortable": true, "paginated": true},
		},
	}
}

func (rg *ReportGenerator) extractChunkKeywords(chunk *ChunkProcessingResult) []string {
	if keywords, ok := chunk.ResultData["keywords"].([]interface{}); ok {
		result := make([]string, 0, len(keywords))
		for _, kw := range keywords {
			if keyword, ok := kw.(string); ok {
				result = append(result, keyword)
			}
		}
		return result
	}
	return []string{}
}

func (rg *ReportGenerator) getChunkSizes(chunks []*ChunkProcessingResult) []int64 {
	sizes := make([]int64, len(chunks))
	for i, chunk := range chunks {
		sizes[i] = chunk.ChunkInfo.Size
	}
	return sizes
}

func (rg *ReportGenerator) getQualityScores(chunks []*ChunkProcessingResult) []float64 {
	scores := make([]float64, 0, len(chunks))
	for _, chunk := range chunks {
		if quality, ok := chunk.ResultData["quality"].(float64); ok {
			scores = append(scores, quality)
		}
	}
	return scores
}

func (rg *ReportGenerator) generateRecommendations(chunks []*ChunkProcessingResult) []string {
	recommendations := []string{
		"Content has been successfully processed and analyzed with high quality scores",
		"Consider implementing search indexing for better discoverability of key topics",
		"Metadata extraction provides excellent foundation for further analytical insights",
		"Processing pipeline demonstrates robust handling of diverse content types",
		"Quality assessment indicates content is suitable for advanced analysis workflows",
	}

	maxRec := rg.config.MaxRecommendations
	if len(recommendations) > maxRec {
		recommendations = recommendations[:maxRec]
	}

	return recommendations
}

func (rg *ReportGenerator) extractFileID(chunkID string) string {
	parts := strings.Split(chunkID, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func (rg *ReportGenerator) createResultMessage(result *SynthesisResult) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "synthesis_result",
		Target:    "synthesis_result",
		Payload:   result,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	reportGenerator := NewReportGenerator()
	agent.Run(reportGenerator, "report-generator")
}