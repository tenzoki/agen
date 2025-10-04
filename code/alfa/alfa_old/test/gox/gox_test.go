package gox_test

import (
	"context"
	"testing"
	"time"

	"alfa/internal/gox"
)

func TestNewManager(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	if mgr == nil {
		t.Fatal("Expected manager to be non-nil")
	}
}

func TestStartStopCell(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           true,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Start a cell
	err = mgr.StartCell("test-cell", "test-project", "/tmp/test-project", nil)
	if err != nil {
		t.Fatalf("Failed to start cell: %v", err)
	}

	// Verify cell is tracked
	cells := mgr.ListCells()
	if len(cells) != 1 {
		t.Fatalf("Expected 1 cell, got %d", len(cells))
	}

	if cells[0].CellID != "test-cell" {
		t.Errorf("Expected cell ID 'test-cell', got '%s'", cells[0].CellID)
	}

	if cells[0].ProjectID != "test-project" {
		t.Errorf("Expected project ID 'test-project', got '%s'", cells[0].ProjectID)
	}

	// Stop the cell
	err = mgr.StopCell("test-cell", "test-project")
	if err != nil {
		t.Fatalf("Failed to stop cell: %v", err)
	}

	// Verify cell is removed
	cells = mgr.ListCells()
	if len(cells) != 0 {
		t.Fatalf("Expected 0 cells after stop, got %d", len(cells))
	}
}

func TestListCells(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Initially no cells
	cells := mgr.ListCells()
	if len(cells) != 0 {
		t.Fatalf("Expected 0 cells initially, got %d", len(cells))
	}

	// Start multiple cells
	err = mgr.StartCell("cell-1", "project-a", "/tmp/project-a", nil)
	if err != nil {
		t.Fatalf("Failed to start cell-1: %v", err)
	}

	err = mgr.StartCell("cell-2", "project-b", "/tmp/project-b", nil)
	if err != nil {
		t.Fatalf("Failed to start cell-2: %v", err)
	}

	// Should have 2 cells
	cells = mgr.ListCells()
	if len(cells) != 2 {
		t.Fatalf("Expected 2 cells, got %d", len(cells))
	}

	// Clean up
	mgr.StopCell("cell-1", "project-a")
	mgr.StopCell("cell-2", "project-b")
}

func TestGetCellInfo(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Start a cell
	vfsRoot := "/tmp/test-vfs"
	err = mgr.StartCell("info-cell", "info-project", vfsRoot, map[string]string{
		"TEST_VAR": "test_value",
	})
	if err != nil {
		t.Fatalf("Failed to start cell: %v", err)
	}
	defer mgr.StopCell("info-cell", "info-project")

	// Get cell info
	info, err := mgr.GetCellInfo("info-cell", "info-project")
	if err != nil {
		t.Fatalf("Failed to get cell info: %v", err)
	}

	if info.CellID != "info-cell" {
		t.Errorf("Expected cell ID 'info-cell', got '%s'", info.CellID)
	}

	if info.ProjectID != "info-project" {
		t.Errorf("Expected project ID 'info-project', got '%s'", info.ProjectID)
	}

	if info.VFSRoot != vfsRoot {
		t.Errorf("Expected VFS root '%s', got '%s'", vfsRoot, info.VFSRoot)
	}
}

func TestDuplicateCellStart(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Start a cell
	err = mgr.StartCell("dup-cell", "dup-project", "/tmp/dup", nil)
	if err != nil {
		t.Fatalf("Failed to start cell: %v", err)
	}
	defer mgr.StopCell("dup-cell", "dup-project")

	// Try to start the same cell again - should error
	err = mgr.StartCell("dup-cell", "dup-project", "/tmp/dup", nil)
	if err == nil {
		t.Fatal("Expected error when starting duplicate cell, got nil")
	}
}

func TestStopNonExistentCell(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Try to stop a cell that doesn't exist
	err = mgr.StopCell("nonexistent", "project")
	if err == nil {
		t.Fatal("Expected error when stopping non-existent cell, got nil")
	}
}

func TestHealthCheck(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	ctx := context.Background()
	if !mgr.IsHealthy(ctx) {
		t.Error("Expected manager to be healthy")
	}
}

func TestPublish(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Publish should not error (placeholder implementation)
	err = mgr.Publish("test:topic", map[string]interface{}{
		"message": "test",
	})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}
}

func TestPublishAndWait(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           false,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// PublishAndWait should return error (placeholder not functional yet)
	_, err = mgr.PublishAndWait("req:topic", "resp:topic", map[string]interface{}{
		"query": "test",
	}, 1*time.Second)

	if err == nil {
		t.Error("Expected error from PublishAndWait placeholder, got nil")
	}
}

func TestSubscribe(t *testing.T) {
	mgr, err := gox.NewManager(gox.Config{
		ConfigPath:      "../../config/gox",
		DefaultDataRoot: "/tmp/alfa-test",
		Debug:           true,
	})
	if err != nil {
		t.Fatalf("Failed to create gox manager: %v", err)
	}
	defer mgr.Close()

	// Subscribe should not panic
	called := false
	mgr.Subscribe("test:events", func(event gox.Event) {
		called = true
	})

	// Note: Events won't actually flow in placeholder implementation
	// This just tests that subscription API works
	time.Sleep(100 * time.Millisecond)

	// In placeholder, called will be false
	if called {
		t.Log("Event handler was called (unexpected in placeholder)")
	}
}
