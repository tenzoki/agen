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

type SummaryGenerator struct {
	agent.DefaultAgentRunner
	config *SummaryConfig
}

type SummaryConfig struct {
	MaxKeywords    int    `json:"max_keywords"`
	MaxTopics      int    `json:"max_topics"`
	SummaryLength  int    `json:"summary_length"`
	OutputFormat   string `json:"output_format"`
	EnableSections bool   `json:"enable_sections"`
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

type DocumentSummary struct {
	Title     string                 `json:"title"`
	Summary   string                 `json:"summary"`
	Keywords  []string               `json:"keywords"`
	Topics    []string               `json:"topics"`
	Language  string                 `json:"language"`
	WordCount int                    `json:"word_count"`
	PageCount int                    `json:"page_count"`
	Sections  []SectionSummary       `json:"sections"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type SectionSummary struct {
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	Keywords  []string `json:"keywords"`
	WordCount int      `json:"word_count"`
	StartPage int      `json:"start_page"`
	EndPage   int      `json:"end_page"`
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

func NewSummaryGenerator() *SummaryGenerator {
	return &SummaryGenerator{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &SummaryConfig{
			MaxKeywords:    20,
			MaxTopics:      10,
			SummaryLength:  500,
			OutputFormat:   "json",
			EnableSections: true,
		},
	}
}

func (sg *SummaryGenerator) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
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

	if request.OutputType != "document_summary" {
		return nil, fmt.Errorf("unsupported output type: %s", request.OutputType)
	}

	sg.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &SynthesisResult{
		JobID:       fmt.Sprintf("summary_%d", time.Now().UnixNano()),
		FileID:      request.FileID,
		OutputType:  request.OutputType,
		ProcessedAt: startTime,
		ChunkCount:  len(request.ChunkIDs),
		ResultData:  make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	chunks, err := sg.loadChunkResults(request.ChunkIDs, base)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load chunks: %v", err)
		result.Success = false
		return sg.createResultMessage(result), nil
	}

	summary, err := sg.generateDocumentSummary(chunks)
	if err != nil {
		result.Error = fmt.Sprintf("failed to generate summary: %v", err)
		result.Success = false
		return sg.createResultMessage(result), nil
	}

	result.ResultData["document_summary"] = summary
	result.ResultData["chunk_count"] = len(chunks)
	result.ResultData["total_size"] = sg.calculateTotalSize(chunks)
	result.Metadata["summarizer"] = "document_summary"
	result.Metadata["generated_at"] = time.Now()
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return sg.createResultMessage(result), nil
}

func (sg *SummaryGenerator) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxKeywords := base.GetConfigInt("max_keywords", 20); maxKeywords > 0 {
		sg.config.MaxKeywords = maxKeywords
	}
	if maxTopics := base.GetConfigInt("max_topics", 10); maxTopics > 0 {
		sg.config.MaxTopics = maxTopics
	}
	if summaryLength := base.GetConfigInt("summary_length", 500); summaryLength > 0 {
		sg.config.SummaryLength = summaryLength
	}
	if outputFormat := base.GetConfigString("output_format", "json"); outputFormat != "" {
		sg.config.OutputFormat = outputFormat
	}
	sg.config.EnableSections = base.GetConfigBool("enable_sections", true)
}

func (sg *SummaryGenerator) loadChunkResults(chunkIDs []string, base *agent.BaseAgent) ([]*ChunkProcessingResult, error) {
	results := make([]*ChunkProcessingResult, 0, len(chunkIDs))

	for _, chunkID := range chunkIDs {
		chunkInfo := ChunkInfo{
			ChunkID:     chunkID,
			FileID:      sg.extractFileID(chunkID),
			SequenceNum: len(results),
			Size:        1024,
			MimeType:    "text/plain",
		}

		resultData := map[string]interface{}{
			"content":    fmt.Sprintf("Content for chunk %s", chunkID),
			"keywords":   []interface{}{"keyword1", "keyword2", "keyword3"},
			"topics":     []interface{}{"topic1", "topic2"},
			"language":   "en",
			"word_count": float64(100),
			"summary":    fmt.Sprintf("Summary for chunk %s", chunkID),
		}

		result := &ChunkProcessingResult{
			ChunkID:     chunkID,
			ChunkInfo:   chunkInfo,
			ProcessedBy: "chunk_processor",
			ResultData:  resultData,
			Metadata:    make(map[string]interface{}),
			ProcessedAt: time.Now(),
			Success:     true,
		}

		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ChunkInfo.SequenceNum < results[j].ChunkInfo.SequenceNum
	})

	return results, nil
}

func (sg *SummaryGenerator) generateDocumentSummary(chunks []*ChunkProcessingResult) (*DocumentSummary, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks to summarize")
	}

	summary := &DocumentSummary{
		Title:     sg.extractTitle(chunks),
		Keywords:  sg.aggregateKeywords(chunks),
		Topics:    sg.extractTopics(chunks),
		Language:  sg.detectLanguage(chunks),
		WordCount: sg.calculateWordCount(chunks),
		PageCount: len(chunks),
		Metadata:  make(map[string]interface{}),
	}

	if sg.config.EnableSections {
		summary.Sections = sg.createSections(chunks)
	}

	summary.Summary = sg.generateSummaryText(chunks, summary.Keywords)

	return summary, nil
}

func (sg *SummaryGenerator) extractTitle(chunks []*ChunkProcessingResult) string {
	if len(chunks) == 0 {
		return "Untitled Document"
	}

	firstChunk := chunks[0]
	if data, ok := firstChunk.ResultData["content"].(string); ok {
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) > 10 && len(line) < 100 {
				return line
			}
		}
	}

	return fmt.Sprintf("Document %s", firstChunk.ChunkInfo.FileID[:8])
}

func (sg *SummaryGenerator) aggregateKeywords(chunks []*ChunkProcessingResult) []string {
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

	type keywordFreq struct {
		keyword string
		freq    int
	}

	var sortedKeywords []keywordFreq
	for keyword, freq := range keywordMap {
		sortedKeywords = append(sortedKeywords, keywordFreq{keyword, freq})
	}

	sort.Slice(sortedKeywords, func(i, j int) bool {
		return sortedKeywords[i].freq > sortedKeywords[j].freq
	})

	maxKeywords := sg.config.MaxKeywords
	if len(sortedKeywords) < maxKeywords {
		maxKeywords = len(sortedKeywords)
	}

	result := make([]string, maxKeywords)
	for i := 0; i < maxKeywords; i++ {
		result[i] = sortedKeywords[i].keyword
	}

	return result
}

func (sg *SummaryGenerator) extractTopics(chunks []*ChunkProcessingResult) []string {
	topics := make(map[string]bool)

	for _, chunk := range chunks {
		if chunkTopics, ok := chunk.ResultData["topics"].([]interface{}); ok {
			for _, topic := range chunkTopics {
				if topicStr, ok := topic.(string); ok {
					topics[topicStr] = true
				}
			}
		}
	}

	result := make([]string, 0, len(topics))
	for topic := range topics {
		result = append(result, topic)
	}

	maxTopics := sg.config.MaxTopics
	if len(result) > maxTopics {
		result = result[:maxTopics]
	}

	return result
}

func (sg *SummaryGenerator) detectLanguage(chunks []*ChunkProcessingResult) string {
	languageCount := make(map[string]int)

	for _, chunk := range chunks {
		if lang, ok := chunk.ResultData["language"].(string); ok {
			languageCount[lang]++
		}
	}

	maxCount := 0
	detectedLang := "unknown"
	for lang, count := range languageCount {
		if count > maxCount {
			maxCount = count
			detectedLang = lang
		}
	}

	return detectedLang
}

func (sg *SummaryGenerator) calculateWordCount(chunks []*ChunkProcessingResult) int {
	totalWords := 0

	for _, chunk := range chunks {
		if wordCount, ok := chunk.ResultData["word_count"].(float64); ok {
			totalWords += int(wordCount)
		}
	}

	return totalWords
}

func (sg *SummaryGenerator) createSections(chunks []*ChunkProcessingResult) []SectionSummary {
	sections := make([]SectionSummary, 0, len(chunks))

	for i, chunk := range chunks {
		section := SectionSummary{
			Title:     fmt.Sprintf("Section %d", i+1),
			Summary:   sg.extractChunkSummary(chunk),
			Keywords:  sg.extractChunkKeywords(chunk),
			WordCount: sg.extractChunkWordCount(chunk),
			StartPage: i + 1,
			EndPage:   i + 1,
		}
		sections = append(sections, section)
	}

	return sections
}

func (sg *SummaryGenerator) generateSummaryText(chunks []*ChunkProcessingResult, keywords []string) string {
	if len(chunks) == 0 {
		return "No content available for summary."
	}

	summaryParts := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if summary := sg.extractChunkSummary(chunk); summary != "" {
			summaryParts = append(summaryParts, summary)
		}
	}

	if len(summaryParts) > 0 {
		fullSummary := strings.Join(summaryParts, " ")
		if len(fullSummary) > sg.config.SummaryLength {
			return fullSummary[:sg.config.SummaryLength] + "..."
		}
		return fullSummary
	}

	maxKeywords := 5
	if len(keywords) < maxKeywords {
		maxKeywords = len(keywords)
	}

	return fmt.Sprintf("Document contains %d sections with key topics: %s",
		len(chunks), strings.Join(keywords[:maxKeywords], ", "))
}

func (sg *SummaryGenerator) extractChunkSummary(chunk *ChunkProcessingResult) string {
	if summary, ok := chunk.ResultData["summary"].(string); ok {
		return summary
	}
	if content, ok := chunk.ResultData["content"].(string); ok {
		if len(content) > 200 {
			return content[:200] + "..."
		}
		return content
	}
	return ""
}

func (sg *SummaryGenerator) extractChunkKeywords(chunk *ChunkProcessingResult) []string {
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

func (sg *SummaryGenerator) extractChunkWordCount(chunk *ChunkProcessingResult) int {
	if wordCount, ok := chunk.ResultData["word_count"].(float64); ok {
		return int(wordCount)
	}
	return 0
}

func (sg *SummaryGenerator) calculateTotalSize(chunks []*ChunkProcessingResult) int64 {
	total := int64(0)
	for _, chunk := range chunks {
		total += chunk.ChunkInfo.Size
	}
	return total
}

func (sg *SummaryGenerator) extractFileID(chunkID string) string {
	parts := strings.Split(chunkID, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func (sg *SummaryGenerator) createResultMessage(result *SynthesisResult) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "synthesis_result",
		Target:    "synthesis_result",
		Payload:   result,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func main() {
	summaryGenerator := NewSummaryGenerator()
	agent.Run(summaryGenerator, "summary-generator")
}