# AGEN Configuration

Configuration directory for AGEN orchestrator, agents, and alfa workbench.

## Structure

```
workbench/config/
├── cellorg.yaml              # Orchestrator configuration
├── pool.yaml                 # Agent type definitions
├── ai-config.json            # Alfa AI provider settings
├── speech-config.json        # Alfa speech settings
├── README.md                 # This file
├── agents/                   # Agent-specific configs (optional)
├── cells/                    # Cell definitions (organized by purpose)
│   ├── pipelines/           # Sequential processing workflows (7 cells)
│   ├── services/            # Backend/supporting services (6 cells)
│   ├── analysis/            # Content analysis workflows (6 cells)
│   ├── synthesis/           # Output generation (5 cells)
│   └── anonymization.yaml   # Special-purpose cell
└── archive/                  # Deprecated/obsolete configs
```

## Core Configuration Files

### cellorg.yaml
Main orchestrator configuration. Defines:
- Support service settings (agent registry)
- Broker service settings (message routing)
- Self-modification settings
- References to pool and cells

**Usage:**
```bash
bin/orchestrator                      # Uses default cellorg.yaml
bin/orchestrator -config=path.yaml   # Custom config
```

### pool.yaml
Agent type definitions. Maps agent types to binaries, capabilities, and deployment strategies.

**Example:**
```yaml
agent_types:
  - agent_type: "text-extractor"
    binary: "bin/text_extractor"
    operator: "spawn"
    capabilities: ["text-extraction", "pdf", "docx"]
```

### ai-config.json
Alfa AI provider configuration (Anthropic/OpenAI).

**Example:**
```json
{
  "provider": "anthropic",
  "model": "claude-sonnet-4.5",
  "api_key": "${ANTHROPIC_API_KEY}"
}
```

### speech-config.json
Alfa speech-to-text and text-to-speech configuration.

## Cell Definitions

Cells are organized by purpose in `cells/`:

### cells/pipelines/
**Sequential processing workflows** - Multi-stage document/file processing

- `document-processing-pipeline.yaml` - General document processing
- `fast-document-processing-pipeline.yaml` - Optimized for speed
- `intelligent-document-processing-pipeline.yaml` - AI-enhanced processing
- `research-paper-processing-pipeline.yaml` - Academic document handling
- `text-extraction-pipeline.yaml` - Text extraction from various formats
- `file-chunking-pipeline.yaml` - Split files into processable chunks
- `file-transform-pipeline.yaml` - Transform file formats/structure

**Usage:**
```bash
bin/orchestrator -config=cells/pipelines/document-processing-pipeline.yaml
```

### cells/services/
**Backend/supporting services** - Infrastructure and support

- `storage.yaml` - OmniStore storage backend
- `storage-service.yaml` - Alternative storage service
- `rag.yaml` - RAG (Retrieval-Augmented Generation) knowledge backend
- `ocr.yaml` - OCR service integration
- `ocr-production.yaml` - Production OCR configuration
- `native-text-extraction.yaml` - Native text extraction service

**Usage:**
```bash
bin/orchestrator -config=cells/services/rag.yaml
```

### cells/analysis/
**Content analysis workflows** - Examine and classify content

- `content-analysis.yaml` - General content analysis
- `academic-analysis.yaml` - Academic content analysis
- `binary-media.yaml` - Binary files and media analysis
- `structured-data.yaml` - Structured data (JSON/XML) analysis
- `text-deep.yaml` - Deep text analysis
- `fast-analysis.yaml` - Fast analysis for quick insights

**Usage:**
```bash
bin/orchestrator -config=cells/analysis/content-analysis.yaml
```

### cells/synthesis/
**Output generation** - Create derived artifacts

- `reporting.yaml` - Generate reports
- `search-ready.yaml` - Prepare search-optimized output
- `data-export.yaml` - Export processed data
- `document-summary.yaml` - Create document summaries
- `full-analysis.yaml` - Comprehensive analysis synthesis

**Usage:**
```bash
bin/orchestrator -config=cells/synthesis/reporting.yaml
```

