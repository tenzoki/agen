package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	alfaconfig "github.com/tenzoki/agen/alfa/internal/config"
	"github.com/tenzoki/agen/alfa/internal/project"
	"github.com/tenzoki/agen/alfa/internal/sandbox"
	"github.com/tenzoki/agen/atomic/vfs"
	cellorchestrator "github.com/tenzoki/agen/cellorg/public/orchestrator"
)

// VoiceController interface for enabling/disabling voice
type VoiceController interface {
	EnableVoice() error
	EnableVoiceInput() error
	EnableVoiceOutput() error
	DisableVoice()
	DisableVoiceInput()
	DisableVoiceOutput()
}

// Dispatcher executes tool operations
type Dispatcher struct {
	vfs             *vfs.VFS
	sandbox         sandbox.Sandbox
	projectManager  *project.Manager
	cellManager     *cellorchestrator.EmbeddedOrchestrator
	config          *alfaconfig.AlfaConfig
	configPath      string
	voiceController VoiceController
	timeout         time.Duration
	useSandbox      bool
	captureOutput   bool
	maxOutputBytes  int
}

// NewDispatcher creates a new tool dispatcher with VFS
func NewDispatcher(projectVFS *vfs.VFS) *Dispatcher {
	return NewDispatcherWithSandbox(projectVFS, nil, false)
}

// NewDispatcherWithSandbox creates a dispatcher with optional sandbox
func NewDispatcherWithSandbox(projectVFS *vfs.VFS, sb sandbox.Sandbox, useSandbox bool) *Dispatcher {
	return &Dispatcher{
		vfs:            projectVFS,
		sandbox:        sb,
		timeout:        30 * time.Second,
		useSandbox:     useSandbox,
		captureOutput:  true,        // Capture output by default
		maxOutputBytes: 10 * 1024,   // 10KB default limit
	}
}

// SetOutputCapture configures output capture behavior
func (d *Dispatcher) SetOutputCapture(enabled bool, maxBytes int) {
	d.captureOutput = enabled
	if maxBytes > 0 {
		d.maxOutputBytes = maxBytes
	}
}

// processOutput handles output according to capture settings
func (d *Dispatcher) processOutput(output string) interface{} {
	if !d.captureOutput {
		return nil
	}

	if len(output) <= d.maxOutputBytes {
		return output
	}

	// Truncate and add notice
	truncated := output[:d.maxOutputBytes]
	remaining := len(output) - d.maxOutputBytes
	return fmt.Sprintf("%s\n\n[Output truncated: %d more bytes not shown. Use --max-output to increase limit.]",
		truncated, remaining)
}

// SetProjectManager sets the project manager for project operations
func (d *Dispatcher) SetProjectManager(pm *project.Manager) {
	d.projectManager = pm
}

// SetCellManager sets the cell manager for cell operations
func (d *Dispatcher) SetCellManager(cm *cellorchestrator.EmbeddedOrchestrator) {
	d.cellManager = cm
}

// SetConfig sets the configuration and config path for runtime config management
func (d *Dispatcher) SetConfig(cfg *alfaconfig.AlfaConfig, configPath string) {
	d.config = cfg
	d.configPath = configPath
}

// SetVoiceController sets the voice controller for runtime voice management
func (d *Dispatcher) SetVoiceController(vc VoiceController) {
	d.voiceController = vc
}

// IsAutoConfirmEnabled returns the current auto-confirm setting
func (d *Dispatcher) IsAutoConfirmEnabled() bool {
	if d.config == nil {
		return false
	}
	return d.config.Execution.AutoConfirm
}

// GetSandbox returns the sandbox instance
func (d *Dispatcher) GetSandbox() sandbox.Sandbox {
	return d.sandbox
}

// UsingSandbox returns whether sandbox is enabled
func (d *Dispatcher) UsingSandbox() bool {
	return d.useSandbox
}

// Action represents a tool action to execute
type Action struct {
	Type   string
	Params map[string]interface{}
}

// Result represents the outcome of a tool execution
type Result struct {
	Action   Action
	Success  bool
	Message  string
	Output   interface{}
	Critical bool
}

