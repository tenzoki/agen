package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tenzoki/agen/alfa/internal/speech"
)

func main() {
	fmt.Println("=== Speech Layer Demo ===\n")

	// Example 1: Load configuration
	fmt.Println("1. Loading speech configuration...")
	cfg, err := speech.LoadConfigWithEnv(speech.GetConfigPath())
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("   STT Model: %s\n", cfg.STT.Model)
	fmt.Printf("   TTS Model: %s, Voice: %s\n\n", cfg.TTS.Model, cfg.TTS.Voice)

	// Check if API key is set
	if cfg.STT.APIKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY not set. Set it to run live API tests.")
		fmt.Println("   Example: export OPENAI_API_KEY='your-key-here'\n")
		fmt.Println("Showing configuration examples without making API calls...\n")
		showConfigurationExamples()
		return
	}

	ctx := context.Background()

	// Example 2: Create TTS client and synthesize speech
	fmt.Println("2. Creating TTS client and synthesizing speech...")
	tts, err := speech.NewTTSFromConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create TTS client: %v", err)
	}
	fmt.Printf("   Provider: %s\n", tts.Provider())

	// Create output directory for audio files
	outputDir := "demo/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	outputPath := filepath.Join(outputDir, "alfa_speech_demo.mp3")

	text := "Hello! This is a demonstration of the speech layer in the Alfa AI workbench."
	fmt.Printf("   Synthesizing: \"%s\"\n", text)

	err = tts.SynthesizeToFile(ctx, text, outputPath)
	if err != nil {
		log.Fatalf("Failed to synthesize speech: %v", err)
	}
	fmt.Printf("   ✓ Audio saved to: %s\n\n", outputPath)

	// Example 3: Create STT client and transcribe the audio we just created
	fmt.Println("3. Creating STT client and transcribing audio...")
	stt, err := speech.NewSTTFromConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create STT client: %v", err)
	}
	fmt.Printf("   Provider: %s\n", stt.Provider())

	transcription, err := stt.TranscribeFile(ctx, outputPath)
	if err != nil {
		log.Fatalf("Failed to transcribe audio: %v", err)
	}

	fmt.Printf("   Transcription: \"%s\"\n", transcription.Text)
	fmt.Printf("   Provider: %s\n", transcription.Provider)
	fmt.Printf("   Duration: %v\n\n", transcription.Duration)

	// Example 4: Different TTS voices
	fmt.Println("4. Testing different TTS voices...")
	voices := []string{"alloy", "echo", "fable", "onyx", "nova", "shimmer"}

	for _, voice := range voices {
		fmt.Printf("   Generating with voice '%s'...\n", voice)

		customTTS := speech.NewOpenAITTS(speech.TTSConfig{
			APIKey: cfg.TTS.APIKey,
			Model:  "tts-1",
			Voice:  voice,
			Speed:  1.0,
			Format: "mp3",
		})

		voiceOutput := filepath.Join(outputDir, fmt.Sprintf("demo_voice_%s.mp3", voice))
		err := customTTS.SynthesizeToFile(ctx, fmt.Sprintf("This is the %s voice.", voice), voiceOutput)
		if err != nil {
			fmt.Printf("   ⚠️  Failed: %v\n", err)
		} else {
			fmt.Printf("   ✓ Saved to: %s\n", voiceOutput)
		}
	}
	fmt.Println()

	// Example 5: Different speech speeds
	fmt.Println("5. Testing different speech speeds...")
	speeds := []float64{0.5, 1.0, 1.5, 2.0}

	for _, speed := range speeds {
		fmt.Printf("   Generating at speed %.1fx...\n", speed)

		customTTS := speech.NewOpenAITTS(speech.TTSConfig{
			APIKey: cfg.TTS.APIKey,
			Model:  "tts-1",
			Voice:  "nova",
			Speed:  speed,
			Format: "mp3",
		})

		speedOutput := filepath.Join(outputDir, fmt.Sprintf("demo_speed_%.1f.mp3", speed))
		err := customTTS.SynthesizeToFile(ctx, "Testing different speech speeds.", speedOutput)
		if err != nil {
			fmt.Printf("   ⚠️  Failed: %v\n", err)
		} else {
			fmt.Printf("   ✓ Saved to: %s\n", speedOutput)
		}
	}
	fmt.Println()

	// Example 6: Different audio formats
	fmt.Println("6. Testing different audio formats...")
	formats := []string{"mp3", "opus", "aac", "flac"}

	for _, format := range formats {
		fmt.Printf("   Generating %s format...\n", format)

		customTTS := speech.NewOpenAITTS(speech.TTSConfig{
			APIKey: cfg.TTS.APIKey,
			Model:  "tts-1",
			Voice:  "alloy",
			Speed:  1.0,
			Format: format,
		})

		formatOutput := filepath.Join(outputDir, fmt.Sprintf("demo_format.%s", format))
		err := customTTS.SynthesizeToFile(ctx, "Testing audio format.", formatOutput)
		if err != nil {
			fmt.Printf("   ⚠️  Failed: %v\n", err)
		} else {
			fmt.Printf("   ✓ Saved to: %s\n", formatOutput)
		}
	}
	fmt.Println()

	// Example 7: Transcription with language hint
	fmt.Println("7. Testing STT with language hint...")
	germanSTT := speech.NewWhisperSTT(speech.STTConfig{
		APIKey:   cfg.STT.APIKey,
		Model:    "whisper-1",
		Language: "de", // German language hint
	})

	germanTTS := speech.NewOpenAITTS(speech.TTSConfig{
		APIKey: cfg.TTS.APIKey,
		Model:  "tts-1",
		Voice:  "alloy",
	})

	germanAudioPath := filepath.Join(outputDir, "demo_german.mp3")
	germanText := "Guten Tag! Dies ist ein Test."
	fmt.Printf("   Synthesizing German text: \"%s\"\n", germanText)

	err = germanTTS.SynthesizeToFile(ctx, germanText, germanAudioPath)
	if err != nil {
		log.Printf("   ⚠️  Failed to synthesize: %v", err)
	} else {
		fmt.Printf("   ✓ Audio saved\n")

		transcription, err := germanSTT.TranscribeFile(ctx, germanAudioPath)
		if err != nil {
			log.Printf("   ⚠️  Failed to transcribe: %v", err)
		} else {
			fmt.Printf("   Transcription: \"%s\"\n", transcription.Text)
		}
	}
	fmt.Println()

	// Example 8: Save custom configuration
	fmt.Println("8. Saving custom configuration...")
	customCfg := speech.DefaultConfig()
	customCfg.TTS.Voice = "nova"
	customCfg.TTS.Speed = 1.2
	customCfg.TTS.Model = "tts-1-hd"

	customConfigPath := "config/custom-speech-config.json"
	if err := speech.SaveConfig(customConfigPath, &customCfg); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}
	fmt.Printf("   ✓ Configuration saved to %s\n\n", customConfigPath)

	fmt.Println("=== Demo Complete ===")
	fmt.Printf("\n📁 Audio files saved to: %s\n", outputDir)
	fmt.Println("   You can play them with: mpg123, ffplay, or any audio player")
}

