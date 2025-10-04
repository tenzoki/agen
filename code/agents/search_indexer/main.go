package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

type SearchIndexer struct {
	agent.DefaultAgentRunner
	config *IndexerConfig
}

type IndexerConfig struct {
	MaxTerms           int     `json:"max_terms"`
	MinTermFrequency   int     `json:"min_term_frequency"`
	CalculatePositions bool    `json:"calculate_positions"`
	IndexFormat        string  `json:"index_format"`
	ScoreThreshold     float64 `json:"score_threshold"`
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

type SearchIndex struct {
	IndexID    string                   `json:"index_id"`
	Terms      map[string][]IndexEntry  `json:"terms"`
	Documents  map[string]DocumentIndex `json:"documents"`
	Metadata   map[string]interface{}   `json:"metadata"`
	CreatedAt  time.Time                `json:"created_at"`
	Statistics IndexStatistics          `json:"statistics"`
}

type IndexEntry struct {
	ChunkID   string  `json:"chunk_id"`
	Frequency int     `json:"frequency"`
	Score     float64 `json:"score"`
	Position  []int   `json:"position"`
}

type DocumentIndex struct {
	FileID     string                 `json:"file_id"`
	Title      string                 `json:"title"`
	Summary    string                 `json:"summary"`
	Keywords   []string               `json:"keywords"`
	ChunkCount int                    `json:"chunk_count"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type IndexStatistics struct {
	TotalTerms     int `json:"total_terms"`
	UniqueTerms    int `json:"unique_terms"`
	TotalDocuments int `json:"total_documents"`
	TotalChunks    int `json:"total_chunks"`
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

func NewSearchIndexer() *SearchIndexer {
	return &SearchIndexer{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &IndexerConfig{
			MaxTerms:           10000,
			MinTermFrequency:   1,
			CalculatePositions: true,
			IndexFormat:        "json",
			ScoreThreshold:     0.1,
		},
	}
}

func (si *SearchIndexer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
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

	if request.OutputType != "search_index" {
		return nil, fmt.Errorf("unsupported output type: %s", request.OutputType)
	}

	si.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &SynthesisResult{
		JobID:       fmt.Sprintf("index_%d", time.Now().UnixNano()),
		FileID:      request.FileID,
		OutputType:  request.OutputType,
		ProcessedAt: startTime,
		ChunkCount:  len(request.ChunkIDs),
		ResultData:  make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	chunks, err := si.loadChunkResults(request.ChunkIDs, base)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load chunks: %v", err)
		result.Success = false
		return si.createResultMessage(result), nil
	}

	index, err := si.buildSearchIndex(chunks)
	if err != nil {
		result.Error = fmt.Sprintf("failed to build index: %v", err)
		result.Success = false
		return si.createResultMessage(result), nil
	}

	result.ResultData["search_index"] = index
	result.Metadata["indexer"] = "search_index"
	result.Metadata["index_id"] = index.IndexID
	result.Metadata["term_count"] = len(index.Terms)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return si.createResultMessage(result), nil
}

func (si *SearchIndexer) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxTerms := base.GetConfigInt("max_terms", 10000); maxTerms > 0 {
		si.config.MaxTerms = maxTerms
	}
	if minFreq := base.GetConfigInt("min_term_frequency", 1); minFreq > 0 {
		si.config.MinTermFrequency = minFreq
	}
	if indexFormat := base.GetConfigString("index_format", "json"); indexFormat != "" {
		si.config.IndexFormat = indexFormat
	}
	si.config.CalculatePositions = base.GetConfigBool("calculate_positions", true)
}

func (si *SearchIndexer) loadChunkResults(chunkIDs []string, base *agent.BaseAgent) ([]*ChunkProcessingResult, error) {
	results := make([]*ChunkProcessingResult, 0, len(chunkIDs))

	for _, chunkID := range chunkIDs {
		chunkInfo := ChunkInfo{
			ChunkID:     chunkID,
			FileID:      si.extractFileID(chunkID),
			SequenceNum: len(results),
			Size:        1024,
			MimeType:    "text/plain",
		}

		resultData := map[string]interface{}{
			"content":    fmt.Sprintf("Content for chunk %s with searchable terms", chunkID),
			"keywords":   []interface{}{"search", "index", "term", "document"},
			"topics":     []interface{}{"indexing", "search"},
			"language":   "en",
			"word_count": float64(100),
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

	return results, nil
}

func (si *SearchIndexer) buildSearchIndex(chunks []*ChunkProcessingResult) (*SearchIndex, error) {
	index := &SearchIndex{
		IndexID:   fmt.Sprintf("idx_%d", time.Now().Unix()),
		Terms:     make(map[string][]IndexEntry),
		Documents: make(map[string]DocumentIndex),
		Metadata:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		Statistics: IndexStatistics{},
	}

	si.buildTermIndex(chunks, index)
	si.buildDocumentIndex(chunks, index)
	si.calculateStatistics(index)

	return index, nil
}

func (si *SearchIndexer) buildTermIndex(chunks []*ChunkProcessingResult, index *SearchIndex) {
	termFreq := make(map[string]map[string]int)

	for _, chunk := range chunks {
		chunkTerms := make(map[string]int)

		if keywords, ok := chunk.ResultData["keywords"].([]interface{}); ok {
			for _, kw := range keywords {
				if keyword, ok := kw.(string); ok {
					keyword = strings.ToLower(strings.TrimSpace(keyword))
					if len(keyword) > 0 {
						chunkTerms[keyword]++
					}
				}
			}
		}

		if content, ok := chunk.ResultData["content"].(string); ok {
			words := strings.Fields(strings.ToLower(content))
			for _, word := range words {
				word = strings.Trim(word, ".,!?;:")
				if len(word) > 2 {
					chunkTerms[word]++
				}
			}
		}

		for term, freq := range chunkTerms {
			if termFreq[term] == nil {
				termFreq[term] = make(map[string]int)
			}
			termFreq[term][chunk.ChunkID] = freq
		}
	}

	for term, chunkFreqs := range termFreq {
		if len(chunkFreqs) < si.config.MinTermFrequency {
			continue
		}

		entries := make([]IndexEntry, 0, len(chunkFreqs))
		for chunkID, freq := range chunkFreqs {
			score := si.calculateTermScore(freq, len(chunkFreqs), len(chunks))
			if score >= si.config.ScoreThreshold {
				entry := IndexEntry{
					ChunkID:   chunkID,
					Frequency: freq,
					Score:     score,
					Position:  []int{0},
				}
				entries = append(entries, entry)
			}
		}

		if len(entries) > 0 {
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Score > entries[j].Score
			})
			index.Terms[term] = entries
		}
	}
}

func (si *SearchIndexer) calculateTermScore(freq, docCount, totalDocs int) float64 {
	tf := float64(freq)
	idf := 1.0
	if docCount > 0 {
		idf = 1.0 + (float64(totalDocs) / float64(docCount))
	}
	return tf * idf
}

func (si *SearchIndexer) buildDocumentIndex(chunks []*ChunkProcessingResult, index *SearchIndex) {
	if len(chunks) == 0 {
		return
	}

	fileID := chunks[0].ChunkInfo.FileID

	docIndex := DocumentIndex{
		FileID:     fileID,
		Title:      si.extractDocumentTitle(chunks),
		Summary:    si.extractDocumentSummary(chunks),
		Keywords:   si.extractDocumentKeywords(chunks),
		ChunkCount: len(chunks),
		Metadata:   make(map[string]interface{}),
	}

	index.Documents[fileID] = docIndex
}

func (si *SearchIndexer) extractDocumentTitle(chunks []*ChunkProcessingResult) string {
	if len(chunks) > 0 {
		return fmt.Sprintf("Document %s", chunks[0].ChunkInfo.FileID[:8])
	}
	return "Unknown Document"
}

func (si *SearchIndexer) extractDocumentSummary(chunks []*ChunkProcessingResult) string {
	summaries := make([]string, 0, min(3, len(chunks)))
	for i, chunk := range chunks {
		if i >= 3 {
			break
		}
		if summary, ok := chunk.ResultData["summary"].(string); ok && summary != "" {
			summaries = append(summaries, summary)
		} else if content, ok := chunk.ResultData["content"].(string); ok && content != "" {
			if len(content) > 100 {
				summaries = append(summaries, content[:100]+"...")
			} else {
				summaries = append(summaries, content)
			}
		}
	}
	return strings.Join(summaries, " ")
}

func (si *SearchIndexer) extractDocumentKeywords(chunks []*ChunkProcessingResult) []string {
	keywordMap := make(map[string]bool)
	for _, chunk := range chunks {
		if keywords, ok := chunk.ResultData["keywords"].([]interface{}); ok {
			for _, kw := range keywords {
				if keyword, ok := kw.(string); ok {
					keywordMap[keyword] = true
				}
			}
		}
	}

	result := make([]string, 0, len(keywordMap))
	for keyword := range keywordMap {
		result = append(result, keyword)
	}
	return result
}

func (si *SearchIndexer) calculateStatistics(index *SearchIndex) {
	index.Statistics.TotalTerms = len(index.Terms)
	index.Statistics.TotalDocuments = len(index.Documents)

	uniqueTerms := make(map[string]bool)
	totalChunks := 0

	for term, entries := range index.Terms {
		uniqueTerms[term] = true
		totalChunks += len(entries)
	}

	index.Statistics.UniqueTerms = len(uniqueTerms)
	index.Statistics.TotalChunks = totalChunks
}

func (si *SearchIndexer) extractFileID(chunkID string) string {
	parts := strings.Split(chunkID, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func (si *SearchIndexer) createResultMessage(result *SynthesisResult) *client.BrokerMessage {
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
	searchIndexer := NewSearchIndexer()
	agent.Run(searchIndexer, "search-indexer")
}