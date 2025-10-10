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
	// Return a mock verification report
	if m.response == "" {
		m.response = `{
  "goal_achieved": true,
  "reasoning": "All steps completed successfully. The warning icon was added to the prompt rendering code and tests passed.",
  "issues": [],
  "next_actions": []
}`
	}

	return &ai.Response{
		Content:      m.response,
		Model:        m.model,
		StopReason:   "end_turn",
		Usage:        ai.Usage{InputTokens: 600, OutputTokens: 200, TotalTokens: 800},
		FinishTime:   time.Now(),
		ResponseTime: 150 * time.Millisecond,
	}, nil
}

func (m *MockLLM) ChatStream(ctx context.Context, messages []ai.Message) (<-chan string, <-chan error) {
	return nil, nil
}

func (m *MockLLM) Model() string {
	return m.model
}

func (m *MockLLM) Provider() string {
	return "mock"
}

// TestVerifierLLMVerification tests LLM-based verification
func TestVerifierLLMVerification(t *testing.T) {
	verifier := NewPEVVerifier()
	verifier.llm = &MockLLM{model: "claude-opus-4"}

	mockBase := &agent.BaseAgent{}

	// Create verify request with successful execution
	verifyReq := VerifyRequest{
		RequestID: "test-001",
		PlanID:    "plan-001",
		Goal:      "Add warning icon when self_modify=true",
		ExecutionResults: map[string]interface{}{
			"all_success": true,
			"step_results": []interface{}{
				map[string]interface{}{
					"step_id":  "step-1",
					"action":   "search",
					"success":  true,
					"output":   []string{"code/alfa/internal/orchestrator/orchestrator.go"},
					"duration": "50ms",
				},
				map[string]interface{}{
					"step_id":  "step-2",
					"action":   "read_file",
					"success":  true,
					"output":   "package main\n...",
					"duration": "5ms",
				},
				map[string]interface{}{
					"step_id":  "step-3",
					"action":   "patch",
					"success":  true,
					"output":   "Patched file",
					"duration": "10ms",
				},
				map[string]interface{}{
					"step_id":  "step-4",
					"action":   "run_tests",
					"success":  true,
					"output":   "PASS\nok  \t0.345s",
					"duration": "350ms",
				},
			},
		},
	}

	// Verify using LLM
	report, err := verifier.verifyResultsWithLLM(verifyReq, mockBase)
	if err != nil {
		t.Fatalf("Failed to verify with LLM: %v", err)
	}

	// Validate report
	if report.RequestID != "test-001" {
		t.Errorf("Expected request_id test-001, got %s", report.RequestID)
	}

	if !report.GoalAchieved {
		t.Error("Expected goal to be achieved")
	}

	if len(report.Issues) != 0 {
		t.Errorf("Expected no issues, got %d", len(report.Issues))
	}

	t.Logf("✓ LLM verification successful")
	t.Logf("  Goal Achieved: %v", report.GoalAchieved)
	t.Logf("  Issues: %d", len(report.Issues))
}

// TestVerifierFailedVerification tests verification of failed execution
func TestVerifierFailedVerification(t *testing.T) {
	verifier := NewPEVVerifier()

	// Mock LLM with failure response
	verifier.llm = &MockLLM{
		model: "claude-opus-4",
		response: `{
  "goal_achieved": false,
  "reasoning": "Step 4 (run_tests) failed with test errors. The implementation is incorrect.",
  "issues": [
    {
      "step_id": "step-4",
      "issue": "Tests failed with compilation error",
      "severity": "critical"
    }
  ],
  "next_actions": [
    {
      "type": "fix",
      "description": "Fix the compilation error in the patched code",
      "priority": "high"
    }
  ]
}`,
	}

	mockBase := &agent.BaseAgent{}

	// Create verify request with failed test
	verifyReq := VerifyRequest{
		RequestID: "test-002",
		PlanID:    "plan-002",
		Goal:      "Add warning icon when self_modify=true",
		ExecutionResults: map[string]interface{}{
			"all_success": false,
			"step_results": []interface{}{
				map[string]interface{}{
					"step_id":  "step-1",
					"action":   "search",
					"success":  true,
					"output":   []string{"file.go"},
					"duration": "50ms",
				},
				map[string]interface{}{
					"step_id":  "step-2",
					"action":   "patch",
					"success":  true,
					"output":   "Patched",
					"duration": "10ms",
				},
				map[string]interface{}{
					"step_id":  "step-4",
					"action":   "run_tests",
					"success":  false,
					"error":    "compilation error",
					"duration": "100ms",
				},
			},
		},
	}

	// Verify
	report, err := verifier.verifyResultsWithLLM(verifyReq, mockBase)
	if err != nil {
		t.Fatalf("Failed to verify: %v", err)
	}

	// Validate report
	if report.GoalAchieved {
		t.Error("Expected goal NOT to be achieved")
	}

	if len(report.Issues) == 0 {
		t.Error("Expected issues to be reported")
	}

	if len(report.NextActions) == 0 {
		t.Error("Expected next actions to be suggested")
	}

	// Check issue severity
	hasCritical := false
	for _, issue := range report.Issues {
		if issue.Severity == "critical" {
			hasCritical = true
			break
		}
	}
	if !hasCritical {
		t.Error("Expected at least one critical issue")
	}

	t.Logf("✓ Failed verification detected correctly")
	t.Logf("  Goal Achieved: %v", report.GoalAchieved)
	t.Logf("  Issues: %d", len(report.Issues))
	t.Logf("  Next Actions: %d", len(report.NextActions))
}

