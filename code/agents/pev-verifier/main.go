package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tenzoki/agen/atomic/ai"
	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// PEVVerifier validates execution results and assesses goal achievement
type PEVVerifier struct {
	agent.DefaultAgentRunner
	model            string
	strictValidation bool
	llm              ai.LLM
	baseAgent        *agent.BaseAgent
	temperature      float64
}

// VerifyRequest from Coordinator
type VerifyRequest struct {
	RequestID       string                 `json:"request_id"`
	PlanID          string                 `json:"plan_id"`
	ExecutionResults map[string]interface{} `json:"execution_results"`
	Goal            string                 `json:"goal"`
}

// StepResult from execution
type StepResult struct {
	StepID   string      `json:"step_id"`
	Action   string      `json:"action"`
	Success  bool        `json:"success"`
	Output   interface{} `json:"output"`
	Error    string      `json:"error,omitempty"`
	Duration string      `json:"duration"`
}

// VerificationReport to Coordinator
type VerificationReport struct {
	ID           string        `json:"id"`
	RequestID    string        `json:"request_id"`
	Type         string        `json:"type"`
	GoalAchieved bool          `json:"goal_achieved"`
	Issues       []Issue       `json:"issues"`
	NextActions  []NextAction  `json:"next_actions"`
	VerifiedAt   time.Time     `json:"verified_at"`
}

// Issue describes a problem found during verification
type Issue struct {
	StepID      string `json:"step_id"`
	Issue       string `json:"issue"`
	Severity    string `json:"severity"` // critical, high, medium, low
}

// NextAction suggests what to do next
type NextAction struct {
	Type        string `json:"type"`        // fix, retry, adjust
	Description string `json:"description"`
	Priority    string `json:"priority"`    // high, medium, low
}

func NewPEVVerifier() *PEVVerifier {
	return &PEVVerifier{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		model:              "claude-opus-4-20250514",
		strictValidation:   true,
		temperature:        0.3,
	}
}

func (v *PEVVerifier) Init(base *agent.BaseAgent) error {
	v.baseAgent = base

	// Load config
	if model := base.GetConfigString("model", ""); model != "" {
		v.model = model
	}
	v.strictValidation = base.GetConfigBool("strict_validation", true)

	base.LogInfo("PEV Verifier initializing with model: %s", v.model)

	// Initialize LLM client
	var err error
	v.llm, err = v.createLLMClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	base.LogInfo("PEV Verifier initialized successfully")
	return nil
}

// createLLMClient creates the appropriate LLM client based on model
func (v *PEVVerifier) createLLMClient() (ai.LLM, error) {
	// Determine provider from model name
	var provider string
	modelLower := strings.ToLower(v.model)
	if strings.Contains(modelLower, "claude") {
		provider = "anthropic"
	} else if strings.Contains(modelLower, "gpt") || strings.Contains(modelLower, "o1") {
		provider = "openai"
	} else {
		provider = "anthropic" // Default to Anthropic for verification
	}

	config := ai.Config{
		Model:       v.model,
		MaxTokens:   64000,
		Temperature: v.temperature,
		Timeout:     180 * time.Second,
		RetryCount:  3,
		RetryDelay:  1 * time.Second,
	}

	// Get API key from environment
	if provider == "anthropic" {
		config.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		if config.APIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
		}
		return ai.NewClaudeClient(config), nil
	} else {
		config.APIKey = os.Getenv("OPENAI_API_KEY")
		if config.APIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not set")
		}
		return ai.NewOpenAIClient(config), nil
	}
}

