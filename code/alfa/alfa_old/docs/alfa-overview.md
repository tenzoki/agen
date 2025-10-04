# Alfa - Technical Overview for AI Integration

## What is Alfa?

Alfa is an AI-powered coding assistant that combines voice control, large language models (LLMs), and sandboxed code execution to provide an interactive development experience. It acts as a conversational interface between developers and their codebase, understanding natural language requests and executing precise code operations.

---

## Core Architecture

### 1. **Orchestrator** (`internal/orchestrator`)
The main control loop that coordinates all components:
- Manages conversation flow between user and AI
- Parses AI responses for executable actions
- Executes actions through the appropriate subsystems
- Handles confirmation mode vs. autonomous mode

### 2. **AI Layer** (`internal/ai`)
LLM integration supporting multiple providers:
- **Anthropic Claude** (claude-3-5-sonnet, etc.)
- **OpenAI GPT-4**
- Configured via `config/ai-config.json`
- API keys loaded from environment variables

### 3. **Multi-Project Management** (`internal/project`)
Manages multiple isolated projects within a single workbench:
- Each project is a separate git repository
- Projects stored in `workbench/projects/`
- Automatic backup to `workbench/.git-remotes/` (bare repos)
- Hot-swapping between projects without restart
- Full CRUD operations: create, list, delete, restore, switch

### 4. **Virtual File System (VFS)** (`internal/vfs`)
Dual VFS architecture for security:
- **Workbench VFS**: Configuration, context, and history
  - Location: `workbench/` root
  - Stores: `config/`, `.alfa/context.json`, `.alfa/history.log`
- **Project VFS**: Isolated per-project file access
  - Location: `workbench/projects/<project-name>/`
  - AI can only access files within active project
  - Path validation prevents directory traversal attacks

### 5. **Tool Dispatcher** (`internal/tools`)
Executes AI-requested operations with VFS isolation:

**Code Operations:**
- `read_file`: Read file contents (VFS-scoped)
- `write_file`: Create or overwrite files (VFS-scoped)
- `patch`: Apply structured code changes using TextPatch format
- `search`: Search codebase for patterns (VFS-scoped)

**Execution Operations:**
- `run_command`: Execute shell commands (optional Docker sandbox)
- `run_tests`: Run test suites (optional Docker sandbox)

**Project Management Operations:**
- `list_projects`: List all projects with metadata
- `create_project`: Create new project + auto-switch
- `delete_project`: Delete project (backup kept)
- `restore_project`: Restore from backup + auto-switch
- `switch_project`: Hot-swap to different project

### 6. **TextPatch System** (`internal/textpatch`)
Structured code modification format:
```json
{
  "action": "patch",
  "file": "path/to/file.go",
  "operations": [
    {"line": 10, "type": "insert", "content": ["new line 1", "new line 2"]},
    {"line": 15, "type": "delete"},
    {"line": 20, "type": "replace", "content": ["updated line"]}
  ]
}
```

### 7. **Version Control (VCR)** (`internal/vcr`)
Git integration wrapper:
- Automatic commits after successful operations
- **Auto-push to local bare repo** after each commit
- Supports branching and history tracking
- Author: "VCR Bot"

### 8. **Docker Sandbox** (`internal/sandbox`)
Optional isolated execution environment:
- CPU limits (default: 1-2 cores)
- Memory limits (default: 512MB-1GB)
- Network isolation (disabled by default)
- Read-only filesystem (except `/tmp`)
- All capabilities dropped, no new privileges

### 9. **Context Manager** (`internal/context`)
Maintains conversation state:
- Tracks active project name
- Stores conversation history (messages)
- Records file modifications
- Persists to `.alfa/context.json`
- Auto-trims to max messages (default: 100)

### 10. **Voice Pipeline** (Optional)
- **STT**: Whisper (OpenAI)
- **TTS**: OpenAI TTS
- **Audio**: Sox-based recording/playback with VAD
- Hybrid text/voice input supported

---

## Request-Response Flow

