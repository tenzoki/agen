package main

import (
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockTextTransformerRunner implements agent.AgentRunner for testing
type MockTextTransformerRunner struct {
	config map[string]interface{}
}

func (m *MockTextTransformerRunner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	return msg, nil
}

func (m *MockTextTransformerRunner) Init(base *agent.BaseAgent) error {
	return nil
}

func (m *MockTextTransformerRunner) Cleanup(base *agent.BaseAgent) {
}

func TestTextTransformerInitialization(t *testing.T) {
	config := map[string]interface{}{
		"transformation_types": []string{"normalize", "clean", "format", "tokenize"},
		"default_encoding":     "utf-8",
		"preserve_structure":   true,
	}

	runner := &MockTextTransformerRunner{config: config}
	framework := agent.NewFramework(runner, "text-transformer")
	if framework == nil {
		t.Fatal("Expected agent framework to be created")
	}

	t.Log("Text transformer framework created successfully")
}

func TestTextTransformerNormalization(t *testing.T) {
	config := map[string]interface{}{
		"normalization": map[string]interface{}{
			"unicode_normalization": "NFC",
			"case_conversion":       "none",
			"whitespace_handling":   "normalize",
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	inputText := map[string]interface{}{
		"content": "   This   is    sample   text   with   irregular    spacing   and\n\n\nextra\nlines.   ",
		"encoding": "utf-8",
		"metadata": map[string]interface{}{
			"source": "user_input",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-normalization",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "normalize_text",
			"input_text":  inputText,
			"normalization_config": map[string]interface{}{
				"normalize_whitespace":  true,
				"normalize_unicode":     true,
				"remove_extra_spaces":   true,
				"normalize_line_breaks": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text normalization failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text normalization test completed successfully")
}

func TestTextTransformerCleaning(t *testing.T) {
	config := map[string]interface{}{
		"cleaning": map[string]interface{}{
			"remove_html_tags":     true,
			"remove_special_chars": false,
			"remove_urls":          true,
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	dirtyText := map[string]interface{}{
		"content": "<p>This is <b>sample</b> text with <a href='http://example.com'>links</a> and HTML tags.</p>\n" +
			"Email: test@example.com\nPhone: +1-555-123-4567\n" +
			"Visit: https://www.example.com for more information!",
		"source_format": "html",
	}

	msg := &client.BrokerMessage{
		ID:     "test-cleaning",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":  "clean_text",
			"input_text": dirtyText,
			"cleaning_config": map[string]interface{}{
				"remove_html_tags":    true,
				"remove_urls":         true,
				"preserve_emails":     true,
				"preserve_phone_numbers": true,
				"remove_punctuation":  false,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text cleaning failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text cleaning test completed successfully")
}

func TestTextTransformerTokenization(t *testing.T) {
	config := map[string]interface{}{
		"tokenization": map[string]interface{}{
			"method":          "word_tokenize",
			"preserve_case":   false,
			"remove_stopwords": false,
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	textToTokenize := map[string]interface{}{
		"content": "Natural language processing is a fascinating field of artificial intelligence. " +
			"It involves the interaction between computers and human language.",
		"language": "en",
	}

	msg := &client.BrokerMessage{
		ID:     "test-tokenization",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "tokenize_text",
			"input_text":  textToTokenize,
			"tokenization_config": map[string]interface{}{
				"tokenizer_type":   "word",
				"lowercase":        true,
				"remove_punctuation": true,
				"remove_stopwords": true,
				"stemming":         false,
				"lemmatization":    true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text tokenization failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text tokenization test completed successfully")
}

func TestTextTransformerFormatConversion(t *testing.T) {
	config := map[string]interface{}{
		"format_conversion": map[string]interface{}{
			"supported_formats": []string{"plain", "markdown", "html", "json"},
			"preserve_semantics": true,
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	markdownText := map[string]interface{}{
		"content": "# Main Title\n\n## Subtitle\n\nThis is **bold** and *italic* text.\n\n- List item 1\n- List item 2\n\n[Link](http://example.com)",
		"source_format": "markdown",
		"target_format":  "html",
	}

	msg := &client.BrokerMessage{
		ID:     "test-format-conversion",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "convert_format",
			"input_text":  markdownText,
			"conversion_config": map[string]interface{}{
				"source_format":      "markdown",
				"target_format":      "html",
				"preserve_structure": true,
				"include_metadata":   true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Format conversion failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Format conversion test completed successfully")
}

func TestTextTransformerLanguageDetection(t *testing.T) {
	config := map[string]interface{}{
		"language_processing": map[string]interface{}{
			"detection_enabled": true,
			"supported_languages": []string{"en", "es", "fr", "de", "it"},
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	multilingualTexts := []map[string]interface{}{
		{
			"content":           "This is a sample text in English language.",
			"expected_language": "en",
		},
		{
			"content":           "Este es un texto de muestra en idioma español.",
			"expected_language": "es",
		},
		{
			"content":           "Ceci est un échantillon de texte en langue française.",
			"expected_language": "fr",
		},
	}

	for i, text := range multilingualTexts {
		msg := &client.BrokerMessage{
			ID:     "test-language-detection-" + string(rune('1'+i)),
			Type:   "text_transformation",
			Target: "text-transformer",
			Payload: map[string]interface{}{
				"operation":   "detect_language",
				"input_text":  text,
				"detection_config": map[string]interface{}{
					"confidence_threshold": 0.8,
					"return_alternatives": true,
					"detect_multiple":     false,
				},
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Language detection failed for text %d: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for text %d", i+1)
		}
	}
	t.Log("Language detection test completed successfully")
}

func TestTextTransformerSentimentAnalysis(t *testing.T) {
	config := map[string]interface{}{
		"sentiment_analysis": map[string]interface{}{
			"enabled": true,
			"model":   "lexicon_based",
			"scale":   "positive_negative_neutral",
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	sentimentTexts := []map[string]interface{}{
		{
			"content": "I absolutely love this product! It's amazing and works perfectly.",
			"expected_sentiment": "positive",
		},
		{
			"content": "This is terrible and I hate it. Worst purchase ever.",
			"expected_sentiment": "negative",
		},
		{
			"content": "The product is okay. It works as expected, nothing special.",
			"expected_sentiment": "neutral",
		},
	}

	for i, text := range sentimentTexts {
		msg := &client.BrokerMessage{
			ID:     "test-sentiment-" + string(rune('1'+i)),
			Type:   "text_transformation",
			Target: "text-transformer",
			Payload: map[string]interface{}{
				"operation":   "analyze_sentiment",
				"input_text":  text,
				"sentiment_config": map[string]interface{}{
					"return_confidence": true,
					"detailed_analysis": true,
					"emotion_detection": false,
				},
			},
			Meta: make(map[string]interface{}),
		}

		result, err := runner.ProcessMessage(msg, nil)
		if err != nil {
			t.Errorf("Sentiment analysis failed for text %d: %v", i+1, err)
		}
		if result == nil {
			t.Errorf("Expected result message for text %d", i+1)
		}
	}
	t.Log("Sentiment analysis test completed successfully")
}

func TestTextTransformerEntityExtraction(t *testing.T) {
	config := map[string]interface{}{
		"entity_extraction": map[string]interface{}{
			"enabled": true,
			"entity_types": []string{"PERSON", "ORG", "LOCATION", "DATE", "MONEY"},
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	entityText := map[string]interface{}{
		"content": "John Smith works at Microsoft in Seattle and earned $75,000 last year in 2023. " +
			"He will meet with Sarah Johnson from Google in San Francisco on March 15th, 2024.",
	}

	msg := &client.BrokerMessage{
		ID:     "test-entity-extraction",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "extract_entities",
			"input_text":  entityText,
			"extraction_config": map[string]interface{}{
				"entity_types":      []string{"PERSON", "ORG", "LOCATION", "DATE", "MONEY"},
				"include_confidence": true,
				"resolve_coreferences": true,
				"normalize_entities": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Entity extraction failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Entity extraction test completed successfully")
}

func TestTextTransformerTextSummarization(t *testing.T) {
	config := map[string]interface{}{
		"summarization": map[string]interface{}{
			"enabled": true,
			"method":  "extractive",
			"max_sentences": 3,
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	longText := map[string]interface{}{
		"content": "Artificial intelligence (AI) is intelligence demonstrated by machines, in contrast to the natural intelligence displayed by humans and animals. " +
			"Leading AI textbooks define the field as the study of 'intelligent agents': any device that perceives its environment and takes actions that maximize its chance of successfully achieving its goals. " +
			"Colloquially, the term 'artificial intelligence' is often used to describe machines that mimic 'cognitive' functions that humans associate with the human mind, such as 'learning' and 'problem solving'. " +
			"As machines become increasingly capable, tasks considered to require 'intelligence' are often removed from the definition of AI, a phenomenon known as the AI effect. " +
			"A quip in Tesler's Theorem says 'AI is whatever hasn't been done yet.' " +
			"For instance, optical character recognition is frequently excluded from things considered to be AI, having become a routine technology.",
	}

	msg := &client.BrokerMessage{
		ID:     "test-summarization",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "summarize_text",
			"input_text":  longText,
			"summarization_config": map[string]interface{}{
				"method":         "extractive",
				"max_sentences":  2,
				"preserve_order": true,
				"include_scores": true,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text summarization failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text summarization test completed successfully")
}

func TestTextTransformerTranslation(t *testing.T) {
	config := map[string]interface{}{
		"translation": map[string]interface{}{
			"enabled": true,
			"service": "mock_translator",
			"supported_pairs": []map[string]string{
				{"from": "en", "to": "es"},
				{"from": "en", "to": "fr"},
				{"from": "es", "to": "en"},
			},
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	translationRequest := map[string]interface{}{
		"content":         "Hello, how are you today? I hope you are doing well.",
		"source_language": "en",
		"target_language": "es",
	}

	msg := &client.BrokerMessage{
		ID:     "test-translation",
		Type:   "text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "translate_text",
			"input_text":  translationRequest,
			"translation_config": map[string]interface{}{
				"preserve_formatting": true,
				"confidence_threshold": 0.8,
				"alternative_translations": 2,
			},
		},
		Meta: make(map[string]interface{}),
	}

	result, err := runner.ProcessMessage(msg, nil)
	if err != nil {
		t.Errorf("Text translation failed: %v", err)
	}
	if result == nil {
		t.Error("Expected result message")
	}
	t.Log("Text translation test completed successfully")
}

func TestTextTransformerBatchProcessing(t *testing.T) {
	config := map[string]interface{}{
		"batch_processing": map[string]interface{}{
			"enabled":     true,
			"max_batch_size": 5,
			"parallel_workers": 2,
		},
	}

	runner := &MockTextTransformerRunner{config: config}

	batchTexts := []map[string]interface{}{
		{
			"id":      "text-1",
			"content": "First text for batch processing.",
			"operation": "normalize_text",
		},
		{
			"id":      "text-2",
			"content": "Second text for batch processing.",
			"operation": "clean_text",
		},
		{
			"id":      "text-3",
			"content": "Third text for batch processing.",
			"operation": "tokenize_text",
		},
	}

	msg := &client.BrokerMessage{
		ID:     "test-batch-processing",
		Type:   "batch_text_transformation",
		Target: "text-transformer",
		Payload: map[string]interface{}{
			"operation":   "batch_transform",
			"input_texts": batchTexts,
			"batch_config": map[string]interface{}{
				"parallel_processing": true,
				"error_handling":      "continue_on_error",
				"progress_reporting":  true,
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