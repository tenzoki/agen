# Plan-Execute-Verify Architecture for ALFA

**Status**: Design Proposal
**Last Updated**: 2025-10-09
**Purpose**: Transform alfa from naive single-turn execution to intelligent iterative workflow

---

## Executive Summary

Current alfa architecture stops after a single successful tool execution. This prevents:
- Multi-step workflows (search → read → patch → test)
- Self-correction when approach fails
- Goal-oriented iteration until completion

**Solution**: Plan-Execute-Verify (PEV) pattern using AGEN's cell/agent architecture.

---

## Current Problem

### Naive Loop (Current)
```
User Request → AI Response → Execute Tools → if(success) STOP
```

**Issues**:
- Stops after first successful action (e.g., search)
- No concept of "goal achieved"
- No self-correction
- No iterative refinement

### Example Failure
```
User: "modify your code to add warning icon"
Alfa: [executes search, finds 10 matches] → STOPS
Expected: → read files → plan changes → patch files → test → commit
```

---

## Plan-Execute-Verify Pattern

### Conceptual Flow

```
┌─────────────────────────────────────────────────────────┐
│                    USER REQUEST                         │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  PLANNING PHASE                                         │
│  - Analyze request                                      │
│  - Break into subtasks                                  │
│  - Create execution plan                                │
│  - Identify success criteria                            │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  EXECUTION PHASE                                        │
│  - Execute tasks sequentially/parallel                  │
│  - Collect results and observations                     │
│  - Handle errors gracefully                             │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  VERIFICATION PHASE                                     │
│  - Check goal achievement                               │
│  - Validate results                                     │
│  - Identify issues/gaps                                 │
└─────────────────────────────────────────────────────────┘
                           ↓
              ┌────────────┴────────────┐
              │   Goal Achieved?        │
              └────────────┬────────────┘
                           │
          ┌────────────────┼────────────────┐
          │ YES            │ NO             │
          ↓                ↓
    ┌─────────┐      ┌────────────┐
    │ SUCCESS │      │  RE-PLAN   │
    │ RESPOND │      │  ADJUST    │
    └─────────┘      └──────┬─────┘
                             │
                             └──→ Back to EXECUTION
```

---

## AGEN Cell-Based Architecture

### Specialized Agents

#### 1. **Planner Agent**
**Purpose**: Strategic planning and task decomposition
**Model**: High-capability (GPT-5, Claude Opus 4.1, Claude Sonnet 4.5)
**Inputs**: User request, context, targetVFS info
**Outputs**: Structured execution plan

**Plan Structure**:
```json
{
  "request_id": "req-001",
  "goal": "Add warning triangle when self_modify=true",
  "target_context": "framework",
  "steps": [
    {
      "id": "step-1",
      "phase": "discovery",
      "action": "search",
      "params": {"query": "input prompt", "pattern": "*.go"},
      "success_criteria": "Find files with prompt rendering"
    },
    {
      "id": "step-2",
      "phase": "analysis",
      "action": "read_file",
      "params": {"path": "code/alfa/internal/orchestrator/orchestrator.go"},
      "depends_on": ["step-1"],
      "success_criteria": "Understand prompt rendering logic"
    },
    {
      "id": "step-3",
      "phase": "implementation",
      "action": "patch",
      "params": {
        "file": "code/alfa/internal/orchestrator/orchestrator.go",
        "operations": [...]
      },
      "depends_on": ["step-2"],
      "success_criteria": "Code modified to show warning icon"
    },
    {
      "id": "step-4",
      "phase": "validation",
      "action": "run_tests",
      "params": {"pattern": "./code/alfa/..."},
      "depends_on": ["step-3"],
      "success_criteria": "All tests pass"
    }
  ],
  "overall_success": "Warning triangle displays when self_modify=true"
}
```

#### 2. **Executor Agent**
**Purpose**: Execute individual plan steps
**Model**: Fast model (GPT-5-mini, Claude Sonnet)
**Inputs**: Plan step, context
**Outputs**: Execution results, observations

**Responsibilities**:
- Execute tool calls (read_file, patch, run_command, etc.)
- Collect detailed results
- Handle errors gracefully
- Report back to Coordinator

#### 3. **Verifier Agent**
**Purpose**: Validate results and assess goal achievement
**Model**: Analytical model (o1, Claude with extended thinking)
**Inputs**: Plan, execution results, goal criteria
**Outputs**: Verification report

**Verification Checks**:
```json
{
  "request_id": "req-001",
  "verification": {
    "goal_achieved": false,
    "completed_steps": ["step-1", "step-2", "step-3"],
    "failed_steps": ["step-4"],
    "issues": [
      {
        "step_id": "step-4",
        "issue": "Tests failed: compilation error in orchestrator.go:123",
        "severity": "critical"
      }
    ],
    "next_actions": [
      {
        "type": "fix",
        "description": "Fix syntax error in added code",
        "priority": "high"
      }
    ]
  }
}
```