```
1. User Input (text or voice)
   ↓
2. Orchestrator receives input
   ↓
3. Context Manager adds to conversation history
   ↓
4. Build messages with system prompt + history
   ↓
5. Send to LLM (Claude or GPT-4)
   ↓
6. Parse LLM response for JSON action blocks
   ↓
7. For each action:
   - Confirm with user (if confirm mode)
   - Execute via Tool Dispatcher
   - Handle project switches automatically
   ↓
8. VCR auto-commits + auto-pushes changes
   ↓
9. Context Manager records results
   ↓
10. Display results to user (text or voice)
```

---

## AI-Accessible Actions

The AI receives a system prompt describing all available actions in JSON format:

### File Operations
```json
{"action": "read_file", "path": "main.go"}
{"action": "write_file", "path": "test.txt", "content": "..."}
{"action": "search", "pattern": "func.*Error"}
```

### Code Patching
```json
{
  "action": "patch",
  "file": "main.go",
  "operations": [{"line": 10, "type": "insert", "content": ["// New comment"]}]
}
```

### Execution
```json
{"action": "run_command", "command": "go build"}
{"action": "run_tests", "pattern": "./..."}
```

### Project Management
```json
{"action": "list_projects"}
{"action": "create_project", "name": "myapp"}
{"action": "delete_project", "name": "old-project"}
{"action": "restore_project", "name": "old-project"}
{"action": "switch_project", "name": "backend"}
```

---

## Directory Structure

```
workbench/                      # Workbench root
├── config/
│   └── ai-config.json         # LLM provider configuration
├── .alfa/
│   ├── context.json           # Conversation + active project
│   └── history.log            # Operation log
├── projects/                  # Multi-project workspace
│   ├── myapp/                 # Project 1 (git repo)
│   │   ├── .git/
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── ...
│   ├── backend/               # Project 2 (git repo)
│   └── frontend/              # Project 3 (git repo)
└── .git-remotes/              # Backup bare repos
    ├── myapp.git/
    ├── backend.git/
    └── frontend.git/
```

---

## Configuration

### AI Config (`config/ai-config.json`)
```json
{
  "default_provider": "anthropic",
  "providers": {
    "anthropic": {
      "model": "claude-3-5-sonnet-20241022",
      "max_tokens": 4096,
      "temperature": 1.0,
      "timeout": 60000000000
    },
    "openai": {
      "model": "gpt-4",
      "max_tokens": 4096,
      "temperature": 1.0,
      "timeout": 60000000000
    }
  }
}
```

### Environment Variables
- `ANTHROPIC_API_KEY`: Claude API key
- `OPENAI_API_KEY`: OpenAI API key (also for Whisper/TTS if voice enabled)

---

## Execution Modes

### 1. **Confirm Mode** (default)
- User approves each action before execution
- Safety-first approach
- CLI: `./alfa` or `./alfa --mode confirm`

### 2. **Allow-All Mode**
- Autonomous execution without confirmations
- CLI: `./alfa --mode allow-all`

### 3. **Headless Mode**
- Autonomous voice agent
- Enables voice + allow-all automatically
- CLI: `./alfa --headless`

---

## Security Model

### Path Isolation
- AI can only access files within active project VFS
- `..` traversal attempts are blocked
- Workbench config/context isolated from AI

### Sandboxing
- Optional Docker-based execution
- Resource limits enforced
- Network isolation available
- Read-only filesystem

### Backup & Recovery
- Every commit auto-pushed to local bare repo
- Deleted projects recoverable
- Full git history preserved

---

## Hot Project Switching

Key innovation: **No restart required** when switching projects.

When switching projects (create, restore, or switch actions):
1. New VFS initialized for target project
2. New VCR initialized for target project
3. Tool Dispatcher updated with new VFS
4. Context Manager records active project
5. All subsequent operations use new project automatically

This enables workflows like:
```
User: "Create project backend"
AI: [creates + switches to backend]
User: "Write main.go"
AI: [writes to backend/main.go]
User: "Switch to frontend"
AI: [switches to frontend]
User: "Write index.html"
AI: [writes to frontend/index.html]
```

