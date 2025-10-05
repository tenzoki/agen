# Code

Main source code directory containing all AGEN modules.

## Intent

Houses core implementation modules for the AGEN system: agents (processing catalog), alfa (AI workbench), atomic (foundation utilities), cellorg (orchestration framework), and omni (unified storage). Each module is independently buildable with clear dependency boundaries.

## Usage

Navigate to specific module for development:
```bash
cd code/agents      # Agent catalog (27 agents)
cd code/alfa        # AI workbench
cd code/atomic      # VFS and VCR utilities
cd code/cellorg     # Cell orchestration framework
cd code/omni        # Unified storage backend
```

Build all modules:
```bash
# From repository root
go build -o bin/orchestrator ./code/cellorg/cmd/orchestrator
go build -o bin/alfa ./code/alfa/cmd/alfa
# Agents built individually (see code/agents/README.md)
```

## Setup

Module dependencies (build order):
```
atomic → (no dependencies)
omni → atomic
cellorg → atomic, omni
agents → cellorg, omni
alfa → cellorg, atomic
```

Install dependencies:
```bash
cd code/<module>
go mod download
```

## Tests

Run tests for all modules:
```bash
cd code
go test ./... -v
```

Run tests for specific module:
```bash
go test ./code/cellorg/... -v
go test ./code/agents/... -v
```

## Demo

See module-specific READMEs:
- [agents/README.md](agents/README.md) - Agent catalog and development
- [alfa/README.md](alfa/README.md) - AI workbench usage
- [atomic/README.md](atomic/README.md) - VFS and VCR utilities
- [cellorg/README.md](cellorg/README.md) - Cell orchestration
- [omni/README.md](omni/README.md) - Unified storage

Architecture documentation: [/reflect/architecture/README.md](../reflect/architecture/README.md)
