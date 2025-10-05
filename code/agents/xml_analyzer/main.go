package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

type XMLAnalyzer struct {
	agent.DefaultAgentRunner
	config *AnalyzerConfig
}

type AnalyzerConfig struct {
	EnableValidation   bool `json:"enable_validation"`
	EnableSchemaGen    bool `json:"enable_schema_gen"`
	EnableNamespaceAnalysis bool `json:"enable_namespace_analysis"`
	MaxDepth           int  `json:"max_depth"`
	MaxElements        int  `json:"max_elements"`
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

type XMLElement struct {
	Name       string                 `json:"name"`
	Attributes map[string]string      `json:"attributes,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Children   []XMLElement           `json:"children,omitempty"`
	Namespace  string                 `json:"namespace,omitempty"`
}

func NewXMLAnalyzer() *XMLAnalyzer {
	return &XMLAnalyzer{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		config: &AnalyzerConfig{
			EnableValidation:        true,
			EnableSchemaGen:         true,
			EnableNamespaceAnalysis: true,
			MaxDepth:                20,
			MaxElements:             1000,
			EnableMinification:      false,
		},
	}
}

func (xa *XMLAnalyzer) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
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

	if request.ContentType != "xml" && request.ContentType != "application/xml" && request.ContentType != "text/xml" {
		return nil, fmt.Errorf("unsupported content type: %s", request.ContentType)
	}

	xa.loadConfigFromAgent(base)

	startTime := time.Now()
	result := &ProcessingResult{
		RequestID:      request.RequestID,
		ChunkID:        request.ChunkID,
		FileID:         request.FileID,
		ChunkIndex:     request.ChunkIndex,
		ProcessedBy:    "xml_analyzer",
		ProcessingType: "xml_analysis",
		ResultData:     make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
		ProcessedAt:    startTime,
	}

	processedXML, analysisData, err := xa.analyzeXML(request.Content)
	if err != nil {
		result.Error = fmt.Sprintf("XML analysis failed: %v", err)
		result.Success = false
		result.ProcessingTime = time.Since(startTime)
		return xa.createResultMessage(result), nil
	}

	result.ResultData["processed_content"] = processedXML
	result.ResultData = mergeMapStringInterface(result.ResultData, analysisData)
	result.Metadata["analyzer"] = "xml_analysis"
	result.Metadata["original_length"] = len(request.Content)
	result.Metadata["processed_length"] = len(processedXML)
	result.Success = true
	result.ProcessingTime = time.Since(startTime)

	return xa.createResultMessage(result), nil
}

func (xa *XMLAnalyzer) loadConfigFromAgent(base *agent.BaseAgent) {
	if maxDepth := base.GetConfigInt("max_depth", 20); maxDepth > 0 {
		xa.config.MaxDepth = maxDepth
	}
	if maxElements := base.GetConfigInt("max_elements", 1000); maxElements > 0 {
		xa.config.MaxElements = maxElements
	}
	xa.config.EnableValidation = base.GetConfigBool("enable_validation", true)
	xa.config.EnableSchemaGen = base.GetConfigBool("enable_schema_gen", true)
	xa.config.EnableNamespaceAnalysis = base.GetConfigBool("enable_namespace_analysis", true)
	xa.config.EnableMinification = base.GetConfigBool("enable_minification", false)
}

func (xa *XMLAnalyzer) analyzeXML(content string) (string, map[string]interface{}, error) {
	analysisData := make(map[string]interface{})

	// Basic XML validation
	decoder := xml.NewDecoder(strings.NewReader(content))
	elementCount := 0
	depth := 0
	maxDepth := 0
	elementNames := make(map[string]int)
	attributeNames := make(map[string]int)
	namespaces := make(map[string]int)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			analysisData["validation_error"] = err.Error()
			analysisData["is_valid"] = false
			return content, analysisData, nil
		}

		switch t := token.(type) {
		case xml.StartElement:
			elementCount++
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}

			elementNames[t.Name.Local]++

			if xa.config.EnableNamespaceAnalysis && t.Name.Space != "" {
				namespaces[t.Name.Space]++
			}

			for _, attr := range t.Attr {
				attributeNames[attr.Name.Local]++
				if xa.config.EnableNamespaceAnalysis && attr.Name.Space != "" {
					namespaces[attr.Name.Space]++
				}
			}

		case xml.EndElement:
			depth--
		}

		if elementCount > xa.config.MaxElements {
			break
		}
	}

	analysisData["is_valid"] = true
	analysisData["element_count"] = elementCount
	analysisData["max_depth"] = maxDepth
	analysisData["unique_elements"] = len(elementNames)
	analysisData["unique_attributes"] = len(attributeNames)

	// Element frequency analysis
	analysisData["element_frequency"] = xa.getTopFrequencies(elementNames, 10)
	analysisData["attribute_frequency"] = xa.getTopFrequencies(attributeNames, 10)

	// Namespace analysis
	if xa.config.EnableNamespaceAnalysis {
		analysisData["namespace_count"] = len(namespaces)
		analysisData["namespaces"] = xa.getTopFrequencies(namespaces, 10)
	}

	// Basic XML statistics
	analysisData["original_size"] = len(content)

	// Process content
	processedContent := content
	if xa.config.EnableMinification {
		processedContent = xa.minifyXML(content)
		analysisData["minified_size"] = len(processedContent)
		analysisData["compression_ratio"] = float64(len(content)) / float64(len(processedContent))
	}

	// Detect XML type and patterns
	xmlType, patterns := xa.detectXMLTypeAndPatterns(content)
	analysisData["xml_type"] = xmlType
	analysisData["patterns"] = patterns

	// Structure analysis
	structure := xa.analyzeXMLStructure(content)
	analysisData["structure"] = structure

	// Encoding detection
	encoding := xa.detectEncoding(content)
	analysisData["encoding"] = encoding

	// Processing metadata
	analysisData["processed_by"] = "xml_analyzer"
	analysisData["processed_at"] = time.Now().Format(time.RFC3339)
	analysisData["processing_type"] = "xml_validation_and_analysis"

	return processedContent, analysisData, nil
}

func (xa *XMLAnalyzer) getTopFrequencies(frequencies map[string]int, limit int) []map[string]interface{} {
	type freq struct {
		name  string
		count int
	}

	var freqs []freq
	for name, count := range frequencies {
		freqs = append(freqs, freq{name, count})
	}

	// Sort by frequency (simple bubble sort for small datasets)
	for i := 0; i < len(freqs)-1; i++ {
		for j := i + 1; j < len(freqs); j++ {
			if freqs[i].count < freqs[j].count {
				freqs[i], freqs[j] = freqs[j], freqs[i]
			}
		}
	}

	// Get top N
	maxItems := limit
	if len(freqs) < maxItems {
		maxItems = len(freqs)
	}

	result := make([]map[string]interface{}, maxItems)
	for i := 0; i < maxItems; i++ {
		result[i] = map[string]interface{}{
			"name":  freqs[i].name,
			"count": freqs[i].count,
		}
	}

	return result
}

func (xa *XMLAnalyzer) minifyXML(content string) string {
	// Simple XML minification by removing unnecessary whitespace
	lines := strings.Split(content, "\n")
	var minified []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			minified = append(minified, trimmed)
		}
	}

	return strings.Join(minified, "")
}

func (xa *XMLAnalyzer) detectXMLTypeAndPatterns(content string) (string, map[string]interface{}) {
	patterns := make(map[string]interface{})
	xmlType := "generic"

	lowerContent := strings.ToLower(content)

	// Detect common XML types
	if strings.Contains(lowerContent, "<!doctype html") || strings.Contains(lowerContent, "<html") {
		xmlType = "html"
		patterns["html_document"] = true
	} else if strings.Contains(lowerContent, "<?xml") {
		xmlType = "xml_document"
		if strings.Contains(lowerContent, "xmlns") {
			patterns["namespace_aware"] = true
		}
	} else if strings.Contains(lowerContent, "<soap:") || strings.Contains(lowerContent, "<s:") {
		xmlType = "soap"
		patterns["soap_message"] = true
	} else if strings.Contains(lowerContent, "<rss") || strings.Contains(lowerContent, "<feed") {
		xmlType = "feed"
		patterns["rss_or_atom"] = true
	} else if strings.Contains(lowerContent, "<configuration") || strings.Contains(lowerContent, "<config") {
		xmlType = "configuration"
		patterns["configuration_file"] = true
	} else if strings.Contains(lowerContent, "<project") && strings.Contains(lowerContent, "maven") {
		xmlType = "maven_pom"
		patterns["maven_project"] = true
	} else if strings.Contains(lowerContent, "<build>") || strings.Contains(lowerContent, "<target") {
		xmlType = "build_script"
		patterns["build_configuration"] = true
	}

	// Detect structural patterns
	if strings.Count(content, "<") > 100 {
		patterns["complex_structure"] = true
	}

	if regexp.MustCompile(`<\w+\s+[^>]*id\s*=`).MatchString(lowerContent) {
		patterns["has_id_attributes"] = true
	}

	if regexp.MustCompile(`<\w+\s+[^>]*class\s*=`).MatchString(lowerContent) {
		patterns["has_class_attributes"] = true
	}

	if strings.Contains(content, "CDATA") {
		patterns["contains_cdata"] = true
	}

	if strings.Contains(content, "<!--") {
		patterns["contains_comments"] = true
	}

	return xmlType, patterns
}

func (xa *XMLAnalyzer) analyzeXMLStructure(content string) map[string]interface{} {
	structure := make(map[string]interface{})

	// Count different structural elements
	structure["element_count"] = strings.Count(content, "<") - strings.Count(content, "</") - strings.Count(content, "<!--")
	structure["closing_tag_count"] = strings.Count(content, "</")
	structure["comment_count"] = strings.Count(content, "<!--")
	structure["cdata_count"] = strings.Count(content, "<![CDATA[")

	// Analyze nesting depth
	depth := 0
	maxDepth := 0
	for _, char := range content {
		if char == '<' {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		} else if char == '>' {
			depth--
		}
	}
	structure["max_nesting_depth"] = maxDepth

	// Check for self-closing tags
	selfClosingPattern := regexp.MustCompile(`<[^>]+/>`)
	selfClosingCount := len(selfClosingPattern.FindAllString(content, -1))
	structure["self_closing_tags"] = selfClosingCount

	// Check for attributes
	attributePattern := regexp.MustCompile(`\s+\w+\s*=\s*["'][^"']*["']`)
	attributeMatches := attributePattern.FindAllString(content, -1)
	structure["total_attributes"] = len(attributeMatches)

	return structure
}

func (xa *XMLAnalyzer) detectEncoding(content string) string {
	// Look for XML declaration with encoding
	xmlDeclPattern := regexp.MustCompile(`<\?xml[^>]*encoding\s*=\s*["']([^"']*)["']`)
	matches := xmlDeclPattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}

	// Check for BOM
	if strings.HasPrefix(content, "\uFEFF") {
		return "UTF-8 with BOM"
	}

	// Default assumption
	return "UTF-8"
}

func (xa *XMLAnalyzer) createResultMessage(result *ProcessingResult) *client.BrokerMessage {
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
	xmlAnalyzer := NewXMLAnalyzer()
	agent.Run(xmlAnalyzer, "xml-analyzer")
}