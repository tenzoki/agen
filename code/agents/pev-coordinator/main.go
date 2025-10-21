package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// PEVCoordinator orchestrates the Plan-Execute-Verify loop
type PEVCoordinator struct {
	agent.DefaultAgentRunner
	maxIterations  int
	activeRequests map[string]*RequestState
	baseAgent      *agent.BaseAgent // Store base agent for direct broker access
}

// RequestState tracks the state of a single user request through PEV cycles
type RequestState struct {
	RequestID         string
	UserRequest       string
	Context           map[string]interface{}
	Iteration         int
	CurrentPhase      string // "planning", "executing", "verifying", "complete", "failed"
	PlanID            string
	GoalAchieved      bool
	CreatedAt         time.Time
	CompletedAt       time.Time // When the request completed (for cleanup)
	LastExecutionResults map[string]interface{} // Store execution results for summary
}

// UserRequest message from alfa orchestrator
type UserRequest struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Content string                 `json:"content"`
	Context map[string]interface{} `json:"context"`
}

// PlanRequest message to Planner
type PlanRequest struct {
	RequestID   string                 `json:"request_id"`
	UserRequest string                 `json:"user_request"`
	Context     map[string]interface{} `json:"context"`
	Iteration   int                    `json:"iteration"`
	PreviousPlan string                `json:"previous_plan,omitempty"`
	PreviousIssues []string            `json:"previous_issues,omitempty"`
}

// ExecutionPlan message from Planner
type ExecutionPlan struct {
	ID            string     `json:"id"`
	RequestID     string     `json:"request_id"`
	Type          string     `json:"type"`
	Goal          string     `json:"goal"`
	TargetContext string     `json:"target_context"`
	Steps         []PlanStep `json:"steps"`
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

// VerificationReport message from Verifier
type VerificationReport struct {
	ID           string       `json:"id"`
	RequestID    string       `json:"request_id"`
	Type         string       `json:"type"`
	GoalAchieved bool         `json:"goal_achieved"`
	Issues       []Issue      `json:"issues"`
	NextActions  []NextAction `json:"next_actions"`
}

// Issue describes a problem found during verification
type Issue struct {
	StepID   string `json:"step_id"`
	Issue    string `json:"issue"`
	Severity string `json:"severity"`
}

// NextAction suggests what to do next
type NextAction struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

func NewPEVCoordinator() *PEVCoordinator {
	return &PEVCoordinator{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		maxIterations:      10,
		activeRequests:     make(map[string]*RequestState),
	}
}

func (c *PEVCoordinator) Init(base *agent.BaseAgent) error {
	c.baseAgent = base
	base.LogInfo("PEV Coordinator initialized (max iterations: %d)", c.maxIterations)

	// Load config
	if maxIter := base.GetConfigInt("max_iterations", 10); maxIter > 0 {
		c.maxIterations = maxIter
	}

	// Start cleanup goroutine for old completed requests
	go c.cleanupOldRequests()

	return nil
}

func (c *PEVCoordinator) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	base.LogDebug("Coordinator received message type: %s", msg.Type)

	switch msg.Type {
	case "user_request":
		return c.handleUserRequest(msg, base)
	case "execution_plan":
		return c.handleExecutionPlan(msg, base)
	case "execution_results":
		return c.handleExecutionResults(msg, base)
	case "verification_report":
		return c.handleVerificationReport(msg, base)
	default:
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

func (c *PEVCoordinator) Cleanup(base *agent.BaseAgent) {
	base.LogInfo("PEV Coordinator cleanup - active requests: %d", len(c.activeRequests))
}

// cleanupOldRequests periodically removes completed requests older than 10 minutes
func (c *PEVCoordinator) cleanupOldRequests() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if c.baseAgent == nil {
			continue
		}

		now := time.Now()
		cleanupCount := 0

		for id, state := range c.activeRequests {
			// Only clean up completed or failed requests
			if state.CurrentPhase != "complete" && state.CurrentPhase != "failed" {
				continue
			}

			// Check if completed more than 10 minutes ago
			if !state.CompletedAt.IsZero() && now.Sub(state.CompletedAt) > 10*time.Minute {
				delete(c.activeRequests, id)
				cleanupCount++
				c.baseAgent.LogDebug("Cleaned up old completed request: %s (completed %v ago)",
					id, now.Sub(state.CompletedAt))
			}
		}

		if cleanupCount > 0 {
			c.baseAgent.LogInfo("Cleaned up %d old completed requests", cleanupCount)
		}
	}
}

