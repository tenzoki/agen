package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// TestExecutorInit verifies executor initialization
func TestExecutorInit(t *testing.T) {
	executor := NewPEVExecutor()

	// Create temp directory for VFS
	tmpDir := t.TempDir()

	// Create mock agent with config
	mockBase := &agent.BaseAgent{}
	// Set the config map directly
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	err := executor.Init(mockBase)
	if err != nil {
		t.Fatalf("Failed to initialize executor: %v", err)
	}

	if executor.vfs == nil {
		t.Error("VFS not initialized")
	}

	if executor.dispatcher == nil {
		t.Error("Dispatcher not initialized")
	}

	t.Logf("✓ Executor initialized successfully")
}

// TestExecutorReadFile tests read_file action
func TestExecutorReadFile(t *testing.T) {
	executor := NewPEVExecutor()

	// Create temp directory with test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, world!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Create plan with read_file step
	plan := ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:     "step-1",
				Action: "read_file",
				Params: map[string]interface{}{
					"path": "test.txt",
				},
			},
		},
	}

	// Execute steps
	results := executor.executeSteps(plan.Steps, mockBase)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Step failed: %s", result.Error)
	}

	if output, ok := result.Output.(string); ok {
		if output != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, output)
		}
	}

	t.Logf("✓ Read file successfully: %d bytes", len(testContent))
}

// TestExecutorWriteFile tests write_file action
func TestExecutorWriteFile(t *testing.T) {
	executor := NewPEVExecutor()

	tmpDir := t.TempDir()

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	testContent := "New file content"

	// Create plan with write_file step
	plan := ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:     "step-1",
				Action: "write_file",
				Params: map[string]interface{}{
					"path":    "output.txt",
					"content": testContent,
				},
			},
		},
	}

	// Execute steps
	results := executor.executeSteps(plan.Steps, mockBase)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Step failed: %s", result.Error)
	}

	// Verify file was created
	content, err := os.ReadFile(filepath.Join(tmpDir, "output.txt"))
	if err != nil {
		t.Errorf("Failed to read created file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
	}

	t.Logf("✓ Wrote file successfully: %d bytes", len(testContent))
}

// TestExecutorPatch tests patch action
func TestExecutorPatch(t *testing.T) {
	executor := NewPEVExecutor()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "code.go")

	// Create test file with multiple lines
	originalContent := `package main

func main() {
	println("Hello")
}
`

	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Create plan with patch step (insert line)
	plan := ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:     "step-1",
				Action: "patch",
				Params: map[string]interface{}{
					"file": "code.go",
					"operations": []interface{}{
						map[string]interface{}{
							"type":    "insert",
							"line":    4,
							"content": "\tprintln(\"World\")",
						},
					},
				},
			},
		},
	}

	// Execute steps
	results := executor.executeSteps(plan.Steps, mockBase)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Patch failed: %s", result.Error)
	}

	// Verify file was patched
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read patched file: %v", err)
	}

	expected := `package main

func main() {
	println("World")
	println("Hello")
}
`

	if string(content) != expected {
		t.Errorf("Patched content mismatch\nExpected:\n%s\nGot:\n%s", expected, string(content))
	}

	t.Logf("✓ Patched file successfully")
}

// TestExecutorSearch tests search action
func TestExecutorSearch(t *testing.T) {
	executor := NewPEVExecutor()

	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.go":    "package main\nfunc main() {}",
		"helper.go":  "package main\nfunc helper() {}",
		"README.md":  "# Project",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Create plan with search step
	plan := ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:     "step-1",
				Action: "search",
				Params: map[string]interface{}{
					"query":   "package main",
					"pattern": "*.go",
				},
			},
		},
	}

	// Execute steps
	results := executor.executeSteps(plan.Steps, mockBase)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Search failed: %s", result.Error)
	}

	// Verify matches found
	if matches, ok := result.Output.([]string); ok {
		if len(matches) != 2 {
			t.Errorf("Expected 2 matches, got %d: %v", len(matches), matches)
		}
	}

	t.Logf("✓ Search completed successfully")
}

