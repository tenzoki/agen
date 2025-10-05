# Projects

User project workspace - VFS-scoped directories for AI-driven project development and modification.

## Intent

Provides isolated project workspace for AI operations. Each project directory is VFS-scoped to ensure safe file operations within project boundaries. AI can create, modify, and manage project files without escaping project scope. VCR (git) integration provides automatic versioning.

## Usage

Create new project:
```bash
mkdir -p myproject
cd myproject
# Initialize project structure
git init
```

AI-scoped operations (from alfa):
```bash
../../bin/alfa --workdir=..
> "Create a new project called data-pipeline"
> "Add error handling to myproject/src/main.go"
> "Commit changes to myproject"
```

Direct project access:
```bash
cd myproject
# Project files are VFS-scoped
# Safe AI file operations within this directory
```

## Setup

No dependencies - workspace directory.

Project isolation:
- Each project = separate VFS scope
- AI file operations confined to project directory
- Path traversal prevention (no `../` escapes)
- VCR auto-commit on modifications

Project structure (recommended):
```
myproject/
├── src/          # Source code
├── config/       # Configuration files
├── tests/        # Test files
├── docs/         # Documentation
└── .git/         # Version control
```

## Tests

Test project isolation:
```bash
# Create test project
mkdir test-project
cd test-project

# Verify VFS scoping
# AI cannot access files outside project directory
```

Validate VCR integration:
```bash
cd myproject
git log          # View AI modification history
git diff HEAD~1  # Review last AI changes
```

## Demo

Project workflow:
```bash
# 1. Create project via AI
../../bin/alfa --workdir=..
> "Create a new project called text-processor"

# 2. AI modifies project
> "Add a main.go file to text-processor with basic structure"
> "Add error handling to processData function"

# 3. Review changes
cd text-processor
git log --oneline
# Shows: AI commits with modification descriptions

# 4. Verify isolation
cat ../other-project/secret.txt
# ERROR: VFS blocks access outside project scope
```

VFS security features:
- Absolute path resolution
- Path traversal prevention
- Directory-scoped operations
- Safe AI file access

See [/reflect/architecture/atomic.md](../../reflect/architecture/atomic.md) for VFS and VCR details.