---

## Natural Language Understanding

The AI understands requests like:
- "Create a new Go project called api-service"
- "Switch to the backend project"
- "Delete the old-project but keep the backup"
- "Restore the deleted test-app"
- "Add error handling to the main function in server.go"
- "Run the tests"
- "Search for all TODO comments"

The system prompt teaches the AI to respond with appropriate JSON actions.

---

## Integration Points for Extensions

### 1. **Adding New Tools/Actions**
Implement in `internal/tools/tools.go`:
```go
func (d *Dispatcher) executeNewAction(action Action) Result {
    // Use d.vfs for file access
    // Use d.projectManager for project operations
    // Use d.sandbox for isolated execution (if enabled)
    // Return Result with success/failure + output
}
```

Add to switch statement in `Execute()` method.

### 2. **Extending System Prompt**
Update `buildSystemPrompt()` in `internal/orchestrator/orchestrator.go`:
- Add new action to AVAILABLE ACTIONS list
- Provide JSON format examples
- Explain behavior and constraints

### 3. **Custom LLM Providers**
Implement `LLM` interface in `internal/ai`:
```go
type LLM interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
    ChatStream(ctx context.Context, messages []Message) (<-chan string, <-chan error)
    Model() string
    Provider() string
}
```

### 4. **VFS Extensions**
The VFS provides:
- `Read(path string) ([]byte, error)`
- `Write(path string, data []byte) error`
- `List(path string) ([]os.FileInfo, error)`
- `Exists(path string) bool`
- `Root() string`

All operations are scoped to VFS root (project directory).

### 5. **Hook into Project Lifecycle**
- Project creation: `internal/project/Manager.Create()`
- Project deletion: `internal/project/Manager.Delete()`
- Project restore: `internal/project/Manager.Restore()`
- Project switch: `internal/orchestrator/Orchestrator.SwitchProject()`

---

## Key Design Principles

1. **VFS Isolation**: AI never accesses files outside active project
2. **Structured Operations**: All actions are JSON-based for parsing reliability
3. **Auto-backup**: Every commit pushed to local bare repo for recovery
4. **Hot-swapping**: No restarts needed when changing projects
5. **Multi-modal**: Text and voice input supported
6. **Provider-agnostic**: Works with Claude, GPT-4, or custom LLMs
7. **Safety layers**: Confirmation mode, sandboxing, path validation

---

## CLI Interface

```bash
# Project management
./alfa --create-project <name>
./alfa --list-projects
./alfa --delete-project <name>
./alfa --restore-project <name>

# Working with projects
./alfa --project <name>              # Select project
./alfa --project <name> --voice      # With voice I/O
./alfa --project <name> --headless   # Autonomous voice agent

# Execution options
./alfa --mode allow-all              # No confirmations
./alfa --sandbox                     # Use Docker sandbox
./alfa --provider anthropic          # Override AI provider
./alfa --max-iterations 20           # Max AI loops per request
```

---

## Summary for AI Integration

When integrating another tool with Alfa, keep in mind:

1. **All AI interactions go through the Orchestrator**
   - It parses JSON actions from LLM responses
   - It executes actions via Tool Dispatcher
   - It manages conversation context

2. **All file operations are VFS-scoped**
   - Tools receive a VFS instance, not raw filesystem
   - All paths are relative to project root
   - Security enforced at VFS layer

3. **Projects are hot-swappable**
   - AI can create/delete/restore/switch projects
   - No manual restart needed
   - VFS and VCR auto-reinitialize

4. **Git integration is automatic**
   - Commits happen after successful operations
   - Auto-push to `.git-remotes/` for backup
   - Managed by VCR wrapper

5. **Natural language → JSON actions**
   - System prompt teaches AI the action formats
   - AI generates JSON blocks in responses
   - Orchestrator extracts and executes them

If your tool extends Alfa, you'll likely:
- Add new actions to Tool Dispatcher
- Update system prompt with new capabilities
- Potentially hook into project lifecycle events
- Use VFS for any file operations
- Return structured Result objects for feedback
