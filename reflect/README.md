# Reflect

Documentation and architecture specifications for AGEN system.

## Intent

Central documentation hub for AGEN architecture, design principles, cell configurations, and historical decisions. Entry point to all architectural documentation for AI reasoning and developer understanding.

## Usage

Start here for architecture understanding:
```bash
# Core architecture documentation
cat reflect/architecture/README.md    # System overview and design principles
cat reflect/architecture/agents.md    # Agent concept and catalog
cat reflect/architecture/cellorg.md   # Cell orchestration framework
cat reflect/architecture/atomic.md    # Foundation utilities (VFS, VCR)
cat reflect/architecture/omni.md      # Unified storage backend
```

Browse cell configurations:
```bash
ls reflect/cells/                     # Organized by category
cat reflect/cells/README.md           # Cell documentation index
cat reflect/cells/pipelines/anonymization-pipeline.md
cat reflect/cells/services/rag.md
```

Review historical decisions:
```bash
ls reflect/archive/                   # Architecture evolution
cat reflect/archive/concept.md        # Original concept
cat reflect/archive/technical-documentation.md
```

## Setup

No dependencies - documentation only.

Directory structure:
- `architecture/` - Core architecture specs (start here)
- `cells/` - Cell configuration documentation (organized by purpose)
  - `pipelines/` - Processing workflows (8 cells)
  - `services/` - Backend services (6 cells)
  - `analysis/` - Content analysis (6 cells)
  - `synthesis/` - Output generation (5 cells)
  - `guides/` - Tutorials and setup (6 docs)
- `archive/` - Historical decisions and evolution

## Tests

Documentation validation:
```bash
# Check for broken links
find reflect -name "*.md" -exec grep -l "](/" {} \;

# Verify referenced paths exist
grep -r "\[.*\](.*\.md)" reflect/architecture/
```

## Demo

Navigation workflow:
1. Start: `reflect/architecture/README.md` - System overview
2. Module: `reflect/architecture/<module>.md` - Specific component
3. Cell index: `reflect/cells/README.md` - Cell categories and index
4. Cell examples: `reflect/cells/<category>/*.md` - Configuration patterns
5. Archive: `reflect/archive/*.md` - Historical context

Cross-references to implementation:
- Architecture docs → `/code/<module>/` implementation
- Cell docs → `/workbench/config/*.yaml` configurations
- Agent docs → `/code/agents/<agent>/` implementation

This directory serves as the knowledge base for AI self-modification and developer onboarding.
