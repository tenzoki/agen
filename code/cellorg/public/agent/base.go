// Package agent provides the base agent framework for all GOX v3 agents.
// This package implements the common functionality required by all agents,
// including connection management, configuration handling, lifecycle management,
// and communication with the support service and broker.
//
// Key Features:
// - Automatic connection setup to support service and broker
// - Configuration management with environment variable support
// - Agent lifecycle state management and reporting
// - Standardized logging with debug support
// - Graceful shutdown handling
// - Agent ID resolution with multiple fallback strategies
//
// All specific agent implementations (file-ingester, text-transformer, etc.)
// should embed BaseAgent to inherit these common capabilities while adding
// their domain-specific processing logic.
package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tenzoki/agen/atomic/vfs"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// BaseAgent provides the foundation for all GOX v3 agents.
// It manages connections to the support service and broker, handles configuration
// loading, and provides lifecycle management capabilities. All specific agent
// types should embed this struct to inherit common functionality.
//
// The BaseAgent handles:
// - Service discovery and connection management
// - Configuration loading from support service
// - Agent registration and state reporting
// - Context management for graceful shutdown
// - Standardized logging with agent identification
// - VFS-scoped file operations for project isolation
//
// Thread Safety: BaseAgent methods are generally thread-safe, but specific
// agent implementations should handle their own concurrency requirements.
type BaseAgent struct {
	// Agent identification
	ID        string // Unique agent identifier (e.g., "file-ingester-001")
	AgentType string // Agent type classification (e.g., "file-ingester")
	Debug     bool   // Enable debug logging

	// Service connections
	SupportClient  *client.SupportClient // Connection to support service
	BrokerClient   *client.BrokerClient  // Connection to message broker
	SupportAddress string                // Support service address for logging and retries

	// Configuration and state management
	Config    map[string]interface{} // Agent configuration from support service
	ctx       context.Context        // Agent context for cancellation
	cancel    context.CancelFunc     // Cancellation function for graceful shutdown
	Lifecycle *LifecycleManager      // Agent lifecycle state manager

	// VFS for project-scoped file operations
	VFS       *vfs.VFS // Virtual file system rooted at project directory
	ProjectID string   // Project identifier for multi-tenant isolation
}

// AgentConfig holds the initialization configuration for creating a new agent.
// This structure contains all the parameters needed to set up an agent's
// connections and basic operational parameters.
type AgentConfig struct {
	ID             string        // Unique agent identifier
	AgentType      string        // Agent type (file-ingester, text-transformer, etc.)
	Debug          bool          // Enable debug logging
	SupportAddress string        // Support service address (e.g., "localhost:8080")
	Capabilities   []string      // List of agent capabilities for service discovery
	RebootTimeout  time.Duration // Timeout for agent restart operations

	// VFS configuration
	ProjectID  string // Project identifier for VFS isolation (optional)
	DataRoot   string // Root directory for VFS (defaults to /var/lib/gox or GOX_DATA_ROOT env)
	VFSEnabled bool   // Enable VFS for this agent (default: true)
	ReadOnly   bool   // Create read-only VFS (for query-only agents)
}

