package context

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Storage handles persisting context to disk
type Storage struct {
	path string
}

// NewStorage creates a new storage instance
func NewStorage(path string) *Storage {
	return &Storage{path: path}
}

// Save writes context data to disk
func (s *Storage) Save(data *StorageData) error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(s.path, jsonData, 0644)
}

// Load reads context data from disk
func (s *Storage) Load() *StorageData {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil
	}

	var storageData StorageData
	if err := json.Unmarshal(data, &storageData); err != nil {
		return nil
	}

	return &storageData
}

// Exists checks if storage file exists
func (s *Storage) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

// Delete removes the storage file
func (s *Storage) Delete() error {
	return os.Remove(s.path)
}
