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

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

// OpenAIClient implements LLM interface for OpenAI's GPT API
type OpenAIClient struct {
	config     Config
	httpClient *http.Client
}

// NewOpenAIClient creates a new OpenAI API client
func NewOpenAIClient(config Config) *OpenAIClient {
	if config.Model == "" {
		config.Model = "gpt-4"
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

	return &OpenAIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// openaiMessage represents a message in OpenAI API format
type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openaiRequest represents the request payload for OpenAI API
type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// openaiResponse represents the response from OpenAI API
type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// openaiError represents an error response from OpenAI API
type openaiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Chat sends messages to OpenAI and returns the response
func (o *OpenAIClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
	startTime := time.Now()

	// Convert messages to OpenAI format
	var apiMessages []openaiMessage
	for _, msg := range messages {
		apiMessages = append(apiMessages, openaiMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	reqBody := openaiRequest{
		Model:       o.config.Model,
		Messages:    apiMessages,
		MaxTokens:   o.config.MaxTokens,
		Temperature: o.config.Temperature,
		Stream:      false,
	}

	var resp *Response
	var err error

	// Retry loop
	for attempt := 0; attempt <= o.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(o.config.RetryDelay * time.Duration(attempt)):
			}
		}

		resp, err = o.makeRequest(ctx, reqBody)
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

	return nil, fmt.Errorf("failed after %d retries: %w", o.config.RetryCount, err)
}

// makeRequest performs the actual HTTP request to OpenAI API
func (o *OpenAIClient) makeRequest(ctx context.Context, reqBody openaiRequest) (*Response, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, &Error{
			Provider: "openai",
			Code:     "marshal_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openaiAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &Error{
			Provider: "openai",
			Code:     "request_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	httpResp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, &Error{
			Provider: "openai",
			Code:     "network_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, &Error{
			Provider: "openai",
			Code:     "read_error",
			Message:  err.Error(),
			Retry:    true,
		}
	}

	// Handle error responses
	if httpResp.StatusCode != http.StatusOK {
		var errResp openaiError
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, &Error{
				Provider: "openai",
				Code:     errResp.Error.Type,
				Message:  errResp.Error.Message,
				Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
			}
		}
		return nil, &Error{
			Provider: "openai",
			Code:     fmt.Sprintf("http_%d", httpResp.StatusCode),
			Message:  string(body),
			Retry:    httpResp.StatusCode >= 500 || httpResp.StatusCode == 429,
		}
	}

	var apiResp openaiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, &Error{
			Provider: "openai",
			Code:     "unmarshal_error",
			Message:  err.Error(),
			Retry:    false,
		}
	}

	// Extract content from first choice
	if len(apiResp.Choices) == 0 {
		return nil, &Error{
			Provider: "openai",
			Code:     "no_choices",
			Message:  "API returned no choices",
			Retry:    false,
		}
	}

	choice := apiResp.Choices[0]

	return &Response{
		Content:    choice.Message.Content,
		Model:      apiResp.Model,
		StopReason: choice.FinishReason,
		Usage: Usage{
			InputTokens:  apiResp.Usage.PromptTokens,
			OutputTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:  apiResp.Usage.TotalTokens,
		},
	}, nil
}

// ChatStream implements streaming chat (not yet implemented)
func (o *OpenAIClient) ChatStream(ctx context.Context, messages []Message) (<-chan string, <-chan error) {
	contentChan := make(chan string)
	errChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errChan)
		errChan <- &Error{
			Provider: "openai",
			Code:     "not_implemented",
			Message:  "streaming not yet implemented",
			Retry:    false,
		}
	}()

	return contentChan, errChan
}

// Model returns the model identifier
func (o *OpenAIClient) Model() string {
	return o.config.Model
}

// Provider returns the provider name
func (o *OpenAIClient) Provider() string {
	return "openai"
}