// NewBaseAgent creates a new base agent instance with full service integration.
// This constructor handles all the complex setup required for agent operation,
// including service discovery, connection establishment, and configuration loading.
//
// Initialization process:
// 1. Connect to support service for infrastructure discovery
// 2. Discover and connect to message broker
// 3. Register agent with support service
// 4. Load cell-specific configuration
// 5. Initialize lifecycle management
// 6. Transition to configured state
//
// The method performs comprehensive error handling for all setup steps,
// ensuring the agent is fully operational before returning.
//
// Parameters:
//   - config: Agent configuration parameters
//
// Returns:
//   - *BaseAgent: Fully initialized agent ready for operation
//   - error: Setup error (connection, registration, or configuration)
//
// Called by: Specific agent main() functions during startup
func NewBaseAgent(config AgentConfig) (*BaseAgent, error) {
	// Connect to support service with retry logic (15-minute timeout)
	supportClient := client.NewSupportClient(config.SupportAddress, config.Debug)

	// Retry connection to GOX support service for up to 15 minutes
	const maxRetryDuration = 15 * time.Minute
	const retryInterval = 10 * time.Second

	log.Printf("Agent %s: Connecting to GOX at %s (will retry for up to 15 minutes)", config.ID, config.SupportAddress)

	startTime := time.Now()
	var lastErr error

	for time.Since(startTime) < maxRetryDuration {
		if err := supportClient.Connect(); err != nil {
			lastErr = err
			log.Printf("Agent %s: Cannot connect to GOX at %s: %v (retrying in %v)",
				config.ID, config.SupportAddress, err, retryInterval)

			// Provide helpful guidance on first connection attempt
			if time.Since(startTime) < retryInterval*2 {
				log.Printf("Agent %s: If GOX is running on a different host, use: %s -gox-host=<hostname>",
					config.ID, os.Args[0])
				if strings.Contains(config.SupportAddress, "localhost") {
					log.Printf("Agent %s: WARNING: Trying to connect to 'localhost' - ensure GOX is running locally or specify the correct hostname",
						config.ID)
				}
			}

			time.Sleep(retryInterval)
			continue
		}

		// Connection successful
		log.Printf("Agent %s: Successfully connected to GOX at %s", config.ID, config.SupportAddress)
		break
	}

	// Check if we timed out
	if time.Since(startTime) >= maxRetryDuration {
		errorMsg := fmt.Sprintf("failed to connect to GOX at %s after %v: %v",
			config.SupportAddress, maxRetryDuration, lastErr)

		// Add helpful guidance in timeout error
		if strings.Contains(config.SupportAddress, "localhost") {
			errorMsg += "\nHINT: If GOX is running on a different host, use: " + os.Args[0] + " -gox-host=<hostname>"
			errorMsg += "\nHINT: Ensure GOX is running and accessible at " + config.SupportAddress
		} else {
			errorMsg += "\nHINT: Verify GOX is running at " + config.SupportAddress
			errorMsg += "\nHINT: To change the address, use: " + os.Args[0] + " -gox-host=<hostname>"
		}

		return nil, fmt.Errorf("%s", errorMsg)
	}

	// Discover broker address through support service
	brokerInfo, err := supportClient.GetBroker()
	if err != nil {
		return nil, fmt.Errorf("failed to get broker info: %w", err)
	}

	// Establish connection to message broker
	brokerAddress := brokerInfo.Address + brokerInfo.Port
	brokerClient := client.NewBrokerClient(brokerAddress, config.ID, config.Debug)
	if err := brokerClient.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to broker: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create lifecycle manager with support notification callback
	supportNotifyCallback := func(agentID, state string) error {
		// This will be called when agent state changes
		return supportClient.ReportStateChange(agentID, state)
	}
	lifecycle := NewLifecycleManager(config.ID, supportNotifyCallback)

	agent := &BaseAgent{
		ID:             config.ID,
		AgentType:      config.AgentType,
		Debug:          config.Debug,
		SupportClient:  supportClient,
		BrokerClient:   brokerClient,
		SupportAddress: config.SupportAddress,
		Config:         make(map[string]interface{}),
		ctx:            ctx,
		cancel:         cancel,
		Lifecycle:      lifecycle,
		ProjectID:      config.ProjectID,
	}

	// Initialize VFS if enabled (default: true unless explicitly disabled)
	vfsEnabled := true
	if !config.VFSEnabled && config.ProjectID == "" && config.DataRoot == "" {
		vfsEnabled = false // Explicitly disabled
	}
	if vfsEnabled {
		if err := agent.initializeVFS(config); err != nil {
			return nil, fmt.Errorf("failed to initialize VFS: %w", err)
		}
	}

	// Register with support service
	registration := client.AgentRegistration{
		ID:           config.ID,
		AgentType:    config.AgentType,
		Protocol:     "tcp",
		Address:      "localhost",
		Port:         "0",
		Codec:        "json",
		Capabilities: config.Capabilities,
	}

	if err := supportClient.RegisterAgent(registration); err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	// Fetch cell-specific configuration
	cellConfig, err := supportClient.GetAgentCellConfig(config.ID)
	if err != nil {
		agent.LogDebug("No cell-specific config available: %v", err)
		// Not a fatal error - agent can work with default config
	} else {
		// Apply cell configuration to agent
		if cellConfig.Config != nil {
			for key, value := range cellConfig.Config {
				agent.Config[key] = value
			}
		}

		// Store ingress/egress information
		agent.Config["ingress"] = cellConfig.Ingress
		agent.Config["egress"] = cellConfig.Egress

		if config.Debug {
			log.Printf("Agent %s loaded cell config: ingress=%s, egress=%s",
				config.ID, cellConfig.Ingress, cellConfig.Egress)
		}

		// Transition to configured state
		if err := agent.Lifecycle.SetState(StateConfigured, "cell configuration loaded"); err != nil {
			agent.LogError("Failed to transition to configured state: %v", err)
		}
	}

	if config.Debug {
		log.Printf("Agent %s initialized successfully", config.ID)
	}

	return agent, nil
}

