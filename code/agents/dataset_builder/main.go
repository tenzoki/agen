package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

type DatasetBuilder struct {
	agent.DefaultAgentRunner
	config *DatasetConfig
}

type DatasetConfig struct {
	OutputFormat      string   `json:"output_format"`
	IncludeMetadata   bool     `json:"include_metadata"`
	GenerateSchema    bool     `json:"generate_schema"`
	NamingScheme      string   `json:"naming_scheme"`
	RequiredFields    []string `json:"required_fields"`
	OptionalFields    []string `json:"optional_fields"`
	MaxRecords        int      `json:"max_records"`
	EnableValidation  bool     `json:"enable_validation"`
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

type Dataset struct {
	DatasetID   string                   `json:"dataset_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Schema      DatasetSchema            `json:"schema"`
	Records     []map[string]interface{} `json:"records"`
	Metadata    map[string]interface{}   `json:"metadata"`
	Statistics  DatasetStatistics        `json:"statistics"`
	CreatedAt   time.Time                `json:"created_at"`
}

type DatasetSchema struct {
	Fields      []FieldDefinition      `json:"fields"`
	Constraints map[string]interface{} `json:"constraints"`
	Indexes     []string               `json:"indexes"`
}

type FieldDefinition struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Constraints map[string]interface{} `json:"constraints"`
}

type DatasetStatistics struct {
	RecordCount   int                       `json:"record_count"`
	FieldCounts   map[string]int            `json:"field_counts"`
	TypeCounts    map[string]int            `json:"type_counts"`
	ValueCounts   map[string]map[string]int `json:"value_counts"`
	Distributions map[string]interface{}    `json:"distributions"`
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

func NewDatasetBuilder() *DatasetBuilder {
	return &DatasetBuilder{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &DatasetConfig{
			OutputFormat:     "json",
			IncludeMetadata:  true,
			GenerateSchema:   true,
			NamingScheme:     "chunk_XXXX",
			RequiredFields:   []string{"chunk_id", "sequence_num", "size", "mime_type"},
			OptionalFields:   []string{"keywords", "sentiment", "language", "processed_at"},
			MaxRecords:       100000,
			EnableValidation: true,
		},
	}
}

func (db *DatasetBuilder) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
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

	if request.OutputType != "dataset" {
		return nil, fmt.Errorf("unsupported output type: %s", request.OutputType)
	}

	db.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &SynthesisResult{
		JobID:       fmt.Sprintf("dataset_%d", time.Now().UnixNano()),
		FileID:      request.FileID,
		OutputType:  request.OutputType,
		ProcessedAt: startTime,
		ChunkCount:  len(request.ChunkIDs),
		ResultData:  make(map[string]interface{}),
		Metadata:    make(map[string]interface{}),
	}

	chunks, err := db.loadChunkResults(request.ChunkIDs, base)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load chunks: %v", err)
		result.Success = false
		return db.createResultMessage(result), nil
	}

	dataset, err := db.buildDataset(chunks)
	if err != nil {
		result.Error = fmt.Sprintf("failed to build dataset: %v", err)
		result.Success = false
		return db.createResultMessage(result), nil
	}

	result.ResultData["dataset"] = dataset
	result.Metadata["builder"] = "dataset"
	result.Metadata["dataset_id"] = dataset.DatasetID
	result.Metadata["record_count"] = len(dataset.Records)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return db.createResultMessage(result), nil
}

func (db *DatasetBuilder) loadConfigFromAgent(base *agent.BaseAgent) {
	if outputFormat := base.GetConfigString("output_format", "json"); outputFormat != "" {
		db.config.OutputFormat = outputFormat
	}
	if namingScheme := base.GetConfigString("naming_scheme", "chunk_XXXX"); namingScheme != "" {
		db.config.NamingScheme = namingScheme
	}
	if maxRecords := base.GetConfigInt("max_records", 100000); maxRecords > 0 {
		db.config.MaxRecords = maxRecords
	}
	db.config.IncludeMetadata = base.GetConfigBool("include_metadata", true)
	db.config.GenerateSchema = base.GetConfigBool("generate_schema", true)
	db.config.EnableValidation = base.GetConfigBool("enable_validation", true)
}

func (db *DatasetBuilder) loadChunkResults(chunkIDs []string, base *agent.BaseAgent) ([]*ChunkProcessingResult, error) {
	results := make([]*ChunkProcessingResult, 0, len(chunkIDs))

	for i, chunkID := range chunkIDs {
		chunkInfo := ChunkInfo{
			ChunkID:     chunkID,
			FileID:      db.extractFileID(chunkID),
			SequenceNum: i,
			Size:        int64(1024 + i*100),
			MimeType:    "text/plain",
		}

		resultData := map[string]interface{}{
			"content":    fmt.Sprintf("Dataset content for chunk %s", chunkID),
			"keywords":   []interface{}{"dataset", "build", "structure", "export"},
			"topics":     []interface{}{"data", "processing"},
			"language":   "en",
			"word_count": float64(175 + i*10),
			"sentiment":  0.6 + float64(i)*0.05,
			"quality":    0.85 + float64(i)*0.01,
		}

		metadata := map[string]interface{}{
			"processing_agent":   "chunk_processor",
			"extraction_method":  "automated",
			"validation_status":  "passed",
			"data_source":        "document_analysis",
			"record_type":        "structured",
		}

		result := &ChunkProcessingResult{
			ChunkID:     chunkID,
			ChunkInfo:   chunkInfo,
			ProcessedBy: "chunk_processor",
			ResultData:  resultData,
			Metadata:    metadata,
			ProcessedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			Success:     true,
		}

		results = append(results, result)
	}

	return results, nil
}

func (db *DatasetBuilder) buildDataset(chunks []*ChunkProcessingResult) (*Dataset, error) {
	dataset := &Dataset{
		DatasetID:   fmt.Sprintf("dataset_%d", time.Now().Unix()),
		Name:        fmt.Sprintf("Chunk Analysis Dataset %s", db.getFileID(chunks)[:8]),
		Description: "Dataset generated from chunk processing results with structured records and metadata",
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
	}

	if db.config.GenerateSchema {
		dataset.Schema = db.generateSchema(chunks)
	}

	dataset.Records = db.generateRecords(chunks)

	if db.config.EnableValidation {
		if err := db.validateDataset(dataset); err != nil {
			return nil, fmt.Errorf("dataset validation failed: %w", err)
		}
	}

	db.calculateStatistics(dataset)

	return dataset, nil
}

func (db *DatasetBuilder) getFileID(chunks []*ChunkProcessingResult) string {
	if len(chunks) > 0 {
		return chunks[0].ChunkInfo.FileID
	}
	return "unknown"
}

func (db *DatasetBuilder) generateSchema(chunks []*ChunkProcessingResult) DatasetSchema {
	fields := []FieldDefinition{
		{
			Name:        "chunk_id",
			Type:        "string",
			Description: "Unique chunk identifier",
			Required:    true,
			Constraints: map[string]interface{}{"unique": true},
		},
		{
			Name:        "sequence_num",
			Type:        "integer",
			Description: "Chunk sequence number in document",
			Required:    true,
			Constraints: map[string]interface{}{"min": 0},
		},
		{
			Name:        "file_id",
			Type:        "string",
			Description: "Source file identifier",
			Required:    true,
			Constraints: map[string]interface{}{},
		},
		{
			Name:        "size",
			Type:        "integer",
			Description: "Chunk size in bytes",
			Required:    true,
			Constraints: map[string]interface{}{"min": 0},
		},
		{
			Name:        "mime_type",
			Type:        "string",
			Description: "Content MIME type",
			Required:    true,
			Constraints: map[string]interface{}{},
		},
		{
			Name:        "keywords",
			Type:        "array",
			Description: "Extracted keywords from content",
			Required:    false,
			Constraints: map[string]interface{}{"item_type": "string"},
		},
		{
			Name:        "language",
			Type:        "string",
			Description: "Detected content language",
			Required:    false,
			Constraints: map[string]interface{}{"pattern": "^[a-z]{2}$"},
		},
		{
			Name:        "sentiment",
			Type:        "number",
			Description: "Content sentiment score",
			Required:    false,
			Constraints: map[string]interface{}{"min": -1.0, "max": 1.0},
		},
		{
			Name:        "word_count",
			Type:        "integer",
			Description: "Number of words in content",
			Required:    false,
			Constraints: map[string]interface{}{"min": 0},
		},
		{
			Name:        "processed_at",
			Type:        "datetime",
			Description: "Processing timestamp",
			Required:    true,
			Constraints: map[string]interface{}{},
		},
	}

	return DatasetSchema{
		Fields:      fields,
		Constraints: map[string]interface{}{"unique": []string{"chunk_id"}},
		Indexes:     []string{"chunk_id", "sequence_num", "file_id", "processed_at"},
	}
}

func (db *DatasetBuilder) generateRecords(chunks []*ChunkProcessingResult) []map[string]interface{} {
	records := make([]map[string]interface{}, 0, len(chunks))

	maxRecords := db.config.MaxRecords
	if len(chunks) < maxRecords {
		maxRecords = len(chunks)
	}

	for i := 0; i < maxRecords; i++ {
		chunk := chunks[i]
		record := map[string]interface{}{
			"chunk_id":     chunk.ChunkID,
			"sequence_num": chunk.ChunkInfo.SequenceNum,
			"file_id":      chunk.ChunkInfo.FileID,
			"size":         chunk.ChunkInfo.Size,
			"mime_type":    chunk.ChunkInfo.MimeType,
			"processed_at": chunk.ProcessedAt,
		}

		if keywords := db.extractKeywords(chunk); len(keywords) > 0 {
			record["keywords"] = keywords
		}

		if sentiment, ok := chunk.ResultData["sentiment"].(float64); ok {
			record["sentiment"] = sentiment
		}

		if language, ok := chunk.ResultData["language"].(string); ok {
			record["language"] = language
		}

		if wordCount, ok := chunk.ResultData["word_count"].(float64); ok {
			record["word_count"] = int(wordCount)
		}

		if quality, ok := chunk.ResultData["quality"].(float64); ok {
			record["quality"] = quality
		}

		if db.config.IncludeMetadata {
			record["metadata"] = chunk.Metadata
		}

		records = append(records, record)
	}

	return records
}

func (db *DatasetBuilder) extractKeywords(chunk *ChunkProcessingResult) []string {
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

func (db *DatasetBuilder) validateDataset(dataset *Dataset) error {
	if dataset.DatasetID == "" {
		return fmt.Errorf("dataset ID is required")
	}

	if len(dataset.Records) == 0 {
		return fmt.Errorf("dataset must contain at least one record")
	}

	requiredFields := map[string]bool{}
	for _, field := range dataset.Schema.Fields {
		if field.Required {
			requiredFields[field.Name] = true
		}
	}

	for i, record := range dataset.Records {
		for field := range requiredFields {
			if _, exists := record[field]; !exists {
				return fmt.Errorf("record %d missing required field: %s", i, field)
			}
		}
	}

	return nil
}

func (db *DatasetBuilder) calculateStatistics(dataset *Dataset) {
	stats := DatasetStatistics{
		RecordCount:   len(dataset.Records),
		FieldCounts:   make(map[string]int),
		TypeCounts:    make(map[string]int),
		ValueCounts:   make(map[string]map[string]int),
		Distributions: make(map[string]interface{}),
	}

	for _, record := range dataset.Records {
		for field, value := range record {
			stats.FieldCounts[field]++
			stats.TypeCounts[fmt.Sprintf("%T", value)]++

			if field == "mime_type" || field == "language" {
				if stats.ValueCounts[field] == nil {
					stats.ValueCounts[field] = make(map[string]int)
				}
				if strVal, ok := value.(string); ok {
					stats.ValueCounts[field][strVal]++
				}
			}
		}
	}

	sentiments := []float64{}
	wordCounts := []int{}
	sizes := []int64{}

	for _, record := range dataset.Records {
		if sentiment, ok := record["sentiment"].(float64); ok {
			sentiments = append(sentiments, sentiment)
		}
		if wordCount, ok := record["word_count"].(int); ok {
			wordCounts = append(wordCounts, wordCount)
		}
		if size, ok := record["size"].(int64); ok {
			sizes = append(sizes, size)
		}
	}

	if len(sentiments) > 0 {
		stats.Distributions["sentiment_avg"] = db.calculateAverage(sentiments)
	}
	if len(wordCounts) > 0 {
		stats.Distributions["word_count_avg"] = db.calculateAverageInt(wordCounts)
	}
	if len(sizes) > 0 {
		stats.Distributions["size_avg"] = db.calculateAverageInt64(sizes)
	}

	stats.Distributions["unique_files"] = len(stats.ValueCounts["mime_type"])
	stats.Distributions["field_coverage"] = float64(len(stats.FieldCounts)) / float64(len(dataset.Schema.Fields))

	dataset.Statistics = stats
}

func (db *DatasetBuilder) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (db *DatasetBuilder) calculateAverageInt(values []int) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

func (db *DatasetBuilder) calculateAverageInt64(values []int64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := int64(0)
	for _, v := range values {
		sum += v
	}
	return float64(sum) / float64(len(values))
}

func (db *DatasetBuilder) extractFileID(chunkID string) string {
	parts := strings.Split(chunkID, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return "unknown"
}

func (db *DatasetBuilder) createResultMessage(result *SynthesisResult) *client.BrokerMessage {
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
	datasetBuilder := NewDatasetBuilder()
	agent.Run(datasetBuilder, "dataset-builder")
}