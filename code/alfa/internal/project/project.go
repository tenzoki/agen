package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Metadata holds information about a project
type Metadata struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	LastModified time.Time `json:"last_modified"`
	LastCommit   string    `json:"last_commit,omitempty"`
	Branch       string    `json:"branch,omitempty"`
	IsGitRepo    bool      `json:"is_git_repo"`
}

// Manager handles multiple projects within a workbench
type Manager struct {
	workbenchRoot string
	projectsDir   string
	remotesDir    string
}

// NewManager creates a new project manager
func NewManager(workbenchRoot string) *Manager {
	return &Manager{
		workbenchRoot: workbenchRoot,
		projectsDir:   filepath.Join(workbenchRoot, "projects"),
		remotesDir:    filepath.Join(workbenchRoot, ".git-remotes"),
	}
}

// EnsureDirectories creates the projects and remotes directories if they don't exist
func (m *Manager) EnsureDirectories() error {
	if err := os.MkdirAll(m.projectsDir, 0755); err != nil {
		return fmt.Errorf("failed to create projects directory: %w", err)
	}
	if err := os.MkdirAll(m.remotesDir, 0755); err != nil {
		return fmt.Errorf("failed to create remotes directory: %w", err)
	}
	return nil
}

// List returns all projects in the workbench
func (m *Manager) List() ([]Metadata, error) {
	entries, err := os.ReadDir(m.projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Metadata{}, nil
		}
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []Metadata
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		meta, err := m.GetMetadata(entry.Name())
		if err != nil {
			// Skip projects we can't read
			continue
		}
		projects = append(projects, meta)
	}

	return projects, nil
}

// GetMetadata retrieves metadata for a specific project
func (m *Manager) GetMetadata(name string) (Metadata, error) {
	projectPath := filepath.Join(m.projectsDir, name)

	info, err := os.Stat(projectPath)
	if err != nil {
		return Metadata{}, fmt.Errorf("project '%s' not found: %w", name, err)
	}

	meta := Metadata{
		Name:         name,
		Path:         projectPath,
		LastModified: info.ModTime(),
	}

	// Check if it's a git repo
	gitDir := filepath.Join(projectPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		meta.IsGitRepo = true

		// Get current branch
		if branch, err := m.getCurrentBranch(projectPath); err == nil {
			meta.Branch = branch
		}

		// Get last commit message
		if commit, err := m.getLastCommit(projectPath); err == nil {
			meta.LastCommit = commit
		}
	}

	return meta, nil
}

// Exists checks if a project exists
func (m *Manager) Exists(name string) bool {
	projectPath := filepath.Join(m.projectsDir, name)
	info, err := os.Stat(projectPath)
	return err == nil && info.IsDir()
}

// Create creates a new project with git initialization
func (m *Manager) Create(name string) error {
	if m.Exists(name) {
		return fmt.Errorf("project '%s' already exists", name)
	}

	projectPath := filepath.Join(m.projectsDir, name)
	remotePath := filepath.Join(m.remotesDir, name+".git")

	// Create project directory
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Initialize git repo
	if err := m.runGitCommand(projectPath, "init"); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}

	// If a backup exists from a previously deleted project, remove it
	if _, err := os.Stat(remotePath); err == nil {
		if err := os.RemoveAll(remotePath); err != nil {
			return fmt.Errorf("failed to remove old backup: %w", err)
		}
	}

	// Create bare remote repo directory
	if err := os.MkdirAll(remotePath, 0755); err != nil {
		return fmt.Errorf("failed to create bare remote directory: %w", err)
	}

	// Create bare remote repo
	if err := m.runGitCommand(remotePath, "init", "--bare"); err != nil {
		return fmt.Errorf("failed to create bare remote: %w", err)
	}

	// Add remote
	if err := m.runGitCommand(projectPath, "remote", "add", "workbench", remotePath); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	// Create initial commit
	readmePath := filepath.Join(projectPath, "README.md")
	if err := os.WriteFile(readmePath, []byte(fmt.Sprintf("# %s\n\nCreated: %s\n", name, time.Now().Format(time.RFC3339))), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	if err := m.runGitCommand(projectPath, "add", "README.md"); err != nil {
		return fmt.Errorf("failed to add README: %w", err)
	}

	if err := m.runGitCommand(projectPath, "commit", "-m", "Initial commit"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	// Push to workbench remote
	if err := m.runGitCommand(projectPath, "push", "-u", "workbench", "main"); err != nil {
		// Try "master" if "main" doesn't exist
		if err := m.runGitCommand(projectPath, "push", "-u", "workbench", "master"); err != nil {
			return fmt.Errorf("failed to push initial commit: %w", err)
		}
	}

	return nil
}

// Delete removes a project (but keeps the bare remote for recovery)
func (m *Manager) Delete(name string) error {
	if !m.Exists(name) {
		return fmt.Errorf("project '%s' does not exist", name)
	}

	projectPath := filepath.Join(m.projectsDir, name)

	// Remove project directory
	if err := os.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// Restore recreates a project from its bare remote
func (m *Manager) Restore(name string) error {
	if m.Exists(name) {
		return fmt.Errorf("project '%s' already exists", name)
	}

	remotePath := filepath.Join(m.remotesDir, name+".git")
	if _, err := os.Stat(remotePath); err != nil {
		return fmt.Errorf("no backup found for project '%s'", name)
	}

	projectPath := filepath.Join(m.projectsDir, name)

	// Clone from bare repo
	cmd := exec.Command("git", "clone", remotePath, projectPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore project: %w", err)
	}

	return nil
}

// GetProjectPath returns the full path to a project
func (m *Manager) GetProjectPath(name string) string {
	return filepath.Join(m.projectsDir, name)
}

// GetRemotePath returns the full path to a project's bare remote
func (m *Manager) GetRemotePath(name string) string {
	return filepath.Join(m.remotesDir, name+".git")
}

// Helper functions

func (m *Manager) runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) getCurrentBranch(projectPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output[:len(output)-1]), nil // Remove trailing newline
}

func (m *Manager) getLastCommit(projectPath string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%s")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
