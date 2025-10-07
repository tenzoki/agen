# Alfa Configuration

Alfa uses a hierarchical configuration system with `alfa.yaml` as the central configuration file.

## Configuration Priority

**CLI arguments always have priority over alfa.yaml settings.**

1. **CLI arguments** (highest priority)
2. **alfa.yaml** (workbench/config/alfa.yaml)
3. **Embedded defaults** (lowest priority)

## Configuration File Location

`workbench/config/alfa.yaml`

- The workbench path is determined by `--workbench` flag or defaults to `workbench` in the current directory
- If `alfa.yaml` doesn't exist, alfa creates it with default settings
- Changes are saved back to `alfa.yaml` on every run

## Configuration Structure

### Workbench
```yaml
workbench:
  path: workbench           # Workbench directory
  project: ""               # Default project (empty = prompt/context)
```

### AI Provider
```yaml
ai:
  provider: anthropic       # Default provider
  config_file: ""           # Path to ai-config.json (optional)
  providers:
    anthropic:
      model: claude-3-5-sonnet-20241022
      max_tokens: 4096
      temperature: 1.0
      timeout: 1m0s
      retry_count: 3
      retry_delay: 1s
    openai:
      model: gpt-4
      max_tokens: 4096
      temperature: 1.0
      timeout: 1m0s
      retry_count: 3
      retry_delay: 1s
```

**API Keys**: Set via environment variables:
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`

### Voice
```yaml
voice:
  enabled: false            # Enable voice input/output
  headless: false           # Headless mode (voice + auto-confirm)
```

**Requirements**: OPENAI_API_KEY and sox

### Execution
```yaml
execution:
  auto_confirm: false       # Auto-confirm operations
  max_iterations: 10        # Max AI iterations
```

### Sandbox
```yaml
sandbox:
  enabled: false            # Use Docker sandbox
  image: golang:1.24-alpine # Docker image
```

**Requires**: Docker

### Cellorg
```yaml
cellorg:
  enabled: false            # Enable cellorg features
  config_path: config       # Cellorg config directory
```

### Output
```yaml
output:
  capture_enabled: true     # Capture command output
  max_size_kb: 10           # Max output size to show AI
```

### Self-Modification
```yaml
self_modify:
  allowed: false            # Allow framework modifications
```

## CLI Arguments

All settings can be overridden via CLI:

| Setting | Flag | Example |
|---------|------|---------|
| Workbench | `--workbench` | `--workbench=/path/to/wb` |
| Project | `--project` | `--project=my-app` |
| Provider | `--provider` | `--provider=openai` |
| AI Config | `--config` | `--config=custom-ai.json` |
| Voice | `--voice` | `--voice` |
| Headless | `--headless` | `--headless` |
| Auto-confirm | `--auto-confirm` | `--auto-confirm` |
| Max Iterations | `--max-iterations` | `--max-iterations=20` |
| Sandbox | `--sandbox` | `--sandbox` |
| Sandbox Image | `--sandbox-image` | `--sandbox-image=python:3.11` |
| Cellorg | `--enable-cellorg` | `--enable-cellorg` |
| Cellorg Config | `--cellorg-config` | `--cellorg-config=/path` |
| Capture Output | `--capture-output` | `--capture-output=false` |
| Max Output | `--max-output` | `--max-output=50` |
| Self-Modify | `--allow-self-modify` | `--allow-self-modify` |

## Usage Examples

### First Run
```bash
# Creates alfa.yaml with defaults
alfa
```

### Use Saved Settings
```bash
# Uses alfa.yaml configuration
alfa
```

### Override Settings
```bash
# Override provider (saved to alfa.yaml)
alfa --provider openai

# Enable voice + auto-confirm
alfa --voice --auto-confirm

# Full configuration
alfa --project=my-app \
     --provider=anthropic \
     --enable-cellorg \
     --allow-self-modify
```

### Project Management
```bash
# List projects
alfa --list-projects

# Create project
alfa --create-project my-app

# Use project
alfa --project my-app

# Delete project (keeps backup)
alfa --delete-project my-app

# Restore project
alfa --restore-project my-app
```

## Configuration Flow

1. **Determine workbench path**: CLI `--workbench` or default `workbench`
2. **Load alfa.yaml**: From `workbench/config/alfa.yaml`
3. **Create if missing**: Uses embedded defaults
4. **Apply CLI overrides**: CLI arguments override alfa.yaml
5. **Save configuration**: Final config saved back to alfa.yaml
6. **Run alfa**: Uses merged configuration

## Benefits

- **Zero configuration**: Works out of the box with defaults
- **Persistent settings**: Save preferences without repeating CLI args
- **Flexible overrides**: CLI args for one-off changes
- **Self-documenting**: alfa.yaml shows all current settings
- **Version control friendly**: Check alfa.yaml into git for team settings

## Related Files

- `workbench/config/alfa.yaml` - Main alfa configuration
- `workbench/config/ai-config.json` - AI provider details (optional)
- `workbench/config/speech-config.json` - Speech settings (legacy)
- `workbench/.alfa/context.json` - Active project context

## See Also

- [Configuration Standards](../../guidelines/references/config-standards.md)
- [Alfa README](../../code/alfa/README.md)