// Execute runs a tool action and returns the result
func (d *Dispatcher) Execute(ctx context.Context, action Action) Result {
	switch action.Type {
	case "read_file":
		return d.executeReadFile(action)
	case "write_file":
		return d.executeWriteFile(action)
	case "run_command":
		return d.executeRunCommand(ctx, action)
	case "run_tests":
		return d.executeRunTests(ctx, action)
	case "search":
		return d.executeSearch(action)
	case "list_projects":
		return d.executeListProjects(action)
	case "create_project":
		return d.executeCreateProject(action)
	case "delete_project":
		return d.executeDeleteProject(action)
	case "restore_project":
		return d.executeRestoreProject(action)
	case "switch_project":
		return d.executeSwitchProject(action)
	case "start_cell":
		return d.executeStartCell(ctx, action)
	case "stop_cell":
		return d.executeStopCell(action)
	case "list_cells":
		return d.executeListCells(action)
	case "query_cell":
		return d.executeQueryCell(ctx, action)
	case "extract_entities":
		return d.executeExtractEntities(ctx, action)
	case "anonymize_text":
		return d.executeAnonymizeText(ctx, action)
	case "deanonymize_text":
		return d.executeDeanonymizeText(ctx, action)
	case "config_get":
		return d.executeConfigGet(action)
	case "config_set":
		return d.executeConfigSet(action)
	case "config_list":
		return d.executeConfigList(action)
	default:
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("unknown action type: %s", action.Type),
		}
	}
}

// executeReadFile reads a file's contents
func (d *Dispatcher) executeReadFile(action Action) Result {
	filePath, ok := action.Params["path"].(string)
	if !ok {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'path' parameter",
		}
	}

	content, err := d.vfs.ReadString(filePath)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to read file: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Read %s (%d bytes)", filePath, len(content)),
		Output:  content,
	}
}

// executeWriteFile creates or overwrites a file
func (d *Dispatcher) executeWriteFile(action Action) Result {
	filePath, ok := action.Params["path"].(string)
	if !ok {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'path' parameter",
		}
	}

	content, ok := action.Params["content"].(string)
	if !ok {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'content' parameter",
		}
	}

	if err := d.vfs.WriteString(content, filePath); err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to write file: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Wrote %s (%d bytes)", filePath, len(content)),
	}
}

// executeRunCommand runs a shell command
func (d *Dispatcher) executeRunCommand(ctx context.Context, action Action) Result {
	command, ok := action.Params["command"].(string)
	if !ok {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'command' parameter",
		}
	}

	// Use sandbox if enabled and available
	if d.useSandbox && d.sandbox != nil && d.sandbox.IsAvailable() {
		return d.executeRunCommandSandboxed(ctx, action, command)
	}

	// Fallback to direct execution
	cmdCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)
	cmd.Dir = d.vfs.Root()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Always print output to user (regardless of capture setting)
	if len(outputStr) > 0 {
		fmt.Print(outputStr)
	}

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("command failed: %v", err),
			Output:  d.processOutput(outputStr),
		}
	}

	message := "Command executed successfully"
	if !d.captureOutput {
		message = "Command executed successfully (output not captured)"
	} else if len(outputStr) > d.maxOutputBytes {
		message = fmt.Sprintf("Command executed successfully (output truncated to %d bytes)", d.maxOutputBytes)
	}

	return Result{
		Action:  action,
		Success: true,
		Message: message,
		Output:  d.processOutput(outputStr),
	}
}

