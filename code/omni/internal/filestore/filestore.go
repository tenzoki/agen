// Package filestore provides content-addressable file storage with deduplication, compression, and encryption
package filestore

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// FileStore interface defines content-addressable file storage operations
type FileStore interface {
	// Basic blob operations
	Store(data []byte, metadata map[string]string) (string, error) // Returns content hash
	StoreWithKey(key string, data []byte, metadata map[string]string) error
	Retrieve(hash string) ([]byte, map[string]string, error)
	Delete(hash string) error
	Exists(hash string) (bool, error)

	// Stream operations for large files
	StoreStream(reader io.Reader, metadata map[string]string) (string, int64, error) // Returns hash and size
	RetrieveStream(hash string) (io.ReadCloser, map[string]string, error)

	// Content operations
	GetContentHash(data []byte) string
	FindByHash(hash string) ([]string, error) // Find all keys with this hash
	FindByPrefix(prefix string) ([]string, error)

	// Deduplication and compression
	EnableCompression(enabled bool)
	EnableDeduplication(enabled bool)
	GetDeduplicationStats() (*DeduplicationStats, error)

	// Metadata operations
	SetMetadata(hash string, metadata map[string]string) error
	GetMetadata(hash string) (map[string]string, error)
	UpdateMetadata(hash string, updates map[string]string) error

	// Statistics
	Stats() (*FileStoreStats, error)
	Close() error
}

// Configuration types
type Config struct {
	StorageDir          string `json:"storage_dir"`
	EnableEncryption    bool   `json:"enable_encryption"`
	EncryptionKey       string `json:"encryption_key,omitempty"`
	EnableCompression   bool   `json:"enable_compression"`
	CompressionLevel    int    `json:"compression_level"`
	EnableDeduplication bool   `json:"enable_deduplication"`
	MaxFileSize         int64  `json:"max_file_size"`
	IndexingEnabled     bool   `json:"indexing_enabled"`
}

// Statistics types
type FileStoreStats struct {
	FileCount        int64               `json:"file_count"`
	TotalSize        int64               `json:"total_size"`
	CompressedSize   int64               `json:"compressed_size"`
	AverageSize      float64             `json:"average_size"`
	ContentTypes     map[string]int64    `json:"content_types"`
	LastAccess       time.Time           `json:"last_access"`
	IndexedFiles     int64               `json:"indexed_files"`
	CompressionRatio float64             `json:"compression_ratio"`
	Deduplication    *DeduplicationStats `json:"deduplication,omitempty"`
}

type DeduplicationStats struct {
	TotalFiles        int64   `json:"total_files"`
	UniqueFiles       int64   `json:"unique_files"`
	DuplicateFiles    int64   `json:"duplicate_files"`
	SpaceSaved        int64   `json:"space_saved"`
	DeduplicationRate float64 `json:"deduplication_rate"`
}

// Implementation
type fileStoreImpl struct {
	config             *Config
	storageDir         string
	metadataDB         *badger.DB
	cipher             cipher.Block
	compressionEnabled bool
	deduplication      bool
	hashIndex          map[string][]string // hash -> []keys
	stats              *FileStoreStats
	mu                 sync.RWMutex
}

// NewFileStore creates a new file storage instance
func NewFileStore(config *Config, dataDir string) (FileStore, error) {
	fs := &fileStoreImpl{
		config:             config,
		storageDir:         filepath.Join(dataDir, config.StorageDir),
		compressionEnabled: config.EnableCompression,
		deduplication:      config.EnableDeduplication,
		hashIndex:          make(map[string][]string),
		stats: &FileStoreStats{
			ContentTypes: make(map[string]int64),
			LastAccess:   time.Now(),
		},
	}

	// Create storage directory
	if err := os.MkdirAll(fs.storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Initialize metadata database
	metadataPath := filepath.Join(fs.storageDir, "metadata")
	opts := badger.DefaultOptions(metadataPath)
	opts.Logger = nil // Disable logging for metadata DB
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata database: %w", err)
	}
	fs.metadataDB = db

	// Initialize encryption if enabled
	if config.EnableEncryption && config.EncryptionKey != "" {
		key := []byte(config.EncryptionKey)
		if len(key) != 32 {
			// Pad or truncate to 32 bytes for AES-256
			paddedKey := make([]byte, 32)
			copy(paddedKey, key)
			key = paddedKey
		}
		cipher, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize encryption: %w", err)
		}
		fs.cipher = cipher
	}

	// Load existing index
	if err := fs.loadIndex(); err != nil {
		return nil, fmt.Errorf("failed to load index: %w", err)
	}

	return fs, nil
}

