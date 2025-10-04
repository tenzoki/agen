# Alfa - AI Coding Assistant

**Alfa** is a voice-controlled AI coding assistant that helps you write, debug, and manage code through natural conversation. It combines speech recognition, large language models, and sandboxed code execution to provide an interactive development experience.

---

## 🎯 Features

- **Voice Control**: Speak naturally to control your development environment
- **Multi-Provider AI**: Support for Anthropic Claude and OpenAI GPT-4
- **Multi-Project Management**: Create, switch, and manage multiple projects in one workbench
- **Docker Sandbox**: Optional Docker-based isolated execution with resource limits
- **Dual VFS Security**: Separate workbench and project virtual file systems
- **Code Patching**: Structured JSON-based patches for precise code modifications
- **Context Management**: Persistent conversation history and file tracking
- **Auto-commit & Push**: Automatic git commits and push to local backup after operations
- **Project Recovery**: Deleted projects can be restored from local backups
- **Dual Modes**: Confirm (manual approval) or Allow-all (autonomous)
- **Gox Integration** (Alpha): Advanced multi-agent workflows via cells (optional)

---

## 🚀 Quick Start

### Prerequisites

```bash
# Install Go 1.24.3+
go version

# For Docker sandbox (optional)
# Install Docker Desktop or docker engine

# For voice features (optional)
brew install sox

# Set API key(s)
export OPENAI_API_KEY="sk-..."
# or
export ANTHROPIC_API_KEY="sk-ant-..."
```

### Build

```bash
go build -o alfa ./cmd/alfa
```

### Run

```bash
# First time? Just run alfa - it will guide you!
./alfa

# Create additional projects
./alfa --create-project myapp

# List all projects
./alfa --list-projects

# Work on a specific project (text mode with confirmations)
./alfa --project myapp

# Voice mode (hybrid text/voice input)
./alfa --project myapp --voice

# Headless mode (autonomous voice agent, no confirmations)
./alfa --project myapp --headless

# If you have only one project, --project is optional
./alfa
```

**First-Time Experience**: When you run `./alfa` with no projects, you'll see a friendly welcome prompt asking you to create your first project. No complicated setup - just enter a name and start coding!

---

## 📖 Usage

### Command Line Options

```bash
./alfa [options]

Project Management:
  --project string           Select project to work on
  --list-projects            List all projects and exit
  --create-project string    Create a new project and exit
  --delete-project string    Delete a project (keeps backup) and exit
  --restore-project string   Restore a deleted project and exit

General Options:
  --workdir string         Working directory (default: current directory)
  --config string          Config file path (default: config/ai-config.json)
  --provider string        AI provider override (anthropic or openai)
  --mode string            Execution mode: confirm or allow-all (default: confirm)
  --max-iterations int     Maximum AI iterations per request (default: 10)
  --voice                  Enable voice input/output
  --headless               Autonomous voice agent (enables --voice and --mode allow-all)
  --sandbox                Use Docker sandbox for command execution
  --sandbox-image string   Docker image for sandbox (default: golang:1.24-alpine)
  --enable-gox             Enable Gox advanced features (cells, RAG, etc.)
  --gox-config string      Path to Gox configuration directory (default: config/gox)
```

### Configuration

Create `config/ai-config.json`:

```json
{
  "default_provider": "openai",
  "providers": {
    "openai": {
      "model": "gpt-4",
      "max_tokens": 4096,
      "temperature": 1.0,
      "timeout": 60000000000,
      "retry_count": 3,
      "retry_delay": 1000000000
    },
    "anthropic": {
      "model": "claude-3-5-sonnet-20241022",
      "max_tokens": 4096,
      "temperature": 1.0,
      "timeout": 60000000000,
      "retry_count": 3,
      "retry_delay": 1000000000
    }
  }
}
```

