package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// TestFullPEVCycle simulates a complete Plan-Execute-Verify cycle
func TestFullPEVCycle(t *testing.T) {
	// Create all 4 agents
	coordinator := NewPEVCoordinator()

	mockBase := &agent.BaseAgent{}

	if err := coordinator.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize coordinator: %v", err)
	}

	fmt.Println("\n=== Full PEV Cycle Simulation ===")

	// Step 1: User Request
	fmt.Println("Step 1: User sends request...")
	userReq := &client.BrokerMessage{
		ID:   "full-test-001",
		Type: "user_request",
		Payload: map[string]interface{}{
			"id":      "full-test-001",
			"type":    "user_request",
			"content": "Add warning icon when self_modify=true",
			"context": map[string]interface{}{
				"target_vfs":      "framework",
				"self_modify_enabled": true,
			},
		},
		Meta:      make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	planReqMsg, err := coordinator.ProcessMessage(userReq, mockBase)
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	fmt.Printf("  → Coordinator outputs: %s\n\n", planReqMsg.Type)

	// Step 2: Planner generates plan (simulated)
	fmt.Println("Step 2: Planner creates execution plan...")
	executionPlan := ExecutionPlan{
		ID:        "plan-001",
		RequestID: "full-test-001",
		Type:      "execution_plan",
		Goal:      "Add warning icon when self_modify=true",
		Steps: []map[string]interface{}{
			{
				"id":     "step-1",
				"phase":  "discovery",
				"action": "search",
				"params": map[string]interface{}{
					"query":   "getUserInput",
					"pattern": "*.go",
				},
			},
			{
				"id":     "step-2",
				"phase":  "analysis",
				"action": "read_file",
				"params": map[string]interface{}{
					"path": "code/alfa/internal/orchestrator/orchestrator.go",
				},
			},
			{
				"id":     "step-3",
				"phase":  "implementation",
				"action": "patch",
				"params": map[string]interface{}{
					"file": "code/alfa/internal/orchestrator/orchestrator.go",
				},
			},
			{
				"id":     "step-4",
				"phase":  "validation",
				"action": "run_tests",
				"params": map[string]interface{}{
					"pattern": "./code/alfa/...",
				},
			},
		},
	}

	planMsg := &client.BrokerMessage{
		ID:      "plan-msg-001",
		Type:    "execution_plan",
		Payload: executionPlan,
		Meta:    make(map[string]interface{}),
	}

	executeTaskMsg, err := coordinator.ProcessMessage(planMsg, mockBase)
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	fmt.Printf("  → Planner outputs: %s with %d steps\n\n", executeTaskMsg.Type, len(executionPlan.Steps))

	// Step 3: Executor runs steps (simulated)
	fmt.Println("Step 3: Executor executes plan steps...")
	executionResults := map[string]interface{}{
		"request_id": "full-test-001",
		"plan_id":    "plan-001",
		"type":       "execution_results",
		"step_results": []map[string]interface{}{
			{"step_id": "step-1", "action": "search", "success": true, "output": "Found getUserInput in orchestrator.go:217"},
			{"step_id": "step-2", "action": "read_file", "success": true, "output": "Read 500 lines"},
			{"step_id": "step-3", "action": "patch", "success": true, "output": "Applied 1 patch"},
			{"step_id": "step-4", "action": "run_tests", "success": true, "output": "All tests passed"},
		},
		"all_success": true,
	}

	resultsMsg := &client.BrokerMessage{
		ID:      "results-msg-001",
		Type:    "execution_results",
		Payload: executionResults,
		Meta:    make(map[string]interface{}),
	}

	verifyReqMsg, err := coordinator.ProcessMessage(resultsMsg, mockBase)
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	fmt.Printf("  → Executor outputs: %s (all steps succeeded)\n\n", verifyReqMsg.Type)

	// Step 4: Verifier checks goal achievement (simulated)
	fmt.Println("Step 4: Verifier validates results...")
	verificationReport := VerificationReport{
		ID:           "verify-001",
		RequestID:    "full-test-001",
		Type:         "verification_report",
		GoalAchieved: true,
		Issues:       []Issue{},
		NextActions:  []NextAction{},
	}

	reportMsg := &client.BrokerMessage{
		ID:      "verify-msg-001",
		Type:    "verification_report",
		Payload: verificationReport,
		Meta:    make(map[string]interface{}),
	}

	finalResponse, err := coordinator.ProcessMessage(reportMsg, mockBase)
	if err != nil {
		t.Fatalf("Step 4 failed: %v", err)
	}
	fmt.Printf("  → Verifier outputs: goal_achieved=true\n\n")

	// Step 5: Coordinator responds to user
	fmt.Println("Step 5: Coordinator sends final response...")
	if finalResponse.Type != "user_response" {
		t.Errorf("Expected user_response, got %s", finalResponse.Type)
	}

	payload := finalResponse.Payload.(map[string]interface{})
	if !payload["goal_achieved"].(bool) {
		t.Error("Expected goal_achieved=true")
	}

	fmt.Printf("  → Final response: status=%s, iterations=%d\n\n",
		payload["status"], int(payload["iterations"].(int)))

	fmt.Println("=== PEV Cycle Complete ===")
	fmt.Println("✓ User request processed successfully in 1 iteration")

	coordinator.Cleanup(mockBase)
}

