# Demos

Demonstration applications showcasing AGEN capabilities - cell orchestration, AI integration, speech interface, and VFS utilities.

## Intent

Provides runnable examples demonstrating core AGEN features: cell-based pipelines, AI workbench integration, voice interaction, filesystem virtualization, and sandboxed execution. Each demo is self-contained and executable.

## Usage

Run demo:
```bash
cd <demo_name>
go run main.go
```

List available demos:
```bash
ls -1
# gox_demo
# gox_anonymization
# ai_demo
# speech_ai
# speech_demo
# voice_interactive
# audio_demo
# vfs_demo
# sandbox_demo
```

## Setup

Dependencies vary by demo:
- Go 1.24.3+ (all demos)
- sox (audio/speech demos)
- Docker (sandbox_demo)
- Speech recognition service (speech demos)

Build demo:
```bash
cd <demo_name>
go build -o demo main.go
./demo
```

## Tests

Test demos:
```bash
# Test all demos
for demo in */; do
    cd $demo
    go test ./... -v
    cd ..
done
```

Test specific demo:
```bash
cd gox_demo
go test ./... -v
```

## Demo

### Cell Orchestration Demos

**gox_demo** - Full cell orchestration demonstration
```bash
cd gox_demo
go run main.go
```
Demonstrates: Multi-agent pipelines, dependency resolution, message routing

**gox_anonymization** - Privacy pipeline with PII anonymization
```bash
cd gox_anonymization
go run main.go
```
Demonstrates: NER agent, anonymization store, privacy-preserving processing

### AI Integration Demos

**ai_demo** - AI workbench and orchestration
```bash
cd ai_demo
go run main.go
```
Demonstrates: AI reasoning, cell modification, self-modification flow

**speech_ai** - Voice + AI integration
```bash
cd speech_ai
go run main.go
```
Demonstrates: Speech interface, AI orchestration, voice-driven workflows

### Speech Interface Demos

**speech_demo** - Speech recognition and synthesis
```bash
cd speech_demo
go run main.go
```
Demonstrates: Voice input/output, audio processing

**voice_interactive** - Interactive voice interface
```bash
cd voice_interactive
go run main.go
```
Demonstrates: Real-time voice interaction, command processing

**audio_demo** - Audio processing utilities
```bash
cd audio_demo
go run main.go
```
Demonstrates: Audio file processing, format conversion

### Utility Demos

**vfs_demo** - Virtual filesystem operations
```bash
cd vfs_demo
go run main.go
```
Demonstrates: VFS scoping, path validation, security features

**sandbox_demo** - Sandboxed code execution
```bash
cd sandbox_demo
go run main.go
```
Demonstrates: Docker isolation, safe code execution

See individual demo directories for detailed READMEs and specific usage instructions.
