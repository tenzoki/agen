package main

import (
	"encoding/json"
	"fmt"
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
	RequestID      string
	UserRequest    string
	Context        map[string]interface{}
	Iteration      int
	CurrentPhase   string // "planning", "executing", "verifying", "complete"
	PlanID         string
	GoalAchieved   bool
	CreatedAt      time.Time
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

// handleUserRequest starts a new PEV cycle for a user request
func (c *PEVCoordinator) handleUserRequest(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	var req UserRequest
	if err := unmarshalPayload(msg.Payload, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user request: %w", err)
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
	c.publishToTopic("plan-requests", planMsg) // Best-effort publish to specific topic

	return planMsg, nil // Also return for framework to handle via egress
}

// handleExecutionPlan receives a plan and triggers Executor
func (c *PEVCoordinator) handleExecutionPlan(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	var plan ExecutionPlan
	if err := unmarshalPayload(msg.Payload, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal execution plan: %w", err)
	}

	state, ok := c.activeRequests[plan.RequestID]
	if !ok {
		return nil, fmt.Errorf("unknown request ID: %s", plan.RequestID)
	}

	base.LogInfo("Received execution plan with %d steps for request %s", len(plan.Steps), plan.RequestID)

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
	c.publishToTopic("execute-tasks", execMsg) // Best-effort publish to specific topic

	return execMsg, nil // Also return for framework to handle via egress
}

// handleExecutionResults receives execution results and triggers Verifier
func (c *PEVCoordinator) handleExecutionResults(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	payload := msg.Payload.(map[string]interface{})
	requestID := payload["request_id"].(string)

	state, ok := c.activeRequests[requestID]
	if !ok {
		return nil, fmt.Errorf("unknown request ID: %s", requestID)
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
	c.publishToTopic("verify-requests", verifyMsg) // Best-effort publish to specific topic

	return verifyMsg, nil // Also return for framework to handle via egress
}

// handleVerificationReport receives verification and decides: continue, re-plan, or complete
func (c *PEVCoordinator) handleVerificationReport(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	var report VerificationReport
	if err := unmarshalPayload(msg.Payload, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verification report: %w", err)
	}

	state, ok := c.activeRequests[report.RequestID]
	if !ok {
		return nil, fmt.Errorf("unknown request ID: %s", report.RequestID)
	}

	base.LogInfo("Received verification report for request %s: goal_achieved=%v", report.RequestID, report.GoalAchieved)

	if report.GoalAchieved {
		// SUCCESS - PEV cycle complete
		state.CurrentPhase = "complete"
		state.GoalAchieved = true

		response := map[string]interface{}{
			"request_id":     report.RequestID,
			"status":         "complete",
			"iterations":     state.Iteration,
			"goal_achieved":  true,
			"message":        "Request completed successfully",
		}

		base.LogInfo("PEV cycle complete for request %s after %d iterations", report.RequestID, state.Iteration)
		delete(c.activeRequests, report.RequestID)

		// Create and publish message
		respMsg := c.createMessage("user_response", response)
		c.publishToTopic("user-responses", respMsg) // Best-effort publish to specific topic

		return respMsg, nil // Also return for framework to handle via egress
	}

	// Goal not achieved - check iteration limit
	if state.Iteration >= c.maxIterations {
		// MAX ITERATIONS REACHED
		state.CurrentPhase = "failed"

		// Format issues for user response
		issueStrings := make([]string, len(report.Issues))
		for i, issue := range report.Issues {
			issueStrings[i] = fmt.Sprintf("[%s] Step %s: %s", issue.Severity, issue.StepID, issue.Issue)
		}

		response := map[string]interface{}{
			"request_id":    report.RequestID,
			"status":        "failed",
			"iterations":    state.Iteration,
			"goal_achieved": false,
			"message":       fmt.Sprintf("Failed after %d iterations", c.maxIterations),
			"issues":        issueStrings,
			"next_actions":  report.NextActions,
		}

		base.LogInfo("Max iterations reached for request %s with %d issues", report.RequestID, len(report.Issues))
		delete(c.activeRequests, report.RequestID)

		// Create and publish message
		failMsg := c.createMessage("user_response", response)
		c.publishToTopic("user-responses", failMsg) // Best-effort publish to specific topic

		return failMsg, nil // Also return for framework to handle via egress
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
	c.publishToTopic("plan-requests", replanMsg) // Best-effort publish to specific topic

	return replanMsg, nil // Also return for framework to handle via egress
}

// Helper functions

// publishToTopic publishes a message directly to a specific topic
func (c *PEVCoordinator) publishToTopic(topic string, msg *client.BrokerMessage) error {
	if c.baseAgent == nil || c.baseAgent.BrokerClient == nil {
		// During testing without broker, just log
		return nil
	}
	return c.baseAgent.BrokerClient.Publish(topic, *msg)
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

func main() {
	coordinator := NewPEVCoordinator()
	if err := agent.Run(coordinator, "pev-coordinator"); err != nil {
		panic(err)
	}
}