func showConfigurationExamples() {
	fmt.Println("Configuration Examples:\n")

	fmt.Println("Creating TTS client:")
	fmt.Println(`  tts := speech.NewOpenAITTS(speech.TTSConfig{
    APIKey: "your-api-key",
    Model:  "tts-1",
    Voice:  "alloy",
    Speed:  1.0,
    Format: "mp3",
  })`)
	fmt.Println()

	fmt.Println("Synthesizing speech:")
	fmt.Println(`  err := tts.SynthesizeToFile(ctx, "Hello world", "output.mp3")`)
	fmt.Println()

	fmt.Println("Creating STT client:")
	fmt.Println(`  stt := speech.NewWhisperSTT(speech.STTConfig{
    APIKey: "your-api-key",
    Model:  "whisper-1",
  })`)
	fmt.Println()

	fmt.Println("Transcribing audio:")
	fmt.Println(`  transcription, err := stt.TranscribeFile(ctx, "audio.mp3")
  fmt.Println(transcription.Text)`)
	fmt.Println()

	fmt.Println("Available TTS voices:")
	fmt.Println("  - alloy   (neutral)")
	fmt.Println("  - echo    (male)")
	fmt.Println("  - fable   (expressive)")
	fmt.Println("  - onyx    (deep male)")
	fmt.Println("  - nova    (female)")
	fmt.Println("  - shimmer (soft female)")
	fmt.Println()

	fmt.Println("Audio formats:")
	fmt.Println("  - mp3  (default)")
	fmt.Println("  - opus (low latency)")
	fmt.Println("  - aac  (compatibility)")
	fmt.Println("  - flac (lossless)")
	fmt.Println("  - wav  (uncompressed)")
	fmt.Println("  - pcm  (raw)")
}