// Package logging provides session-based logging for Alfa and agents.
// It enables clean CLI output while preserving detailed logs in session files.
package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionLogger manages logging to both file and console with selective output.
// Debug/verbose logs go to file only, while critical user-facing messages
// go to both file and console.
type SessionLogger struct {
	sessionFile *os.File
	mu          sync.Mutex
	sessionPath string
	quietMode   bool // If true, only errors and explicit user messages go to console
}

// New creates a new session logger.
// logDir: directory where session log files are stored
// quietMode: if true, suppress debug output to console (file only)
func New(logDir string, quietMode bool) (*SessionLogger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate session filename with timestamp
	sessionID := time.Now().Format("20060102-150405")
	sessionPath := filepath.Join(logDir, fmt.Sprintf("session-%s.log", sessionID))

	// Open session log file
	file, err := os.OpenFile(sessionPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create session log file: %w", err)
	}

	logger := &SessionLogger{
		sessionFile: file,
		sessionPath: sessionPath,
		quietMode:   quietMode,
	}

	// Write session header
	logger.writeToFile("=== Alfa Session Started ===\n")
	logger.writeToFile("Session ID: %s\n", sessionID)
	logger.writeToFile("Time: %s\n", time.Now().Format(time.RFC3339))
	logger.writeToFile("Log file: %s\n", sessionPath)
	logger.writeToFile("===============================\n\n")

	// Redirect standard log package to session file for clean CLI
	// This captures all log.Printf() calls from libraries and agents
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime) // Standard log format

	return logger, nil
}

// Close closes the session log file.
func (s *SessionLogger) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sessionFile != nil {
		s.writeToFile("\n=== Session Ended ===\n")
		s.writeToFile("Time: %s\n", time.Now().Format(time.RFC3339))
		return s.sessionFile.Close()
	}
	return nil
}

// GetSessionPath returns the path to the current session log file.
func (s *SessionLogger) GetSessionPath() string {
	return s.sessionPath
}

// Debug writes a debug message to the session file only (not console).
func (s *SessionLogger) Debug(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	s.writeToFile("[%s] DEBUG: %s\n", timestamp, fmt.Sprintf(format, args...))
}

// Info writes an info message to the session file only (not console in quiet mode).
func (s *SessionLogger) Info(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	s.writeToFile("[%s] INFO: %s\n", timestamp, message)

	// Also print to console if not in quiet mode
	if !s.quietMode {
		fmt.Println(message)
	}
}

// UserMessage writes a user-facing message to both file and console.
// This is for important messages that the user should see in the CLI.
func (s *SessionLogger) UserMessage(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	s.writeToFile("[%s] USER: %s\n", timestamp, message)
	fmt.Println(message)
}

// Error writes an error message to both file and console.
func (s *SessionLogger) Error(format string, args ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	message := fmt.Sprintf(format, args...)
	s.writeToFile("[%s] ERROR: %s\n", timestamp, message)
	fmt.Fprintf(os.Stderr, "❌ Error: %s\n", message)
}

// LogUserInput records user input to the session file.
func (s *SessionLogger) LogUserInput(input string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	s.writeToFile("\n[%s] USER INPUT:\n%s\n\n", timestamp, input)
}

// LogAIResponse records AI response to the session file.
func (s *SessionLogger) LogAIResponse(response string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	s.writeToFile("[%s] AI RESPONSE:\n%s\n\n", timestamp, response)
}

// LogPEVEvent records PEV cycle events to the session file.
func (s *SessionLogger) LogPEVEvent(event string, details string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	timestamp := time.Now().Format("15:04:05")
	s.writeToFile("[%s] PEV: %s\n", timestamp, event)
	if details != "" {
		s.writeToFile("  Details: %s\n", details)
	}
}

// writeToFile writes to the session file (must be called with lock held).
func (s *SessionLogger) writeToFile(format string, args ...interface{}) {
	if s.sessionFile != nil {
		fmt.Fprintf(s.sessionFile, format, args...)
		s.sessionFile.Sync() // Ensure immediate write for real-time viewing
	}
}

// SetQuietMode enables or disables quiet mode.
func (s *SessionLogger) SetQuietMode(quiet bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.quietMode = quiet
}

// Global session logger instance (initialized by orchestrator)
var globalLogger *SessionLogger
var globalMu sync.Mutex

// SetGlobalLogger sets the global session logger instance.
func SetGlobalLogger(logger *SessionLogger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = logger
}

// GetGlobalLogger returns the global session logger instance.
func GetGlobalLogger() *SessionLogger {
	globalMu.Lock()
	defer globalMu.Unlock()
	return globalLogger
}

// GlobalDebug writes to global logger if available, otherwise falls back to log.Printf.
func GlobalDebug(format string, args ...interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Debug(format, args...)
	} else {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// GlobalInfo writes to global logger if available, otherwise falls back to log.Printf.
func GlobalInfo(format string, args ...interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Info(format, args...)
	} else {
		log.Printf("[INFO] "+format, args...)
	}
}

// GlobalError writes to global logger if available, otherwise falls back to log.Printf.
func GlobalError(format string, args ...interface{}) {
	if logger := GetGlobalLogger(); logger != nil {
		logger.Error(format, args...)
	} else {
		log.Printf("[ERROR] "+format, args...)
	}
}
