package sandbox_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/tenzoki/agen/alfa/internal/sandbox"
)

func TestDockerSandbox_IsAvailable(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available, skipping test")
	}
}

func TestDockerSandbox_SimpleCommand(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command: "echo 'Hello from sandbox'",
		Timeout: 5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	if !strings.Contains(result.Stdout, "Hello from sandbox") {
		t.Errorf("Expected stdout to contain 'Hello from sandbox', got: %s", result.Stdout)
	}
}

func TestDockerSandbox_ExitCode(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command: "exit 42",
		Timeout: 5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)

	// Should get result even with non-zero exit
	if result == nil {
		t.Fatal("Expected result even with non-zero exit code")
	}

	if result.ExitCode != 42 {
		t.Errorf("Expected exit code 42, got %d", result.ExitCode)
	}

	// err should be nil since execution itself succeeded
	if err != nil {
		t.Logf("Got error (acceptable): %v", err)
	}
}

func TestDockerSandbox_Timeout(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command: "sleep 10",
		Timeout: 1 * time.Second,
	}

	result, err := sb.Execute(ctx, req)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if result == nil {
		t.Fatal("Expected result even with timeout")
	}

	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for timeout, got %d", result.ExitCode)
	}
}

func TestDockerSandbox_ResourceLimits(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command:  "cat /proc/cpuinfo | grep processor | wc -l",
		CPULimit: 0.5,
		MemoryMB: 256,
		Timeout:  5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", result.ExitCode, result.Stderr)
	}

	// Just verify it ran successfully
	t.Logf("CPU count output: %s", result.Stdout)
}

func TestDockerSandbox_NetworkDisabled(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command:    "ping -c 1 8.8.8.8",
		NetworkOff: true,
		Timeout:    5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)

	// Should fail because network is disabled
	if err == nil && result.ExitCode == 0 {
		t.Error("Expected ping to fail with network disabled")
	}
}

func TestDockerSandbox_EnvironmentVariables(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command: "echo $MY_VAR",
		Env: map[string]string{
			"MY_VAR": "test_value",
		},
		Timeout: 5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result.Stdout, "test_value") {
		t.Errorf("Expected stdout to contain 'test_value', got: %s", result.Stdout)
	}
}

func TestDockerSandbox_WorkDir(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	// Create temp dir for test
	tmpDir := t.TempDir()

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command: "ls -la",
		WorkDir: tmpDir,
		Timeout: 5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d. Stderr: %s", result.ExitCode, result.Stderr)
	}

	t.Logf("WorkDir contents: %s", result.Stdout)
}

func TestDockerSandbox_StderrCapture(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()
	req := sandbox.ExecuteRequest{
		Command: "echo 'stdout message' && echo 'stderr message' >&2",
		Timeout: 5 * time.Second,
	}

	result, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result.Stdout, "stdout message") {
		t.Errorf("Expected stdout to contain 'stdout message', got: %s", result.Stdout)
	}

	if !strings.Contains(result.Stderr, "stderr message") {
		t.Errorf("Expected stderr to contain 'stderr message', got: %s", result.Stderr)
	}
}

func TestDockerSandbox_CustomImage(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	ctx := context.Background()

	// Use alpine image (smaller, faster)
	req := sandbox.ExecuteRequest{
		Command: "cat /etc/os-release",
		Image:   "alpine:latest",
		Timeout: 10 * time.Second,
	}

	result, err := sb.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result.Stdout, "Alpine") {
		t.Errorf("Expected Alpine in output, got: %s", result.Stdout)
	}
}

func TestDockerSandbox_Cleanup(t *testing.T) {
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	if !sb.IsAvailable() {
		t.Skip("Docker not available")
	}

	err := sb.Cleanup()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
}
