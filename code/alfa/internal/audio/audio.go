package audio

import (
	"fmt"
	"time"
)

// Recorder defines the interface for audio recording
type Recorder interface {
	// Record captures audio from microphone and saves to file
	Record(outputPath string, duration time.Duration) error

	// RecordUntilSilence records until silence is detected
	RecordUntilSilence(outputPath string, maxDuration time.Duration) error

	// IsAvailable checks if recording is available on this system
	IsAvailable() bool
}

// Player defines the interface for audio playback
type Player interface {
	// Play plays an audio file
	Play(filePath string) error

	// PlayAsync plays audio in background, returns immediately
	PlayAsync(filePath string) error

	// Stop stops current playback
	Stop() error

	// IsAvailable checks if playback is available on this system
	IsAvailable() bool
}

// RecordConfig holds configuration for audio recording
type RecordConfig struct {
	SampleRate int    // Sample rate in Hz (e.g., 16000, 44100)
	Channels   int    // Number of channels (1=mono, 2=stereo)
	Format     string // Audio format (wav, mp3, flac)
	Encoding   string // Encoding type (pcm_s16le, etc.)
}

// DefaultRecordConfig returns default recording configuration
func DefaultRecordConfig() RecordConfig {
	return RecordConfig{
		SampleRate: 16000, // 16kHz is optimal for speech
		Channels:   1,     // Mono for speech
		Format:     "wav",
		Encoding:   "pcm_s16le",
	}
}

// Error represents an audio processing error
type Error struct {
	Operation string // Operation that failed (e.g., "record", "play")
	Message   string // Human-readable error message
	Err       error  // Underlying error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("audio %s: %s (%v)", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("audio %s: %s", e.Operation, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}