// executeRunCommandSandboxed runs command in Docker sandbox
func (d *Dispatcher) executeRunCommandSandboxed(ctx context.Context, action Action, command string) Result {
	req := sandbox.ExecuteRequest{
		Command:  command,
		WorkDir:  d.vfs.Root(),
		Timeout:  d.timeout,
		CPULimit: 1.0,
		MemoryMB: 512,
	}

	result, err := d.sandbox.Execute(ctx, req)
	outputStr := result.Stdout + "\n" + result.Stderr

	// Always print output to user
	if len(result.Stdout) > 0 {
		fmt.Print(result.Stdout)
	}
	if len(result.Stderr) > 0 {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("sandboxed execution failed: %v", err),
			Output:  d.processOutput(outputStr),
		}
	}

	if result.ExitCode != 0 {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("command exited with code %d", result.ExitCode),
			Output:  d.processOutput(outputStr),
		}
	}

	message := fmt.Sprintf("Sandboxed execution completed in %v", result.Duration)
	if !d.captureOutput {
		message += " (output not captured)"
	} else if len(result.Stdout) > d.maxOutputBytes {
		message = fmt.Sprintf("Sandboxed execution completed in %v (output truncated to %d bytes)",
			result.Duration, d.maxOutputBytes)
	}

	return Result{
		Action:  action,
		Success: true,
		Message: message,
		Output:  d.processOutput(result.Stdout),
	}
}

// executeRunTests runs the test suite
func (d *Dispatcher) executeRunTests(ctx context.Context, action Action) Result {
	pattern, ok := action.Params["pattern"].(string)
	if !ok {
		pattern = "./..."
	}

	testTimeout := d.timeout * 2 // Tests get more time

	// Use sandbox if enabled and available
	if d.useSandbox && d.sandbox != nil && d.sandbox.IsAvailable() {
		return d.executeRunTestsSandboxed(ctx, action, pattern, testTimeout)
	}

	// Fallback to direct execution
	cmdCtx, cancel := context.WithTimeout(ctx, testTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "go", "test", "-v", pattern)
	cmd.Dir = d.vfs.Root()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Always print output to user
	if len(outputStr) > 0 {
		fmt.Print(outputStr)
	}

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Tests failed",
			Output:  d.processOutput(outputStr),
		}
	}

	message := "All tests passed"
	if !d.captureOutput {
		message = "All tests passed (output not captured)"
	} else if len(outputStr) > d.maxOutputBytes {
		message = fmt.Sprintf("All tests passed (output truncated to %d bytes)", d.maxOutputBytes)
	}

	return Result{
		Action:  action,
		Success: true,
		Message: message,
		Output:  d.processOutput(outputStr),
	}
}

// executeRunTestsSandboxed runs tests in Docker sandbox
func (d *Dispatcher) executeRunTestsSandboxed(ctx context.Context, action Action, pattern string, timeout time.Duration) Result {
	req := sandbox.ExecuteRequest{
		Command:  fmt.Sprintf("go test -v %s", pattern),
		WorkDir:  d.vfs.Root(),
		Timeout:  timeout,
		CPULimit: 2.0, // Tests can use more CPU
		MemoryMB: 1024, // Tests get more memory
	}

	result, err := d.sandbox.Execute(ctx, req)
	outputStr := result.Stdout + "\n" + result.Stderr

	// Always print output to user
	if len(result.Stdout) > 0 {
		fmt.Print(result.Stdout)
	}
	if len(result.Stderr) > 0 {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("sandboxed test execution failed: %v", err),
			Output:  d.processOutput(outputStr),
		}
	}

	if result.ExitCode != 0 {
		return Result{
			Action:  action,
			Success: false,
			Message: "Tests failed",
			Output:  d.processOutput(outputStr),
		}
	}

	message := fmt.Sprintf("All tests passed (sandboxed, %v)", result.Duration)
	if !d.captureOutput {
		message += " (output not captured)"
	} else if len(result.Stdout) > d.maxOutputBytes {
		message = fmt.Sprintf("All tests passed (sandboxed, %v, output truncated to %d bytes)",
			result.Duration, d.maxOutputBytes)
	}

	return Result{
		Action:  action,
		Success: true,
		Message: message,
		Output:  d.processOutput(result.Stdout),
	}
}