// Store stores data and returns the content hash
func (fs *fileStoreImpl) Store(data []byte, metadata map[string]string) (string, error) {
	hash := fs.GetContentHash(data)

	// Check if already exists and deduplication is enabled
	if fs.deduplication {
		if exists, _ := fs.Exists(hash); exists {
			fs.updateStats(hash, int64(len(data)), metadata)
			return hash, nil
		}
	}

	// Process data (compress, encrypt)
	processedData, err := fs.processData(data)
	if err != nil {
		return "", fmt.Errorf("failed to process data: %w", err)
	}

	// Store file
	filePath := fs.getFilePath(hash)
	if err := fs.ensureDir(filepath.Dir(filePath)); err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, processedData, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Store metadata
	if err := fs.storeMetadata(hash, metadata, int64(len(data))); err != nil {
		return "", err
	}

	fs.updateStats(hash, int64(len(data)), metadata)
	return hash, nil
}

// StoreWithKey stores data with a specific key
func (fs *fileStoreImpl) StoreWithKey(key string, data []byte, metadata map[string]string) error {
	hash, err := fs.Store(data, metadata)
	if err != nil {
		return err
	}

	// Associate key with hash
	fs.mu.Lock()
	if keys, exists := fs.hashIndex[hash]; exists {
		// Check if key already exists
		for _, k := range keys {
			if k == key {
				fs.mu.Unlock()
				return nil
			}
		}
		fs.hashIndex[hash] = append(keys, key)
	} else {
		fs.hashIndex[hash] = []string{key}
	}
	fs.mu.Unlock()

	// Store key-hash mapping in metadata DB
	return fs.metadataDB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("key:"+key), []byte(hash))
	})
}

// Retrieve retrieves data by hash
func (fs *fileStoreImpl) Retrieve(hash string) ([]byte, map[string]string, error) {
	filePath := fs.getFilePath(hash)

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("file not found: %s", hash)
		}
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Process data (decrypt, decompress)
	originalData, err := fs.unprocessData(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process data: %w", err)
	}

	// Get metadata
	metadata, err := fs.getMetadata(hash)
	if err != nil {
		return nil, nil, err
	}

	fs.stats.LastAccess = time.Now()
	return originalData, metadata, nil
}

// Delete removes a file by hash
func (fs *fileStoreImpl) Delete(hash string) error {
	filePath := fs.getFilePath(hash)

	// Remove file
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	// Remove metadata
	return fs.metadataDB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("meta:" + hash))
	})
}

// Exists checks if a file exists by hash
func (fs *fileStoreImpl) Exists(hash string) (bool, error) {
	filePath := fs.getFilePath(hash)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// StoreStream stores data from a reader
func (fs *fileStoreImpl) StoreStream(reader io.Reader, metadata map[string]string) (string, int64, error) {
	// Read all data to calculate hash
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read stream: %w", err)
	}

	hash, err := fs.Store(data, metadata)
	return hash, int64(len(data)), err
}

// RetrieveStream retrieves data as a stream
func (fs *fileStoreImpl) RetrieveStream(hash string) (io.ReadCloser, map[string]string, error) {
	data, metadata, err := fs.Retrieve(hash)
	if err != nil {
		return nil, nil, err
	}

	return io.NopCloser(strings.NewReader(string(data))), metadata, nil
}

