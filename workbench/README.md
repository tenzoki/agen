# Workbench

Operational workspace for AGEN system - cell configurations, demos, and user projects.

## Intent

Provides isolated workspace for AI operations with cell configurations (config/), demonstration applications (demos/), and user project directories (projects/). VFS-scoped to ensure safe AI file operations within workbench boundaries.

## Usage

Launch alfa workbench:
```bash
cd workbench
../bin/alfa --workdir=.
```

Run orchestrator with cell configuration:
```bash
../bin/orchestrator -config=./config/text-pipeline.yaml
```

Execute demo:
```bash
cd demos/gox_demo
go run main.go
```

Create new project:
```bash
mkdir -p projects/myproject
cd projects/myproject
# VFS-scoped operations ensure isolation
```

## Setup

Dependencies:
- Orchestrator binary: `bin/orchestrator`
- Alfa binary: `bin/alfa` (optional)
- Agent binaries in `bin/` directory

Directory structure:
- `config/` - Cell configurations (30+ YAML files)
- `demos/` - Demo applications (9 demos)
- `projects/` - User project workspace

## Tests

Test cell configurations:
```bash
# Validate YAML syntax
for file in config/*.yaml; do
    ../bin/orchestrator -config=$file -validate
done
```

Run demo tests:
```bash
cd demos/gox_demo
go test ./... -v
```

## Demo

Available demos in `demos/`:
- `gox_demo/` - Full cell orchestration demo
- `gox_anonymization/` - Privacy pipeline demo
- `ai_demo/` - AI integration demo
- `speech_ai/` - Voice + AI demo
- `speech_demo/` - Speech interface demo
- `voice_interactive/` - Voice interaction demo
- `audio_demo/` - Audio processing demo
- `vfs_demo/` - VFS utilities demo
- `sandbox_demo/` - Sandboxed execution demo

Example workflow:
```bash
# 1. Configure cell
cp config/document-processing-pipeline.yaml config/my-pipeline.yaml
# 2. Edit configuration
vim config/my-pipeline.yaml
# 3. Deploy cell
../bin/orchestrator -config=./config/my-pipeline.yaml
# 4. Process files
cp input/*.pdf input/
```

See subdirectory READMEs:
- [config/README.md](config/README.md) - Cell configurations
- [demos/README.md](demos/README.md) - Demo applications
- [projects/README.md](projects/README.md) - Project workspace