#### 4. **Coordinator Agent**
**Purpose**: Orchestrate PEV loop, manage state
**Model**: Fast coordinator (GPT-5-mini)
**Inputs**: All agent outputs
**Outputs**: Orchestration decisions

**Responsibilities**:
- Trigger Planner for new request or re-plan
- Schedule Executor for parallel/sequential tasks
- Invoke Verifier after execution
- Decide: continue, re-plan, or complete
- Manage iteration limits (max 10 cycles)

---

## Cell Definition

**File**: `workbench/config/cells/alfa/plan-execute-verify.yaml`

```yaml
cell:
  id: "alfa:plan-execute-verify"
  description: "Plan-Execute-Verify loop for intelligent code modification"
  debug: true

  orchestration:
    startup_timeout: "30s"
    shutdown_timeout: "15s"
    max_retries: 3
    retry_delay: "3s"

  agents:
    # Coordinator - Orchestrates the PEV loop
    - id: "pev-coordinator-001"
      agent_type: "pev-coordinator"
      ingress: "sub:user-requests"
      egress: "pub:pev-commands"
      config:
        max_iterations: 10
        timeout_per_iteration: "5m"
        model: "gpt-5-mini"

    # Planner - Strategic planning
    - id: "pev-planner-001"
      agent_type: "pev-planner"
      ingress: "sub:plan-requests"
      egress: "pub:execution-plans"
      dependencies: ["pev-coordinator-001"]
      config:
        model: "claude-sonnet-4-5-20250929"
        temperature: 0.7
        max_tokens: 64000
        data_path: "./data/planner"

    # Executor - Execute plan steps
    - id: "pev-executor-001"
      agent_type: "pev-executor"
      ingress: "sub:execute-tasks"
      egress: "pub:execution-results"
      dependencies: ["pev-planner-001"]
      config:
        model: "gpt-5-mini"
        temperature: 0
        max_tokens: 128000
        tools_enabled: true
        data_path: "./data/executor"

    # Verifier - Validate results
    - id: "pev-verifier-001"
      agent_type: "pev-verifier"
      ingress: "sub:verify-requests"
      egress: "pub:verification-reports"
      dependencies: ["pev-executor-001"]
      config:
        model: "o1"
        temperature: 0
        max_tokens: 100000
        strict_validation: true
        data_path: "./data/verifier"

    # Knowledge Store - Persistent state
    - id: "pev-knowledge-001"
      agent_type: "knowledge-store"
      ingress: "sub:knowledge-ops"
      egress: "pub:knowledge-results"
      config:
        data_path: "./data/pev-knowledge"
        enable_graph: true
        enable_search: true
```

---

## Communication Flow

### Topics

**Request Flow**:
```
user-requests        → Coordinator receives user input
plan-requests        → Coordinator → Planner
execution-plans      → Planner → Coordinator
execute-tasks        → Coordinator → Executor
execution-results    → Executor → Coordinator
verify-requests      → Coordinator → Verifier
verification-reports → Verifier → Coordinator
pev-commands         → Coordinator → System
knowledge-ops        → Any agent → Knowledge Store
knowledge-results    → Knowledge Store → Any agent
```

### Message Formats

**User Request**:
```json
{
  "id": "req-001",
  "type": "user_request",
  "content": "modify your code to add warning icon when self_modify=true",
  "context": {
    "target_vfs": "framework",
    "target_root": "/Users/kai/.../agen",
    "self_modify_enabled": true
  }
}
```

**Execution Plan** (Planner → Coordinator):
```json
{
  "id": "plan-001",
  "request_id": "req-001",
  "type": "execution_plan",
  "plan": { ... }  // See Plan Structure above
}
```

**Verification Report** (Verifier → Coordinator):
```json
{
  "id": "verify-001",
  "request_id": "req-001",
  "type": "verification_report",
  "goal_achieved": false,
  "issues": [...],
  "next_actions": [...]
}
```

---

## Storage Strategy

### OmniStore Usage

Each agent uses OmniStore for:

1. **KV Store**: Persist agent state, configuration
2. **Graph Store**: Model relationships (requests → plans → steps → results)
3. **Files Store**: Cache intermediate artifacts
4. **Search**: Full-text search over plans, results

**Graph Schema**:
```
(Request) -[:HAS_PLAN]→ (Plan)
(Plan) -[:HAS_STEP]→ (Step)
(Step) -[:EXECUTED_AS]→ (Execution)
(Execution) -[:PRODUCED]→ (Result)
(Result) -[:VERIFIED_BY]→ (Verification)
(Verification) -[:TRIGGERS]→ (ReplanDecision)
```

