package speech

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const ttsAPIURL = "https://api.openai.com/v1/audio/speech"

// OpenAITTS implements TTS interface using OpenAI's TTS API
type OpenAITTS struct {
	config     TTSConfig
	httpClient *http.Client
}

// NewOpenAITTS creates a new OpenAI TTS client
func NewOpenAITTS(config TTSConfig) *OpenAITTS {
	if config.Model == "" {
		config.Model = "tts-1"
	}
	if config.Voice == "" {
		config.Voice = "alloy"
	}
	if config.Speed == 0 {
		config.Speed = 1.0
	}
	if config.Format == "" {
		config.Format = "mp3"
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

	return &OpenAITTS{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// ttsRequest represents the request payload for TTS API
type ttsRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

// ttsErrorResponse represents an error response from TTS API
type ttsErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Synthesize converts text to audio
func (t *OpenAITTS) Synthesize(ctx context.Context, text string) (io.ReadCloser, error) {
	var result io.ReadCloser
	var err error

	// Retry loop
	for attempt := 0; attempt <= t.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(t.config.RetryDelay * time.Duration(attempt)):
			}
		}

		result, err = t.makeRequest(ctx, text)
		if err == nil {
			return result, nil
		}

		// Check if error is retryable
		if speechErr, ok := err.(*Error); ok && !speechErr.Retry {
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", t.config.RetryCount, err)
}

// SynthesizeToFile converts text to audio and saves to file
func (t *OpenAITTS) SynthesizeToFile(ctx context.Context, text string, outputPath string) error {
	audio, err := t.Synthesize(ctx, text)
	if err != nil {
		return err
	}
	defer audio.Close()

	outFile, err := os.Create(outputPath)
	if err != nil {
		return &Error{
			Provider: "openai-tts",
			Type:     "tts",
			Code:     "file_error",
			Message:  fmt.Sprintf("failed to create output file: %v", err),
			Retry:    false,
		}
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, audio); err != nil {
		return &Error{
			Provider: "openai-tts",
			Type:     "tts",
			Code:     "write_error",
			Message:  fmt.Sprintf("failed to write audio data: %v", err),
			Retry:    false,
		}
	}

	return nil
}

// makeRequest performs the actual HTTP request to TTS API
func (t *OpenAITTS) makeRequest(ctx context.Context, text string) (io.ReadCloser, error) {
	reqBody := ttsRequest{
		Model:          t.config.Model,
		Input:          text,
		Voice:          t.config.Voice,
		ResponseFormat: t.config.Format,
		Speed:          t.config.Speed,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &Error{
			Provider: "openai-tts",
			Type:     "tts",
			Code:     "marshal_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ttsAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &Error{
			Provider: "openai-tts",
			Type:     "tts",
			Code:     "request_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.config.APIKey)

	httpResp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Provider: "openai-tts",
			Type:     "tts",
			Code:     "network_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}

	// Check for error response
	if httpResp.StatusCode != http.StatusOK {
		defer httpResp.Body.Close()
		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, &Error{
				Provider: "openai-tts",
				Type:     "tts",
				Code:     "read_error",
				Message:  err.Error(),
				Retry:    true,
			}
		}

		var errResp ttsErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return nil, &Error{
				Provider: "openai-tts",
				Type:     "tts",
				Code:     errResp.Error.Type,
				Message:  errResp.Error.Message,
				Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
			}
		}
		return nil, &Error{
			Provider: "openai-tts",
			Type:     "tts",
			Code:     fmt.Sprintf("http_%d", httpResp.StatusCode),
			Message:  string(respBody),
			Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
		}
	}

	// Return the audio stream (caller is responsible for closing)
	return httpResp.Body, nil
}

// Provider returns the provider name
func (t *OpenAITTS) Provider() string {
	return "openai-tts"
}