package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// DockerSandbox implements Sandbox using Docker
type DockerSandbox struct {
	config Config
}

// NewDockerSandbox creates a new Docker-based sandbox
func NewDockerSandbox(config Config) *DockerSandbox {
	return &DockerSandbox{
		config: config,
	}
}

// IsAvailable checks if Docker is installed and running
func (s *DockerSandbox) IsAvailable() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

// Execute runs a command in a Docker container with resource limits
func (s *DockerSandbox) Execute(ctx context.Context, req ExecuteRequest) (*ExecuteResult, error) {
	if !s.IsAvailable() {
		return nil, &Error{
			Operation: "execute",
			Message:   "Docker not available. Install Docker Desktop or docker engine.",
		}
	}

	// Apply defaults
	if req.Image == "" {
		req.Image = s.config.DefaultImage
	}
	if req.Timeout == 0 {
		req.Timeout = s.config.DefaultTimeout
	}
	if req.CPULimit == 0 {
		req.CPULimit = s.config.DefaultCPULimit
	}
	if req.MemoryMB == 0 {
		req.MemoryMB = s.config.DefaultMemoryMB
	}

	// Build docker run command
	args := []string{"run", "--rm"}

	// Resource limits
	args = append(args, "--cpus", fmt.Sprintf("%.2f", req.CPULimit))
	args = append(args, "--memory", fmt.Sprintf("%dm", req.MemoryMB))

	// Network
	if req.NetworkOff || s.config.NetworkOff {
		args = append(args, "--network", "none")
	}

	// Mount working directory
	if req.WorkDir != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/workspace", req.WorkDir))
		args = append(args, "-w", "/workspace")
	}

	// Environment variables
	for key, value := range req.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Security options
	args = append(args, "--security-opt", "no-new-privileges")
	args = append(args, "--cap-drop", "ALL")
	args = append(args, "--read-only")
	args = append(args, "--tmpfs", "/tmp:rw,noexec,nosuid,size=100m")

	// Image and command
	args = append(args, req.Image)
	args = append(args, "sh", "-c", req.Command)

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// Execute
	startTime := time.Now()
	cmd := exec.CommandContext(execCtx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	result := &ExecuteResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	// Handle errors
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = &Error{
				Operation: "execute",
				Message:   fmt.Sprintf("execution timeout after %v", req.Timeout),
				Err:       execCtx.Err(),
			}
			result.ExitCode = -1
			return result, result.Error
		}

		// Check if it's an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = &Error{
				Operation: "execute",
				Message:   fmt.Sprintf("docker execution failed: %v", err),
				Err:       err,
			}
			return result, result.Error
		}
	}

	return result, nil
}

// Cleanup removes stopped containers (if not using --rm)
func (s *DockerSandbox) Cleanup() error {
	// With --rm flag, containers auto-cleanup
	// This method is here for manual cleanup if needed
	cmd := exec.Command("docker", "container", "prune", "-f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &Error{
			Operation: "cleanup",
			Message:   fmt.Sprintf("cleanup failed: %s", string(output)),
			Err:       err,
		}
	}
	return nil
}

// ExecuteStreaming runs a command and streams output in real-time
func (s *DockerSandbox) ExecuteStreaming(ctx context.Context, req ExecuteRequest, stdoutWriter, stderrWriter io.Writer) (*ExecuteResult, error) {
	if !s.IsAvailable() {
		return nil, &Error{
			Operation: "execute",
			Message:   "Docker not available",
		}
	}

	// Apply defaults
	if req.Image == "" {
		req.Image = s.config.DefaultImage
	}
	if req.Timeout == 0 {
		req.Timeout = s.config.DefaultTimeout
	}
	if req.CPULimit == 0 {
		req.CPULimit = s.config.DefaultCPULimit
	}
	if req.MemoryMB == 0 {
		req.MemoryMB = s.config.DefaultMemoryMB
	}

	// Build docker run command
	args := []string{"run", "--rm"}
	args = append(args, "--cpus", fmt.Sprintf("%.2f", req.CPULimit))
	args = append(args, "--memory", fmt.Sprintf("%dm", req.MemoryMB))

	if req.NetworkOff || s.config.NetworkOff {
		args = append(args, "--network", "none")
	}

	if req.WorkDir != "" {
		args = append(args, "-v", fmt.Sprintf("%s:/workspace", req.WorkDir))
		args = append(args, "-w", "/workspace")
	}

	for key, value := range req.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	args = append(args, "--security-opt", "no-new-privileges")
	args = append(args, "--cap-drop", "ALL")
	args = append(args, "--read-only")
	args = append(args, "--tmpfs", "/tmp:rw,noexec,nosuid,size=100m")

	args = append(args, req.Image)
	args = append(args, "sh", "-c", req.Command)

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	// Execute with streaming output
	startTime := time.Now()
	cmd := exec.CommandContext(execCtx, "docker", args...)

	// Capture output while streaming
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdoutBuf, stdoutWriter)
	cmd.Stderr = io.MultiWriter(&stderrBuf, stderrWriter)

	err := cmd.Run()
	duration := time.Since(startTime)

	result := &ExecuteResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		Duration: duration,
	}

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			result.Error = &Error{
				Operation: "execute",
				Message:   fmt.Sprintf("timeout after %v", req.Timeout),
				Err:       execCtx.Err(),
			}
			result.ExitCode = -1
			return result, result.Error
		}

		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.Error = &Error{
				Operation: "execute",
				Message:   fmt.Sprintf("execution failed: %v", err),
				Err:       err,
			}
			return result, result.Error
		}
	}

	return result, nil
}

// PullImage pulls a Docker image if not present
func (s *DockerSandbox) PullImage(ctx context.Context, image string) error {
	if !s.IsAvailable() {
		return &Error{
			Operation: "pull",
			Message:   "Docker not available",
		}
	}

	// Check if image exists
	checkCmd := exec.Command("docker", "image", "inspect", image)
	if checkCmd.Run() == nil {
		return nil // Image already exists
	}

	// Pull image
	cmd := exec.CommandContext(ctx, "docker", "pull", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &Error{
			Operation: "pull",
			Message:   fmt.Sprintf("failed to pull image %s: %s", image, string(output)),
			Err:       err,
		}
	}

	return nil
}

// GetImageSize returns the size of a Docker image in MB
func (s *DockerSandbox) GetImageSize(image string) (int64, error) {
	cmd := exec.Command("docker", "image", "inspect", image, "--format", "{{.Size}}")
	output, err := cmd.Output()
	if err != nil {
		return 0, &Error{
			Operation: "inspect",
			Message:   fmt.Sprintf("failed to inspect image %s", image),
			Err:       err,
		}
	}

	sizeBytes, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, err
	}

	return sizeBytes / (1024 * 1024), nil // Convert to MB
}
