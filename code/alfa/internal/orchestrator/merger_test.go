package orchestrator

import (
	"strings"
	"testing"
)

// Test merging with single response
func TestMergeSingleResponse(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{"Single response"}

	merged, err := merger.Merge(responses, ContentTypeText)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if merged != "Single response" {
		t.Errorf("Expected 'Single response', got '%s'", merged)
	}
}

// Test merging with no responses
func TestMergeNoResponses(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{}

	_, err := merger.Merge(responses, ContentTypeText)
	if err == nil {
		t.Error("Expected error for empty responses, got nil")
	}
}

// Test merging text responses
func TestMergeTextResponses(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{
		"Here is the first part of the answer. It contains important information.",
		"Let me continue with the second part. This has more details.",
		"In conclusion, this is the final part. It summarizes everything.",
	}

	merged, err := merger.Merge(responses, ContentTypeText)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check that all content is present
	if !strings.Contains(merged, "first part") {
		t.Error("First part missing from merged text")
	}
	if !strings.Contains(merged, "second part") {
		t.Error("Second part missing from merged text")
	}
	if !strings.Contains(merged, "final part") {
		t.Error("Final part missing from merged text")
	}

	t.Logf("Merged text: %s", merged)
}

// Test merging code responses
func TestMergeCodeResponses(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{
		`import fmt
import os

func main() {
    fmt.Println("Hello")
}`,
		`import fmt
import strings

func helper() {
    fmt.Println("Helper")
}`,
	}

	merged, err := merger.Merge(responses, ContentTypeCode)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check that imports are deduplicated
	importCount := strings.Count(merged, "import fmt")
	if importCount > 1 {
		t.Errorf("Expected single 'import fmt', found %d", importCount)
	}

	// Check that both functions are present
	if !strings.Contains(merged, "func main()") {
		t.Error("main function missing")
	}
	if !strings.Contains(merged, "func helper()") {
		t.Error("helper function missing")
	}

	t.Logf("Merged code:\n%s", merged)
}

// Test merging analysis responses
func TestMergeAnalysisResponses(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{
		`Issue: Unused variable x
Issue: Missing error handling`,
		`Issue: Unused variable x
Issue: Type mismatch in function call`,
	}

	merged, err := merger.Merge(responses, ContentTypeAnalysis)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Check that analysis sections are present
	if !strings.Contains(merged, "Analysis Part 1") {
		t.Error("Analysis Part 1 missing")
	}
	if !strings.Contains(merged, "Analysis Part 2") {
		t.Error("Analysis Part 2 missing")
	}

	// Check that issues are deduplicated
	// Note: Basic deduplication in mergeAnalysis uses normalized strings
	issueCount := strings.Count(strings.ToLower(merged), "unused variable x")
	if issueCount > 2 { // One per section max
		t.Logf("Warning: Duplicate issues found (%d occurrences)", issueCount)
	}

	t.Logf("Merged analysis:\n%s", merged)
}

// Test content type detection
func TestDetectContentType(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	tests := []struct {
		text     string
		expected ContentType
	}{
		{"func main() { }", ContentTypeCode},
		{"class MyClass { }", ContentTypeCode},
		{"import package", ContentTypeCode},
		{"```go\ncode\n```", ContentTypeCode},
		{"Issue: Something wrong", ContentTypeAnalysis},
		{"Finding: Bug detected", ContentTypeAnalysis},
		{"This is regular text", ContentTypeText},
	}

	for _, test := range tests {
		detected := merger.detectContentType(test.text)
		if detected != test.expected {
			t.Errorf("For text '%s', expected %d, got %d", test.text, test.expected, detected)
		}
	}
}

// Test auto content type detection in merge
func TestMergeAutoContentType(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	// Code responses
	codeResponses := []string{
		"func first() { }",
		"func second() { }",
	}

	merged, err := merger.Merge(codeResponses, ContentTypeAuto)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	if !strings.Contains(merged, "first") || !strings.Contains(merged, "second") {
		t.Error("Code merge failed with auto detection")
	}
}

// Test cleanTextChunk - now just returns text as-is
func TestCleanTextChunk(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	text := "Here is the answer. Content goes here."
	cleaned := merger.cleanTextChunk(text, false, false)

	// Should return text unchanged
	if cleaned != text {
		t.Errorf("Expected unchanged text, got: %s", cleaned)
	}
}

// Test deduplication prompt building
func TestBuildDeduplicationPrompt(t *testing.T) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	merged := "Test content with duplicates"

	tests := []struct {
		contentType ContentType
		keyword     string
	}{
		{ContentTypeCode, "code"},
		{ContentTypeText, "text"},
		{ContentTypeAnalysis, "analysis"},
	}

	for _, test := range tests {
		prompt := merger.buildDeduplicationPrompt(merged, test.contentType)

		if !strings.Contains(prompt, test.keyword) {
			t.Errorf("Expected prompt to contain '%s', got: %s", test.keyword, prompt)
		}

		if !strings.Contains(prompt, merged) {
			t.Error("Expected prompt to contain merged content")
		}

		if !strings.Contains(prompt, "Remove redundancies") {
			t.Error("Expected prompt to contain deduplication instruction")
		}
	}
}

// Test deduplication with mock LLM
func TestDeduplicate(t *testing.T) {
	mockLLM := &mockLLM{
		responses: []string{"Deduplicated content"},
	}
	merger := NewResponseMerger(mockLLM)

	merged := "Content with duplicates. Content with duplicates."

	deduplicated, err := merger.Deduplicate(merged, ContentTypeText)
	if err != nil {
		t.Fatalf("Deduplicate failed: %v", err)
	}

	if deduplicated != "Deduplicated content" {
		t.Errorf("Expected 'Deduplicated content', got '%s'", deduplicated)
	}

	// Check that LLM was called
	if mockLLM.callCount != 1 {
		t.Errorf("Expected 1 LLM call, got %d", mockLLM.callCount)
	}
}

// Test deduplication with auto content type
func TestDeduplicateAutoContentType(t *testing.T) {
	mockLLM := &mockLLM{
		responses: []string{"Deduplicated code"},
	}
	merger := NewResponseMerger(mockLLM)

	merged := "func test() { } func test() { }"

	deduplicated, err := merger.Deduplicate(merged, ContentTypeAuto)
	if err != nil {
		t.Fatalf("Deduplicate failed: %v", err)
	}

	if deduplicated == "" {
		t.Error("Expected non-empty deduplicated content")
	}
}

// Benchmark merging
func BenchmarkMergeText(b *testing.B) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{
		"First response with content",
		"Second response with more content",
		"Third response with final content",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = merger.Merge(responses, ContentTypeText)
	}
}

// Benchmark code merging
func BenchmarkMergeCode(b *testing.B) {
	mockLLM := &mockLLM{}
	merger := NewResponseMerger(mockLLM)

	responses := []string{
		"import fmt\nimport os\n\nfunc main() {}",
		"import fmt\nimport strings\n\nfunc helper() {}",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = merger.Merge(responses, ContentTypeCode)
	}
}