func (v *PEVVerifier) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "verify_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	var req VerifyRequest
	if err := unmarshalPayload(msg.Payload, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify request: %w", err)
	}

	base.LogInfo("Verifying execution results for request %s", req.RequestID)

	// Use LLM-based verification with fallback to heuristics
	report, err := v.verifyResultsWithLLM(req, base)
	if err != nil {
		base.LogError("LLM verification failed: %v", err)
		// Fallback to heuristic verification
		report = v.verifyResultsHeuristic(req, base)
		base.LogInfo("Using fallback heuristic verification")
	}

	base.LogInfo("Verification complete: goal_achieved=%v, issues=%d",
		report.GoalAchieved, len(report.Issues))

	return v.createMessage("verification_report", report), nil
}

func (v *PEVVerifier) Cleanup(base *agent.BaseAgent) {
	base.LogInfo("PEV Verifier cleanup")
}

// verifyResultsWithLLM uses LLM to deeply analyze execution results
func (v *PEVVerifier) verifyResultsWithLLM(req VerifyRequest, base *agent.BaseAgent) (VerificationReport, error) {
	// Build verification prompts
	systemPrompt := v.buildVerificationSystemPrompt()
	userPrompt := v.buildVerificationUserPrompt(req)

	base.LogDebug("Calling LLM for verification...")

	// Call LLM
	ctx := context.Background()
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := v.llm.Chat(ctx, messages)
	if err != nil {
		return VerificationReport{}, fmt.Errorf("LLM call failed: %w", err)
	}

	base.LogInfo("LLM response received (%d tokens, %dms)",
		response.Usage.TotalTokens, response.ResponseTime.Milliseconds())

	// Parse verification report from LLM response
	report, err := v.parseVerificationReport(response.Content, req)
	if err != nil {
		return VerificationReport{}, fmt.Errorf("failed to parse verification report: %w", err)
	}

	return report, nil
}

// verifyResultsHeuristic analyzes execution results using hardcoded heuristics
// Used as fallback when LLM fails
func (v *PEVVerifier) verifyResultsHeuristic(req VerifyRequest, base *agent.BaseAgent) VerificationReport {
	reportID := fmt.Sprintf("verify-%d", time.Now().UnixNano())

	// Extract step results
	var stepResults []StepResult
	if results, ok := req.ExecutionResults["step_results"].([]interface{}); ok {
		for _, r := range results {
			var stepResult StepResult
			if err := unmarshalPayload(r, &stepResult); err != nil {
				base.LogError("Failed to unmarshal step result: %v", err)
				continue
			}
			stepResults = append(stepResults, stepResult)
		}
	}

	base.LogDebug("Analyzing %d step results", len(stepResults))

	// Phase 1: Simple heuristic verification
	issues := v.findIssues(stepResults, base)
	nextActions := v.suggestNextActions(issues, req.Goal, base)

	// Goal is achieved if:
	// 1. All steps succeeded
	// 2. No critical issues found
	allSuccess := req.ExecutionResults["all_success"].(bool)
	hasCriticalIssues := v.hasCriticalIssues(issues)

	goalAchieved := allSuccess && !hasCriticalIssues

	return VerificationReport{
		ID:           reportID,
		RequestID:    req.RequestID,
		Type:         "verification_report",
		GoalAchieved: goalAchieved,
		Issues:       issues,
		NextActions:  nextActions,
		VerifiedAt:   time.Now(),
	}
}

// findIssues analyzes step results to find problems
func (v *PEVVerifier) findIssues(stepResults []StepResult, base *agent.BaseAgent) []Issue {
	issues := make([]Issue, 0)

	for _, result := range stepResults {
		if !result.Success {
			severity := "high"
			if result.Action == "run_tests" {
				severity = "critical"
			}

			issues = append(issues, Issue{
				StepID:   result.StepID,
				Issue:    fmt.Sprintf("Step '%s' failed: %s", result.Action, result.Error),
				Severity: severity,
			})

			base.LogDebug("Found issue in step %s: %s", result.StepID, result.Error)
		}
	}

	// Additional heuristics
	if len(stepResults) == 0 {
		issues = append(issues, Issue{
			StepID:   "overall",
			Issue:    "No steps were executed",
			Severity: "critical",
		})
	}

	return issues
}