// GetContentHash calculates SHA256 hash of data
func (fs *fileStoreImpl) GetContentHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// FindByHash finds all keys associated with a hash
func (fs *fileStoreImpl) FindByHash(hash string) ([]string, error) {
	fs.mu.RLock()
	keys := fs.hashIndex[hash]
	fs.mu.RUnlock()
	return keys, nil
}

// FindByPrefix finds keys with a given prefix
func (fs *fileStoreImpl) FindByPrefix(prefix string) ([]string, error) {
	var keys []string
	err := fs.metadataDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefixBytes := []byte("key:" + prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := string(item.Key()[4:]) // Remove "key:" prefix
			keys = append(keys, key)
		}
		return nil
	})

	return keys, err
}

// EnableCompression enables or disables compression
func (fs *fileStoreImpl) EnableCompression(enabled bool) {
	fs.mu.Lock()
	fs.compressionEnabled = enabled
	fs.mu.Unlock()
}

// EnableDeduplication enables or disables deduplication
func (fs *fileStoreImpl) EnableDeduplication(enabled bool) {
	fs.mu.Lock()
	fs.deduplication = enabled
	fs.mu.Unlock()
}

// GetDeduplicationStats returns deduplication statistics
func (fs *fileStoreImpl) GetDeduplicationStats() (*DeduplicationStats, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	totalFiles := int64(0)
	uniqueFiles := int64(len(fs.hashIndex))

	for _, keys := range fs.hashIndex {
		totalFiles += int64(len(keys))
	}

	duplicateFiles := totalFiles - uniqueFiles
	deduplicationRate := 0.0
	if totalFiles > 0 {
		deduplicationRate = float64(duplicateFiles) / float64(totalFiles)
	}

	return &DeduplicationStats{
		TotalFiles:        totalFiles,
		UniqueFiles:       uniqueFiles,
		DuplicateFiles:    duplicateFiles,
		SpaceSaved:        fs.stats.TotalSize - fs.stats.CompressedSize,
		DeduplicationRate: deduplicationRate,
	}, nil
}

// SetMetadata sets metadata for a hash
func (fs *fileStoreImpl) SetMetadata(hash string, metadata map[string]string) error {
	return fs.storeMetadata(hash, metadata, 0)
}

// GetMetadata gets metadata for a hash
func (fs *fileStoreImpl) GetMetadata(hash string) (map[string]string, error) {
	return fs.getMetadata(hash)
}

// UpdateMetadata updates metadata for a hash
func (fs *fileStoreImpl) UpdateMetadata(hash string, updates map[string]string) error {
	existing, err := fs.getMetadata(hash)
	if err != nil {
		return err
	}

	for k, v := range updates {
		existing[k] = v
	}

	return fs.storeMetadata(hash, existing, 0)
}

// Stats returns file store statistics
func (fs *fileStoreImpl) Stats() (*FileStoreStats, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Update stats
	fs.calculateStats()

	return fs.stats, nil
}

// Close closes the file store
func (fs *fileStoreImpl) Close() error {
	return fs.metadataDB.Close()
}

// Helper methods

func (fs *fileStoreImpl) getFilePath(hash string) string {
	// Create a 3-level directory structure: aa/bb/cc/hash
	if len(hash) < 6 {
		return filepath.Join(fs.storageDir, "misc", hash)
	}
	return filepath.Join(fs.storageDir, hash[:2], hash[2:4], hash[4:6], hash)
}

