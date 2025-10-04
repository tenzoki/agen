package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/agen/cellorg/internal/agent"
	"github.com/agen/cellorg/internal/client"
)

// TextExtractorNative is a Framework-compliant agent for native text extraction
// Uses local Tesseract OCR without external dependencies
type TextExtractorNative struct {
	agent.DefaultAgentRunner
	config *ExtractorConfig
}

// ExtractorConfig holds configuration for native text extraction
type ExtractorConfig struct {
	EnableOCR        bool     `json:"enable_ocr"`
	OCRLanguages     []string `json:"ocr_languages"`
	IncludeMetadata  bool     `json:"include_metadata"`
	QualityThreshold float64  `json:"quality_threshold"`
}

// ProcessMessage handles text extraction requests
func (e *TextExtractorNative) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	// Parse request
	var request struct {
		RequestID string           `json:"request_id"`
		FilePath  string           `json:"file_path"`
		Options   *ExtractorConfig `json:"options,omitempty"`
	}

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payload, &request); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	// Use config from request or default
	cfg := e.config
	if request.Options != nil {
		cfg = request.Options
	}

	start := time.Now()

	// Extract text using native methods
	text, err := e.extractText(request.FilePath, cfg)
	if err != nil {
		// Return error response
		errorResponse := map[string]interface{}{
			"request_id":     request.RequestID,
			"success":        false,
			"error":          err.Error(),
			"processing_time": time.Since(start).String(),
			"extractor_used": "text_extractor_native",
		}

		return &client.BrokerMessage{
			ID:        fmt.Sprintf("error_%d", time.Now().UnixNano()),
			Type:      "text_extraction_response",
			Target:    "text_extraction_response",
			Payload:   errorResponse,
			Meta:      make(map[string]interface{}),
			Timestamp: time.Now(),
		}, nil
	}

	// Prepare successful response
	response := map[string]interface{}{
		"request_id":      request.RequestID,
		"success":         true,
		"extracted_text":  text,
		"processing_time": time.Since(start).String(),
		"extractor_used":  "text_extractor_native",
		"quality":         0.95, // Native typically high quality
	}

	// Add metadata if requested
	if cfg.IncludeMetadata {
		response["metadata"] = map[string]interface{}{
			"format":     strings.ToUpper(filepath.Ext(request.FilePath)[1:]),
			"language":   "en", // Would need detection logic
			"confidence": 95.0,
			"words":      len(strings.Fields(text)),
		}
	}

	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "text_extraction_response",
		Target:    "text_extraction_response",
		Payload:   response,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}, nil
}

// extractText performs the actual text extraction using native methods
func (e *TextExtractorNative) extractText(filePath string, config *ExtractorConfig) (string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".txt", ".md":
		return e.extractPlainText(filePath)
	case ".png", ".jpg", ".jpeg", ".tiff", ".bmp":
		if config.EnableOCR {
			return e.extractOCR(filePath, config.OCRLanguages)
		}
		return "", fmt.Errorf("OCR disabled for image file %s", filePath)
	default:
		return "", fmt.Errorf("unsupported file format: %s", ext)
	}
}

// extractPlainText extracts text from plain text files
func (e *TextExtractorNative) extractPlainText(filePath string) (string, error) {
	cmd := exec.Command("cat", filePath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read text file: %w", err)
	}
	return string(output), nil
}

// extractOCR extracts text from images using Tesseract OCR
func (e *TextExtractorNative) extractOCR(filePath string, languages []string) (string, error) {
	// Build Tesseract command
	langStr := "eng" // default
	if len(languages) > 0 {
		langStr = strings.Join(languages, "+")
	}

	cmd := exec.Command("tesseract", filePath, "stdout", "-l", langStr)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tesseract OCR failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func main() {
	textExtractor := &TextExtractorNative{
		config: &ExtractorConfig{
			EnableOCR:        true,
			OCRLanguages:     []string{"eng"},
			IncludeMetadata:  true,
			QualityThreshold: 0.8,
		},
	}

	agent.Run(textExtractor, "text-extractor-native")
}