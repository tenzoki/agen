package agent

import (
	"os"
	"path/filepath"
)

// StandardConfigResolver follows AGEN's universal config convention
//
// Resolution order (highest priority first):
// 1. Command-line flag (--config=/path/to/file)
// 2. Environment variable AGEN_CONFIG_PATH
// 3. Environment variable AGEN_WORKBENCH_DIR/config/agents/<name>.yaml
// 4. CWD-relative: ./config/<name>.yaml (most natural for standalone)
// 5. CWD-relative AGEN convention: ./workbench/config/agents/<name>.yaml
// 6. Binary-relative: <binary-dir>/config/<name>.yaml (portable bundles)
// 7. No config found (returns empty string, use embedded defaults)
type StandardConfigResolver struct {
	AgentName  string
	ConfigFlag *string // Optional: pointer to flag.String() result
}

// Resolve returns the config file path following AGEN conventions
// Returns empty string if no config file found (caller should use embedded defaults)
func (r *StandardConfigResolver) Resolve() (string, error) {
	// 1. Command-line flag (explicit override)
	if r.ConfigFlag != nil && *r.ConfigFlag != "" {
		return *r.ConfigFlag, nil
	}

	// 2. Explicit environment variable (specific file)
	if path := os.Getenv("AGEN_CONFIG_PATH"); path != "" {
		if fileExists(path) {
			return path, nil
		}
	}

	// 3. Workbench environment variable (orchestrator sets this when spawning)
	if workbench := os.Getenv("AGEN_WORKBENCH_DIR"); workbench != "" {
		path := filepath.Join(workbench, "config", "agents", r.AgentName+".yaml")
		if fileExists(path) {
			return path, nil
		}
	}

	// 4. CWD-relative simple pattern (most natural for standalone)
	path := filepath.Join("config", r.AgentName+".yaml")
	if fileExists(path) {
		return path, nil
	}

	// 5. CWD-relative AGEN convention (for AGEN development)
	path = filepath.Join("workbench", "config", "agents", r.AgentName+".yaml")
	if fileExists(path) {
		return path, nil
	}

	// 6. Binary-relative (for portable agent bundles)
	binaryDir := filepath.Dir(os.Args[0])
	path = filepath.Join(binaryDir, "config", r.AgentName+".yaml")
	if fileExists(path) {
		return path, nil
	}

	// 7. No config file found - caller should use embedded defaults
	return "", nil
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// LoadConfigWithDefaults loads config using standard resolution, falling back to provided defaults
//
// Example usage:
//   type MyConfig struct { ... }
//   defaults := MyConfig{Port: 8080, Debug: false}
//
//   config, err := agent.LoadConfigWithDefaults("my_agent", &configFlag, defaults,
//       func(path string) (MyConfig, error) {
//           var cfg MyConfig
//           data, _ := os.ReadFile(path)
//           yaml.Unmarshal(data, &cfg)
//           return cfg, nil
//       })
func LoadConfigWithDefaults[T any](
	agentName string,
	configFlag *string,
	defaults T,
	loader func(string) (T, error),
) (T, error) {
	resolver := StandardConfigResolver{
		AgentName:  agentName,
		ConfigFlag: configFlag,
	}

	path, err := resolver.Resolve()
	if err != nil {
		return defaults, err
	}

	// No config file found, use defaults
	if path == "" {
		return defaults, nil
	}

	// Load from file
	return loader(path)
}

// GetConfigPath is a convenience function that just returns the resolved path
// Useful when you want to know where config came from
func GetConfigPath(agentName string, configFlag *string) string {
	resolver := StandardConfigResolver{
		AgentName:  agentName,
		ConfigFlag: configFlag,
	}
	path, _ := resolver.Resolve()
	return path
}
