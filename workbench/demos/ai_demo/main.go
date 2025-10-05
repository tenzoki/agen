package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tenzoki/agen/alfa/internal/ai"
)

func main() {
	fmt.Println("=== AI Layer Demo ===\n")

	// Example 1: Load configuration from file with environment variables
	fmt.Println("1. Loading configuration from config/ai-config.json...")
	cfg, err := ai.LoadConfigWithEnv(ai.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("   Default provider: %s\n\n", cfg.DefaultProvider)

	// Example 2: Create LLM client from config (uses default provider)
	fmt.Println("2. Creating LLM client from config...")
	llm, err := ai.NewLLMFromConfig(cfg, "") // Empty string uses default provider
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}
	fmt.Printf("   Provider: %s\n", llm.Provider())
	fmt.Printf("   Model: %s\n\n", llm.Model())

	// Example 3: Simple chat with the LLM
	fmt.Println("3. Sending a simple chat request...")
	ctx := context.Background()
	messages := []ai.Message{
		{Role: "user", Content: "What is the capital of France? Answer in one word."},
	}

	response, err := llm.Chat(ctx, messages)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}

	fmt.Printf("   Response: %s\n", response.Content)
	fmt.Printf("   Model used: %s\n", response.Model)
	fmt.Printf("   Stop reason: %s\n", response.StopReason)
	fmt.Printf("   Tokens - Input: %d, Output: %d, Total: %d\n",
		response.Usage.InputTokens,
		response.Usage.OutputTokens,
		response.Usage.TotalTokens)
	fmt.Printf("   Response time: %v\n\n", response.ResponseTime)

	// Example 4: Multi-turn conversation
	fmt.Println("4. Multi-turn conversation...")
	conversation := []ai.Message{
		{Role: "user", Content: "I'm learning Go. What's a good first project?"},
	}

	resp1, err := llm.Chat(ctx, conversation)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	fmt.Printf("   Assistant: %s\n\n", resp1.Content[:100]+"...") // First 100 chars

	// Continue the conversation
	conversation = append(conversation,
		ai.Message{Role: "assistant", Content: resp1.Content},
		ai.Message{Role: "user", Content: "How long would that take to build?"},
	)

	resp2, err := llm.Chat(ctx, conversation)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	fmt.Printf("   Assistant: %s\n\n", resp2.Content[:100]+"...")

	// Example 5: Using system messages (Claude specific)
	fmt.Println("5. Using system messages...")
	systemMessages := []ai.Message{
		{Role: "system", Content: "You are a helpful coding assistant. Keep responses very brief."},
		{Role: "user", Content: "Explain what a goroutine is in one sentence."},
	}

	resp3, err := llm.Chat(ctx, systemMessages)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	fmt.Printf("   Response: %s\n\n", resp3.Content)

	// Example 6: Create a specific provider client directly
	fmt.Println("6. Creating Claude client directly...")
	claudeClient := ai.NewClaudeClient(ai.Config{
		APIKey:      cfg.Providers["anthropic"].APIKey,
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
		RetryCount:  2,
		RetryDelay:  500 * time.Millisecond,
	})

	resp4, err := claudeClient.Chat(ctx, []ai.Message{
		{Role: "user", Content: "Say 'Hello from Claude!'"},
	})
	if err != nil {
		log.Fatalf("Claude chat failed: %v", err)
	}
	fmt.Printf("   %s: %s\n\n", claudeClient.Provider(), resp4.Content)

	// Example 7: Create OpenAI client directly
	fmt.Println("7. Creating OpenAI client directly...")
	openaiClient := ai.NewOpenAIClient(ai.Config{
		APIKey:      cfg.Providers["openai"].APIKey,
		Model:       "gpt-4",
		MaxTokens:   1000,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
		RetryCount:  2,
		RetryDelay:  500 * time.Millisecond,
	})

	resp5, err := openaiClient.Chat(ctx, []ai.Message{
		{Role: "user", Content: "Say 'Hello from OpenAI!'"},
	})
	if err != nil {
		log.Fatalf("OpenAI chat failed: %v", err)
	}
	fmt.Printf("   %s: %s\n\n", openaiClient.Provider(), resp5.Content)

	// Example 8: Error handling
	fmt.Println("8. Demonstrating error handling...")
	badClient := ai.NewClaudeClient(ai.Config{
		APIKey: "invalid-key",
		Model:  "claude-3-5-sonnet-20241022",
	})

	_, err = badClient.Chat(ctx, []ai.Message{
		{Role: "user", Content: "This will fail"},
	})
	if err != nil {
		if aiErr, ok := err.(*ai.Error); ok {
			fmt.Printf("   Error type: *ai.Error\n")
			fmt.Printf("   Provider: %s\n", aiErr.Provider)
			fmt.Printf("   Code: %s\n", aiErr.Code)
			fmt.Printf("   Message: %s\n", aiErr.Message)
			fmt.Printf("   Retryable: %v\n\n", aiErr.Retry)
		}
	}

	// Example 9: Using context with timeout
	fmt.Println("9. Using context with timeout...")
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp6, err := llm.Chat(ctxWithTimeout, []ai.Message{
		{Role: "user", Content: "Quick question: what's 2+2?"},
	})
	if err != nil {
		log.Fatalf("Chat with timeout failed: %v", err)
	}
	fmt.Printf("   Response: %s\n\n", resp6.Content)

	// Example 10: Saving configuration
	fmt.Println("10. Saving custom configuration...")
	customCfg := ai.DefaultConfig()
	customCfg.DefaultProvider = "openai"
	customCfg.Providers["openai"] = ai.Config{
		Model:       "gpt-4-turbo",
		MaxTokens:   8000,
		Temperature: 0.5,
		Timeout:     120 * time.Second,
		RetryCount:  5,
		RetryDelay:  2 * time.Second,
	}

	if err := ai.SaveConfig("config/custom-config.json", &customCfg); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}
	fmt.Println("   Configuration saved to config/custom-config.json")

	fmt.Println("\n=== Demo Complete ===")
}