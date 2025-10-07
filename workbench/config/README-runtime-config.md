# Runtime Configuration Management

Alfa supports changing configuration settings during runtime without restarting the session.

## Available Tools

### 1. config_list
Lists all current configuration settings grouped by category.

**Usage:**
```
User: "Show me all configuration settings"
User: "What are my current settings?"
User: "List the configuration"
```

**Example Output:**
```
[Workbench]
  workbench.path              = workbench
  workbench.project           = my-app

[AI Provider]
  ai.provider                 = anthropic
  ai.config_file              =

[Voice]
  voice.enabled               = false
  voice.headless              = false

[Execution]
  execution.auto_confirm      = false
  execution.max_iterations    = 10

[Sandbox]
  sandbox.enabled             = false
  sandbox.image               = golang:1.24-alpine

[Cellorg]
  cellorg.enabled             = false
  cellorg.config_path         = config

[Output]
  output.capture_enabled      = true
  output.max_size_kb          = 10

[Self-Modify]
  self_modify.allowed         = false

Config file: workbench/config/alfa.yaml
```

### 2. config_get
Retrieves a specific configuration setting.

**Parameters:**
- `key` (required): The configuration key to retrieve

**Usage:**
```
User: "What is the max iterations setting?"
User: "Get the ai.provider setting"
User: "Show me the output.max_size_kb value"
```

**Tool call:**
```json
{
  "type": "config_get",
  "params": {
    "key": "execution.max_iterations"
  }
}
```

**Response:**
```json
{
  "key": "execution.max_iterations",
  "value": "10"
}
```

### 3. config_set
Updates a configuration setting and saves it to alfa.yaml.

**Parameters:**
- `key` (required): The configuration key to update
- `value` (required): The new value (as string)

**Usage:**
```
User: "Change max iterations to 20"
User: "Set the provider to openai"
User: "Enable auto-confirm mode"
User: "Increase max output size to 50KB"
```

**Tool call:**
```json
{
  "type": "config_set",
  "params": {
    "key": "execution.max_iterations",
    "value": "20"
  }
}
```

**Response:**
```json
{
  "key": "execution.max_iterations",
  "value": "20",
  "needs_restart": false
}
```

## Configuration Keys

### Workbench
- `workbench.path` - Workbench directory path
- `workbench.project` - Default project name

### AI Provider
- `ai.provider` - AI provider (anthropic or openai) *[requires restart]*
- `ai.config_file` - Path to ai-config.json *[requires restart]*

### Voice
- `voice.enabled` - Enable voice input/output (true/false)
- `voice.headless` - Headless mode (true/false)

### Execution
- `execution.auto_confirm` - Auto-confirm operations (true/false)
- `execution.max_iterations` - Maximum AI iterations (number)

### Sandbox
- `sandbox.enabled` - Use Docker sandbox (true/false) *[requires restart]*
- `sandbox.image` - Docker image name *[requires restart]*

### Cellorg
- `cellorg.enabled` - Enable cellorg features (true/false) *[requires restart]*
- `cellorg.config_path` - Cellorg config directory *[requires restart]*

### Output
- `output.capture_enabled` - Capture command output (true/false) **[immediate]**
- `output.max_size_kb` - Max output size in KB (number) **[immediate]**

### Self-Modify
- `self_modify.allowed` - Allow framework modifications (true/false)

## Settings Behavior

### Immediate Effect ⚡
These settings apply immediately without restart:
- `output.capture_enabled`
- `output.max_size_kb`

### Next Operation
These settings apply on the next relevant operation:
- `execution.auto_confirm`
- `execution.max_iterations`
- `self_modify.allowed`

### Requires Restart 🔄
These settings require restarting alfa to take effect:
- `ai.provider`
- `ai.config_file`
- `workbench.path`
- `sandbox.enabled`
- `sandbox.image`
- `cellorg.enabled`
- `cellorg.config_path`

