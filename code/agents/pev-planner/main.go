package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tenzoki/agen/atomic/ai"
	"github.com/tenzoki/agen/cellorg/public/agent"
	"github.com/tenzoki/agen/cellorg/public/client"
)

// PEVPlanner generates execution plans from user requests
type PEVPlanner struct {
	agent.DefaultAgentRunner
	model            string
	temperature      float64
	llm              ai.LLM
	baseAgent        *agent.BaseAgent
	logIntermediates bool
	logPath          string
}

// PlanRequest from Coordinator
type PlanRequest struct {
	RequestID      string                 `json:"request_id"`
	UserRequest    string                 `json:"user_request"`
	Context        map[string]interface{} `json:"context"`
	Iteration      int                    `json:"iteration"`
	PreviousPlan   string                 `json:"previous_plan,omitempty"`
	PreviousIssues []string               `json:"previous_issues,omitempty"`
}

// ExecutionPlan to Coordinator
type ExecutionPlan struct {
	ID                string                   `json:"id"`
	RequestID         string                   `json:"request_id"`
	Type              string                   `json:"type"`
	Goal              string                   `json:"goal"`
	TargetContext     string                   `json:"target_context"`
	Steps             []PlanStep               `json:"steps"`
	OverallSuccess    string                   `json:"overall_success"`
}

// PlanStep defines a single step in the execution plan
type PlanStep struct {
	ID              string                 `json:"id"`
	Phase           string                 `json:"phase"` // discovery, analysis, implementation, validation
	Action          string                 `json:"action"` // search, read_file, patch, run_tests, etc.
	Params          map[string]interface{} `json:"params"`
	DependsOn       []string               `json:"depends_on"`
	SuccessCriteria string                 `json:"success_criteria"`
}

func NewPEVPlanner() *PEVPlanner {
	return &PEVPlanner{
		DefaultAgentRunner: agent.DefaultAgentRunner{},
		model:              "gpt-5",
		temperature:        0.7,
	}
}

func (p *PEVPlanner) Init(base *agent.BaseAgent) error {
	p.baseAgent = base

	// Load config
	if model := base.GetConfigString("model", ""); model != "" {
		p.model = model
	}
	p.logIntermediates = base.GetConfigBool("log_intermediates", false)
	p.logPath = base.GetConfigString("log_path", "tmp")

	// Resolve log path relative to CELLORG_DATA_ROOT if available
	if dataRoot := os.Getenv("CELLORG_DATA_ROOT"); dataRoot != "" {
		p.logPath = filepath.Join(dataRoot, p.logPath)
	}

	base.LogInfo("PEV Planner initializing with model: %s", p.model)
	if p.logIntermediates {
		base.LogInfo("LLM intermediates logging enabled: %s", p.logPath)
		// Create log directory
		os.MkdirAll(p.logPath, 0755)
	}

	// Initialize LLM client
	var err error
	p.llm, err = p.createLLMClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	base.LogInfo("PEV Planner initialized successfully")
	return nil
}