**Benefits**:
- Track full execution history
- Analyze patterns (which approaches succeed)
- Learn from failures (avoid repeating mistakes)
- Provide explainability (why did it do X?)

---

## Model Selection Strategy

### Rationale for Different Models

| Agent | Model | Reasoning |
|-------|-------|-----------|
| **Planner** | Claude Sonnet 4.5 / GPT-5 | Strategic thinking, task decomposition requires high capability |
| **Executor** | GPT-5-mini / Claude Haiku | Fast execution, straightforward tool use, cost-effective for repetitive tasks |
| **Verifier** | o1 / Claude Opus | Deep reasoning to validate results, catch subtle issues |
| **Coordinator** | GPT-5-mini | Simple orchestration logic, low latency needed |

### Cost-Performance Tradeoff

**High-value operations** (planning, verification): Premium models
**High-volume operations** (execution steps): Fast, cheap models

**Example Cost Profile** (per request with 5 iterations):
- Planner: 1-2 calls × $expensive = $X
- Executor: 20-30 calls × $cheap = $Y << $X
- Verifier: 5 calls × $expensive = $Z
- Coordinator: 50 calls × $cheap = $W << $X

**Total**: Dominated by Planner + Verifier, but they're called rarely.

---

## Integration with Alfa

### Current Orchestrator Replacement

**Before** (`orchestrator.go`):
```go
func (o *Orchestrator) processRequest(ctx context.Context, userInput string) error {
    // Naive: AI → Actions → Execute → if(success) STOP
    for iteration < maxIterations {
        response := llm.Chat(messages)
        actions := parseActions(response)
        results := executeActions(actions)

        if isComplete(results) {  // ← PROBLEM: stops on success
            return nil
        }
    }
}
```

**After** (PEV-enabled):
```go
func (o *Orchestrator) processRequest(ctx context.Context, userInput string) error {
    // Publish request to PEV cell
    request := &client.BrokerMessage{
        ID: generateID(),
        Payload: map[string]interface{}{
            "content": userInput,
            "context": o.getTargetContext(),  // framework vs project
        },
    }

    // Publish to user-requests topic
    o.cellBroker.Publish("user-requests", request)

    // Subscribe to responses
    responseChan := o.cellBroker.Subscribe("user-responses")

    // Wait for completion (with timeout)
    select {
    case response := <-responseChan:
        return o.handlePEVResponse(response)
    case <-time.After(10 * time.Minute):
        return fmt.Errorf("PEV timeout")
    }
}
```

### Alfa Becomes a Cell Orchestrator

Instead of:
- Alfa directly executing tools ❌

We have:
- Alfa orchestrates PEV cell ✅
- PEV agents do the heavy lifting ✅
- Alfa just publishes request and waits for result ✅

---

## Implementation Phases

### Phase 1: Agent Skeletons (Week 1)
- [x] Implement 4 specialized agents (Coordinator, Planner, Executor, Verifier)
- [x] 3-method pattern (Init, ProcessMessage, Cleanup)
- [x] Basic pub/sub communication
- [x] No AI yet - hardcoded logic for testing

### Phase 2: PEV Cell (Week 1-2)
- [x] Create cell YAML definition
- [x] Wire up pub/sub topics
- [x] Test message flow with mock data
- [x] Verify agents can start/stop

### Phase 3: Planner Intelligence (Week 2)
- [x] Integrate LLM (GPT-5)
- [x] Implement plan generation from user request
- [x] Store plans in OmniStore
- [x] Test: Can it create valid plans?

### Phase 4: Executor Tools (Week 2-3)
- [x] Implement tool dispatcher in Executor
- [x] Connect to existing alfa tools (read_file, patch, etc.)
- [x] Execute plan steps sequentially
- [x] Report results back

### Phase 5: Verifier Logic (Week 3)
- [x] Integrate analytical LLM (o1)
- [x] Implement goal-checking logic
- [x] Generate verification reports
- [x] Decide: continue, re-plan, or done

### Phase 6: Coordinator Loop (Week 3-4)
- [x] Implement PEV loop orchestration
- [x] Handle re-planning on failure
- [x] Enforce iteration limits
- [x] Graceful termination

### Phase 7: Alfa Integration (Week 4)
- [x] Replace naive loop with PEV cell invocation
- [x] Pass targetVFS context to PEV
- [x] Handle responses
- [x] UI for showing progress

### Phase 8: Knowledge Graph (Week 5)
- [x] Model request → plan → execution graph
- [x] Store in OmniStore KV + Search
- [x] Query: "What worked last time for similar requests?"
- [x] Learning from history (MVP complete)

---

## Success Criteria