func (fs *fileStoreImpl) ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func (fs *fileStoreImpl) processData(data []byte) ([]byte, error) {
	result := data

	// Compress if enabled
	if fs.compressionEnabled {
		var buf strings.Builder
		gzipWriter, err := gzip.NewWriterLevel(&buf, fs.config.CompressionLevel)
		if err != nil {
			return nil, err
		}
		_, err = gzipWriter.Write(data)
		if err != nil {
			return nil, err
		}
		gzipWriter.Close()
		result = []byte(buf.String())
	}

	// Encrypt if enabled
	if fs.cipher != nil {
		// Simple GCM encryption
		gcm, err := cipher.NewGCM(fs.cipher)
		if err != nil {
			return nil, err
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		ciphertext := gcm.Seal(nonce, nonce, result, nil)
		result = ciphertext
	}

	return result, nil
}

func (fs *fileStoreImpl) unprocessData(data []byte) ([]byte, error) {
	result := data

	// Decrypt if enabled
	if fs.cipher != nil {
		gcm, err := cipher.NewGCM(fs.cipher)
		if err != nil {
			return nil, err
		}

		nonceSize := gcm.NonceSize()
		if len(data) < nonceSize {
			return nil, fmt.Errorf("ciphertext too short")
		}

		nonce, ciphertext := data[:nonceSize], data[nonceSize:]
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, err
		}
		result = plaintext
	}

	// Decompress if enabled
	if fs.compressionEnabled {
		gzipReader, err := gzip.NewReader(strings.NewReader(string(result)))
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()

		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			return nil, err
		}
		result = decompressed
	}

	return result, nil
}

func (fs *fileStoreImpl) storeMetadata(hash string, metadata map[string]string, size int64) error {
	meta := FileMetadata{
		Hash:       hash,
		Metadata:   metadata,
		Size:       size,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Compressed: fs.compressionEnabled,
		Encrypted:  fs.cipher != nil,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return fs.metadataDB.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("meta:"+hash), data)
	})
}

func (fs *fileStoreImpl) getMetadata(hash string) (map[string]string, error) {
	var metadata map[string]string

	err := fs.metadataDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("meta:" + hash))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var meta FileMetadata
			if err := json.Unmarshal(val, &meta); err != nil {
				return err
			}
			metadata = meta.Metadata
			return nil
		})
	})

	return metadata, err
}

func (fs *fileStoreImpl) loadIndex() error {
	return fs.metadataDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			if strings.HasPrefix(key, "key:") {
				// Key-hash mapping
				keyName := key[4:]
				err := item.Value(func(val []byte) error {
					hash := string(val)
					fs.mu.Lock()
					if keys, exists := fs.hashIndex[hash]; exists {
						fs.hashIndex[hash] = append(keys, keyName)
					} else {
						fs.hashIndex[hash] = []string{keyName}
					}
					fs.mu.Unlock()
					return nil
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (fs *fileStoreImpl) calculateStats() {
	fileCount := int64(0)
	totalSize := int64(0)

	filepath.WalkDir(fs.storageDir, func(path string, dirEntry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !dirEntry.IsDir() && dirEntry.Name() != "MANIFEST" && dirEntry.Name() != "KEYREGISTRY" {
			info, err := dirEntry.Info()
			if err == nil {
				fileCount++
				totalSize += info.Size()
			}
		}
		return nil
	})

	fs.stats.FileCount = fileCount
	fs.stats.TotalSize = totalSize
	if fileCount > 0 {
		fs.stats.AverageSize = float64(totalSize) / float64(fileCount)
	}
}

func (fs *fileStoreImpl) updateStats(hash string, size int64, metadata map[string]string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.stats.FileCount++
	fs.stats.TotalSize += size

	if contentType, exists := metadata["content_type"]; exists {
		fs.stats.ContentTypes[contentType]++
	}
}

// FileMetadata represents metadata stored for each file
type FileMetadata struct {
	Hash       string            `json:"hash"`
	Metadata   map[string]string `json:"metadata"`
	Size       int64             `json:"size"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Compressed bool              `json:"compressed"`
	Encrypted  bool              `json:"encrypted"`
}

// DefaultConfig returns a default configuration for the file store
func DefaultConfig() *Config {
	return &Config{
		StorageDir:          "files",
		EnableEncryption:    false,
		EnableCompression:   true,
		CompressionLevel:    6,
		EnableDeduplication: true,
		MaxFileSize:         100 * 1024 * 1024, // 100MB
		IndexingEnabled:     true,
	}
}
