package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// AlfaConfig represents the complete alfa configuration
type AlfaConfig struct {
	// Workbench configuration
	Workbench WorkbenchConfig `yaml:"workbench" json:"workbench"`

	// AI provider configuration
	AI AIConfig `yaml:"ai" json:"ai"`

	// Voice configuration
	Voice VoiceConfig `yaml:"voice" json:"voice"`

	// Execution configuration
	Execution ExecutionConfig `yaml:"execution" json:"execution"`

	// Sandbox configuration
	Sandbox SandboxConfig `yaml:"sandbox" json:"sandbox"`

	// Cellorg configuration
	Cellorg CellorgConfig `yaml:"cellorg" json:"cellorg"`

	// Output configuration
	Output OutputConfig `yaml:"output" json:"output"`

	// Self-modification configuration
	SelfModify SelfModifyConfig `yaml:"self_modify" json:"self_modify"`
}

// WorkbenchConfig defines workbench settings
type WorkbenchConfig struct {
	Path    string `yaml:"path" json:"path"`         // Workbench directory path
	Project string `yaml:"project" json:"project"`   // Default project name
}

// AIConfig defines AI provider settings
type AIConfig struct {
	Provider      string                    `yaml:"provider" json:"provider"`             // "anthropic" or "openai"
	SelectedModel string                    `yaml:"selected_model" json:"selected_model"` // Selected model (optional, uses provider default if empty)
	ConfigFile    string                    `yaml:"config_file" json:"config_file"`       // Path to ai-config.json (optional)
	Providers     map[string]ProviderConfig `yaml:"providers" json:"providers"`           // Provider-specific configs
}

// ProviderConfig defines provider-specific settings with multiple model support
type ProviderConfig struct {
	DefaultModel string                 `yaml:"default_model" json:"default_model"` // Default model for this provider
	Models       map[string]ModelConfig `yaml:"models" json:"models"`               // Available models with their configs
}

// ModelConfig defines configuration for a specific model
type ModelConfig struct {
	MaxTokens   int           `yaml:"max_tokens" json:"max_tokens"`
	Temperature float64       `yaml:"temperature" json:"temperature"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	RetryCount  int           `yaml:"retry_count" json:"retry_count"`
	RetryDelay  time.Duration `yaml:"retry_delay" json:"retry_delay"`
	Description string        `yaml:"description,omitempty" json:"description,omitempty"` // Optional model description
}

// VoiceConfig defines voice input/output settings
type VoiceConfig struct {
	InputEnabled  bool `yaml:"input_enabled" json:"input_enabled"`   // Enable voice input (STT)
	OutputEnabled bool `yaml:"output_enabled" json:"output_enabled"` // Enable voice output (TTS)
	Headless      bool `yaml:"headless" json:"headless"`             // Headless mode (voice + auto-confirm)
}

// ExecutionConfig defines execution behavior
type ExecutionConfig struct {
	AutoConfirm   bool `yaml:"auto_confirm" json:"auto_confirm"`     // Auto-confirm operations
	MaxIterations int  `yaml:"max_iterations" json:"max_iterations"` // Max AI iterations
}

// SandboxConfig defines sandbox execution settings
type SandboxConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"` // Use Docker sandbox
	Image   string `yaml:"image" json:"image"`     // Docker image
}

