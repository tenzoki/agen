// Package main provides the central GOX framework orchestrator that coordinates
// the entire agent-based processing system.
//
// GOX is an agent orchestration framework that enables distributed processing
// pipelines through message-passing between specialized agents. The main entry
// point manages the lifecycle of core services (Support, Broker) and deploys
// agent instances based on YAML configuration.
//
// Architecture Overview:
// - Support Service: Agent registry and health monitoring
// - Broker Service: Message routing and pub/sub coordination
// - Agent Deployer: Spawns and manages agent instances
// - Configuration System: YAML-based pipeline and agent definitions
//
// Called by: External processes (CLI, containers, orchestration systems)
// Calls: All internal GOX framework services and agent operators
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tenzoki/agen/cellorg/internal/broker"
	"github.com/tenzoki/agen/cellorg/internal/config"
	"github.com/tenzoki/agen/cellorg/internal/deployer"
	"github.com/tenzoki/agen/cellorg/internal/support"
)

// main is the entry point for the GOX orchestrator.
//
// Configuration Loading Strategy:
// 1. Command line argument: Uses specified config file path
// 2. Default file: Attempts to load gox.yaml from config directory
// 3. Hardcoded defaults: Falls back to built-in configuration
//
// Called by: Operating system process execution
// Calls: config.Load(), getDefaultConfig(), service initialization functions
func main() {
	var cfg *config.Config
	var configSource string

	// Determine config source using priority hierarchy
	if len(os.Args) >= 2 {
		// Use provided config file path from command line
		configFile := os.Args[1]
		loadedCfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("Failed to load config from %s: %v", configFile, err)
		}
		cfg = loadedCfg
		configSource = fmt.Sprintf("config file: %s", configFile)
	} else {
		// Try to load gox.yaml from config directory as default
		if _, err := os.Stat("config/gox.yaml"); err == nil {
			loadedCfg, err := config.Load("config/gox.yaml")
			if err != nil {
				log.Printf("Warning: config/gox.yaml exists but failed to load: %v", err)
				log.Printf("Using hardcoded defaults instead")
				cfg = getDefaultConfig()
				configSource = "hardcoded defaults (config/gox.yaml failed to parse)"
			} else {
				cfg = loadedCfg
				configSource = "config/gox.yaml (default)"
			}
		} else {
			// Use hardcoded defaults when no config file is available
			log.Printf("No config file specified and config/gox.yaml not found")
			cfg = getDefaultConfig()
			configSource = "hardcoded defaults"
		}
	}

	// Inform user about configuration source and debug status
	log.Printf("Starting Gox using %s", configSource)

	if cfg.Debug {
		log.Printf("Debug enabled for app: %s", cfg.AppName)
	}

	// Initialize cancellation context for graceful shutdown coordination
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start Support Service first - it provides agent registry and health monitoring
	// that other services depend on for agent discovery and lifecycle management
	supportService := support.NewService(cfg.Support)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := supportService.Start(ctx); err != nil {
			log.Printf("Support service error: %v", err)
		}
	}()

	// Brief delay to ensure Support service is ready before Broker attempts connection
	time.Sleep(100 * time.Millisecond)

	// Start Broker Service - handles message routing between agents using pub/sub
	brokerService := broker.NewService(cfg.Broker)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := brokerService.Start(ctx); err != nil {
			log.Printf("Broker service error: %v", err)
		}
	}()

	// Register broker connection details with support service for agent discovery
	brokerInfo := broker.Info{
		Protocol: cfg.Broker.Protocol,
		Address:  "localhost",
		Port:     cfg.Broker.Port,
		Codec:    cfg.Broker.Codec,
	}
	if err := supportService.SetBrokerAddress(brokerInfo); err != nil {
		log.Printf("Failed to register broker with support: %v", err)
	}

	log.Printf("Gox v3 started: %s", cfg.AppName)
	log.Printf("Support service on: %s", cfg.Support.Port)
	log.Printf("Broker service on: %s (%s/%s)", cfg.Broker.Port, cfg.Broker.Protocol, cfg.Broker.Codec)

	// Wait for services to be fully ready before agent deployment
	time.Sleep(500 * time.Millisecond)

	// Create agent deployer that manages agent lifecycle and process spawning
	agentDeployer := deployer.NewAgentDeployer("localhost"+cfg.Support.Port, cfg.Debug)

	// Load pool configuration (agent type definitions) and register with support service
	var poolConfig *config.PoolConfig
	if len(cfg.Pool) > 0 {
		poolFile := cfg.Pool[0]
		// Resolve relative paths against base directory
		if len(cfg.BaseDir) > 0 && poolFile[0] != '/' {
			poolFile = cfg.BaseDir[0] + "/" + poolFile
		}
		// Register agent types from config/pool.yaml with support service for discovery
		if err := supportService.LoadAgentTypesFromFile(poolFile); err != nil {
			log.Printf("Warning: failed to load agent types: %v", err)
		}

		// Also load pool config for deployer to understand agent deployment strategies
		var err error
		poolConfig, err = cfg.LoadPool()
		if err != nil {
			log.Printf("Warning: failed to load pool config for deployer: %v", err)
		} else {
			if err := agentDeployer.LoadPool(poolConfig); err != nil {
				log.Printf("Warning: failed to configure deployer: %v", err)
			}
		}
	}

	// Load cells configuration (agent instances and pipelines) and deploy agents
	cellsConfig, err := cfg.LoadCells()
	if err != nil {
		log.Printf("Warning: failed to load cells configuration: %v", err)
	} else if len(cellsConfig.Cells) > 0 {
		log.Printf("Loaded %d cells from configuration", len(cellsConfig.Cells))

		// Deploy agents from cells, respecting dependencies and orchestration settings
		totalAgents := 0
		for _, cell := range cellsConfig.Cells {
			log.Printf("Deploying agents from cell: %s", cell.ID)
			for _, agent := range cell.Agents {
				totalAgents++
				// Deploy agent based on its operator strategy (spawn, call, await)
				if err := agentDeployer.DeployAgent(ctx, agent); err != nil {
					log.Printf("Failed to deploy agent %s: %v", agent.ID, err)
					// Continue with other agents even if one fails to maintain pipeline resilience
				}
			}
		}
		log.Printf("Deployment complete: %d agents processed", totalAgents)
	}

	// Handle graceful shutdown signals from operating system
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until shutdown signal received or context cancelled
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %s, shutting down...", sig)
	case <-ctx.Done():
		log.Printf("Context cancelled, shutting down...")
	}

	// Stop all deployed agents before shutting down core services
	log.Printf("Stopping deployed agents...")
	agentDeployer.StopAll()

	// Cancel context to signal all services to shut down
	cancel()

	// Wait for services to shut down gracefully with timeout protection
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All services shut down successfully")
	case <-time.After(10 * time.Second):
		log.Println("Shutdown timeout exceeded")
	}
}