// Stop gracefully shuts down the agent and cleans up all resources.
// This method performs orderly shutdown of all agent connections and
// notifies the support service of the agent's termination.
//
// Shutdown process:
// 1. Cancel agent context to signal shutdown to all goroutines
// 2. Disconnect from message broker
// 3. Disconnect from support service
// 4. Clean up any remaining resources
//
// The method is idempotent and can be called multiple times safely.
//
// Returns:
//   - error: Shutdown error or nil on successful cleanup
//
// Called by: Agent signal handlers, main() cleanup, or orchestrator
func (a *BaseAgent) Stop() error {
	if a.Debug {
		log.Printf("Agent %s shutting down", a.ID)
	}

	// Cancel agent context to signal shutdown to all goroutines
	a.cancel()

	// Disconnect from broker to stop message processing
	if a.BrokerClient != nil {
		a.BrokerClient.Disconnect()
	}

	// Disconnect from support service to deregister agent
	if a.SupportClient != nil {
		a.SupportClient.Disconnect()
	}

	return nil
}

// GetConfigString retrieves a string configuration value
func (a *BaseAgent) GetConfigString(key, defaultValue string) string {
	if value, exists := a.Config[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetConfigBool retrieves a boolean configuration value
func (a *BaseAgent) GetConfigBool(key string, defaultValue bool) bool {
	if value, exists := a.Config[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// GetConfigInt retrieves an integer configuration value
func (a *BaseAgent) GetConfigInt(key string, defaultValue int) int {
	if value, exists := a.Config[key]; exists {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// GetIngress returns the agent's ingress configuration
func (a *BaseAgent) GetIngress() string {
	return a.GetConfigString("ingress", "")
}

// GetEgress returns the agent's egress configuration
func (a *BaseAgent) GetEgress() string {
	return a.GetConfigString("egress", "")
}

// GetSupportAddress returns the support service address
func (a *BaseAgent) GetSupportAddress() string {
	return a.SupportAddress
}

// Log helper functions
func (a *BaseAgent) LogInfo(format string, args ...interface{}) {
	log.Printf("Agent %s: "+format, append([]interface{}{a.ID}, args...)...)
}

func (a *BaseAgent) LogDebug(format string, args ...interface{}) {
	if a.Debug {
		log.Printf("Agent %s [DEBUG]: "+format, append([]interface{}{a.ID}, args...)...)
	}
}

func (a *BaseAgent) LogError(format string, args ...interface{}) {
	log.Printf("Agent %s [ERROR]: "+format, append([]interface{}{a.ID}, args...)...)
}

// Context returns the agent's context for cancellation
func (a *BaseAgent) Context() context.Context {
	return a.ctx
}

// GetAgentID resolves the agent's unique identifier using multiple fallback strategies.
// This function provides flexible agent ID resolution to support different deployment
// scenarios from development to production environments.
//
// Resolution priority:
// 1. Command-line argument: --id=custom-agent-id
// 2. Environment variable: GOX_AGENT_ID
// 3. Auto-generated: agentType-hostname-pid
//
// The auto-generated ID ensures uniqueness across multiple agent instances
// running on the same or different machines.
//
// Parameters:
//   - agentType: Agent type for auto-generation (e.g., "file-ingester")
//
// Returns:
//   - string: Resolved agent ID
//
// Called by: Agent main() functions during initialization
func GetAgentID(agentType string) string {
	// Priority 1: CLI argument --id=agent-id
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--id=") {
			return strings.TrimPrefix(arg, "--id=")
		}
	}

	// Priority 2: Environment variable
	if id := os.Getenv("GOX_AGENT_ID"); id != "" {
		return id
	}

	// Priority 3: Auto-generate unique ID using hostname and process ID
	hostname, _ := os.Hostname()
	pid := os.Getpid()
	return fmt.Sprintf("%s-%s-%d", agentType, hostname, pid)
}

// GetAgentType resolves agent type from CLI args or environment
func GetAgentType(defaultType string) string {
	// Priority 1: CLI argument --type=agent-type
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "--type=") {
			return strings.TrimPrefix(arg, "--type=")
		}
	}

	// Priority 2: Environment variable
	if agentType := os.Getenv("GOX_AGENT_TYPE"); agentType != "" {
		return agentType
	}

	// Priority 3: Default
	return defaultType
}

// Helper to get configuration from environment
func GetEnvConfig(key, defaultValue string) string {
	if value := os.Getenv("GOX_" + key); value != "" {
		return value
	}
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDebugFromEnv checks for debug flag
func GetDebugFromEnv() bool {
	if os.Getenv("GOX_DEBUG") == "true" {
		return true
	}
	for _, arg := range os.Args {
		if arg == "--debug" {
			return true
		}
	}
	return false
}

// initializeVFS creates and initializes the VFS for the agent
func (b *BaseAgent) initializeVFS(config AgentConfig) error {
	// Determine data root
	dataRoot := config.DataRoot
	if dataRoot == "" {
		dataRoot = os.Getenv("GOX_DATA_ROOT")
	}
	if dataRoot == "" {
		dataRoot = "/var/lib/gox"
	}

	// Determine project ID
	projectID := config.ProjectID
	if projectID == "" {
		projectID = os.Getenv("GOX_PROJECT_ID")
	}
	if projectID == "" {
		projectID = "default" // Default project for backward compatibility
	}

	// Create VFS root path: dataRoot/projects/projectID
	vfsRoot := filepath.Join(dataRoot, "projects", projectID)

	// Initialize VFS
	projectVFS, err := vfs.NewVFS(vfsRoot, config.ReadOnly)
	if err != nil {
		return fmt.Errorf("failed to create VFS at %s: %w", vfsRoot, err)
	}

	b.VFS = projectVFS
	b.ProjectID = projectID

	if b.Debug {
		log.Printf("Agent %s: VFS initialized at %s (project: %s, readonly: %v)",
			b.ID, vfsRoot, projectID, config.ReadOnly)
	}

	return nil
}

// VFS Helper Methods for convenient access

// ReadFile reads a file from the VFS
func (b *BaseAgent) ReadFile(parts ...string) ([]byte, error) {
	if b.VFS == nil {
		return nil, fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.Read(parts...)
}

// ReadFileString reads a file from the VFS as a string
func (b *BaseAgent) ReadFileString(parts ...string) (string, error) {
	if b.VFS == nil {
		return "", fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.ReadString(parts...)
}

// WriteFile writes content to a file in the VFS
func (b *BaseAgent) WriteFile(content []byte, parts ...string) error {
	if b.VFS == nil {
		return fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.Write(content, parts...)
}

// WriteFileString writes a string to a file in the VFS
func (b *BaseAgent) WriteFileString(content string, parts ...string) error {
	if b.VFS == nil {
		return fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.WriteString(content, parts...)
}

// FileExists checks if a file exists in the VFS
func (b *BaseAgent) FileExists(parts ...string) bool {
	if b.VFS == nil {
		return false
	}
	return b.VFS.Exists(parts...)
}

// MkdirVFS creates a directory in the VFS
func (b *BaseAgent) MkdirVFS(parts ...string) error {
	if b.VFS == nil {
		return fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.Mkdir(parts...)
}

// DeleteFile deletes a file or directory in the VFS
func (b *BaseAgent) DeleteFile(parts ...string) error {
	if b.VFS == nil {
		return fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.Delete(parts...)
}

// ListFiles lists files in a directory in the VFS
func (b *BaseAgent) ListFiles(parts ...string) ([]os.FileInfo, error) {
	if b.VFS == nil {
		return nil, fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.List(parts...)
}

// VFSPath returns the absolute path for relative VFS parts
func (b *BaseAgent) VFSPath(parts ...string) (string, error) {
	if b.VFS == nil {
		return "", fmt.Errorf("VFS not initialized for agent %s", b.ID)
	}
	return b.VFS.Path(parts...)
}

// VFSRoot returns the VFS root directory
func (b *BaseAgent) VFSRoot() string {
	if b.VFS == nil {
		return ""
	}
	return b.VFS.Root()
}

// SetVFSRoot sets a custom VFS root for the agent.
// This is used by the embedded orchestrator to override the VFS root per cell instance,
// allowing multiple projects to use the same cell definition with different VFS roots.
//
// Parameters:
//   - vfsRoot: Absolute path to the VFS root directory
//   - readOnly: Whether the VFS should be read-only
//
// Returns:
//   - error: VFS initialization error or nil on success
//
// Called by: pkg/orchestrator during cell startup with custom VFS roots
func (b *BaseAgent) SetVFSRoot(vfsRoot string, readOnly bool) error {
	// Create new VFS instance with custom root
	projectVFS, err := vfs.NewVFS(vfsRoot, readOnly)
	if err != nil {
		return fmt.Errorf("failed to create VFS at %s: %w", vfsRoot, err)
	}

	// Replace existing VFS
	b.VFS = projectVFS

	if b.Debug {
		log.Printf("Agent %s: VFS root set to %s (readonly: %v)", b.ID, vfsRoot, readOnly)
	}

	return nil
}