### cells/anonymization.yaml
**Special-purpose cell** - PII/privacy anonymization pipeline using NER

**Usage:**
```bash
bin/orchestrator -config=cells/anonymization.yaml
```

## Agent Configuration (Optional)

### agents/*.yaml
Agent-specific configuration overrides. Only create these if you need to override agent defaults.

**When to use:**
- Override model paths
- Adjust performance settings
- Enable debug mode
- Custom agent behavior

**Example** `agents/ner_agent.yaml`:
```yaml
model_path: "/custom/path/to/model"
batch_size: 32
debug: true
```

## Agent Configuration Resolution

Agents follow this resolution order (highest priority first):

1. **Command-line flag**: `--config=/path/to/config.yaml`
2. **Environment variable**: `AGEN_CONFIG_PATH=/path/to/config.yaml`
3. **Orchestrator workbench**: `$AGEN_WORKBENCH_DIR/config/agents/<agent>.yaml`
4. **CWD-relative simple**: `./config/<agent>.yaml`
5. **CWD-relative AGEN**: `./workbench/config/agents/<agent>.yaml`
6. **Binary-relative**: `<binary-dir>/config/<agent>.yaml`
7. **Embedded defaults**: Always works, no config needed

See [guidelines/references/config-standards.md](../../guidelines/references/config-standards.md) for details.

## Usage Examples

### Run Orchestrator with Default Config
```bash
bin/orchestrator
```

### Run Specific Cell
```bash
bin/orchestrator -config=cells/pipelines/document-processing-pipeline.yaml
bin/orchestrator -config=cells/services/rag.yaml
bin/orchestrator -config=cells/analysis/content-analysis.yaml
```

### Dry Run (Validate Config)
```bash
bin/orchestrator -config=cells/pipelines/document-processing-pipeline.yaml -dry-run
```

### Run Alfa Workbench
```bash
bin/alfa                              # Default config
bin/alfa --config=custom-ai.json     # Custom AI config
bin/alfa --project=my-project        # With specific project
```

### Run Standalone Agent
```bash
bin/ner_agent                        # Finds config automatically
bin/ner_agent --config=my-config.yaml  # Custom config
AGEN_WORKBENCH_DIR=/path bin/ner_agent # Via environment
```

## Environment Variables

- `AGEN_CONFIG_PATH`: Explicit config file path
- `AGEN_WORKBENCH_DIR`: Workbench directory (orchestrator sets this for spawned agents)
- `AGEN_FRAMEWORK_DIR`: Framework root directory
- `AGEN_SELF_MODIFICATION`: Enable/disable self-modification (true/false)
- `ANTHROPIC_API_KEY`: Anthropic API key (alfa)
- `OPENAI_API_KEY`: OpenAI API key (alfa)

## Self-Modification

AGEN can modify its own code when needed. Configuration in `cellorg.yaml`:

```yaml
self_modification:
  enabled: true              # Allow framework modifications
  require_tests: true        # Run tests before committing
  auto_commit: true          # Auto-commit successful changes
```

When enabled, AGEN can:
- Add new agents to `code/agents/`
- Update configuration files
- Improve framework code
- Update documentation

Changes are tracked in separate git repositories:
- Framework changes → `agen/.git`
- Workbench changes → `workbench/.git`
- Project changes → `workbench/projects/<name>/.git`

## Finding Cells

### By Purpose
- **Need to process documents?** → `cells/pipelines/`
- **Need backend services?** → `cells/services/`
- **Need to analyze content?** → `cells/analysis/`
- **Need to generate output?** → `cells/synthesis/`
- **Need PII anonymization?** → `cells/anonymization.yaml`

### Full Documentation
- Cell definitions: [reflect/cells/](../../reflect/cells/)
- Configuration syntax: [reflect/architecture/cellorg.md](../../reflect/architecture/cellorg.md)
- Agent development: [guidelines/references/agent-patterns.md](../../guidelines/references/agent-patterns.md)

## Archive

`archive/` contains deprecated configurations:
- `cells.yaml` - Old monolithic cell definitions (34KB)
- Duplicate/obsolete configs

Keep for reference but do not use in production.
