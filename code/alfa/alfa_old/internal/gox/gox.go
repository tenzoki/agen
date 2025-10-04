package gox

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager wraps Gox orchestrator for Alfa integration
// NOTE: This is a placeholder until pkg/orchestrator is published in gox
type Manager struct {
	config        Config
	cells         map[string]*CellInfo // key: "cellID:projectID"
	mu            sync.RWMutex
	eventHandlers map[string][]EventHandler
	eventChannels map[string]chan Event
}

// Config for Gox manager
type Config struct {
	ConfigPath      string
	DefaultDataRoot string
	SupportPort     string
	BrokerPort      string
	Debug           bool
}

// CellInfo tracks running cells
type CellInfo struct {
	CellID    string
	ProjectID string
	VFSRoot   string
	StartedAt time.Time
}

// EventHandler is a callback for cell events
type EventHandler func(event Event)

// Event represents a Gox event
type Event struct {
	Topic     string
	ProjectID string
	Data      map[string]interface{}
	Timestamp time.Time
}

// NewManager creates a new Gox manager
func NewManager(config Config) (*Manager, error) {
	// Set defaults
	if config.SupportPort == "" {
		config.SupportPort = ":9000"
	}
	if config.BrokerPort == "" {
		config.BrokerPort = ":9001"
	}
	if config.ConfigPath == "" {
		config.ConfigPath = "config/gox"
	}

	// NOTE: Placeholder implementation until pkg/orchestrator is available
	// The actual gox embedded orchestrator will be integrated when the package is published
	mgr := &Manager{
		config:        config,
		cells:         make(map[string]*CellInfo),
		eventHandlers: make(map[string][]EventHandler),
		eventChannels: make(map[string]chan Event),
	}

	if config.Debug {
		log.Println("[Gox Manager] Initialized (placeholder - awaiting pkg/orchestrator publication)")
		log.Println("[Gox Manager] Cell management API ready, but agents won't be deployed until gox orchestrator is integrated")
	}

	return mgr, nil
}

// StartCell starts a Gox cell for a project
func (m *Manager) StartCell(cellID, projectID, vfsRoot string, env map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := cellID + ":" + projectID
	if _, exists := m.cells[key]; exists {
		return fmt.Errorf("cell %s already running for project %s", cellID, projectID)
	}

	// Add GOX_PROJECT_ID and GOX_DATA_ROOT to environment
	if env == nil {
		env = make(map[string]string)
	}
	env["GOX_PROJECT_ID"] = projectID
	env["GOX_DATA_ROOT"] = vfsRoot

	// NOTE: Placeholder - actual cell starting will happen when pkg/orchestrator is integrated
	// For now, just track the cell request
	if m.config.Debug {
		log.Printf("[Gox Manager] Cell start requested: %s (project: %s, vfs: %s)", cellID, projectID, vfsRoot)
		log.Printf("[Gox Manager] Placeholder: Cell will be started when gox pkg/orchestrator is integrated")
	}

	// Track cell
	m.cells[key] = &CellInfo{
		CellID:    cellID,
		ProjectID: projectID,
		VFSRoot:   vfsRoot,
		StartedAt: time.Now(),
	}

	if m.config.Debug {
		log.Printf("[Gox Manager] Started cell %s for project %s (VFS: %s)", cellID, projectID, vfsRoot)
	}

	return nil
}

// StopCell stops a running cell
func (m *Manager) StopCell(cellID, projectID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := cellID + ":" + projectID
	if _, exists := m.cells[key]; !exists {
		return fmt.Errorf("cell %s not running for project %s", cellID, projectID)
	}

	// NOTE: Placeholder - actual cell stopping will happen when pkg/orchestrator is integrated
	if m.config.Debug {
		log.Printf("[Gox Manager] Cell stop requested: %s (project: %s)", cellID, projectID)
	}

	// Remove from tracking
	delete(m.cells, key)

	if m.config.Debug {
		log.Printf("[Gox Manager] Stopped cell %s for project %s", cellID, projectID)
	}

	return nil
}

// ListCells returns all running cells
func (m *Manager) ListCells() []CellInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cells := make([]CellInfo, 0, len(m.cells))
	for _, cell := range m.cells {
		cells = append(cells, *cell)
	}

	return cells
}

// Subscribe subscribes to events from cells
func (m *Manager) Subscribe(topic string, handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventHandlers[topic] = append(m.eventHandlers[topic], handler)

	// Create event channel if it doesn't exist
	if _, exists := m.eventChannels[topic]; !exists {
		m.eventChannels[topic] = make(chan Event, 100)

		// NOTE: Placeholder - will be connected to actual broker when pkg/orchestrator is integrated
		if m.config.Debug {
			log.Printf("[Gox Manager] Event subscription created for topic: %s", topic)
			log.Printf("[Gox Manager] Placeholder: Events will flow when gox pkg/orchestrator is integrated")
		}
	}
}

// Publish publishes an event to cells
func (m *Manager) Publish(topic string, data map[string]interface{}) error {
	// NOTE: Placeholder - actual publishing will happen when pkg/orchestrator is integrated
	if m.config.Debug {
		log.Printf("[Gox Manager] Event publish requested for topic: %s", topic)
	}
	return nil
}

// PublishAndWait publishes an event and waits for response
func (m *Manager) PublishAndWait(requestTopic, responseTopic string, data map[string]interface{}, timeout time.Duration) (*Event, error) {
	// NOTE: Placeholder - actual pub/wait will happen when pkg/orchestrator is integrated
	if m.config.Debug {
		log.Printf("[Gox Manager] PublishAndWait requested: %s -> %s", requestTopic, responseTopic)
		log.Printf("[Gox Manager] Placeholder: Will execute when gox pkg/orchestrator is integrated")
	}

	return nil, fmt.Errorf("gox pkg/orchestrator not yet integrated - query_cell is not functional yet")
}

// Close stops all cells and shuts down the orchestrator
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Debug {
		log.Printf("[Gox Manager] Shutting down (%d cells running)", len(m.cells))
	}

	// Stop all cells
	for key := range m.cells {
		delete(m.cells, key)
	}

	// Close event channels
	for _, ch := range m.eventChannels {
		close(ch)
	}

	// NOTE: Placeholder - actual shutdown will happen when pkg/orchestrator is integrated
	return nil
}

// IsHealthy checks if Gox services are running
func (m *Manager) IsHealthy(ctx context.Context) bool {
	// NOTE: Placeholder - will check actual services when pkg/orchestrator is integrated
	// For now, manager is always "healthy" (initialized)
	return true
}

// GetCellInfo returns info about a specific cell
func (m *Manager) GetCellInfo(cellID, projectID string) (*CellInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := cellID + ":" + projectID
	cell, exists := m.cells[key]
	if !exists {
		return nil, fmt.Errorf("cell %s not running for project %s", cellID, projectID)
	}

	return cell, nil
}
