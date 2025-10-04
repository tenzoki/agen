# Alfa Integration - Quick Start Guide

**5-Minute Setup** | **Updated**: 2025-10-02
**Repository**: https://github.com/tenzoki/gox

---

## Step 1: Add Gox Dependency (30 seconds)

```bash
cd /path/to/alfa
go get github.com/tenzoki/gox@latest
```

---

## Step 2: Create Minimal Config (2 minutes)

Create directory:
```bash
mkdir -p /etc/alfa/gox
```

**`/etc/alfa/gox/gox.yaml`**:
```yaml
app_name: "alfa-gox"
debug: true
support:
  port: ":9000"
broker:
  port: ":9001"
basedir: ["/etc/alfa/gox"]
pool: ["pool.yaml"]
cells: ["cells.yaml"]
```

**`/etc/alfa/gox/pool.yaml`**:
```yaml
pool:
  agent_types: []  # Will add agents later
```

**`/etc/alfa/gox/cells.yaml`**:
```yaml
# Empty for now - cells defined per project
```

---

## Step 3: Basic Integration (2 minutes)

**`alfa/internal/workspace/gox.go`**:
```go
package workspace

import (
    "github.com/tenzoki/gox/pkg/orchestrator"
)

type Workspace struct {
    gox *orchestrator.EmbeddedOrchestrator
}

func NewWorkspace() (*Workspace, error) {
    gox, err := orchestrator.NewEmbedded(orchestrator.Config{
        ConfigPath:      "/etc/alfa/gox",
        DefaultDataRoot: "/var/lib/alfa",
        Debug:           true,
    })
    if err != nil {
        return nil, err
    }

    return &Workspace{gox: gox}, nil
}

func (w *Workspace) Close() error {
    return w.gox.Close()
}
```

**`alfa/cmd/alfa/main.go`**:
```go
func main() {
    ws, err := workspace.NewWorkspace()
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    // Your Alfa logic here
}
```

---

## Step 4: Test It (30 seconds)

```bash
go build ./cmd/alfa
./alfa
```

**Expected output**:
```
[Gox Embedded] Initialized with agent deployment support
```

---

## Step 5: Add Event Subscriptions (Optional)

```go
func (w *Workspace) OpenProject(path string) error {
    projectID := filepath.Base(path)

    // Subscribe to project events
    events := w.gox.Subscribe(projectID + ":*")
    go func() {
        for event := range events {
            log.Printf("[%s] Event: %v", projectID, event.Data)
        }
    }()

    return nil
}
```

---

## Next Steps

### To Actually Deploy Agents:

1. **Build Gox agents**:
   ```bash
   cd /path/to/gox
   make build
   ```

2. **Start Gox services** (temporary - until Phase 3):
   ```bash
   cd /path/to/gox
   ./build/gox --config=/etc/alfa/gox/gox.yaml
   ```

3. **Define a cell** in `/etc/alfa/gox/cells.yaml`

4. **Start cell** in Alfa:
   ```go
   w.gox.StartCell("my-cell", orchestrator.CellOptions{
       ProjectID: "project-123",
       VFSRoot:   "/path/to/project",
   })
   ```

---

## Troubleshooting

### "Failed to load gox.yaml"
**Fix**: Check ConfigPath points to directory with gox.yaml

### "Failed to load pool.yaml"
**Fix**: Ensure pool.yaml exists (can be empty)

### All tests pass but nothing happens?
**Normal**: Without cells defined and agents built, orchestrator just initializes

---

## Complete Example

```go
package main

import (
    "github.com/tenzoki/gox/pkg/orchestrator"
    "log"
    "time"
)

func main() {
    // Initialize
    gox, _ := orchestrator.NewEmbedded(orchestrator.Config{
        ConfigPath: "/etc/alfa/gox",
        Debug:      true,
    })
    defer gox.Close()

    // Subscribe to events
    events := gox.Subscribe("test:*")
    go func() {
        for event := range events {
            log.Printf("Event: %v", event.Data)
        }
    }()

    // Publish event
    gox.Publish("test:hello", map[string]interface{}{
        "message": "Hello from Alfa!",
    })

    // Wait a bit
    time.Sleep(100 * time.Millisecond)

    log.Println("Integration working!")
}
```

**Run it**:
```bash
go run main.go
```

**Output**:
```
[Gox Embedded] Initialized with agent deployment support
Event: map[message:Hello from Alfa!]
Integration working!
[Gox Embedded] Shut down
```

---

## What Works Right Now

✅ Initialize orchestrator
✅ Subscribe to events
✅ Publish events
✅ Multi-project event isolation
✅ Configuration loading

⏳ Agent deployment (needs external Gox)
⏳ Agent → Alfa events (Phase 3)

---

## Resources

- **Full Docs**: `docs/INTEGRATION-READINESS.md`
- **API Reference**: `pkg/orchestrator/README.md`
- **Examples**: `test/integration/alfa_integration_test.go`
- **Test Coverage**: `docs/test-coverage-summary.md`

---

**Ready?** Start with Step 1! 🚀