// executeSearch searches for text in files
func (d *Dispatcher) executeSearch(action Action) Result {
	query, ok := action.Params["query"].(string)
	if !ok {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'query' parameter",
		}
	}

	pattern, ok := action.Params["pattern"].(string)
	if !ok {
		pattern = "*.go"
	}

	// Find matching files
	var matches []string
	err := d.vfs.Walk(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			// Skip hidden and vendor directories
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches pattern
		matched, _ := filepath.Match(pattern, info.Name())
		if !matched {
			return nil
		}

		// Read and search file
		relPath, _ := filepath.Rel(d.vfs.Root(), path)
		content, err := d.vfs.ReadString(relPath)
		if err != nil {
			return nil
		}

		if strings.Contains(content, query) {
			matches = append(matches, relPath)
		}

		return nil
	})

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("search failed: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Found %d matches", len(matches)),
		Output:  matches,
	}
}

// executeListProjects lists all projects in the workbench
func (d *Dispatcher) executeListProjects(action Action) Result {
	if d.projectManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "project manager not available",
		}
	}

	projects, err := d.projectManager.List()
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to list projects: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Found %d projects", len(projects)),
		Output:  projects,
	}
}

// executeCreateProject creates a new project
func (d *Dispatcher) executeCreateProject(action Action) Result {
	if d.projectManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "project manager not available",
		}
	}

	name, ok := action.Params["name"].(string)
	if !ok || name == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'name' parameter",
		}
	}

	if err := d.projectManager.Create(name); err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to create project: %v", err),
		}
	}

	meta, _ := d.projectManager.GetMetadata(name)

	// Mark as needing project switch
	result := Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Project '%s' created successfully. Switching to it now...", name),
		Output: map[string]interface{}{
			"project_name": name,
			"metadata":     meta,
			"switch_to":    name,
		},
		Critical: true,
	}

	return result
}

// executeDeleteProject deletes a project (keeps backup)
func (d *Dispatcher) executeDeleteProject(action Action) Result {
	if d.projectManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "project manager not available",
		}
	}

	name, ok := action.Params["name"].(string)
	if !ok || name == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'name' parameter",
		}
	}

	if !d.projectManager.Exists(name) {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("project '%s' does not exist", name),
		}
	}

	// Delete the project
	if err := d.projectManager.Delete(name); err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to delete project: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Project '%s' deleted successfully. Backup kept in .git-remotes/ for recovery.", name),
		Output: map[string]interface{}{
			"deleted_project": name,
			"can_restore":     true,
		},
		Critical: false,
	}
}

// executeRestoreProject restores a deleted project from backup
func (d *Dispatcher) executeRestoreProject(action Action) Result {
	if d.projectManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "project manager not available",
		}
	}

	name, ok := action.Params["name"].(string)
	if !ok || name == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'name' parameter",
		}
	}

	if d.projectManager.Exists(name) {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("project '%s' already exists - cannot restore over existing project", name),
		}
	}

	// Restore the project
	if err := d.projectManager.Restore(name); err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to restore project: %v", err),
		}
	}

	meta, _ := d.projectManager.GetMetadata(name)

	// Mark as needing project switch
	result := Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Project '%s' restored successfully. Switching to it now...", name),
		Output: map[string]interface{}{
			"project_name": name,
			"metadata":     meta,
			"switch_to":    name,
		},
		Critical: true,
	}

	return result
}

// executeSwitchProject requests a project switch
func (d *Dispatcher) executeSwitchProject(action Action) Result {
	if d.projectManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "project manager not available",
		}
	}

	name, ok := action.Params["name"].(string)
	if !ok || name == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'name' parameter",
		}
	}

	if !d.projectManager.Exists(name) {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("project '%s' does not exist", name),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Switching to project '%s'...", name),
		Output: map[string]interface{}{
			"switch_to": name,
		},
		Critical: true,
	}
}

// executeStartCell starts a cell for a project
func (d *Dispatcher) executeStartCell(ctx context.Context, action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Cellorg is not available. Advanced features are disabled.",
		}
	}

	cellID, ok := action.Params["cell_id"].(string)
	if !ok || cellID == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'cell_id' parameter",
		}
	}

	projectID, ok := action.Params["project_id"].(string)
	if !ok || projectID == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'project_id' parameter",
		}
	}

	// Get VFS root for the project
	vfsRoot := d.vfs.Root()

	// Extract environment variables if provided
	env := make(map[string]string)
	if envParam, ok := action.Params["environment"].(map[string]interface{}); ok {
		for k, v := range envParam {
			if str, ok := v.(string); ok {
				env[k] = str
			}
		}
	}

	// Create cell options
	opts := cellorchestrator.CellOptions{
		ProjectID:   projectID,
		VFSRoot:     vfsRoot,
		Environment: env,
	}

	// Start cell
	err := d.cellManager.StartCell(cellID, opts)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to start cell: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Cell '%s' started for project '%s'", cellID, projectID),
		Output: map[string]interface{}{
			"cell_id":    cellID,
			"project_id": projectID,
			"vfs_root":   vfsRoot,
		},
	}
}

