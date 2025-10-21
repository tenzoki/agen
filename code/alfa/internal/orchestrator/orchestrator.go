package orchestrator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tenzoki/agen/alfa/internal/ai"
	"github.com/tenzoki/agen/alfa/internal/audio"
	alfacontext "github.com/tenzoki/agen/alfa/internal/context"
	"github.com/tenzoki/agen/alfa/internal/keyboard"
	"github.com/tenzoki/agen/alfa/internal/knowledge"
	"github.com/tenzoki/agen/alfa/internal/project"
	"github.com/tenzoki/agen/alfa/internal/speech"
	"github.com/tenzoki/agen/alfa/internal/textpatch"
	"github.com/tenzoki/agen/alfa/internal/tools"
	"github.com/tenzoki/agen/atomic/logging"
	"github.com/tenzoki/agen/atomic/vcr"
	"github.com/tenzoki/agen/atomic/vfs"
	cellorchestrator "github.com/tenzoki/agen/cellorg/public/orchestrator"
)

// Mode represents the execution mode
type Mode int

const (
	ModeConfirm     Mode = iota // Ask before each operation
	ModeAutoConfirm             // Execute without confirmation
)

// Orchestrator coordinates all components
type Orchestrator struct {
	llm            ai.LLM
	contextMgr     *alfacontext.Manager
	toolDispatcher *tools.Dispatcher
	vcr            *vcr.Vcr
	targetVFS      *vfs.VFS
	targetName     string
	projectManager *project.Manager
	workbenchRoot  string
	frameworkRoot  string
	cellManager    *cellorchestrator.EmbeddedOrchestrator

	stt      speech.STT
	tts      speech.TTS
	recorder audio.Recorder
	player   audio.Player

	mode            Mode
	maxIterations   int
	allowSelfModify bool

	conversationID string
	running        bool

	sessionLogger *logging.SessionLogger
}

// Config holds orchestrator configuration
type Config struct {
	LLM             ai.LLM
	ContextManager  *alfacontext.Manager
	ToolDispatcher  *tools.Dispatcher
	VCR             *vcr.Vcr
	TargetVFS       *vfs.VFS
	TargetName      string
	ProjectManager  *project.Manager
	WorkbenchRoot   string
	FrameworkRoot   string
	CellManager     *cellorchestrator.EmbeddedOrchestrator
	STT             speech.STT
	TTS             speech.TTS
	Recorder        audio.Recorder
	Player          audio.Player
	Mode            Mode
	MaxIterations   int
	AllowSelfModify bool
}

// New creates a new orchestrator instance
func New(cfg Config) *Orchestrator {
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 10
	}

	// Calculate framework root if not provided (workbench parent directory)
	frameworkRoot := cfg.FrameworkRoot
	if frameworkRoot == "" && cfg.WorkbenchRoot != "" {
		frameworkRoot = filepath.Dir(cfg.WorkbenchRoot)
	}

	// Create session logger
	// Store logs in project-specific logs directory (e.g., workbench/projects/p1/logs)
	logDir := filepath.Join(cfg.TargetVFS.Root(), "logs")
	sessionLogger, err := logging.New(logDir, true) // quietMode=true for clean CLI
	if err != nil {
		// Non-fatal: fall back to console logging
		fmt.Fprintf(os.Stderr, "Warning: Failed to create session logger: %v\n", err)
		fmt.Fprintf(os.Stderr, "Debug output will only be shown in console.\n")
	} else {
		// Set as global logger for agents to use
		logging.SetGlobalLogger(sessionLogger)
		// Session log path is now shown in main.go startup banner (if desired)

		// If cellorg is available, configure it to redirect agent output to session log
		if cfg.CellManager != nil {
			// Get the log file handle for agent output redirection
			sessionLogFile, err := os.OpenFile(sessionLogger.GetSessionPath(), os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				cfg.CellManager.SetAgentLogFile(sessionLogFile)
			}
		}
	}

	return &Orchestrator{
		llm:             cfg.LLM,
		contextMgr:      cfg.ContextManager,
		toolDispatcher:  cfg.ToolDispatcher,
		vcr:             cfg.VCR,
// Hardcoded patch for testing
		targetVFS:       cfg.TargetVFS,
		targetName:      cfg.TargetName,
		projectManager:  cfg.ProjectManager,
		workbenchRoot:   cfg.WorkbenchRoot,
		frameworkRoot:   frameworkRoot,
		cellManager:     cfg.CellManager,
		stt:             cfg.STT,
		tts:             cfg.TTS,
		recorder:        cfg.Recorder,
		player:          cfg.Player,
		mode:            cfg.Mode,
		maxIterations:   cfg.MaxIterations,
		allowSelfModify: cfg.AllowSelfModify,
		conversationID:  generateID(),
		sessionLogger:   sessionLogger,
	}
}

// Run starts the orchestrator's main interaction loop
func (o *Orchestrator) Run(ctx context.Context) error {
	o.running = true
	defer func() {
		o.running = false
		// Close session logger on exit
		if o.sessionLogger != nil {
			o.sessionLogger.Close()
		}
	}()

	systemPrompt := o.buildSystemPrompt()

	// Main interaction loop starts (banner already shown in main.go)
	for o.running {
		userInput, err := o.getUserInput(ctx)
		if err != nil {
			// Handle Ctrl+C gracefully
			if err == ErrInputCancelled {
				fmt.Println("\n👋 Goodbye!")
				return nil
			}
			return err
		}

		if userInput == "" {
			continue
		}

		if userInput == "exit" || userInput == "quit" {
			break
		}

		if userInput == "clear" {
			o.contextMgr.Clear()
			fmt.Println("✓ Context cleared")
			continue
		}

		// Log user input to session file
		if o.sessionLogger != nil {
			o.sessionLogger.LogUserInput(userInput)
		}

		err = o.processRequest(ctx, userInput, systemPrompt)
		if err != nil {
			o.respond(ctx, fmt.Sprintf("❌ Error: %v", err))
			continue
		}
	}

	fmt.Println("\n👋 Goodbye!")
	return nil
}