// createLLMClient creates the appropriate LLM client based on model
func (p *PEVPlanner) createLLMClient() (ai.LLM, error) {
	// Determine provider from model name
	var provider string
	modelLower := strings.ToLower(p.model)
	if strings.Contains(modelLower, "claude") {
		provider = "anthropic"
	} else if strings.Contains(modelLower, "gpt") || strings.Contains(modelLower, "o1") {
		provider = "openai"
	} else {
		provider = "openai" // Default to OpenAI
	}

	config := ai.Config{
		Model:       p.model,
		MaxTokens:   64000,
		Temperature: p.temperature,
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

func (p *PEVPlanner) ProcessMessage(msg *client.BrokerMessage, base *agent.BaseAgent) (*client.BrokerMessage, error) {
	if msg.Type != "plan_request" {
		return nil, fmt.Errorf("unsupported message type: %s", msg.Type)
	}

	var req PlanRequest
	if err := unmarshalPayload(msg.Payload, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plan request: %w", err)
	}

	base.LogInfo("Creating execution plan for request %s (iteration %d)", req.RequestID, req.Iteration)

	// Generate plan using LLM
	plan, err := p.createAIPlan(req, base)
	if err != nil {
		base.LogError("Failed to create AI plan: %v", err)
		// Fallback to hardcoded plan
		plan = p.createHardcodedPlan(req, base)
		base.LogInfo("Using fallback hardcoded plan")
	}

	base.LogInfo("Generated plan with %d steps", len(plan.Steps))

	// Debug: Log plan details before sending
	if len(plan.Steps) == 0 {
		base.LogError("CRITICAL: Planner generated plan with ZERO steps!")
	} else {
		base.LogInfo("\n=== Plan Generated ===")
		base.LogInfo("Plan ID: %s", plan.ID)
		base.LogInfo("Request ID: %s", plan.RequestID)
		base.LogInfo("Goal: %s", plan.Goal)
		base.LogInfo("Steps: %d steps", len(plan.Steps))
		for i, step := range plan.Steps {
			base.LogInfo("  %d. [%s] %s (phase: %s)", i+1, step.ID, step.Action, step.Phase)
		}
		base.LogInfo("======================\n")
	}

	// Debug: Write to file for inspection
	if f, err := os.OpenFile("/tmp/pev-logs/planner.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintf(f, "\n=== Plan Generated ===\n")
		fmt.Fprintf(f, "Steps count: %d\n", len(plan.Steps))
		planJSON, _ := json.MarshalIndent(plan, "", "  ")
		fmt.Fprintf(f, "Plan JSON:\n%s\n", string(planJSON))
		f.Close()
	}

	return p.createMessage("execution_plan", plan), nil
}

func (p *PEVPlanner) Cleanup(base *agent.BaseAgent) {
	base.LogInfo("PEV Planner cleanup")
}

// createAIPlan generates an execution plan using LLM
func (p *PEVPlanner) createAIPlan(req PlanRequest, base *agent.BaseAgent) (ExecutionPlan, error) {
	// Build system prompt for plan generation
	systemPrompt := p.buildPlanningSystemPrompt()

	// Build user prompt with request details
	userPrompt := p.buildPlanningUserPrompt(req)

	base.LogDebug("Calling LLM for plan generation...")

	// Call LLM
	ctx := context.Background()
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := p.llm.Chat(ctx, messages)
	if err != nil {
		return ExecutionPlan{}, fmt.Errorf("LLM call failed: %w", err)
	}

	base.LogInfo("LLM response received (%d tokens, %dms)",
		response.Usage.TotalTokens, response.ResponseTime.Milliseconds())

	// Log intermediates if enabled
	if p.logIntermediates {
		timestamp := time.Now().Format("20060102-150405")
		iteration := req.Iteration

		// Write prompt
		promptFile := fmt.Sprintf("%s/%s-planner-iter%d-prompt.txt", p.logPath, timestamp, iteration)
		promptContent := fmt.Sprintf("=== SYSTEM ===\n%s\n\n=== USER ===\n%s\n", systemPrompt, userPrompt)
		os.WriteFile(promptFile, []byte(promptContent), 0644)

		// Write completion
		completionFile := fmt.Sprintf("%s/%s-planner-iter%d-completion.txt", p.logPath, timestamp, iteration)
		os.WriteFile(completionFile, []byte(response.Content), 0644)

		base.LogDebug("Wrote LLM intermediates to %s", p.logPath)
	}

	// Parse JSON response
	plan, err := p.parsePlanFromResponse(response.Content, req)
	if err != nil {
		return ExecutionPlan{}, fmt.Errorf("failed to parse plan: %w", err)
	}

	// Store plan for future reference
	if storageErr := p.storePlan(plan, base); storageErr != nil {
		base.LogError("Failed to store plan: %v", storageErr)
		// Continue anyway - storage failure shouldn't block planning
	}

	return plan, nil
}

// buildPlanningSystemPrompt creates the system prompt for plan generation
func (p *PEVPlanner) buildPlanningSystemPrompt() string {
	return `You are an expert planning agent for the AGEN framework. Your job is to create detailed execution plans for code modification tasks.

**Available Tools:**
- search: Search codebase for files/patterns
- read_file: Read a file's contents
- write_file: Create or overwrite a file
- patch: Modify file with specific operations
- run_command: Execute shell command
- run_tests: Run tests

**Phases:**
- discovery: Find relevant files/code
- analysis: Understand current implementation
- implementation: Make code changes
- validation: Test and verify changes

**Your Response Must Be Valid JSON:**
{
  "goal": "Clear statement of what needs to be accomplished",
  "target_context": "framework" or "project",
  "steps": [
    {
      "id": "step-1",
      "phase": "discovery|analysis|implementation|validation",
      "action": "search|read_file|patch|run_tests|etc",
      "params": { /* action-specific parameters */ },
      "depends_on": ["step-id"],
      "success_criteria": "How to know this step succeeded"
    }
  ]
}

**Important:**
- Create 3-8 steps (not too many, not too few)
- Each step must have clear success criteria
- Use depends_on to ensure proper sequencing
- For patches, be specific about what to change
- Always include validation (tests) as final step
- Think step-by-step: discover → understand → implement → verify`
}

// buildPlanningUserPrompt creates the user prompt with request details
func (p *PEVPlanner) buildPlanningUserPrompt(req PlanRequest) string {
	var prompt strings.Builder

	// Target context
	targetContext := "framework"
	if req.Context != nil {
		if tc, ok := req.Context["target_vfs"].(string); ok {
			targetContext = tc
		}
	}

	prompt.WriteString(fmt.Sprintf("**User Request:** %s\n\n", req.UserRequest))
	prompt.WriteString(fmt.Sprintf("**Target Context:** %s\n", targetContext))
	prompt.WriteString(fmt.Sprintf("**Request ID:** %s\n", req.RequestID))
	prompt.WriteString(fmt.Sprintf("**Iteration:** %d\n\n", req.Iteration))

	// If this is a re-plan, include previous issues
	if req.Iteration > 1 && len(req.PreviousIssues) > 0 {
		prompt.WriteString("**Previous Attempt Failed Due To:**\n")
		for i, issue := range req.PreviousIssues {
			prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, issue))
		}
		prompt.WriteString("\n**Adjust the plan to address these issues.**\n\n")
	}

	if targetContext == "framework" {
		prompt.WriteString("**Note:** You are modifying the AGEN framework code itself (self-modification).\n")
		prompt.WriteString("Framework code is in: code/alfa/, code/cellorg/, code/omni/, code/atomic/\n\n")
	}

	prompt.WriteString("Generate an execution plan as JSON.")

	return prompt.String()
}

