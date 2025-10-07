package knowledge

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// scanAgents scans all agent README files
func (e *Extractor) scanAgents() ([]AgentInfo, error) {
	agentsDir := filepath.Join(e.frameworkRoot, "code", "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, err
	}

	var agents []AgentInfo
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "testutil" {
			continue
		}

		agentDir := filepath.Join(agentsDir, entry.Name())
		readmePath := filepath.Join(agentDir, "README.md")

		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			continue // No README
		}

		agent, err := e.parseAgentREADME(readmePath, entry.Name())
		if err != nil {
			fmt.Printf("  Warning: failed to parse %s: %v\n", readmePath, err)
			continue
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// parseAgentREADME parses an agent README file
func (e *Extractor) parseAgentREADME(path, dirName string) (AgentInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return AgentInfo{}, err
	}
	defer file.Close()

	agent := AgentInfo{
		Directory: dirName,
		Type:      strings.ReplaceAll(dirName, "_", "-"), // snake_case -> kebab-case
		Config:    make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var sectionContent []string
	var inCodeBlock bool

	for scanner.Scan() {
		line := scanner.Text()

		// Track code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines inside code blocks
		if inCodeBlock {
			continue
		}

		// Detect sections
		if strings.HasPrefix(line, "# ") {
			agent.Name = strings.TrimPrefix(line, "# ")
		} else if strings.HasPrefix(line, "## ") {
			// Save previous section
			if currentSection != "" {
				e.saveAgentSection(&agent, currentSection, strings.Join(sectionContent, "\n"))
			}
			currentSection = strings.TrimPrefix(line, "## ")
			sectionContent = []string{}
		} else if currentSection != "" && line != "" {
			sectionContent = append(sectionContent, line)
		}
	}

	// Save last section
	if currentSection != "" {
		e.saveAgentSection(&agent, currentSection, strings.Join(sectionContent, "\n"))
	}

	return agent, scanner.Err()
}

// saveAgentSection saves a parsed section to agent info
func (e *Extractor) saveAgentSection(agent *AgentInfo, section, content string) {
	content = strings.TrimSpace(content)
	switch section {
	case "Intent":
		agent.Intent = content
	case "Usage":
		agent.Usage = content
	}
}

// scanCells scans all cell documentation files
func (e *Extractor) scanCells() ([]CellInfo, error) {
	cellsDir := filepath.Join(e.frameworkRoot, "reflect", "cells")

	categories := []string{"pipelines", "services", "analysis", "synthesis"}
	var cells []CellInfo

	for _, category := range categories {
		categoryDir := filepath.Join(cellsDir, category)
		entries, err := os.ReadDir(categoryDir)
		if err != nil {
			continue // Category doesn't exist
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			cellPath := filepath.Join(categoryDir, entry.Name())
			cell, err := e.parseCellDoc(cellPath, category)
			if err != nil {
				fmt.Printf("  Warning: failed to parse %s: %v\n", cellPath, err)
				continue
			}

			cells = append(cells, cell)
		}
	}

	return cells, nil
}

// parseCellDoc parses a cell documentation file
func (e *Extractor) parseCellDoc(path, category string) (CellInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return CellInfo{}, err
	}
	defer file.Close()

	cell := CellInfo{
		Category: category,
		FilePath: path,
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	var sectionContent []string
	var inCodeBlock bool

	for scanner.Scan() {
		line := scanner.Text()

		// Track code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines inside code blocks
		if inCodeBlock {
			continue
		}

		if strings.HasPrefix(line, "# ") {
			cell.ID = strings.TrimPrefix(line, "# ")
		} else if strings.HasPrefix(line, "## ") {
			if currentSection != "" {
				e.saveCellSection(&cell, currentSection, strings.Join(sectionContent, "\n"))
			}
			currentSection = strings.TrimPrefix(line, "## ")
			sectionContent = []string{}
		} else if currentSection != "" && line != "" {
			sectionContent = append(sectionContent, line)
		}
	}

	if currentSection != "" {
		e.saveCellSection(&cell, currentSection, strings.Join(sectionContent, "\n"))
	}

	return cell, scanner.Err()
}

// saveCellSection saves a parsed section to cell info
func (e *Extractor) saveCellSection(cell *CellInfo, section, content string) {
	content = strings.TrimSpace(content)
	switch section {
	case "Intent", "Purpose":
		cell.Purpose = content
	case "Description":
		cell.Description = content
	}
}

// parsePool parses the pool.yaml file
func (e *Extractor) parsePool() (*PoolInfo, error) {
	poolPath := filepath.Join(e.frameworkRoot, "workbench", "config", "pool.yaml")

	data, err := os.ReadFile(poolPath)
	if err != nil {
		return nil, err
	}

	var config struct {
		Pool struct {
			AgentTypes []struct {
				AgentType    string   `yaml:"agent_type"`
				Binary       string   `yaml:"binary"`
				Operator     string   `yaml:"operator"`
				Capabilities []string `yaml:"capabilities"`
				Description  string   `yaml:"description"`
			} `yaml:"agent_types"`
		} `yaml:"pool"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	pool := &PoolInfo{
		AgentTypes: make(map[string]PoolAgent),
	}

	for _, at := range config.Pool.AgentTypes {
		pool.AgentTypes[at.AgentType] = PoolAgent{
			Type:         at.AgentType,
			Binary:       at.Binary,
			Operator:     at.Operator,
			Capabilities: at.Capabilities,
			Description:  at.Description,
		}
	}

	return pool, nil
}
