package audio

import (
	"fmt"
	"os/exec"
	"time"
)

// SoxRecorder implements Recorder using the sox command-line tool
type SoxRecorder struct {
	config RecordConfig
}

// NewSoxRecorder creates a new sox-based recorder
func NewSoxRecorder(config RecordConfig) *SoxRecorder {
	return &SoxRecorder{
		config: config,
	}
}

// IsAvailable checks if sox is installed
func (r *SoxRecorder) IsAvailable() bool {
	_, err := exec.LookPath("sox")
	return err == nil
}

// Record captures audio from microphone for specified duration
func (r *SoxRecorder) Record(outputPath string, duration time.Duration) error {
	if !r.IsAvailable() {
		return &Error{
			Operation: "record",
			Message:   "sox not found. Install with: brew install sox",
		}
	}

	// sox -d output.wav trim 0 5
	// -d = default input device (microphone)
	// trim 0 N = record for N seconds
	args := []string{
		"-d",                                  // Default input device
		"-r", fmt.Sprintf("%d", r.config.SampleRate), // Sample rate
		"-c", fmt.Sprintf("%d", r.config.Channels),   // Channels
		"-b", "16",                            // Bits per sample
		outputPath,                            // Output file
		"trim", "0", fmt.Sprintf("%.2f", duration.Seconds()), // Duration
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

// RecordUntilSilence records until silence is detected or max duration reached
func (r *SoxRecorder) RecordUntilSilence(outputPath string, maxDuration time.Duration) error {
	if !r.IsAvailable() {
		return &Error{
			Operation: "record",
			Message:   "sox not found. Install with: brew install sox",
		}
	}

	// sox -d output.wav silence 1 0.1 1% 1 2.0 1%
	// silence = stop recording on silence
	// First silence: 1 0.1 1% (wait for sound to start)
	// Second silence: 1 2.0 1% (stop after 2 seconds of silence below 1%)
	args := []string{
		"-d", // Default input device
		"-r", fmt.Sprintf("%d", r.config.SampleRate),
		"-c", fmt.Sprintf("%d", r.config.Channels),
		"-b", "16",
		outputPath,
		"silence",
		"1", "0.1", "1%", // Wait for sound to start (above 1% volume for 0.1s)
		"1", "2.0", "1%", // Stop after 2 seconds of silence (below 1% volume)
	}

	// Add timeout wrapper
	cmd := exec.Command("timeout", fmt.Sprintf("%.0f", maxDuration.Seconds()), "sox")
	cmd.Args = append(cmd.Args, args[1:]...) // Add sox args after timeout duration

	// Fallback if timeout command not available (macOS doesn't have it by default)
	if _, err := exec.LookPath("timeout"); err != nil {
		// Use sox without timeout
		cmd = exec.Command("sox", args...)

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
			return nil // Timeout is not an error, just max duration reached
		}
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Exit code 124 from timeout means max duration reached (not an error)
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 124 {
			return nil
		}
		return &Error{
			Operation: "record",
			Message:   fmt.Sprintf("sox failed: %s", string(output)),
			Err:       err,
		}
	}

	return nil
}