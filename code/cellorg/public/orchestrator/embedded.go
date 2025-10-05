package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/broker"
	"github.com/tenzoki/agen/cellorg/internal/config"
	"github.com/tenzoki/agen/cellorg/internal/deployer"
	"github.com/tenzoki/agen/cellorg/internal/support"
)

// EmbeddedOrchestrator runs Gox orchestrator in-process
//
// Phase 2 implementation with actual agent deployment.
//
// Capabilities:
// - Event bridge for pub/sub communication
// - Cell lifecycle tracking
// - VFS root management per project
// - Actual agent deployment via internal deployer
// - Configuration loading from cells.yaml
type EmbeddedOrchestrator struct {
	config        Config
	eventBridge   *EventBridge
	cells         map[string]*RunningCell
	cellsMutex    sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	agentDeployer *deployer.AgentDeployer
	goxConfig     *config.Config
	cellsConfig   *config.CellsConfig
	poolConfig    *config.PoolConfig

	// Embedded services (Phase 3)
	supportService *support.Service
	brokerService  *broker.Service
	servicesReady  chan struct{}
}

// RunningCell represents a cell instance running for a project
type RunningCell struct {
	CellID    string
	ProjectID string
	VFSRoot   string
	Status    CellStatus
	StartedAt time.Time
	Options   CellOptions
}

// NewEmbedded creates a new embedded orchestrator
//
// This creates an orchestrator instance that manages:
// - Cell lifecycle state
// - Event bridging (Gox topics → Go channels)
// - VFS root tracking per project
// - Agent deployment via internal deployer
// - Configuration loading from cells.yaml and pool.yaml
//
// Phase 2: Loads actual configuration and sets up deployer.
// TODO: Wire to support/broker services for full integration.
func NewEmbedded(cfg Config) (*EmbeddedOrchestrator, error) {
	// Set defaults
	if cfg.ConfigPath == "" {
		cfg.ConfigPath = "./config"
	}
	if cfg.DefaultDataRoot == "" {
		cfg.DefaultDataRoot = "/var/lib/gox"
	}
	if cfg.SupportPort == "" {
		cfg.SupportPort = ":9000"
	}
	if cfg.BrokerPort == "" {
		cfg.BrokerPort = ":9001"
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())

	// Create embedded orchestrator
	eo := &EmbeddedOrchestrator{
		config: cfg,
		cells:  make(map[string]*RunningCell),
		ctx:    ctx,
		cancel: cancel,
	}

	// Create event bridge
	eo.eventBridge = &EventBridge{
		subscribers: make(map[string][]chan Event),
	}

	// Load Gox configuration (gox.yaml)
	goxConfigPath := cfg.ConfigPath + "/gox.yaml"
	goxConfig, err := config.Load(goxConfigPath)
	if err != nil {
		// Use defaults if config not found
		if cfg.Debug {
			fmt.Printf("[Gox Embedded] Warning: Could not load gox.yaml: %v, using defaults\n", err)
		}
		goxConfig = &config.Config{
			Support: config.SupportConfig{Port: cfg.SupportPort},
			Broker:  config.BrokerConfig{Port: cfg.BrokerPort},
			BaseDir: []string{cfg.ConfigPath},
			Pool:    []string{"pool.yaml"},
			Cells:   []string{"cells.yaml"},
		}
	}
	eo.goxConfig = goxConfig

	// Load pool configuration
	poolConfig, err := goxConfig.LoadPool()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("[Gox Embedded] Warning: Could not load pool.yaml: %v\n", err)
		}
	}
	eo.poolConfig = poolConfig

	// Load cells configuration
	cellsConfig, err := goxConfig.LoadCells()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("[Gox Embedded] Warning: Could not load cells.yaml: %v\n", err)
		}
	}
	eo.cellsConfig = cellsConfig

	// Start embedded services (Phase 3)
	eo.servicesReady = make(chan struct{})

	// Create support service
	eo.supportService = support.NewService(support.SupportConfig{
		Port:  cfg.SupportPort,
		Debug: cfg.Debug,
	})

	// Load agent types into support service
	if poolConfig != nil {
		// Convert pool config to the format support service expects
		poolPath := fmt.Sprintf("%s/pool.yaml", cfg.ConfigPath)
		if err := eo.supportService.LoadAgentTypesFromFile(poolPath); err != nil && cfg.Debug {
			fmt.Printf("[Gox Embedded] Warning: Could not load agent types: %v\n", err)
		}
	}

	// Create broker service
	eo.brokerService = broker.NewService(struct {
		Port, Protocol, Codec string
		Debug                 bool
	}{
		Port:     cfg.BrokerPort,
		Protocol: "tcp",
		Codec:    "json",
		Debug:    cfg.Debug,
	})

	// Start services as goroutines
	go func() {
		if cfg.Debug {
			fmt.Println("[Gox Embedded] Starting support service...")
		}
		if err := eo.supportService.Start(eo.ctx); err != nil && eo.ctx.Err() == nil {
			fmt.Printf("[Gox Embedded] Support service error: %v\n", err)
		}
	}()

	go func() {
		if cfg.Debug {
			fmt.Println("[Gox Embedded] Starting broker service...")
		}
		if err := eo.brokerService.Start(eo.ctx); err != nil && eo.ctx.Err() == nil {
			fmt.Printf("[Gox Embedded] Broker service error: %v\n", err)
		}
	}()

	// Give services time to start
	time.Sleep(100 * time.Millisecond)
	close(eo.servicesReady)

	// Create agent deployer (connects to embedded services)
	supportAddress := "localhost" + cfg.SupportPort
	eo.agentDeployer = deployer.NewAgentDeployer(supportAddress, cfg.Debug)

	// Load pool config into deployer
	if poolConfig != nil {
		if err := eo.agentDeployer.LoadPool(poolConfig); err != nil {
			return nil, fmt.Errorf("failed to load pool config: %w", err)
		}
	}

	// Wire EventBridge to broker (Phase 3)
	// Note: Currently EventBridge works in-memory for Alfa↔Alfa
	// Full agent→Alfa wiring requires broker API extension
	// For now, embedded services allow agents to communicate via broker
	// while Alfa uses EventBridge for its own events

	if cfg.Debug {
		fmt.Println("[Gox Embedded] Initialized with embedded services")
		fmt.Println("[Gox Embedded] Services: Support(" + cfg.SupportPort + "), Broker(" + cfg.BrokerPort + ")")
	}

	return eo, nil
}

