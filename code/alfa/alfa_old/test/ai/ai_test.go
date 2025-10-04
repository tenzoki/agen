package ai_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"alfa/internal/ai"
)

func TestDefaultConfig(t *testing.T) {
	cfg := ai.DefaultConfig()

	if cfg.DefaultProvider != "anthropic" {
		t.Errorf("Expected default provider to be 'anthropic', got '%s'", cfg.DefaultProvider)
	}

	if _, ok := cfg.Providers["anthropic"]; !ok {
		t.Error("Expected anthropic provider to be present")
	}

	if _, ok := cfg.Providers["openai"]; !ok {
		t.Error("Expected openai provider to be present")
	}

	anthropicCfg := cfg.Providers["anthropic"]
	if anthropicCfg.Model == "" {
		t.Error("Expected anthropic model to be set")
	}
	if anthropicCfg.MaxTokens == 0 {
		t.Error("Expected anthropic max_tokens to be set")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a test config
	cfg := ai.DefaultConfig()
	cfg.DefaultProvider = "openai"
	cfg.Providers["anthropic"] = ai.Config{
		APIKey:      "test-key-123",
		Model:       "claude-3-opus-20240229",
		MaxTokens:   8000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
		RetryCount:  5,
		RetryDelay:  2 * time.Second,
	}

	// Save config
	if err := ai.SaveConfig(configPath, &cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := ai.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.DefaultProvider != "openai" {
		t.Errorf("Expected default provider 'openai', got '%s'", loadedCfg.DefaultProvider)
	}

	anthropicCfg := loadedCfg.Providers["anthropic"]
	if anthropicCfg.APIKey != "test-key-123" {
		t.Errorf("Expected API key 'test-key-123', got '%s'", anthropicCfg.APIKey)
	}
	if anthropicCfg.Model != "claude-3-opus-20240229" {
		t.Errorf("Expected model 'claude-3-opus-20240229', got '%s'", anthropicCfg.Model)
	}
	if anthropicCfg.MaxTokens != 8000 {
		t.Errorf("Expected max tokens 8000, got %d", anthropicCfg.MaxTokens)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Create a basic config file without API keys
	cfg := ai.DefaultConfig()
	if err := ai.SaveConfig(configPath, &cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Set environment variables
	os.Setenv("ANTHROPIC_API_KEY", "env-anthropic-key")
	os.Setenv("OPENAI_API_KEY", "env-openai-key")
	defer func() {
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	// Load config with env
	loadedCfg, err := ai.LoadConfigWithEnv(configPath)
	if err != nil {
		t.Fatalf("Failed to load config with env: %v", err)
	}

	anthropicCfg := loadedCfg.Providers["anthropic"]
	if anthropicCfg.APIKey != "env-anthropic-key" {
		t.Errorf("Expected API key from env 'env-anthropic-key', got '%s'", anthropicCfg.APIKey)
	}

	openaiCfg := loadedCfg.Providers["openai"]
	if openaiCfg.APIKey != "env-openai-key" {
		t.Errorf("Expected API key from env 'env-openai-key', got '%s'", openaiCfg.APIKey)
	}
}

func TestNewLLMFromConfig(t *testing.T) {
	cfg := ai.DefaultConfig()
	cfg.Providers["anthropic"] = ai.Config{
		APIKey:      "test-key",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   4096,
		Temperature: 1.0,
		Timeout:     60 * time.Second,
		RetryCount:  3,
		RetryDelay:  1 * time.Second,
	}

	// Test creating Claude client
	llm, err := ai.NewLLMFromConfig(&cfg, "anthropic")
	if err != nil {
		t.Fatalf("Failed to create LLM from config: %v", err)
	}

	if llm.Provider() != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got '%s'", llm.Provider())
	}

	if llm.Model() != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got '%s'", llm.Model())
	}
}

func TestNewLLMFromConfigMissingAPIKey(t *testing.T) {
	cfg := ai.DefaultConfig()
	cfg.Providers["anthropic"] = ai.Config{
		Model: "claude-3-5-sonnet-20241022",
		// APIKey is missing
	}

	_, err := ai.NewLLMFromConfig(&cfg, "anthropic")
	if err == nil {
		t.Error("Expected error when API key is missing")
	}
}

func TestNewLLMFromConfigInvalidProvider(t *testing.T) {
	cfg := ai.DefaultConfig()

	_, err := ai.NewLLMFromConfig(&cfg, "invalid-provider")
	if err == nil {
		t.Error("Expected error for invalid provider")
	}
}

func TestClaudeClientDefaults(t *testing.T) {
	client := ai.NewClaudeClient(ai.Config{})

	if client.Model() != "claude-3-5-sonnet-20241022" {
		t.Errorf("Expected default model 'claude-3-5-sonnet-20241022', got '%s'", client.Model())
	}

	if client.Provider() != "anthropic" {
		t.Errorf("Expected provider 'anthropic', got '%s'", client.Provider())
	}
}

func TestOpenAIClientDefaults(t *testing.T) {
	client := ai.NewOpenAIClient(ai.Config{})

	if client.Model() != "gpt-4" {
		t.Errorf("Expected default model 'gpt-4', got '%s'", client.Model())
	}

	if client.Provider() != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", client.Provider())
	}
}

func TestClaudeClientInvalidAPIKey(t *testing.T) {
	t.Skip("Skipping Anthropic API test - no API key available")

	client := ai.NewClaudeClient(ai.Config{
		APIKey: "invalid-key",
		Model:  "claude-3-5-sonnet-20241022",
	})

	ctx := context.Background()
	messages := []ai.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Error("Expected error with invalid API key")
	}

	// Check that it's an AI error
	if _, ok := err.(*ai.Error); !ok {
		t.Errorf("Expected *ai.Error, got %T", err)
	}
}

func TestOpenAIClientInvalidAPIKey(t *testing.T) {
	client := ai.NewOpenAIClient(ai.Config{
		APIKey: "invalid-key",
		Model:  "gpt-4",
	})

	ctx := context.Background()
	messages := []ai.Message{
		{Role: "user", Content: "Hello"},
	}

	_, err := client.Chat(ctx, messages)
	if err == nil {
		t.Error("Expected error with invalid API key")
	}

	// Check that it's an AI error
	if _, ok := err.(*ai.Error); !ok {
		t.Errorf("Expected *ai.Error, got %T", err)
	}
}

func TestErrorType(t *testing.T) {
	err := &ai.Error{
		Provider: "test-provider",
		Code:     "test-code",
		Message:  "test message",
		Retry:    true,
	}

	expected := "test-provider: test message"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}