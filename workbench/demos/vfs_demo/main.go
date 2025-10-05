package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tenzoki/agen/alfa/internal/vfs"
)

func main() {
	fmt.Println("=== VFS Demo: Workbench vs Project Separation ===\n")

	// Create workbench VFS (read/write - for config, context, history)
	workbench, err := vfs.NewVFS("./demo_workbench", false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Workbench VFS created at: %s\n", workbench.Root())

	// Create project VFS (read/write - for source code, sandboxed AI operations)
	project, err := vfs.NewVFS("./demo_workbench/project", false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Project VFS created at: %s\n\n", project.Root())

	// === Workbench Operations ===
	fmt.Println("--- Workbench Operations (Config, Context, History) ---")

	// Store AI config
	configContent := `{
  "default_provider": "openai",
  "providers": {
    "openai": {
      "model": "gpt-4",
      "max_tokens": 4096
    }
  }
}`
	if err := workbench.WriteString(configContent, "config", "ai-config.json"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Wrote AI config to workbench")

	// Store conversation context
	contextData := `{
  "conversation_id": "20250930-123456",
  "messages": [
    {"role": "user", "content": "Fix the bug in main.go"},
    {"role": "assistant", "content": "I'll analyze main.go"}
  ]
}`
	if err := workbench.WriteString(contextData, ".alfa", "context.json"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Stored context in workbench")

	// Store operation history
	historyEntry := "2025-09-30 12:34:56 | Applied patch to main.go\n"
	if err := workbench.Append([]byte(historyEntry), ".alfa", "history.log"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Appended to history log\n")

	// === Project Operations (Sandboxed) ===
	fmt.Println("--- Project Operations (Sandboxed Code) ---")

	// Create project structure
	if err := project.Mkdir("cmd", "myapp"); err != nil {
		log.Fatal(err)
	}
	if err := project.Mkdir("internal", "logic"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Created project directory structure")

	// Write source files
	mainGo := `package main

import "fmt"

func main() {
    fmt.Println("Hello from Alfa!")
}
`
	if err := project.WriteString(mainGo, "cmd", "myapp", "main.go"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Created cmd/myapp/main.go")

	logicGo := `package logic

func Calculate(x, y int) int {
    return x + y
}
`
	if err := project.WriteString(logicGo, "internal", "logic", "logic.go"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Created internal/logic/logic.go")

	goMod := `module myapp

go 1.24.3
`
	if err := project.WriteString(goMod, "go.mod"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Created go.mod\n")

	// === Security: Path Traversal Prevention ===
	fmt.Println("--- Security: Path Traversal Prevention ---")

	// Attempt to escape project VFS (should fail)
	err = project.WriteString("malicious", "..", "..", "etc", "passwd")
	if err != nil {
		fmt.Printf("✓ Path traversal blocked: %v\n", err)
	} else {
		fmt.Println("✗ SECURITY FAILURE: Path traversal succeeded!")
	}

	// Attempt to access workbench config from project VFS (should fail)
	_, err = project.Path("..", "config", "ai-config.json")
	if err != nil {
		fmt.Printf("✓ Access to workbench blocked: %v\n\n", err)
	} else {
		fmt.Println("✗ SECURITY FAILURE: Accessed workbench from project!")
	}

	// === Read-Only VFS Demo ===
	fmt.Println("--- Read-Only VFS Demo ---")

	// Create a read-only view of the project
	readOnlyProject, err := vfs.NewVFS(project.Root(), true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created read-only project VFS\n")

	// Can read files
	content, err := readOnlyProject.ReadString("cmd", "myapp", "main.go")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Read file: %d bytes\n", len(content))

	// Cannot write files
	err = readOnlyProject.WriteString("modified", "cmd", "myapp", "main.go")
	if err != nil {
		fmt.Printf("✓ Write blocked on read-only VFS: %v\n\n", err)
	} else {
		fmt.Println("✗ SECURITY FAILURE: Write succeeded on read-only VFS!")
	}

	// === List Project Contents ===
	fmt.Println("--- Project Contents ---")

	err = project.Walk(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Printf("  %s\n", path)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n--- Demo Complete ---")
	fmt.Println("Workbench and project are properly isolated.")
	fmt.Println("All AI operations will be sandboxed to the project VFS.")
}
