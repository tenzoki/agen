package speech

import (
	"context"
	"io"
	"time"
)

// STT defines the interface for Speech-to-Text conversion
type STT interface {
	// Transcribe converts audio to text
	Transcribe(ctx context.Context, audio io.Reader, format string) (*Transcription, error)

	// TranscribeFile converts an audio file to text
	TranscribeFile(ctx context.Context, filePath string) (*Transcription, error)

	// Provider returns the provider name (e.g., "openai-whisper")
	Provider() string
}

// TTS defines the interface for Text-to-Speech conversion
type TTS interface {
	// Synthesize converts text to audio
	Synthesize(ctx context.Context, text string) (io.ReadCloser, error)

	// SynthesizeToFile converts text to audio and saves to file
	SynthesizeToFile(ctx context.Context, text string, outputPath string) error

	// Provider returns the provider name (e.g., "openai-tts")
	Provider() string
}

// Transcription represents the result of speech-to-text conversion
type Transcription struct {
	Text     string        // The transcribed text
	Language string        // Detected language (e.g., "en", "de")
	Duration time.Duration // Audio duration
	Provider string        // Provider used for transcription
}

// SynthesisConfig holds configuration for TTS
type SynthesisConfig struct {
	Voice  string  // Voice identifier (e.g., "alloy", "echo", "nova")
	Model  string  // TTS model (e.g., "tts-1", "tts-1-hd")
	Speed  float64 // Speech speed (0.25 to 4.0)
	Format string  // Output format (e.g., "mp3", "opus", "aac", "flac")
}

// STTConfig holds configuration for STT
type STTConfig struct {
	APIKey      string        // API authentication key
	Model       string        // STT model (e.g., "whisper-1")
	Language    string        // Optional language hint (e.g., "en", "de")
	Temperature float64       // Sampling temperature (0.0-1.0)
	Timeout     time.Duration // Request timeout
	RetryCount  int           // Number of retries on failure
	RetryDelay  time.Duration // Delay between retries
}

// TTSConfig holds configuration for TTS
type TTSConfig struct {
	APIKey      string        // API authentication key
	Model       string        // TTS model (e.g., "tts-1", "tts-1-hd")
	Voice       string        // Voice to use (e.g., "alloy", "echo", "fable", "onyx", "nova", "shimmer")
	Speed       float64       // Speech speed (0.25 to 4.0)
	Format      string        // Output format (mp3, opus, aac, flac, wav, pcm)
	Timeout     time.Duration // Request timeout
	RetryCount  int           // Number of retries on failure
	RetryDelay  time.Duration // Delay between retries
}

// Error represents a speech processing error
type Error struct {
	Provider string // Which provider caused the error
	Type     string // Error type (e.g., "stt", "tts")
	Code     string // Error code from provider
	Message  string // Human-readable error message
	Retry    bool   // Whether the error is retryable
}

func (e *Error) Error() string {
	return e.Provider + " (" + e.Type + "): " + e.Message
}