// executeStopCell stops a running cell
func (d *Dispatcher) executeStopCell(action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Cellorg is not available",
		}
	}

	cellID, ok := action.Params["cell_id"].(string)
	if !ok || cellID == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'cell_id' parameter",
		}
	}

	projectID, ok := action.Params["project_id"].(string)
	if !ok || projectID == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'project_id' parameter",
		}
	}

	err := d.cellManager.StopCell(cellID, projectID)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to stop cell: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Cell '%s' stopped for project '%s'", cellID, projectID),
	}
}

// executeListCells lists all running cells
func (d *Dispatcher) executeListCells(action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: true,
			Message: "No cells running (Gox not available)",
			Output:  []interface{}{},
		}
	}

	cells := d.cellManager.ListCells()

	cellsOutput := make([]map[string]interface{}, 0, len(cells))
	for _, cell := range cells {
		cellsOutput = append(cellsOutput, map[string]interface{}{
			"cell_id":    cell.CellID,
			"project_id": cell.ProjectID,
			"vfs_root":   cell.VFSRoot,
			"started_at": cell.StartedAt.Format(time.RFC3339),
		})
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Found %d running cells", len(cells)),
		Output:  cellsOutput,
	}
}

// executeQueryCell sends a query to a cell and waits for response
func (d *Dispatcher) executeQueryCell(ctx context.Context, action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Cellorg is not available",
		}
	}

	projectID, ok := action.Params["project_id"].(string)
	if !ok || projectID == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'project_id' parameter",
		}
	}

	query, ok := action.Params["query"].(string)
	if !ok || query == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'query' parameter",
		}
	}

	// Default timeout to 10 seconds
	timeout := 10 * time.Second
	if timeoutParam, ok := action.Params["timeout"].(float64); ok {
		timeout = time.Duration(timeoutParam) * time.Second
	}

	// Build request/response topics
	requestTopic := projectID + ":queries"
	responseTopic := projectID + ":query-results"

	// Allow custom topics
	if reqTopic, ok := action.Params["request_topic"].(string); ok {
		requestTopic = reqTopic
	}
	if respTopic, ok := action.Params["response_topic"].(string); ok {
		responseTopic = respTopic
	}

	// Prepare query data
	queryData := map[string]interface{}{
		"query":      query,
		"project_id": projectID,
	}

	// Add any additional parameters
	if params, ok := action.Params["params"].(map[string]interface{}); ok {
		for k, v := range params {
			queryData[k] = v
		}
	}

	// Send query and wait for response
	event, err := d.cellManager.PublishAndWait(requestTopic, responseTopic, queryData, timeout)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("query failed: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: "Query completed successfully",
		Output:  event.Data,
	}
}

