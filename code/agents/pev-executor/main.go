package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tenzoki/agen/atomic/tools"
	"github.com/tenzoki/agen/atomic/vfs"
	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// PEVExecutor executes plan steps using available tools
type PEVExecutor struct {
	agent.DefaultAgentRunner
	model         string
	toolsEnabled  bool
	dispatcher    *tools.Dispatcher
	vfs           *vfs.VFS
	baseAgent     *agent.BaseAgent
	executedPlans map[string]bool // Track executed plan IDs for idempotency
}

// ExecuteTask from Coordinator
type ExecuteTask struct {
	RequestID string        `json:"request_id"`
	PlanID    string        `json:"plan_id"`
	Plan      ExecutionPlan `json:"plan"`
}

// ExecutionPlan structure
type ExecutionPlan struct {
	ID            string                   `json:"id"`
	RequestID     string                   `json:"request_id"`
	Type          string                   `json:"type"`
	Goal          string                   `json:"goal"`
	TargetContext string                   `json:"target_context"`
	Steps         []PlanStep               `json:"steps"`
}

// PlanStep defines a single step
type PlanStep struct {
	ID              string                 `json:"id"`
	Phase           string                 `json:"phase"`
	Action          string                 `json:"action"`
	Params          map[string]interface{} `json:"params"`
	DependsOn       []string               `json:"depends_on"`
	SuccessCriteria string                 `json:"success_criteria"`
}

// ExecutionResults to Coordinator
type ExecutionResults struct {
	RequestID      string           `json:"request_id"`
	PlanID         string           `json:"plan_id"`
	Type           string           `json:"type"`
	StepResults    []StepResult     `json:"step_results"`
	AllSuccess     bool             `json:"all_success"`
	ExecutionTime  time.Duration    `json:"execution_time"`
}

