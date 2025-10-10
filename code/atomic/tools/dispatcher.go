package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/tenzoki/agen/atomic/vfs"
)

// Dispatcher executes tool operations with minimal dependencies
type Dispatcher struct {
	vfs     *vfs.VFS
	timeout time.Duration
}

// NewDispatcher creates a new lightweight tool dispatcher
func NewDispatcher(projectVFS *vfs.VFS) *Dispatcher {
	return &Dispatcher{
		vfs:     projectVFS,
		timeout: 30 * time.Second,
	}
}

// Action represents a tool action to execute
type Action struct {
	Type   string
	Params map[string]interface{}
}

// Result represents the outcome of a tool execution
type Result struct {
	Action  Action
	Success bool
	Message string
	Output  interface{}
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

	cmdCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sh", "-c", command)
	cmd.Dir = d.vfs.Root()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: fmt.Sprintf("command failed: %v", err),
			Output:  outputStr,
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: "Command executed successfully",
		Output:  outputStr,
	}
}

// executeRunTests runs the test suite
func (d *Dispatcher) executeRunTests(ctx context.Context, action Action) Result {
	pattern, ok := action.Params["pattern"].(string)
	if !ok {
		pattern = "./..."
	}

	testTimeout := d.timeout * 2

	cmdCtx, cancel := context.WithTimeout(ctx, testTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "go", "test", "-v", pattern)
	cmd.Dir = d.vfs.Root()

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return Result{
			Action:  action,
			Success: false,
			Message: "Tests failed",
			Output:  outputStr,
		}
	}

	return Result{
		Action:  action,
		Success: true,
		Message: "All tests passed",
		Output:  outputStr,
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
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		matched, _ := filepath.Match(pattern, info.Name())
		if !matched {
			return nil
		}

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