// executeExtractEntities extracts named entities from text using NER cell
func (d *Dispatcher) executeExtractEntities(ctx context.Context, action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Cellorg is not available. Named entity recognition requires --enable-cellorg flag.",
		}
	}

	text, ok := action.Params["text"].(string)
	if !ok || text == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'text' parameter",
		}
	}

	projectID, ok := action.Params["project_id"].(string)
	if !ok || projectID == "" {
		projectID = "default"
	}

	// Ensure NER cell is started
	cellID := "nlp:entity-extraction"
	cells := d.cellManager.ListCells()
	cellRunning := false
	for _, cell := range cells {
		if cell.CellID == cellID && cell.ProjectID == projectID {
			cellRunning = true
			break
		}
	}

	if !cellRunning {
		// Start the NER cell
		vfsRoot := d.vfs.Root()
		env := make(map[string]string)

		opts := cellorchestrator.CellOptions{
			ProjectID:   projectID,
			VFSRoot:     vfsRoot,
			Environment: env,
		}

		if err := d.cellManager.StartCell(cellID, opts); err != nil {
			return Result{
				Action:  action,
				Success: false,
				Message: fmt.Sprintf("failed to start NER cell: %v", err),
			}
		}

		// Give cell time to initialize
		time.Sleep(2 * time.Second)
	}

	// Send text for NER processing
	nerRequest := map[string]interface{}{
		"text":       text,
		"project_id": projectID,
	}

	// Optional language hint
	if language, ok := action.Params["language"].(string); ok {
		nerRequest["language"] = language
	}

	// Query the NER cell
	timeout := 30 * time.Second
	if timeoutParam, ok := action.Params["timeout"].(float64); ok {
		timeout = time.Duration(timeoutParam) * time.Second
	}

	event, err := d.cellManager.PublishAndWait(
		"text:for-ner",
		"entities:results",
		nerRequest,
		timeout,
	)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("NER extraction failed: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: "Entity extraction completed",
		Output:  event.Data,
	}
}

// executeAnonymizeText anonymizes text by replacing entities with pseudonyms
func (d *Dispatcher) executeAnonymizeText(ctx context.Context, action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Cellorg is not available. Anonymization requires --enable-cellorg flag.",
		}
	}

	text, ok := action.Params["text"].(string)
	if !ok || text == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'text' parameter",
		}
	}

	projectID, ok := action.Params["project_id"].(string)
	if !ok || projectID == "" {
		projectID = "default"
	}

	// Ensure anonymization pipeline cell is started
	cellID := "privacy:anonymization-pipeline"
	cells := d.cellManager.ListCells()
	cellRunning := false
	for _, cell := range cells {
		if cell.CellID == cellID && cell.ProjectID == projectID {
			cellRunning = true
			break
		}
	}

	if !cellRunning {
		// Start the anonymization pipeline cell
		vfsRoot := d.vfs.Root()
		env := make(map[string]string)

		opts := cellorchestrator.CellOptions{
			ProjectID:   projectID,
			VFSRoot:     vfsRoot,
			Environment: env,
		}

		if err := d.cellManager.StartCell(cellID, opts); err != nil {
			return Result{
				Action:  action,
				Success: false,
				Message: fmt.Sprintf("failed to start anonymization cell: %v", err),
			}
		}

		// Give cell time to initialize (storage + NER + anonymizer agents)
		time.Sleep(3 * time.Second)
	}

	// Step 1: Extract entities using NER
	nerRequest := map[string]interface{}{
		"text":       text,
		"project_id": projectID,
	}

	timeout := 30 * time.Second
	if timeoutParam, ok := action.Params["timeout"].(float64); ok {
		timeout = time.Duration(timeoutParam) * time.Second
	}

	nerEvent, err := d.cellManager.PublishAndWait(
		"text:for-analysis",
		"entities:detected",
		nerRequest,
		timeout,
	)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("NER extraction failed: %v", err),
		}
	}

	// Extract entities from NER response
	entities, ok := nerEvent.Data["entities"].([]interface{})
	if !ok || len(entities) == 0 {
		return Result{
			Action:  action,
			Success: false,
			Message: "no entities found in NER response",
		}
	}

	// Step 2: Send to anonymizer
	anonRequest := map[string]interface{}{
		"text":       text,
		"entities":   entities,
		"project_id": projectID,
	}

	anonEvent, err := d.cellManager.PublishAndWait(
		"entities:detected",
		"text:anonymized",
		anonRequest,
		timeout,
	)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("anonymization failed: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: "Text anonymized successfully",
		Output:  anonEvent.Data,
	}
}

