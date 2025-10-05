package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	anthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
)

// ClaudeClient implements LLM interface for Anthropic's Claude API
type ClaudeClient struct {
	config     Config
	httpClient *http.Client
}

// NewClaudeClient creates a new Claude API client
func NewClaudeClient(config Config) *ClaudeClient {
	if config.Model == "" {
		config.Model = "claude-3-5-sonnet-20241022"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}
	if config.Temperature == 0 {
		config.Temperature = 1.0
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

	return &ClaudeClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// claudeMessage represents a message in Claude API format
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeRequest represents the request payload for Claude API
type claudeRequest struct {
	Model       string          `json:"model"`
	Messages    []claudeMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature,omitempty"`
	System      string          `json:"system,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// claudeResponse represents the response from Claude API
type claudeResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Role         string `json:"role"`
	Content      []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// claudeError represents an error response from Claude API
type claudeError struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends messages to Claude and returns the response
func (c *ClaudeClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
	startTime := time.Now()

	// Extract system message if present
	var systemMsg string
	var apiMessages []claudeMessage
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMsg = msg.Content
		} else {
			apiMessages = append(apiMessages, claudeMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	reqBody := claudeRequest{
		Model:       c.config.Model,
		Messages:    apiMessages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		System:      systemMsg,
		Stream:      false,
	}

	var resp *Response
	var err error

	// Retry loop
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.config.RetryDelay * time.Duration(attempt)):
			}
		}

		resp, err = c.makeRequest(ctx, reqBody)
		if err == nil {
			resp.ResponseTime = time.Since(startTime)
			resp.FinishTime = time.Now()
			return resp, nil
		}

		// Check if error is retryable
		if aiErr, ok := err.(*Error); ok && !aiErr.Retry {
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", c.config.RetryCount, err)
}

// makeRequest performs the actual HTTP request to Claude API
func (c *ClaudeClient) makeRequest(ctx context.Context, reqBody claudeRequest) (*Response, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &Error{
			Provider: "anthropic",
			Code:     "marshal_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &Error{
			Provider: "anthropic",
			Code:     "request_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Provider: "anthropic",
			Code:     "network_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &Error{
			Provider: "anthropic",
			Code:     "read_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}

	// Handle error responses
	if httpResp.StatusCode != http.StatusOK {
		var errResp claudeError
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, &Error{
				Provider: "anthropic",
				Code:     errResp.Error.Type,
				Message:  errResp.Error.Message,
				Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
			}
		}
		return nil, &Error{
			Provider: "anthropic",
			Code:     fmt.Sprintf("http_%d", httpResp.StatusCode),
			Message:  string(body),
			Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
		}
	}

	var apiResp claudeResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &Error{
			Provider: "anthropic",
			Code:     "unmarshal_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	// Extract text content
	var content string
	for _, c := range apiResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &Response{
		Content:    content,
		Model:      apiResp.Model,
		StopReason: apiResp.StopReason,
		Usage: Usage{
			InputTokens:  apiResp.Usage.InputTokens,
			OutputTokens: apiResp.Usage.OutputTokens,
			TotalTokens:  apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		},
	}, nil
}

// ChatStream implements streaming chat (not yet implemented)
func (c *ClaudeClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, <-chan error) {
	contentChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errChan)
		errChan <- &Error{
			Provider: "anthropic",
			Code:     "not_implemented",
			Message:  "streaming not yet implemented",
			Retry:    false,
		}
	}()

	return contentChan, errChan
}

// Model returns the model identifier
func (c *ClaudeClient) Model() string {
	return c.config.Model
}

// Provider returns the provider name
func (c *ClaudeClient) Provider() string {
	return "anthropic"
}