# AGEN Configuration Conventions

This document defines the standard configuration patterns that ALL AGEN binaries must follow.

## Principles

1. **Zero-configuration startup**: All binaries work with embedded defaults
2. **Workbench-centric**: Configuration lives in `workbench/config/`
3. **Explicit overrides**: CLI flags and environment variables take precedence
4. **Consistent resolution**: Same lookup order for all binaries

## Standard Configuration Resolution

All AGEN binaries follow this **7-level resolution order** (highest priority first):

### 1. Command-line Flag
```bash
bin/agent --config=/absolute/path/to/config.yaml
```
- **Priority**: Highest
- **Use case**: One-off overrides, testing
- **Format**: Absolute or relative path

### 2. Environment Variable
```bash
export AGEN_CONFIG_PATH=/path/to/config.yaml
bin/agent
```
- **Priority**: High
- **Use case**: CI/CD, containerized deployments
- **Variable**: `AGEN_CONFIG_PATH`

### 3. Orchestrator Workbench
```bash
export AGEN_WORKBENCH_DIR=/path/to/workbench
bin/agent
```
- **Priority**: Medium-high
- **Path**: `$AGEN_WORKBENCH_DIR/config/agents/<agent-name>.yaml`
- **Use case**: Agents spawned by orchestrator
- **Variable**: `AGEN_WORKBENCH_DIR`

### 4. CWD-relative Simple
```bash
./bin/agent
# Looks for: ./config/<agent-name>.yaml
```
- **Priority**: Medium
- **Path**: `./config/<agent-name>.yaml`
- **Use case**: Project-local configuration
- **Most natural for manual invocation**

### 5. CWD-relative AGEN Convention
```bash
./bin/agent
# Looks for: ./workbench/config/agents/<agent-name>.yaml
```
- **Priority**: Medium-low
- **Path**: `./workbench/config/agents/<agent-name>.yaml`
- **Use case**: Running from AGEN root

### 6. Binary-relative
```bash
/opt/agen/bin/agent
# Looks for: /opt/agen/bin/config/<agent-name>.yaml
```
- **Priority**: Low
- **Path**: `<binary-dir>/config/<agent-name>.yaml`
- **Use case**: Installed binaries with adjacent config

### 7. Embedded Defaults
- **Priority**: Lowest (fallback)
- **Always available**: Zero-configuration deployment
- **Use case**: First run, containers, testing

## Directory Structure

### Workbench Layout
```
workbench/
├── config/                    # Configuration root
│   ├── cellorg.yaml          # Main orchestrator config
│   ├── pool.yaml             # Agent type definitions
│   ├── ai-config.json        # AI provider settings (alfa)
│   ├── speech-config.json    # Speech settings (alfa)
│   ├── cells/                # Cell definitions (organized by purpose)
│   │   ├── pipelines/        # Processing workflows (8 cells)
│   │   ├── services/         # Backend services (6 cells)
│   │   ├── analysis/         # Content analysis (6 cells)
│   │   └── synthesis/        # Output generation (5 cells)
│   └── agents/               # Agent configs (OPTIONAL)
│       ├── ner_agent.yaml
│       └── *.yaml
├── projects/                 # User projects (VFS isolated)
│   └── <name>/
└── demos/                    # Demo applications
```

### Framework Layout
```
agen/
├── code/                     # Source code
├── workbench/               # Default workbench
├── bin/                     # Built binaries
└── reflect/                 # Documentation
```

## Environment Variables

### Standard Variables
- **`AGEN_CONFIG_PATH`**: Explicit config file path (highest priority after CLI)
- **`AGEN_WORKBENCH_DIR`**: Workbench directory (set by orchestrator for spawned agents)
- **`AGEN_FRAMEWORK_DIR`**: Framework root directory
- **`AGEN_SELF_MODIFICATION`**: Enable/disable self-modification (`true`/`false`)

### Provider-specific Variables
- **`ANTHROPIC_API_KEY`**: Anthropic API key (alfa)
- **`OPENAI_API_KEY`**: OpenAI API key (alfa)

