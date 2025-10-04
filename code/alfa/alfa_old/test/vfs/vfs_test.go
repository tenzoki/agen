package vfs_test

import (
	"os"
	"path/filepath"
	"testing"

	"alfa/internal/vfs"
)

func TestNewVFS(t *testing.T) {
	tmpDir := t.TempDir()

	v, err := vfs.NewVFS(tmpDir, false)
	if err != nil {
		t.Fatalf("NewVFS failed: %v", err)
	}

	if v.Root() != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, v.Root())
	}

	if v.IsReadOnly() {
		t.Error("Expected read-write VFS")
	}
}

func TestNewVFS_ReadOnly(t *testing.T) {
	tmpDir := t.TempDir()

	v, err := vfs.NewVFS(tmpDir, true)
	if err != nil {
		t.Fatalf("NewVFS failed: %v", err)
	}

	if !v.IsReadOnly() {
		t.Error("Expected read-only VFS")
	}
}

func TestVFS_WriteAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	content := "Hello, World!"
	err := v.WriteString(content, "test.txt")
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	read, err := v.ReadString("test.txt")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if read != content {
		t.Errorf("Expected %q, got %q", content, read)
	}
}

func TestVFS_ReadOnly_Write(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, true)

	err := v.WriteString("test", "test.txt")
	if err == nil {
		t.Error("Expected write to fail on read-only VFS")
	}
}

func TestVFS_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	// Attempt path traversal
	err := v.WriteString("malicious", "..", "..", "etc", "passwd")
	if err == nil {
		t.Error("Expected path traversal to be blocked")
	}

	_, err = v.Path("..", "..", "etc", "passwd")
	if err == nil {
		t.Error("Expected path traversal validation to fail")
	}
}

func TestVFS_PathTraversal_Dots(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	// Try various path traversal attacks
	testCases := [][]string{
		{"../etc/passwd"},
		{"..", "etc", "passwd"},
		{"foo", "..", "..", "etc"},
	}

	for _, tc := range testCases {
		err := v.WriteString("malicious", tc...)
		if err == nil {
			t.Errorf("Path traversal not blocked: %v", tc)
		}
	}
}

func TestVFS_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	if v.Exists("nonexistent.txt") {
		t.Error("Expected file to not exist")
	}

	v.WriteString("content", "existing.txt")

	if !v.Exists("existing.txt") {
		t.Error("Expected file to exist")
	}
}

func TestVFS_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	v.WriteString("content", "todelete.txt")

	if !v.Exists("todelete.txt") {
		t.Fatal("File should exist before delete")
	}

	err := v.Delete("todelete.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if v.Exists("todelete.txt") {
		t.Error("File should not exist after delete")
	}
}

func TestVFS_Delete_ReadOnly(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file first
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	v, _ := vfs.NewVFS(tmpDir, true)

	err := v.Delete("test.txt")
	if err == nil {
		t.Error("Expected delete to fail on read-only VFS")
	}
}

func TestVFS_Mkdir(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	err := v.Mkdir("subdir", "nested")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	if !v.IsDir("subdir", "nested") {
		t.Error("Expected directory to exist")
	}
}

func TestVFS_IsDir(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	v.Mkdir("testdir")
	v.WriteString("content", "testfile.txt")

	if !v.IsDir("testdir") {
		t.Error("Expected testdir to be a directory")
	}

	if v.IsDir("testfile.txt") {
		t.Error("Expected testfile.txt to not be a directory")
	}
}

func TestVFS_IsFile(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	v.Mkdir("testdir")
	v.WriteString("content", "testfile.txt")

	if !v.IsFile("testfile.txt") {
		t.Error("Expected testfile.txt to be a file")
	}

	if v.IsFile("testdir") {
		t.Error("Expected testdir to not be a file")
	}
}

func TestVFS_List(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	v.WriteString("content1", "file1.txt")
	v.WriteString("content2", "file2.txt")
	v.Mkdir("subdir")

	entries, err := v.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestVFS_Copy(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	content := "original content"
	v.WriteString(content, "original.txt")

	err := v.Copy([]string{"original.txt"}, []string{"copy.txt"})
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	copied, err := v.ReadString("copy.txt")
	if err != nil {
		t.Fatalf("Read copy failed: %v", err)
	}

	if copied != content {
		t.Errorf("Expected %q, got %q", content, copied)
	}
}

func TestVFS_Move(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	content := "movable content"
	v.WriteString(content, "before.txt")

	err := v.Move([]string{"before.txt"}, []string{"after.txt"})
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	if v.Exists("before.txt") {
		t.Error("Original file should not exist after move")
	}

	if !v.Exists("after.txt") {
		t.Error("Moved file should exist")
	}

	moved, _ := v.ReadString("after.txt")
	if moved != content {
		t.Errorf("Expected %q, got %q", content, moved)
	}
}

func TestVFS_Append(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	v.WriteString("line1\n", "log.txt")
	v.Append([]byte("line2\n"), "log.txt")
	v.Append([]byte("line3\n"), "log.txt")

	content, _ := v.ReadString("log.txt")
	expected := "line1\nline2\nline3\n"

	if content != expected {
		t.Errorf("Expected %q, got %q", expected, content)
	}
}

func TestVFS_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	err := v.WriteString("nested content", "level1", "level2", "level3", "file.txt")
	if err != nil {
		t.Fatalf("Write to nested directory failed: %v", err)
	}

	if !v.Exists("level1", "level2", "level3", "file.txt") {
		t.Error("Nested file should exist")
	}

	content, _ := v.ReadString("level1", "level2", "level3", "file.txt")
	if content != "nested content" {
		t.Errorf("Expected 'nested content', got %q", content)
	}
}

func TestVFS_Walk(t *testing.T) {
	tmpDir := t.TempDir()
	v, _ := vfs.NewVFS(tmpDir, false)

	v.WriteString("content1", "file1.txt")
	v.WriteString("content2", "subdir", "file2.txt")
	v.WriteString("content3", "subdir", "nested", "file3.txt")

	count := 0
	err := v.Walk(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected to find 3 files, found %d", count)
	}
}
