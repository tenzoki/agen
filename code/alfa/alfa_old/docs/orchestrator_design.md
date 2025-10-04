# Orchestrator Design

## Overview

The Orchestrator is the central control system that coordinates all components (AI, Speech, Context, Tools, Safety, VCR) to enable interactive voice/text-based coding assistance.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      ORCHESTRATOR                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Input Handler                             │  │
│  │  - Text input (CLI)                                    │  │
│  │  - Voice input (STT)                                   │  │
│  │  - Mode detection (confirm vs allow-all)              │  │
│  └───────────────────────────────────────────────────────┘  │
│                            ↓                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           Context Manager Integration                  │  │
│  │  - Retrieve relevant context                          │  │
│  │  - Format conversation history                        │  │
│  │  - Maintain working set of files                      │  │
│  └───────────────────────────────────────────────────────┘  │
│                            ↓                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              AI Layer Communication                    │  │
│  │  - Build prompt with context                          │  │
│  │  - Send to LLM (Claude/OpenAI)                        │  │
│  │  - Receive structured response                        │  │
│  └───────────────────────────────────────────────────────┘  │
│                            ↓                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           Response Parser/Validator                    │  │
│  │  - Parse JSON responses                               │  │
│  │  - Validate tool calls                                │  │
│  │  - Validate patches                                   │  │
│  │  - Handle invalid output → reprompt                   │  │
│  └───────────────────────────────────────────────────────┘  │
│                            ↓                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Tool Dispatcher                           │  │
│  │  - Route to appropriate tool                          │  │
│  │  - Execute with safety checks                         │  │
│  │  - Collect results                                    │  │
│  └───────────────────────────────────────────────────────┘  │
│                            ↓                                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │          Result Handler & Feedback Loop                │  │
│  │  - Update context with results                        │  │
│  │  - Auto-commit on success                             │  │
│  │  - Format response for user                           │  │
│  │  - TTS output (optional)                              │  │
│  │  - Continue conversation or exit                      │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Orchestrator Struct

```go
package orchestrator

import (
    "alfa/internal/ai"
    "alfa/internal/context"
    "alfa/internal/speech"
    "alfa/internal/tools"
    "alfa/internal/vcr"
    "context"
)

type Mode int

const (
    ModeConfirm Mode = iota  // Ask before each operation
    ModeAllowAll             // Execute without confirmation
)

type Orchestrator struct {
    // Core dependencies
    llm            ai.LLM
    contextMgr     *context.Manager
    toolDispatcher *tools.Dispatcher
    vcr            *vcr.Vcr

    // Optional speech components
    stt speech.STT
    tts speech.TTS

    // Configuration
    mode           Mode
    workdir        string
    maxIterations  int  // Prevent infinite loops

    // State
    conversationID string
    running        bool
}

type Config struct {
    LLM            ai.LLM
    ContextManager *context.Manager
    ToolDispatcher *tools.Dispatcher
    VCR            *vcr.Vcr
    STT            speech.STT  // optional
    TTS            speech.TTS  // optional
    Mode           Mode
    Workdir        string
    MaxIterations  int
}

func New(cfg Config) *Orchestrator {
    if cfg.MaxIterations == 0 {
        cfg.MaxIterations = 10
    }

    return &Orchestrator{
        llm:            cfg.LLM,
        contextMgr:     cfg.ContextManager,
        toolDispatcher: cfg.ToolDispatcher,
        vcr:            cfg.VCR,
        stt:            cfg.STT,
        tts:            cfg.TTS,
        mode:           cfg.Mode,
        workdir:        cfg.Workdir,
        maxIterations:  cfg.MaxIterations,
        conversationID: generateID(),
    }
}
```

---

### 2. Main Control Loop

