package audio_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tenzoki/agen/alfa/internal/audio"
)

func TestDefaultRecordConfig(t *testing.T) {
	cfg := audio.DefaultRecordConfig()

	if cfg.SampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", cfg.SampleRate)
	}

	if cfg.Channels != 1 {
		t.Errorf("Expected 1 channel, got %d", cfg.Channels)
	}

	if cfg.Format != "wav" {
		t.Errorf("Expected format 'wav', got '%s'", cfg.Format)
	}
}

func TestSoxRecorderIsAvailable(t *testing.T) {
	recorder := audio.NewSoxRecorder(audio.DefaultRecordConfig())

	// This test will pass if sox is installed, skip otherwise
	if !recorder.IsAvailable() {
		t.Skip("Sox not installed, skipping recorder tests")
	}
}

func TestSoxPlayerIsAvailable(t *testing.T) {
	player := audio.NewSoxPlayer()

	// This test checks for sox 'play' or 'afplay'
	if !player.IsAvailable() {
		t.Skip("No audio player available, skipping player tests")
	}
}

func TestVADRecorderIsAvailable(t *testing.T) {
	recorder := audio.NewVADRecorder(audio.DefaultRecordConfig())

	if !recorder.IsAvailable() {
		t.Skip("Sox not installed, skipping VAD tests")
	}
}

func TestSoxRecorderRecord(t *testing.T) {
	t.Skip("Skipping recording test - requires microphone input")

	recorder := audio.NewSoxRecorder(audio.DefaultRecordConfig())

	if !recorder.IsAvailable() {
		t.Skip("Sox not installed")
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test_record.wav")

	// Record 1 second of audio
	err := recorder.Record(outputPath, 1*time.Second)
	if err != nil {
		t.Fatalf("Recording failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Check file has some content
	info, _ := os.Stat(outputPath)
	if info.Size() == 0 {
		t.Error("Output file is empty")
	}
}

func TestSoxPlayerPlay(t *testing.T) {
	t.Skip("Skipping playback test - requires audio output and user verification")

	player := audio.NewSoxPlayer()

	if !player.IsAvailable() {
		t.Skip("No audio player available")
	}

	// This test would require an actual audio file to play
	// Skipping for automated tests
}

func TestErrorType(t *testing.T) {
	err := &audio.Error{
		Operation: "test-operation",
		Message:   "test message",
	}

	expected := "audio test-operation: test message"
	if err.Error() != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestRecorderNotAvailable(t *testing.T) {
	recorder := audio.NewSoxRecorder(audio.DefaultRecordConfig())

	if recorder.IsAvailable() {
		t.Skip("Sox is installed, cannot test unavailable case")
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "test.wav")

	err := recorder.Record(outputPath, 1*time.Second)
	if err == nil {
		t.Error("Expected error when sox not available")
	}

	if audioErr, ok := err.(*audio.Error); ok {
		if audioErr.Operation != "record" {
			t.Errorf("Expected operation 'record', got '%s'", audioErr.Operation)
		}
	} else {
		t.Errorf("Expected *audio.Error, got %T", err)
	}
}