// executeDeanonymizeText restores original text from anonymized version
func (d *Dispatcher) executeDeanonymizeText(ctx context.Context, action Action) Result {
	if d.cellManager == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Cellorg is not available. Deanonymization requires --enable-cellorg flag.",
		}
	}

	anonymizedText, ok := action.Params["anonymized_text"].(string)
	if !ok || anonymizedText == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing or invalid 'anonymized_text' parameter",
		}
	}

	projectID, ok := action.Params["project_id"].(string)
	if !ok || projectID == "" {
		projectID = "default"
	}

	// Mappings should be provided to restore text
	mappings, ok := action.Params["mappings"].(map[string]interface{})
	if !ok || len(mappings) == 0 {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'mappings' parameter (obtain from anonymize_text output)",
		}
	}

	// Reverse the mappings: pseudonym -> original
	reversedMappings := make(map[string]string)
	for original, pseudonym := range mappings {
		if pseudonymStr, ok := pseudonym.(string); ok {
			reversedMappings[pseudonymStr] = original
		}
	}

	// Replace pseudonyms with original text
	restoredText := anonymizedText
	for pseudonym, original := range reversedMappings {
		restoredText = strings.ReplaceAll(restoredText, pseudonym, original)
	}

	return Result{
		Action:  action,
		Success: true,
		Message: "Text deanonymized successfully",
		Output: map[string]interface{}{
			"original_text":   anonymizedText,
			"restored_text":   restoredText,
			"replacements":    len(reversedMappings),
		},
	}
}

// executeConfigGet retrieves a configuration setting
func (d *Dispatcher) executeConfigGet(action Action) Result {
	if d.config == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Configuration not available",
		}
	}

	key, ok := action.Params["key"].(string)
	if !ok || key == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'key' parameter",
		}
	}

	value, err := d.config.GetSetting(key)
	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to get setting: %v", err),
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: fmt.Sprintf("Retrieved setting: %s = %s", key, value),
		Output: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	}
}

// executeConfigSet updates a configuration setting
func (d *Dispatcher) executeConfigSet(action Action) Result {
	if d.config == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Configuration not available",
		}
	}

	key, ok := action.Params["key"].(string)
	if !ok || key == "" {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'key' parameter",
		}
	}

	value, ok := action.Params["value"].(string)
	if !ok {
		return Result{
			Action:  action,
			Success: false,
			Message: "missing 'value' parameter",
		}
	}

	// Update the setting
	if err := d.config.UpdateSetting(key, value); err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("failed to update setting: %v", err),
		}
	}

	// Save to disk
	if d.configPath != "" {
		if err := alfaconfig.SaveConfig(d.configPath, d.config); err != nil {
			return Result{
				Action:  action,
				Success: false,
				Message: fmt.Sprintf("setting updated but failed to save: %v", err),
			}
		}
	}

	// Apply immediate changes based on the setting
	message := fmt.Sprintf("Updated setting: %s = %s (saved to alfa.yaml)", key, value)
	needsRestart := false

	// Check if setting requires restart
	restartSettings := []string{
		"ai.provider",
		"ai.config_file",
		"workbench.path",
		"sandbox.enabled",
		"sandbox.image",
		"cellorg.enabled",
		"cellorg.config_path",
		"voice.headless", // voice.input_enabled and voice.output_enabled can be changed at runtime
	}

	for _, rs := range restartSettings {
		if key == rs {
			needsRestart = true
			message += " (restart alfa to apply changes)"
			break
		}
	}

	// Settings that can be applied immediately
	switch key {
	case "execution.auto_confirm":
		message += " (will apply on next operation)"
	case "execution.max_iterations":
		message += " (will apply on next request)"
	case "output.capture_enabled", "output.max_size_kb":
		// Update dispatcher settings
		d.captureOutput = d.config.Output.CaptureEnabled
		d.maxOutputBytes = d.config.Output.MaxSizeKB * 1024
		message += " (applied immediately)"
	case "voice.input_enabled":
		// Enable or disable voice input at runtime
		if d.voiceController != nil {
			if d.config.Voice.InputEnabled {
				if err := d.voiceController.EnableVoiceInput(); err != nil {
					return Result{
						Action:  action,
						Success: false,
						Message: fmt.Sprintf("failed to enable voice input: %v", err),
					}
				}
				message += " (voice input activated)"
			} else {
				d.voiceController.DisableVoiceInput()
				message += " (voice input deactivated)"
			}
		} else {
			message += " (restart alfa to apply changes)"
		}
	case "voice.output_enabled":
		// Enable or disable voice output at runtime
		if d.voiceController != nil {
			if d.config.Voice.OutputEnabled {
				if err := d.voiceController.EnableVoiceOutput(); err != nil {
					return Result{
						Action:  action,
						Success: false,
						Message: fmt.Sprintf("failed to enable voice output: %v", err),
					}
				}
				message += " (voice output activated)"
			} else {
				d.voiceController.DisableVoiceOutput()
				message += " (voice output deactivated)"
			}
		} else {
			message += " (restart alfa to apply changes)"
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: message,
		Output: map[string]interface{}{
			"key":           key,
			"value":         value,
			"needs_restart": needsRestart,
		},
	}
}