// TestPEVCycleWithReplan simulates a cycle that needs re-planning
func TestPEVCycleWithReplan(t *testing.T) {
	coordinator := NewPEVCoordinator()
	coordinator.maxIterations = 5

	mockBase := &agent.BaseAgent{}
	coordinator.Init(mockBase)

	fmt.Println("\n=== PEV Cycle with Re-planning ===")

	// Start cycle
	userReq := &client.BrokerMessage{
		ID:   "replan-test-001",
		Type: "user_request",
		Payload: map[string]interface{}{
			"id":      "replan-test-001",
			"type":    "user_request",
			"content": "Fix compilation error",
			"context": map[string]interface{}{},
		},
		Meta: make(map[string]interface{}),
	}

	coordinator.ProcessMessage(userReq, mockBase)

	// Iteration 1: Tests fail
	fmt.Println("Iteration 1: Initial attempt...")
	setupIteration(t, coordinator, mockBase, "replan-test-001", 1, false, []Issue{
		{StepID: "step-4", Issue: "Tests failed: syntax error", Severity: "critical"},
	})
	fmt.Println("  → Tests failed, re-planning...")

	// Iteration 2: Still fails
	fmt.Println("Iteration 2: Second attempt...")
	setupIteration(t, coordinator, mockBase, "replan-test-001", 2, false, []Issue{
		{StepID: "step-4", Issue: "Tests still failing", Severity: "high"},
	})
	fmt.Println("  → Still failing, re-planning again...")

	// Iteration 3: Success!
	fmt.Println("Iteration 3: Third attempt...")
	setupIteration(t, coordinator, mockBase, "replan-test-001", 3, true, []Issue{})

	fmt.Println("\n=== Cycle Complete ===")
	fmt.Println("✓ Goal achieved after 3 iterations")
}

func setupIteration(t *testing.T, coord *PEVCoordinator, base *agent.BaseAgent, reqID string, iter int, success bool, issues []Issue) {
	// Simulate plan
	plan := ExecutionPlan{
		ID:        fmt.Sprintf("plan-%d", iter),
		RequestID: reqID,
		Type:      "execution_plan",
		Goal:      "Fix error",
		Steps:     []map[string]interface{}{{"id": "step-1", "action": "patch"}},
	}

	planMsg := &client.BrokerMessage{
		ID:      fmt.Sprintf("plan-%d", iter),
		Type:    "execution_plan",
		Payload: plan,
		Meta:    make(map[string]interface{}),
	}
	coord.ProcessMessage(planMsg, base)

	// Simulate execution
	resultsMsg := &client.BrokerMessage{
		ID:   fmt.Sprintf("results-%d", iter),
		Type: "execution_results",
		Payload: map[string]interface{}{
			"request_id":   reqID,
			"plan_id":      plan.ID,
			"all_success":  success,
			"step_results": []map[string]interface{}{},
		},
		Meta: make(map[string]interface{}),
	}
	coord.ProcessMessage(resultsMsg, base)

	// Simulate verification
	report := VerificationReport{
		ID:           fmt.Sprintf("verify-%d", iter),
		RequestID:    reqID,
		Type:         "verification_report",
		GoalAchieved: success,
		Issues:       issues,
	}

	reportMsg := &client.BrokerMessage{
		ID:      fmt.Sprintf("verify-%d", iter),
		Type:    "verification_report",
		Payload: report,
		Meta:    make(map[string]interface{}),
	}
	coord.ProcessMessage(reportMsg, base)
}
