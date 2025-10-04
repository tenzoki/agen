package orchestrator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"alfa/internal/ai"
	"alfa/internal/audio"
	alfacontext "alfa/internal/context"
	"alfa/internal/gox"
	"alfa/internal/project"
	"alfa/internal/speech"
	"alfa/internal/textpatch"
	"alfa/internal/tools"
	"alfa/internal/vcr"
	"alfa/internal/vfs"
)

// Mode represents the execution mode
type Mode int

const (
	ModeConfirm  Mode = iota // Ask before each operation
	ModeAllowAll             // Execute without confirmation
)

// Orchestrator coordinates all components
type Orchestrator struct {
	llm            ai.LLM
	contextMgr     *alfacontext.Manager
	toolDispatcher *tools.Dispatcher
	vcr            *vcr.Vcr
	projectVFS     *vfs.VFS
	projectManager *project.Manager
	workbenchRoot  string
	goxManager     *gox.Manager

	stt      speech.STT
	tts      speech.TTS
	recorder audio.Recorder
	player   audio.Player

	mode          Mode
	maxIterations int

	conversationID string
	running        bool
}

// Config holds orchestrator configuration
type Config struct {
	LLM            ai.LLM
	ContextManager *alfacontext.Manager
	ToolDispatcher *tools.Dispatcher
	VCR            *vcr.Vcr
	ProjectVFS     *vfs.VFS
	ProjectManager *project.Manager
	WorkbenchRoot  string
	GoxManager     *gox.Manager
	STT            speech.STT
	TTS            speech.TTS
	Recorder       audio.Recorder
	Player         audio.Player
	Mode           Mode
	MaxIterations  int
}

// New creates a new orchestrator instance
func New(cfg Config) *Orchestrator {
	if cfg.MaxIterations == 0 {
		cfg.MaxIterations = 10
	}

	return &Orchestrator{
		llm:            cfg.LLM,
		contextMgr:     cfg.ContextManager,
		toolDispatcher: cfg.ToolDispatcher,
		vcr:            cfg.VCR,
		projectVFS:     cfg.ProjectVFS,
		projectManager: cfg.ProjectManager,
		workbenchRoot:  cfg.WorkbenchRoot,
		goxManager:     cfg.GoxManager,
		stt:            cfg.STT,
		tts:            cfg.TTS,
		recorder:       cfg.Recorder,
		player:         cfg.Player,
		mode:           cfg.Mode,
		maxIterations:  cfg.MaxIterations,
		conversationID: generateID(),
	}
}

// Run starts the orchestrator's main interaction loop
func (o *Orchestrator) Run(ctx context.Context) error {
	o.running = true
	defer func() { o.running = false }()

	systemPrompt := o.buildSystemPrompt()

	fmt.Println("🤖 Alfa AI Coding Assistant")
	fmt.Println("Type 'exit' or 'quit' to end the session")
	fmt.Println()

	for o.running {
		userInput, err := o.getUserInput(ctx)
		if err != nil {
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
	o.contextMgr.AddUserMessage(userInput)

	iteration := 0
	for iteration < o.maxIterations {
		iteration++

		messages := o.buildMessages(systemPrompt)

		response, err := o.llm.Chat(ctx, messages)
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

			o.respond(ctx, o.formatResults(results))
			return nil
		}

		o.contextMgr.AddToolResults(results)
	}

	return fmt.Errorf("max iterations (%d) reached", o.maxIterations)
}

// getUserInput gets input from text or voice
func (o *Orchestrator) getUserInput(ctx context.Context) (string, error) {
	if o.stt != nil && o.recorder != nil {
		fmt.Print("\n🎤 Press Enter to speak (or type to use text mode): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input != "" {
			return input, nil
		}

		// Voice input mode
		fmt.Println("🔴 Recording... (speak now, will auto-stop after 2s silence)")

		// Create temp audio file
		audioPath := filepath.Join(o.projectVFS.Root(), "output", "recording.wav")
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
	fmt.Print("\n> ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// respond sends output to user
func (o *Orchestrator) respond(ctx context.Context, message string) {
	fmt.Println(message)

	if o.tts != nil && o.player != nil && message != "" {
		go func() {
			audioPath := filepath.Join(o.projectVFS.Root(), "output", "response.mp3")
			os.MkdirAll(filepath.Dir(audioPath), 0755)

			err := o.tts.SynthesizeToFile(ctx, message, audioPath)
			if err != nil {
				return
			}

			// Play audio
			o.player.PlayAsync(audioPath)
		}()
	}
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
		if o.mode == ModeConfirm {
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
	case "read_file", "write_file", "run_command", "run_tests", "search", "list_projects", "create_project", "delete_project", "restore_project", "switch_project":
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
	absPath, err := o.projectVFS.Path(filePath)
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
	fmt.Printf("\n⚠️  AI wants to: %s\n", action.Type)
	fmt.Printf("Details: %+v\n", action.Params)
	fmt.Print("Proceed? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

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
	goxAvailable := o.goxManager != nil

	capabilities := `You are an AI coding assistant with access to tools and the ability to modify code.

CAPABILITIES:
1. Read and analyze code files
2. Apply patches to fix bugs or add features
3. Run tests and commands
4. Search through the codebase
5. Generate git commits for changes`

	if goxAvailable {
		capabilities += `
6. Start and manage Gox cells for advanced workflows
7. Query RAG systems for semantic code search
8. Coordinate multi-agent processing pipelines
9. Extract named entities from text (NER) in 100+ languages
10. Anonymize sensitive data with reversible pseudonyms (GDPR compliance)`
	}

	return capabilities + `

RESPONSE FORMAT:
Always structure your responses as:
1. Natural language explanation of what you'll do
2. JSON action blocks wrapped in ` + "```json ... ```" + ` for operations

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
- switch_project: Request to switch to another project` +
	func() string {
		if goxAvailable {
			return `

Advanced Features (Gox Cells):
- start_cell: Start a Gox cell for advanced workflows
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
` + "```" + func() string {
		if goxAvailable {
			return `

GOX CELL MANAGEMENT FORMATS:

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

GUIDELINES:
- Always explain your reasoning before taking actions
- Test changes after modifications
- Keep patches focused and atomic
- Include error handling
- Follow existing code style

IMPORTANT: All file operations are sandboxed to the project directory.
You cannot access files outside the project.

Project root: ` + o.projectVFS.Root() + `
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
	o.projectVFS = newVFS
	o.vcr = newVCR

	// Update tool dispatcher with new VFS
	o.toolDispatcher = tools.NewDispatcherWithSandbox(newVFS, o.toolDispatcher.GetSandbox(), o.toolDispatcher.UsingSandbox())
	o.toolDispatcher.SetProjectManager(o.projectManager)

	// Update context manager
	o.contextMgr.SetActiveProject(projectName)

	fmt.Printf("\n✅ Switched to project '%s'\n", projectName)
	fmt.Printf("   Path: %s\n\n", projectPath)

	return nil
}

// generateID creates a unique ID
func generateID() string {
	return time.Now().Format("20060102-150405")
}