// suggestNextActions provides recommendations based on issues found
func (v *PEVVerifier) suggestNextActions(issues []Issue, goal string, base *agent.BaseAgent) []NextAction {
	actions := make([]NextAction, 0)

	if len(issues) == 0 {
		return actions
	}

	// Group issues by type
	testFailures := 0
	patchFailures := 0
	otherFailures := 0

	for _, issue := range issues {
		if strings.Contains(issue.Issue, "test") {
			testFailures++
		} else if strings.Contains(issue.Issue, "patch") {
			patchFailures++
		} else {
			otherFailures++
		}
	}

	// Suggest actions based on issue patterns
	if testFailures > 0 {
		actions = append(actions, NextAction{
			Type:        "fix",
			Description: "Fix code to pass tests",
			Priority:    "high",
		})
	}

	if patchFailures > 0 {
		actions = append(actions, NextAction{
			Type:        "adjust",
			Description: "Adjust patch strategy or target different files",
			Priority:    "high",
		})
	}

	if otherFailures > 0 {
		actions = append(actions, NextAction{
			Type:        "retry",
			Description: "Retry failed steps with different approach",
			Priority:    "medium",
		})
	}

	base.LogDebug("Suggested %d next actions", len(actions))

	return actions
}

// buildVerificationSystemPrompt creates the system prompt for verification
func (v *PEVVerifier) buildVerificationSystemPrompt() string {
	return `You are an expert verification agent for the AGEN framework. Your job is to analyze execution results and determine if the user's goal was achieved.

**Your Task:**
1. Review the execution plan's goal
2. Analyze each step's execution results
3. Determine if the goal was achieved
4. Identify any issues or failures
5. Suggest next actions if goal not achieved

**Analysis Criteria:**
- **Success**: All critical steps succeeded AND goal criteria met
- **Partial**: Some steps succeeded but goal not fully achieved
- **Failure**: Critical steps failed OR goal not achieved

**Issue Severity Levels:**
- **critical**: Blocks goal achievement completely (e.g., test failures, compilation errors)
- **high**: Significant problems that likely prevent goal (e.g., partial failures, missing outputs)
- **medium**: Issues that might affect goal (e.g., warnings, suboptimal implementations)
- **low**: Minor issues that don't block goal (e.g., style issues, non-critical warnings)

**Next Action Types:**
- **fix**: Fix code to resolve errors (for test failures, compilation errors)
- **adjust**: Adjust approach or strategy (for wrong approach, incorrect files)
- **retry**: Retry with same approach (for transient failures)
- **continue**: Continue with additional steps (for partial completion)

**Your Response Must Be Valid JSON:**
{
  "goal_achieved": true|false,
  "reasoning": "Detailed explanation of why goal was/wasn't achieved",
  "issues": [
    {
      "step_id": "step-3",
      "issue": "Description of the problem",
      "severity": "critical|high|medium|low"
    }
  ],
  "next_actions": [
    {
      "type": "fix|adjust|retry|continue",
      "description": "What to do next",
      "priority": "high|medium|low"
    }
  ]
}

**Important:**
- Be thorough in analysis
- Consider all execution results
- If any tests failed, goal is NOT achieved
- If goal requires specific output/behavior, verify it actually happened
- Provide actionable next steps
- Focus on whether the GOAL was achieved, not just if steps succeeded`
}

