package vcr

import (
	"fmt"
	"path/filepath"
	"strings"
)

// VCRManager manages multiple VCR instances for framework, workbench, and project repos
type VCRManager struct {
	frameworkVCR *Vcr // AGEN framework modifications
	workbenchVCR *Vcr // Workbench config/state (optional)
	projectVCR   *Vcr // Current project work

	frameworkRoot string
	workbenchRoot string
	projectRoot   string
}

// NewVCRManager creates a new VCR manager with separate tracking for framework, workbench, and project
func NewVCRManager(frameworkRoot, workbenchRoot, projectRoot string) (*VCRManager, error) {
	// Normalize paths
	frameworkRoot, _ = filepath.Abs(frameworkRoot)
	workbenchRoot, _ = filepath.Abs(workbenchRoot)
	projectRoot, _ = filepath.Abs(projectRoot)

	return &VCRManager{
		frameworkVCR:  NewVcr("agen-framework", frameworkRoot),
		workbenchVCR:  NewVcr("workbench", workbenchRoot),
		projectVCR:    NewVcr("project", projectRoot),
		frameworkRoot: frameworkRoot,
		workbenchRoot: workbenchRoot,
		projectRoot:   projectRoot,
	}, nil
}

// GetVCRForPath returns the appropriate VCR based on file path
func (m *VCRManager) GetVCRForPath(filePath string) *Vcr {
	absPath, _ := filepath.Abs(filePath)

	// Framework modifications (code/, reflect/, guidelines/, drivers/)
	if strings.HasPrefix(absPath, m.frameworkRoot+"/code/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/reflect/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/guidelines/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/drivers/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/trained/") {
		return m.frameworkVCR
	}

	// Workbench modifications (workbench/config/, workbench/demos/)
	if strings.HasPrefix(absPath, m.workbenchRoot+"/config/") ||
		strings.HasPrefix(absPath, m.workbenchRoot+"/demos/") {
		return m.workbenchVCR
	}

	// Project modifications (everything in project root)
	return m.projectVCR
}

// Commit commits changes to the appropriate repository based on file path
func (m *VCRManager) Commit(filePath, message string) string {
	vcr := m.GetVCRForPath(filePath)
	return vcr.Commit(message)
}

// CommitMultiple commits multiple files, grouping them by repository
// Returns map of repo type to commit hash
func (m *VCRManager) CommitMultiple(files []string, message string) map[string]string {
	// Determine which VCRs are affected
	hasFramework := false
	hasWorkbench := false
	hasProject := false

	for _, file := range files {
		vcr := m.GetVCRForPath(file)
		switch vcr {
		case m.frameworkVCR:
			hasFramework = true
		case m.workbenchVCR:
			hasWorkbench = true
		case m.projectVCR:
			hasProject = true
		}
	}

	commits := make(map[string]string)

	// Commit to each affected VCR once
	if hasFramework {
		frameworkMsg := fmt.Sprintf("[Framework] %s", message)
		hash := m.frameworkVCR.Commit(frameworkMsg)
		if hash != "" && hash != "?" {
			commits["framework"] = hash
		}
	}

	if hasWorkbench {
		workbenchMsg := fmt.Sprintf("[Workbench] %s", message)
		hash := m.workbenchVCR.Commit(workbenchMsg)
		if hash != "" && hash != "?" {
			commits["workbench"] = hash
		}
	}

	if hasProject {
		hash := m.projectVCR.Commit(message)
		if hash != "" && hash != "?" {
			commits["project"] = hash
		}
	}

	return commits
}

// FrameworkVCR returns the framework VCR instance
func (m *VCRManager) FrameworkVCR() *Vcr {
	return m.frameworkVCR
}

// WorkbenchVCR returns the workbench VCR instance
func (m *VCRManager) WorkbenchVCR() *Vcr {
	return m.workbenchVCR
}

// ProjectVCR returns the project VCR instance
func (m *VCRManager) ProjectVCR() *Vcr {
	return m.projectVCR
}

// GetRepoType returns a string indicating which repo a file belongs to
func (m *VCRManager) GetRepoType(filePath string) string {
	vcr := m.GetVCRForPath(filePath)
	switch vcr {
	case m.frameworkVCR:
		return "framework"
	case m.workbenchVCR:
		return "workbench"
	case m.projectVCR:
		return "project"
	default:
		return "unknown"
	}
}
