# Atomic

Foundational utilities with zero AGEN dependencies - VFS (Virtual File System) and VCR (Version Control Repository).

## Intent

Provides reusable, dependency-free utilities for filesystem virtualization and git operations. Atomic components are building blocks used by all other AGEN modules (alfa, cellorg, agents, omni). See [/reflect/architecture/atomic.md](../../reflect/architecture/atomic.md) for detailed architecture.

## Usage

VFS - Directory-scoped filesystem operations:
```go
import "github.com/tenzoki/agen/atomic/vfs"

// Create scoped VFS (all operations confined to base path)
projectVFS := vfs.NewVFS("/workbench/projects/myproject")

// Safe operations (path traversal blocked)
content, _ := projectVFS.Read("src/main.go")  // OK
content, _ := projectVFS.Read("../secret")    // ERROR: blocked

// File operations
projectVFS.Write("config.yaml", data)
files, _ := projectVFS.List("*.go")
```

VCR - Git operations wrapper:
```go
import "github.com/tenzoki/agen/atomic/vcr"

// Initialize repository
vcr := vcr.NewVcr("alfa-ai", "/workbench/projects/myproject")

// Auto-commit changes
hash := vcr.Commit("AI: Added error handling")  // Returns: "a3f5b2c"

// Query history
history, _ := vcr.History(10)  // Last 10 commits
```

## Setup

Dependencies:
- github.com/go-git/go-git/v5 (VCR only)

Install:
```bash
cd code/atomic
go mod download
```

Build:
```bash
go build -v ./vfs ./vcr
```

## Tests

Run all tests:
```bash
go test ./... -v
```

Test specific component:
```bash
go test ./vfs -v    # VFS tests (path validation, security)
go test ./vcr -v    # VCR tests (commit, history, branches)
```

## Demo

VFS demo:
```bash
cd ../../workbench/demos/vfs_demo
go run main.go
```

Demonstrates:
- Directory-scoped operations
- Security validation (path traversal prevention)
- Path resolution and normalization

See [/reflect/architecture/atomic.md](../../reflect/architecture/atomic.md) for integration patterns and use cases.
