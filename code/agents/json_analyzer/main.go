package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

type JSONAnalyzer struct {
	agent.DefaultAgentRunner
	config *AnalyzerConfig
}

type AnalyzerConfig struct {
	EnableValidation   bool `json:"enable_validation"`
	EnableSchemaGen    bool `json:"enable_schema_gen"`
	EnableKeyAnalysis  bool `json:"enable_key_analysis"`
	MaxDepth           int  `json:"max_depth"`
	MaxKeys            int  `json:"max_keys"`
	EnableMinification bool `json:"enable_minification"`
}

type ChunkProcessingRequest struct {
	RequestID   string                 `json:"request_id"`
	ChunkID     string                 `json:"chunk_id"`
	FileID      string                 `json:"file_id"`
	ChunkIndex  int                    `json:"chunk_index"`
	ContentType string                 `json:"content_type"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	Options     map[string]interface{} `json:"options"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ProcessingResult struct {
	RequestID      string                 `json:"request_id"`
	ChunkID        string                 `json:"chunk_id"`
	FileID         string                 `json:"file_id"`
	ChunkIndex     int                    `json:"chunk_index"`
	ProcessedBy    string                 `json:"processed_by"`
	ProcessingType string                 `json:"processing_type"`
	ResultData     map[string]interface{} `json:"result_data"`
	Metadata       map[string]interface{} `json:"metadata"`
	ProcessedAt    time.Time              `json:"processed_at"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
}

func NewJSONAnalyzer() *JSONAnalyzer {
	return &JSONAnalyzer{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &AnalyzerConfig{
			EnableValidation:   true,
			EnableSchemaGen:    true,
			EnableKeyAnalysis:  true,
			MaxDepth:           20,
			MaxKeys:            1000,
			EnableMinification: false,
		},
	}
}

func (ja *JSONAnalyzer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "chunk_processing_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var request ChunkProcessingRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	if request.ContentType != "json" && request.ContentType != "application/json" {
		return nil, fmt.Errorf("unsupported content type: %s", request.ContentType)
	}

	ja.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &ProcessingResult{
		RequestID:      request.RequestID,
		ChunkID:        request.ChunkID,
		FileID:         request.FileID,
		ChunkIndex:     request.ChunkIndex,
		ProcessedBy:    "json_analyzer",
		ProcessingType: "json_analysis",
		ResultData:     make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		ProcessedAt:    startTime,
	}

	processedJSON, analysisData, err := ja.analyzeJSON(request.Content)
	if err != nil {
		result.Error = fmt.Sprintf("JSON analysis failed: %v", err)
		result.Success = false
		result.ProcessingTime = time.Since(startTime)
		return ja.createResultMessage(result), nil
	}

	result.ResultData["processed_content"] = processedJSON
	result.ResultData = mergeMapStringInterface(result.ResultData, analysisData)
	result.Metadata["analyzer"] = "json_analysis"
	result.Metadata["original_length"] = len(request.Content)
	result.Metadata["processed_length"] = len(processedJSON)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return ja.createResultMessage(result), nil
}

func (ja *JSONAnalyzer) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxDepth := base.GetConfigInt("max_depth", 20); maxDepth > 0 {
		ja.config.MaxDepth = maxDepth
	}
	if maxKeys := base.GetConfigInt("max_keys", 1000); maxKeys > 0 {
		ja.config.MaxKeys = maxKeys
	}
	ja.config.EnableValidation = base.GetConfigBool("enable_validation", true)
	ja.config.EnableSchemaGen = base.GetConfigBool("enable_schema_gen", true)
	ja.config.EnableKeyAnalysis = base.GetConfigBool("enable_key_analysis", true)
	ja.config.EnableMinification = base.GetConfigBool("enable_minification", false)
}

func (ja *JSONAnalyzer) analyzeJSON(content string) (string, map[string]interface{}, error) {
	analysisData := make(map[string]interface{})

	// Validate JSON
	var jsonData interface{}
	err := json.Unmarshal([]byte(content), &jsonData)
	if err != nil {
		analysisData["validation_error"] = err.Error()
		analysisData["is_valid"] = false
		return content, analysisData, nil
	}

	analysisData["is_valid"] = true

	// Basic JSON statistics
	analysisData["original_size"] = len(content)

	// Minify if enabled
	processedContent := content
	if ja.config.EnableMinification {
		minified, err := json.Marshal(jsonData)
		if err == nil {
			processedContent = string(minified)
			analysisData["minified_size"] = len(processedContent)
			analysisData["compression_ratio"] = float64(len(content)) / float64(len(processedContent))
		}
	}

	// Analyze structure
	if ja.config.EnableKeyAnalysis {
		keyStats := ja.analyzeKeys(jsonData, 0)
		analysisData["key_statistics"] = keyStats
	}

	// Generate schema if enabled
	if ja.config.EnableSchemaGen {
		schema := ja.generateSchema(jsonData)
		analysisData["schema"] = schema
	}

	// Detect JSON type and complexity
	jsonType, complexity := ja.detectJSONTypeAndComplexity(jsonData)
	analysisData["json_type"] = jsonType
	analysisData["complexity"] = complexity

	// Extract value types and counts
	valueTypes := ja.analyzeValueTypes(jsonData)
	analysisData["value_types"] = valueTypes

	// Detect patterns
	patterns := ja.detectPatterns(jsonData)
	analysisData["patterns"] = patterns

	// Processing metadata
	analysisData["processed_by"] = "json_analyzer"
	analysisData["processed_at"] = time.Now().Format(time.RFC3339)
	analysisData["processing_type"] = "json_validation_and_analysis"

	return processedContent, analysisData, nil
}

func (ja *JSONAnalyzer) analyzeKeys(data interface{}, depth int) map[string]interface{} {
	stats := map[string]interface{}{
		"total_keys":  0,
		"max_depth":   depth,
		"key_lengths": make(map[string]int),
		"duplicates":  make([]string, 0),
	}

	if depth > ja.config.MaxDepth {
		return stats
	}

	keysSeen := make(map[string]bool)
	ja.collectKeys(data, depth, keysSeen, stats)

	stats["total_keys"] = len(keysSeen)
	return stats
}

func (ja *JSONAnalyzer) collectKeys(data interface{}, depth int, keysSeen map[string]bool, stats map[string]interface{}) {
	if depth > ja.config.MaxDepth {
		return
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if currentDepth, ok := stats["max_depth"].(int); !ok || depth > currentDepth {
			stats["max_depth"] = depth
		}

		for key, value := range v {
			if keysSeen[key] {
				if duplicates, ok := stats["duplicates"].([]string); ok {
					found := false
					for _, dup := range duplicates {
						if dup == key {
							found = true
							break
						}
					}
					if !found {
						stats["duplicates"] = append(duplicates, key)
					}
				}
			}
			keysSeen[key] = true

			if keyLengths, ok := stats["key_lengths"].(map[string]int); ok {
				keyLengths[key] = len(key)
			}

			ja.collectKeys(value, depth+1, keysSeen, stats)
		}
	case []interface{}:
		for _, item := range v {
			ja.collectKeys(item, depth+1, keysSeen, stats)
		}
	}
}

func (ja *JSONAnalyzer) generateSchema(data interface{}) map[string]interface{} {
	return ja.generateSchemaForValue(data)
}

func (ja *JSONAnalyzer) generateSchemaForValue(value interface{}) map[string]interface{} {
	schema := make(map[string]interface{})

	switch v := value.(type) {
	case nil:
		schema["type"] = "null"
	case bool:
		schema["type"] = "boolean"
	case float64:
		schema["type"] = "number"
	case string:
		schema["type"] = "string"
		schema["minLength"] = len(v)
		schema["maxLength"] = len(v)
		if ja.isEmail(v) {
			schema["format"] = "email"
		} else if ja.isURL(v) {
			schema["format"] = "uri"
		} else if ja.isDate(v) {
			schema["format"] = "date-time"
		}
	case []interface{}:
		schema["type"] = "array"
		if len(v) > 0 {
			schema["items"] = ja.generateSchemaForValue(v[0])
		}
		schema["minItems"] = len(v)
		schema["maxItems"] = len(v)
	case map[string]interface{}:
		schema["type"] = "object"
		properties := make(map[string]interface{})
		required := make([]string, 0)
		for key, val := range v {
			properties[key] = ja.generateSchemaForValue(val)
			required = append(required, key)
		}
		schema["properties"] = properties
		schema["required"] = required
	default:
		schema["type"] = reflect.TypeOf(value).String()
	}

	return schema
}

func (ja *JSONAnalyzer) detectJSONTypeAndComplexity(data interface{}) (string, string) {
	switch v := data.(type) {
	case map[string]interface{}:
		keyCount := len(v)
		depth := ja.calculateDepth(data, 0)

		if keyCount <= 5 && depth <= 2 {
			return "object", "simple"
		} else if keyCount <= 20 && depth <= 4 {
			return "object", "moderate"
		} else {
			return "object", "complex"
		}
	case []interface{}:
		arrayLen := len(v)
		if arrayLen == 0 {
			return "array", "empty"
		}

		depth := ja.calculateDepth(data, 0)
		if arrayLen <= 10 && depth <= 2 {
			return "array", "simple"
		} else if arrayLen <= 100 && depth <= 4 {
			return "array", "moderate"
		} else {
			return "array", "complex"
		}
	case string, float64, bool:
		return "primitive", "simple"
	case nil:
		return "null", "simple"
	default:
		return "unknown", "unknown"
	}
}

func (ja *JSONAnalyzer) calculateDepth(data interface{}, currentDepth int) int {
	maxDepth := currentDepth

	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			depth := ja.calculateDepth(value, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	case []interface{}:
		for _, item := range v {
			depth := ja.calculateDepth(item, currentDepth+1)
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth
}

func (ja *JSONAnalyzer) analyzeValueTypes(data interface{}) map[string]int {
	types := make(map[string]int)
	ja.countValueTypes(data, types)
	return types
}

func (ja *JSONAnalyzer) countValueTypes(data interface{}, types map[string]int) {
	switch v := data.(type) {
	case nil:
		types["null"]++
	case bool:
		types["boolean"]++
	case float64:
		types["number"]++
	case string:
		types["string"]++
	case []interface{}:
		types["array"]++
		for _, item := range v {
			ja.countValueTypes(item, types)
		}
	case map[string]interface{}:
		types["object"]++
		for _, value := range v {
			ja.countValueTypes(value, types)
		}
	default:
		types["unknown"]++
	}
}

func (ja *JSONAnalyzer) detectPatterns(data interface{}) map[string]interface{} {
	patterns := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		// Detect common object patterns
		if ja.looksLikeAPIResponse(v) {
			patterns["api_response"] = true
		}
		if ja.looksLikeConfig(v) {
			patterns["configuration"] = true
		}
		if ja.looksLikeMetadata(v) {
			patterns["metadata"] = true
		}
	case []interface{}:
		// Detect array patterns
		if ja.looksLikeDataset(v) {
			patterns["dataset"] = true
		}
		if ja.looksLikeTimeSeries(v) {
			patterns["time_series"] = true
		}
	}

	return patterns
}

func (ja *JSONAnalyzer) looksLikeAPIResponse(obj map[string]interface{}) bool {
	commonAPIKeys := []string{"status", "data", "message", "error", "code", "result"}
	matches := 0
	for _, key := range commonAPIKeys {
		if _, exists := obj[key]; exists {
			matches++
		}
	}
	return matches >= 2
}

func (ja *JSONAnalyzer) looksLikeConfig(obj map[string]interface{}) bool {
	commonConfigKeys := []string{"settings", "config", "options", "preferences", "parameters"}
	for _, key := range commonConfigKeys {
		if _, exists := obj[key]; exists {
			return true
		}
	}
	return false
}

func (ja *JSONAnalyzer) looksLikeMetadata(obj map[string]interface{}) bool {
	commonMetaKeys := []string{"created_at", "updated_at", "id", "name", "type", "version"}
	matches := 0
	for _, key := range commonMetaKeys {
		if _, exists := obj[key]; exists {
			matches++
		}
	}
	return matches >= 3
}

func (ja *JSONAnalyzer) looksLikeDataset(arr []interface{}) bool {
	if len(arr) < 2 {
		return false
	}

	// Check if all elements have similar structure
	firstType := reflect.TypeOf(arr[0])
	for _, item := range arr[1:] {
		if reflect.TypeOf(item) != firstType {
			return false
		}
	}
	return true
}

func (ja *JSONAnalyzer) looksLikeTimeSeries(arr []interface{}) bool {
	if len(arr) < 2 {
		return false
	}

	timestampFound := 0
	for _, item := range arr {
		if obj, ok := item.(map[string]interface{}); ok {
			for key := range obj {
				if strings.Contains(strings.ToLower(key), "time") ||
				   strings.Contains(strings.ToLower(key), "date") ||
				   strings.Contains(strings.ToLower(key), "timestamp") {
					timestampFound++
					break
				}
			}
		}
	}
	return timestampFound >= len(arr)/2
}

func (ja *JSONAnalyzer) isEmail(s string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(s)
}

func (ja *JSONAnalyzer) isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "ftp://")
}

func (ja *JSONAnalyzer) isDate(s string) bool {
	dateFormats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"01-02-2006",
	}

	for _, format := range dateFormats {
		if _, err := time.Parse(format, s); err == nil {
			return true
		}
	}
	return false
}

func (ja *JSONAnalyzer) createResultMessage(result *ProcessingResult) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "chunk_processing_result",
		Target:    "chunk_processing_result",
		Payload:   result,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func mergeMapStringInterface(dst, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func main() {
	jsonAnalyzer := NewJSONAnalyzer()
	agent.Run(jsonAnalyzer, "json-analyzer")
}