// TestExecutorDependencies tests dependency handling
func TestExecutorDependencies(t *testing.T) {
	executor := NewPEVExecutor()

	tmpDir := t.TempDir()

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Create plan with dependencies
	plan := ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:     "step-1",
				Action: "write_file",
				Params: map[string]interface{}{
					"path":    "data.txt",
					"content": "initial data",
				},
				DependsOn: []string{},
			},
			{
				ID:     "step-2",
				Action: "read_file",
				Params: map[string]interface{}{
					"path": "data.txt",
				},
				DependsOn: []string{"step-1"},
			},
		},
	}

	// Execute steps
	results := executor.executeSteps(plan.Steps, mockBase)

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Both should succeed
	for i, result := range results {
		if !result.Success {
			t.Errorf("Step %d failed: %s", i+1, result.Error)
		}
	}

	t.Logf("✓ Dependencies handled correctly")
}

// TestExecutorProcessMessage tests message processing
func TestExecutorProcessMessage(t *testing.T) {
	executor := NewPEVExecutor()

	tmpDir := t.TempDir()

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Create broker message with execute task
	msg := &client.BrokerMessage{
		Type: "execute_task",
		Payload: map[string]interface{}{
			"request_id": "test-001",
			"plan_id":    "plan-001",
			"plan": map[string]interface{}{
				"id":             "plan-001",
				"request_id":     "test-001",
				"type":           "execution_plan",
				"goal":           "Test execution",
				"target_context": "project",
				"steps": []interface{}{
					map[string]interface{}{
						"id":               "step-1",
						"phase":            "implementation",
						"action":           "write_file",
						"params": map[string]interface{}{
							"path":    "test.txt",
							"content": "test content",
						},
						"depends_on":       []interface{}{},
						"success_criteria": "File created",
					},
				},
			},
		},
	}

	// Process message
	response, err := executor.ProcessMessage(msg, mockBase)
	if err != nil {
		t.Fatalf("Failed to process message: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response message")
	}

	if response.Type != "execution_results" {
		t.Errorf("Expected execution_results, got %s", response.Type)
	}

	// Check payload
	results, ok := response.Payload.(ExecutionResults)
	if !ok {
		t.Fatal("Response payload is not ExecutionResults")
	}

	if results.RequestID != "test-001" {
		t.Errorf("Expected request_id test-001, got %s", results.RequestID)
	}

	if !results.AllSuccess {
		t.Error("Expected all steps to succeed")
	}

	t.Logf("✓ Message processed successfully")
	t.Logf("  Request ID: %s", results.RequestID)
	t.Logf("  Plan ID: %s", results.PlanID)
	t.Logf("  Steps: %d/%d succeeded", len(results.StepResults), len(results.StepResults))
}

// TestExecutorSimulationMode tests execution with tools disabled
func TestExecutorSimulationMode(t *testing.T) {
	executor := NewPEVExecutor()
	executor.toolsEnabled = false // Disable tools for simulation

	tmpDir := t.TempDir()

	mockBase := &agent.BaseAgent{}
	if mockBase.Config == nil {
		mockBase.Config = make(map[string]interface{})
	}
	mockBase.Config["vfs_root"] = tmpDir
	mockBase.Config["tools_enabled"] = false

	if err := executor.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}

	// Create plan with various actions
	plan := ExecutionPlan{
		Steps: []PlanStep{
			{ID: "step-1", Action: "search", Params: map[string]interface{}{}},
			{ID: "step-2", Action: "read_file", Params: map[string]interface{}{}},
			{ID: "step-3", Action: "patch", Params: map[string]interface{}{}},
		},
	}

	// Execute steps in simulation mode
	results := executor.executeSteps(plan.Steps, mockBase)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// All should succeed (simulated)
	for i, result := range results {
		if !result.Success {
			t.Errorf("Simulated step %d failed: %s", i+1, result.Error)
		}

		if output, ok := result.Output.(string); ok {
			if output == "" {
				t.Errorf("Expected simulated output for step %d", i+1)
			}
		}
	}

	t.Logf("✓ Simulation mode works correctly")
}
