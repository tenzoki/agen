package vfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VFSManager manages multiple VFS instances for framework, workbench, and project access
type VFSManager struct {
	frameworkVFS *VFS // Full AGEN repo access
	workbenchVFS *VFS // Workbench access
	projectVFS   *VFS // Project isolation (default)

	frameworkRoot string
	workbenchRoot string
	projectRoot   string

	selfModificationEnabled bool
}

// NewVFSManager creates a new VFS manager with multiple root contexts
func NewVFSManager(frameworkRoot, workbenchRoot, projectRoot string, allowSelfMod bool) (*VFSManager, error) {
	// Normalize paths
	frameworkRoot, _ = filepath.Abs(frameworkRoot)
	workbenchRoot, _ = filepath.Abs(workbenchRoot)
	projectRoot, _ = filepath.Abs(projectRoot)

	// Create VFS instances (all read-write)
	frameworkVFS, err := NewVFS(frameworkRoot, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create framework VFS: %w", err)
	}

	workbenchVFS, err := NewVFS(workbenchRoot, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create workbench VFS: %w", err)
	}

	projectVFS, err := NewVFS(projectRoot, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create project VFS: %w", err)
	}

	return &VFSManager{
		frameworkVFS:            frameworkVFS,
		workbenchVFS:            workbenchVFS,
		projectVFS:              projectVFS,
		frameworkRoot:           frameworkRoot,
		workbenchRoot:           workbenchRoot,
		projectRoot:             projectRoot,
		selfModificationEnabled: allowSelfMod,
	}, nil
}

// GetVFS returns appropriate VFS based on path
func (m *VFSManager) GetVFS(filePath string) (*VFS, error) {
	absPath, _ := filepath.Abs(filePath)

	// Framework modification request
	if strings.HasPrefix(absPath, m.frameworkRoot+"/code/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/reflect/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/guidelines/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/drivers/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/trained/") {

		if !m.selfModificationEnabled {
			return nil, fmt.Errorf("self-modification disabled: cannot modify %s", filePath)
		}
		return m.frameworkVFS, nil
	}

	// Workbench modification
	if strings.HasPrefix(absPath, m.workbenchRoot) {
		return m.workbenchVFS, nil
	}

	// Project modification (default)
	return m.projectVFS, nil
}

// Read reads a file using the appropriate VFS
func (m *VFSManager) Read(filePath string) ([]byte, error) {
	vfs, err := m.GetVFS(filePath)
	if err != nil {
		return nil, err
	}
	// Convert absolute path to relative path within the VFS root
	relPath, err := filepath.Rel(vfs.Root(), filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	return vfs.Read(relPath)
}

// Write writes a file using the appropriate VFS
func (m *VFSManager) Write(filePath string, content []byte) error {
	vfs, err := m.GetVFS(filePath)
	if err != nil {
		return err
	}
	// Convert absolute path to relative path within the VFS root
	relPath, err := filepath.Rel(vfs.Root(), filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	return vfs.Write(content, relPath)
}

// Exists checks if a file exists using the appropriate VFS
func (m *VFSManager) Exists(filePath string) bool {
	vfs, err := m.GetVFS(filePath)
	if err != nil {
		return false
	}
	// Convert absolute path to relative path within the VFS root
	relPath, err := filepath.Rel(vfs.Root(), filePath)
	if err != nil {
		return false
	}
	return vfs.Exists(relPath)
}

// List lists files using the appropriate VFS
func (m *VFSManager) List(dirPath string) ([]os.FileInfo, error) {
	vfs, err := m.GetVFS(dirPath)
	if err != nil {
		return nil, err
	}
	// Convert absolute path to relative path within the VFS root
	relPath, err := filepath.Rel(vfs.Root(), dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path: %w", err)
	}
	return vfs.List(relPath)
}

// Delete deletes a file using the appropriate VFS
func (m *VFSManager) Delete(filePath string) error {
	vfs, err := m.GetVFS(filePath)
	if err != nil {
		return err
	}
	// Convert absolute path to relative path within the VFS root
	relPath, err := filepath.Rel(vfs.Root(), filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}
	return vfs.Delete(relPath)
}

// FrameworkVFS returns the framework VFS instance
func (m *VFSManager) FrameworkVFS() *VFS {
	return m.frameworkVFS
}

// WorkbenchVFS returns the workbench VFS instance
func (m *VFSManager) WorkbenchVFS() *VFS {
	return m.workbenchVFS
}

// ProjectVFS returns the project VFS instance
func (m *VFSManager) ProjectVFS() *VFS {
	return m.projectVFS
}

// EnableSelfModification enables framework modifications
func (m *VFSManager) EnableSelfModification() {
	m.selfModificationEnabled = true
}

// DisableSelfModification disables framework modifications
func (m *VFSManager) DisableSelfModification() {
	m.selfModificationEnabled = false
}

// IsSelfModificationEnabled returns whether self-modification is enabled
func (m *VFSManager) IsSelfModificationEnabled() bool {
	return m.selfModificationEnabled
}

// GetContextType returns the context type for a given path
func (m *VFSManager) GetContextType(filePath string) string {
	absPath, _ := filepath.Abs(filePath)

	if strings.HasPrefix(absPath, m.frameworkRoot+"/code/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/reflect/") ||
		strings.HasPrefix(absPath, m.frameworkRoot+"/guidelines/") {
		return "framework"
	}

	if strings.HasPrefix(absPath, m.workbenchRoot) {
		return "workbench"
	}

	return "project"
}