// handleUserRequest starts a new PEV cycle for a user request
func (c *PEVCoordinator) handleUserRequest(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	var req UserRequest
	if err := unmarshalPayload(msg.Payload, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user request: %w", err)
	}

	// Check if request is already being processed (idempotency)
	if _, exists := c.activeRequests[req.ID]; exists {
		base.LogInfo("Request %s already being processed, ignoring duplicate", req.ID)
		return nil, nil
	}

	base.LogInfo("Starting PEV cycle for request: %s", req.ID)

	// Create request state
	state := &RequestState{
		RequestID:    req.ID,
		UserRequest:  req.Content,
		Context:      req.Context,
		Iteration:    1,
		CurrentPhase: "planning",
		CreatedAt:    time.Now(),
	}
	c.activeRequests[req.ID] = state

	// Send plan request to Planner
	planReq := PlanRequest{
		RequestID:   req.ID,
		UserRequest: req.Content,
		Context:     req.Context,
		Iteration:   1,
	}

	base.LogInfo("Sending plan request for iteration %d", state.Iteration)

	// Create and publish message
	planMsg := c.createMessage("plan_request", planReq)
	c.publishToTopic("plan-requests", planMsg)

	return nil, nil // Already published, don't send via egress
}

// handleExecutionPlan receives a plan and triggers Executor
func (c *PEVCoordinator) handleExecutionPlan(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	var plan ExecutionPlan
	if err := unmarshalPayload(msg.Payload, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution plan: %w", err)
	}

	state, ok := c.activeRequests[plan.RequestID]
	if !ok {
		base.LogDebug("Received execution plan for unknown request %s (likely already completed), ignoring", plan.RequestID)
		return nil, nil
	}

	// Check if request already completed
	if state.CurrentPhase == "complete" || state.CurrentPhase == "failed" {
		base.LogDebug("Received execution plan for already completed request %s, ignoring", plan.RequestID)
		return nil, nil
	}

	// Check if this plan was already processed (idempotency)
	if state.PlanID == plan.ID {
		base.LogDebug("Plan %s already processed for request %s, ignoring duplicate", plan.ID, plan.RequestID)
		return nil, nil
	}

	base.LogInfo("Received execution plan with %d steps for request %s", len(plan.Steps), plan.RequestID)

	// Debug: Log plan steps
	if len(plan.Steps) == 0 {
		base.LogError("Received plan with ZERO steps from planner!")
	} else {
		base.LogInfo("\n=== Received Plan ===")
		base.LogInfo("Plan ID: %s", plan.ID)
		base.LogInfo("Steps: %d", len(plan.Steps))
		for i, step := range plan.Steps {
			base.LogInfo("  %d. [%s] %s (phase: %s)", i+1, step.ID, step.Action, step.Phase)
		}
		base.LogInfo("======================\n")
	}

	// Debug: Write to file for inspection
	if f, err := os.OpenFile("/tmp/pev-logs/coordinator.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "\n=== Received Plan ===\n")
		fmt.Fprintf(f, "Steps count: %d\n", len(plan.Steps))
		planJSON, _ := json.MarshalIndent(plan, "", "  ")
		fmt.Fprintf(f, "Plan JSON:\n%s\n", string(planJSON))
		f.Close()
	}

	state.PlanID = plan.ID
	state.CurrentPhase = "executing"

	// Send execute request to Executor
	executeReq := map[string]interface{}{
		"request_id": plan.RequestID,
		"plan_id":    plan.ID,
		"plan":       plan,
	}

	base.LogInfo("Triggering execution for plan %s", plan.ID)

	// Create and publish message
	execMsg := c.createMessage("execute_task", executeReq)
	c.publishToTopic("execute-tasks", execMsg)

	return nil, nil // Already published, don't send via egress
}

