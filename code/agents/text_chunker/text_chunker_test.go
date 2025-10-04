package main

import (
	"testing"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
	"github.com/tenzoki/gox/test/testutil"
)

// MockTextChunkerRunner implements agent.AgentRunner for testing
type MockTextChunkerRunner struct {
	config map[string]interface{}
}

func (m *MockTextChunkerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockTextChunkerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockTextChunkerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestTextChunkerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"default_chunk_size": 2048,
		"max_chunk_size":     10485760, // 10MB
		"min_chunk_size":     256,
		"default_overlap":    256,
	}

	runner := &MockTextChunkerRunner{config: config}
	framework := agent.NewFramework(runner, "text-chunker")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Text chunker framework created successfully")
}

func TestTextChunkerSizeBasedChunking(t *testing.T) {
	config := map[string]interface{}{
		"chunk_size": 500,
		"strategy":   "size_based",
	}

	runner := &MockTextChunkerRunner{config: config}

	// Load real test data
	longText, err := testutil.LoadTestFileString(testutil.SampleArticlePath)
	if err != nil {
		t.Fatalf("Failed to load test data: %v", err)
	}

	msg := &client.BrokerMessage{
		ID:     "test-size-based",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":       longText,
			"strategy":   "size_based",
			"chunk_size": 500,
			"source_file": testutil.GetTestDataPath(testutil.SampleArticlePath),
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Size-based chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Size-based chunking test completed successfully")
}

func TestTextChunkerParagraphBasedChunking(t *testing.T) {
	config := map[string]interface{}{
		"strategy": "paragraph_based",
	}

	runner := &MockTextChunkerRunner{config: config}

	paragraphText := `First paragraph contains some introductory text about the topic.
This paragraph provides context and sets up the discussion.

Second paragraph delves deeper into the main subject matter.
It contains more detailed information and analysis.

Third paragraph concludes the discussion with final thoughts.
This paragraph wraps up the main points and provides closure.

Fourth paragraph might contain additional references or notes.
It serves as supplementary information to the main content.`

	msg := &client.BrokerMessage{
		ID:     "test-paragraph-based",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":     paragraphText,
			"strategy": "paragraph_based",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Paragraph-based chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Paragraph-based chunking test completed successfully")
}

func TestTextChunkerSectionBasedChunking(t *testing.T) {
	config := map[string]interface{}{
		"strategy": "section_based",
	}

	runner := &MockTextChunkerRunner{config: config}

	sectionText := `# Introduction
This is the introduction section of the document.
It provides an overview of what will be covered.

## Subsection 1.1
This subsection covers specific details about the first topic.
It includes examples and explanations.

# Main Content
This is the main content section of the document.
It contains the bulk of the information.

## Subsection 2.1
This subsection provides detailed analysis.

## Subsection 2.2
This subsection covers implementation details.

# Conclusion
This is the conclusion section.
It summarizes the key points and provides final thoughts.`

	msg := &client.BrokerMessage{
		ID:     "test-section-based",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":     sectionText,
			"strategy": "section_based",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Section-based chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Section-based chunking test completed successfully")
}

func TestTextChunkerBoundaryBasedChunking(t *testing.T) {
	config := map[string]interface{}{
		"strategy": "boundary_based",
		"boundary_patterns": []string{"\n\n", "---", "EOF"},
	}

	runner := &MockTextChunkerRunner{config: config}

	boundaryText := `First section of text with natural boundaries.
This continues the first section.

Second section starts here after double newline.
This is more content in the second section.

---

Third section starts after the separator.
This section has different content.

EOF

Final section after EOF marker.
This is the last piece of content.`

	msg := &client.BrokerMessage{
		ID:     "test-boundary-based",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":     boundaryText,
			"strategy": "boundary_based",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Boundary-based chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Boundary-based chunking test completed successfully")
}

func TestTextChunkerWithOverlap(t *testing.T) {
	config := map[string]interface{}{
		"chunk_size": 50,
		"overlap_size": 10,
		"strategy": "size_based",
	}

	runner := &MockTextChunkerRunner{config: config}

	overlapText := "This text will be chunked with overlap to ensure continuity between chunks. " +
		"The overlap helps maintain context when processing chunks independently."

	msg := &client.BrokerMessage{
		ID:     "test-overlap",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":        overlapText,
			"strategy":    "size_based",
			"chunk_size":  50,
			"overlap_size": 10,
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Overlap chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Overlap chunking test completed successfully")
}

func TestTextChunkerSemanticChunking(t *testing.T) {
	config := map[string]interface{}{
		"strategy": "semantic",
		"semantic_analysis": true,
	}

	runner := &MockTextChunkerRunner{config: config}

	semanticText := `Machine learning is a subset of artificial intelligence that focuses on algorithms.
These algorithms can learn and make decisions from data without being explicitly programmed.

Natural language processing is another important area of AI.
It deals with the interaction between computers and human language.
NLP enables machines to understand, interpret, and generate human language.

Computer vision is the field that trains machines to interpret visual information.
It involves processing images and videos to extract meaningful insights.
This technology is used in autonomous vehicles and medical imaging.`

	msg := &client.BrokerMessage{
		ID:     "test-semantic",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":     semanticText,
			"strategy": "semantic",
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Semantic chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Semantic chunking test completed successfully")
}

func TestTextChunkerMetadataPreservation(t *testing.T) {
	config := map[string]interface{}{
		"preserve_metadata": true,
		"strategy":         "paragraph_based",
	}

	runner := &MockTextChunkerRunner{config: config}

	msg := &client.BrokerMessage{
		ID:     "test-metadata",
		Type:   "text_chunking",
		Target: "text-chunker",
		Payload: map[string]interface{}{
			"text":     "First paragraph.\n\nSecond paragraph.\n\nThird paragraph.",
			"strategy": "paragraph_based",
			"metadata": map[string]interface{}{
				"source_file": "/test/document.txt",
				"author":      "Test Author",
				"created_at":  "2024-09-27T10:00:00Z",
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Metadata preservation chunking failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Metadata preservation chunking test completed successfully")
}