```go
// Run starts the orchestrator's main interaction loop
func (o *Orchestrator) Run(ctx context.Context) error {
    o.running = true
    defer func() { o.running = false }()

    // Initialize system prompt
    systemPrompt := o.buildSystemPrompt()

    for o.running {
        // 1. Get user input
        userInput, err := o.getUserInput(ctx)
        if err != nil {
            return err
        }

        if userInput == "" || userInput == "exit" || userInput == "quit" {
            break
        }

        // 2. Process the request
        err = o.processRequest(ctx, userInput, systemPrompt)
        if err != nil {
            o.respond(ctx, fmt.Sprintf("Error: %v", err))
            continue
        }
    }

    return nil
}

// processRequest handles a single user request through multiple AI iterations
func (o *Orchestrator) processRequest(ctx context.Context, userInput string, systemPrompt string) error {
    // Add user message to context
    o.contextMgr.AddUserMessage(userInput)

    iteration := 0
    for iteration < o.maxIterations {
        iteration++

        // 1. Build messages with context
        messages := o.buildMessages(systemPrompt)

        // 2. Call AI
        response, err := o.llm.Chat(ctx, messages)
        if err != nil {
            return fmt.Errorf("AI error: %w", err)
        }

        // 3. Add assistant response to context
        o.contextMgr.AddAssistantMessage(response.Content)

        // 4. Parse response for actions
        actions, textResponse, err := o.parseResponse(response.Content)
        if err != nil {
            // Invalid output, reprompt
            o.contextMgr.AddUserMessage("Invalid output format. Please provide valid JSON for tool calls or patches.")
            continue
        }

        // 5. If no actions, just respond and done
        if len(actions) == 0 {
            o.respond(ctx, textResponse)
            return nil
        }

        // 6. Execute actions
        results, err := o.executeActions(ctx, actions)
        if err != nil {
            return err
        }

        // 7. Check if we're done (all actions successful, no continuation needed)
        if o.isComplete(results) {
            // Auto-commit if successful operations occurred
            if o.hasFileModifications(results) {
                commitMsg := o.generateCommitMessage(actions, results)
                o.vcr.Commit(commitMsg)
            }

            o.respond(ctx, o.formatResults(results))
            return nil
        }

        // 8. Feed results back to AI for next iteration
        o.contextMgr.AddToolResults(results)
    }

    return fmt.Errorf("max iterations (%d) reached", o.maxIterations)
}
```

---

### 3. Input/Output Handling

```go
// getUserInput gets input from text or voice
func (o *Orchestrator) getUserInput(ctx context.Context) (string, error) {
    if o.stt != nil {
        // Voice input mode
        fmt.Println("\n🎤 Press Enter to speak (or type 'text' for text mode)...")
        var input string
        fmt.Scanln(&input)

        if input == "text" {
            return o.getTextInput()
        }

        // Record audio and transcribe
        audioFile, err := o.recordAudio(ctx)
        if err != nil {
            return "", err
        }

        transcription, err := o.stt.TranscribeFile(ctx, audioFile)
        if err != nil {
            return "", err
        }

        fmt.Printf("You said: %s\n", transcription.Text)
        return transcription.Text, nil
    }

    // Text input mode
    return o.getTextInput()
}

func (o *Orchestrator) getTextInput() (string, error) {
    fmt.Print("\n> ")
    var input string
    scanner := bufio.NewScanner(os.Stdin)
    if scanner.Scan() {
        input = scanner.Text()
    }
    return input, scanner.Err()
}

// respond sends output to user via text or voice
func (o *Orchestrator) respond(ctx context.Context, message string) {
    fmt.Println(message)

    if o.tts != nil && message != "" {
        // Also speak the response
        go func() {
            audioPath := filepath.Join(o.workdir, "output", "response.mp3")
            err := o.tts.SynthesizeToFile(ctx, message, audioPath)
            if err == nil {
                // Play audio (implementation depends on platform)
                playAudio(audioPath)
            }
        }()
    }
}
```

---

### 4. Response Parsing

```go
// ActionType represents different kinds of actions the AI can request
type ActionType string

const (
    ActionPatch      ActionType = "patch"
    ActionToolCall   ActionType = "tool_call"
    ActionReadFile   ActionType = "read_file"
    ActionWriteFile  ActionType = "write_file"
    ActionRunCommand ActionType = "run_command"
    ActionRunTests   ActionType = "run_tests"
    ActionSearch     ActionType = "search"
)

type Action struct {
    Type   ActionType
    Params map[string]interface{}
}

// parseResponse extracts structured actions and text from AI response
func (o *Orchestrator) parseResponse(content string) ([]Action, string, error) {
    // Expected format:
    // 1. Text explanation/response
    // 2. JSON block(s) with actions
    //
    // Example:
    // I'll fix the bug in main.go by updating line 42.
    // ```json
    // {
    //   "action": "patch",
    //   "file": "main.go",
    //   "operations": [
    //     {"line": 42, "type": "replace", "content": ["fixed line"]}
    //   ]
    // }
    // ```

    var actions []Action
    var textParts []string

    // Split by code blocks
    parts := extractCodeBlocks(content)

    for _, part := range parts {
        if part.IsCode && part.Language == "json" {
            // Try to parse as action
            action, err := parseAction(part.Content)
            if err != nil {
                return nil, "", fmt.Errorf("invalid action JSON: %w", err)
            }
            actions = append(actions, action)
        } else {
            textParts = append(textParts, part.Content)
        }
    }

    textResponse := strings.Join(textParts, "\n")
    return actions, textResponse, nil
}