// handleExecutionResults receives execution results and triggers Verifier
func (c *PEVCoordinator) handleExecutionResults(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	payload := msg.Payload.(map[string]interface{})
	requestID := payload["request_id"].(string)

	state, ok := c.activeRequests[requestID]
	if !ok {
		base.LogDebug("Received execution results for unknown request %s (likely already completed), ignoring", requestID)
		return nil, nil
	}

	// Check if request already completed
	if state.CurrentPhase == "complete" || state.CurrentPhase == "failed" {
		base.LogDebug("Received execution results for already completed request %s, ignoring", requestID)
		return nil, nil
	}

	// Check if already verifying this execution (idempotency)
	if state.CurrentPhase == "verifying" {
		base.LogDebug("Request %s already verifying, ignoring duplicate execution results", requestID)
		return nil, nil
	}

	// Debug: check if step_results exists in payload
	stepResults, hasStepResults := payload["step_results"]
	if hasStepResults {
		if results, ok := stepResults.([]interface{}); ok {
			base.LogInfo("Received execution results for request %s: %d steps", requestID, len(results))
		} else {
			base.LogError("step_results is not an array: %T", stepResults)
		}
	} else {
		base.LogError("NO step_results in execution results payload! Keys: %v", getKeys(payload))
	}

	state.CurrentPhase = "verifying"

	// Store execution results for later summary generation
	state.LastExecutionResults = payload
	base.LogInfo("Stored execution results with keys: %v", getKeys(payload))

	// Send verify request to Verifier
	// Note: payload is already the ExecutionResults struct as a map
	// It contains: request_id, plan_id, step_results, all_success, etc.
	verifyReq := map[string]interface{}{
		"request_id":        requestID,
		"plan_id":           state.PlanID,
		"execution_results": payload, // Pass the entire payload as execution_results
		"goal":              state.UserRequest,
	}

	base.LogInfo("Triggering verification for request %s", requestID)

	// Create and publish message
	verifyMsg := c.createMessage("verify_request", verifyReq)
	c.publishToTopic("verify-requests", verifyMsg)

	return nil, nil // Already published, don't send via egress
}