// wireEventBridgeToBroker sets up event forwarding from broker to EventBridge
// TODO Phase 3: Implement once broker has external subscription API
func (eo *EmbeddedOrchestrator) wireEventBridgeToBroker() {
	// This will be implemented when we add OnPublish() hook to broker
	// For now, EventBridge handles Alfa-to-Alfa events
	// Agents communicate via the embedded broker
}

// StartCell starts a cell for a specific project
//
// Phase 2: Actually deploys agents with custom VFS root and environment.
func (eo *EmbeddedOrchestrator) StartCell(cellID string, opts CellOptions) error {
	eo.cellsMutex.Lock()
	defer eo.cellsMutex.Unlock()

	// Check if cell already running for this project
	key := fmt.Sprintf("%s-%s", cellID, opts.ProjectID)
	if _, exists := eo.cells[key]; exists {
		return fmt.Errorf("cell %s already running for project %s", cellID, opts.ProjectID)
	}

	// Set defaults
	if opts.VFSRoot == "" {
		opts.VFSRoot = eo.config.DefaultDataRoot
	}

	// Find cell configuration
	var cellConfig *config.Cell
	if eo.cellsConfig != nil {
		for i := range eo.cellsConfig.Cells {
			if eo.cellsConfig.Cells[i].ID == cellID {
				cellConfig = &eo.cellsConfig.Cells[i]
				break
			}
		}
	}

	if cellConfig == nil {
		return fmt.Errorf("cell %s not found in cells.yaml", cellID)
	}

	if eo.config.Debug {
		fmt.Printf("[Gox Embedded] Starting cell %s for project %s (VFS root: %s)\n",
			cellID, opts.ProjectID, opts.VFSRoot)
		fmt.Printf("[Gox Embedded] Cell has %d agents\n", len(cellConfig.Agents))
	}

	// Deploy each agent in the cell
	for _, agent := range cellConfig.Agents {
		// Create modified agent config with project-specific settings
		agentCopy := agent

		// Build custom environment for VFS root injection
		customEnv := make(map[string]string)
		customEnv["GOX_DATA_ROOT"] = opts.VFSRoot
		customEnv["GOX_PROJECT_ID"] = opts.ProjectID

		// Add user-provided environment variables
		for key, value := range opts.Environment {
			customEnv[key] = value
		}

		if eo.config.Debug {
			fmt.Printf("[Gox Embedded] Deploying agent %s (type: %s) with VFS root: %s\n",
				agent.ID, agent.AgentType, opts.VFSRoot)
		}

		// Deploy the agent with custom environment
		if err := eo.agentDeployer.DeployAgentWithEnv(eo.ctx, agentCopy, customEnv); err != nil {
			return fmt.Errorf("failed to deploy agent %s: %w", agent.ID, err)
		}
	}

	// Create running cell record
	runningCell := &RunningCell{
		CellID:    cellID,
		ProjectID: opts.ProjectID,
		VFSRoot:   opts.VFSRoot,
		Status:    CellStatusRunning,
		StartedAt: time.Now(),
		Options:   opts,
	}

	// Store running cell
	eo.cells[key] = runningCell

	if eo.config.Debug {
		fmt.Printf("[Gox Embedded] Cell %s started successfully for project %s\n",
			cellID, opts.ProjectID)
	}

	return nil
}

