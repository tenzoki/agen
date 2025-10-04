package audio

import (
	"fmt"
	"os/exec"
	"time"
)

// VADRecorder implements Voice Activity Detection using sox
type VADRecorder struct {
	config RecordConfig
}

// NewVADRecorder creates a new VAD-enabled recorder
func NewVADRecorder(config RecordConfig) *VADRecorder {
	return &VADRecorder{
		config: config,
	}
}

// IsAvailable checks if sox is installed
func (r *VADRecorder) IsAvailable() bool {
	_, err := exec.LookPath("sox")
	return err == nil
}

// RecordWithVAD records with advanced voice activity detection
// - silenceThresholdStart: volume % to consider as silence at start (e.g., "1%")
// - silenceThresholdEnd: volume % to consider as silence at end (e.g., "1%")
// - silenceDurationStart: how long to wait for speech to start (e.g., 0.5s)
// - silenceDurationEnd: how long silence before stopping (e.g., 2.0s)
func (r *VADRecorder) RecordWithVAD(
	outputPath string,
	maxDuration time.Duration,
	silenceThresholdStart string,
	silenceDurationStart float64,
	silenceThresholdEnd string,
	silenceDurationEnd float64,
) error {
	if !r.IsAvailable() {
		return &Error{
			Operation: "record",
			Message:   "sox not found. Install with: brew install sox",
		}
	}

	// sox -d output.wav silence 1 0.5 1% 1 2.0 1%
	// First silence params: wait for speech to start
	//   1 = at least 1 period of silence at start
	//   0.5 = duration of silence (0.5 seconds)
	//   1% = threshold below which is considered silence
	// Second silence params: detect end of speech
	//   1 = at least 1 period of silence at end
	//   2.0 = duration of silence (2.0 seconds)
	//   1% = threshold below which is considered silence
	args := []string{
		"-d", // Default input device
		"-r", fmt.Sprintf("%d", r.config.SampleRate),
		"-c", fmt.Sprintf("%d", r.config.Channels),
		"-b", "16",
		outputPath,
		"silence",
		"1", fmt.Sprintf("%.2f", silenceDurationStart), silenceThresholdStart,
		"1", fmt.Sprintf("%.2f", silenceDurationEnd), silenceThresholdEnd,
	}

	cmd := exec.Command("sox", args...)

	// Start the command
	if err := cmd.Start(); err != nil {
		return &Error{
			Operation: "record",
			Message:   "failed to start recording",
			Err:       err,
		}
	}

	// Wait with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			return &Error{
				Operation: "record",
				Message:   "recording failed",
				Err:       err,
			}
		}
		return nil
	case <-time.After(maxDuration):
		cmd.Process.Kill()
		return nil // Timeout is not an error
	}
}

// Record implements basic recording (delegates to VADRecorder with defaults)
func (r *VADRecorder) Record(outputPath string, duration time.Duration) error {
	// Just record for fixed duration without VAD
	args := []string{
		"-d",
		"-r", fmt.Sprintf("%d", r.config.SampleRate),
		"-c", fmt.Sprintf("%d", r.config.Channels),
		"-b", "16",
		outputPath,
		"trim", "0", fmt.Sprintf("%.2f", duration.Seconds()),
	}

	cmd := exec.Command("sox", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &Error{
			Operation: "record",
			Message:   fmt.Sprintf("sox failed: %s", string(output)),
			Err:       err,
		}
	}

	return nil
}

// RecordUntilSilence records with default VAD settings
func (r *VADRecorder) RecordUntilSilence(outputPath string, maxDuration time.Duration) error {
	return r.RecordWithVAD(
		outputPath,
		maxDuration,
		"1%",  // Start threshold
		0.3,   // Wait 0.3s for speech to start
		"1%",  // End threshold
		2.0,   // Stop after 2.0s of silence
	)
}

// RecordUntilSilenceAggressive uses more aggressive VAD (stops sooner)
func (r *VADRecorder) RecordUntilSilenceAggressive(outputPath string, maxDuration time.Duration) error {
	return r.RecordWithVAD(
		outputPath,
		maxDuration,
		"2%",  // Higher threshold at start (more sensitive)
		0.2,   // Shorter wait for speech
		"2%",  // Higher threshold at end
		1.0,   // Stop after 1.0s of silence (faster)
	)
}

// RecordUntilSilenceRelaxed uses relaxed VAD (stops later, good for slow speakers)
func (r *VADRecorder) RecordUntilSilenceRelaxed(outputPath string, maxDuration time.Duration) error {
	return r.RecordWithVAD(
		outputPath,
		maxDuration,
		"0.5%", // Lower threshold at start (less sensitive)
		0.5,    // Longer wait for speech
		"0.5%", // Lower threshold at end
		3.0,    // Stop after 3.0s of silence (more patient)
	)
}