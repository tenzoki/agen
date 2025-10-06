# AGEN Testing Patterns

**Intended Audience**: AI/LLM
**Purpose**: Testing standards and patterns for AGEN framework

---

## Test Organization

### Co-located Tests

Tests live alongside source code:
```
code/agents/ner_agent/
├── main.go
├── main_test.go          ← Test file
└── README.md
```

**Naming convention**: `<source>_test.go`

### Test Directories

For integration and e2e tests:
```
test/
├── agents/               # Agent integration tests
│   ├── ner_agent_test.go
│   └── text_chunker_test.go
├── integration/          # System integration tests
│   └── vfs_isolation_test.go
└── pkg/orchestrator/     # Orchestrator tests
    └── embedded_test.go
```

---

## Test Types

### 1. Unit Tests

Test individual functions/methods in isolation:

```go
func TestChunkText(t *testing.T) {
    input := "Hello, World!"
    chunks := ChunkText(input, 5)

    if len(chunks) != 3 {
        t.Errorf("expected 3 chunks, got %d", len(chunks))
    }
}
```

**Requirements**:
- Fast (< 1ms per test)
- No external dependencies
- No file I/O (use in-memory)
- No network calls

### 2. Integration Tests

Test component interactions:

```go
func TestAgentWithStorage(t *testing.T) {
    // Setup
    store := storage.NewMemoryStore()
    agent := NewAgent(store)

    // Test
    env := envelope.New("test", map[string]interface{}{
        "text": "sample",
    })

    err := agent.ProcessEnvelope(context.Background(), env)

    // Verify
    if err != nil {
        t.Fatalf("ProcessEnvelope failed: %v", err)
    }

    // Check storage state
    result, _ := store.Get("key")
    if result != "expected" {
        t.Errorf("unexpected result: %v", result)
    }
}
```

**Requirements**:
- Test real component interactions
- Use test doubles for external services
- Clean up resources (defer cleanup)
- Reasonable timeout (< 10s per test)

### 3. Agent Tests

Test full agent lifecycle:

```go
func TestAgentLifecycle(t *testing.T) {
    // Create test support registry
    reg := support.NewServiceRegistry()
    reg.RegisterBroker(broker.NewMemoryBroker())
    reg.RegisterStorage(storage.NewMemoryStore())

    // Initialize agent
    agent := NewMyAgent()
    err := agent.Initialize(reg)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }
    defer agent.Shutdown()

    // Process envelope
    env := envelope.New("test-topic", map[string]interface{}{
        "key": "value",
    })

    err = agent.ProcessEnvelope(context.Background(), env)
    if err != nil {
        t.Errorf("ProcessEnvelope failed: %v", err)
    }
}
```

**Requirements**:
- Test Initialize, ProcessEnvelope, Shutdown
- Use in-memory support services
- Test error conditions
- Verify agent state

---

## Test Patterns

### Table-Driven Tests

For testing multiple inputs:

```go
func TestChunkSizes(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        chunkSize int
        want      int
    }{
        {"empty", "", 10, 0},
        {"single chunk", "hello", 10, 1},
        {"multiple chunks", "hello world", 5, 3},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := len(ChunkText(tt.input, tt.chunkSize))
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

### Test Helpers

Extract common setup:

```go
func newTestAgent(t *testing.T) (*MyAgent, func()) {
    t.Helper()

    reg := support.NewServiceRegistry()
    reg.RegisterBroker(broker.NewMemoryBroker())

    agent := NewMyAgent()
    err := agent.Initialize(reg)
    if err != nil {
        t.Fatalf("failed to create test agent: %v", err)
    }

    cleanup := func() {
        agent.Shutdown()
    }

    return agent, cleanup
}

func TestWithHelper(t *testing.T) {
    agent, cleanup := newTestAgent(t)
    defer cleanup()

    // Use agent...
}
```

### Context and Timeouts

Always use context with timeout:

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := agent.ProcessEnvelope(ctx, env)
    if err != nil {
        t.Fatalf("operation failed: %v", err)
    }
}
```

---

## Coverage Requirements

### Minimum Coverage

- **Agents**: 70% coverage
- **Core framework**: 80% coverage
- **Utilities**: 60% coverage

### Check coverage:

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### What to test:

✓ **Always test:**
- Public API functions
- Error conditions
- Edge cases (empty input, nil, etc.)
- Critical business logic