// handleVerificationReport receives verification and decides: continue, re-plan, or complete
func (c *PEVCoordinator) handleVerificationReport(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	var report VerificationReport
	if err := unmarshalPayload(msg.Payload, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification report: %w", err)
	}

	state, ok := c.activeRequests[report.RequestID]
	if !ok {
		// Request already completed or unknown
		return nil, nil
	}

	// Check if already processed verification (idempotency)
	if state.CurrentPhase == "complete" || state.CurrentPhase == "planning" {
		base.LogInfo("Request %s already in phase %s, ignoring duplicate verification report", report.RequestID, state.CurrentPhase)
		return nil, nil
	}

	base.LogInfo("Received verification report for request %s: goal_achieved=%v", report.RequestID, report.GoalAchieved)

	if report.GoalAchieved {
		// SUCCESS - PEV cycle complete
		state.CurrentPhase = "complete"
		state.GoalAchieved = true
		state.CompletedAt = time.Now()

		// Generate human-friendly summary
		summary := summarizeAccomplishments(state.UserRequest, state.LastExecutionResults, base)

		response := map[string]interface{}{
			"type":           "user_response",
			"request_id":     report.RequestID,
			"status":         "complete",
			"iterations":     state.Iteration,
			"goal_achieved":  true,
			"message":        summary,
		}

		base.LogInfo("PEV cycle complete for request %s after %d iterations", report.RequestID, state.Iteration)
		// Don't delete immediately - let cleanup goroutine handle it after grace period

		// Create and publish message to alfa-responses topic (not pev-bus)
		// This way alfa doesn't see all the internal PEV messages
		respMsg := c.createMessage("user_response", response)
		base.LogInfo("Publishing completion response to alfa-responses")
		if err := c.publishToTopic("alfa-responses", respMsg); err != nil {
			return nil, fmt.Errorf("failed to publish completion response: %w", err)
		}
		base.LogInfo("Successfully published completion response")

		return nil, nil // Already published, don't send via egress
	}

	// Goal not achieved - check iteration limit
	if state.Iteration >= c.maxIterations {
		// MAX ITERATIONS REACHED
		state.CurrentPhase = "failed"
		state.CompletedAt = time.Now()

		// Format issues for user response
		issueStrings := make([]string, len(report.Issues))
		for i, issue := range report.Issues {
			issueStrings[i] = fmt.Sprintf("[%s] Step %s: %s", issue.Severity, issue.StepID, issue.Issue)
		}

		// Generate summary of partial accomplishments
		partialSummary := summarizeAccomplishments(state.UserRequest, state.LastExecutionResults, base)
		failureMsg := fmt.Sprintf("Could not complete your request after %d iterations. %s", c.maxIterations, partialSummary)
		if len(report.Issues) > 0 {
			failureMsg += fmt.Sprintf(" Main issue: %s", issueStrings[0])
		}

		response := map[string]interface{}{
			"type":          "user_response",
			"request_id":    report.RequestID,
			"status":        "failed",
			"iterations":    state.Iteration,
			"goal_achieved": false,
			"message":       failureMsg,
			"issues":        issueStrings,
			"next_actions":  report.NextActions,
		}

		base.LogInfo("Max iterations reached for request %s with %d issues", report.RequestID, len(report.Issues))
		// Don't delete immediately - let cleanup goroutine handle it after grace period

		// Create and publish message to alfa-responses topic (not pev-bus)
		failMsg := c.createMessage("user_response", response)
		base.LogInfo("Publishing failure response to alfa-responses")
		if err := c.publishToTopic("alfa-responses", failMsg); err != nil {
			return nil, fmt.Errorf("failed to publish failure response: %w", err)
		}
		base.LogInfo("Successfully published failure response")

		return nil, nil // Already published, don't send via egress
	}

	// RE-PLAN - iterate again
	state.Iteration++
	state.CurrentPhase = "planning"

	// Convert issues to strings for planner
	issueStrings := make([]string, len(report.Issues))
	for i, issue := range report.Issues {
		issueStrings[i] = fmt.Sprintf("[%s] Step %s: %s", issue.Severity, issue.StepID, issue.Issue)
	}

	planReq := PlanRequest{
		RequestID:      report.RequestID,
		UserRequest:    state.UserRequest,
		Context:        state.Context,
		Iteration:      state.Iteration,
		PreviousPlan:   state.PlanID,
		PreviousIssues: issueStrings,
	}

	base.LogInfo("Re-planning for request %s (iteration %d/%d) with %d issues",
		report.RequestID, state.Iteration, c.maxIterations, len(report.Issues))

	// Create and publish message
	replanMsg := c.createMessage("plan_request", planReq)
	c.publishToTopic("plan-requests", replanMsg)

	return nil, nil // Already published, don't send via egress
}

// Helper functions

// publishToTopic publishes a message directly to a specific topic
func (c *PEVCoordinator) publishToTopic(topic string, msg *client.BrokerMessage) error {
	if c.baseAgent == nil || c.baseAgent.BrokerClient == nil {
		// During testing without broker, just log
		c.baseAgent.LogDebug("Skipping publish to %s (no broker client)", topic)
		return nil
	}
	c.baseAgent.LogDebug("Publishing message %s (type: %s) to topic: %s", msg.ID, msg.Type, topic)
	err := c.baseAgent.BrokerClient.Publish(topic, *msg)
	if err != nil {
		c.baseAgent.LogError("Failed to publish to %s: %v", topic, err)
	} else {
		c.baseAgent.LogDebug("Successfully published to topic: %s", topic)
	}
	return err
}

