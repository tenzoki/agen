package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

type OCRHTTPStub struct {
	agent.DefaultAgentRunner
	config *OCRConfig
}

type OCRConfig struct {
	ServiceURL      string        `json:"service_url"`
	RequestTimeout  time.Duration `json:"request_timeout"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	HealthCheckURL  string        `json:"health_check_url"`
	SupportedFormat []string      `json:"supported_formats"`
}

type OCRRequest struct {
	RequestID    string                 `json:"request_id"`
	FilePath     string                 `json:"file_path,omitempty"`
	FileData     []byte                 `json:"file_data,omitempty"`
	FileName     string                 `json:"file_name,omitempty"`
	Options      OCROptions             `json:"options,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ReplyTo      string                 `json:"reply_to,omitempty"`
}

type OCROptions struct {
	Languages    []string `json:"languages,omitempty"`
	PSM          int      `json:"psm,omitempty"`
	OEM          int      `json:"oem,omitempty"`
	Preprocess   bool     `json:"preprocess,omitempty"`
	MinConfidence float64 `json:"min_confidence,omitempty"`
}

type OCRResponse struct {
	RequestID      string                 `json:"request_id"`
	Success        bool                   `json:"success"`
	Text           string                 `json:"text,omitempty"`
	Confidence     float64                `json:"confidence,omitempty"`
	WordCount      int                    `json:"word_count,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Language       string                 `json:"language,omitempty"`
	Error          string                 `json:"error,omitempty"`
	ServiceInfo    map[string]interface{} `json:"service_info,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type HTTPOCRServiceResponse struct {
	Text           string  `json:"text"`
	Confidence     float64 `json:"confidence"`
	WordCount      int     `json:"word_count"`
	Language       string  `json:"language"`
	ProcessingTime float64 `json:"processing_time"`
	Status         string  `json:"status"`
	Error          string  `json:"error,omitempty"`
}

func NewOCRHTTPStub() *OCRHTTPStub {
	return &OCRHTTPStub{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
	}
}

func (o *OCRHTTPStub) Init(base *agent.BaseAgent) error {
	config, err := o.loadConfiguration(base)
	if err != nil {
		return fmt.Errorf("failed to load OCR HTTP configuration: %w", err)
	}
	o.config = config

	// Health check on initialization
	if err := o.checkServiceHealth(); err != nil {
		base.LogError("OCR HTTP service health check failed during init: %v", err)
		base.LogInfo("Will retry health checks on demand...")
	} else {
		base.LogInfo("OCR HTTP service is healthy: %s", config.ServiceURL)
	}

	base.LogInfo("OCR HTTP Stub Agent initialized")
	base.LogInfo("Service URL: %s", config.ServiceURL)
	base.LogInfo("Request timeout: %v", config.RequestTimeout)
	base.LogInfo("Max retries: %d", config.MaxRetries)
	base.LogInfo("Supported formats: %v", config.SupportedFormat)

	return nil
}

func (o *OCRHTTPStub) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "ocr_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	startTime := time.Now()

	payload, err := json.Marshal(msg.Payload)
	if err != nil {
		return o.createErrorResponse("", "failed to marshal payload", startTime), nil
	}

	var request OCRRequest
	if err := json.Unmarshal(payload, &request); err != nil {
		return o.createErrorResponse("", "failed to unmarshal request", startTime), nil
	}

	if request.RequestID == "" {
		request.RequestID = fmt.Sprintf("ocr_req_%d", time.Now().UnixNano())
	}

	base.LogInfo("Processing OCR request %s via HTTP service", request.RequestID)

	// Validate request
	if request.FilePath == "" && len(request.FileData) == 0 {
		return o.createErrorResponse(request.RequestID, "either file_path or file_data must be provided", startTime), nil
	}

	// Perform OCR via HTTP service
	response := o.performOCRRequest(request, base, startTime)

	return o.createResultMessage(response), nil
}