// parsePlanFromResponse extracts ExecutionPlan from LLM response
func (p *PEVPlanner) parsePlanFromResponse(content string, req PlanRequest) (ExecutionPlan, error) {
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
	var planData struct {
		Goal          string     `json:"goal"`
		TargetContext string     `json:"target_context"`
		Steps         []PlanStep `json:"steps"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &planData); err != nil {
		return ExecutionPlan{}, fmt.Errorf("JSON parse error: %w\nContent: %s", err, jsonStr)
	}

	// Validate plan
	if len(planData.Steps) == 0 {
		return ExecutionPlan{}, fmt.Errorf("plan has no steps")
	}

	// Create execution plan
	plan := ExecutionPlan{
		ID:             fmt.Sprintf("plan-%d", time.Now().UnixNano()),
		RequestID:      req.RequestID,
		Type:           "execution_plan",
		Goal:           planData.Goal,
		TargetContext:  planData.TargetContext,
		Steps:          planData.Steps,
		OverallSuccess: "All steps completed successfully and goal achieved",
	}

	return plan, nil
}

// storePlan stores the generated plan for future reference and learning
func (p *PEVPlanner) storePlan(plan ExecutionPlan, base *agent.BaseAgent) error {
	// Get data path from config (relative to VFS root)
	dataPath := base.GetConfigString("data_path", "data/planner")

	// Resolve relative to CELLORG_DATA_ROOT if available
	if dataRoot := os.Getenv("CELLORG_DATA_ROOT"); dataRoot != "" {
		dataPath = filepath.Join(dataRoot, dataPath)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Serialize plan to JSON
	planJSON, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	// Store plan with timestamp
	filename := fmt.Sprintf("%s/plan-%s-%d.json", dataPath, plan.RequestID, time.Now().Unix())
	if err := os.WriteFile(filename, planJSON, 0644); err != nil {
		return fmt.Errorf("failed to write plan file: %w", err)
	}

	base.LogDebug("Stored plan to: %s", filename)

	// TODO: When OmniStore is available, also store in:
	// - KV Store: plans/{plan.ID} → JSON
	// - Graph Store: (Request)-[:HAS_PLAN]->(Plan)-[:HAS_STEP]->(Step)
	// - Search Index: Full-text search over plans

	return nil
}

// createHardcodedPlan generates a simple plan for testing (Phase 1/2)
// Used as fallback if LLM fails
func (p *PEVPlanner) createHardcodedPlan(req PlanRequest, base *agent.BaseAgent) ExecutionPlan {
	planID := fmt.Sprintf("plan-%d", time.Now().UnixNano())

	// Extract target context (framework vs project)
	targetContext := "framework"
	if req.Context != nil {
		if tc, ok := req.Context["target_vfs"].(string); ok {
			targetContext = tc
		}
	}

	base.LogDebug("Creating plan for target context: %s", targetContext)

	// For Phase 1, create a simple hardcoded plan
	// This demonstrates the plan structure without actual AI
	steps := []PlanStep{
		{
			ID:              "step-1",
			Phase:           "discovery",
			Action:          "search",
			Params: map[string]interface{}{
				"query":   "prompt rendering",
				"pattern": "*.go",
			},
			DependsOn:       []string{},
			SuccessCriteria: "Find files related to user prompt",
		},
		{
			ID:              "step-2",
			Phase:           "analysis",
			Action:          "read_file",
			Params: map[string]interface{}{
				"path": "code/alfa/internal/orchestrator/orchestrator.go",
			},
			DependsOn:       []string{"step-1"},
			SuccessCriteria: "Understand current implementation",
		},
		{
			ID:              "step-3",
			Phase:           "implementation",
			Action:          "patch",
			Params: map[string]interface{}{
				"file": "code/alfa/internal/orchestrator/orchestrator.go",
				"operations": []map[string]interface{}{
					{
						"type":   "insert",
						"line":   100,
						"content": "// Hardcoded patch for testing",
					},
				},
			},
			DependsOn:       []string{"step-2"},
			SuccessCriteria: "Code modified successfully",
		},
		{
			ID:              "step-4",
			Phase:           "validation",
			Action:          "run_tests",
			Params: map[string]interface{}{
				"pattern": "./code/alfa/...",
			},
			DependsOn:       []string{"step-3"},
			SuccessCriteria: "All tests pass",
		},
	}

	// If re-planning (iteration > 1), add a note about previous issues
	if req.Iteration > 1 && len(req.PreviousIssues) > 0 {
		base.LogDebug("Re-planning due to issues: %v", req.PreviousIssues)
		// In Phase 3, the LLM will use this information to adjust the plan
	}

	return ExecutionPlan{
		ID:             planID,
		RequestID:      req.RequestID,
		Type:           "execution_plan",
		Goal:           req.UserRequest,
		TargetContext:  targetContext,
		Steps:          steps,
		OverallSuccess: "User request satisfied",
	}
}

// Helper functions

func (p *PEVPlanner) createMessage(msgType string, payload interface{}) *client.BrokerMessage {
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
	planner := NewPEVPlanner()
	if err := agent.Run(planner, "pev-planner"); err != nil {
		panic(err)
	}
}
