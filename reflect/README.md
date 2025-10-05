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
ls reflect/cells/                     # 30+ cell configurations
cat reflect/cells/anonymization-pipeline.md
cat reflect/cells/document-processing-pipeline.md
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
- `cells/` - Cell configuration documentation
- `agents/` - Agent-specific documentation
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
3. Cell examples: `reflect/cells/*.md` - Configuration patterns
4. Archive: `reflect/archive/*.md` - Historical context

Cross-references to implementation:
- Architecture docs → `/code/<module>/` implementation
- Cell docs → `/workbench/config/*.yaml` configurations
- Agent docs → `/code/agents/<agent>/` implementation

This directory serves as the knowledge base for AI self-modification and developer onboarding.