API keys are loaded from environment variables:
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`

---

## 🏗️ Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                    ORCHESTRATOR                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ Speech Layer │  │   AI Layer   │  │ Tool Layer   │  │
│  │  (STT/TTS)   │  │(Claude/GPT-4)│  │ (VFS-based)  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Context    │  │     VCR      │  │  TextPatch   │  │
│  │   Manager    │  │  (Git Auto)  │  │   (Patches)  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Virtual File System (VFS)

Alfa uses a dual VFS design for security with multi-project support:

```
workbench/                    # Workbench VFS (config, context, history)
├── config/
│   └── ai-config.json        # AI provider configuration
├── .alfa/
│   ├── context.json          # Conversation history + active project
│   └── history.log           # Operation log
├── projects/                 # Multiple projects (Project VFS)
│   ├── myapp/                # Each project is a git repo
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── go.mod
│   │   └── [your code]
│   ├── backend/
│   └── frontend/
└── .git-remotes/             # Local bare repos (backup/recovery)
    ├── myapp.git/
    ├── backend.git/
    └── frontend.git/
```

**Security Model**:
- **Project VFS**: AI can only access/modify files in selected project directory
- **Workbench VFS**: Config and context isolated from AI operations
- **Path Validation**: All `..` traversal attempts blocked
- **Auto-backup**: Each commit automatically pushes to local bare repo
- **Recovery**: Deleted projects can be restored from `.git-remotes/`

---

## 🎤 Voice Mode

### Interactive Voice

Press Enter to start recording. Alfa will:
1. Record your speech (auto-stops after 2s silence)
2. Transcribe using Whisper
3. Send to AI for processing
4. Speak the response using TTS
5. Execute requested actions (with confirmation in default mode)

### Headless Mode

Fully autonomous voice-controlled agent:

```bash
./alfa --headless
```

- No confirmations required
- Continuous voice interaction
- Automatic execution of all operations

---

## 🛠️ Available Actions

Alfa understands and executes the following operations:

### Code Patching
```json
{
  "action": "patch",
  "file": "main.go",
  "operations": [
    {"line": 10, "type": "insert", "content": ["new line"]},
    {"line": 15, "type": "delete"},
    {"line": 20, "type": "replace", "content": ["updated line"]}
  ]
}
```

### File Operations
- `read_file`: Read file contents
- `write_file`: Create or overwrite files
- `search`: Search codebase for patterns

### Execution
- `run_command`: Execute shell commands (optionally in Docker sandbox)
- `run_tests`: Run test suites (optionally in Docker sandbox)

### Project Management (AI-Enabled)
- `list_projects`: List all projects in workbench
- `create_project`: Create a new project and **automatically switch to it**
- `delete_project`: Delete a project (backup kept in `.git-remotes/` for recovery)
- `restore_project`: Restore a deleted project from backup and **automatically switch to it**
- `switch_project`: Switch to another project **in real-time** (no restart needed)

```json
{
  "action": "list_projects"
}
```

```json
{
  "action": "create_project",
  "name": "frontend"
}
```

```json
{
  "action": "delete_project",
  "name": "old-project"
}
```

```json
{
  "action": "restore_project",
  "name": "old-project"
}
```

```json
{
  "action": "switch_project",
  "name": "backend"
}
```

**Example AI Conversation:**
```
User: "Show me all my projects"
AI: [executes list_projects action]

User: "Create a new project called api-service"
AI: [executes create_project with name="api-service"]
    ✅ Switched to project 'api-service'
    [AI can now work on api-service immediately]

User: "Delete the old-project"
AI: [executes delete_project with name="old-project"]
    ✅ Project 'old-project' deleted successfully
    Backup kept in .git-remotes/ for recovery

User: "Restore old-project"
AI: [executes restore_project with name="old-project"]
    ✅ Project 'old-project' restored successfully
    ✅ Switched to project 'old-project'
    [AI can now work on old-project immediately]

User: "Switch to the backend project"
AI: [executes switch_project with name="backend"]
    ✅ Switched to project 'backend'
    [AI can now work on backend immediately]
