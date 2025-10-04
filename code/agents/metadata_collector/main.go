package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

type MetadataCollector struct {
	agent.DefaultAgentRunner
	config *CollectorConfig
}

type CollectorConfig struct {
	IncludeChunkMetadata bool   `json:"include_chunk_metadata"`
	IncludeFileMetadata  bool   `json:"include_file_metadata"`
	GenerateSchema       bool   `json:"generate_schema"`
	OutputFormat         string `json:"output_format"`
	MaxMetadataSize      int64  `json:"max_metadata_size"`
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

type MetadataCollection struct {
	FileMetadata  map[string]interface{} `json:"file_metadata"`
	ChunkMetadata []ChunkMetadata        `json:"chunk_metadata"`
	Statistics    MetadataStatistics     `json:"statistics"`
	Schema        map[string]interface{} `json:"schema"`
	CreatedAt     time.Time              `json:"created_at"`
}

type ChunkMetadata struct {
	ChunkID   string                 `json:"chunk_id"`
	Type      string                 `json:"type"`
	Size      int64                  `json:"size"`
	Metadata  map[string]interface{} `json:"metadata"`
	Tags      []string               `json:"tags"`
	CreatedAt time.Time              `json:"created_at"`
}

type MetadataStatistics struct {
	TotalChunks  int                    `json:"total_chunks"`
	TypeCounts   map[string]int         `json:"type_counts"`
	TotalSize    int64                  `json:"total_size"`
	AverageSize  float64                `json:"average_size"`
	TagFrequency map[string]int         `json:"tag_frequency"`
	Properties   map[string]interface{} `json:"properties"`
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

func NewMetadataCollector() *MetadataCollector {
	return &MetadataCollector{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &CollectorConfig{
			IncludeChunkMetadata: true,
			IncludeFileMetadata:  true,
			GenerateSchema:       true,
			OutputFormat:         "json",
			MaxMetadataSize:      10485760,
		},
	}
}

func (mc *MetadataCollector) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
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

	if request.OutputType != "metadata_collection" {
		return nil, fmt.Errorf("unsupported output type: %s", request.OutputType)
	}

	mc.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &SynthesisResult{
		JobID:       fmt.Sprintf("metadata_%d", time.Now().UnixNano()),
		FileID:      request.FileID,
		OutputType:  request.OutputType,
		ProcessedAt: startTime,
		ChunkCount:  len(request.ChunkIDs),
		ResultData:  make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	chunks, err := mc.loadChunkResults(request.ChunkIDs, base)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load chunks: %v", err)
		result.Success = false
		return mc.createResultMessage(result), nil
	}

	collection, err := mc.collectMetadata(chunks)
	if err != nil {
		result.Error = fmt.Sprintf("failed to collect metadata: %v", err)
		result.Success = false
		return mc.createResultMessage(result), nil
	}

	result.ResultData["metadata_collection"] = collection
	result.Metadata["collector"] = "metadata_collection"
	result.Metadata["chunk_count"] = len(chunks)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return mc.createResultMessage(result), nil
}

func (mc *MetadataCollector) loadConfigFromAgent(base *agent.BaseAgent) {
	if outputFormat := base.GetConfigString("output_format", "json"); outputFormat != "" {
		mc.config.OutputFormat = outputFormat
	}
	if maxSize := base.GetConfigInt("max_metadata_size", 10485760); maxSize > 0 {
		mc.config.MaxMetadataSize = int64(maxSize)
	}
	mc.config.IncludeChunkMetadata = base.GetConfigBool("include_chunk_metadata", true)
	mc.config.IncludeFileMetadata = base.GetConfigBool("include_file_metadata", true)
	mc.config.GenerateSchema = base.GetConfigBool("generate_schema", true)
}

func (mc *MetadataCollector) loadChunkResults(chunkIDs []string, base *agent.BaseAgent) ([]*ChunkProcessingResult, error) {
	results := make([]*ChunkProcessingResult, 0, len(chunkIDs))

	for _, chunkID := range chunkIDs {
		chunkInfo := ChunkInfo{
			ChunkID:     chunkID,
			FileID:      mc.extractFileID(chunkID),
			SequenceNum: len(results),
			Size:        1024,
			MimeType:    "text/plain",
		}

		resultData := map[string]interface{}{
			"content":    fmt.Sprintf("Content for chunk %s", chunkID),
			"keywords":   []interface{}{"metadata", "collection", "processing"},
			"language":   "en",
			"word_count": float64(150),
			"sentiment":  0.5,
		}

		metadata := map[string]interface{}{
			"processing_agent":  "chunk_processor",
			"processing_time":   1500,
			"extraction_method": "nlp",
			"quality_score":     0.95,
			"content_type":      "text",
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

func (mc *MetadataCollector) collectMetadata(chunks []*ChunkProcessingResult) (*MetadataCollection, error) {
	collection := &MetadataCollection{
		FileMetadata:  make(map[string]interface{}),
		ChunkMetadata: make([]ChunkMetadata, 0, len(chunks)),
		Schema:        make(map[string]interface{}),
		CreatedAt:     time.Now(),
	}

	if mc.config.IncludeFileMetadata {
		mc.collectFileMetadata(chunks, collection)
	}

	if mc.config.IncludeChunkMetadata {
		mc.collectChunkMetadata(chunks, collection)
	}

	if mc.config.GenerateSchema {
		mc.generateSchema(collection)
	}

	mc.calculateStatistics(collection)

	return collection, nil
}

func (mc *MetadataCollector) collectFileMetadata(chunks []*ChunkProcessingResult, collection *MetadataCollection) {
	if len(chunks) == 0 {
		return
	}

	fileInfo := chunks[0].ChunkInfo
	collection.FileMetadata["file_id"] = fileInfo.FileID
	collection.FileMetadata["mime_type"] = fileInfo.MimeType
	collection.FileMetadata["chunk_count"] = len(chunks)
	collection.FileMetadata["total_size"] = mc.calculateTotalSize(chunks)
	collection.FileMetadata["created_at"] = time.Now()

	languages := make(map[string]int)
	contentTypes := make(map[string]int)

	for _, chunk := range chunks {
		if lang, ok := chunk.ResultData["language"].(string); ok {
			languages[lang]++
		}
		contentTypes[chunk.ChunkInfo.MimeType]++
	}

	collection.FileMetadata["languages"] = languages
	collection.FileMetadata["content_types"] = contentTypes

	if len(chunks) > 0 {
		collection.FileMetadata["first_processed"] = chunks[0].ProcessedAt
		collection.FileMetadata["last_processed"] = chunks[len(chunks)-1].ProcessedAt
	}
}

func (mc *MetadataCollector) collectChunkMetadata(chunks []*ChunkProcessingResult, collection *MetadataCollection) {
	for _, chunk := range chunks {
		metadata := ChunkMetadata{
			ChunkID:   chunk.ChunkID,
			Type:      chunk.ChunkInfo.MimeType,
			Size:      chunk.ChunkInfo.Size,
			Metadata:  make(map[string]interface{}),
			Tags:      mc.extractTags(chunk),
			CreatedAt: chunk.ProcessedAt,
		}

		for key, value := range chunk.Metadata {
			metadata.Metadata[key] = value
		}

		if keywords, ok := chunk.ResultData["keywords"].([]interface{}); ok {
			metadata.Metadata["keywords"] = keywords
		}

		if sentiment, ok := chunk.ResultData["sentiment"].(float64); ok {
			metadata.Metadata["sentiment"] = sentiment
		}

		if language, ok := chunk.ResultData["language"].(string); ok {
			metadata.Metadata["language"] = language
		}

		if wordCount, ok := chunk.ResultData["word_count"].(float64); ok {
			metadata.Metadata["word_count"] = wordCount
		}

		collection.ChunkMetadata = append(collection.ChunkMetadata, metadata)
	}
}

func (mc *MetadataCollector) extractTags(chunk *ChunkProcessingResult) []string {
	tags := []string{}

	if keywords, ok := chunk.ResultData["keywords"].([]interface{}); ok {
		for _, kw := range keywords {
			if keyword, ok := kw.(string); ok {
				tags = append(tags, keyword)
			}
		}
	}

	if chunk.ChunkInfo.MimeType != "text/plain" {
		tags = append(tags, chunk.ChunkInfo.MimeType)
	}

	if chunk.Success {
		tags = append(tags, "processed")
	} else {
		tags = append(tags, "failed")
	}

	return tags
}

func (mc *MetadataCollector) generateSchema(collection *MetadataCollection) {
	schema := make(map[string]interface{})

	fieldTypes := make(map[string]map[string]bool)

	for _, chunk := range collection.ChunkMetadata {
		for key, value := range chunk.Metadata {
			if fieldTypes[key] == nil {
				fieldTypes[key] = make(map[string]bool)
			}
			fieldTypes[key][fmt.Sprintf("%T", value)] = true
		}
	}

	for field, types := range fieldTypes {
		typeList := make([]string, 0, len(types))
		for typ := range types {
			typeList = append(typeList, typ)
		}
		schema[field] = typeList
	}

	schema["_metadata"] = map[string]interface{}{
		"generated_at":   time.Now(),
		"total_fields":   len(fieldTypes),
		"chunk_count":    len(collection.ChunkMetadata),
		"schema_version": "1.0",
	}

	collection.Schema = schema
}

func (mc *MetadataCollector) calculateStatistics(collection *MetadataCollection) {
	stats := MetadataStatistics{
		TotalChunks:  len(collection.ChunkMetadata),
		TypeCounts:   make(map[string]int),
		TagFrequency: make(map[string]int),
		Properties:   make(map[string]interface{}),
	}

	var totalSize int64
	qualityScores := make([]float64, 0)

	for _, chunk := range collection.ChunkMetadata {
		stats.TypeCounts[chunk.Type]++
		totalSize += chunk.Size

		for _, tag := range chunk.Tags {
			stats.TagFrequency[tag]++
		}

		if score, ok := chunk.Metadata["quality_score"].(float64); ok {
			qualityScores = append(qualityScores, score)
		}
	}

	stats.TotalSize = totalSize
	if stats.TotalChunks > 0 {
		stats.AverageSize = float64(totalSize) / float64(stats.TotalChunks)
	}

	if len(qualityScores) > 0 {
		var sum float64
		for _, score := range qualityScores {
			sum += score
		}
		stats.Properties["average_quality_score"] = sum / float64(len(qualityScores))
	}

	stats.Properties["unique_types"] = len(stats.TypeCounts)
	stats.Properties["unique_tags"] = len(stats.TagFrequency)

	collection.Statistics = stats
}

func (mc *MetadataCollector) calculateTotalSize(chunks []*ChunkProcessingResult) int64 {
	total := int64(0)
	for _, chunk := range chunks {
		total += chunk.ChunkInfo.Size
	}
	return total
}

func (mc *MetadataCollector) extractFileID(chunkID string) string {
	parts := strings.Split(chunkID, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func (mc *MetadataCollector) createResultMessage(result *SynthesisResult) *client.BrokerMessage {
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
	metadataCollector := NewMetadataCollector()
	agent.Run(metadataCollector, "metadata-collector")
}