### Functional Requirements
- ✅ Multi-step workflows complete automatically
- ✅ Self-correction on errors
- ✅ Iterates until goal achieved
- ✅ Works for both framework and project modification
- ✅ Handles complex requests (5+ steps)

### Performance Requirements
- Max 10 iterations per request
- Response time: < 5 minutes for typical request
- Cost: < $0.50 per request (average)

### Quality Requirements
- Code modifications pass tests before commit
- Verification catches >90% of errors before user sees them
- Clear explanation of what was done and why

---

## Example Scenario

### Request
```
"Modify your code so that when self_modify=true,
a warning triangle (⚠️) appears at the start of the input prompt"
```

### Execution Trace

**Iteration 1**:

1. **Planner** creates plan:
   - Step 1: Search for "prompt" in orchestrator
   - Step 2: Read orchestrator.go
   - Step 3: Find getUserInput() function
   - Step 4: Patch to add warning icon
   - Step 5: Build and test

2. **Executor** executes steps 1-4:
   - Searches → finds orchestrator.go:217
   - Reads file → understands prompt rendering
   - Applies patch → adds `if allowSelfModify { fmt.Print("⚠️ ") }`
   - Build succeeds

3. **Verifier** checks results:
   - ❌ "Warning only shows in text mode, not voice mode"
   - Recommends: "Also check voice input handling"

**Iteration 2** (Re-plan):

1. **Planner** adjusts:
   - Step 6: Check voice input code path
   - Step 7: Patch voice prompt too

2. **Executor** executes:
   - Reads voice input code
   - Patches both text and voice prompts

3. **Verifier** checks:
   - ✅ "Warning displays in both modes"
   - ✅ "Tests pass"
   - ✅ "Goal achieved"

**Result**: Request complete in 2 iterations, 7 steps total.

---

## Advantages over Naive Approach

| Aspect | Naive Loop | PEV Architecture |
|--------|-----------|------------------|
| **Planning** | None - reactive only | Explicit plan with steps |
| **Execution** | Stops after first success | Continues until goal |
| **Error Handling** | User sees errors | Self-correction via re-planning |
| **Multi-step** | Requires AI to batch all actions | Natural sequential/parallel execution |
| **Verification** | None | Explicit goal checking |
| **Learning** | None | History stored in graph |
| **Explainability** | "It ran search" | "Plan: search → read → patch → verify" |
| **Cost** | Expensive model for everything | Right model for right task |

---

## Risks and Mitigations

### Risk 1: Infinite Loops
**Mitigation**: Hard limit of 10 iterations. Coordinator enforces.

### Risk 2: Cost Explosion
**Mitigation**: Use cheap models for high-volume (Executor). Monitor cost per request.

### Risk 3: Complex State Management
**Mitigation**: OmniStore handles all state. Single source of truth.

### Risk 4: Latency
**Mitigation**: Parallel execution where possible. Fast models for Executor.

### Risk 5: Integration Complexity
**Mitigation**: Phase implementation. Test each agent independently first.

---

## Future Enhancements

### Phase 9: Meta-Learning
- Analyze which plans succeed/fail
- Extract patterns: "For requests like X, plan Y works best"
- Auto-suggest plans based on history

### Phase 10: Parallel Execution
- Executor spawns multiple workers for independent steps
- Dependency graph determines execution order

### Phase 11: Human-in-the-Loop
- Planner asks user to clarify ambiguous requests
- Verifier requests user validation for critical changes

### Phase 12: Multi-Agent Collaboration
- Different Executors specialize (one for Go, one for YAML, etc.)
- Coordinator routes steps to specialists

---

## Conclusion

Plan-Execute-Verify transforms alfa from a naive tool-executor into an intelligent, goal-oriented system that:

1. **Plans** before acting
2. **Executes** methodically
3. **Verifies** results
4. **Iterates** until success

By leveraging AGEN's cell/agent architecture, we get:
- **Modularity**: Each agent has one job
- **Scalability**: Add more agents as needed
- **Maintainability**: Test agents independently
- **Observability**: Track execution in graph store
- **Cost-efficiency**: Right model for right task

This architecture positions alfa as a true **agentic coding assistant** capable of handling complex, multi-step code modifications with minimal user intervention.

---

## References

- **ReAct Pattern**: [react-lm.github.io](https://react-lm.github.io/)
- **Claude Code Architecture**: [Anthropic Engineering Blog](https://www.anthropic.com/engineering/claude-code-best-practices)
- **Agentic Workflows 2025**: [Weaviate Blog](https://weaviate.io/blog/what-are-agentic-workflows)
- **AGEN Architecture**: `guidelines/references/architecture.md`
- **Agent Patterns**: `guidelines/references/agent-patterns.md`
