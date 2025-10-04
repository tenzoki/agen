package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"alfa/internal/ai"
	"alfa/internal/audio"
	"alfa/internal/speech"
)

func main() {
	fmt.Println("=== Interactive Voice Assistant Demo ===\n")

	// Check for sox
	recorder := audio.NewSoxRecorder(audio.DefaultRecordConfig())
	if !recorder.IsAvailable() {
		fmt.Println("⚠️  Sox not found. Please install it:")
		fmt.Println("   brew install sox")
		return
	}

	player := audio.NewSoxPlayer()
	if !player.IsAvailable() {
		fmt.Println("⚠️  Audio player not found. Please install sox:")
		fmt.Println("   brew install sox")
		return
	}

	fmt.Println("✓ Audio recording/playback available\n")

	// Load configurations
	fmt.Println("Loading configurations...")
	aiCfg, err := ai.LoadConfigWithEnv(ai.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load AI config: %v", err)
	}

	speechCfg, err := speech.LoadConfigWithEnv(speech.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load speech config: %v", err)
	}

	// Check API keys
	if speechCfg.STT.APIKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY not set.")
		fmt.Println("   Set it to use voice features: export OPENAI_API_KEY='your-key-here'")
		return
	}

	// Create clients
	fmt.Println("Creating AI and speech clients...")
	llm, err := ai.NewLLMFromConfig(aiCfg, "")
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	tts, err := speech.NewTTSFromConfig(speechCfg)
	if err != nil {
		log.Fatalf("Failed to create TTS: %v", err)
	}

	stt, err := speech.NewSTTFromConfig(speechCfg)
	if err != nil {
		log.Fatalf("Failed to create STT: %v", err)
	}

	fmt.Printf("✓ LLM: %s (%s)\n", llm.Provider(), llm.Model())
	fmt.Printf("✓ TTS: %s\n", tts.Provider())
	fmt.Printf("✓ STT: %s\n\n", stt.Provider())

	// Create output directory
	outputDir := "demo/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)

	// Conversation history
	conversation := []ai.Message{
		{Role: "system", Content: "You are a helpful voice assistant. Keep responses concise and natural for spoken conversation."},
	}

	fmt.Println("=== Voice Assistant Ready ===")
	fmt.Println("\nCommands:")
	fmt.Println("  [Enter]  - Record your question (stops after 2 seconds of silence)")
	fmt.Println("  'text'   - Type a question instead of speaking")
	fmt.Println("  'quit'   - Exit")
	fmt.Println()

	turnNumber := 0

	for {
		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" || input == "exit" {
			fmt.Println("\nGoodbye!")
			break
		}

		turnNumber++
		var userText string

		if input == "" {
			// Voice input mode
			fmt.Println("\n🎤 Recording... (speak now, will auto-stop after 2s silence)")

			recordPath := filepath.Join(outputDir, fmt.Sprintf("user_%d.wav", turnNumber))
			err := recorder.RecordUntilSilence(recordPath, 30*time.Second)
			if err != nil {
				log.Printf("❌ Recording failed: %v\n", err)
				continue
			}

			fmt.Println("✓ Recording complete")

			// Transcribe
			fmt.Println("🔄 Transcribing...")
			transcription, err := stt.TranscribeFile(ctx, recordPath)
			if err != nil {
				log.Printf("❌ Transcription failed: %v\n", err)
				continue
			}

			userText = transcription.Text
			fmt.Printf("📝 You said: \"%s\"\n", userText)

		} else if strings.HasPrefix(input, "text ") {
			// Text input mode
			userText = strings.TrimPrefix(input, "text ")
			fmt.Printf("📝 Text input: \"%s\"\n", userText)
		} else {
			// Treat any other input as text
			userText = input
			fmt.Printf("📝 Text input: \"%s\"\n", userText)
		}

		if userText == "" {
			fmt.Println("⚠️  No input detected")
			continue
		}

		// Add to conversation
		conversation = append(conversation, ai.Message{
			Role:    "user",
			Content: userText,
		})

		// Get AI response
		fmt.Println("🤖 AI thinking...")
		response, err := llm.Chat(ctx, conversation)
		if err != nil {
			log.Printf("❌ AI failed: %v\n", err)
			continue
		}

		fmt.Printf("💬 Assistant: %s\n", response.Content)
		fmt.Printf("   (tokens: %d in, %d out, %v latency)\n",
			response.Usage.InputTokens,
			response.Usage.OutputTokens,
			response.ResponseTime)

		// Add AI response to conversation
		conversation = append(conversation, ai.Message{
			Role:    "assistant",
			Content: response.Content,
		})

		// Synthesize and play response
		fmt.Println("🔊 Synthesizing speech...")
		responsePath := filepath.Join(outputDir, fmt.Sprintf("assistant_%d.mp3", turnNumber))
		err = tts.SynthesizeToFile(ctx, response.Content, responsePath)
		if err != nil {
			log.Printf("❌ TTS failed: %v\n", err)
			continue
		}

		fmt.Println("▶️  Playing response...")
		err = player.Play(responsePath)
		if err != nil {
			log.Printf("❌ Playback failed: %v\n", err)
		}

		fmt.Println("✓ Response complete")
	}

	fmt.Printf("\n📁 Audio files saved to: %s\n", outputDir)
	fmt.Println("Session complete.")
}