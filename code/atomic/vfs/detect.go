package vfs

import (
	"fmt"
	"os"
	"path/filepath"
)

// DetectFrameworkRoot detects AGEN framework root by looking for marker files
//
// Searches upward from current directory looking for:
// - go.work (workspace file)
// - code/cellorg directory
// - guidelines directory
// - workbench directory
//
// Requires at least 2 markers to confirm it's the AGEN root
func DetectFrameworkRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		markers := 0

		// Check for go.work
		if fileExists(filepath.Join(current, "go.work")) {
			markers++
		}

		// Check for code/cellorg directory
		if dirExists(filepath.Join(current, "code", "cellorg")) {
			markers++
		}

		// Check for guidelines directory
		if dirExists(filepath.Join(current, "guidelines")) {
			markers++
		}

		// Check for workbench directory
		if dirExists(filepath.Join(current, "workbench")) {
			markers++
		}

		// Need at least 2 markers to confirm it's AGEN root
		if markers >= 2 {
			return current, nil
		}

		// Move up one directory
		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root without finding AGEN
			return "", fmt.Errorf("not inside AGEN framework (searched from %s to root)", os.Getenv("PWD"))
		}
		current = parent
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