// executeConfigList lists all configuration settings
func (d *Dispatcher) executeConfigList(action Action) Result {
	if d.config == nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Configuration not available",
		}
	}

	settings := d.config.ListSettings()

	// Format settings for display
	var output strings.Builder
	output.WriteString("═══════════════════════════════════════════════\n")
	output.WriteString("ALFA CONFIGURATION STATUS\n")
	output.WriteString("═══════════════════════════════════════════════\n\n")

	// Group by category
	categories := map[string][]string{
		"Workbench":      {"workbench.path", "workbench.project"},
		"AI Provider":    {"ai.provider", "ai.config_file"},
		"Voice":          {"voice.input_enabled", "voice.output_enabled", "voice.headless"},
		"Execution":      {"execution.auto_confirm", "execution.max_iterations"},
		"Sandbox":        {"sandbox.enabled", "sandbox.image"},
		"Cellorg":        {"cellorg.enabled", "cellorg.config_path"},
		"Output":         {"output.capture_enabled", "output.max_size_kb"},
		"Self-Modify":    {"self_modify.allowed"},
	}

	for category, keys := range categories {
		output.WriteString(fmt.Sprintf("[%s]\n", category))
		for _, key := range keys {
			if value, ok := settings[key]; ok {
				output.WriteString(fmt.Sprintf("  %-28s = %s\n", key, value))
			}
		}
		output.WriteString("\n")
	}

	// Add detailed AI provider information
	if d.config.AI.Provider != "" {
		output.WriteString("───────────────────────────────────────────────\n")
		output.WriteString("[AI Providers - Detailed]\n\n")

		// Show configured providers
		for providerName, provider := range d.config.AI.Providers {
			active := ""
			if providerName == d.config.AI.Provider {
				active = " ★ ACTIVE"
			}

			output.WriteString(fmt.Sprintf("Provider: %s%s\n", providerName, active))
			output.WriteString(fmt.Sprintf("  Default Model: %s\n", provider.DefaultModel))

			// Show available models
			if len(provider.Models) > 0 {
				output.WriteString("  Available Models:\n")
				for modelName, modelCfg := range provider.Models {
					activeModel := ""
					if providerName == d.config.AI.Provider && modelName == d.config.AI.SelectedModel {
						activeModel = " ← SELECTED"
					} else if providerName == d.config.AI.Provider && modelName == provider.DefaultModel && d.config.AI.SelectedModel == "" {
						activeModel = " ← DEFAULT"
					}

					output.WriteString(fmt.Sprintf("    • %s%s\n", modelName, activeModel))
					output.WriteString(fmt.Sprintf("      Description:  %s\n", modelCfg.Description))
					output.WriteString(fmt.Sprintf("      Max Tokens:   %d\n", modelCfg.MaxTokens))
					output.WriteString(fmt.Sprintf("      Temperature:  %.1f\n", modelCfg.Temperature))
					output.WriteString(fmt.Sprintf("      Timeout:      %s\n", modelCfg.Timeout))
				}
			}
			output.WriteString("\n")
		}
	}

	output.WriteString("───────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("Config file: %s\n", d.configPath))

	return Result{
		Action:  action,
		Success: true,
		Message: "Configuration settings retrieved",
		Output:  output.String(),
	}
}