// CellorgConfig defines cellorg integration settings
type CellorgConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`         // Enable cellorg features
	ConfigPath string `yaml:"config_path" json:"config_path"` // Cellorg config directory
}

// OutputConfig defines output capture settings
type OutputConfig struct {
	CaptureEnabled bool `yaml:"capture_enabled" json:"capture_enabled"` // Capture command output
	MaxSizeKB      int  `yaml:"max_size_kb" json:"max_size_kb"`         // Max output size in KB
}

// SelfModifyConfig defines self-modification settings
type SelfModifyConfig struct {
	Allowed bool `yaml:"allowed" json:"allowed"` // Allow framework modifications
}

// DefaultConfig returns the default alfa configuration
func DefaultConfig() AlfaConfig {
	return AlfaConfig{
		Workbench: WorkbenchConfig{
			Path:    "workbench",
			Project: "",
		},
		AI: AIConfig{
			Provider:      "anthropic",
			SelectedModel: "", // Uses provider's default_model
			ConfigFile:    "", // Will use workbench/config/ai-config.json by default
			Providers: map[string]ProviderConfig{
				"anthropic": {
					DefaultModel: "claude-3-5-sonnet-20241022",
					Models: map[string]ModelConfig{
						"claude-3-5-sonnet-20241022": {
							MaxTokens:   4096,
							Temperature: 1.0,
							Timeout:     60 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "Most intelligent model, balanced performance",
						},
						"claude-3-opus-20240229": {
							MaxTokens:   4096,
							Temperature: 1.0,
							Timeout:     60 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "Powerful model for complex tasks",
						},
						"claude-3-sonnet-20240229": {
							MaxTokens:   4096,
							Temperature: 1.0,
							Timeout:     60 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "Balanced model for most tasks",
						},
					},
				},
				"openai": {
					DefaultModel: "gpt-5",
					Models: map[string]ModelConfig{
						"gpt-5": {
							MaxTokens:   128000,
							Temperature: 0, // Not used - GPT-5 doesn't support temperature
							Timeout:     180 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "GPT-5: Unified system with smart routing, 272K input, 94.6% AIME",
						},
						"gpt-5-mini": {
							MaxTokens:   128000,
							Temperature: 0, // Not used - GPT-5 doesn't support temperature
							Timeout:     120 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "GPT-5 Mini: Fast variant, half the cost of GPT-4o",
						},
						"gpt-5-nano": {
							MaxTokens:   128000,
							Temperature: 0, // Not used - GPT-5 doesn't support temperature
							Timeout:     120 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "GPT-5 Nano: Smallest GPT-5 variant",
						},
						"gpt-5-chat": {
							MaxTokens:   128000,
							Temperature: 0, // Not used - GPT-5 doesn't support temperature
							Timeout:     120 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "GPT-5 Chat: Optimized for chat applications",
						},
						"gpt-4o": {
							MaxTokens:   4096,
							Temperature: 1.0,
							Timeout:     60 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "GPT-4o: Previous generation multimodal flagship",
						},
						"gpt-4o-mini": {
							MaxTokens:   16384,
							Temperature: 1.0,
							Timeout:     60 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "GPT-4o Mini: Previous generation small model",
						},
						"o1": {
							MaxTokens:   100000,
							Temperature: 0, // Not used - o1 doesn't support temperature
							Timeout:     180 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "o1: Legacy reasoning model",
						},
						"o1-preview": {
							MaxTokens:   32768,
							Temperature: 0, // Not used - o1 doesn't support temperature
							Timeout:     120 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "o1 Preview: Legacy reasoning model (preview)",
						},
						"o1-mini": {
							MaxTokens:   65536,
							Temperature: 0, // Not used - o1 doesn't support temperature
							Timeout:     120 * time.Second,
							RetryCount:  3,
							RetryDelay:  1 * time.Second,
							Description: "o1 Mini: Legacy reasoning for code/math/science",
						},
					},
				},
			},
		},
		Voice: VoiceConfig{
			InputEnabled:  false,
			OutputEnabled: true, // Voice output enabled by default when voice is active
			Headless:      false,
		},
		Execution: ExecutionConfig{
			AutoConfirm:   false,
			MaxIterations: 10,
		},
		Sandbox: SandboxConfig{
			Enabled: false,
			Image:   "golang:1.24-alpine",
		},
		Cellorg: CellorgConfig{
			Enabled:    true,
			ConfigPath: "config",
		},
		Output: OutputConfig{
			CaptureEnabled: true,
			MaxSizeKB:      10,
		},
		SelfModify: SelfModifyConfig{
			Allowed: false,
		},
	}
}

// LoadConfig loads alfa configuration from file
func LoadConfig(path string) (*AlfaConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves alfa configuration to file
func SaveConfig(path string, cfg *AlfaConfig) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// UpdateSetting updates a specific configuration setting
func (c *AlfaConfig) UpdateSetting(key, value string) error {
	switch key {
	// Workbench settings
	case "workbench.path":
		c.Workbench.Path = value
	case "workbench.project":
		c.Workbench.Project = value

	// AI settings
	case "ai.provider":
		if value != "anthropic" && value != "openai" {
			return fmt.Errorf("invalid provider: %s (must be 'anthropic' or 'openai')", value)
		}
		c.AI.Provider = value
	case "ai.config_file":
		c.AI.ConfigFile = value

	// Voice settings
	case "voice.input_enabled":
		c.Voice.InputEnabled = (value == "true")
	case "voice.output_enabled":
		c.Voice.OutputEnabled = (value == "true")
	case "voice.headless":
		c.Voice.Headless = (value == "true")

	// Execution settings
	case "execution.auto_confirm":
		c.Execution.AutoConfirm = (value == "true")
	case "execution.max_iterations":
		var iterations int
		if _, err := fmt.Sscanf(value, "%d", &iterations); err != nil {
			return fmt.Errorf("invalid max_iterations: %s", value)
		}
		c.Execution.MaxIterations = iterations

	// Sandbox settings
	case "sandbox.enabled":
		c.Sandbox.Enabled = (value == "true")
	case "sandbox.image":
		c.Sandbox.Image = value

	// Cellorg settings
	case "cellorg.enabled":
		c.Cellorg.Enabled = (value == "true")
	case "cellorg.config_path":
		c.Cellorg.ConfigPath = value

	// Output settings
	case "output.capture_enabled":
		c.Output.CaptureEnabled = (value == "true")
	case "output.max_size_kb":
		var sizeKB int
		if _, err := fmt.Sscanf(value, "%d", &sizeKB); err != nil {
			return fmt.Errorf("invalid max_size_kb: %s", value)
		}
		c.Output.MaxSizeKB = sizeKB

	// Self-modify settings
	case "self_modify.allowed":
		c.SelfModify.Allowed = (value == "true")

	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

// GetSetting retrieves a specific configuration setting
func (c *AlfaConfig) GetSetting(key string) (string, error) {
	switch key {
	// Workbench settings
	case "workbench.path":
		return c.Workbench.Path, nil
	case "workbench.project":
		return c.Workbench.Project, nil

	// AI settings
	case "ai.provider":
		return c.AI.Provider, nil
	case "ai.config_file":
		return c.AI.ConfigFile, nil

	// Voice settings
	case "voice.input_enabled":
		return fmt.Sprintf("%t", c.Voice.InputEnabled), nil
	case "voice.output_enabled":
		return fmt.Sprintf("%t", c.Voice.OutputEnabled), nil
	case "voice.headless":
		return fmt.Sprintf("%t", c.Voice.Headless), nil

	// Execution settings
	case "execution.auto_confirm":
		return fmt.Sprintf("%t", c.Execution.AutoConfirm), nil
	case "execution.max_iterations":
		return fmt.Sprintf("%d", c.Execution.MaxIterations), nil

	// Sandbox settings
	case "sandbox.enabled":
		return fmt.Sprintf("%t", c.Sandbox.Enabled), nil
	case "sandbox.image":
		return c.Sandbox.Image, nil

	// Cellorg settings
	case "cellorg.enabled":
		return fmt.Sprintf("%t", c.Cellorg.Enabled), nil
	case "cellorg.config_path":
		return c.Cellorg.ConfigPath, nil

	// Output settings
	case "output.capture_enabled":
		return fmt.Sprintf("%t", c.Output.CaptureEnabled), nil
	case "output.max_size_kb":
		return fmt.Sprintf("%d", c.Output.MaxSizeKB), nil

	// Self-modify settings
	case "self_modify.allowed":
		return fmt.Sprintf("%t", c.SelfModify.Allowed), nil

	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// ListSettings returns all configuration settings as key-value pairs
func (c *AlfaConfig) ListSettings() map[string]string {
	return map[string]string{
		"workbench.path":           c.Workbench.Path,
		"workbench.project":        c.Workbench.Project,
		"ai.provider":              c.AI.Provider,
		"ai.config_file":           c.AI.ConfigFile,
		"voice.input_enabled":      fmt.Sprintf("%t", c.Voice.InputEnabled),
		"voice.output_enabled":     fmt.Sprintf("%t", c.Voice.OutputEnabled),
		"voice.headless":           fmt.Sprintf("%t", c.Voice.Headless),
		"execution.auto_confirm":   fmt.Sprintf("%t", c.Execution.AutoConfirm),
		"execution.max_iterations": fmt.Sprintf("%d", c.Execution.MaxIterations),
		"sandbox.enabled":          fmt.Sprintf("%t", c.Sandbox.Enabled),
		"sandbox.image":            c.Sandbox.Image,
		"cellorg.enabled":          fmt.Sprintf("%t", c.Cellorg.Enabled),
		"cellorg.config_path":      c.Cellorg.ConfigPath,
		"output.capture_enabled":   fmt.Sprintf("%t", c.Output.CaptureEnabled),
		"output.max_size_kb":       fmt.Sprintf("%d", c.Output.MaxSizeKB),
		"self_modify.allowed":      fmt.Sprintf("%t", c.SelfModify.Allowed),
	}
}

// ResolveConfigPath resolves the path to alfa.yaml
// Priority: workbench/config/alfa.yaml
func ResolveConfigPath(workbenchPath string) string {
	return filepath.Join(workbenchPath, "config", "alfa.yaml")
}

// LoadOrCreate loads alfa.yaml or creates it with defaults
func LoadOrCreate(workbenchPath string) (*AlfaConfig, error) {
	configPath := ResolveConfigPath(workbenchPath)

	// Try to load existing config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		// Check if file doesn't exist using errors.Is for wrapped errors
		if errors.Is(err, os.ErrNotExist) {
			// Create default config
			cfg := DefaultConfig()
			cfg.Workbench.Path = workbenchPath

			// Save it
			if err := SaveConfig(configPath, &cfg); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}

			fmt.Printf("✨ Created default configuration: %s\n", configPath)
			return &cfg, nil
		}
		return nil, err
	}

	return cfg, nil
}

// ApplyCLIOverrides applies CLI arguments to configuration
func (c *AlfaConfig) ApplyCLIOverrides(flags CLIFlags) {
	// Workbench overrides
	if flags.Workbench != "" {
		c.Workbench.Path = flags.Workbench
	}
	if flags.Project != "" {
		c.Workbench.Project = flags.Project
	}

	// AI overrides
	if flags.Provider != "" {
		c.AI.Provider = flags.Provider
	}
	if flags.ConfigFile != "" {
		c.AI.ConfigFile = flags.ConfigFile
	}

	// Voice overrides
	if flags.Voice {
		c.Voice.InputEnabled = true
		c.Voice.OutputEnabled = true
	}
	if flags.Headless {
		c.Voice.Headless = true
		c.Voice.InputEnabled = true
		c.Voice.OutputEnabled = true
		c.Execution.AutoConfirm = true
	}

	// Execution overrides
	if flags.AutoConfirm {
		c.Execution.AutoConfirm = true
	}
	if flags.MaxIterations > 0 {
		c.Execution.MaxIterations = flags.MaxIterations
	}

	// Sandbox overrides
	if flags.Sandbox {
		c.Sandbox.Enabled = true
	}
	if flags.SandboxImage != "" {
		c.Sandbox.Image = flags.SandboxImage
	}

	// Cellorg overrides
	if flags.EnableCellorg {
		c.Cellorg.Enabled = true
	}
	if flags.CellorgConfig != "" {
		c.Cellorg.ConfigPath = flags.CellorgConfig
	}

	// Output overrides
	if flags.CaptureOutput != nil {
		c.Output.CaptureEnabled = *flags.CaptureOutput
	}
	if flags.MaxOutputKB > 0 {
		c.Output.MaxSizeKB = flags.MaxOutputKB
	}

	// Self-modify overrides
	if flags.AllowSelfModify {
		c.SelfModify.Allowed = true
	}
}

// CLIFlags represents command-line flags
type CLIFlags struct {
	Workbench       string
	ConfigFile      string
	AutoConfirm     bool
	Provider        string
	MaxIterations   int
	Voice           bool
	Headless        bool
	Sandbox         bool
	SandboxImage    string
	Project         string
	EnableCellorg   bool
	CellorgConfig   string
	CaptureOutput   *bool
	MaxOutputKB     int
	AllowSelfModify bool
}

// GetActiveModelConfig returns the configuration for the currently selected model
func (c *AlfaConfig) GetActiveModelConfig() (ModelConfig, string, error) {
	providerCfg, ok := c.AI.Providers[c.AI.Provider]
	if !ok {
		return ModelConfig{}, "", fmt.Errorf("provider not found: %s", c.AI.Provider)
	}

	// Determine which model to use
	modelName := c.AI.SelectedModel
	if modelName == "" {
		modelName = providerCfg.DefaultModel
	}

	modelCfg, ok := providerCfg.Models[modelName]
	if !ok {
		return ModelConfig{}, "", fmt.Errorf("model not found: %s for provider %s", modelName, c.AI.Provider)
	}

	return modelCfg, modelName, nil
}

// MarshalYAML implements custom YAML marshaling for ModelConfig to handle time.Duration
func (m ModelConfig) MarshalYAML() (interface{}, error) {
	result := map[string]interface{}{
		"max_tokens":  m.MaxTokens,
		"temperature": m.Temperature,
		"timeout":     m.Timeout.String(),
		"retry_count": m.RetryCount,
		"retry_delay": m.RetryDelay.String(),
	}
	if m.Description != "" {
		result["description"] = m.Description
	}
	return result, nil
}

// UnmarshalYAML implements custom YAML unmarshaling for ModelConfig to handle time.Duration
func (m *ModelConfig) UnmarshalYAML(node *yaml.Node) error {
	type rawModel struct {
		MaxTokens   int     `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     string  `yaml:"timeout"`
		RetryCount  int     `yaml:"retry_count"`
		RetryDelay  string  `yaml:"retry_delay"`
		Description string  `yaml:"description"`
	}

	var raw rawModel
	if err := node.Decode(&raw); err != nil {
		return err
	}

	timeout, err := time.ParseDuration(raw.Timeout)
	if err != nil {
		timeout = 60 * time.Second // Default on parse error
	}

	retryDelay, err := time.ParseDuration(raw.RetryDelay)
	if err != nil {
		retryDelay = 1 * time.Second // Default on parse error
	}

	m.MaxTokens = raw.MaxTokens
	m.Temperature = raw.Temperature
	m.Timeout = timeout
	m.RetryCount = raw.RetryCount
	m.RetryDelay = retryDelay
	m.Description = raw.Description

	return nil
}
