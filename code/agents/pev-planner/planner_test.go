package main

import (
	"context"
	"testing"
	"time"

	"github.com/tenzoki/agen/atomic/ai"
	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// MockLLM implements ai.LLM for testing
type MockLLM struct {
	model    string
	response string
}

func (m *MockLLM) Chat(ctx context.Context, messages []ai.Message) (*ai.Response, error) {
	// Return a mock JSON plan
	if m.response == "" {
		m.response = `{
  "goal": "Add warning triangle when self_modify=true",
  "target_context": "framework",
  "steps": [
    {
      "id": "step-1",
      "phase": "discovery",
      "action": "search",
      "params": {"query": "getUserInput", "pattern": "*.go"},
      "depends_on": [],
      "success_criteria": "Find prompt rendering code"
    },
    {
      "id": "step-2",
      "phase": "analysis",
      "action": "read_file",
      "params": {"path": "code/alfa/internal/orchestrator/orchestrator.go"},
      "depends_on": ["step-1"],
      "success_criteria": "Understand current implementation"
    },
    {
      "id": "step-3",
      "phase": "implementation",
      "action": "patch",
      "params": {
        "file": "code/alfa/internal/orchestrator/orchestrator.go",
        "operations": [{"type": "insert", "line": 217, "content": "fmt.Print(\"⚠️  \")"}]
      },
      "depends_on": ["step-2"],
      "success_criteria": "Warning icon added to prompt"
    },
    {
      "id": "step-4",
      "phase": "validation",
      "action": "run_tests",
      "params": {"pattern": "./code/alfa/..."},
      "depends_on": ["step-3"],
      "success_criteria": "All tests pass"
    }
  ]
}`
	}

	return &ai.Response{
		Content:      m.response,
		Model:        m.model,
		StopReason:   "end_turn",
		Usage:        ai.Usage{InputTokens: 500, OutputTokens: 300, TotalTokens: 800},
		FinishTime:   time.Now(),
		ResponseTime: 100 * time.Millisecond,
	}, nil
}

func (m *MockLLM) ChatStream(ctx context.Context, messages []ai.Message) (<-chan string, <-chan error) {
	// Not implemented for testing
	return nil, nil
}

func (m *MockLLM) Model() string {
	return m.model
}

func (m *MockLLM) Provider() string {
	return "mock"
}

func TestPEVPlannerAIPlan(t *testing.T) {
	planner := NewPEVPlanner()
	planner.llm = &MockLLM{model: "gpt-5"}

	mockBase := &agent.BaseAgent{}

	// Create a plan request
	planRequest := PlanRequest{
		RequestID:   "test-001",
		UserRequest: "Add warning triangle when self_modify=true",
		Context: map[string]interface{}{
			"target_vfs": "framework",
		},
		Iteration: 1,
	}

	// Generate plan using AI
	plan, err := planner.createAIPlan(planRequest, mockBase)
	if err != nil {
		t.Fatalf("Failed to create AI plan: %v", err)
	}

	// Validate plan
	if plan.RequestID != "test-001" {
		t.Errorf("Expected request_id test-001, got %s", plan.RequestID)
	}

	if plan.Goal == "" {
		t.Error("Plan goal is empty")
	}

	if len(plan.Steps) == 0 {
		t.Error("Plan has no steps")
	}

	if len(plan.Steps) != 4 {
		t.Errorf("Expected 4 steps, got %d", len(plan.Steps))
	}

	// Check phases are present
	phases := make(map[string]bool)
	for _, step := range plan.Steps {
		phases[step.Phase] = true
	}

	expectedPhases := []string{"discovery", "analysis", "implementation", "validation"}
	for _, phase := range expectedPhases {
		if !phases[phase] {
			t.Errorf("Missing phase: %s", phase)
		}
	}

	t.Logf("✓ AI plan generated successfully with %d steps", len(plan.Steps))
	t.Logf("  Goal: %s", plan.Goal)
	t.Logf("  Target: %s", plan.TargetContext)
}

func TestPEVPlannerFallback(t *testing.T) {
	planner := NewPEVPlanner()
	// Don't set LLM - will fail and use hardcoded fallback

	mockBase := &agent.BaseAgent{}

	// Create plan request
	planRequest := PlanRequest{
		RequestID:   "test-002",
		UserRequest: "Test fallback plan",
		Context: map[string]interface{}{
			"target_vfs": "framework",
		},
		Iteration: 1,
	}

	// This should fail and use hardcoded fallback
	plan := planner.createHardcodedPlan(planRequest, mockBase)

	// Validate fallback plan
	if plan.RequestID != "test-002" {
		t.Errorf("Expected request_id test-002, got %s", plan.RequestID)
	}

	if len(plan.Steps) == 0 {
		t.Error("Fallback plan has no steps")
	}

	t.Logf("✓ Fallback plan generated with %d steps", len(plan.Steps))
}

func TestPEVPlannerProcessMessage(t *testing.T) {
	planner := NewPEVPlanner()
	planner.llm = &MockLLM{model: "gpt-5"}

	mockBase := &agent.BaseAgent{}

	// Create broker message
	msg := &client.BrokerMessage{
		ID:   "msg-001",
		Type: "plan_request",
		Payload: map[string]interface{}{
			"request_id":   "test-003",
			"user_request": "Add logging to main function",
			"context": map[string]interface{}{
				"target_vfs": "project",
			},
			"iteration": 1,
		},
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// Process message
	response, err := planner.ProcessMessage(msg, mockBase)
	if err != nil {
		t.Fatalf("Failed to process message: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response message")
	}

	if response.Type != "execution_plan" {
		t.Errorf("Expected execution_plan, got %s", response.Type)
	}

	// Check payload
	plan, ok := response.Payload.(ExecutionPlan)
	if !ok {
		t.Fatal("Response payload is not ExecutionPlan")
	}

	if plan.RequestID != "test-003" {
		t.Errorf("Expected request_id test-003, got %s", plan.RequestID)
	}

	t.Logf("✓ Message processed successfully")
	t.Logf("  Plan ID: %s", plan.ID)
	t.Logf("  Steps: %d", len(plan.Steps))
}

func TestPEVPlannerReplan(t *testing.T) {
	planner := NewPEVPlanner()
	planner.llm = &MockLLM{model: "claude-sonnet-4-5-20250929"}

	mockBase := &agent.BaseAgent{}

	// Create re-plan request (iteration 2)
	planRequest := PlanRequest{
		RequestID:   "test-004",
		UserRequest: "Fix compilation error",
		Context: map[string]interface{}{
			"target_vfs": "framework",
		},
		Iteration:      2,
		PreviousPlan:   "plan-001",
		PreviousIssues: []string{"Tests failed: syntax error", "Missing import statement"},
	}

	// Generate plan
	plan, err := planner.createAIPlan(planRequest, mockBase)
	if err != nil {
		t.Fatalf("Failed to create replan: %v", err)
	}

	// Validate replan
	if plan.RequestID != "test-004" {
		t.Errorf("Expected request_id test-004, got %s", plan.RequestID)
	}

	if len(plan.Steps) == 0 {
		t.Error("Replan has no steps")
	}

	t.Logf("✓ Replan generated successfully (iteration %d)", planRequest.Iteration)
	t.Logf("  Previous issues: %v", planRequest.PreviousIssues)
	t.Logf("  New plan steps: %d", len(plan.Steps))
}
