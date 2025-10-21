package orchestrator

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
)

var ErrInputCancelled = errors.New("input cancelled")

// getMultiLineInput shows a multi-line input prompt using readline
// Returns the input string or error
//
// Usage:
// - Enter for newline (continue typing)
// - Ctrl+D on empty line to submit
// - Ctrl+C to cancel
func getMultiLineInput(prompt string) (string, error) {
	// Create readline instance with custom config
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		InterruptPrompt: "^C",
		EOFPrompt:       "",
		HistoryLimit:    100,
		// Disable built-in history file for now (can enable later)
		// HistoryFile: filepath.Join(os.TempDir(), ".alfa_history"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to initialize input: %w", err)
	}
	defer rl.Close()

	// Show instructions once at the beginning
	fmt.Println("(Enter for newline, Ctrl+D on empty line to submit, Ctrl+C to cancel)")

	var lines []string
	lineNum := 0

	for {
		// For first line, use the provided prompt
		// For subsequent lines, use continuation prompt
		if lineNum > 0 {
			rl.SetPrompt("  ") // Indented continuation
		}

		line, err := rl.Readline()

		if err == readline.ErrInterrupt {
			// Ctrl+C - cancel input
			fmt.Println() // Newline after ^C
			return "", ErrInputCancelled
		} else if err == io.EOF {
			// Ctrl+D - submit
			// If Ctrl+D on empty line, submit the accumulated input
			if line == "" && lineNum > 0 {
				// Submit
				break
			} else if line == "" && lineNum == 0 {
				// Ctrl+D on first line with no input - cancel
				fmt.Println()
				return "", ErrInputCancelled
			}
			// Ctrl+D with text on line - add the line and continue
			lines = append(lines, line)
			lineNum++
			continue
		} else if err != nil {
			// Other errors
			return "", fmt.Errorf("input error: %w", err)
		}

		// Normal line - add to input
		lines = append(lines, line)
		lineNum++
	}

	fmt.Println() // Blank line after submission for visual separation
	return strings.Join(lines, "\n"), nil
}
