package audio

import (
	"os/exec"
	"sync"
)

// SoxPlayer implements Player using the sox 'play' command
type SoxPlayer struct {
	currentCmd *exec.Cmd
	mu         sync.Mutex
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

	var cmd *exec.Cmd

	// Try afplay first (native macOS)
	if _, err := exec.LookPath("afplay"); err == nil {
		cmd = exec.Command("afplay", filePath)
	} else {
		// Fallback to sox play
		cmd = exec.Command("play", "-q", filePath) // -q for quiet
	}

	// Store command so it can be stopped
	p.mu.Lock()
	p.currentCmd = cmd
	p.mu.Unlock()

	// Start the process
	if err := cmd.Start(); err != nil {
		p.mu.Lock()
		p.currentCmd = nil
		p.mu.Unlock()
		return &Error{
			Operation: "play",
			Message:   "failed to start playback",
			Err:       err,
		}
	}

	// Wait for completion (can be interrupted by Stop())
	err := cmd.Wait()
	p.mu.Lock()
	p.currentCmd = nil
	p.mu.Unlock()

	if err != nil {
		// If killed by Stop(), don't return error
		if err.Error() == "signal: killed" {
			return nil
		}
		return &Error{
			Operation: "play",
			Message:   "playback failed",
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

	var cmd *exec.Cmd

	// Try afplay first (native macOS)
	if _, err := exec.LookPath("afplay"); err == nil {
		cmd = exec.Command("afplay", filePath)
	} else {
		// Fallback to sox play
		cmd = exec.Command("play", "-q", filePath)
	}

	p.mu.Lock()
	p.currentCmd = cmd
	p.mu.Unlock()

	if err := cmd.Start(); err != nil {
		p.mu.Lock()
		p.currentCmd = nil
		p.mu.Unlock()
		return &Error{
			Operation: "play",
			Message:   "failed to start playback",
			Err:       err,
		}
	}

	return nil
}

// Stop stops current playback
func (p *SoxPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

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

	return nil
}