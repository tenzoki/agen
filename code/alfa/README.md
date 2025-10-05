# Alfa

AI workbench module - speech interface, AI orchestration, and project management for AGEN system.

## Intent

Provides AI-driven workbench for building and modifying complex workflows through natural language. Integrates speech interface, AI reasoning layer, project isolation (VFS), and version control (VCR) to enable self-modifying cell-based applications. See [/reflect/architecture/README.md](../../reflect/architecture/README.md) for core design principles.

## Usage

Launch workbench:
```bash
cd code/alfa
go build -o ../../bin/alfa ./cmd/alfa
../../bin/alfa --workdir=../../workbench
```

Interact via natural language:
```
> "Process all files in input/ through text-pipeline"
> "Create a new cell for document anonymization"
> "Show me the last 5 commits in this project"
```

Run with speech interface:
```bash
../../bin/alfa --workdir=../../workbench --speech
```

## Setup

Dependencies:
- cellorg framework (github.com/tenzoki/agen/cellorg)
- atomic utilities (github.com/tenzoki/agen/atomic)
- Optional: sox (for audio processing)
- Optional: speech recognition service

Build:
```bash
cd code/alfa
go build -o ../../bin/alfa ./cmd/alfa
```

Configuration:
- Workbench directory: `--workdir` flag (default: `./workbench`)
- Speech config: `workbench/config/ai-config.json`

## Tests

Run unit tests:
```bash
go test ./... -v
```

Integration tests:
```bash
go test ./internal/... -integration
```

## Demo

Demo applications in workbench:
- `/workbench/demos/ai_demo` - AI orchestration demo
- `/workbench/demos/speech_ai` - Voice + AI integration
- `/workbench/demos/speech_demo` - Speech interface only

Run demo:
```bash
cd ../../workbench/demos/ai_demo
go run main.go
```

See [/reflect/architecture/README.md](../../reflect/architecture/README.md) for self-modification flow and architecture details.
