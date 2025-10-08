package orchestrator

import (
	"fmt"
	"os"
	"time"

	"github.com/tenzoki/agen/alfa/internal/audio"
	"github.com/tenzoki/agen/alfa/internal/speech"
)

// VoiceComponents holds all voice-related components
type VoiceComponents struct {
	STT      speech.STT
	TTS      speech.TTS
	Recorder audio.Recorder
	Player   audio.Player
}

// InitializeVoice creates and initializes all voice components
func InitializeVoice() (*VoiceComponents, error) {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	vc := &VoiceComponents{}

	// Create STT (Whisper)
	vc.STT = speech.NewWhisperSTT(speech.STTConfig{
		APIKey:  openaiKey,
		Model:   "whisper-1",
		Timeout: 60 * time.Second,
	})

	// Create TTS
	vc.TTS = speech.NewOpenAITTS(speech.TTSConfig{
		APIKey:  openaiKey,
		Model:   "tts-1",
		Voice:   "alloy",
		Speed:   1.0,
		Format:  "mp3",
		Timeout: 60 * time.Second,
	})

	// Create audio recorder
	recorder := audio.NewVADRecorder(audio.DefaultRecordConfig())
	if recorder.IsAvailable() {
		vc.Recorder = recorder
	} else {
		fmt.Println("⚠️  Warning: sox not found. Voice input disabled.")
		fmt.Println("   Install with: brew install sox")
	}

	// Create audio player
	player := audio.NewSoxPlayer()
	if player.IsAvailable() {
		vc.Player = player
	} else {
		fmt.Println("⚠️  Warning: No audio player found. Voice output disabled.")
	}

	return vc, nil
}

// EnableVoiceInput activates voice input (STT)
func (o *Orchestrator) EnableVoiceInput() error {
	if o.stt != nil {
		return nil // Already enabled
	}

	vc, err := InitializeVoice()
	if err != nil {
		return err
	}

	o.stt = vc.STT
	o.recorder = vc.Recorder

	if o.recorder != nil {
		fmt.Println("🎤 Voice input enabled")
	}

	return nil
}

// EnableVoiceOutput activates voice output (TTS)
func (o *Orchestrator) EnableVoiceOutput() error {
	if o.tts != nil {
		return nil // Already enabled
	}

	vc, err := InitializeVoice()
	if err != nil {
		return err
	}

	o.tts = vc.TTS
	o.player = vc.Player

	if o.player != nil {
		fmt.Println("🔊 Voice output enabled")
	}

	return nil
}

// EnableVoice activates voice input/output
func (o *Orchestrator) EnableVoice() error {
	vc, err := InitializeVoice()
	if err != nil {
		return err
	}

	// Enable both input and output
	if o.stt == nil {
		o.stt = vc.STT
		o.recorder = vc.Recorder
	}
	if o.tts == nil {
		o.tts = vc.TTS
		o.player = vc.Player
	}

	if o.recorder != nil || o.player != nil {
		fmt.Println("🎤 Voice mode enabled")
	}

	return nil
}

// DisableVoice deactivates voice input/output
func (o *Orchestrator) DisableVoice() {
	o.stt = nil
	o.tts = nil
	o.recorder = nil
	o.player = nil
	fmt.Println("🔇 Voice mode disabled")
}

// DisableVoiceInput deactivates voice input only
func (o *Orchestrator) DisableVoiceInput() {
	o.stt = nil
	o.recorder = nil
	fmt.Println("🔇 Voice input disabled")
}

// DisableVoiceOutput deactivates voice output only
func (o *Orchestrator) DisableVoiceOutput() {
	o.tts = nil
	o.player = nil
	fmt.Println("🔇 Voice output disabled")
}

// IsVoiceEnabled returns true if voice is currently active
func (o *Orchestrator) IsVoiceEnabled() bool {
	return o.tts != nil || o.stt != nil
}
