# Atomic

**Target Audience**: AI/LLM
**Purpose**: Foundation utilities specification

Foundational utilities with zero AGEN dependencies - VFS abstraction and VCR version control wrapper.

**Key Principle**: Zero external dependencies (stdlib only) - See `guidelines/references/architecture.md`

## Intent

Provide reusable, dependency-free utilities for filesystem virtualization and git operations. Atomic components are building blocks used by all other AGEN modules.

## Components

### VFS (Virtual File System)

**Purpose:** Directory-scoped filesystem operations with path validation and security.

**Core Functions:**
```go
type VFS struct {
    basePath string  // Root directory - all paths scoped within
}

// Path operations
func (v *VFS) Abs(path string) (string, error)      // Resolve to absolute path
func (v *VFS) Validate(path string) error            // Security check (no traversal)
func (v *VFS) Join(elem ...string) string           // Path joining

// File operations
func (v *VFS) Read(path string) ([]byte, error)     // Read file
func (v *VFS) Write(path string, data []byte) error // Write file
func (v *VFS) List(pattern string) ([]string, error) // Glob pattern matching
func (v *VFS) Exists(path string) bool              // File existence check
```

**Security Model:**
- All paths validated against base directory
- Prevents directory traversal attacks (`../` blocked)
- Automatic absolute path resolution
- Path normalization and cleaning

**Usage Pattern:**
```go
// Create scoped VFS
projectVFS := vfs.NewVFS("/workbench/projects/myproject")

// All operations confined to base path
content, _ := projectVFS.Read("src/main.go")  // OK: /workbench/projects/myproject/src/main.go
content, _ := projectVFS.Read("../secret")    // ERROR: path traversal blocked
```

**Use Cases:**
- AI-accessible file operations (prevents escaping project scope)
- Multi-project isolation (each project = separate VFS)
- Configuration management (workbench vs project VFS)
- Sandboxed code execution

### VCR (Version Control Repository)

**Purpose:** Go-git wrapper for commit operations and history management.

**Core Functions:**
```go
type Vcr struct {
    workdir string
    repo    *git.Repository
}

// Repository operations
func NewVcr(user, workdir string) *Vcr                    // Initialize/open repo
func (v *Vcr) Commit(message string) string               // Create commit, return hash
func (v *Vcr) History(maxResults int) ([]CommitInfo, error) // Get commit log
func (v *Vcr) Diff(commitID string) (string, error)      // Get commit diff
func (v *Vcr) ListBranches() ([]string, error)           // List all branches
func (v *Vcr) CurrentBranch() (string, error)            // Get active branch
```

**Automatic Features:**
- Repository initialization on first use
- Default branch creation ("A")
- Auto-add all changes before commit
- Commit hash return for tracking

**Usage Pattern:**
```go
// Initialize VCR
vcr := vcr.NewVcr("alfa-ai", "/workbench/projects/myproject")

// Auto-commit after modifications
vcr.Commit("AI: Added error handling to processData()")
// Returns: "a3f5b2c" (commit hash)

// Query history
history, _ := vcr.History(10)  // Last 10 commits
for _, commit := range history {
    fmt.Printf("%s: %s\n", commit.Hash[:7], commit.Message)
}
```

**Use Cases:**
- AI code modification tracking (auto-commit after each change)
- Project backup (bare repos in `.git-remotes/`)
- Undo/redo operations (branch management)
- Code change auditing (full history of AI modifications)

## Integration Points

**Alfa Integration:**
- Workbench VFS: Configuration and context storage
- Project VFS: Per-project isolation for AI operations
- VCR: Auto-commit after successful tool execution

**Cellorg Integration:**
- VFS: Config file loading (cells.yaml, pool.yaml)
- VCR: Agent deployment versioning

**Omni Integration:**
- VFS: Database path management
- VCR: Schema migration tracking

## Module Structure

```
code/atomic/
├── go.mod                    # Module: github.com/tenzoki/agen/atomic
├── vfs/
│   ├── vfs.go               # VFS implementation
│   └── vfs_test.go          # VFS tests
└── vcr/
    ├── vcr.go               # VCR implementation
    └── vcr_test.go          # VCR tests
```

## Dependencies

**External:**
- `github.com/go-git/go-git/v5` - Git operations

**Internal:**
- None (zero AGEN dependencies by design)

## Setup

```bash
# Install dependencies
go mod download

# Build
go build -v ./code/atomic/...

# Test
go test ./code/atomic/vfs ./code/atomic/vcr
```

## Tests

**VFS Tests** (`vfs/vfs_test.go`):
- Path validation and security
- Read/Write operations
- Glob pattern matching
- Error handling

**VCR Tests** (`vcr/vcr_test.go`):
- Repository initialization
- Commit operations
- History retrieval
- Branch management

**Run Tests:**
```bash
go test ./code/atomic/... -v
```

## Demo

**VFS Demo** (`workbench/demos/vfs_demo/main.go`):
- Directory-scoped operations
- Security validation
- Path resolution

**Usage:**
```bash
go run workbench/demos/vfs_demo/main.go
```