// buildVerificationUserPrompt creates the user prompt with execution details
func (v *PEVVerifier) buildVerificationUserPrompt(req VerifyRequest) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("**Goal:** %s\n\n", req.Goal))
	prompt.WriteString(fmt.Sprintf("**Request ID:** %s\n", req.RequestID))
	prompt.WriteString(fmt.Sprintf("**Plan ID:** %s\n\n", req.PlanID))

	prompt.WriteString("**Execution Results:**\n\n")

	// Extract and format step results
	var stepResults []StepResult
	if results, ok := req.ExecutionResults["step_results"].([]interface{}); ok {
		for _, r := range results {
			var stepResult StepResult
			if err := unmarshalPayload(r, &stepResult); err != nil {
				continue
			}
			stepResults = append(stepResults, stepResult)
		}
	}

	if len(stepResults) == 0 {
		prompt.WriteString("⚠️  No steps were executed!\n\n")
	} else {
		for i, result := range stepResults {
			status := "✅ SUCCESS"
			if !result.Success {
				status = "❌ FAILED"
			}

			prompt.WriteString(fmt.Sprintf("%d. **Step %s** (%s) - %s\n", i+1, result.StepID, result.Action, status))
			if result.Success {
				prompt.WriteString(fmt.Sprintf("   Duration: %s\n", result.Duration))
				if result.Output != nil {
					// Truncate output if too long
					outputStr := fmt.Sprintf("%v", result.Output)
					if len(outputStr) > 500 {
						outputStr = outputStr[:500] + "... [truncated]"
					}
					prompt.WriteString(fmt.Sprintf("   Output: %s\n", outputStr))
				}
			} else {
				prompt.WriteString(fmt.Sprintf("   Error: %s\n", result.Error))
			}
			prompt.WriteString("\n")
		}
	}

	// Overall success flag
	allSuccess, _ := req.ExecutionResults["all_success"].(bool)
	prompt.WriteString(fmt.Sprintf("**All Steps Succeeded:** %v\n\n", allSuccess))

	prompt.WriteString("Analyze these results and determine if the goal was achieved. Provide your assessment as JSON.")

	return prompt.String()
}

// parseVerificationReport extracts VerificationReport from LLM response
func (v *PEVVerifier) parseVerificationReport(content string, req VerifyRequest) (VerificationReport, error) {
	// Extract JSON from response (handle markdown code blocks)
	jsonStr := content
	if strings.Contains(content, "```json") {
		parts := strings.Split(content, "```json")
		if len(parts) > 1 {
			jsonParts := strings.Split(parts[1], "```")
			if len(jsonParts) > 0 {
				jsonStr = strings.TrimSpace(jsonParts[0])
			}
		}
	} else if strings.Contains(content, "```") {
		parts := strings.Split(content, "```")
		if len(parts) > 1 {
			jsonStr = strings.TrimSpace(parts[1])
		}
	}

	// Parse JSON
	var reportData struct {
		GoalAchieved bool         `json:"goal_achieved"`
		Reasoning    string       `json:"reasoning"`
		Issues       []Issue      `json:"issues"`
		NextActions  []NextAction `json:"next_actions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &reportData); err != nil {
		return VerificationReport{}, fmt.Errorf("JSON parse error: %w\nContent: %s", err, jsonStr)
	}

	// Create verification report
	report := VerificationReport{
		ID:           fmt.Sprintf("verify-%d", time.Now().UnixNano()),
		RequestID:    req.RequestID,
		Type:         "verification_report",
		GoalAchieved: reportData.GoalAchieved,
		Issues:       reportData.Issues,
		NextActions:  reportData.NextActions,
		VerifiedAt:   time.Now(),
	}

	// Ensure issues and next actions are not nil
	if report.Issues == nil {
		report.Issues = []Issue{}
	}
	if report.NextActions == nil {
		report.NextActions = []NextAction{}
	}

	return report, nil
}

// hasCriticalIssues checks if any issues are critical
func (v *PEVVerifier) hasCriticalIssues(issues []Issue) bool {
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return true
		}
	}
	return false
}

// Helper functions

func (v *PEVVerifier) createMessage(msgType string, payload interface{}) *client.BrokerMessage {
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
	verifier := NewPEVVerifier()
	if err := agent.Run(verifier, "pev-verifier"); err != nil {
		panic(err)
	}
}