When you change these, alfa will indicate: "(restart alfa to apply changes)"

## Example Conversations

### Example 1: Change Max Iterations
```
User: I need more iterations, change it to 25
AI: I'll update the max iterations setting to 25.
    [executes config_set with key="execution.max_iterations", value="25"]
    Updated setting: execution.max_iterations = 25 (saved to alfa.yaml)
    The change will apply on your next request.

User: Great, now help me with...
```

### Example 2: Enable Auto-Confirm
```
User: Turn on auto-confirm mode
AI: I'll enable auto-confirm for you.
    [executes config_set with key="execution.auto_confirm", value="true"]
    Updated setting: execution.auto_confirm = true (saved to alfa.yaml)
    Auto-confirm will apply on the next operation.
```

### Example 3: Increase Output Size
```
User: I need to see more output, increase to 100KB
AI: I'll increase the max output size to 100KB.
    [executes config_set with key="output.max_size_kb", value="100"]
    Updated setting: output.max_size_kb = 100 (saved to alfa.yaml) (applied immediately)
    Output capture is now set to 100KB.
```

### Example 4: View All Settings
```
User: Show me all my configuration
AI: Here are your current configuration settings:
    [executes config_list]

    [Workbench]
      workbench.path              = workbench
      workbench.project           = my-app

    [AI Provider]
      ai.provider                 = anthropic
      ai.config_file              =

    [Execution]
      execution.auto_confirm      = true
      execution.max_iterations    = 25

    ... (full list shown)
```

### Example 5: Change Provider (Requires Restart)
```
User: Switch to OpenAI
AI: I'll change the AI provider to OpenAI.
    [executes config_set with key="ai.provider", value="openai"]
    Updated setting: ai.provider = openai (saved to alfa.yaml) (restart alfa to apply changes)

    Note: You'll need to restart alfa for this change to take effect.
    Make sure you have OPENAI_API_KEY set in your environment.
```

## Persistence

All configuration changes are:
1. **Applied to the running config** (in memory)
2. **Saved to alfa.yaml** (on disk)
3. **Preserved across sessions** (will be loaded on next alfa start)

The configuration file is located at: `workbench/config/alfa.yaml`

## Error Handling

### Invalid Key
```
User: Set foo.bar to value
AI: [executes config_set with key="foo.bar", value="value"]
    Failed: unknown configuration key: foo.bar
```

### Invalid Value
```
User: Set max iterations to abc
AI: [executes config_set with key="execution.max_iterations", value="abc"]
    Failed: invalid max_iterations: abc
```

### Invalid Provider
```
User: Set provider to gemini
AI: [executes config_set with key="ai.provider", value="gemini"]
    Failed: invalid provider: gemini (must be 'anthropic' or 'openai')
```

## Integration with AI

The AI assistant can automatically:
- **Interpret natural language**: "turn on auto-confirm" → config_set
- **Retrieve settings**: "what's my max iterations?" → config_get
- **Show all settings**: "show me my config" → config_list
- **Make changes**: "increase output to 50kb" → config_set
- **Warn about restarts**: "Note: restart required for this change"

## Technical Notes

### Value Types
All values are stored and retrieved as strings. The config system handles type conversion:
- **Boolean**: "true" or "false"
- **Integer**: "10", "25", "100"
- **String**: Direct values

### Validation
The config system validates:
- Key names (must be recognized)
- Boolean values (must be "true" or "false")
- Integer values (must be parseable numbers)
- Provider values (must be "anthropic" or "openai")

### Thread Safety
Configuration updates are:
- Applied atomically to the in-memory config
- Immediately saved to disk
- Applied to relevant subsystems (like output capture)

## See Also

- [Alfa Configuration Guide](README-alfa-config.md)
- [Configuration Example](alfa.yaml.example)
- [Tool Dispatcher](../../code/alfa/internal/tools/tools.go)