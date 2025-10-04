package speech

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
	"time"
)

const whisperAPIURL = "https://api.openai.com/v1/audio/transcriptions"

// WhisperSTT implements STT interface using OpenAI's Whisper API
type WhisperSTT struct {
	config     STTConfig
	httpClient *http.Client
}

// NewWhisperSTT creates a new Whisper STT client
func NewWhisperSTT(config STTConfig) *WhisperSTT {
	if config.Model == "" {
		config.Model = "whisper-1"
	}
	if config.Temperature == 0 {
		config.Temperature = 0.0
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	return &WhisperSTT{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// whisperResponse represents the response from Whisper API
type whisperResponse struct {
	Text string `json:"text"`
}

// whisperErrorResponse represents an error response from Whisper API
type whisperErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Transcribe converts audio to text
func (w *WhisperSTT) Transcribe(ctx context.Context, audio io.Reader, format string) (*Transcription, error) {
	startTime := time.Now()

	var result *Transcription
	var err error

	// Retry loop
	for attempt := 0; attempt <= w.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(w.config.RetryDelay * time.Duration(attempt)):
			}
		}

		result, err = w.makeRequest(ctx, audio, format, "audio."+format)
		if err == nil {
			result.Duration = time.Since(startTime)
			return result, nil
		}

		// Check if error is retryable
		if speechErr, ok := err.(*Error); ok && !speechErr.Retry {
			return nil, err
		}

		// Reset reader if possible for retry
		if seeker, ok := audio.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", w.config.RetryCount, err)
}

// TranscribeFile converts an audio file to text
func (w *WhisperSTT) TranscribeFile(ctx context.Context, filePath string) (*Transcription, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "file_error",
			Message:  fmt.Sprintf("failed to open file: %v", err),
			Retry:    false,
		}
	}
	defer file.Close()

	format := filepath.Ext(filePath)
	if len(format) > 0 {
		format = format[1:] // Remove leading dot
	}

	return w.Transcribe(ctx, file, format)
}

// makeRequest performs the actual HTTP request to Whisper API
func (w *WhisperSTT) makeRequest(ctx context.Context, audio io.Reader, format, filename string) (*Transcription, error) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file field
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "form_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	if _, err := io.Copy(part, audio); err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "copy_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	// Add other fields
	writer.WriteField("model", w.config.Model)
	if w.config.Language != "" {
		writer.WriteField("language", w.config.Language)
	}
	if w.config.Temperature > 0 {
		writer.WriteField("temperature", fmt.Sprintf("%.2f", w.config.Temperature))
	}

	if err := writer.Close(); err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "form_close_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", whisperAPIURL, body)
	if err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "request_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+w.config.APIKey)

	// Execute request
	httpResp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "network_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "read_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}

	// Handle error responses
	if httpResp.StatusCode != http.StatusOK {
		var errResp whisperErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, &Error{
				Provider: "openai-whisper",
				Type:     "stt",
				Code:     errResp.Error.Type,
				Message:  errResp.Error.Message,
				Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
			}
		}
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     fmt.Sprintf("http_%d", httpResp.StatusCode),
			Message:  string(respBody),
			Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
		}
	}

	// Parse response
	var apiResp whisperResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, &Error{
			Provider: "openai-whisper",
			Type:     "stt",
			Code:     "unmarshal_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	return &Transcription{
		Text:     apiResp.Text,
		Language: w.config.Language,
		Provider: "openai-whisper",
	}, nil
}

// Provider returns the provider name
func (w *WhisperSTT) Provider() string {
	return "openai-whisper"
}