```

**Hot Project Switching**: When the AI creates, restores, or switches projects, Alfa automatically reinitializes the VFS, VCR, and context for the new project - no restart required!

**Safe Deletion & Recovery**: Deleted projects are backed up in `.git-remotes/` and can be restored by the AI with `restore_project` or via CLI with `alfa --restore-project <name>`

---

## 🧪 Testing

Run the test suite:

```bash
# All tests
go test ./...

# Specific package
go test ./internal/vfs
go test ./internal/ai
go test ./internal/sandbox
go test ./internal/gox
go test ./test/textpatch
go test ./test/gox
```

### Demo Applications

Explore individual components:

```bash
# AI layer
go run demo/ai/main.go

# Speech (STT/TTS)
go run demo/speech/main.go

# Voice interactive assistant
go run demo/voice_interactive/main.go

# VFS demonstration
go run demo/vfs_demo/main.go

# Docker sandbox demonstration
go run demo/sandbox_demo/main.go

# Gox integration demonstration
go run demo/gox_demo/main.go
```

---

## 📚 Project Structure

```
alfa/
├── cmd/
│   └── alfa/              # Main CLI entry point
├── internal/
│   ├── ai/                # LLM clients (Claude, OpenAI)
│   ├── audio/             # Audio recording/playback (sox)
│   ├── context/           # Context manager (tracks active project)
│   ├── gox/               # Gox orchestrator wrapper (cell management)
│   ├── orchestrator/      # Main control loop
│   ├── project/           # Project manager (multi-project support)
│   ├── sandbox/           # Docker sandbox
│   ├── speech/            # STT/TTS (Whisper, OpenAI TTS)
│   ├── textpatch/         # Code patching
│   ├── tools/             # Tool dispatcher (VFS-based)
│   ├── vcr/               # Git wrapper (auto-push to local remote)
│   └── vfs/               # Virtual file system
├── test/                  # Test suites
├── demo/                  # Demonstration programs
├── config/                # Configuration files
└── docs/                  # Documentation
```

---

## 🔧 Gox Integration (Alpha)

Alfa integrates with [Gox](https://github.com/tenzoki/gox) to support advanced multi-agent workflows via cells.

**Status**: Alpha - Placeholder implementation awaiting `pkg/orchestrator` publication

### Features

#### Core Cell Management
- **Cell Management**: Start/stop cells (agent networks)
- **Event System**: Pub/sub communication with cells
- **VFS Isolation**: Per-project cell isolation
- **AI Integration**: AI can manage cells via JSON actions

#### NER & Anonymization (New)
- **Named Entity Recognition**: Extract PERSON, ORG, LOC entities from text (100+ languages)
- **Text Anonymization**: Replace sensitive data with reversible pseudonyms
- **GDPR Compliance**: Privacy-preserving text processing with mapping storage
- **Multilingual Support**: XLM-RoBERTa model for cross-language entity extraction

### Prerequisites for NER/Anonymization

Before using NER and anonymization features, you must:

1. **Install ONNXRuntime** (required for model inference)
   ```bash
   brew install onnxruntime  # macOS
   ```

2. **Download Models** (~3GB, see `docs/gox-models-integration.md`)
   ```bash
   # Clone latest gox
   cd /tmp
   git clone https://github.com/tenzoki/gox.git
   cd gox/models

   # Setup Python environment
   python3 -m venv venv
   source venv/bin/activate
   pip install -r requirements.txt

   # Download and convert models
   python download_and_convert.py

   # Copy models to Alfa workbench
   mkdir -p /path/to/alfa/workbench/models/ner
   cp /tmp/gox/models/ner/*.onnx /path/to/alfa/workbench/models/ner/
   cp /tmp/gox/models/ner/tokenizer.json /path/to/alfa/workbench/models/ner/
   ```

3. **Set Environment Variables** (for CGO compilation)
   ```bash
   export CGO_CFLAGS="-I/opt/homebrew/include"
   export CGO_LDFLAGS="-L/opt/homebrew/lib -lonnxruntime"
   export DYLD_LIBRARY_PATH="/opt/homebrew/lib:$DYLD_LIBRARY_PATH"
   ```

**Note**: See `docs/gox-models-integration.md` for complete installation instructions and troubleshooting.

### Usage

```bash
# Enable Gox features
./alfa --enable-gox --project myproject

# AI can now use cell actions:
# - start_cell
# - stop_cell
# - list_cells
# - query_cell

# NER & Anonymization actions:
# - extract_entities
# - anonymize_text
# - deanonymize_text
```

### Example AI Workflows

#### Cell Management
```
User: "Start RAG cell for this project"
AI: [starts rag:knowledge-backend cell]

User: "Find authentication code"
AI: [queries cell, receives context, provides answer]
```

#### Named Entity Recognition
```
User: "Extract entities from this text: Angela Merkel met Emmanuel Macron in Berlin"
AI: [Starts NER cell, extracts entities]
Result:
  - Angela Merkel (PERSON)
  - Emmanuel Macron (PERSON)
  - Berlin (LOC)
```

#### Text Anonymization
```
User: "Anonymize this customer support ticket: John Smith called about his order"
AI: [Starts anonymization pipeline]
Original: "John Smith called about his order"
Anonymized: "PERSON_123456 called about his order"
Mappings: {"John Smith": "PERSON_123456"}

User: "Now restore the original text"
AI: [Deanonymizes using mappings]
Restored: "John Smith called about his order"
```

For full documentation, see:
- [docs/gox-integration.md](docs/gox-integration.md) - Cell management guide
- [docs/gox-models-integration.md](docs/gox-models-integration.md) - NER/anonymization setup

---

## 🔒 Security

- **Dual VFS Isolation**: Separate workbench (config/context) and project (code) file systems
- **Path Validation**: Prevents directory traversal attacks (`..` blocking)
- **Docker Sandbox**: Optional containerized execution with resource limits
  - CPU limits (default: 1-2 cores)
  - Memory limits (default: 512MB-1GB)
  - Network isolation (default: disabled)
  - Read-only filesystem (except /tmp)
  - All capabilities dropped
  - No new privileges
- **Confirmation Mode**: Manual approval before operations (default)
- **Read-only VFS**: Optional read-only mode for analysis tasks
- **No External Access**: AI cannot access files outside project directory

---

## 📝 Development

### Adding New AI Providers

Implement the `LLM` interface in `internal/ai`:

```go
type LLM interface {
    Chat(ctx context.Context, messages []Message) (*Response, error)
    ChatStream(ctx context.Context, messages []Message) (<-chan string, <-chan error)
    Model() string
    Provider() string
}
```

### Adding New Tools

Add tool implementations in `internal/tools/tools.go`:

```go
func (d *Dispatcher) executeNewTool(action Action) Result {
    // Implementation using d.vfs for file access
}
```

---

## 🐛 Troubleshooting

### Voice Not Working

```bash
# Check sox installation
which sox

# Install if missing
brew install sox

# Verify OPENAI_API_KEY is set
echo $OPENAI_API_KEY
```

### Audio Playback Issues (macOS)

Alfa automatically uses `afplay` (native macOS) as a fallback if sox isn't available.

### Docker Sandbox Not Working

```bash
# Check Docker installation
docker --version

# Test Docker is running
docker ps

# Pull default sandbox image
docker pull golang:1.24-alpine
```

### Context Not Persisting

Check that `.alfa/` directory is writable:
```bash
ls -la .alfa/
```

---

## 📄 License

This project is licensed under the [European Union Public Licence v1.2 (EUPL)](https://joinup.ec.europa.eu/collection/eupl/eupl-text-eupl-12).

---

## 🙏 Acknowledgments

- [Anthropic Claude](https://www.anthropic.com/) - AI provider
- [OpenAI](https://openai.com/) - GPT-4, Whisper, TTS
- [Sox](http://sox.sourceforge.net/) - Audio processing

---

## 📬 Support

For issues, feature requests, or questions, please open an issue on the repository.