func parseAction(jsonStr string) (Action, error) {
    var raw map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
        return Action{}, err
    }

    actionType, ok := raw["action"].(string)
    if !ok {
        return Action{}, fmt.Errorf("missing 'action' field")
    }

    return Action{
        Type:   ActionType(actionType),
        Params: raw,
    }, nil
}
```

---

### 5. Action Execution

```go
// executeActions runs all requested actions with user confirmation if needed
func (o *Orchestrator) executeActions(ctx context.Context, actions []Action) ([]ActionResult, error) {
    var results []ActionResult

    for _, action := range actions {
        // Confirm with user if in confirm mode
        if o.mode == ModeConfirm {
            if !o.confirmAction(action) {
                results = append(results, ActionResult{
                    Action:  action,
                    Success: false,
                    Message: "User cancelled operation",
                })
                continue
            }
        }

        // Execute the action
        result := o.executeAction(ctx, action)
        results = append(results, result)

        // Stop on critical errors
        if !result.Success && result.Critical {
            break
        }
    }

    return results, nil
}

type ActionResult struct {
    Action   Action
    Success  bool
    Message  string
    Output   interface{}
    Critical bool  // If true, stop processing further actions
}

func (o *Orchestrator) executeAction(ctx context.Context, action Action) ActionResult {
    switch action.Type {
    case ActionPatch:
        return o.executePatch(action)
    case ActionToolCall:
        return o.toolDispatcher.Execute(ctx, action)
    case ActionReadFile:
        return o.executeReadFile(action)
    case ActionWriteFile:
        return o.executeWriteFile(action)
    case ActionRunCommand:
        return o.executeRunCommand(ctx, action)
    case ActionRunTests:
        return o.executeRunTests(ctx, action)
    case ActionSearch:
        return o.executeSearch(action)
    default:
        return ActionResult{
            Action:  action,
            Success: false,
            Message: fmt.Sprintf("unknown action type: %s", action.Type),
        }
    }
}

func (o *Orchestrator) executePatch(action Action) ActionResult {
    filePath := action.Params["file"].(string)
    operations := action.Params["operations"]

    // Convert to JSON for textpatch
    opsJSON, err := json.Marshal(operations)
    if err != nil {
        return ActionResult{
            Action:  action,
            Success: false,
            Message: fmt.Errorf("invalid patch operations: %w", err).Error(),
        }
    }

    // Apply patch
    err = textpatch.PatchFile(filePath, string(opsJSON))
    if err != nil {
        return ActionResult{
            Action:  action,
            Success: false,
            Message: fmt.Errorf("patch failed: %w", err).Error(),
        }
    }

    // Update context with file modification
    o.contextMgr.RecordFileModification(filePath, string(opsJSON))

    return ActionResult{
        Action:  action,
        Success: true,
        Message: fmt.Sprintf("Successfully patched %s", filePath),
    }
}

func (o *Orchestrator) confirmAction(action Action) bool {
    fmt.Printf("\n⚠️  AI wants to perform: %s\n", action.Type)
    fmt.Printf("Details: %+v\n", action.Params)
    fmt.Print("Proceed? [Y/n]: ")

    var response string
    fmt.Scanln(&response)

    return response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}
