package context

import (
	"alfa/internal/ai"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// Manager maintains conversation context and file state
type Manager struct {
	workdir          string
	conversationID   string
	activeProject    string
	messages         []ai.Message
	fileModifications map[string][]FileModification
	openFiles        map[string]bool
	mu               sync.RWMutex
	storage          *Storage
}

// FileModification records a change to a file
type FileModification struct {
	Timestamp time.Time
	PatchJSON string
	Summary   string
}

// Config holds configuration for the context manager
type Config struct {
	Workdir        string
	ConversationID string
	StoragePath    string // Path to persist context
	MaxMessages    int    // Max messages to keep in memory
}

// NewManager creates a new context manager
func NewManager(workdir string) *Manager {
	return NewManagerWithConfig(Config{
		Workdir:     workdir,
		MaxMessages: 100,
	})
}

// NewManagerWithConfig creates a new context manager with custom config
func NewManagerWithConfig(cfg Config) *Manager {
	if cfg.ConversationID == "" {
		cfg.ConversationID = generateID()
	}
	if cfg.MaxMessages == 0 {
		cfg.MaxMessages = 100
	}
	if cfg.StoragePath == "" {
		cfg.StoragePath = filepath.Join(cfg.Workdir, ".alfa", "context.json")
	}

	m := &Manager{
		workdir:           cfg.Workdir,
		conversationID:    cfg.ConversationID,
		messages:          make([]ai.Message, 0),
		fileModifications: make(map[string][]FileModification),
		openFiles:         make(map[string]bool),
		storage:           NewStorage(cfg.StoragePath),
	}

	// Try to load existing context
	m.load()

	return m
}

// AddUserMessage adds a user message to the conversation
func (m *Manager) AddUserMessage(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, ai.Message{
		Role:    "user",
		Content: content,
	})

	m.trim()
	m.save()
}

// AddAssistantMessage adds an assistant message to the conversation
func (m *Manager) AddAssistantMessage(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, ai.Message{
		Role:    "assistant",
		Content: content,
	})

	m.trim()
	m.save()
}

// AddSystemMessage adds a system message to the conversation
func (m *Manager) AddSystemMessage(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, ai.Message{
		Role:    "system",
		Content: content,
	})

	m.trim()
	m.save()
}

// AddToolResults adds tool execution results as a user message
func (m *Manager) AddToolResults(results interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Format results as JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		data = []byte("Error formatting results")
	}

	content := "Tool execution results:\n```json\n" + string(data) + "\n```"

	m.messages = append(m.messages, ai.Message{
		Role:    "user",
		Content: content,
	})

	m.trim()
	m.save()
}

// RecordFileModification records a modification to a file
func (m *Manager) RecordFileModification(filePath, patchJSON string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	relPath, _ := filepath.Rel(m.workdir, filePath)

	mod := FileModification{
		Timestamp: time.Now(),
		PatchJSON: patchJSON,
		Summary:   summarizePatch(patchJSON),
	}

	m.fileModifications[relPath] = append(m.fileModifications[relPath], mod)
	m.openFiles[relPath] = true

	m.save()
}

// GetRecentHistory returns the N most recent messages
func (m *Manager) GetRecentHistory(n int) []ai.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n <= 0 || n > len(m.messages) {
		n = len(m.messages)
	}

	start := len(m.messages) - n
	history := make([]ai.Message, n)
	copy(history, m.messages[start:])

	return history
}

// GetOpenFiles returns the list of currently open files
func (m *Manager) GetOpenFiles() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make([]string, 0, len(m.openFiles))
	for file := range m.openFiles {
		files = append(files, file)
	}
	return files
}

// GetFileHistory returns modification history for a file
func (m *Manager) GetFileHistory(filePath string) []FileModification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	relPath, _ := filepath.Rel(m.workdir, filePath)
	return m.fileModifications[relPath]
}

// CloseFile removes a file from the open files set
func (m *Manager) CloseFile(filePath string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	relPath, _ := filepath.Rel(m.workdir, filePath)
	delete(m.openFiles, relPath)

	m.save()
}

// Clear resets all context
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = make([]ai.Message, 0)
	m.fileModifications = make(map[string][]FileModification)
	m.openFiles = make(map[string]bool)

	m.save()
}

// trim removes old messages if over the limit
func (m *Manager) trim() {
	maxMessages := 100 // TODO: make configurable

	if len(m.messages) > maxMessages {
		// Keep the most recent messages
		m.messages = m.messages[len(m.messages)-maxMessages:]
	}
}

// save persists context to disk
func (m *Manager) save() {
	m.storage.Save(m.toStorageFormat())
}

// load restores context from disk
func (m *Manager) load() {
	data := m.storage.Load()
	if data != nil {
		m.fromStorageFormat(data)
	}
}

// StorageData represents the serializable context state
type StorageData struct {
	ConversationID    string                       `json:"conversation_id"`
	ActiveProject     string                       `json:"active_project,omitempty"`
	Messages          []ai.Message                 `json:"messages"`
	FileModifications map[string][]FileModification `json:"file_modifications"`
	OpenFiles         []string                     `json:"open_files"`
	Timestamp         time.Time                    `json:"timestamp"`
}

func (m *Manager) toStorageFormat() *StorageData {
	openFilesList := make([]string, 0, len(m.openFiles))
	for file := range m.openFiles {
		openFilesList = append(openFilesList, file)
	}

	return &StorageData{
		ConversationID:    m.conversationID,
		ActiveProject:     m.activeProject,
		Messages:          m.messages,
		FileModifications: m.fileModifications,
		OpenFiles:         openFilesList,
		Timestamp:         time.Now(),
	}
}

func (m *Manager) fromStorageFormat(data *StorageData) {
	m.conversationID = data.ConversationID
	m.activeProject = data.ActiveProject
	m.messages = data.Messages
	m.fileModifications = data.FileModifications

	m.openFiles = make(map[string]bool)
	for _, file := range data.OpenFiles {
		m.openFiles[file] = true
	}
}

// SetActiveProject sets the active project name
func (m *Manager) SetActiveProject(projectName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeProject = projectName
	m.save()
}

// GetActiveProject returns the active project name
func (m *Manager) GetActiveProject() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeProject
}

// summarizePatch creates a brief summary of patch operations
func summarizePatch(patchJSON string) string {
	var ops []map[string]interface{}
	if err := json.Unmarshal([]byte(patchJSON), &ops); err != nil {
		return "patch"
	}

	if len(ops) == 1 {
		op := ops[0]
		return fmt.Sprintf("%s at line %.0f", op["type"].(string), op["line"].(float64))
	}

	return fmt.Sprintf("%d operations", len(ops))
}

// generateID creates a unique conversation ID
func generateID() string {
	return time.Now().Format("20060102-150405")
}