## Binary-Specific Configurations

### Orchestrator (`bin/orchestrator`)
```bash
# Default workbench
bin/orchestrator

# Custom config
bin/orchestrator -config=workbench/config/cellorg.yaml

# Specific cell
bin/orchestrator -config=workbench/config/cells/pipelines/anonymization-pipeline.yaml

# Dry run (validate)
bin/orchestrator -config=workbench/config/cellorg.yaml -dry-run
```

**Config file**: `workbench/config/cellorg.yaml`

**Required sections**:
- `support`: Support service settings
- `broker`: Message routing settings
- `pool`: References to pool.yaml
- `cells`: References to cell definitions
- `self_modification`: Self-modification settings

### Alfa (`bin/alfa`)
```bash
# Default workbench
bin/alfa

# Custom workbench
bin/alfa --workbench=/path/to/workbench

# With project
bin/alfa --project=my-app

# Custom AI config
bin/alfa --config=/path/to/ai-config.json
```

**Config files**:
- AI: `workbench/config/ai-config.json`
- Speech: `workbench/config/speech-config.json`

**Resolution**:
- AI config: `--config` flag → `$workbench/config/ai-config.json` → embedded defaults
- Speech config: `$workbench/config/speech-config.json` → embedded defaults

### Agents (`bin/*_agent`)
```bash
# Standalone (embedded defaults)
bin/ner_agent

# With custom config
bin/ner_agent --config=my-config.yaml

# Spawned by orchestrator (uses AGEN_WORKBENCH_DIR)
AGEN_WORKBENCH_DIR=/path/to/workbench bin/ner_agent
```

**Standard agent config resolution**:
Uses `StandardConfigResolver` from `cellorg/public/agent/config.go`

**Implementation requirement**:
```go
resolver := agent.NewStandardConfigResolver("agent-name", configFlag)
configPath, err := resolver.Resolve()
```

## Implementation Checklist

All binaries MUST:

- [ ] Support `--config` or `-config` flag
- [ ] Implement standard 7-level resolution order
- [ ] Have embedded defaults (work without any config)
- [ ] Respect `AGEN_CONFIG_PATH` environment variable
- [ ] Respect `AGEN_WORKBENCH_DIR` when spawned by orchestrator
- [ ] Document config file format in binary's README
- [ ] Handle missing config gracefully (use defaults)
- [ ] Validate config and provide clear error messages

## Testing Configuration Resolution

### Test Cases
1. **No config**: Binary runs with embedded defaults
2. **CLI flag**: `--config=path` overrides everything
3. **Environment**: `AGEN_CONFIG_PATH=path` works
4. **Workbench**: Config found in `$AGEN_WORKBENCH_DIR/config/agents/`
5. **CWD simple**: Config found in `./config/`
6. **CWD AGEN**: Config found in `./workbench/config/agents/`
7. **Binary-relative**: Config found adjacent to binary

### Validation Script
```bash
# Test embedded defaults
bin/agent

# Test CLI override
bin/agent --config=test.yaml

# Test environment
export AGEN_CONFIG_PATH=test.yaml
bin/agent
unset AGEN_CONFIG_PATH

# Test workbench
export AGEN_WORKBENCH_DIR=workbench
bin/agent
unset AGEN_WORKBENCH_DIR
```

## Migration Guide

### For New Binaries
1. Use `agent.NewStandardConfigResolver()` from cellorg
2. Implement embedded defaults
3. Accept `--config` flag
4. Test all 7 resolution levels

### For Existing Binaries
1. Add StandardConfigResolver integration
2. Remove hardcoded paths
3. Move configs to `workbench/config/agents/`
4. Update documentation

## See Also

- [workbench/config/README.md](../workbench/config/README.md) - Configuration file formats
- [immutable-principles.md](./immutable-principles.md) - Core AGEN principles
- [cellorg/public/agent/config.go](../code/cellorg/public/agent/config.go) - StandardConfigResolver implementation