// StopCell stops a running cell and terminates all its agents
func (eo *EmbeddedOrchestrator) StopCell(cellID string, projectID string) error {
	eo.cellsMutex.Lock()
	defer eo.cellsMutex.Unlock()

	key := fmt.Sprintf("%s-%s", cellID, projectID)
	runningCell, exists := eo.cells[key]
	if !exists {
		return fmt.Errorf("cell %s not running for project %s", cellID, projectID)
	}

	// Find cell configuration to get agent IDs
	var cellConfig *config.Cell
	if eo.cellsConfig != nil {
		for i := range eo.cellsConfig.Cells {
			if eo.cellsConfig.Cells[i].ID == cellID {
				cellConfig = &eo.cellsConfig.Cells[i]
				break
			}
		}
	}

	// Stop all agents in the cell
	if cellConfig != nil {
		for _, agent := range cellConfig.Agents {
			if err := eo.agentDeployer.StopAgent(agent.ID); err != nil {
				if eo.config.Debug {
					fmt.Printf("[Gox Embedded] Warning: Failed to stop agent %s: %v\n", agent.ID, err)
				}
			}
		}
	}

	runningCell.Status = CellStatusStopped

	// Remove from running cells
	delete(eo.cells, key)

	if eo.config.Debug {
		fmt.Printf("[Gox Embedded] Stopped cell %s for project %s\n", cellID, projectID)
	}

	return nil
}

// StopAll stops all running cells
func (eo *EmbeddedOrchestrator) StopAll() error {
	eo.cellsMutex.Lock()
	cellKeys := make([]string, 0, len(eo.cells))
	for key := range eo.cells {
		cellKeys = append(cellKeys, key)
	}
	eo.cellsMutex.Unlock()

	// Stop each cell
	for _, key := range cellKeys {
		parts := parseKey(key)
		if len(parts) >= 2 {
			eo.StopCell(parts[0], parts[1])
		}
	}

	return nil
}

// Subscribe returns a channel that receives events matching the topic pattern
func (eo *EmbeddedOrchestrator) Subscribe(topicPattern string) <-chan Event {
	return eo.eventBridge.Subscribe(topicPattern)
}

// Unsubscribe closes a subscription
func (eo *EmbeddedOrchestrator) Unsubscribe(topicPattern string, ch <-chan Event) {
	eo.eventBridge.Unsubscribe(topicPattern, ch)
}