// TestVerifierHeuristicFallback tests fallback to heuristic verification
func TestVerifierHeuristicFallback(t *testing.T) {
	verifier := NewPEVVerifier()
	// Don't set LLM - will fail and use heuristic fallback

	mockBase := &agent.BaseAgent{}

	// Create verify request
	verifyReq := VerifyRequest{
		RequestID: "test-003",
		PlanID:    "plan-003",
		Goal:      "Test heuristic fallback",
		ExecutionResults: map[string]interface{}{
			"all_success": true,
			"step_results": []interface{}{
				map[string]interface{}{
					"step_id":  "step-1",
					"action":   "write_file",
					"success":  true,
					"output":   "File written",
					"duration": "10ms",
				},
			},
		},
	}

	// Use heuristic verification directly
	report := verifier.verifyResultsHeuristic(verifyReq, mockBase)

	// Validate report
	if report.RequestID != "test-003" {
		t.Errorf("Expected request_id test-003, got %s", report.RequestID)
	}

	// Heuristic should pass if all_success is true
	if !report.GoalAchieved {
		t.Error("Expected goal to be achieved (heuristic)")
	}

	t.Logf("✓ Heuristic verification works")
}

// TestVerifierProcessMessage tests message processing
func TestVerifierProcessMessage(t *testing.T) {
	verifier := NewPEVVerifier()
	verifier.llm = &MockLLM{model: "claude-opus-4"}

	mockBase := &agent.BaseAgent{}

	// Create broker message with verify request
	msg := &client.BrokerMessage{
		Type: "verify_request",
		Payload: map[string]interface{}{
			"request_id": "test-004",
			"plan_id":    "plan-004",
			"goal":       "Test message processing",
			"execution_results": map[string]interface{}{
				"all_success": true,
				"step_results": []interface{}{
					map[string]interface{}{
						"step_id":  "step-1",
						"action":   "write_file",
						"success":  true,
						"output":   "Done",
						"duration": "10ms",
					},
				},
			},
		},
	}

	// Process message
	response, err := verifier.ProcessMessage(msg, mockBase)
	if err != nil {
		t.Fatalf("Failed to process message: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response message")
	}

	if response.Type != "verification_report" {
		t.Errorf("Expected verification_report, got %s", response.Type)
	}

	// Check payload
	report, ok := response.Payload.(VerificationReport)
	if !ok {
		t.Fatal("Response payload is not VerificationReport")
	}

	if report.RequestID != "test-004" {
		t.Errorf("Expected request_id test-004, got %s", report.RequestID)
	}

	t.Logf("✓ Message processed successfully")
	t.Logf("  Report ID: %s", report.ID)
	t.Logf("  Goal Achieved: %v", report.GoalAchieved)
}

// TestVerifierJSONParsing tests JSON extraction from markdown
func TestVerifierJSONParsing(t *testing.T) {
	verifier := NewPEVVerifier()

	// Test with markdown-wrapped JSON
	markdownResponse := "```json\n{\n  \"goal_achieved\": true,\n  \"reasoning\": \"Test\",\n  \"issues\": [],\n  \"next_actions\": []\n}\n```"

	req := VerifyRequest{
		RequestID: "test-005",
		PlanID:    "plan-005",
		Goal:      "Test JSON parsing",
	}

	report, err := verifier.parseVerificationReport(markdownResponse, req)
	if err != nil {
		t.Fatalf("Failed to parse markdown JSON: %v", err)
	}

	if !report.GoalAchieved {
		t.Error("Expected goal_achieved to be true")
	}

	// Test with plain JSON
	plainResponse := `{"goal_achieved": false, "reasoning": "Failed", "issues": [], "next_actions": []}`

	report2, err := verifier.parseVerificationReport(plainResponse, req)
	if err != nil {
		t.Fatalf("Failed to parse plain JSON: %v", err)
	}

	if report2.GoalAchieved {
		t.Error("Expected goal_achieved to be false")
	}

	t.Logf("✓ JSON parsing works for both markdown and plain formats")
}

// TestVerifierIssueSeverity tests issue severity detection
func TestVerifierIssueSeverity(t *testing.T) {
	verifier := NewPEVVerifier()

	issues := []Issue{
		{StepID: "step-1", Issue: "Minor issue", Severity: "low"},
		{StepID: "step-2", Issue: "Important issue", Severity: "high"},
	}

	if verifier.hasCriticalIssues(issues) {
		t.Error("Expected no critical issues")
	}

	// Add critical issue
	issues = append(issues, Issue{StepID: "step-3", Issue: "Critical", Severity: "critical"})

	if !verifier.hasCriticalIssues(issues) {
		t.Error("Expected critical issue to be detected")
	}

	t.Logf("✓ Issue severity detection works")
}
