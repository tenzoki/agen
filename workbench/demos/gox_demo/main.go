package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tenzoki/agen/alfa/internal/gox"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║   Alfa Gox Integration Demo                  ║")
	fmt.Println("╚═══════════════════════════════════════════════╝\n")

	// Create Gox manager
	fmt.Println("🔧 Initializing Gox Manager...")
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "config/gox",
		DefaultDataRoot: "/tmp/alfa-gox-demo",
		Debug:           true,
	})
	if err != nil {
		log.Fatalf("Failed to create Gox manager: %v", err)
	}
	defer func() {
		fmt.Println("\n🔧 Shutting down Gox Manager...")
		mgr.Close()
		fmt.Println("✅ Gox Manager shutdown complete")
	}()

	fmt.Println("✅ Gox Manager initialized\n")

	// Demonstrate cell management
	demonstrateCellManagement(mgr)

	// Demonstrate event system
	demonstrateEventSystem(mgr)

	// Show status
	showStatus(mgr)
}

func demonstrateCellManagement(mgr *gox.Manager) {
	fmt.Println("📦 Cell Management Demo")
	fmt.Println("─────────────────────────────────────────────\n")

	// Start a RAG cell
	fmt.Println("1. Starting RAG cell for project 'demo-project'...")
	err := mgr.StartCell(
		"rag:knowledge-backend",
		"demo-project",
		"/tmp/demo-project",
		map[string]string{
			"OPENAI_API_KEY": "sk-demo",
		},
	)
	if err != nil {
		log.Printf("   ⚠️  Error: %v", err)
	} else {
		fmt.Println("   ✅ RAG cell started successfully")
	}

	time.Sleep(500 * time.Millisecond)

	// Start another cell
	fmt.Println("\n2. Starting document processing cell...")
	err = mgr.StartCell(
		"pipeline:document-processing",
		"demo-project",
		"/tmp/demo-project",
		nil,
	)
	if err != nil {
		log.Printf("   ⚠️  Error: %v", err)
	} else {
		fmt.Println("   ✅ Document processing cell started")
	}

	time.Sleep(500 * time.Millisecond)

	// List running cells
	fmt.Println("\n3. Listing running cells...")
	cells := mgr.ListCells()
	if len(cells) == 0 {
		fmt.Println("   No cells running")
	} else {
		for _, cell := range cells {
			fmt.Printf("   - Cell: %s\n", cell.CellID)
			fmt.Printf("     Project: %s\n", cell.ProjectID)
			fmt.Printf("     VFS Root: %s\n", cell.VFSRoot)
			fmt.Printf("     Started: %s ago\n", time.Since(cell.StartedAt).Round(time.Second))
		}
	}

	time.Sleep(500 * time.Millisecond)

	// Get specific cell info
	fmt.Println("\n4. Getting RAG cell info...")
	info, err := mgr.GetCellInfo("rag:knowledge-backend", "demo-project")
	if err != nil {
		log.Printf("   ⚠️  Error: %v", err)
	} else {
		fmt.Printf("   Cell ID: %s\n", info.CellID)
		fmt.Printf("   Project: %s\n", info.ProjectID)
		fmt.Printf("   VFS Root: %s\n", info.VFSRoot)
	}

	time.Sleep(500 * time.Millisecond)

	// Stop cells
	fmt.Println("\n5. Stopping cells...")
	err = mgr.StopCell("rag:knowledge-backend", "demo-project")
	if err != nil {
		log.Printf("   ⚠️  Error stopping RAG cell: %v", err)
	} else {
		fmt.Println("   ✅ RAG cell stopped")
	}

	err = mgr.StopCell("pipeline:document-processing", "demo-project")
	if err != nil {
		log.Printf("   ⚠️  Error stopping document cell: %v", err)
	} else {
		fmt.Println("   ✅ Document processing cell stopped")
	}

	fmt.Println("")
}

func demonstrateEventSystem(mgr *gox.Manager) {
	fmt.Println("📡 Event System Demo")
	fmt.Println("─────────────────────────────────────────────\n")

	// Subscribe to events
	fmt.Println("1. Subscribing to events...")

	eventReceived := false
	mgr.Subscribe("demo:events", func(event gox.Event) {
		fmt.Printf("   📩 Event received on topic '%s'\n", event.Topic)
		fmt.Printf("      Project: %s\n", event.ProjectID)
		fmt.Printf("      Data: %v\n", event.Data)
		eventReceived = true
	})
	fmt.Println("   ✅ Subscribed to 'demo:events'")

	time.Sleep(500 * time.Millisecond)

	// Publish an event
	fmt.Println("\n2. Publishing event...")
	err := mgr.Publish("demo:events", map[string]interface{}{
		"message": "Hello from Alfa!",
		"timestamp": time.Now().Unix(),
	})
	if err != nil {
		log.Printf("   ⚠️  Error: %v", err)
	} else {
		fmt.Println("   ✅ Event published")
	}

	time.Sleep(500 * time.Millisecond)

	// Try PublishAndWait (will fail in placeholder)
	fmt.Println("\n3. Testing PublishAndWait (synchronous query)...")
	_, err = mgr.PublishAndWait(
		"demo:query-requests",
		"demo:query-results",
		map[string]interface{}{
			"query": "find authentication code",
		},
		2*time.Second,
	)
	if err != nil {
		fmt.Printf("   ⚠️  Expected error (placeholder): %v\n", err)
	}

	if !eventReceived {
		fmt.Println("\n   ℹ️  Note: Events don't flow in placeholder implementation")
		fmt.Println("      Will work when gox pkg/orchestrator is integrated")
	}

	fmt.Println("")
}

func showStatus(mgr *gox.Manager) {
	fmt.Println("🔍 Manager Status")
	fmt.Println("─────────────────────────────────────────────\n")

	// Health check
	healthy := mgr.IsHealthy(nil)
	if healthy {
		fmt.Println("   ✅ Gox Manager is healthy")
	} else {
		fmt.Println("   ⚠️  Gox Manager is not healthy")
	}

	// Show running cells
	cells := mgr.ListCells()
	fmt.Printf("   📦 Running cells: %d\n", len(cells))

	fmt.Println("")
}