func (c *PEVCoordinator) createMessage(msgType string, payload interface{}) *client.BrokerMessage {
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

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// summarizeAccomplishments creates a human-friendly summary of what was accomplished
func summarizeAccomplishments(userRequest string, executionResults map[string]interface{}, base *agent.BaseAgent) string {
	// Always log for debugging
	log.Printf("[SUMMARY DEBUG] Called with executionResults=%v", executionResults != nil)

	if executionResults == nil {
		log.Printf("[SUMMARY DEBUG] executionResults is nil")
		return "Completed your request."
	}

	// Debug: log what keys are available
	log.Printf("[SUMMARY DEBUG] Execution results keys: %v", getKeys(executionResults))

	// Extract step_results
	stepResults, ok := executionResults["step_results"].([]interface{})
	if !ok {
		log.Printf("[SUMMARY DEBUG] step_results not found or wrong type: %T", executionResults["step_results"])
		return "Completed your request."
	}

	if len(stepResults) == 0 {
		log.Printf("[SUMMARY DEBUG] step_results is empty")
		return "Completed your request."
	}

	log.Printf("[SUMMARY DEBUG] Found %d step results", len(stepResults))

	// Track accomplishments
	var filesCreated []string
	var filesModified []string
	var commandsRun []string
	var testsRun bool

	for i, stepInterface := range stepResults {
		step, ok := stepInterface.(map[string]interface{})
		if !ok {
			log.Printf("[SUMMARY DEBUG] Step %d: not a map: %T", i, stepInterface)
			continue
		}

		success, _ := step["success"].(bool)
		if !success {
			log.Printf("[SUMMARY DEBUG] Step %d: success=false, skipping", i)
			continue // Skip failed steps
		}

		action, _ := step["action"].(string)
		params, _ := step["params"].(map[string]interface{})

		log.Printf("[SUMMARY DEBUG] Step %d: action=%s, params keys=%v", i, action, getKeys(params))

		switch action {
		case "write_file":
			if path, ok := params["path"].(string); ok {
				filesCreated = append(filesCreated, path)
			}
		case "patch":
			if file, ok := params["file"].(string); ok {
				filesModified = append(filesModified, file)
			}
		case "run_command":
			if cmd, ok := params["command"].(string); ok {
				commandsRun = append(commandsRun, cmd)
				if len(cmd) > 4 && cmd[:4] == "test" {
					testsRun = true
				}
			}
		case "run_tests":
			testsRun = true
		}
	}

	// Build summary
	var parts []string

	if len(filesCreated) > 0 {
		if len(filesCreated) == 1 {
			parts = append(parts, fmt.Sprintf("created %s", filesCreated[0]))
		} else {
			parts = append(parts, fmt.Sprintf("created %d files (%s, ...)", len(filesCreated), filesCreated[0]))
		}
	}

	if len(filesModified) > 0 {
		if len(filesModified) == 1 {
			parts = append(parts, fmt.Sprintf("modified %s", filesModified[0]))
		} else {
			parts = append(parts, fmt.Sprintf("modified %d files", len(filesModified)))
		}
	}

	if len(commandsRun) > 0 && !testsRun {
		parts = append(parts, fmt.Sprintf("ran %d command(s)", len(commandsRun)))
	}

	if testsRun {
		parts = append(parts, "ran tests")
	}

	log.Printf("[SUMMARY DEBUG] Accomplishments - files created: %v, modified: %v, commands: %v, tests: %v",
		filesCreated, filesModified, commandsRun, testsRun)

	if len(parts) == 0 {
		log.Printf("[SUMMARY DEBUG] No accomplishments found, returning generic message")
		return "Completed your request."
	}

	// Format final message
	summary := "I " + parts[0]
	for i := 1; i < len(parts); i++ {
		if i == len(parts)-1 {
			summary += " and " + parts[i]
		} else {
			summary += ", " + parts[i]
		}
	}
	summary += "."

	log.Printf("[SUMMARY DEBUG] Generated summary: %s", summary)
	return summary
}

func main() {
	coordinator := NewPEVCoordinator()
	if err := agent.Run(coordinator, "pev-coordinator"); err != nil {
		panic(err)
	}
}