// StepResult captures the result of executing one plan step
type StepResult struct {
	StepID    string                 `json:"step_id"`
	Action    string                 `json:"action"`
	Params    map[string]interface{} `json:"params"`
	Success   bool                   `json:"success"`
	Output    interface{}            `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
}

func NewPEVExecutor() *PEVExecutor {
	return &PEVExecutor{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		model:              "gpt-5-mini",
		toolsEnabled:       true,
	}
}

func (e *PEVExecutor) Init(base *agent.BaseAgent) error {
	e.baseAgent = base
	e.executedPlans = make(map[string]bool)

	// Load config
	if model := base.GetConfigString("model", ""); model != "" {
		e.model = model
	}
	e.toolsEnabled = base.GetConfigBool("tools_enabled", true)

	// Get VFS root path - prefer CELLORG_DATA_ROOT (set by cellorg for project isolation)
	vfsRoot := os.Getenv("CELLORG_DATA_ROOT")
	if vfsRoot == "" {
		// Fall back to config value (for standalone mode or self-modification)
		vfsRoot = base.GetConfigString("vfs_root", ".")
	}

	// Create VFS for the executor (readOnly=false for write operations)
	var err error
	e.vfs, err = vfs.NewVFS(vfsRoot, false)
	if err != nil {
		return fmt.Errorf("failed to create VFS: %w", err)
	}

	// Initialize tool dispatcher
	e.dispatcher = tools.NewDispatcher(e.vfs)

	base.LogInfo("PEV Executor initialized (model: %s, tools: %v, vfs_root: %s)",
		e.model, e.toolsEnabled, vfsRoot)

	return nil
}

func (e *PEVExecutor) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "execute_task" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	var task ExecuteTask
	if err := unmarshalPayload(msg.Payload, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execute task: %w", err)
	}

	// Check if this plan was already executed (idempotency)
	if e.executedPlans[task.PlanID] {
		base.LogInfo("Plan %s already executed, ignoring duplicate", task.PlanID)
		return nil, nil
	}

	base.LogInfo("Executing plan %s with %d steps", task.PlanID, len(task.Plan.Steps))

	// Debug: Log plan details
	if len(task.Plan.Steps) == 0 {
		base.LogError("Plan has ZERO steps! This will result in no execution.")
	} else {
		base.LogInfo("Plan steps:")
		for i, step := range task.Plan.Steps {
			base.LogInfo("  Step %d: id=%s, action=%s, phase=%s", i+1, step.ID, step.Action, step.Phase)
		}
	}

	startTime := time.Now()

	// Execute all steps sequentially (Phase 1: hardcoded simulation)
	// Phase 4 will implement actual tool execution
	stepResults := e.executeSteps(task.Plan.Steps, base)

	// Debug: Log results
	base.LogInfo("executeSteps returned %d results (expected %d)", len(stepResults), len(task.Plan.Steps))

	// Debug: Write to file for inspection
	if f, err := os.OpenFile("/tmp/pev-logs/executor.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "\n=== Received Task ===\n")
		fmt.Fprintf(f, "Plan ID: %s\n", task.PlanID)
		fmt.Fprintf(f, "Steps count in task.Plan: %d\n", len(task.Plan.Steps))
		for i, step := range task.Plan.Steps {
			fmt.Fprintf(f, "  Step %d: id=%s, action=%s\n", i+1, step.ID, step.Action)
		}
		fmt.Fprintf(f, "Results count: %d\n", len(stepResults))
		for i, result := range stepResults {
			fmt.Fprintf(f, "  Result %d: step_id=%s, success=%v\n", i+1, result.StepID, result.Success)
		}
		f.Close()
	}

	// Check if all steps succeeded
	allSuccess := true
	for _, result := range stepResults {
		if !result.Success {
			allSuccess = false
			break
		}
	}

	results := ExecutionResults{
		RequestID:     task.RequestID,
		PlanID:        task.PlanID,
		Type:          "execution_results",
		StepResults:   stepResults,
		AllSuccess:    allSuccess,
		ExecutionTime: time.Since(startTime),
	}

	base.LogInfo("Execution complete: %d/%d steps succeeded",
		e.countSuccessful(stepResults), len(stepResults))

	// Mark plan as executed
	e.executedPlans[task.PlanID] = true

	return e.createMessage("execution_results", results), nil
}

func (e *PEVExecutor) Cleanup(base *agent.BaseAgent) {
	base.LogInfo("PEV Executor cleanup")
}

// executeSteps runs all plan steps sequentially
// Phase 1: Hardcoded simulation
// Phase 4: Will connect to alfa's tool dispatcher
func (e *PEVExecutor) executeSteps(steps []PlanStep, base *agent.BaseAgent) []StepResult {
	results := make([]StepResult, 0, len(steps))
	executedSteps := make(map[string]bool)

	for _, step := range steps {
		// Check dependencies
		if !e.dependenciesMet(step.DependsOn, executedSteps) {
			base.LogError("Dependencies not met for step %s", step.ID)
			results = append(results, StepResult{
				StepID:  step.ID,
				Action:  step.Action,
				Params:  step.Params,
				Success: false,
				Error:   "dependencies not met",
			})
			continue
		}

		// Execute step (hardcoded for Phase 1)
		result := e.executeStep(step, base)
		results = append(results, result)

		// Mark as executed if successful
		if result.Success {
			executedSteps[step.ID] = true
		}
	}

	return results
}

// executeStep executes a single plan step using real tools
func (e *PEVExecutor) executeStep(step PlanStep, base *agent.BaseAgent) StepResult {
	startTime := time.Now()

	base.LogDebug("Executing step %s: %s", step.ID, step.Action)

	// If tools are disabled, simulate execution
	if !e.toolsEnabled {
		return e.simulateStep(step, base, startTime)
	}

	// Map plan action to tool action
	var toolResult tools.Result

	switch step.Action {
	case "search":
		toolResult = e.dispatcher.Execute(context.Background(), tools.Action{
			Type:   "search",
			Params: step.Params,
		})

	case "read_file":
		toolResult = e.dispatcher.Execute(context.Background(), tools.Action{
			Type:   "read_file",
			Params: step.Params,
		})

	case "write_file":
		toolResult = e.dispatcher.Execute(context.Background(), tools.Action{
			Type:   "write_file",
			Params: step.Params,
		})

	case "patch":
		// Patch is implemented as read → modify → write
		toolResult = e.executePatch(step, base)

	case "run_command":
		toolResult = e.dispatcher.Execute(context.Background(), tools.Action{
			Type:   "run_command",
			Params: step.Params,
		})

	case "run_tests":
		toolResult = e.dispatcher.Execute(context.Background(), tools.Action{
			Type:   "run_tests",
			Params: step.Params,
		})

	default:
		toolResult = tools.Result{
			Success: false,
			Message: fmt.Sprintf("unsupported action: %s", step.Action),
		}
	}

	// Convert tool result to step result
	result := StepResult{
		StepID:   step.ID,
		Action:   step.Action,
		Params:   step.Params,
		Success:  toolResult.Success,
		Output:   toolResult.Output,
		Duration: time.Since(startTime),
	}

	if !toolResult.Success {
		result.Error = toolResult.Message
		base.LogError("Step %s failed: %s", step.ID, toolResult.Message)
	} else {
		base.LogDebug("Step %s completed successfully", step.ID)
	}

	return result
}

// simulateStep simulates execution when tools are disabled
func (e *PEVExecutor) simulateStep(step PlanStep, base *agent.BaseAgent, startTime time.Time) StepResult {
	time.Sleep(100 * time.Millisecond)

	result := StepResult{
		StepID:   step.ID,
		Action:   step.Action,
		Params:   step.Params,
		Success:  true,
		Output:   fmt.Sprintf("Simulated output for %s", step.Action),
		Duration: time.Since(startTime),
	}

	// Simulate occasional failures for testing
	if step.Action == "run_tests" && time.Now().UnixNano()%10 < 3 {
		result.Success = false
		result.Error = "simulated test failure"
	}

	return result
}

// executePatch applies patch operations to a file
func (e *PEVExecutor) executePatch(step PlanStep, base *agent.BaseAgent) tools.Result {
	// Extract file path
	filePath, ok := step.Params["file"].(string)
	if !ok {
		return tools.Result{
			Success: false,
			Message: "missing 'file' parameter in patch",
		}
	}

	// Extract operations
	opsInterface, ok := step.Params["operations"]
	if !ok {
		return tools.Result{
			Success: false,
			Message: "missing 'operations' parameter in patch",
		}
	}

	// Convert operations to proper format
	var operations []map[string]interface{}
	switch ops := opsInterface.(type) {
	case []interface{}:
		for _, op := range ops {
			if opMap, ok := op.(map[string]interface{}); ok {
				operations = append(operations, opMap)
			}
		}
	case []map[string]interface{}:
		operations = ops
	default:
		return tools.Result{
			Success: false,
			Message: fmt.Sprintf("invalid operations format: %T", opsInterface),
		}
	}

	// Read current file content
	content, err := e.vfs.ReadString(filePath)
	if err != nil {
		return tools.Result{
			Success: false,
			Message: fmt.Sprintf("failed to read file for patch: %v", err),
		}
	}

	// Split into lines for line-based operations
	lines := strings.Split(content, "\n")

	// Apply operations
	for _, op := range operations {
		opType, _ := op["type"].(string)

		// Extract line number (handle both int and float64 from JSON)
		line := 0
		switch v := op["line"].(type) {
		case float64:
			line = int(v)
		case int:
			line = v
		case int64:
			line = int(v)
		}

		contentToApply, _ := op["content"].(string)

		switch opType {
		case "insert":
			// Insert line at position (1-indexed)
			if line <= 0 || line > len(lines)+1 {
				return tools.Result{
					Success: false,
					Message: fmt.Sprintf("invalid line number for insert: %d", line),
				}
			}
			// Convert to 0-indexed
			idx := line - 1
			lines = append(lines[:idx], append([]string{contentToApply}, lines[idx:]...)...)

		case "replace":
			// Replace line at position (1-indexed)
			if line <= 0 || line > len(lines) {
				return tools.Result{
					Success: false,
					Message: fmt.Sprintf("invalid line number for replace: %d", line),
				}
			}
			lines[line-1] = contentToApply

		case "delete":
			// Delete line at position (1-indexed)
			if line <= 0 || line > len(lines) {
				return tools.Result{
					Success: false,
					Message: fmt.Sprintf("invalid line number for delete: %d", line),
				}
			}
			lines = append(lines[:line-1], lines[line:]...)

		default:
			return tools.Result{
				Success: false,
				Message: fmt.Sprintf("unknown patch operation type: %s", opType),
			}
		}
	}

	// Write back modified content
	modifiedContent := strings.Join(lines, "\n")
	if err := e.vfs.WriteString(modifiedContent, filePath); err != nil {
		return tools.Result{
			Success: false,
			Message: fmt.Sprintf("failed to write patched file: %v", err),
		}
	}

	base.LogDebug("Applied %d patch operations to %s", len(operations), filePath)

	return tools.Result{
		Success: true,
		Message: fmt.Sprintf("Applied %d patch operations to %s", len(operations), filePath),
		Output:  fmt.Sprintf("Patched %s with %d operations", filePath, len(operations)),
	}
}

// dependenciesMet checks if all dependencies have been successfully executed
func (e *PEVExecutor) dependenciesMet(dependencies []string, executed map[string]bool) bool {
	for _, dep := range dependencies {
		if !executed[dep] {
			return false
		}
	}
	return true
}

// countSuccessful counts how many steps succeeded
func (e *PEVExecutor) countSuccessful(results []StepResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}

// Helper functions

func (e *PEVExecutor) createMessage(msgType string, payload interface{}) *client.BrokerMessage {
	return &client.BrokerMessage{
		ID:        fmt.Sprintf("%s_%d", msgType, time.Now().UnixNano()),
		Type:      msgType,
		Target:    msgType,
		Payload:   payload,
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

func unmarshalPayload(payload interface{}, target interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func main() {
	executor := NewPEVExecutor()
	if err := agent.Run(executor, "pev-executor"); err != nil {
		panic(err)
	}
}
