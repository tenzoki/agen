package speech

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConfigFile represents the speech configuration file structure
type ConfigFile struct {
	STT STTConfig `json:"stt"` // Speech-to-Text config
	TTS TTSConfig `json:"tts"` // Text-to-Speech config
}

// DefaultConfig returns a default speech configuration
func DefaultConfig() ConfigFile {
	return ConfigFile{
		STT: STTConfig{
			Model:       "whisper-1",
			Language:    "",
			Temperature: 0.0,
			Timeout:     60 * time.Second,
			RetryCount:  3,
			RetryDelay:  1 * time.Second,
		},
		TTS: TTSConfig{
			Model:      "tts-1",
			Voice:      "alloy",
			Speed:      1.0,
			Format:     "mp3",
			Timeout:    60 * time.Second,
			RetryCount: 3,
			RetryDelay: 1 * time.Second,
		},
	}
}

// LoadConfig loads speech configuration from a file
func LoadConfig(path string) (*ConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults
	defaults := DefaultConfig()

	// STT defaults
	if cfg.STT.Model == "" {
		cfg.STT.Model = defaults.STT.Model
	}
	if cfg.STT.Timeout == 0 {
		cfg.STT.Timeout = defaults.STT.Timeout
	}
	if cfg.STT.RetryCount == 0 {
		cfg.STT.RetryCount = defaults.STT.RetryCount
	}
	if cfg.STT.RetryDelay == 0 {
		cfg.STT.RetryDelay = defaults.STT.RetryDelay
	}

	// TTS defaults
	if cfg.TTS.Model == "" {
		cfg.TTS.Model = defaults.TTS.Model
	}
	if cfg.TTS.Voice == "" {
		cfg.TTS.Voice = defaults.TTS.Voice
	}
	if cfg.TTS.Speed == 0 {
		cfg.TTS.Speed = defaults.TTS.Speed
	}
	if cfg.TTS.Format == "" {
		cfg.TTS.Format = defaults.TTS.Format
	}
	if cfg.TTS.Timeout == 0 {
		cfg.TTS.Timeout = defaults.TTS.Timeout
	}
	if cfg.TTS.RetryCount == 0 {
		cfg.TTS.RetryCount = defaults.TTS.RetryCount
	}
	if cfg.TTS.RetryDelay == 0 {
		cfg.TTS.RetryDelay = defaults.TTS.RetryDelay
	}

	return &cfg, nil
}

// LoadConfigWithEnv loads configuration and merges with environment variables
func LoadConfigWithEnv(path string) (*ConfigFile, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		// If file doesn't exist, start with defaults
		if os.IsNotExist(err) {
			defaultCfg := DefaultConfig()
			cfg = &defaultCfg
		} else {
			return nil, err
		}
	}

	// Override API key from environment (both STT and TTS use same OpenAI key)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		cfg.STT.APIKey = apiKey
		cfg.TTS.APIKey = apiKey
	}

	return cfg, nil
}

// SaveConfig saves speech configuration to a file
func SaveConfig(path string, cfg *ConfigFile) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	return "config/speech-config.json"
}

// NewSTTFromConfig creates an STT client from configuration
func NewSTTFromConfig(cfg *ConfigFile) (STT, error) {
	if cfg.STT.APIKey == "" {
		return nil, fmt.Errorf("STT API key not set")
	}

	return NewWhisperSTT(cfg.STT), nil
}

// NewTTSFromConfig creates a TTS client from configuration
func NewTTSFromConfig(cfg *ConfigFile) (TTS, error) {
	if cfg.TTS.APIKey == "" {
		return nil, fmt.Errorf("TTS API key not set")
	}

	return NewOpenAITTS(cfg.TTS), nil
}