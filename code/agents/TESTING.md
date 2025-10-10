# PEV Agent Testing Guide

## Quick Start

```bash
# Run all coordinator tests
cd code/agents/pev-coordinator
go test -v

# Expected: All tests pass ✓
```

---

## Test Options

### **Option 1: Unit Tests** (Fastest)

Test individual agent components in isolation.

```bash
cd code/agents/pev-coordinator
go test -v
```

**What it tests:**
- ✓ User request → plan_request flow
- ✓ Execution plan → execute_task flow
- ✓ Verification success → user_response
- ✓ Verification failure → re-planning
- ✓ Max iterations limit enforcement

**Output:**
```
=== RUN   TestPEVCoordinatorFlow
=== RUN   TestPEVCoordinatorFlow/HandleUserRequest
    ✓ User request processed, plan_request generated
=== RUN   TestPEVCoordinatorFlow/HandleExecutionPlan
    ✓ Execution plan processed, execute_task generated
=== RUN   TestPEVCoordinatorFlow/HandleVerificationReportSuccess
    ✓ Verification report processed, success response generated
=== RUN   TestPEVCoordinatorFlow/HandleVerificationReportReplan
    ✓ Re-planning triggered, iteration incremented
--- PASS: TestPEVCoordinatorFlow (0.00s)
=== RUN   TestPEVCoordinatorMaxIterations
    ✓ Max iterations enforced, failure response generated
--- PASS: TestPEVCoordinatorMaxIterations (0.00s)
PASS
```

---

### **Option 2: Integration Tests** (Full PEV Cycle)

Simulates complete Plan-Execute-Verify cycles with all 4 agents.

```bash
cd code/agents/pev-coordinator
go test -v -run "TestFull|TestPEVCycleWith"
```

**What it tests:**
- ✓ Complete PEV cycle (success in 1 iteration)
- ✓ PEV cycle with re-planning (3 iterations to success)
- ✓ Message flow between agents
- ✓ State management across iterations

**Output:**
```
=== RUN   TestFullPEVCycle

=== Full PEV Cycle Simulation ===
Step 1: User sends request...
  → Coordinator outputs: plan_request

Step 2: Planner creates execution plan...
  → Planner outputs: execute_task with 4 steps

Step 3: Executor executes plan steps...
  → Executor outputs: verify_request (all steps succeeded)

Step 4: Verifier validates results...
  → Verifier outputs: goal_achieved=true

Step 5: Coordinator sends final response...
  → Final response: status=complete, iterations=1

=== PEV Cycle Complete ===
✓ User request processed successfully in 1 iteration
--- PASS: TestFullPEVCycle (0.00s)

=== RUN   TestPEVCycleWithReplan

=== PEV Cycle with Re-planning ===
Iteration 1: Initial attempt...
  → Tests failed, re-planning...
Iteration 2: Second attempt...
  → Still failing, re-planning again...
Iteration 3: Third attempt...

=== Cycle Complete ===
✓ Goal achieved after 3 iterations
--- PASS: TestPEVCycleWithReplan (0.00s)
PASS
```

---

### **Option 3: Manual Testing with Cellorg** (Live Agents)

Run agents with the cellorg orchestrator for live testing.

#### Prerequisites:
1. Build all agents (already done):
```bash
ls -lh bin/pev-*
# Should show 4 binaries: coordinator, planner, executor, verifier
```

2. Cell configuration exists:
```bash
cat workbench/config/cells/alfa/plan-execute-verify.yaml
```

#### Start Orchestrator:
```bash
# Terminal 1: Start cellorg orchestrator
cd /Users/kai/Dropbox/qboot/projects/E10-agen/agen
./bin/orchestrator --config workbench/config/cells/alfa/plan-execute-verify.yaml
```

#### Start Agents (4 separate terminals):
```bash
# Terminal 2: Coordinator
cd /Users/kai/Dropbox/qboot/projects/E10-agen/agen
./bin/pev-coordinator --agent-id pev-coordinator-001 --gox-host localhost

# Terminal 3: Planner
./bin/pev-planner --agent-id pev-planner-001 --gox-host localhost

# Terminal 4: Executor
./bin/pev-executor --agent-id pev-executor-001 --gox-host localhost

# Terminal 5: Verifier
./bin/pev-verifier --agent-id pev-verifier-001 --gox-host localhost
```

#### Send Test Request:
```bash
# Terminal 6: Send test message via broker
# (This will be easier once alfa integration is complete)
```

**Note:** Full manual testing will be easier after Phase 7 (Alfa Integration).

---

## Test Results Summary

| Test Type | Duration | Coverage | Purpose |
|-----------|----------|----------|---------|
| **Unit Tests** | <1s | Individual components | Fast feedback during development |
| **Integration Tests** | <1s | Full PEV cycle | Verify agent interactions |
| **Manual Testing** | Variable | Live system | End-to-end validation |

---

## Current Test Status (Phase 1)

✅ **Completed:**
- Coordinator unit tests (5 test cases)
- Full PEV cycle simulation
- Re-planning scenario test
- Max iterations enforcement
- All agents build successfully

⏳ **Pending (Phase 2+):**
- Live agent communication via GOX
- Real tool execution (Phase 4)
- LLM-based planning (Phase 3)
- Verification with o1/Claude (Phase 5)

---

## Troubleshooting

### Tests fail with "no such file or directory"
```bash
# Ensure you're in the right directory
cd /Users/kai/Dropbox/qboot/projects/E10-agen/agen/code/agents/pev-coordinator
pwd  # Should end with: /code/agents/pev-coordinator
```

### Import errors
```bash
# Update dependencies
go mod tidy
```

### Agent binaries not found
```bash
# Rebuild agents
cd /Users/kai/Dropbox/qboot/projects/E10-agen/agen
go build -o bin/pev-coordinator code/agents/pev-coordinator/main.go
go build -o bin/pev-planner code/agents/pev-planner/main.go
go build -o bin/pev-executor code/agents/pev-executor/main.go
go build -o bin/pev-verifier code/agents/pev-verifier/main.go
```

---

## Next Steps

**Phase 2: Wire up pub/sub topics**
- [ ] Test message flow with cellorg orchestrator
- [ ] Verify agents can start/stop gracefully
- [ ] Test complete PEV cycle with live agents

**Phase 3: Add LLM intelligence to Planner**
- [ ] Integrate Claude Sonnet 4.5
- [ ] Replace hardcoded plans with AI-generated plans
- [ ] Test plan quality

**Phase 4: Connect Executor to actual tools**
- [ ] Integrate alfa's tool dispatcher
- [ ] Test read_file, patch, run_tests execution
- [ ] Handle tool errors gracefully

**Phase 5: Add LLM verification**
- [ ] Integrate o1 for deep reasoning
- [ ] Validate goal achievement
- [ ] Generate actionable next steps

---

## Contact

For questions or issues with testing, check:
- `guidelines/tasks.md` - Implementation roadmap
- `code/agents/pev-coordinator/coordinator_test.go` - Test examples
- `code/agents/pev-coordinator/integration_test.go` - Full cycle examples
