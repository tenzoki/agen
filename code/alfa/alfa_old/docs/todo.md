# TODO - Alfa Development Roadmap

This document tracks pending features, improvements, and known limitations.

---

## 🚧 High Priority

### Safety Layer (Docker Sandbox)
- [x] Implement Docker-based sandbox for code execution
- [x] Add resource limits (CPU, memory, execution time)
  - [x] `--cpus=1`
  - [x] `--memory=512m`
  - [x] Execution timeouts
- [x] Collect stdout/stderr/exit code from sandboxed execution
- [x] Return structured results to AI Layer
- [x] Integration with Tool Dispatcher for `run_command` and `run_tests`

### Streaming Support
- [ ] Implement `ChatStream` for Claude client
- [ ] Implement `ChatStream` for OpenAI client
- [ ] Add streaming response display in orchestrator
- [ ] Support streaming in voice mode (stream → TTS in chunks)

---

## 🎯 Medium Priority

### Context Management Enhancements
- [ ] Implement context trimming with retrieval
  - [ ] Summarize older edits
  - [ ] Keep only relevant snippets beyond token limit
- [ ] Add vector search for large codebases (RAG)
  - [ ] Embed code snippets
  - [ ] Retrieve relevant context based on query
- [ ] Implement context persistence options
  - [ ] SQLite backend (alternative to JSON)
  - [ ] Redis backend (for distributed setup)
- [ ] Add context visualization/debugging
  - [ ] Show token usage per message
  - [ ] Display context window utilization

### Advanced Orchestration
- [ ] Multi-turn tool use within single request
  - [ ] Allow AI to chain multiple operations
  - [ ] Implement agentic loops (plan → execute → verify → iterate)
- [ ] Add operation rollback/undo
  - [ ] Track all file modifications
  - [ ] Implement undo stack
  - [ ] Add `undo` command
- [ ] Improve error recovery
  - [ ] Better reprompt strategies
  - [ ] Suggest fixes for common errors
  - [ ] Auto-retry with corrections

### Tool Enhancements
- [ ] Add more tool types
  - [ ] `git` operations (branch, merge, diff)
  - [ ] `lint` code (golangci-lint, eslint, etc.)
  - [ ] `format` code (gofmt, prettier, etc.)
  - [ ] `debug` (set breakpoints, inspect variables)
  - [ ] `profile` (CPU, memory profiling)
- [ ] Improve search functionality
  - [ ] Regex support
  - [ ] Case-insensitive options
  - [ ] Multi-file context search
  - [ ] Semantic code search

---

## 📦 Low Priority / Nice-to-Have

### Voice Improvements
- [ ] Support multiple STT providers
  - [ ] Local Whisper (whisper.cpp)
  - [ ] Azure Speech Services
  - [ ] Google Speech-to-Text
- [ ] Support multiple TTS providers
  - [ ] Local TTS (Coqui, Piper)
  - [ ] Azure TTS
  - [ ] Google TTS
- [ ] Voice activity detection (VAD) improvements
  - [ ] Better silence detection
  - [ ] Noise cancellation
  - [ ] Echo cancellation
- [ ] Custom wake word support
  - [ ] "Hey Alfa" activation
  - [ ] Continuous listening mode

### Configuration
- [ ] Web-based configuration UI
- [ ] Per-project configuration overrides
- [ ] Environment-specific configs (dev/staging/prod)
- [ ] API key management (secure storage)

### Documentation
- [ ] Video tutorials
- [ ] Interactive quickstart guide
- [ ] API reference documentation
- [ ] Architecture deep-dive
- [ ] Contributing guidelines

### Testing
- [ ] Increase test coverage
  - [ ] Integration tests for orchestrator
  - [ ] E2E tests for voice pipeline
  - [ ] Stress tests for context management
- [ ] Add benchmarks
  - [ ] VFS performance
  - [ ] Context retrieval speed
  - [ ] Patch application speed
- [ ] CI/CD pipeline
  - [ ] Automated testing on commit
  - [ ] Release automation

### Monitoring & Observability
- [ ] Add structured logging
- [ ] Implement metrics collection
  - [ ] Request latency
  - [ ] Token usage
  - [ ] Error rates
- [ ] Add tracing support (OpenTelemetry)
- [ ] Cost tracking for API usage

### Platform Support
- [ ] Linux audio support (ALSA/PulseAudio)
- [ ] Windows audio support
- [ ] Cross-platform path handling improvements
- [ ] Docker image for containerized deployment

---

## 🐛 Known Issues / Limitations

### Current Limitations
- [ ] Single request-response cycle (no multi-turn agentic loops within one request)
- [ ] No streaming support (responses are buffered)
- [ ] Limited context window handling (no smart truncation)
- [ ] No RAG/vector search for large codebases
- [ ] Voice mode requires macOS or Linux (sox dependency)

### Bug Fixes Needed
- [ ] Handle malformed JSON responses from AI more gracefully
- [ ] Improve error messages when API keys are invalid
- [ ] Fix potential race conditions in audio playback
- [ ] Better handling of network timeouts/retries
- [ ] Validate patch operations before applying (prevent corruption)

---

## 🎨 Future Enhancements

### Advanced AI Features
- [ ] Multi-agent collaboration
  - [ ] Separate agents for planning, coding, testing, reviewing
  - [ ] Agent communication protocol
- [ ] Code review agent
  - [ ] Automated code review on commits
  - [ ] Security vulnerability scanning
  - [ ] Best practice suggestions
- [ ] Test generation
  - [ ] Auto-generate unit tests for functions
  - [ ] Property-based testing suggestions
- [ ] Documentation generation
  - [ ] Auto-generate docstrings/comments
  - [ ] API documentation from code

### IDE Integration
- [ ] VS Code extension
- [ ] IntelliJ plugin
- [ ] Neovim/Vim plugin
- [ ] Language Server Protocol (LSP) support

### Collaboration Features
- [ ] Multi-user support
- [ ] Session sharing
- [ ] Code review workflow integration
- [ ] Team context management

---

## 📋 Completion Status

**Implemented**:
- ✅ AI Layer (Claude, OpenAI)
- ✅ Speech Layer (Whisper STT, OpenAI TTS)
- ✅ Audio System (sox recording/playback, VAD)
- ✅ Instruction Layer (TextPatch format)
- ✅ Orchestrator (control loop, action parsing)
- ✅ Context Manager (persistence, history tracking)
- ✅ Tool Dispatcher (VFS-based file operations)
- ✅ VFS (dual VFS, security, path validation)
- ✅ Version Control Integration (VCR git wrapper)
- ✅ Voice Mode (full STT → AI → TTS pipeline)
- ✅ Headless Mode (autonomous voice agent)
- ✅ Safety Layer (Docker sandbox with resource limits)

**Not Implemented**:
- ❌ Streaming Support (ChatStream)
- ❌ Advanced Context (RAG, vector search)
- ❌ Multi-turn agentic loops

**Estimated Completion**: ~75% of planned features implemented

---

Last Updated: 2025-10-01
