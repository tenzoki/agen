package audio

import (
	"fmt"
	"os/exec"
)

// SoxPlayer implements Player using the sox 'play' command
type SoxPlayer struct {
	currentCmd *exec.Cmd
}

// NewSoxPlayer creates a new sox-based player
func NewSoxPlayer() *SoxPlayer {
	return &SoxPlayer{}
}

// IsAvailable checks if sox play command is installed
func (p *SoxPlayer) IsAvailable() bool {
	_, err := exec.LookPath("play")
	if err == nil {
		return true
	}
	// Try afplay on macOS
	_, err = exec.LookPath("afplay")
	return err == nil
}

// Play plays an audio file and waits for completion
func (p *SoxPlayer) Play(filePath string) error {
	if !p.IsAvailable() {
		return &Error{
			Operation: "play",
			Message:   "no audio player found. Install sox (brew install sox) or use afplay",
		}
	}

	// Try afplay first (native macOS)
	if _, err := exec.LookPath("afplay"); err == nil {
		cmd := exec.Command("afplay", filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return &Error{
				Operation: "play",
				Message:   fmt.Sprintf("afplay failed: %s", string(output)),
				Err:       err,
			}
		}
		return nil
	}

	// Fallback to sox play
	cmd := exec.Command("play", "-q", filePath) // -q for quiet
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &Error{
			Operation: "play",
			Message:   fmt.Sprintf("play failed: %s", string(output)),
			Err:       err,
		}
	}

	return nil
}

// PlayAsync plays audio in background, returns immediately
func (p *SoxPlayer) PlayAsync(filePath string) error {
	if !p.IsAvailable() {
		return &Error{
			Operation: "play",
			Message:   "no audio player found. Install sox (brew install sox) or use afplay",
		}
	}

	// Try afplay first (native macOS)
	if _, err := exec.LookPath("afplay"); err == nil {
		cmd := exec.Command("afplay", filePath)
		p.currentCmd = cmd
		if err := cmd.Start(); err != nil {
			return &Error{
				Operation: "play",
				Message:   "failed to start afplay",
				Err:       err,
			}
		}
		return nil
	}

	// Fallback to sox play
	cmd := exec.Command("play", "-q", filePath)
	p.currentCmd = cmd
	if err := cmd.Start(); err != nil {
		return &Error{
			Operation: "play",
			Message:   "failed to start play",
			Err:       err,
		}
	}

	return nil
}

// Stop stops current playback
func (p *SoxPlayer) Stop() error {
	if p.currentCmd == nil || p.currentCmd.Process == nil {
		return nil // Nothing to stop
	}

	if err := p.currentCmd.Process.Kill(); err != nil {
		return &Error{
			Operation: "stop",
			Message:   "failed to stop playback",
			Err:       err,
		}
	}

	p.currentCmd = nil
	return nil
}