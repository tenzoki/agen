package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Extractor coordinates knowledge extraction
type Extractor struct {
	frameworkRoot string
	manualDir     string
	manifestPath  string
}

// NewExtractor creates a new knowledge extractor
func NewExtractor(frameworkRoot string) *Extractor {
	manualDir := filepath.Join(frameworkRoot, "reflect", "manual")
	return &Extractor{
		frameworkRoot: frameworkRoot,
		manualDir:     manualDir,
		manifestPath:  filepath.Join(manualDir, ".manifest"),
	}
}

// IsStale checks if the knowledge base needs regeneration
func (e *Extractor) IsStale() bool {
	// Check if manifest exists
	manifest, err := e.loadManifest()
	if err != nil {
		return true // No manifest = needs extraction
	}

	// Check if any source is newer than manifest
	for _, source := range manifest.Sources {
		info, err := os.Stat(source.Path)
		if err != nil {
			continue // Source deleted? Need regeneration
		}
		if info.ModTime().After(manifest.Timestamp) {
			return true
		}
	}

	// Check if output files exist
	outputs := []string{"agents.md", "cells.md", "capabilities.md"}
	for _, output := range outputs {
		path := filepath.Join(e.manualDir, output)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return true
		}
	}

	return false
}

// Extract performs full knowledge extraction
func (e *Extractor) Extract() error {
	fmt.Printf("🔍 Extracting knowledge base...\n")
	startTime := time.Now()

	// Ensure manual directory exists
	if err := os.MkdirAll(e.manualDir, 0755); err != nil {
		return fmt.Errorf("failed to create manual directory: %w", err)
	}

	// Scan sources
	fmt.Printf("  Scanning agents...\n")
	agents, err := e.scanAgents()
	if err != nil {
		return fmt.Errorf("failed to scan agents: %w", err)
	}

	fmt.Printf("  Scanning cells...\n")
	cells, err := e.scanCells()
	if err != nil {
		return fmt.Errorf("failed to scan cells: %w", err)
	}

	fmt.Printf("  Parsing pool.yaml...\n")
	pool, err := e.parsePool()
	if err != nil {
		return fmt.Errorf("failed to parse pool: %w", err)
	}

	// Enrich agents with pool info
	for i := range agents {
		if poolAgent, exists := pool.AgentTypes[agents[i].Type]; exists {
			agents[i].Description = poolAgent.Description
			agents[i].Operator = poolAgent.Operator
			agents[i].Binary = poolAgent.Binary
			if len(agents[i].Capabilities) == 0 {
				agents[i].Capabilities = poolAgent.Capabilities
			}
		}
	}

	// Generate documentation
	fmt.Printf("  Generating agents.md...\n")
	if err := e.generateAgentsDoc(agents); err != nil {
		return fmt.Errorf("failed to generate agents.md: %w", err)
	}

	fmt.Printf("  Generating cells.md...\n")
	if err := e.generateCellsDoc(cells); err != nil {
		return fmt.Errorf("failed to generate cells.md: %w", err)
	}

	fmt.Printf("  Generating capabilities.md...\n")
	if err := e.generateCapabilitiesDoc(agents, cells); err != nil {
		return fmt.Errorf("failed to generate capabilities.md: %w", err)
	}

	// Save manifest
	fmt.Printf("  Saving manifest...\n")
	if err := e.saveManifest(agents, cells); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ Knowledge base extracted: %d agents, %d cells (%v)\n", len(agents), len(cells), elapsed)

	return nil
}

// LoadAgentsDoc loads the agents documentation
func (e *Extractor) LoadAgentsDoc() (string, error) {
	return e.loadDoc("agents.md")
}

// LoadCellsDoc loads the cells documentation
func (e *Extractor) LoadCellsDoc() (string, error) {
	return e.loadDoc("cells.md")
}

// LoadCapabilitiesDoc loads the capabilities documentation
func (e *Extractor) LoadCapabilitiesDoc() (string, error) {
	return e.loadDoc("capabilities.md")
}

// loadDoc loads a documentation file
func (e *Extractor) loadDoc(filename string) (string, error) {
	path := filepath.Join(e.manualDir, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", filename, err)
	}
	return string(content), nil
}
