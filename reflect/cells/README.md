# Cell Documentation

**Target Audience**: AI/LLM and developers
**Purpose**: Detailed specifications for all AGEN cell configurations

This directory contains technical documentation for each cell defined in `workbench/config/cells/`. Documentation is organized to mirror the config structure.

## Structure

```
reflect/cells/
├── pipelines/       # Sequential processing workflows (8 cells)
├── services/        # Backend/supporting services (6 cells)
├── analysis/        # Content analysis workflows (6 cells)
├── synthesis/       # Output generation (5 cells)
└── guides/          # Tutorials and setup guides (6 docs)
```

## Cell Categories

### pipelines/
**Sequential processing workflows** - Multi-stage document/file processing

Each pipeline doc describes:
- Agent composition and dependencies
- Input/output specifications
- Configuration options
- Usage examples

Files: `anonymization-pipeline.md`, `document-processing-pipeline.md`, `fast-document-processing-pipeline.md`, `file-chunking-pipeline.md`, `file-transform-pipeline.md`, `intelligent-document-processing-pipeline.md`, `research-paper-processing-pipeline.md`, `text-extraction-pipeline.md`

### services/
**Backend/supporting services** - Infrastructure and support

Each service doc describes:
- Service purpose and capabilities
- Integration patterns
- Configuration requirements
- Performance characteristics

Files: `storage.md`, `rag.md`, `ocr.md`, `ocr-production.md`, `storage-service.md`, `native-text-extraction.md`

### analysis/
**Content analysis workflows** - Examine and classify content

Each analysis doc describes:
- Analysis techniques and algorithms
- Output formats
- Accuracy and performance
- Use cases

Files: `content-analysis.md`, `academic-analysis.md`, `binary-media.md`, `structured-data.md`, `text-deep.md`, `fast-analysis.md`

### synthesis/
**Output generation** - Create derived artifacts

Each synthesis doc describes:
- Synthesis strategy
- Input requirements
- Output formats
- Quality metrics

Files: `reporting.md`, `search-ready.md`, `data-export.md`, `document-summary.md`, `full-analysis.md`

### guides/
**Tutorials and setup** - How-to guides for complex features

Specialized documentation for setup, troubleshooting, and advanced usage.

Files: `anonymization-quickstart.md`, `anonymization-setup.md`, `gox_anonymization_concept.md`, `all-in-one-ocr-container.md`, `ner-agent-complete.md`, `rag-implementation-progress.md`

## Relationship to Config

Each cell documentation file corresponds to a YAML config:

```
reflect/cells/pipelines/anonymization-pipeline.md
  ↓
workbench/config/cells/pipelines/anonymization-pipeline.yaml
```

**Documentation conventions:**
- Filename matches config: `<name>.md` ↔ `<name>.yaml`
- Category matches directory: same subdirectory in both trees
- Content describes implementation: architecture, agents, data flow

## Usage

**Find cell documentation:**
```bash
# By category
ls reflect/cells/pipelines/
ls reflect/cells/services/

# By name
find reflect/cells -name "*anonymization*"
```

**Read before using a cell:**
```bash
# 1. Read documentation
cat reflect/cells/pipelines/anonymization-pipeline.md

# 2. Review config
cat workbench/config/cells/pipelines/anonymization-pipeline.yaml

# 3. Run cell
bin/orchestrator -config=workbench/config/cells/pipelines/anonymization-pipeline.yaml
```

## Contributing

When adding new cells:

1. Create YAML config in `workbench/config/cells/<category>/`
2. Create matching doc in `reflect/cells/<category>/`
3. Use same category and base filename
4. Follow documentation template (see existing docs)
5. Update this README if adding new categories

## See Also

- [Configuration Guide](../../workbench/config/README.md) - How to use and organize configs
- [Cellorg Architecture](../architecture/cellorg.md) - Cell framework design
- [Agent Patterns](../../guidelines/references/agent-patterns.md) - How to create agents
