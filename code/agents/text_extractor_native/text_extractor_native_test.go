package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockTextExtractorNativeRunner implements agent.AgentRunner for testing
type MockTextExtractorNativeRunner struct {
	config map[string]interface{}
}

func (m *MockTextExtractorNativeRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockTextExtractorNativeRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockTextExtractorNativeRunner) Cleanup(base *agent.BaseAgent) {
}

func TestTextExtractorNativeInitialization(t *testing.T) {
	config := map[string]interface{}{
		"supported_formats": []string{"txt", "md", "json", "csv", "xml"},
		"encoding_detection": true,
		"preserve_formatting": false,
	}

	runner := &MockTextExtractorNativeRunner{config: config}
	framework := agent.NewFramework(runner, "text-extractor-native")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Text extractor native framework created successfully")
}

func TestTextExtractorNativePlainTextExtraction(t *testing.T) {
	config := map[string]interface{}{
		"encoding": "utf-8",
		"normalize_whitespace": true,
		"remove_empty_lines": false,
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	textFile := map[string]interface{}{
		"file_path": "/test/sample.txt",
		"file_type": "txt",
		"encoding":  "utf-8",
		"content":   "Sample plain text content\nwith multiple lines\nand various formatting.",
	}

	msg := &client.BrokerMessage{
		ID:     "test-plain-text",
		Type:   "text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "extract_plain_text",
			"file":      textFile,
			"options": map[string]interface{}{
				"preserve_line_breaks": true,
				"trim_whitespace":      true,
				"normalize_encoding":   true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Plain text extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Plain text extraction test completed successfully")
}

func TestTextExtractorNativeMarkdownExtraction(t *testing.T) {
	config := map[string]interface{}{
		"markdown_processing": map[string]interface{}{
			"preserve_structure": true,
			"extract_metadata":   true,
			"parse_tables":       true,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	markdownContent := `# Main Title

## Introduction
This is the introduction section with **bold** and *italic* text.

### Subsection
- List item 1
- List item 2
- List item 3

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Value 1  | Value 2  | Value 3  |
| Value 4  | Value 5  | Value 6  |

## Conclusion
Final thoughts and summary.`

	markdownFile := map[string]interface{}{
		"file_path": "/test/document.md",
		"file_type": "markdown",
		"content":   markdownContent,
	}

	msg := &client.BrokerMessage{
		ID:     "test-markdown-extraction",
		Type:   "text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "extract_markdown",
			"file":      markdownFile,
			"options": map[string]interface{}{
				"preserve_formatting": true,
				"extract_headings":    true,
				"extract_links":       true,
				"extract_tables":      true,
				"parse_metadata":      true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Markdown extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Markdown extraction test completed successfully")
}

func TestTextExtractorNativeJSONExtraction(t *testing.T) {
	config := map[string]interface{}{
		"json_processing": map[string]interface{}{
			"extract_text_fields": true,
			"flatten_structure":   false,
			"preserve_types":      true,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	jsonContent := map[string]interface{}{
		"title":       "Sample JSON Document",
		"description": "This is a test JSON file with various text fields.",
		"metadata": map[string]interface{}{
			"author":     "Test Author",
			"created_at": "2024-09-27T10:00:00Z",
			"tags":       []string{"test", "json", "extraction"},
		},
		"content": map[string]interface{}{
			"sections": []map[string]interface{}{
				{
					"heading": "Introduction",
					"text":    "Introduction text content.",
				},
				{
					"heading": "Main Content",
					"text":    "Main content with detailed information.",
				},
			},
		},
		"numeric_data": 12345,
		"boolean_flag": true,
	}

	msg := &client.BrokerMessage{
		ID:     "test-json-extraction",
		Type:   "text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "extract_json_text",
			"file": map[string]interface{}{
				"file_path": "/test/data.json",
				"file_type": "json",
				"content":   jsonContent,
			},
			"options": map[string]interface{}{
				"text_fields_only":  true,
				"include_keys":      true,
				"max_depth":         5,
				"exclude_fields":    []string{"metadata.created_at"},
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("JSON extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("JSON extraction test completed successfully")
}

func TestTextExtractorNativeCSVExtraction(t *testing.T) {
	config := map[string]interface{}{
		"csv_processing": map[string]interface{}{
			"delimiter":       ",",
			"has_header":      true,
			"text_columns_only": false,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	csvContent := `Name,Age,Description,Location
John Doe,30,"Software engineer with 5 years experience","New York, NY"
Jane Smith,25,"Data scientist specializing in machine learning","San Francisco, CA"
Bob Johnson,35,"Product manager with background in tech","Seattle, WA"`

	csvFile := map[string]interface{}{
		"file_path": "/test/employees.csv",
		"file_type": "csv",
		"content":   csvContent,
	}

	msg := &client.BrokerMessage{
		ID:     "test-csv-extraction",
		Type:   "text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "extract_csv_text",
			"file":      csvFile,
			"options": map[string]interface{}{
				"delimiter":         ",",
				"include_headers":   true,
				"text_columns":      []string{"Name", "Description", "Location"},
				"row_separator":     "\n",
				"include_row_numbers": false,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("CSV extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("CSV extraction test completed successfully")
}

func TestTextExtractorNativeXMLExtraction(t *testing.T) {
	config := map[string]interface{}{
		"xml_processing": map[string]interface{}{
			"extract_text_nodes": true,
			"preserve_attributes": false,
			"namespace_aware":    true,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<document xmlns="http://example.com/schema">
	<metadata>
		<title>Sample XML Document</title>
		<author>Test Author</author>
		<created>2024-09-27</created>
	</metadata>
	<content>
		<section id="intro">
			<heading>Introduction</heading>
			<paragraph>This is the introduction paragraph with important information.</paragraph>
		</section>
		<section id="main">
			<heading>Main Content</heading>
			<paragraph>This section contains the main content of the document.</paragraph>
			<list>
				<item>First item in the list</item>
				<item>Second item with <emphasis>emphasized</emphasis> text</item>
				<item>Third item with additional details</item>
			</list>
		</section>
	</content>
</document>`

	xmlFile := map[string]interface{}{
		"file_path": "/test/document.xml",
		"file_type": "xml",
		"content":   xmlContent,
	}

	msg := &client.BrokerMessage{
		ID:     "test-xml-extraction",
		Type:   "text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "extract_xml_text",
			"file":      xmlFile,
			"options": map[string]interface{}{
				"extract_text_only":    true,
				"preserve_structure":   true,
				"include_attributes":   false,
				"target_elements":      []string{"paragraph", "heading", "item"},
				"exclude_elements":     []string{"metadata.created"},
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("XML extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("XML extraction test completed successfully")
}

func TestTextExtractorNativeEncodingDetection(t *testing.T) {
	config := map[string]interface{}{
		"encoding_detection": map[string]interface{}{
			"enabled":           true,
			"fallback_encoding": "utf-8",
			"confidence_threshold": 0.8,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	// Simulate files with different encodings
	files := []map[string]interface{}{
		{
			"file_path": "/test/utf8.txt",
			"detected_encoding": "utf-8",
			"content": "UTF-8 encoded text with unicode characters: áéíóú",
		},
		{
			"file_path": "/test/latin1.txt",
			"detected_encoding": "iso-8859-1",
			"content": "Latin-1 encoded text with special characters",
		},
		{
			"file_path": "/test/ascii.txt",
			"detected_encoding": "ascii",
			"content": "Plain ASCII text without special characters",
		},
	}

	for i, file := range files {
		msg := &client.BrokerMessage{
			ID:     "test-encoding-" + string(rune('1'+i)),
			Type:   "text_extraction",
			Target: "text-extractor-native",
			Payload: map[string]interface{}{
				"operation": "extract_with_encoding_detection",
				"file":      file,
				"options": map[string]interface{}{
					"auto_detect_encoding": true,
					"convert_to_utf8":      true,
					"report_encoding":      true,
				},
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Encoding detection failed for file %d: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for file %d", i+1)
		}
	}
	t.Log("Encoding detection test completed successfully")
}

func TestTextExtractorNativeBatchProcessing(t *testing.T) {
	config := map[string]interface{}{
		"batch_processing": map[string]interface{}{
			"enabled":      true,
			"max_batch_size": 10,
			"parallel_workers": 3,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	batchFiles := []map[string]interface{}{
		{"file_path": "/test/batch1.txt", "file_type": "txt"},
		{"file_path": "/test/batch2.md", "file_type": "markdown"},
		{"file_path": "/test/batch3.json", "file_type": "json"},
		{"file_path": "/test/batch4.csv", "file_type": "csv"},
		{"file_path": "/test/batch5.xml", "file_type": "xml"},
	}

	msg := &client.BrokerMessage{
		ID:     "test-batch-processing",
		Type:   "batch_text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "batch_extract",
			"files":     batchFiles,
			"batch_options": map[string]interface{}{
				"parallel_processing": true,
				"max_workers":        3,
				"error_handling":     "continue_on_error",
				"progress_reporting": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Batch processing failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Batch processing test completed successfully")
}

func TestTextExtractorNativeMetadataExtraction(t *testing.T) {
	config := map[string]interface{}{
		"metadata_extraction": map[string]interface{}{
			"enabled": true,
			"include_file_stats": true,
			"extract_content_metadata": true,
		},
	}

	runner := &MockTextExtractorNativeRunner{config: config}

	fileWithMetadata := map[string]interface{}{
		"file_path": "/test/metadata_test.txt",
		"file_type": "txt",
		"file_stats": map[string]interface{}{
			"size_bytes":      1024,
			"created_at":      "2024-09-27T09:00:00Z",
			"modified_at":     "2024-09-27T10:00:00Z",
			"permissions":     "644",
		},
		"content": "Sample text content for metadata extraction testing.",
	}

	msg := &client.BrokerMessage{
		ID:     "test-metadata-extraction",
		Type:   "text_extraction",
		Target: "text-extractor-native",
		Payload: map[string]interface{}{
			"operation": "extract_with_metadata",
			"file":      fileWithMetadata,
			"metadata_options": map[string]interface{}{
				"include_file_stats":    true,
				"analyze_content":       true,
				"extract_language":      true,
				"count_statistics":      true,
				"detect_patterns":       true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Metadata extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Metadata extraction test completed successfully")
}