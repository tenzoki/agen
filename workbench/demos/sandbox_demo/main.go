package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tenzoki/agen/alfa/internal/sandbox"
)

func main() {
	fmt.Println("=== Docker Sandbox Demo ===\n")

	// Create sandbox with default config
	sb := sandbox.NewDockerSandbox(sandbox.DefaultConfig())

	// Check if Docker is available
	if !sb.IsAvailable() {
		log.Fatal("Docker is not available. Please install Docker Desktop or docker engine.")
	}

	fmt.Println("✓ Docker is available\n")

	ctx := context.Background()

	// Demo 1: Simple command
	fmt.Println("--- Demo 1: Simple Command ---")
	result1, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo 'Hello from Docker sandbox!'",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Exit Code: %d\n", result1.ExitCode)
		fmt.Printf("Output: %s\n", result1.Stdout)
		fmt.Printf("Duration: %v\n\n", result1.Duration)
	}

	// Demo 2: Resource limits
	fmt.Println("--- Demo 2: Resource Limits (512MB RAM, 1 CPU) ---")
	result2, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command:  "cat /proc/meminfo | grep MemTotal && cat /proc/cpuinfo | grep processor | wc -l",
		CPULimit: 1.0,
		MemoryMB: 512,
		Timeout:  5 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Exit Code: %d\n", result2.ExitCode)
		fmt.Printf("Output:\n%s\n", result2.Stdout)
		fmt.Printf("Duration: %v\n\n", result2.Duration)
	}

	// Demo 3: Network disabled
	fmt.Println("--- Demo 3: Network Disabled ---")
	result3, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command:    "ping -c 1 -W 1 8.8.8.8 || echo 'Network disabled as expected'",
		NetworkOff: true,
		Timeout:    5 * time.Second,
	})
	if err != nil {
		fmt.Printf("Expected failure (network disabled): %v\n", err)
	}
	fmt.Printf("Exit Code: %d\n", result3.ExitCode)
	fmt.Printf("Output: %s", result3.Stdout)
	fmt.Printf("Stderr: %s\n\n", result3.Stderr)

	// Demo 4: Environment variables
	fmt.Println("--- Demo 4: Environment Variables ---")
	result4, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo \"MY_VAR is: $MY_VAR\" && echo \"ANOTHER_VAR is: $ANOTHER_VAR\"",
		Env: map[string]string{
			"MY_VAR":      "secret_value",
			"ANOTHER_VAR": "another_secret",
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Output:\n%s\n", result4.Stdout)
	}

	// Demo 5: Timeout
	fmt.Println("--- Demo 5: Timeout (1s limit, 10s sleep) ---")
	result5, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo 'Starting...' && sleep 10 && echo 'This should not print'",
		Timeout: 1 * time.Second,
	})
	if err != nil {
		fmt.Printf("Expected timeout error: %v\n", err)
	}
	fmt.Printf("Exit Code: %d\n", result5.ExitCode)
	fmt.Printf("Output: %s\n", result5.Stdout)

	// Demo 6: Exit codes
	fmt.Println("--- Demo 6: Non-zero Exit Code ---")
	result6, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo 'About to fail...' && exit 42",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("Note: %v\n", err)
	}
	fmt.Printf("Exit Code: %d\n", result6.ExitCode)
	fmt.Printf("Output: %s\n", result6.Stdout)

	// Demo 7: Stdout and Stderr
	fmt.Println("--- Demo 7: Stdout and Stderr Separation ---")
	result7, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo 'This goes to stdout' && echo 'This goes to stderr' >&2",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Stdout: %s", result7.Stdout)
		fmt.Printf("Stderr: %s\n", result7.Stderr)
	}

	// Demo 8: Custom Docker image (Alpine)
	fmt.Println("--- Demo 8: Custom Docker Image (Alpine) ---")
	result8, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "cat /etc/os-release | grep PRETTY_NAME",
		Image:   "alpine:latest",
		Timeout: 10 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Output: %s\n", result8.Stdout)
		fmt.Printf("Duration: %v\n\n", result8.Duration)
	}

	// Demo 9: Security - read-only filesystem
	fmt.Println("--- Demo 9: Security - Read-only Filesystem ---")
	result9, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo 'test' > /test.txt 2>&1 || echo 'Cannot write to root (expected)'",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("Note: %v\n", err)
	}
	fmt.Printf("Output: %s", result9.Stdout)
	fmt.Printf("Stderr: %s\n", result9.Stderr)

	// Demo 10: /tmp is writable
	fmt.Println("--- Demo 10: /tmp is Writable (100MB limit) ---")
	result10, err := sb.Execute(ctx, sandbox.ExecuteRequest{
		Command: "echo 'test' > /tmp/test.txt && cat /tmp/test.txt && rm /tmp/test.txt && echo 'Success!'",
		Timeout: 5 * time.Second,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Output: %s\n", result10.Stdout)
	}

	fmt.Println("--- Demo Complete ---")
	fmt.Println("\nKey Security Features:")
	fmt.Println("✓ Resource limits (CPU, Memory)")
	fmt.Println("✓ Network isolation")
	fmt.Println("✓ Read-only filesystem (except /tmp)")
	fmt.Println("✓ No new privileges")
	fmt.Println("✓ All capabilities dropped")
	fmt.Println("✓ Execution timeout")
	fmt.Println("✓ Auto-cleanup (--rm flag)")
}
