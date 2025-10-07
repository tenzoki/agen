package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// loadManifest loads the extraction manifest
func (e *Extractor) loadManifest() (*Manifest, error) {
	data, err := os.ReadFile(e.manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// saveManifest saves the extraction manifest
func (e *Extractor) saveManifest(agents []AgentInfo, cells []CellInfo) error {
	// Collect all source files
	sources := []SourceFile{}

	// Add agent READMEs
	agentsDir := filepath.Join(e.frameworkRoot, "code", "agents")
	for _, agent := range agents {
		readmePath := filepath.Join(agentsDir, agent.Directory, "README.md")
		info, err := os.Stat(readmePath)
		if err != nil {
			continue
		}
		sources = append(sources, SourceFile{
			Path:    readmePath,
			ModTime: info.ModTime(),
		})
	}

	// Add cell docs
	for _, cell := range cells {
		info, err := os.Stat(cell.FilePath)
		if err != nil {
			continue
		}
		sources = append(sources, SourceFile{
			Path:    cell.FilePath,
			ModTime: info.ModTime(),
		})
	}

	// Add pool.yaml
	poolPath := filepath.Join(e.frameworkRoot, "workbench", "config", "pool.yaml")
	poolInfo, err := os.Stat(poolPath)
	if err == nil {
		sources = append(sources, SourceFile{
			Path:    poolPath,
			ModTime: poolInfo.ModTime(),
		})
	}

	manifest := Manifest{
		Timestamp:   time.Now(),
		Sources:     sources,
		AgentCount:  len(agents),
		CellCount:   len(cells),
		GeneratedBy: "alfa knowledge extractor",
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(e.manifestPath, data, 0644)
}

// MarkStale removes the manifest to force regeneration
func (e *Extractor) MarkStale() error {
	if err := os.Remove(e.manifestPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to mark knowledge base as stale: %w", err)
	}
	return nil
}

// GetManifestInfo returns manifest information if available
func (e *Extractor) GetManifestInfo() (string, error) {
	manifest, err := e.loadManifest()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Last extracted: %s (%d agents, %d cells)",
		manifest.Timestamp.Format("2006-01-02 15:04:05"),
		manifest.AgentCount,
		manifest.CellCount), nil
}