// Publish publishes an event to a topic
//
// Phase 1: Events are forwarded to subscribers only (in-memory)
// Phase 2: Will publish to actual Gox broker
func (eo *EmbeddedOrchestrator) Publish(topic string, data interface{}) error {
	// Create event
	event := Event{
		Topic:     topic,
		ProjectID: extractProjectIDFromTopic(topic),
		Timestamp: time.Now(),
		Source:    "host_application",
	}

	// Convert data to map
	if dataMap, ok := data.(map[string]interface{}); ok {
		event.Data = dataMap
	} else {
		event.Data = map[string]interface{}{"payload": data}
	}

	// Forward to subscribers (simplified - no broker yet)
	eo.eventBridge.mutex.RLock()
	defer eo.eventBridge.mutex.RUnlock()

	for pattern, subscribers := range eo.eventBridge.subscribers {
		if eo.eventBridge.topicMatches(topic, pattern) {
			for _, subscriber := range subscribers {
				select {
				case subscriber <- event:
				default:
					// Channel full, drop event
				}
			}
		}
	}

	// TODO Phase 2: Publish to actual broker
	// eo.broker.Publish(topic, data)

	return nil
}

// PublishAndWait publishes a request and waits for a response
func (eo *EmbeddedOrchestrator) PublishAndWait(
	requestTopic string,
	responseTopic string,
	data interface{},
	timeout time.Duration,
) (*Event, error) {
	// Subscribe to response topic
	responseCh := eo.Subscribe(responseTopic)
	defer eo.Unsubscribe(responseTopic, responseCh)

	// Publish request
	if err := eo.Publish(requestTopic, data); err != nil {
		return nil, fmt.Errorf("failed to publish request: %w", err)
	}

	// Wait for response
	select {
	case event := <-responseCh:
		return &event, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for response on topic %s", responseTopic)
	}
}

// ListCells returns information about all running cells
func (eo *EmbeddedOrchestrator) ListCells() []CellInfo {
	eo.cellsMutex.RLock()
	defer eo.cellsMutex.RUnlock()

	cells := make([]CellInfo, 0, len(eo.cells))
	for _, runningCell := range eo.cells {
		cells = append(cells, CellInfo{
			CellID:    runningCell.CellID,
			ProjectID: runningCell.ProjectID,
			VFSRoot:   runningCell.VFSRoot,
			Status:    runningCell.Status,
			StartedAt: runningCell.StartedAt,
		})
	}

	return cells
}

// CellStatusInfo returns the status of a specific cell
func (eo *EmbeddedOrchestrator) CellStatusInfo(cellID string, projectID string) (*CellInfo, error) {
	eo.cellsMutex.RLock()
	defer eo.cellsMutex.RUnlock()

	key := fmt.Sprintf("%s-%s", cellID, projectID)
	runningCell, exists := eo.cells[key]
	if !exists {
		return nil, fmt.Errorf("cell %s not running for project %s", cellID, projectID)
	}

	return &CellInfo{
		CellID:    runningCell.CellID,
		ProjectID: runningCell.ProjectID,
		VFSRoot:   runningCell.VFSRoot,
		Status:    runningCell.Status,
		StartedAt: runningCell.StartedAt,
	}, nil
}

// Close shuts down the embedded orchestrator and all embedded services
func (eo *EmbeddedOrchestrator) Close() error {
	// Stop all cells
	eo.StopAll()

	// Close event bridge
	if eo.eventBridge != nil {
		eo.eventBridge.Close()
	}

	// Cancel context (this stops support and broker services)
	eo.cancel()

	// Give services time to shut down gracefully
	time.Sleep(50 * time.Millisecond)

	if eo.config.Debug {
		fmt.Println("[Gox Embedded] Shut down")
	}

	return nil
}

// Helper functions

func parseKey(key string) []string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '-' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return []string{key}
}

func extractProjectIDFromTopic(topic string) string {
	// Extract from "project-id:topic-name" format
	for i, c := range topic {
		if c == ':' {
			return topic[:i]
		}
	}
	return "default"
}