// getDefaultConfig returns hardcoded default configuration for GOX framework.
//
// This fallback configuration is used when:
// - No command line config file is specified
// - gox.yaml is not found in current directory
// - gox.yaml exists but contains parsing errors
//
// Default Configuration Includes:
// - Support service on port 9000 (agent registry and health monitoring)
// - Broker service on port 9001 with TCP/JSON protocol
// - Debug mode enabled for development visibility
// - Standard YAML file locations (config/pool.yaml, config/cells.yaml)
// - Conservative timeout values for production stability
//
// Called by: main() when no valid configuration file is available
// Calls: None (pure configuration data)
func getDefaultConfig() *config.Config {
	return &config.Config{
		AppName: "gox-default",
		Debug:   true,
		Support: config.SupportConfig{
			Port:  ":9000", // Agent registry and health monitoring service
			Debug: true,
		},
		Broker: config.BrokerConfig{
			Port:     ":9001", // Message routing and pub/sub service
			Protocol: "tcp",   // TCP transport for reliable message delivery
			Codec:    "json",  // JSON encoding for human-readable debugging
			Debug:    true,
		},
		BaseDir:                   []string{"config/"},    // Config directory as base for relative paths
		Pool:                      []string{"pool.yaml"},  // Agent type definitions
		Cells:                     []string{"cells.yaml"}, // Agent instance and pipeline configurations
		AwaitTimeoutSeconds:       300,                    // 5 minutes for agent startup
		AwaitSupportRebootSeconds: 300,                    // 5 minutes for support service recovery
	}
}