⚠️ **Careful testing:**
- Private functions (test via public API when possible)
- Generated code (if critical)

✗ **Don't test:**
- Third-party libraries
- Obvious getters/setters
- Trivial code

---

## Test Data

### Location

Centralized test data:
```
testdata/
├── documents/           # Sample documents
├── json/               # JSON test files
├── images/             # Test images
└── config/             # Test configurations
```

### Usage

```go
func TestLoadDocument(t *testing.T) {
    data, err := os.ReadFile("../../testdata/documents/sample.txt")
    if err != nil {
        t.Fatalf("failed to load test data: %v", err)
    }

    // Use data...
}
```

---

## Mocking and Test Doubles

### Interface-based mocking

```go
// Define interface
type Storage interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
}

// Create mock
type MockStorage struct {
    data map[string]interface{}
}

func (m *MockStorage) Get(key string) (interface{}, error) {
    val, ok := m.data[key]
    if !ok {
        return nil, errors.New("not found")
    }
    return val, nil
}

func (m *MockStorage) Set(key string, value interface{}) error {
    m.data[key] = value
    return nil
}

// Use in test
func TestWithMock(t *testing.T) {
    mock := &MockStorage{data: make(map[string]interface{})}
    agent := NewAgent(mock)

    // Test...
}
```

---

## Running Tests

### All tests
```bash
go test ./...
```

### Specific package
```bash
go test ./code/agents/ner_agent
```

### Verbose output
```bash
go test -v ./...
```

### With race detection
```bash
go test -race ./...
```

### Parallel execution
```bash
go test -parallel 4 ./...
```

### Short tests only (skip long-running)
```bash
go test -short ./...
```

Mark long tests:
```go
func TestLongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping long test")
    }
    // ...
}
```

---

## CI/CD Integration

Tests run automatically on:
- Every commit (local check)
- Pull requests (full suite)
- Before releases (extended tests)

### Test reports

Location: `reflect/test-reports/`
- `latest-summary.txt` - Quick status
- `latest-report.md` - Detailed results

### Self-test protocol

When starting work:
1. Check `reflect/test-reports/latest-summary.txt`
2. If failures found, read `latest-report.md`
3. Fix failures before proceeding
4. Report outcome to human

---

## Best Practices

### DO

✓ Write tests before committing
✓ Test error conditions
✓ Use table-driven tests for multiple cases
✓ Clean up resources (defer cleanup)
✓ Use meaningful test names
✓ Test public API, not implementation
✓ Keep tests simple and readable
✓ Run tests frequently

### DON'T

✗ Skip writing tests
✗ Test only happy path
✗ Leave resource leaks
✗ Use sleep() for synchronization
✗ Depend on test execution order
✗ Test implementation details
✗ Write flaky tests
✗ Commit without running tests

---

## Examples

### Complete Agent Test

```go
package myagent_test

import (
    "context"
    "testing"
    "time"

    "github.com/tenzoki/agen/cellorg/internal/broker"
    "github.com/tenzoki/agen/cellorg/internal/envelope"
    "github.com/tenzoki/agen/cellorg/internal/storage"
    "github.com/tenzoki/agen/cellorg/internal/support"
)

func TestMyAgent(t *testing.T) {
    // Setup
    reg := support.NewServiceRegistry()
    reg.RegisterBroker(broker.NewMemoryBroker())
    reg.RegisterStorage(storage.NewMemoryStore())

    agent := NewMyAgent()

    // Initialize
    err := agent.Initialize(reg)
    if err != nil {
        t.Fatalf("Initialize failed: %v", err)
    }
    defer agent.Shutdown()

    // Test cases
    tests := []struct {
        name    string
        input   map[string]interface{}
        wantErr bool
    }{
        {
            name: "valid input",
            input: map[string]interface{}{
                "text": "hello world",
            },
            wantErr: false,
        },
        {
            name: "missing field",
            input: map[string]interface{}{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()

            env := envelope.New("test-topic", tt.input)
            err := agent.ProcessEnvelope(ctx, env)

            if (err != nil) != tt.wantErr {
                t.Errorf("ProcessEnvelope() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## Reference

- Go testing package: https://pkg.go.dev/testing
- Table-driven tests: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
- Test fixtures: https://github.com/go-testfixtures/testfixtures

---

**See Also**:
- [Agent Development Guide](agent-patterns.md) - Agent patterns
- [Architecture Principles](architecture.md) - Framework architecture
