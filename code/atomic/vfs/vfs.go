package vfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// VFS represents a virtual file system rooted at a specific directory
type VFS struct {
	root     string
	readonly bool
}

// NewVFS initializes a VFS with the given root directory
func NewVFS(root string, readonly bool) (*VFS, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("invalid root path: %w", err)
	}

	// Ensure root exists
	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root: %w", err)
	}

	return &VFS{
		root:     abs,
		readonly: readonly,
	}, nil
}

// Root returns the absolute root path
func (vfs *VFS) Root() string {
	return vfs.root
}

// IsReadOnly returns whether this VFS is read-only
func (vfs *VFS) IsReadOnly() bool {
	return vfs.readonly
}

// validatePath ensures path is within VFS root and doesn't escape
func (vfs *VFS) validatePath(parts ...string) (string, error) {
	rel := filepath.Join(parts...)

	// Check for path traversal attempts
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", rel)
	}

	// Build absolute path
	abs := filepath.Join(vfs.root, rel)

	// Clean the path
	abs = filepath.Clean(abs)

	// Ensure it's still within root
	if !strings.HasPrefix(abs, vfs.root) {
		return "", fmt.Errorf("path outside VFS root: %s", rel)
	}

	return abs, nil
}

// Path returns the absolute path for relative parts
func (vfs *VFS) Path(parts ...string) (string, error) {
	return vfs.validatePath(parts...)
}

// Read reads a file's contents
func (vfs *VFS) Read(parts ...string) ([]byte, error) {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}

// ReadString reads a file as a string
func (vfs *VFS) ReadString(parts ...string) (string, error) {
	data, err := vfs.Read(parts...)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Write writes content to a file
func (vfs *VFS) Write(content []byte, parts ...string) error {
	if vfs.readonly {
		return fmt.Errorf("VFS is read-only")
	}

	path, err := vfs.validatePath(parts...)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, content, 0644)
}

// WriteString writes a string to a file
func (vfs *VFS) WriteString(content string, parts ...string) error {
	return vfs.Write([]byte(content), parts...)
}

// Append appends content to a file
func (vfs *VFS) Append(content []byte, parts ...string) error {
	if vfs.readonly {
		return fmt.Errorf("VFS is read-only")
	}

	path, err := vfs.validatePath(parts...)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	return err
}

// Delete removes a file or directory
func (vfs *VFS) Delete(parts ...string) error {
	if vfs.readonly {
		return fmt.Errorf("VFS is read-only")
	}

	path, err := vfs.validatePath(parts...)
	if err != nil {
		return err
	}

	return os.RemoveAll(path)
}

// Exists checks if a path exists
func (vfs *VFS) Exists(parts ...string) bool {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// IsDir checks if path is a directory
func (vfs *VFS) IsDir(parts ...string) bool {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// IsFile checks if path is a regular file
func (vfs *VFS) IsFile(parts ...string) bool {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

// List returns entries in a directory
func (vfs *VFS) List(parts ...string) ([]os.FileInfo, error) {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	infos := make([]os.FileInfo, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos[i] = info
	}

	return infos, nil
}

// Walk traverses the VFS directory tree
func (vfs *VFS) Walk(fn filepath.WalkFunc, parts ...string) error {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return err
	}

	return filepath.Walk(path, fn)
}

// Copy copies a file within the VFS
func (vfs *VFS) Copy(srcParts []string, dstParts []string) error {
	if vfs.readonly {
		return fmt.Errorf("VFS is read-only")
	}

	srcPath, err := vfs.validatePath(srcParts...)
	if err != nil {
		return err
	}

	dstPath, err := vfs.validatePath(dstParts...)
	if err != nil {
		return err
	}

	// Open source file
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy contents
	_, err = io.Copy(dst, src)
	return err
}

// Move moves/renames a file within the VFS
func (vfs *VFS) Move(srcParts []string, dstParts []string) error {
	if vfs.readonly {
		return fmt.Errorf("VFS is read-only")
	}

	srcPath, err := vfs.validatePath(srcParts...)
	if err != nil {
		return err
	}

	dstPath, err := vfs.validatePath(dstParts...)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	return os.Rename(srcPath, dstPath)
}

// Mkdir creates a directory
func (vfs *VFS) Mkdir(parts ...string) error {
	if vfs.readonly {
		return fmt.Errorf("VFS is read-only")
	}

	path, err := vfs.validatePath(parts...)
	if err != nil {
		return err
	}

	return os.MkdirAll(path, 0755)
}

// Stat returns file info
func (vfs *VFS) Stat(parts ...string) (os.FileInfo, error) {
	path, err := vfs.validatePath(parts...)
	if err != nil {
		return nil, err
	}

	return os.Stat(path)
}
