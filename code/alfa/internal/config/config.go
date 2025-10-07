package config

import (
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
	Provider    string                 `yaml:"provider" json:"provider"`         // "anthropic" or "openai"
	ConfigFile  string                 `yaml:"config_file" json:"config_file"`   // Path to ai-config.json (optional)
	Providers   map[string]ProviderConfig `yaml:"providers" json:"providers"`    // Provider-specific configs
}

// ProviderConfig defines provider-specific settings
type ProviderConfig struct {
	Model       string        `yaml:"model" json:"model"`
	MaxTokens   int           `yaml:"max_tokens" json:"max_tokens"`
	Temperature float64       `yaml:"temperature" json:"temperature"`
	Timeout     time.Duration `yaml:"timeout" json:"timeout"`
	RetryCount  int           `yaml:"retry_count" json:"retry_count"`
	RetryDelay  time.Duration `yaml:"retry_delay" json:"retry_delay"`
}

// VoiceConfig defines voice input/output settings
type VoiceConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`     // Enable voice mode
	Headless bool   `yaml:"headless" json:"headless"`   // Headless mode (voice + auto-confirm)
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
			Provider:   "anthropic",
			ConfigFile: "", // Will use workbench/config/ai-config.json by default
			Providers: map[string]ProviderConfig{
				"anthropic": {
					Model:       "claude-3-5-sonnet-20241022",
					MaxTokens:   4096,
					Temperature: 1.0,
					Timeout:     60 * time.Second,
					RetryCount:  3,
					RetryDelay:  1 * time.Second,
				},
				"openai": {
					Model:       "gpt-4",
					MaxTokens:   4096,
					Temperature: 1.0,
					Timeout:     60 * time.Second,
					RetryCount:  3,
					RetryDelay:  1 * time.Second,
				},
			},
		},
		Voice: VoiceConfig{
			Enabled:  false,
			Headless: false,
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
			Enabled:    false,
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
	case "voice.enabled":
		c.Voice.Enabled = (value == "true")
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
	case "voice.enabled":
		return fmt.Sprintf("%t", c.Voice.Enabled), nil
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
		"voice.enabled":            fmt.Sprintf("%t", c.Voice.Enabled),
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
		if os.IsNotExist(err) {
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
		c.Voice.Enabled = true
	}
	if flags.Headless {
		c.Voice.Headless = true
		c.Voice.Enabled = true
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

// MarshalYAML implements custom YAML marshaling to handle time.Duration
func (c *AlfaConfig) MarshalYAML() (interface{}, error) {
	type Alias AlfaConfig

	// Convert to intermediate structure with duration as strings
	type DurationConfig struct {
		Model       string `yaml:"model"`
		MaxTokens   int    `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     string `yaml:"timeout"`
		RetryCount  int    `yaml:"retry_count"`
		RetryDelay  string `yaml:"retry_delay"`
	}

	providers := make(map[string]DurationConfig)
	for name, p := range c.AI.Providers {
		providers[name] = DurationConfig{
			Model:       p.Model,
			MaxTokens:   p.MaxTokens,
			Temperature: p.Temperature,
			Timeout:     p.Timeout.String(),
			RetryCount:  p.RetryCount,
			RetryDelay:  p.RetryDelay.String(),
		}
	}

	return &struct {
		Workbench  WorkbenchConfig `yaml:"workbench"`
		AI         struct {
			Provider   string                    `yaml:"provider"`
			ConfigFile string                    `yaml:"config_file"`
			Providers  map[string]DurationConfig `yaml:"providers"`
		} `yaml:"ai"`
		Voice      VoiceConfig      `yaml:"voice"`
		Execution  ExecutionConfig  `yaml:"execution"`
		Sandbox    SandboxConfig    `yaml:"sandbox"`
		Cellorg    CellorgConfig    `yaml:"cellorg"`
		Output     OutputConfig     `yaml:"output"`
		SelfModify SelfModifyConfig `yaml:"self_modify"`
	}{
		Workbench: c.Workbench,
		AI: struct {
			Provider   string                    `yaml:"provider"`
			ConfigFile string                    `yaml:"config_file"`
			Providers  map[string]DurationConfig `yaml:"providers"`
		}{
			Provider:   c.AI.Provider,
			ConfigFile: c.AI.ConfigFile,
			Providers:  providers,
		},
		Voice:      c.Voice,
		Execution:  c.Execution,
		Sandbox:    c.Sandbox,
		Cellorg:    c.Cellorg,
		Output:     c.Output,
		SelfModify: c.SelfModify,
	}, nil
}

// UnmarshalYAML implements custom YAML unmarshaling to handle time.Duration
func (c *AlfaConfig) UnmarshalYAML(node *yaml.Node) error {
	type Alias AlfaConfig

	type DurationConfig struct {
		Model       string `yaml:"model"`
		MaxTokens   int    `yaml:"max_tokens"`
		Temperature float64 `yaml:"temperature"`
		Timeout     string `yaml:"timeout"`
		RetryCount  int    `yaml:"retry_count"`
		RetryDelay  string `yaml:"retry_delay"`
	}

	aux := &struct {
		Workbench  WorkbenchConfig `yaml:"workbench"`
		AI         struct {
			Provider   string                    `yaml:"provider"`
			ConfigFile string                    `yaml:"config_file"`
			Providers  map[string]DurationConfig `yaml:"providers"`
		} `yaml:"ai"`
		Voice      VoiceConfig      `yaml:"voice"`
		Execution  ExecutionConfig  `yaml:"execution"`
		Sandbox    SandboxConfig    `yaml:"sandbox"`
		Cellorg    CellorgConfig    `yaml:"cellorg"`
		Output     OutputConfig     `yaml:"output"`
		SelfModify SelfModifyConfig `yaml:"self_modify"`
	}{}

	if err := node.Decode(aux); err != nil {
		return err
	}

	c.Workbench = aux.Workbench
	c.AI.Provider = aux.AI.Provider
	c.AI.ConfigFile = aux.AI.ConfigFile
	c.Voice = aux.Voice
	c.Execution = aux.Execution
	c.Sandbox = aux.Sandbox
	c.Cellorg = aux.Cellorg
	c.Output = aux.Output
	c.SelfModify = aux.SelfModify

	// Convert duration strings to time.Duration
	c.AI.Providers = make(map[string]ProviderConfig)
	for name, p := range aux.AI.Providers {
		timeout, _ := time.ParseDuration(p.Timeout)
		retryDelay, _ := time.ParseDuration(p.RetryDelay)

		c.AI.Providers[name] = ProviderConfig{
			Model:       p.Model,
			MaxTokens:   p.MaxTokens,
			Temperature: p.Temperature,
			Timeout:     timeout,
			RetryCount:  p.RetryCount,
			RetryDelay:  retryDelay,
		}
	}

	return nil
}