// processRequest handles a single user request through multiple AI iterations
func (o *Orchestrator) processRequest(ctx context.Context, userInput string, systemPrompt string) error {
	// Use PEV cell if cellorg is available
	if o.cellManager != nil {
		return o.processRequestWithPEV(ctx, userInput)
	}

	// Fallback to naive loop if cellorg not available
	return o.processRequestNaive(ctx, userInput, systemPrompt)
}

// processRequestWithPEV handles a request using the Plan-Execute-Verify cell
func (o *Orchestrator) processRequestWithPEV(ctx context.Context, userInput string) error {
	o.contextMgr.AddUserMessage(userInput)

	// Check if PEV cell is running
	cells := o.cellManager.ListCells()

	// Find PEV cell
	var pevCellRunning bool
	for _, cell := range cells {
		if cell.CellID == "alfa:plan-execute-verify" {
			pevCellRunning = true
			break
		}
	}

	if !pevCellRunning {
		// Log to session file only
		if o.sessionLogger != nil {
			o.sessionLogger.LogPEVEvent("Starting PEV cell", "Cell: alfa:plan-execute-verify")
		}

		// Start the PEV cell
		opts := cellorchestrator.CellOptions{
			ProjectID:   o.targetName,
			VFSRoot:     o.targetVFS.Root(),
			Environment: make(map[string]string),
		}

		// Pass framework root for self-modification mode
		if o.allowSelfModify {
			opts.Environment["FRAMEWORK_ROOT"] = o.frameworkRoot
		}

		// Pass session log path to agents via environment
		if o.sessionLogger != nil {
			opts.Environment["ALFA_SESSION_LOG"] = o.sessionLogger.GetSessionPath()
		}

		if err := o.cellManager.StartCell("alfa:plan-execute-verify", opts); err != nil {
			return fmt.Errorf("failed to start PEV cell: %w", err)
		}

		// Log to session file only
		if o.sessionLogger != nil {
			o.sessionLogger.LogPEVEvent("PEV cell started", "Agents initializing...")
		}
		// Give agents a moment to initialize
		time.Sleep(2 * time.Second)
	}

	// Create user request message
	requestID := generateID()
	targetContext := o.getTargetContext()

	userRequest := map[string]interface{}{
		"id":      requestID,
		"type":    "user_request",
		"content": userInput,
		"context": targetContext,
	}

	// Subscribe to BOTH alfa-responses (final) and pev-bus (progress) BEFORE publishing
	responseChan := o.cellManager.Subscribe("alfa-responses")
	defer o.cellManager.Unsubscribe("alfa-responses", responseChan)

	// Also subscribe to pev-bus to get progress updates
	progressChan := o.cellManager.Subscribe("pev-bus")
	defer o.cellManager.Unsubscribe("pev-bus", progressChan)

	// No need to echo - with readline, input is already visible in the terminal

	// Publish to pev-bus using event bridge (will be routed to agents)
	// Start animated thinking indicator
	thinking := NewThinkingIndicator()
	thinking.Start()
	defer thinking.Stop()

	if o.sessionLogger != nil {
		o.sessionLogger.LogPEVEvent("Publishing user request", fmt.Sprintf("Request ID: %s", requestID))
	}
	if err := o.cellManager.Publish("pev-bus", userRequest); err != nil {
		thinking.Stop()
		return fmt.Errorf("failed to publish user request: %w", err)
	}

	// Wait for responses with timeout
	timeout := time.After(10 * time.Minute)
	var currentIteration int

	for {
		select {
		case response := <-responseChan:
			// Handle final user response
			handled, done, err := o.handlePEVEventMessage(ctx, &response, requestID, &currentIteration)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
			if !handled {
				// Ignore messages not for this request
				continue
			}

		case progress := <-progressChan:
			// Handle progress updates (update thinking indicator)
			o.handlePEVProgressMessage(&progress, requestID, thinking)

		case <-timeout:
			return fmt.Errorf("PEV request timeout (10 minutes)")

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// getTargetContext returns context about the current target (framework vs project)
func (o *Orchestrator) getTargetContext() map[string]interface{} {
	return map[string]interface{}{
		"target_vfs":      o.targetName,
		"target_root":     o.targetVFS.Root(),
		"self_modify":     o.allowSelfModify,
		"workbench_root":  o.workbenchRoot,
		"framework_root":  o.frameworkRoot,
	}
}

// handlePEVEventMessage processes a single event from the PEV cell
// Returns (handled, done, error)
// - handled: true if this message was for our request
// - done: true if request is complete (success or failure)
func (o *Orchestrator) handlePEVEventMessage(ctx context.Context, event *cellorchestrator.Event, requestID string, currentIteration *int) (bool, bool, error) {
	// Extract payload from event data
	payload := event.Data
	if payload == nil {
		return false, false, nil
	}

	// Check if this message is for our request
	msgRequestID, ok := payload["request_id"].(string)
	if !ok || msgRequestID != requestID {
		return false, false, nil
	}

	// Get message type from payload
	msgType, _ := payload["type"].(string)

	// Handle different message types
	// Note: Alfa now only receives user_response messages on alfa-responses topic
	// All other internal PEV messages stay on pev-bus
	switch msgType {
	case "user_response":
		// Final response from coordinator (thinking indicator will be stopped by defer)
		return o.handlePEVResponse(ctx, payload)

	default:
		// Unknown message type, ignore
		return false, false, nil
	}
}

// handlePEVResponse processes the final user_response from the PEV coordinator
// Returns (handled, done, error)
func (o *Orchestrator) handlePEVResponse(ctx context.Context, payload map[string]interface{}) (bool, bool, error) {
	status, _ := payload["status"].(string)
	goalAchieved, _ := payload["goal_achieved"].(bool)
	iterations, _ := payload["iterations"].(int)
	message, _ := payload["message"].(string)

	if status == "complete" && goalAchieved {
		// Success!
		if o.sessionLogger != nil {
			o.sessionLogger.LogPEVEvent("Request completed successfully",
				fmt.Sprintf("Iterations: %d, Goal achieved: %v", iterations, goalAchieved))
			if message != "" {
				o.sessionLogger.LogAIResponse(message)
			}
		}

		// Show result to user
		if message != "" {
			fmt.Printf("\n✅ %s\n", message)
		} else {
			fmt.Println("\n✅ Request completed successfully")
		}

		// Record success in context
		o.contextMgr.AddAssistantMessage(fmt.Sprintf("Task completed successfully in %d iterations", iterations))
		return true, true, nil
	}

	if status == "failed" {
		// Failed after max iterations
		fmt.Printf("\n❌ Request failed after %d iteration(s)\n", iterations)
		if o.sessionLogger != nil {
			o.sessionLogger.LogPEVEvent("Request failed",
				fmt.Sprintf("Iterations: %d, Max iterations reached", iterations))
		}
		if message != "" {
			fmt.Println(message)
			if o.sessionLogger != nil {
				o.sessionLogger.Info("Failure message: %s", message)
			}
		}

		// Show issues if present
		if issues, ok := payload["issues"].([]interface{}); ok && len(issues) > 0 {
			fmt.Println("\nIssues encountered:")
			for _, issue := range issues {
				if issueStr, ok := issue.(string); ok {
					fmt.Printf("  • %s\n", issueStr)
					if o.sessionLogger != nil {
						o.sessionLogger.Info("Issue: %s", issueStr)
					}
				}
			}
		}

		// Show next actions if present
		if nextActions, ok := payload["next_actions"].([]interface{}); ok && len(nextActions) > 0 {
			fmt.Println("\nSuggested next actions:")
			for _, action := range nextActions {
				if actionMap, ok := action.(map[string]interface{}); ok {
					desc, _ := actionMap["description"].(string)
					priority, _ := actionMap["priority"].(string)
					fmt.Printf("  • [%s] %s\n", priority, desc)
					if o.sessionLogger != nil {
						o.sessionLogger.Info("Next action [%s]: %s", priority, desc)
					}
				}
			}
		}

		// Record failure in context
		o.contextMgr.AddAssistantMessage(fmt.Sprintf("Task failed after %d iterations. Max iterations reached.", iterations))
		return true, true, fmt.Errorf("PEV request failed after %d iterations", iterations)
	}

	// Unknown status
	return true, true, fmt.Errorf("unexpected PEV status: %s", status)
}

// handlePEVProgressMessage processes progress updates from the PEV pipeline
// Updates the thinking indicator with current phase and action details
func (o *Orchestrator) handlePEVProgressMessage(event *cellorchestrator.Event, requestID string, thinking *ThinkingIndicator) {
	payload := event.Data
	if payload == nil {
		return
	}

	// Check if this message is for our request
	msgRequestID, ok := payload["request_id"].(string)
	if !ok || msgRequestID != requestID {
		return // Not for us
	}

	// Get message type
	msgType, _ := payload["type"].(string)

	// Parse phase and detail based on message type
	var phase, detail string

	switch msgType {
	case "planning_started":
		phase = "Planning"
		if content, ok := payload["content"].(string); ok {
			detail = content
		} else {
			detail = "Analyzing request..."
		}

	case "plan_ready":
		phase = "Planning"
		detail = "Plan complete"

	case "execution_started":
		phase = "Executing"
		if action, ok := payload["action"].(string); ok {
			detail = action
		} else {
			detail = "Running actions..."
		}

	case "action_started":
		phase = "Executing"
		if actionType, ok := payload["action_type"].(string); ok {
			// Make action type more readable
			switch actionType {
			case "read_file":
				if file, ok := payload["file"].(string); ok {
					detail = fmt.Sprintf("Reading %s", file)
				} else {
					detail = "Reading file..."
				}
			case "write_file":
				if file, ok := payload["file"].(string); ok {
					detail = fmt.Sprintf("Writing %s", file)
				} else {
					detail = "Writing file..."
				}
			case "patch":
				if file, ok := payload["file"].(string); ok {
					detail = fmt.Sprintf("Patching %s", file)
				} else {
					detail = "Patching file..."
				}
			case "run_command":
				if cmd, ok := payload["command"].(string); ok {
					// Truncate long commands
					if len(cmd) > 40 {
						cmd = cmd[:40] + "..."
					}
					detail = fmt.Sprintf("Running: %s", cmd)
				} else {
					detail = "Running command..."
				}
			case "run_tests":
				detail = "Running tests..."
			default:
				detail = fmt.Sprintf("%s...", actionType)
			}
		} else {
			detail = "Executing action..."
		}

	case "action_completed":
		// Don't update on completion, wait for next action
		return

	case "verification_started":
		phase = "Verifying"
		detail = "Checking results..."

	case "verification_complete":
		phase = "Verifying"
		if success, ok := payload["success"].(bool); ok {
			if success {
				detail = "Verification passed"
			} else {
				detail = "Verification failed, retrying..."
			}
		} else {
			detail = "Complete"
		}

	default:
		// Unknown message type, don't update
		return
	}

	// Update the thinking indicator
	if phase != "" {
		thinking.UpdateProgress(phase, detail)
	}
}

// processRequestNaive is the original naive loop implementation (fallback)
func (o *Orchestrator) processRequestNaive(ctx context.Context, userInput string, systemPrompt string) error {
	o.contextMgr.AddUserMessage(userInput)

	// No need to echo - with readline, input is already visible in the terminal

	iteration := 0
	for iteration < o.maxIterations {
		iteration++

		messages := o.buildMessages(systemPrompt)

		// Show thinking indicator
		thinking := NewThinkingIndicator()
		thinking.Start()

		response, err := o.llm.Chat(ctx, messages)

		// Stop thinking indicator
		thinking.Stop()

		if err != nil {
			return fmt.Errorf("AI error: %w", err)
		}

		o.contextMgr.AddAssistantMessage(response.Content)

		actions, textResponse, err := o.parseResponse(response.Content)
		if err != nil {
			o.contextMgr.AddUserMessage("Invalid output format. Please provide valid JSON for actions.")
			continue
		}

		if len(actions) == 0 {
			o.respond(ctx, textResponse)
			return nil
		}

		results, err := o.executeActions(ctx, actions)
		if err != nil {
			return err
		}

		if o.isComplete(results) {
			if o.hasFileModifications(results) {
				commitMsg := o.generateCommitMessage(actions, results)
				o.vcr.Commit(commitMsg)
			}

			o.respondWithResults(ctx, results)
			return nil
		}

		o.contextMgr.AddToolResults(results)
	}

	return fmt.Errorf("max iterations (%d) reached", o.maxIterations)
}

// getUserInput gets input from text or voice
func (o *Orchestrator) getUserInput(ctx context.Context) (string, error) {
	// Show warning if self-modification is enabled
	if o.allowSelfModify {
		fmt.Print("⚠️  Self-modification enabled")
	}

	if o.stt != nil && o.recorder != nil {
		// For voice mode, use simple bufio (just check for Enter vs text)
		fmt.Print("\n🎤 Press Enter to speak (or type to use text mode): ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		input := strings.TrimSpace(line)

		if input != "" {
			return input, nil
		}

		// Voice input mode
		fmt.Println("🔴 Recording... (speak now, will auto-stop after 2s silence)")

		// Create temp audio file
		audioPath := filepath.Join(o.targetVFS.Root(), "output", "recording.wav")
		os.MkdirAll(filepath.Dir(audioPath), 0755)

		// Record until silence (max 30 seconds)
		err := o.recorder.RecordUntilSilence(audioPath, 30*time.Second)
		if err != nil {
			fmt.Printf("⚠️  Recording failed: %v\n", err)
			return o.getTextInput()
		}

		fmt.Println("✓ Recording complete, transcribing...")

		// Transcribe
		transcription, err := o.stt.TranscribeFile(ctx, audioPath)
		if err != nil {
			fmt.Printf("⚠️  Transcription failed: %v\n", err)
			return o.getTextInput()
		}

		fmt.Printf("You said: %s\n", transcription.Text)
		return transcription.Text, nil
	}

	return o.getTextInput()
}

func (o *Orchestrator) getTextInput() (string, error) {
	// Use readline for natural terminal input (supports multiline, history, etc.)
	input, err := getMultiLineInput("> ")
	if err != nil {
		// Propagate all errors including ErrInputCancelled (Ctrl+C should exit app)
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// respond sends output to user
func (o *Orchestrator) respond(ctx context.Context, message string) {
	fmt.Println(message)

	if o.tts != nil && o.player != nil && message != "" {
		o.speakWithInterrupt(ctx, message)
	}
}

// respondWithResults sends formatted results to user with intelligent voice output
func (o *Orchestrator) respondWithResults(ctx context.Context, results []tools.Result) {
	displayMessage := o.formatResults(results)
	fmt.Println(displayMessage)

	// For voice, prefer command output over success messages
	voiceMessage := o.extractVoiceContent(results)

	if o.tts != nil && o.player != nil && voiceMessage != "" {
		o.speakWithInterrupt(ctx, voiceMessage)
	}
}

// extractVoiceContent extracts the most relevant content for voice output
func (o *Orchestrator) extractVoiceContent(results []tools.Result) string {
	// Check if there's command output to speak
	for _, r := range results {
		if r.Success && r.Action.Type == "run_command" {
			if output, ok := r.Output.(string); ok && output != "" {
				// Clean up the output for voice (remove excessive whitespace)
				output = strings.TrimSpace(output)
				if output != "" {
					return output
				}
			}
		}
	}

	// Fallback to formatted results
	return o.formatResults(results)
}

// speakWithInterrupt speaks text with ESC key interrupt support
func (o *Orchestrator) speakWithInterrupt(ctx context.Context, text string) {
	audioPath := filepath.Join(o.targetVFS.Root(), "output", "response.mp3")
	os.MkdirAll(filepath.Dir(audioPath), 0755)

	// Generate audio file
	err := o.tts.SynthesizeToFile(ctx, text, audioPath)
	if err != nil {
		return
	}

	// Setup keyboard listener for ESC
	listener := keyboard.NewListener()
	var wg sync.WaitGroup
	stopped := false
	var mu sync.Mutex

	listener.OnEscape(func() {
		mu.Lock()
		defer mu.Unlock()
		if !stopped {
			stopped = true
			o.player.Stop()
			fmt.Println("\r[Voice stopped]                    ")
		}
	})

	// Start listener
	if err := listener.Start(); err == nil {
		defer listener.Stop()
		fmt.Print("[Press ESC to stop] ")
	}

	// Play audio in background (blocking call in goroutine)
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.player.Play(audioPath) // Use blocking Play() so we can interrupt it
	}()

	// Wait for playback to complete
	wg.Wait()

	// Clean up display
	mu.Lock()
	if !stopped {
		fmt.Print("\r                    \r")
	}
	mu.Unlock()
}

// Action represents an action to execute
type Action struct {
	Type   string                 `json:"action"`
	Params map[string]interface{} `json:"-"`
	Raw    map[string]interface{} `json:"-"`
}

// parseResponse extracts actions and text from AI response
func (o *Orchestrator) parseResponse(content string) ([]Action, string, error) {
	var actions []Action
	var textParts []string

	parts := extractCodeBlocks(content)

	for _, part := range parts {
		if part.IsCode && part.Language == "json" {
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
		Type:   actionType,
		Params: raw,
		Raw:    raw,
	}, nil
}

// executeActions runs all requested actions
func (o *Orchestrator) executeActions(ctx context.Context, actions []Action) ([]tools.Result, error) {
	var results []tools.Result

	for _, action := range actions {
		// Check current auto-confirm setting (may have changed at runtime)
		autoConfirm := o.toolDispatcher.IsAutoConfirmEnabled()

		if !autoConfirm && o.mode == ModeConfirm {
			if !o.confirmAction(action) {
				results = append(results, tools.Result{
					Action: tools.Action{
						Type:   action.Type,
						Params: action.Params,
					},
					Success: false,
					Message: "User cancelled operation",
				})
				continue
			}
		}

		result := o.executeAction(ctx, action)
		results = append(results, result)

		if !result.Success && result.Critical {
			break
		}
	}

	return results, nil
}

// executeAction executes a single action
func (o *Orchestrator) executeAction(ctx context.Context, action Action) tools.Result {
	switch action.Type {
	case "patch":
		return o.executePatch(action)
	case "read_file", "write_file", "run_command", "run_tests", "search", "list_projects", "create_project", "delete_project", "restore_project", "switch_project", "config_list", "config_get", "config_set":
		result := o.toolDispatcher.Execute(ctx, tools.Action{
			Type:   action.Type,
			Params: action.Params,
		})

		// Check if we need to switch projects
		if result.Success && (action.Type == "create_project" || action.Type == "restore_project" || action.Type == "switch_project") {
			if outputMap, ok := result.Output.(map[string]interface{}); ok {
				if switchTo, ok := outputMap["switch_to"].(string); ok && switchTo != "" {
					// Perform the project switch
					if err := o.SwitchProject(switchTo); err != nil {
						result.Success = false
						result.Message = fmt.Sprintf("%s\nFailed to switch: %v", result.Message, err)
					}
				}
			}
		}

		// Check if we deleted the active project - clear context if so
		if result.Success && action.Type == "delete_project" {
			if deletedName, ok := action.Params["name"].(string); ok {
				currentProject := o.contextMgr.GetActiveProject()
				if currentProject == deletedName {
					// Clear active project since it was just deleted
					o.contextMgr.SetActiveProject("")
					result.Message += "\n\nNote: This was your active project. Please use --project to select another project next time you start alfa."
				}
			}
		}

		return result
	default:
		return tools.Result{
			Action: tools.Action{
				Type:   action.Type,
				Params: action.Params,
			},
			Success: false,
			Message: fmt.Sprintf("unknown action type: %s", action.Type),
		}
	}
}

// executePatch applies a code patch
func (o *Orchestrator) executePatch(action Action) tools.Result {
	filePath, ok := action.Params["file"].(string)
	if !ok {
		return tools.Result{
			Success: false,
			Message: "missing 'file' parameter",
		}
	}

	operations, ok := action.Params["operations"]
	if !ok {
		return tools.Result{
			Success: false,
			Message: "missing 'operations' parameter",
		}
	}

	opsJSON, err := json.Marshal(operations)
	if err != nil {
		return tools.Result{
			Success: false,
			Message: fmt.Sprintf("invalid patch operations: %v", err),
		}
	}

	// Get absolute path within VFS
	absPath, err := o.targetVFS.Path(filePath)
	if err != nil {
		return tools.Result{
			Success: false,
			Message: fmt.Sprintf("invalid path: %v", err),
		}
	}

	err = textpatch.PatchFile(absPath, string(opsJSON))
	if err != nil {
		return tools.Result{
			Success: false,
			Message: fmt.Sprintf("patch failed: %v", err),
		}
	}

	o.contextMgr.RecordFileModification(absPath, string(opsJSON))

	return tools.Result{
		Success: true,
		Message: fmt.Sprintf("✓ Patched %s", filePath),
	}
}

// confirmAction asks user to confirm an action
func (o *Orchestrator) confirmAction(action Action) bool {
	fmt.Printf("\n⚠️  Alfa wants to: %s\n", action.Type)
	fmt.Printf("Details: %+v\n", action.Params)
	fmt.Print("Proceed? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	response := strings.TrimSpace(strings.ToLower(input))

	return response == "" || response == "y" || response == "yes"
}

// isComplete checks if all actions completed successfully
func (o *Orchestrator) isComplete(results []tools.Result) bool {
	for _, r := range results {
		if !r.Success {
			return false
		}
	}
	return true
}

// hasFileModifications checks if any file was modified
func (o *Orchestrator) hasFileModifications(results []tools.Result) bool {
	for _, r := range results {
		if r.Success && (r.Action.Type == "patch" || r.Action.Type == "write_file") {
			return true
		}
	}
	return false
}

// formatResults formats execution results for display
func (o *Orchestrator) formatResults(results []tools.Result) string {
	var lines []string
	for _, r := range results {
		if r.Success {
			lines = append(lines, "✓ "+r.Message)
		} else {
			lines = append(lines, "✗ "+r.Message)
		}

		// Show Output if present and non-empty
		if r.Output != nil {
			if outputStr, ok := r.Output.(string); ok && outputStr != "" {
				lines = append(lines, outputStr)
			}
		}
	}
	return strings.Join(lines, "\n")
}

// generateCommitMessage creates a commit message from actions
func (o *Orchestrator) generateCommitMessage(actions []Action, results []tools.Result) string {
	var operations []string
	var files []string

	for i, result := range results {
		if !result.Success {
			continue
		}

		action := actions[i]
		switch action.Type {
		case "patch":
			if file, ok := action.Params["file"].(string); ok {
				files = append(files, file)
				operations = append(operations, fmt.Sprintf("modified %s", file))
			}
		case "write_file":
			if file, ok := action.Params["path"].(string); ok {
				files = append(files, file)
				operations = append(operations, fmt.Sprintf("created %s", file))
			}
		}
	}

	if len(operations) == 0 {
		return "AI assistant changes"
	}

	summary := strings.Join(operations, ", ")
	return fmt.Sprintf("AI: %s", summary)
}

// buildSystemPrompt constructs the system prompt
func (o *Orchestrator) buildSystemPrompt() string {
	cellorgAvailable := o.cellManager != nil

	capabilities := `You are an AI coding assistant with access to tools and the ability to modify code.

RESPONSE STYLE:
- Be CONCISE and DIRECT - avoid unnecessary explanations
- For simple operations (config changes, file reads, commands): minimal explanation, just do it
- For complex changes (code modifications, refactoring): brief explanation of approach
- Let the tool execution results speak for themselves
- Only elaborate when explicitly asked or when complexity requires it

CAPABILITIES:
1. Read and analyze code files
2. Apply patches to fix bugs or add features
3. Run tests and commands
4. Search through the codebase
5. Generate git commits for changes`

	capNum := 6
	if o.allowSelfModify {
		capabilities += fmt.Sprintf(`
%d. **SELF-MODIFICATION**: Modify the framework's own codebase (code/ directory)`, capNum)
		capNum++
	}

	if cellorgAvailable {
		capabilities += fmt.Sprintf(`
%d. Start and manage cells for advanced workflows
%d. Query RAG systems for semantic code search
%d. Coordinate multi-agent processing pipelines
%d. Extract named entities from text (NER) in 100+ languages
%d. Anonymize sensitive data with reversible pseudonyms (GDPR compliance)`, capNum, capNum+1, capNum+2, capNum+3, capNum+4)
	}

	// Add examples based on target type
	var examples string
	if o.allowSelfModify {
		examples = `

SELF-MODIFICATION MODE - CRITICAL UNDERSTANDING:

You are ALFA, an AI coding assistant.
YOUR source code is located in: code/alfa/
When the user says:
- "your code" / "deinen code"
- "alfa code" / "alfa's code"
- "modify yourself" / "modifiziere dich"
- "the framework" / "das framework"

→ They mean: Modify files in code/alfa/, code/atomic/, code/omni/, etc.

FRAMEWORK OPERATIONS - WORKFLOW:

When asked to modify framework code, follow these steps:
1. Search for the relevant code (optional, if you don't know where it is)
2. Read the file(s) to understand current implementation
3. Plan the changes
4. Apply patch(es) to modify the code
5. Test if needed

Example workflow - "Add feature X to alfa":

Step 1 - Search (if needed):
` + "```json" + `
{"action": "search", "query": "relevant keyword", "pattern": "*.go"}
` + "```" + `

Step 2 - Read the file:
` + "```json" + `
{"action": "read_file", "path": "code/alfa/internal/orchestrator/orchestrator.go"}
` + "```" + `

Step 3 - Modify the code:
` + "```json" + `
{"action": "patch", "file": "code/alfa/internal/orchestrator/orchestrator.go", "operations": [
  {"line": 100, "type": "insert", "content": ["// New code here"]}
]}
` + "```" + `

Step 4 - Test:
` + "```json" + `
{"action": "run_tests", "pattern": "./code/alfa/..."}
` + "```" + `

IMPORTANT: Don't stop after search! Continue with read_file and patch.
You can directly read and modify ANY file in code/alfa/, code/atomic/, code/omni/, etc.
All paths are relative to TARGET root.`
	} else {
		examples = `

PROJECT OPERATIONS - EXAMPLES:

Read project file:
` + "```json" + `
{"action": "read_file", "path": "main.go"}
` + "```" + `

Create new file:
` + "```json" + `
{"action": "write_file", "path": "utils.go", "content": "package main\n..."}
` + "```" + `

All paths are relative to TARGET root (your project directory).`
	}

	contextInfo := fmt.Sprintf(`

═══════════════════════════════════════════════════
TARGET CONTEXT
═══════════════════════════════════════════════════

Target: %s
Root:   %s

All file operations (read_file, write_file, patch) operate on TARGET.
Commands (run_command, run_tests) execute in TARGET directory.
%s

═══════════════════════════════════════════════════
`, o.targetName, o.targetVFS.Root(), examples)

	return capabilities + contextInfo + `

RESPONSE FORMAT:
Structure your responses as:
1. Brief statement of what you're doing (1 line for simple ops, 2-3 lines for complex)
2. JSON action blocks wrapped in ` + "```json ... ```" + ` for operations
3. No verbose explanations unless complexity requires it

Examples:
Simple operation: "Enabling auto-confirm mode."
Complex operation: "I'll refactor the authentication logic to use middleware. This will make the code more maintainable and easier to test."

AVAILABLE ACTIONS:

Basic Operations:
- patch: Apply code changes
- read_file: Read file contents
- write_file: Create new file
- run_command: Execute shell command
- run_tests: Run test suite
- search: Search codebase

Project Management:
- list_projects: List all projects in workbench
- create_project: Create a new project
- delete_project: Delete a project (backup kept for recovery)
- restore_project: Restore a deleted project from backup
- switch_project: Request to switch to another project

Configuration Management:
- config_list: List all current configuration settings
- config_get: Get a specific configuration setting
- config_set: Update a configuration setting (saved to alfa.yaml)` +
	func() string {
		if cellorgAvailable {
			return `

Advanced Features (Cellorg Cells):
- start_cell: Start a cell for advanced workflows
- stop_cell: Stop a running cell
- list_cells: List all running cells
- query_cell: Send a query to a cell and wait for response

NER & Anonymization:
- extract_entities: Extract named entities (PERSON, ORG, LOC) from text
- anonymize_text: Replace sensitive entities with reversible pseudonyms
- deanonymize_text: Restore original text from anonymized version`
		}
		return ""
	}() + `

PATCH ACTION FORMAT:
` + "```json" + `
{
  "action": "patch",
  "file": "path/to/file.go",
  "operations": [
    {"line": 10, "type": "insert", "content": ["new line 1", "new line 2"]},
    {"line": 15, "type": "delete"},
    {"line": 20, "type": "replace", "content": ["replacement line"]}
  ]
}
` + "```" + `

READ FILE FORMAT:
` + "```json" + `
{
  "action": "read_file",
  "path": "path/to/file.go"
}
` + "```" + `

RUN TESTS FORMAT:
` + "```json" + `
{
  "action": "run_tests",
  "pattern": "./..."
}
` + "```" + `

PROJECT MANAGEMENT FORMATS:
` + "```json" + `
{
  "action": "list_projects"
}
` + "```" + `

` + "```json" + `
{
  "action": "create_project",
  "name": "myproject"
}
` + "```" + `
(Note: After creating a project, you are automatically switched to it and can start working immediately)

` + "```json" + `
{
  "action": "delete_project",
  "name": "old-project"
}
` + "```" + `
(Note: Project files are deleted but backup is kept in .git-remotes/ for recovery)

` + "```json" + `
{
  "action": "restore_project",
  "name": "deleted-project"
}
` + "```" + `
(Note: Restores from .git-remotes/ backup and automatically switches to it)

` + "```json" + `
{
  "action": "switch_project",
  "name": "existing-project"
}
` + "```" + `
(Note: Project switching happens in real-time - no restart needed)

IMPORTANT: After creating or switching projects, all subsequent operations will work on the new project automatically.

CONFIGURATION MANAGEMENT FORMATS:

List all configuration settings:
` + "```json" + `
{
  "action": "config_list"
}
` + "```" + `

Get a specific setting:
` + "```json" + `
{
  "action": "config_get",
  "key": "execution.max_iterations"
}
` + "```" + `

Update a setting (saved to alfa.yaml):
` + "```json" + `
{
  "action": "config_set",
  "key": "execution.auto_confirm",
  "value": "true"
}
` + "```" + `

Available configuration keys:
- workbench.path, workbench.project
- ai.provider (anthropic/openai), ai.config_file
- voice.input_enabled (true/false), voice.output_enabled (true/false), voice.headless (true/false)
- execution.auto_confirm (true/false), execution.max_iterations (number)
- sandbox.enabled (true/false), sandbox.image (string)
- cellorg.enabled (true/false), cellorg.config_path (string)
- output.capture_enabled (true/false), output.max_size_kb (number)
- self_modify.allowed (true/false)

(Note: Some settings require restart, others apply immediately. The AI will indicate which.)
` + "```" + func() string {
		if cellorgAvailable {
			return `

CELLORG CELL MANAGEMENT FORMATS:

Start a cell:
` + "```json" + `
{
  "action": "start_cell",
  "cell_id": "rag:knowledge-backend",
  "project_id": "my-project",
  "environment": {
    "OPENAI_API_KEY": "sk-..."
  }
}
` + "```" + `

Stop a cell:
` + "```json" + `
{
  "action": "stop_cell",
  "cell_id": "rag:knowledge-backend",
  "project_id": "my-project"
}
` + "```" + `

List running cells:
` + "```json" + `
{
  "action": "list_cells"
}
` + "```" + `

Query a cell (RAG, processing, etc.):
` + "```json" + `
{
  "action": "query_cell",
  "project_id": "my-project",
  "query": "find authentication code",
  "params": {
    "top_k": 5
  },
  "timeout": 10
}
` + "```" + `

IMPORTANT CELL PATTERNS:
- ALWAYS use cells as functional units (not individual agents)
- Each cell is a network of agents working together
- Cells are isolated per project via VFS
- File patterns in cells are relative to project root
- Common cells: "rag:knowledge-backend" for code search

Example workflow with cells:
1. Start a RAG cell for the project
2. Query the cell for relevant code context
3. Use the context to make informed code changes
4. Stop the cell when done

` + "```"
		}
		return ""
	}() + `

` + func() string {
		// Load core framework rules if self-modification is enabled
		if !o.allowSelfModify {
			return `
═══════════════════════════════════════════════════
FRAMEWORK MODIFICATION RESTRICTIONS
═══════════════════════════════════════════════════

IMPORTANT: You are CURRENTLY NOT allowed to modify framework code (code/ directory).
- You can read framework code for understanding
- You can work on user projects (workbench/projects/)
- Framework code modifications are disabled

HOW TO ENABLE SELF-MODIFICATION:
If the user asks about modifying the framework or your own code, inform them:
"I'm currently not allowed to modify the framework code (code/ directory).
To enable self-modification, you can:
1. Set the config: self_modify.allowed = true
2. Or restart alfa with --allow-self-modify flag

Would you like me to enable self-modification now?"

═══════════════════════════════════════════════════

`
		}

		coreRulesPath := filepath.Join(o.workbenchRoot, "../guidelines/core-rules.txt")
		content, err := os.ReadFile(coreRulesPath)
		if err != nil {
			// Core rules not found - not critical, continue without them
			return `
═══════════════════════════════════════════════════
SELF-MODIFICATION ENABLED
═══════════════════════════════════════════════════

You ARE ALLOWED to modify the framework code (code/ directory).
This includes:
- Alfa's own codebase (code/alfa/)
- Cellorg orchestrator (code/cellorg/)
- Atomic libraries (code/atomic/)
- Omni storage (code/omni/)
- Agents (code/agents/)

IMPORTANT: Exercise caution and follow best practices.
If core rules are missing, ask the user before making significant changes.

═══════════════════════════════════════════════════

`
		}
		return `
═══════════════════════════════════════════════════
SELF-MODIFICATION ENABLED - FRAMEWORK MODIFICATION RULES
═══════════════════════════════════════════════════

You ARE ALLOWED to modify the framework code (code/ directory).
This includes:
- Alfa's own codebase (code/alfa/)
- Cellorg orchestrator (code/cellorg/)
- Atomic libraries (code/atomic/)
- Omni storage (code/omni/)
- Agents (code/agents/)

` + string(content) + `

BEFORE modifying framework code (code/), ALWAYS:
1. Use read_file to consult relevant reference:
   - Architecture changes: read_file("guidelines/references/architecture.md")
   - Adding agent: read_file("guidelines/references/agent-patterns.md")
   - Config changes: read_file("guidelines/references/config-standards.md")
2. Verify compliance with core rules above
3. If uncertain, ask the user before proceeding

═══════════════════════════════════════════════════

`
	}() + func() string {
		// Load knowledge base if cellorg is enabled
		if !cellorgAvailable {
			return ""
		}

		frameworkRoot := filepath.Dir(o.workbenchRoot)
		extractor := knowledge.NewExtractor(frameworkRoot)

		// Try to load generated knowledge docs
		capabilities, err := extractor.LoadCapabilitiesDoc()
		if err != nil {
			return "" // Knowledge not yet extracted
		}

		agents, err := extractor.LoadAgentsDoc()
		if err != nil {
			return ""
		}

		cells, err := extractor.LoadCellsDoc()
		if err != nil {
			return ""
		}

		return `
═══════════════════════════════════════════════════
AGEN FRAMEWORK KNOWLEDGE BASE
═══════════════════════════════════════════════════

` + capabilities + `

═══════════════════════════════════════════════════
AVAILABLE AGENTS
═══════════════════════════════════════════════════

` + agents + `

═══════════════════════════════════════════════════
AVAILABLE CELLS
═══════════════════════════════════════════════════

` + cells + `

═══════════════════════════════════════════════════

`
	}() + `
GUIDELINES:
- Always explain your reasoning before taking actions
- Test changes after modifications
- Keep patches focused and atomic
- Include error handling
- Follow existing code style

IMPORTANT: All file operations are sandboxed to the project directory.
You cannot access files outside the target directory.

TARGET: ` + o.targetName + `
Target root: ` + o.targetVFS.Root() + `

NOTE: All file operations operate within this TARGET directory.
`
}

// buildMessages constructs the message list for AI
func (o *Orchestrator) buildMessages(systemPrompt string) []ai.Message {
	messages := []ai.Message{
		{Role: "system", Content: systemPrompt},
	}

	history := o.contextMgr.GetRecentHistory(20)
	messages = append(messages, history...)

	return messages
}

// CodeBlock represents a parsed code block
type CodeBlock struct {
	IsCode   bool
	Language string
	Content  string
}

// extractCodeBlocks parses markdown code blocks from text
func extractCodeBlocks(content string) []CodeBlock {
	var blocks []CodeBlock
	lines := strings.Split(content, "\n")

	var currentBlock *CodeBlock
	var blockLines []string

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if currentBlock != nil {
				// End of code block
				currentBlock.Content = strings.Join(blockLines, "\n")
				blocks = append(blocks, *currentBlock)
				currentBlock = nil
				blockLines = nil
			} else {
				// Start of code block
				lang := strings.TrimPrefix(line, "```")
				lang = strings.TrimSpace(lang)
				currentBlock = &CodeBlock{
					IsCode:   true,
					Language: lang,
				}
			}
		} else {
			if currentBlock != nil {
				blockLines = append(blockLines, line)
			} else {
				// Regular text
				if len(blockLines) > 0 {
					blocks = append(blocks, CodeBlock{
						IsCode:  false,
						Content: strings.Join(blockLines, "\n"),
					})
					blockLines = nil
				}
				blockLines = append(blockLines, line)
			}
		}
	}

	// Add remaining text
	if len(blockLines) > 0 {
		blocks = append(blocks, CodeBlock{
			IsCode:  false,
			Content: strings.Join(blockLines, "\n"),
		})
	}

	return blocks
}

// SwitchProject switches to a different project at runtime
func (o *Orchestrator) SwitchProject(projectName string) error {
	if o.allowSelfModify {
		return fmt.Errorf("project switching not available in self-modification mode (target is framework)")
	}

	if o.projectManager == nil {
		return fmt.Errorf("project manager not available")
	}

	if !o.projectManager.Exists(projectName) {
		return fmt.Errorf("project '%s' does not exist", projectName)
	}

	// Get new project path
	projectPath := o.projectManager.GetProjectPath(projectName)

	// Create new VFS for the project
	newVFS, err := vfs.NewVFS(projectPath, false)
	if err != nil {
		return fmt.Errorf("failed to create VFS for project '%s': %w", projectName, err)
	}

	// Create new VCR for the project
	newVCR := vcr.NewVcr("assistant", projectPath)

	// Update orchestrator state
	o.targetVFS = newVFS
	o.targetName = projectName
	o.vcr = newVCR

	// Update tool dispatcher with new VFS
	o.toolDispatcher = tools.NewDispatcherWithSandbox(newVFS, o.toolDispatcher.GetSandbox(), o.toolDispatcher.UsingSandbox())
	o.toolDispatcher.SetProjectManager(o.projectManager)

	// Update context manager
	o.contextMgr.SetActiveProject(projectName)

	fmt.Printf("\n✅ Switched to project '%s'\n", projectName)
	fmt.Printf("   Target: %s\n", projectName)
	fmt.Printf("   Path:   %s\n\n", projectPath)

	return nil
}

// generateID creates a unique ID
func generateID() string {
	return time.Now().Format("20060102-150405")
}
