package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"alfa/internal/audio"
)

func main() {
	fmt.Println("=== Audio System Test ===\n")

	// Check if sox is available
	recorder := audio.NewSoxRecorder(audio.DefaultRecordConfig())
	player := audio.NewSoxPlayer()

	fmt.Println("1. Checking audio system availability...")
	if !recorder.IsAvailable() {
		fmt.Println("   ❌ Sox recorder not available")
		fmt.Println("   Install with: brew install sox")
		return
	}
	fmt.Println("   ✓ Sox recorder available")

	if !player.IsAvailable() {
		fmt.Println("   ❌ Audio player not available")
		fmt.Println("   Install with: brew install sox")
		return
	}
	fmt.Println("   ✓ Audio player available")

	// Create output directory
	outputDir := "demo/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Test 1: Fixed duration recording
	fmt.Println("\n2. Testing fixed duration recording...")
	fmt.Println("   Recording 3 seconds... (speak now!)")

	recordPath := filepath.Join(outputDir, "test_fixed.wav")
	err := recorder.Record(recordPath, 3*time.Second)
	if err != nil {
		log.Fatalf("Recording failed: %v", err)
	}
	fmt.Printf("   ✓ Recorded to: %s\n", recordPath)

	// Play it back
	fmt.Println("   Playing back...")
	err = player.Play(recordPath)
	if err != nil {
		log.Printf("   ⚠️  Playback failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Playback complete")
	}

	// Test 2: Voice Activity Detection
	fmt.Println("\n3. Testing Voice Activity Detection...")
	fmt.Println("   Recording with auto-stop (speak, then pause for 2 seconds)...")
	fmt.Println("   Max duration: 30 seconds")

	vadRecorder := audio.NewVADRecorder(audio.DefaultRecordConfig())
	vadPath := filepath.Join(outputDir, "test_vad.wav")

	err = vadRecorder.RecordUntilSilence(vadPath, 30*time.Second)
	if err != nil {
		log.Fatalf("VAD recording failed: %v", err)
	}
	fmt.Printf("   ✓ Recorded to: %s\n", vadPath)

	// Play it back
	fmt.Println("   Playing back...")
	err = player.Play(vadPath)
	if err != nil {
		log.Printf("   ⚠️  Playback failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Playback complete")
	}

	// Test 3: Aggressive VAD (stops faster)
	fmt.Println("\n4. Testing Aggressive VAD...")
	fmt.Println("   Recording with fast auto-stop (stops after 1 second silence)...")

	aggressivePath := filepath.Join(outputDir, "test_aggressive.wav")
	err = vadRecorder.RecordUntilSilenceAggressive(aggressivePath, 30*time.Second)
	if err != nil {
		log.Fatalf("Aggressive VAD recording failed: %v", err)
	}
	fmt.Printf("   ✓ Recorded to: %s\n", aggressivePath)

	fmt.Println("   Playing back...")
	err = player.Play(aggressivePath)
	if err != nil {
		log.Printf("   ⚠️  Playback failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Playback complete")
	}

	// Test 4: Relaxed VAD (stops later)
	fmt.Println("\n5. Testing Relaxed VAD...")
	fmt.Println("   Recording with slow auto-stop (stops after 3 seconds silence)...")
	fmt.Println("   Good for slow speakers or thinking pauses")

	relaxedPath := filepath.Join(outputDir, "test_relaxed.wav")
	err = vadRecorder.RecordUntilSilenceRelaxed(relaxedPath, 30*time.Second)
	if err != nil {
		log.Fatalf("Relaxed VAD recording failed: %v", err)
	}
	fmt.Printf("   ✓ Recorded to: %s\n", relaxedPath)

	fmt.Println("   Playing back...")
	err = player.Play(relaxedPath)
	if err != nil {
		log.Printf("   ⚠️  Playback failed: %v\n", err)
	} else {
		fmt.Println("   ✓ Playback complete")
	}

	// Summary
	fmt.Println("\n=== Audio System Test Complete ===")
	fmt.Printf("\n📁 Audio files saved to: %s\n", outputDir)
	fmt.Println("\nFiles created:")
	fmt.Println("  - test_fixed.wav       (3 second fixed recording)")
	fmt.Println("  - test_vad.wav         (VAD with 2s silence detection)")
	fmt.Println("  - test_aggressive.wav  (VAD with 1s silence detection)")
	fmt.Println("  - test_relaxed.wav     (VAD with 3s silence detection)")
	fmt.Println("\nYou can play them with:")
	fmt.Println("  afplay demo/output/test_fixed.wav")
	fmt.Println("  play demo/output/test_vad.wav")
}