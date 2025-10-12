package orchestrator

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/broker"
	"github.com/tenzoki/agen/cellorg/internal/config"
	"github.com/tenzoki/agen/cellorg/internal/deployer"
	"github.com/tenzoki/agen/cellorg/internal/support"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// EmbeddedOrchestrator runs cellorg orchestrator in-process
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
	config         Config
	eventBridge    *EventBridge
	cells          map[string]*RunningCell
	cellsMutex     sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	agentDeployer  *deployer.AgentDeployer
	cellorgConfig  *config.Config
	cellsConfig    *config.CellsConfig
	poolConfig     *config.PoolConfig

	// Embedded services (Phase 3)
	supportService *support.Service
	brokerService  *broker.Service
	brokerClient   *client.BrokerClient
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
// - Event bridging (cellorg topics → Go channels)
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
		cfg.DefaultDataRoot = "/var/lib/cellorg"
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

	// Load Cellorg configuration (cellorg.yaml)
	cellorgConfigPath := cfg.ConfigPath + "/cellorg.yaml"
	cellorgConfig, err := config.Load(cellorgConfigPath)
	if err != nil {
		// Use defaults if config not found
		if cfg.Debug {
			fmt.Printf("[Cellorg Embedded] Warning: Could not load cellorg.yaml: %v, using defaults\n", err)
		}
		cellorgConfig = &config.Config{
			Debug:   cfg.Debug,
			Support: config.SupportConfig{Port: cfg.SupportPort},
			Broker:  config.BrokerConfig{Port: cfg.BrokerPort},
			BaseDir: []string{cfg.ConfigPath},
			Pool:    []string{"pool.yaml"},
			Cells:   []string{"cells.yaml"},
		}
	}
	eo.cellorgConfig = cellorgConfig

	// Load pool configuration
	poolConfig, err := cellorgConfig.LoadPool()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("[Cellorg Embedded] Warning: Could not load pool.yaml: %v\n", err)
		}
	}
	eo.poolConfig = poolConfig

	// Load cells configuration
	cellsConfig, err := cellorgConfig.LoadCells()
	if err != nil {
		if cfg.Debug {
			fmt.Printf("[Cellorg Embedded] Warning: Could not load cells: %v\n", err)
		}
	}
	eo.cellsConfig = cellsConfig

	if cfg.Debug && cellsConfig != nil {
		fmt.Printf("[Cellorg Embedded] Loaded %d cells:\n", len(cellsConfig.Cells))
		for _, cell := range cellsConfig.Cells {
			fmt.Printf("[Cellorg Embedded]   - %s (%d agents)\n", cell.ID, len(cell.Agents))
		}
	}

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
			fmt.Printf("[Cellorg Embedded] Warning: Could not load agent types: %v\n", err)
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
			fmt.Println("[Cellorg Embedded] Starting support service...")
		}
		if err := eo.supportService.Start(eo.ctx); err != nil && eo.ctx.Err() == nil {
			fmt.Printf("[Cellorg Embedded] Support service error: %v\n", err)
		}
	}()

	go func() {
		if cfg.Debug {
			fmt.Println("[Cellorg Embedded] Starting broker service...")
		}
		if err := eo.brokerService.Start(eo.ctx); err != nil && eo.ctx.Err() == nil {
			fmt.Printf("[Cellorg Embedded] Broker service error: %v\n", err)
		}
	}()

	// Give services time to start
	time.Sleep(100 * time.Millisecond)

	// Register broker with support service so agents can discover it
	brokerInfo := broker.Info{
		Protocol: "tcp",
		Address:  "localhost",
		Port:     cfg.BrokerPort,
		Codec:    "json",
	}
	if err := eo.supportService.SetBrokerAddress(brokerInfo); err != nil {
		return nil, fmt.Errorf("failed to register broker with support: %w", err)
	}

	if cfg.Debug {
		fmt.Printf("[Cellorg Embedded] Broker registered with support service: %s%s\n", brokerInfo.Address, brokerInfo.Port)
	}

	close(eo.servicesReady)

	// Create broker client for alfa to communicate with agents
	brokerAddress := "localhost" + cfg.BrokerPort
	eo.brokerClient = client.NewBrokerClient(brokerAddress, "alfa-orchestrator", cfg.Debug)
	if err := eo.brokerClient.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect broker client: %w", err)
	}

	if cfg.Debug {
		fmt.Printf("[Cellorg Embedded] Broker client connected to %s\n", brokerAddress)
	}

	// Create agent deployer (connects to embedded services)
	supportAddress := "localhost" + cfg.SupportPort
	// Determine framework root from ConfigPath (ConfigPath is workbench/config)
	frameworkRoot := filepath.Dir(filepath.Dir(cfg.ConfigPath))
	eo.agentDeployer = deployer.NewAgentDeployer(supportAddress, frameworkRoot, cfg.Debug)

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
		fmt.Println("[Cellorg Embedded] Initialized with embedded services")
		fmt.Println("[Cellorg Embedded] Services: Support(" + cfg.SupportPort + "), Broker(" + cfg.BrokerPort + ")")
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
		fmt.Printf("[Cellorg Embedded] Starting cell %s for project %s (VFS root: %s)\n",
			cellID, opts.ProjectID, opts.VFSRoot)
		fmt.Printf("[Cellorg Embedded] Cell has %d agents\n", len(cellConfig.Agents))
	}

	// Deploy each agent in the cell
	for _, agent := range cellConfig.Agents {
		// Create modified agent config with project-specific settings
		agentCopy := agent

		// Build custom environment for VFS root injection
		customEnv := make(map[string]string)
		customEnv["CELLORG_DATA_ROOT"] = opts.VFSRoot
		customEnv["CELLORG_PROJECT_ID"] = opts.ProjectID

		// Add user-provided environment variables
		for key, value := range opts.Environment {
			customEnv[key] = value
		}

		if eo.config.Debug {
			fmt.Printf("[Cellorg Embedded] Deploying agent %s (type: %s) with VFS root: %s\n",
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
		fmt.Printf("[Cellorg Embedded] Cell %s started successfully for project %s\n",
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
					fmt.Printf("[Cellorg Embedded] Warning: Failed to stop agent %s: %v\n", agent.ID, err)
				}
			}
		}
	}

	runningCell.Status = CellStatusStopped

	// Remove from running cells
	delete(eo.cells, key)

	if eo.config.Debug {
		fmt.Printf("[Cellorg Embedded] Stopped cell %s for project %s\n", cellID, projectID)
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

// Subscribe returns a channel that receives events from the broker
func (eo *EmbeddedOrchestrator) Subscribe(topicPattern string) <-chan Event {
	// Subscribe to broker using broker client
	if eo.brokerClient != nil {
		brokerCh, err := eo.brokerClient.Subscribe(topicPattern)
		if err != nil {
			// Return empty channel on error
			emptyCh := make(chan Event)
			close(emptyCh)
			return emptyCh
		}

		// Create Event channel to return to caller
		eventCh := make(chan Event, 100)

		// Start goroutine to convert BrokerMessages to Events
		go func() {
			defer close(eventCh)
			for brokerMsg := range brokerCh {
				event := Event{
					Topic:     brokerMsg.Target,
					ProjectID: extractProjectIDFromTopic(brokerMsg.Target),
					Timestamp: time.Now(),
					Source:    "broker",
				}

				// Convert payload to map
				if dataMap, ok := brokerMsg.Payload.(map[string]interface{}); ok {
					event.Data = dataMap
				} else {
					event.Data = map[string]interface{}{"payload": brokerMsg.Payload}
				}

				select {
				case eventCh <- event:
				case <-eo.ctx.Done():
					return
				}
			}
		}()

		return eventCh
	}

	// Fallback to EventBridge if broker client not available
	return eo.eventBridge.Subscribe(topicPattern)
}

// Unsubscribe closes a subscription
func (eo *EmbeddedOrchestrator) Unsubscribe(topicPattern string, ch <-chan Event) {
	// Note: Broker client subscriptions are automatically cleaned up when channel is closed
	// EventBridge subscriptions are handled by EventBridge itself
	// For now, this is a no-op, but we keep the method for API compatibility
}

// Publish publishes an event to a topic via the broker
//
// Publishes to the actual cellorg broker so agents can receive messages
func (eo *EmbeddedOrchestrator) Publish(topic string, data interface{}) error {
	// Publish to broker using broker client
	if eo.brokerClient != nil {
		// Create BrokerMessage
		brokerMsg := client.BrokerMessage{
			ID:        fmt.Sprintf("alfa_%d", time.Now().UnixNano()),
			Type:      "event",
			Target:    fmt.Sprintf("pub:%s", topic),
			Payload:   data,
			Meta:      make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		// Extract type from data if present
		if dataMap, ok := data.(map[string]interface{}); ok {
			if msgType, exists := dataMap["type"]; exists {
				if msgTypeStr, ok := msgType.(string); ok {
					brokerMsg.Type = msgTypeStr
				}
			}
		}

		return eo.brokerClient.Publish(topic, brokerMsg)
	}

	return fmt.Errorf("broker client not initialized")
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

	// Disconnect broker client
	if eo.brokerClient != nil {
		eo.brokerClient.Disconnect()
	}

	// Close event bridge
	if eo.eventBridge != nil {
		eo.eventBridge.Close()
	}

	// Cancel context (this stops support and broker services)
	eo.cancel()

	// Give services time to shut down gracefully
	time.Sleep(50 * time.Millisecond)

	if eo.config.Debug {
		fmt.Println("[Cellorg Embedded] Shut down")
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