```

---

### 6. System Prompt Builder

```go
func (o *Orchestrator) buildSystemPrompt() string {
    return `You are an AI coding assistant with access to tools and the ability to modify code.

CAPABILITIES:
1. Read and analyze code files
2. Apply patches to fix bugs or add features
3. Run tests and commands in a sandbox
4. Search through the codebase
5. Generate git commits for changes

RESPONSE FORMAT:
Always structure your responses as:
1. Natural language explanation of what you'll do
2. JSON action blocks wrapped in \`\`\`json ... \`\`\` for operations

AVAILABLE ACTIONS:
- patch: Apply code changes
- read_file: Read file contents
- write_file: Create new file
- run_command: Execute shell command (sandboxed)
- run_tests: Run test suite
- search: Search codebase

PATCH ACTION FORMAT:
\`\`\`json
{
  "action": "patch",
  "file": "path/to/file.go",
  "operations": [
    {"line": 10, "type": "insert", "content": ["new line 1", "new line 2"]},
    {"line": 15, "type": "delete"},
    {"line": 20, "type": "replace", "content": ["replacement line"]}
  ]
}
\`\`\`

TOOL CALL FORMAT:
\`\`\`json
{
  "action": "tool_call",
  "tool": "run_tests",
  "args": {"pattern": "./..."}
}
\`\`\`

GUIDELINES:
- Always explain your reasoning before taking actions
- Test changes after modifications
- Keep patches focused and atomic
- Include error handling
- Follow existing code style

Current working directory: ` + o.workdir + `
`
}

func (o *Orchestrator) buildMessages(systemPrompt string) []ai.Message {
    messages := []ai.Message{
        {Role: "system", Content: systemPrompt},
    }

    // Get conversation history from context manager
    history := o.contextMgr.GetRecentHistory(20) // Last 20 messages
    messages = append(messages, history...)

    return messages
}
```

---

### 7. Commit Message Generation

```go
func (o *Orchestrator) generateCommitMessage(actions []Action, results []ActionResult) string {
    // Analyze what was done
    var operations []string
    var files []string

    for i, result := range results {
        if !result.Success {
            continue
        }

        action := actions[i]
        switch action.Type {
        case ActionPatch:
            file := action.Params["file"].(string)
            files = append(files, file)
            operations = append(operations, fmt.Sprintf("modified %s", file))
        case ActionWriteFile:
            file := action.Params["path"].(string)
            files = append(files, file)
            operations = append(operations, fmt.Sprintf("created %s", file))
        }
    }

    if len(operations) == 0 {
        return "AI assistant changes"
    }

    summary := strings.Join(operations, ", ")
    return fmt.Sprintf("AI: %s", summary)
}
```

---

## Usage Example

```go
package main

import (
    "context"
    "os"

    "alfa/internal/ai"
    "alfa/internal/context"
    "alfa/internal/orchestrator"
    "alfa/internal/speech"
    "alfa/internal/tools"
    "alfa/internal/vcr"
)

func main() {
    ctx := context.Background()
    workdir, _ := os.Getwd()

    // Setup components
    llm := ai.NewClaudeClient(ai.Config{
        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
        Model:  "claude-3-5-sonnet-20241022",
    })

    contextMgr := context.NewManager(workdir)
    toolDispatcher := tools.NewDispatcher(workdir)
    vcrInstance := vcr.NewVcr("assistant", workdir)

    // Optional: Add speech
    var stt speech.STT
    var tts speech.TTS
    if os.Getenv("ENABLE_VOICE") == "true" {
        stt = speech.NewWhisperClient(/* config */)
        tts = speech.NewOpenAITTS(/* config */)
    }

    // Create orchestrator
    orch := orchestrator.New(orchestrator.Config{
        LLM:            llm,
        ContextManager: contextMgr,
        ToolDispatcher: toolDispatcher,
        VCR:            vcrInstance,
        STT:            stt,
        TTS:            tts,
        Mode:           orchestrator.ModeConfirm,
        Workdir:        workdir,
    })

    // Run main loop
    if err := orch.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

---

## Key Design Decisions

1. **Iteration-based processing**: AI may need multiple calls to complete a task (plan → execute → verify)

2. **Structured output parsing**: AI must emit JSON for actions, making responses machine-parsable

3. **Confirmation modes**: Safety through user confirmation vs speed through auto-execution

4. **Context integration**: Orchestrator doesn't manage context directly, delegates to ContextManager

5. **Modular tool execution**: ToolDispatcher handles the "how" of execution, Orchestrator handles the "when/why"

6. **Auto-commit on success**: Automatic version control after successful file modifications

7. **Error recovery**: Invalid outputs trigger reprompts rather than failures

8. **Optional voice**: STT/TTS are optional, system works with text-only mode

---

## Next Steps

1. Implement `internal/context` package (Context Manager)
2. Implement `internal/tools` package (Tool Dispatcher)
3. Build the Orchestrator following this design
4. Create main CLI entry point in `cmd/alfa/main.go`
5. Add Safety Layer integration for sandboxed execution
