package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tenzoki/agen/alfa/internal/ai"
	"github.com/tenzoki/agen/alfa/internal/speech"
)

func main() {
	fmt.Println("=== Speech + AI Integration Demo ===\n")

	// Load configurations
	fmt.Println("1. Loading configurations...")
	aiCfg, err := ai.LoadConfigWithEnv(ai.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load AI config: %v", err)
	}

	speechCfg, err := speech.LoadConfigWithEnv(speech.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load speech config: %v", err)
	}

	fmt.Printf("   AI Provider: %s (%s)\n", aiCfg.DefaultProvider, aiCfg.Providers[aiCfg.DefaultProvider].Model)
	fmt.Printf("   TTS: %s (voice: %s)\n", speechCfg.TTS.Model, speechCfg.TTS.Voice)
	fmt.Printf("   STT: %s\n\n", speechCfg.STT.Model)

	// Check if API keys are set
	if speechCfg.STT.APIKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY not set for speech APIs.")
		fmt.Println("   Set it to run the demo: export OPENAI_API_KEY='your-key-here'\n")
		return
	}

	// Create clients
	fmt.Println("2. Creating AI and speech clients...")
	llm, err := ai.NewLLMFromConfig(aiCfg, "")
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	tts, err := speech.NewTTSFromConfig(speechCfg)
	if err != nil {
		log.Fatalf("Failed to create TTS client: %v", err)
	}

	stt, err := speech.NewSTTFromConfig(speechCfg)
	if err != nil {
		log.Fatalf("Failed to create STT client: %v", err)
	}

	fmt.Printf("   LLM: %s\n", llm.Provider())
	fmt.Printf("   TTS: %s\n", tts.Provider())
	fmt.Printf("   STT: %s\n\n", stt.Provider())

	ctx := context.Background()

	// Create output directory for audio files
	outputDir := "demo/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Demonstration: Voice-based Q&A loop
	fmt.Println("=== Voice-based Q&A Demo ===\n")

	questions := []string{
		"What is the capital of France?",
		"Explain recursion in one sentence.",
		"What's the meaning of life?",
	}

	for i, question := range questions {
		fmt.Printf("Question %d: %s\n", i+1, question)

		// Step 1: Synthesize the question to audio
		questionAudio := filepath.Join(outputDir, fmt.Sprintf("question_%d.mp3", i+1))
		fmt.Printf("   → Synthesizing question to audio...\n")
		err := tts.SynthesizeToFile(ctx, question, questionAudio)
		if err != nil {
			log.Printf("   ⚠️  TTS failed: %v\n", err)
			continue
		}
		fmt.Printf("   ✓ Question audio: %s\n", questionAudio)

		// Step 2: Transcribe the question back (simulating user speaking)
		fmt.Printf("   → Transcribing audio to text...\n")
		transcription, err := stt.TranscribeFile(ctx, questionAudio)
		if err != nil {
			log.Printf("   ⚠️  STT failed: %v\n", err)
			continue
		}
		fmt.Printf("   ✓ Transcribed: \"%s\"\n", transcription.Text)

		// Step 3: Send transcribed text to AI
		fmt.Printf("   → Asking AI...\n")
		messages := []ai.Message{
			{Role: "user", Content: transcription.Text},
		}

		response, err := llm.Chat(ctx, messages)
		if err != nil {
			log.Printf("   ⚠️  AI failed: %v\n", err)
			continue
		}
		fmt.Printf("   ✓ AI Response: %s\n", response.Content)
		fmt.Printf("     (Tokens: %d input, %d output, %v latency)\n",
			response.Usage.InputTokens,
			response.Usage.OutputTokens,
			response.ResponseTime)

		// Step 4: Synthesize AI response to audio
		fmt.Printf("   → Synthesizing AI response to audio...\n")
		answerAudio := filepath.Join(outputDir, fmt.Sprintf("answer_%d.mp3", i+1))
		err = tts.SynthesizeToFile(ctx, response.Content, answerAudio)
		if err != nil {
			log.Printf("   ⚠️  TTS failed: %v\n", err)
			continue
		}
		fmt.Printf("   ✓ Answer audio: %s\n\n", answerAudio)
	}

	// Demonstration: Multi-turn conversation with voice
	fmt.Println("=== Multi-turn Voice Conversation Demo ===\n")

	conversation := []ai.Message{
		{Role: "system", Content: "You are a helpful assistant. Keep your responses concise."},
	}

	turns := []string{
		"Tell me a programming joke.",
		"Explain why that's funny.",
		"Tell me another one.",
	}

	for i, userInput := range turns {
		fmt.Printf("Turn %d\n", i+1)
		fmt.Printf("   User: %s\n", userInput)

		// User speaks -> TTS
		userAudio := filepath.Join(outputDir, fmt.Sprintf("conversation_user_%d.mp3", i+1))
		fmt.Printf("   → Synthesizing user speech...\n")
		err := tts.SynthesizeToFile(ctx, userInput, userAudio)
		if err != nil {
			log.Printf("   ⚠️  Failed: %v\n", err)
			continue
		}

		// STT -> Text
		fmt.Printf("   → Transcribing user speech...\n")
		transcription, err := stt.TranscribeFile(ctx, userAudio)
		if err != nil {
			log.Printf("   ⚠️  Failed: %v\n", err)
			continue
		}

		// Add to conversation
		conversation = append(conversation, ai.Message{
			Role:    "user",
			Content: transcription.Text,
		})

		// AI responds
		fmt.Printf("   → AI processing...\n")
		response, err := llm.Chat(ctx, conversation)
		if err != nil {
			log.Printf("   ⚠️  Failed: %v\n", err)
			continue
		}

		fmt.Printf("   Assistant: %s\n", response.Content)

		// Add AI response to conversation
		conversation = append(conversation, ai.Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// AI speaks -> TTS
		assistantAudio := filepath.Join(outputDir, fmt.Sprintf("conversation_assistant_%d.mp3", i+1))
		fmt.Printf("   → Synthesizing assistant speech...\n")
		err = tts.SynthesizeToFile(ctx, response.Content, assistantAudio)
		if err != nil {
			log.Printf("   ⚠️  Failed: %v\n", err)
		} else {
			fmt.Printf("   ✓ Audio saved: %s\n", assistantAudio)
		}

		fmt.Println()
	}

	// Demonstration: Code explanation with voice
	fmt.Println("=== Code Explanation with Voice Demo ===\n")

	codeSnippet := `func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}`

	fmt.Println("Code:")
	fmt.Println(codeSnippet)
	fmt.Println()

	codeQuestion := "Explain this Go code and suggest an improvement."
	fmt.Printf("User asks: \"%s\"\n", codeQuestion)

	// Synthesize question
	codeQuestionAudio := filepath.Join(outputDir, "code_question.mp3")
	err = tts.SynthesizeToFile(ctx, codeQuestion, codeQuestionAudio)
	if err != nil {
		log.Printf("⚠️  TTS failed: %v\n", err)
	}

	// AI analyzes code
	messages := []ai.Message{
		{Role: "system", Content: "You are a helpful programming tutor. Be concise but thorough."},
		{Role: "user", Content: fmt.Sprintf("Here's some code:\n\n%s\n\n%s", codeSnippet, codeQuestion)},
	}

	response, err := llm.Chat(ctx, messages)
	if err != nil {
		log.Fatalf("AI failed: %v", err)
	}

	fmt.Printf("\nAI Explanation:\n%s\n\n", response.Content)

	// Synthesize explanation
	explanationAudio := filepath.Join(outputDir, "code_explanation.mp3")
	fmt.Println("→ Synthesizing explanation to audio...")
	err = tts.SynthesizeToFile(ctx, response.Content, explanationAudio)
	if err != nil {
		log.Printf("⚠️  TTS failed: %v\n", err)
	} else {
		fmt.Printf("✓ Explanation audio: %s\n", explanationAudio)
	}

	fmt.Println("\n=== Demo Complete ===")
	fmt.Printf("\n📁 Audio files saved to: %s\n", outputDir)
	fmt.Println("\nThis demonstrates the complete pipeline:")
	fmt.Println("  User speaks → STT → AI processes → TTS → User hears response")
}