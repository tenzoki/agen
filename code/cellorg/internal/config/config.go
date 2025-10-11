package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppName string `yaml:"app_name"`
	Debug   bool   `yaml:"debug"`

	Support SupportConfig `yaml:"support"`
	Broker  BrokerConfig  `yaml:"broker"`

	BaseDir []string `yaml:"basedir"`
	Pool    []string `yaml:"pool"`
	Cells   []string `yaml:"cells"`

	AwaitTimeoutSeconds       int `yaml:"await-timeout_seconds"`
	AwaitSupportRebootSeconds int `yaml:"await_support_reboot_seconds"`
}

type SupportConfig struct {
	Port  string `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

type BrokerConfig struct {
	Port     string `yaml:"port"`
	Protocol string `yaml:"protocol"`
	Codec    string `yaml:"codec"`
	Debug    bool   `yaml:"debug"`
}

type PoolConfig struct {
	AgentTypes []AgentTypeConfig `yaml:"agent_types"`
}

type AgentTypeConfig struct {
	AgentType    string   `yaml:"agent_type"`
	Binary       string   `yaml:"binary"`
	Operator     string   `yaml:"operator"`
	Capabilities []string `yaml:"capabilities"`
	Description  string   `yaml:"description"`
}

type CellsConfig struct {
	Cells []Cell `yaml:"cells,omitempty"`
}

type Cell struct {
	ID          string      `yaml:"id"`
	Description string      `yaml:"description"`
	Debug       bool        `yaml:"debug"`
	Agents      []CellAgent `yaml:"agents"`
}

type CellAgent struct {
	ID        string                 `yaml:"id"`
	AgentType string                 `yaml:"agent_type"`
	Ingress   string                 `yaml:"ingress"`
	Egress    string                 `yaml:"egress"`
	Config    map[string]interface{} `yaml:"config,omitempty"`
}

func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Support.Port == "" {
		config.Support.Port = ":9000"
	}
	if config.Broker.Port == "" {
		config.Broker.Port = ":9001"
	}
	if config.Broker.Protocol == "" {
		config.Broker.Protocol = "tcp"
	}
	if config.Broker.Codec == "" {
		config.Broker.Codec = "json"
	}
	if config.AwaitTimeoutSeconds == 0 {
		config.AwaitTimeoutSeconds = 300
	}
	if config.AwaitSupportRebootSeconds == 0 {
		config.AwaitSupportRebootSeconds = 300
	}

	// Validate configuration values
	if config.AwaitTimeoutSeconds < 0 {
		return nil, fmt.Errorf("await timeout seconds cannot be negative: %d", config.AwaitTimeoutSeconds)
	}
	if config.AwaitSupportRebootSeconds < 0 {
		return nil, fmt.Errorf("await support reboot seconds cannot be negative: %d", config.AwaitSupportRebootSeconds)
	}

	return &config, nil
}

func (c *Config) LoadPool() (*PoolConfig, error) {
	if len(c.Pool) == 0 {
		return &PoolConfig{}, nil
	}

	poolFile := c.Pool[0]
	if !filepath.IsAbs(poolFile) {
		if len(c.BaseDir) > 0 {
			poolFile = filepath.Join(c.BaseDir[0], poolFile)
		}
	}

	data, err := os.ReadFile(poolFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read pool file %s: %w", poolFile, err)
	}

	var poolConfig struct {
		Pool PoolConfig `yaml:"pool"`
	}
	if err := yaml.Unmarshal(data, &poolConfig); err != nil {
		return nil, fmt.Errorf("failed to parse pool file %s: %w", poolFile, err)
	}

	return &poolConfig.Pool, nil
}

func (c *Config) LoadCells() (*CellsConfig, error) {
	if len(c.Cells) == 0 {
		return &CellsConfig{}, nil
	}

	var cells []Cell

	// Process each cells pattern (supports globs)
	for _, cellsPattern := range c.Cells {
		originalPattern := cellsPattern
		if !filepath.IsAbs(cellsPattern) {
			if len(c.BaseDir) > 0 {
				cellsPattern = filepath.Join(c.BaseDir[0], cellsPattern)
			}
		}

		// Expand glob pattern
		matches, err := filepath.Glob(cellsPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid glob pattern %s: %w", cellsPattern, err)
		}

		if c.Debug {
			fmt.Printf("[Config] Pattern '%s' -> '%s' matched %d files\n", originalPattern, cellsPattern, len(matches))
		}

		// Load each matched file
		for _, cellsFile := range matches {
			if c.Debug {
				fmt.Printf("[Config] Loading cell file: %s\n", cellsFile)
			}

			data, err := os.ReadFile(cellsFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read cells file %s: %w", cellsFile, err)
			}

			// Handle multiple YAML documents separated by ---
			decoder := yaml.NewDecoder(bytes.NewReader(data))
			for {
				var cellDoc struct {
					Cell Cell `yaml:"cell"`
				}
				if err := decoder.Decode(&cellDoc); err != nil {
					if err.Error() == "EOF" {
						break
					}
					return nil, fmt.Errorf("failed to parse cells file %s: %w", cellsFile, err)
				}
				if cellDoc.Cell.ID != "" {
					if c.Debug {
						fmt.Printf("[Config]   Found cell: %s\n", cellDoc.Cell.ID)
					}
					cells = append(cells, cellDoc.Cell)
				}
			}
		}
	}

	return &CellsConfig{Cells: cells}, nil
}

// Helper function to convert timeout strings to integers
func ParseTimeout(timeoutStr string) (int, error) {
	if timeoutStr == "" {
		return 0, nil
	}
	return strconv.Atoi(timeoutStr)
}

// ValidateConfiguration validates that cells reference valid agent types and binaries exist
func ValidateConfiguration(pool *PoolConfig, cells *CellsConfig) error {
	// Build agent type lookup map
	agentTypes := make(map[string]*AgentTypeConfig)
	for i := range pool.AgentTypes {
		agentType := &pool.AgentTypes[i]
		agentTypes[agentType.AgentType] = agentType
	}

	// Validate each cell
	var errors []string

	for _, cell := range cells.Cells {
		// Validate each agent in cell
		for _, agent := range cell.Agents {
			// Check if agent type exists in pool
			agentTypeDef, exists := agentTypes[agent.AgentType]
			if !exists {
				errors = append(errors, fmt.Sprintf(
					"cell '%s': agent '%s' references unknown agent_type '%s' (not found in pool.yaml)",
					cell.ID, agent.ID, agent.AgentType))
				continue
			}

			// Check if binary exists (only for spawn/call operators, not await)
			if agentTypeDef.Operator != "await" && agentTypeDef.Binary != "" {
				if !fileExists(agentTypeDef.Binary) {
					errors = append(errors, fmt.Sprintf(
						"cell '%s': agent '%s' (type '%s') references binary '%s' which does not exist",
						cell.ID, agent.ID, agent.AgentType, agentTypeDef.Binary))
				}
			}
		}
	}

	if len(errors) > 0 {
		errMsg := "Configuration validation failed:\n"
		for _, err := range errors {
			errMsg += "  - " + err + "\n"
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
