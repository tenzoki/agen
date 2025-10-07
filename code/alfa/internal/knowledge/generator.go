package knowledge

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// generateAgentsDoc generates the agents.md documentation
func (e *Extractor) generateAgentsDoc(agents []AgentInfo) error {
	var sb strings.Builder

	sb.WriteString("# AGEN Agent Capabilities\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Agents:** %d  \n\n", len(agents)))
	sb.WriteString("Complete catalog of available agents and their capabilities.\n\n")
	sb.WriteString("---\n\n")

	// Sort agents by name
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Name < agents[j].Name
	})

	for _, agent := range agents {
		sb.WriteString(fmt.Sprintf("## %s\n\n", agent.Name))
		sb.WriteString(fmt.Sprintf("**Type:** `%s`  \n", agent.Type))
		sb.WriteString(fmt.Sprintf("**Operator:** %s  \n", agent.Operator))

		if agent.Description != "" {
			sb.WriteString(fmt.Sprintf("**Description:** %s  \n\n", agent.Description))
		}

		if agent.Intent != "" {
			sb.WriteString("**Intent:**  \n")
			sb.WriteString(agent.Intent)
			sb.WriteString("\n\n")
		}

		if len(agent.Capabilities) > 0 {
			sb.WriteString("**Capabilities:**  \n")
			for _, cap := range agent.Capabilities {
				sb.WriteString(fmt.Sprintf("- %s\n", cap))
			}
			sb.WriteString("\n")
		}

		if agent.Usage != "" {
			sb.WriteString("**Usage:**  \n")
			// Limit usage to first 5 lines to keep compact
			lines := strings.Split(agent.Usage, "\n")
			for i, line := range lines {
				if i >= 5 {
					sb.WriteString("...\n")
					break
				}
				sb.WriteString(line)
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	path := filepath.Join(e.manualDir, "agents.md")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// generateCellsDoc generates the cells.md documentation
func (e *Extractor) generateCellsDoc(cells []CellInfo) error {
	var sb strings.Builder

	sb.WriteString("# AGEN Cell Patterns\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Cells:** %d  \n\n", len(cells)))
	sb.WriteString("Available cell configurations organized by purpose.\n\n")
	sb.WriteString("---\n\n")

	// Group by category
	categories := map[string][]CellInfo{
		"pipelines": {},
		"services":  {},
		"analysis":  {},
		"synthesis": {},
	}

	for _, cell := range cells {
		categories[cell.Category] = append(categories[cell.Category], cell)
	}

	// Generate by category
	categoryOrder := []string{"pipelines", "services", "analysis", "synthesis"}
	categoryTitles := map[string]string{
		"pipelines": "Processing Pipelines",
		"services":  "Backend Services",
		"analysis":  "Content Analysis",
		"synthesis": "Output Synthesis",
	}

	for _, category := range categoryOrder {
		cells := categories[category]
		if len(cells) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", categoryTitles[category]))

		sort.Slice(cells, func(i, j int) bool {
			return cells[i].ID < cells[j].ID
		})

		for _, cell := range cells {
			sb.WriteString(fmt.Sprintf("### %s\n\n", cell.ID))

			if cell.Purpose != "" {
				sb.WriteString(cell.Purpose)
				sb.WriteString("\n\n")
			} else if cell.Description != "" {
				sb.WriteString(cell.Description)
				sb.WriteString("\n\n")
			}

			// Extract config path
			configPath := strings.Replace(cell.FilePath, e.frameworkRoot+"/reflect/cells/", "workbench/config/cells/", 1)
			configPath = strings.Replace(configPath, ".md", ".yaml", 1)

			sb.WriteString(fmt.Sprintf("**Config:** `%s`\n\n", configPath))
		}

		sb.WriteString("---\n\n")
	}

	path := filepath.Join(e.manualDir, "cells.md")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// generateCapabilitiesDoc generates the capabilities.md high-level guide
func (e *Extractor) generateCapabilitiesDoc(agents []AgentInfo, cells []CellInfo) error {
	var sb strings.Builder

	sb.WriteString("# AGEN Capabilities Overview\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("High-level guide to what AGEN can do.\n\n")
	sb.WriteString("---\n\n")

	sb.WriteString("## What is AGEN?\n\n")
	sb.WriteString("AGEN is a cell-based orchestration framework for building distributed processing pipelines with self-modification capabilities.\n\n")

	sb.WriteString("## Core Capabilities\n\n")
	sb.WriteString("### Document Processing\n")
	sb.WriteString("- Extract text from PDF, DOCX, XLSX, images\n")
	sb.WriteString("- OCR processing (native Tesseract + HTTP service)\n")
	sb.WriteString("- Multi-format document transformation\n")
	sb.WriteString("- Intelligent chunking and context enrichment\n\n")

	sb.WriteString("### Content Analysis\n")
	sb.WriteString("- Text analysis (sentiment, keywords, language)\n")
	sb.WriteString("- Structured data analysis (JSON, XML)\n")
	sb.WriteString("- Binary and image analysis\n")
	sb.WriteString("- Academic content processing\n\n")

	sb.WriteString("### Advanced Processing\n")
	sb.WriteString("- Named Entity Recognition (NER) with multilingual support\n")
	sb.WriteString("- PII detection and anonymization (GDPR-compliant)\n")
	sb.WriteString("- RAG (Retrieval-Augmented Generation) with vector embeddings\n")
	sb.WriteString("- Search indexing and full-text search\n\n")

	sb.WriteString("### Output Generation\n")
	sb.WriteString("- Document summarization\n")
	sb.WriteString("- Report generation with charts and tables\n")
	sb.WriteString("- Dataset creation and export\n")
	sb.WriteString("- Search-ready index generation\n\n")

	sb.WriteString("## Available Resources\n\n")
	sb.WriteString(fmt.Sprintf("- **Agents:** %d specialized processing agents\n", len(agents)))

	// Count cells by category
	categoryCount := make(map[string]int)
	for _, cell := range cells {
		categoryCount[cell.Category]++
	}

	sb.WriteString(fmt.Sprintf("- **Pipelines:** %d processing workflows\n", categoryCount["pipelines"]))
	sb.WriteString(fmt.Sprintf("- **Services:** %d backend services\n", categoryCount["services"]))
	sb.WriteString(fmt.Sprintf("- **Analysis:** %d analysis workflows\n", categoryCount["analysis"]))
	sb.WriteString(fmt.Sprintf("- **Synthesis:** %d output generators\n\n", categoryCount["synthesis"]))

	sb.WriteString("## How to Use\n\n")
	sb.WriteString("### Run a Cell\n")
	sb.WriteString("```bash\n")
	sb.WriteString("bin/orchestrator -config=workbench/config/cells/pipelines/<cell>.yaml\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Via Alfa\n")
	sb.WriteString("```bash\n")
	sb.WriteString("bin/alfa --enable-cellorg\n")
	sb.WriteString("> \"Process documents in input/ through anonymization pipeline\"\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Self-Modification\n\n")
	sb.WriteString("AGEN can modify its own codebase when enabled:\n")
	sb.WriteString("- Add new agents dynamically\n")
	sb.WriteString("- Modify cell configurations\n")
	sb.WriteString("- Update processing pipelines\n")
	sb.WriteString("- All changes are version-controlled via VCR\n\n")

	path := filepath.Join(e.manualDir, "capabilities.md")
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
