package main

import (
	"testing"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

func TestPEVCoordinatorFlow(t *testing.T) {
	coordinator := NewPEVCoordinator()

	// Mock BaseAgent (minimal setup for testing)
	mockBase := &agent.BaseAgent{}

	// Initialize coordinator
	if err := coordinator.Init(mockBase); err != nil {
		t.Fatalf("Failed to initialize coordinator: %v", err)
	}

	// Test 1: Handle user request
	t.Run("HandleUserRequest", func(t *testing.T) {
		userReq := &client.BrokerMessage{
			ID:   "test-req-001",
			Type: "user_request",
			Payload: map[string]interface{}{
				"id":      "test-req-001",
				"type":    "user_request",
				"content": "modify code to add warning icon",
				"context": map[string]interface{}{
					"target_vfs": "framework",
				},
			},
			Meta:      make(map[string]interface{}),
			Timestamp: time.Now(),
		}

		result, err := coordinator.ProcessMessage(userReq, mockBase)
		if err != nil {
			t.Errorf("Failed to process user request: %v", err)
		}
		if result == nil {
			t.Error("Expected plan_request message")
		}
		if result.Type != "plan_request" {
			t.Errorf("Expected plan_request, got %s", result.Type)
		}

		t.Logf("✓ User request processed, plan_request generated")
	})

	// Test 2: Handle execution plan
	t.Run("HandleExecutionPlan", func(t *testing.T) {
		// First create a request state
		userReq := &client.BrokerMessage{
			ID:   "test-req-002",
			Type: "user_request",
			Payload: map[string]interface{}{
				"id":      "test-req-002",
				"type":    "user_request",
				"content": "test request",
				"context": map[string]interface{}{},
			},
			Meta:      make(map[string]interface{}),
			Timestamp: time.Now(),
		}
		coordinator.ProcessMessage(userReq, mockBase)

		// Now send execution plan
		plan := ExecutionPlan{
			ID:        "plan-001",
			RequestID: "test-req-002",
			Type:      "execution_plan",
			Goal:      "test goal",
			Steps: []map[string]interface{}{
				{"id": "step-1", "action": "search"},
				{"id": "step-2", "action": "read_file"},
			},
		}

		planMsg := &client.BrokerMessage{
			ID:      "plan-msg-001",
			Type:    "execution_plan",
			Payload: plan,
			Meta:    make(map[string]interface{}),
		}

		result, err := coordinator.ProcessMessage(planMsg, mockBase)
		if err != nil {
			t.Errorf("Failed to process execution plan: %v", err)
		}
		if result == nil {
			t.Error("Expected execute_task message")
		}
		if result.Type != "execute_task" {
			t.Errorf("Expected execute_task, got %s", result.Type)
		}

		t.Logf("✓ Execution plan processed, execute_task generated")
	})

	// Test 3: Handle verification report - goal achieved
	t.Run("HandleVerificationReportSuccess", func(t *testing.T) {
		// Setup request state
		coordinator.activeRequests["test-req-003"] = &RequestState{
			RequestID:    "test-req-003",
			UserRequest:  "test",
			Iteration:    1,
			CurrentPhase: "verifying",
		}

		report := VerificationReport{
			ID:           "verify-001",
			RequestID:    "test-req-003",
			Type:         "verification_report",
			GoalAchieved: true,
			Issues:       []Issue{},
			NextActions:  []NextAction{},
		}

		reportMsg := &client.BrokerMessage{
			ID:      "report-msg-001",
			Type:    "verification_report",
			Payload: report,
			Meta:    make(map[string]interface{}),
		}

		result, err := coordinator.ProcessMessage(reportMsg, mockBase)
		if err != nil {
			t.Errorf("Failed to process verification report: %v", err)
		}
		if result == nil {
			t.Error("Expected user_response message")
		}
		if result.Type != "user_response" {
			t.Errorf("Expected user_response, got %s", result.Type)
		}

		// Check response payload
		payload := result.Payload.(map[string]interface{})
		if !payload["goal_achieved"].(bool) {
			t.Error("Expected goal_achieved to be true")
		}

		t.Logf("✓ Verification report processed, success response generated")
	})

	// Test 4: Handle verification report - re-plan needed
	t.Run("HandleVerificationReportReplan", func(t *testing.T) {
		coordinator.activeRequests["test-req-004"] = &RequestState{
			RequestID:    "test-req-004",
			UserRequest:  "test",
			Iteration:    1,
			CurrentPhase: "verifying",
			Context:      make(map[string]interface{}),
		}

		report := VerificationReport{
			ID:           "verify-002",
			RequestID:    "test-req-004",
			Type:         "verification_report",
			GoalAchieved: false,
			Issues: []Issue{
				{StepID: "step-3", Issue: "Tests failed", Severity: "critical"},
			},
			NextActions: []NextAction{
				{Type: "fix", Description: "Fix test failures", Priority: "high"},
			},
		}

		reportMsg := &client.BrokerMessage{
			ID:      "report-msg-002",
			Type:    "verification_report",
			Payload: report,
			Meta:    make(map[string]interface{}),
		}

		result, err := coordinator.ProcessMessage(reportMsg, mockBase)
		if err != nil {
			t.Errorf("Failed to process verification report: %v", err)
		}
		if result.Type != "plan_request" {
			t.Errorf("Expected plan_request for re-planning, got %s", result.Type)
		}

		// Check iteration incremented
		state := coordinator.activeRequests["test-req-004"]
		if state.Iteration != 2 {
			t.Errorf("Expected iteration 2, got %d", state.Iteration)
		}

		t.Logf("✓ Re-planning triggered, iteration incremented")
	})

	coordinator.Cleanup(mockBase)
}

func TestPEVCoordinatorMaxIterations(t *testing.T) {
	coordinator := NewPEVCoordinator()
	mockBase := &agent.BaseAgent{}
	coordinator.Init(mockBase)

	// Set max iterations AFTER Init (which loads config)
	coordinator.maxIterations = 3

	// Setup request at max iterations
	coordinator.activeRequests["test-req-max"] = &RequestState{
		RequestID:    "test-req-max",
		UserRequest:  "test",
		Iteration:    3, // At max
		CurrentPhase: "verifying",
		Context:      make(map[string]interface{}),
	}

	report := VerificationReport{
		ID:           "verify-max",
		RequestID:    "test-req-max",
		Type:         "verification_report",
		GoalAchieved: false,
		Issues: []Issue{
			{StepID: "step-2", Issue: "Still failing", Severity: "high"},
		},
		NextActions: []NextAction{
			{Type: "retry", Description: "Try different approach", Priority: "high"},
		},
	}

	reportMsg := &client.BrokerMessage{
		ID:      "report-msg-max",
		Type:    "verification_report",
		Payload: report,
		Meta:    make(map[string]interface{}),
	}

	result, err := coordinator.ProcessMessage(reportMsg, mockBase)
	if err != nil {
		t.Errorf("Failed to process verification report: %v", err)
	}

	if result.Type != "user_response" {
		t.Errorf("Expected user_response (failure), got %s", result.Type)
	}

	payload := result.Payload.(map[string]interface{})
	if payload["status"].(string) != "failed" {
		t.Errorf("Expected status=failed, got %s", payload["status"])
	}

	t.Logf("✓ Max iterations enforced, failure response generated")
}
