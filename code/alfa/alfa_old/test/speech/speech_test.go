package speech_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"alfa/internal/speech"
)

func TestDefaultConfig(t *testing.T) {
	cfg := speech.DefaultConfig()

	if cfg.STT.Model != "whisper-1" {
		t.Errorf("Expected STT model 'whisper-1', got '%s'", cfg.STT.Model)
	}

	if cfg.TTS.Model != "tts-1" {
		t.Errorf("Expected TTS model 'tts-1', got '%s'", cfg.TTS.Model)
	}

	if cfg.TTS.Voice != "alloy" {
		t.Errorf("Expected TTS voice 'alloy', got '%s'", cfg.TTS.Voice)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "speech-config.json")

	// Create a test config
	cfg := speech.DefaultConfig()
	cfg.STT.Model = "whisper-1"
	cfg.STT.Language = "en"
	cfg.STT.Temperature = 0.5
	cfg.TTS.Voice = "nova"
	cfg.TTS.Speed = 1.2

	// Save config
	if err := speech.SaveConfig(configPath, &cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config
	loadedCfg, err := speech.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.STT.Language != "en" {
		t.Errorf("Expected STT language 'en', got '%s'", loadedCfg.STT.Language)
	}

	if loadedCfg.STT.Temperature != 0.5 {
		t.Errorf("Expected STT temperature 0.5, got %.2f", loadedCfg.STT.Temperature)
	}

	if loadedCfg.TTS.Voice != "nova" {
		t.Errorf("Expected TTS voice 'nova', got '%s'", loadedCfg.TTS.Voice)
	}

	if loadedCfg.TTS.Speed != 1.2 {
		t.Errorf("Expected TTS speed 1.2, got %.2f", loadedCfg.TTS.Speed)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "speech-config.json")

	// Create a basic config file without API key
	cfg := speech.DefaultConfig()
	if err := speech.SaveConfig(configPath, &cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Set environment variable
	os.Setenv("OPENAI_API_KEY", "test-api-key-123")
	defer os.Unsetenv("OPENAI_API_KEY")

	// Load config with env
	loadedCfg, err := speech.LoadConfigWithEnv(configPath)
	if err != nil {
		t.Fatalf("Failed to load config with env: %v", err)
	}

	if loadedCfg.STT.APIKey != "test-api-key-123" {
		t.Errorf("Expected STT API key from env, got '%s'", loadedCfg.STT.APIKey)
	}

	if loadedCfg.TTS.APIKey != "test-api-key-123" {
		t.Errorf("Expected TTS API key from env, got '%s'", loadedCfg.TTS.APIKey)
	}
}

func TestNewSTTFromConfig(t *testing.T) {
	cfg := speech.DefaultConfig()
	cfg.STT.APIKey = "test-key"

	stt, err := speech.NewSTTFromConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to create STT from config: %v", err)
	}

	if stt.Provider() != "openai-whisper" {
		t.Errorf("Expected provider 'openai-whisper', got '%s'", stt.Provider())
	}
}

func TestNewSTTFromConfigMissingAPIKey(t *testing.T) {
	cfg := speech.DefaultConfig()
	// APIKey is not set

	_, err := speech.NewSTTFromConfig(&cfg)
	if err == nil {
		t.Error("Expected error when STT API key is missing")
	}
}

func TestNewTTSFromConfig(t *testing.T) {
	cfg := speech.DefaultConfig()
	cfg.TTS.APIKey = "test-key"

	tts, err := speech.NewTTSFromConfig(&cfg)
	if err != nil {
		t.Fatalf("Failed to create TTS from config: %v", err)
	}

	if tts.Provider() != "openai-tts" {
		t.Errorf("Expected provider 'openai-tts', got '%s'", tts.Provider())
	}
}

func TestNewTTSFromConfigMissingAPIKey(t *testing.T) {
	cfg := speech.DefaultConfig()
	// APIKey is not set

	_, err := speech.NewTTSFromConfig(&cfg)
	if err == nil {
		t.Error("Expected error when TTS API key is missing")
	}
}

func TestWhisperSTTDefaults(t *testing.T) {
	stt := speech.NewWhisperSTT(speech.STTConfig{})

	if stt.Provider() != "openai-whisper" {
		t.Errorf("Expected provider 'openai-whisper', got '%s'", stt.Provider())
	}
}

func TestOpenAITTSDefaults(t *testing.T) {
	tts := speech.NewOpenAITTS(speech.TTSConfig{})

	if tts.Provider() != "openai-tts" {
		t.Errorf("Expected provider 'openai-tts', got '%s'", tts.Provider())
	}
}

func TestWhisperSTTInvalidAPIKey(t *testing.T) {
	t.Skip("Skipping API test - requires valid OpenAI API key")

	stt := speech.NewWhisperSTT(speech.STTConfig{
		APIKey: "invalid-key",
		Model:  "whisper-1",
	})

	ctx := context.Background()
	audio := strings.NewReader("fake audio data")

	_, err := stt.Transcribe(ctx, audio, "mp3")
	if err == nil {
		t.Error("Expected error with invalid API key")
	}

	// Check that it's a speech error
	if _, ok := err.(*speech.Error); !ok {
		t.Errorf("Expected *speech.Error, got %T", err)
	}
}

func TestOpenAITTSInvalidAPIKey(t *testing.T) {
	t.Skip("Skipping API test - requires valid OpenAI API key")

	tts := speech.NewOpenAITTS(speech.TTSConfig{
		APIKey: "invalid-key",
		Model:  "tts-1",
	})

	ctx := context.Background()

	_, err := tts.Synthesize(ctx, "Hello world")
	if err == nil {
		t.Error("Expected error with invalid API key")
	}

	// Check that it's a speech error
	if _, ok := err.(*speech.Error); !ok {
		t.Errorf("Expected *speech.Error, got %T", err)
	}
}

func TestErrorType(t *testing.T) {
	err := &speech.Error{
		Provider: "test-provider",
		Type:     "test-type",
		Code:     "test-code",
		Message:  "test message",
		Retry:    true,
	}

	expected := "test-provider (test-type): test message"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestWhisperSTTTranscribeFileNotFound(t *testing.T) {
	stt := speech.NewWhisperSTT(speech.STTConfig{
		APIKey: "test-key",
	})

	ctx := context.Background()
	_, err := stt.TranscribeFile(ctx, "/nonexistent/file.mp3")

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	if speechErr, ok := err.(*speech.Error); ok {
		if speechErr.Code != "file_error" {
			t.Errorf("Expected error code 'file_error', got '%s'", speechErr.Code)
		}
	}
}

func TestOpenAITTSSynthesizeToFile(t *testing.T) {
	t.Skip("Skipping API test - requires valid OpenAI API key")

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp3")

	tts := speech.NewOpenAITTS(speech.TTSConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Model:  "tts-1",
		Voice:  "alloy",
	})

	ctx := context.Background()
	err := tts.SynthesizeToFile(ctx, "Hello world", outputPath)
	if err != nil {
		t.Fatalf("Failed to synthesize to file: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestContextTimeout(t *testing.T) {
	stt := speech.NewWhisperSTT(speech.STTConfig{
		APIKey:  "test-key",
		Timeout: 1 * time.Millisecond, // Very short timeout
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	audio := strings.NewReader("fake audio")
	_, err := stt.Transcribe(ctx, audio, "mp3")

	if err == nil {
		t.Error("Expected timeout error")
	}
}