func (o *OCRHTTPStub) performOCRRequest(request OCRRequest, base *agent.BaseAgent, startTime time.Time) OCRResponse {
	// Prepare file data
	var fileData []byte
	var fileName string
	var err error

	if request.FilePath != "" {
		fileData, err = os.ReadFile(request.FilePath)
		if err != nil {
			return OCRResponse{
				RequestID:      request.RequestID,
				Success:        false,
				Error:          fmt.Sprintf("failed to read file: %v", err),
				ProcessingTime: time.Since(startTime),
			}
		}
		fileName = filepath.Base(request.FilePath)
	} else {
		fileData = request.FileData
		fileName = request.FileName
		if fileName == "" {
			fileName = "document.png" // Default filename
		}
	}

	// Check file format
	if !o.isSupportedFormat(fileName) {
		return OCRResponse{
			RequestID:      request.RequestID,
			Success:        false,
			Error:          fmt.Sprintf("unsupported file format: %s", filepath.Ext(fileName)),
			ProcessingTime: time.Since(startTime),
		}
	}

	// Perform HTTP request with retries
	var lastErr error
	for attempt := 0; attempt <= o.config.MaxRetries; attempt++ {
		if attempt > 0 {
			base.LogInfo("OCR HTTP request retry %d/%d for request %s", attempt, o.config.MaxRetries, request.RequestID)
			time.Sleep(o.config.RetryDelay)
		}

		response, err := o.sendOCRRequest(fileData, fileName, request.Options, base)
		if err != nil {
			lastErr = err
			continue
		}

		// Success
		ocrResponse := OCRResponse{
			RequestID:      request.RequestID,
			Success:        true,
			Text:           response.Text,
			Confidence:     response.Confidence,
			WordCount:      response.WordCount,
			Language:       response.Language,
			ProcessingTime: time.Since(startTime),
			ServiceInfo: map[string]interface{}{
				"service_url":        o.config.ServiceURL,
				"service_processing_time": response.ProcessingTime,
				"service_status":     response.Status,
			},
			Metadata: map[string]interface{}{
				"file_name":     fileName,
				"file_size":     len(fileData),
				"attempts":      attempt + 1,
				"ocr_languages": request.Options.Languages,
				"psm":           request.Options.PSM,
				"oem":           request.Options.OEM,
			},
		}

		base.LogInfo("OCR completed for request %s: %d characters extracted (confidence: %.1f%%)",
			request.RequestID, len(response.Text), response.Confidence)

		return ocrResponse
	}

	// All retries failed
	return OCRResponse{
		RequestID:      request.RequestID,
		Success:        false,
		Error:          fmt.Sprintf("OCR HTTP request failed after %d retries: %v", o.config.MaxRetries+1, lastErr),
		ProcessingTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"file_name": fileName,
			"file_size": len(fileData),
			"attempts":  o.config.MaxRetries + 1,
		},
	}
}

func (o *OCRHTTPStub) sendOCRRequest(fileData []byte, fileName string, options OCROptions, base *agent.BaseAgent) (*HTTPOCRServiceResponse, error) {
	// Create multipart form request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file
	fileWriter, err := writer.CreateFormFile("image", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := fileWriter.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %w", err)
	}

	// Add OCR options
	if len(options.Languages) > 0 {
		languagesStr := strings.Join(options.Languages, "+")
		writer.WriteField("languages", languagesStr)
	}

	if options.PSM > 0 {
		writer.WriteField("psm", fmt.Sprintf("%d", options.PSM))
	}

	if options.OEM > 0 {
		writer.WriteField("oem", fmt.Sprintf("%d", options.OEM))
	}

	if options.Preprocess {
		writer.WriteField("preprocess", "true")
	}

	writer.Close()

	// Create HTTP request
	ctx, cancel := context.WithTimeout(context.Background(), o.config.RequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.ServiceURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "GOX-OCR-HTTP-Stub/1.0")

	// Send request
	client := &http.Client{
		Timeout: o.config.RequestTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var ocrResponse HTTPOCRServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocrResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ocrResponse.Status != "" && ocrResponse.Status != "success" {
		return nil, fmt.Errorf("OCR service returned error: %s", ocrResponse.Error)
	}

	return &ocrResponse, nil
}

func (o *OCRHTTPStub) checkServiceHealth() error {
	healthURL := o.config.HealthCheckURL
	if healthURL == "" {
		// Derive health URL from service URL
		baseURL := strings.TrimSuffix(o.config.ServiceURL, "/ocr")
		healthURL = baseURL + "/health"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (o *OCRHTTPStub) isSupportedFormat(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, supportedExt := range o.config.SupportedFormat {
		if ext == strings.ToLower(supportedExt) {
			return true
		}
	}
	return false
}

func (o *OCRHTTPStub) loadConfiguration(base *agent.BaseAgent) (*OCRConfig, error) {
	config := &OCRConfig{
		ServiceURL:     base.GetConfigString("service_url", "http://localhost:8080/ocr"),
		RequestTimeout: time.Duration(base.GetConfigInt("request_timeout", 300000000000)), // 300 seconds
		MaxRetries:     base.GetConfigInt("max_retries", 3),
		RetryDelay:     time.Duration(base.GetConfigInt("retry_delay", 5000000000)), // 5 seconds
		HealthCheckURL: base.GetConfigString("health_check_url", ""),
		SupportedFormat: []string{".png", ".jpg", ".jpeg", ".tiff", ".bmp", ".pdf"},
	}

	// Override supported formats if specified
	if formatsStr := base.GetConfigString("supported_formats", ""); formatsStr != "" {
		config.SupportedFormat = strings.Split(formatsStr, ",")
		for i, format := range config.SupportedFormat {
			config.SupportedFormat[i] = strings.TrimSpace(format)
		}
	}

	return config, nil
}

func (o *OCRHTTPStub) createErrorResponse(requestID, errorMsg string, startTime time.Time) *client.BrokerMessage {
	response := OCRResponse{
		RequestID:      requestID,
		Success:        false,
		Error:          errorMsg,
		ProcessingTime: time.Since(startTime),
	}

	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "ocr_response",
		Target:    "ocr_response",
		Payload:   response,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func (o *OCRHTTPStub) createResultMessage(response OCRResponse) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("result_%d", time.Now().UnixNano()),
		Type:      "ocr_response",
		Target:    "ocr_response",
		Payload:   response,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func main() {
	stub := NewOCRHTTPStub()
	agent.Run(stub, "ocr-http-stub")
}