package omnistore

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOmniStore_FileStorageIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Test file storage access
	fileStore := store.Files()
	require.NotNil(t, fileStore)

	// Test basic file operations
	testData := []byte("Hello from OmniStore!")
	metadata := map[string]string{
		"content_type": "text/plain",
		"source":       "omnistore_test",
	}

	hash, err := fileStore.Store(testData, metadata)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Retrieve data
	retrievedData, retrievedMetadata, err := fileStore.Retrieve(hash)
	require.NoError(t, err)
	assert.Equal(t, testData, retrievedData)
	assert.Equal(t, metadata, retrievedMetadata)

	// Test with key
	err = fileStore.StoreWithKey("test_document", testData, metadata)
	require.NoError(t, err)

	keys, err := fileStore.FindByHash(hash)
	require.NoError(t, err)
	assert.Contains(t, keys, "test_document")
}

func TestOmniStore_CrossComponentQuery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Store some test data in different components

	// KV data
	kvStore := store.KV()
	err = kvStore.Set("user:123", []byte(`{"name": "Alice", "age": 30}`))
	require.NoError(t, err)

	// File data
	fileStore := store.Files()
	testFile := []byte("User profile photo data")
	fileMetadata := map[string]string{
		"user_id":      "123",
		"content_type": "image/jpeg",
	}
	fileHash, err := fileStore.Store(testFile, fileMetadata)
	require.NoError(t, err)

	// Test cross-component query
	crossQuery := &CrossQueryRequest{
		KVQuery: &KVQueryComponent{
			Keys: []string{"user:123"},
		},
		FileQuery: &FileQueryComponent{
			Hashes: []string{fileHash},
		},
	}

	result, err := store.ExecuteCrossQuery(crossQuery)
	require.NoError(t, err)

	// Verify results
	assert.NotNil(t, result.KVResults)
	assert.Contains(t, result.KVResults, "user:123")

	assert.NotNil(t, result.FileResults)
	assert.Len(t, result.FileResults, 1)
	assert.Equal(t, fileHash, result.FileResults[0].Hash)

	assert.Greater(t, result.ExecutionTime, int64(0))
	assert.Contains(t, result.ComponentTimes, "kv")
	assert.Contains(t, result.ComponentTimes, "files")
}

func TestOmniStore_Stats(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Add some data to generate stats
	kvStore := store.KV()
	err = kvStore.Set("test_key", []byte("test_value"))
	require.NoError(t, err)

	fileStore := store.Files()
	testData := []byte("Statistics test data")
	_, err = fileStore.Store(testData, map[string]string{"type": "test"})
	require.NoError(t, err)

	stats, err := store.GetStats()
	require.NoError(t, err)

	assert.NotNil(t, stats.KV)
	assert.NotNil(t, stats.Graph)
	assert.NotNil(t, stats.Files)
	assert.Greater(t, stats.TotalStorageSize, int64(0))
	assert.Greater(t, stats.Uptime, int64(0))
}

func TestOmniStore_Health(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	health, err := store.GetHealth()
	require.NoError(t, err)

	assert.Equal(t, Healthy, health.Status)
	assert.Contains(t, health.Components, "kv")
	assert.Contains(t, health.Components, "graph")
	assert.Contains(t, health.Components, "files")
	assert.Contains(t, health.Components, "search")

	for componentName, componentHealth := range health.Components {
		assert.Equal(t, Healthy, componentHealth.Status, "Component %s should be healthy", componentName)
	}
}

func TestOmniStore_Configuration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test with custom config
	config := DefaultConfig()
	config.DataDir = tempDir
	config.Files.EnableCompression = true
	config.Files.EnableDeduplication = true
	config.Files.MaxFileSize = 1024 * 1024 // 1MB

	store, err := NewOmniStore(config)
	require.NoError(t, err)
	defer store.Close()

	// Test configuration access
	storeConfig := store.GetConfig()
	assert.Equal(t, tempDir, storeConfig.DataDir)
	assert.True(t, storeConfig.Files.EnableCompression)
	assert.True(t, storeConfig.Files.EnableDeduplication)
	assert.Equal(t, int64(1024*1024), storeConfig.Files.MaxFileSize)

	// Test configuration update
	newConfig := *config
	newConfig.Performance.MaxConnections = 200
	err = store.UpdateConfig(&newConfig)
	require.NoError(t, err)

	updatedConfig := store.GetConfig()
	assert.Equal(t, 200, updatedConfig.Performance.MaxConnections)
}

func TestOmniStore_FileStreamOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	fileStore := store.Files()

	// Test stream storage
	testContent := "This is a stream test with more content to test streaming functionality"
	reader := strings.NewReader(testContent)
	metadata := map[string]string{
		"content_type": "text/plain",
		"stream_test":  "true",
	}

	hash, size, err := fileStore.StoreStream(reader, metadata)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Equal(t, int64(len(testContent)), size)

	// Test stream retrieval
	readCloser, retrievedMetadata, err := fileStore.RetrieveStream(hash)
	require.NoError(t, err)
	defer readCloser.Close()

	retrievedContent := make([]byte, len(testContent))
	n, err := readCloser.Read(retrievedContent)
	require.NoError(t, err)
	assert.Equal(t, len(testContent), n)
	assert.Equal(t, testContent, string(retrievedContent))
	assert.Equal(t, metadata, retrievedMetadata)
}

func TestOmniStore_FileDeduplication(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	config := DefaultConfig()
	config.DataDir = tempDir
	config.Files.EnableDeduplication = true

	store, err := NewOmniStore(config)
	require.NoError(t, err)
	defer store.Close()

	fileStore := store.Files()

	// Store the same content multiple times
	testData := []byte("Duplicate content for testing")
	metadata1 := map[string]string{"version": "1"}
	metadata2 := map[string]string{"version": "2"}

	hash1, err := fileStore.Store(testData, metadata1)
	require.NoError(t, err)

	hash2, err := fileStore.Store(testData, metadata2)
	require.NoError(t, err)

	// Should get same hash for same content
	assert.Equal(t, hash1, hash2)

	// Store with different keys
	err = fileStore.StoreWithKey("doc1", testData, metadata1)
	require.NoError(t, err)

	err = fileStore.StoreWithKey("doc2", testData, metadata2)
	require.NoError(t, err)

	// Both keys should map to same hash
	keys, err := fileStore.FindByHash(hash1)
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "doc1")
	assert.Contains(t, keys, "doc2")

	// Check deduplication stats
	stats, err := fileStore.GetDeduplicationStats()
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats.TotalFiles)
	assert.Equal(t, int64(1), stats.UniqueFiles)
	assert.Equal(t, int64(1), stats.DuplicateFiles)
}

func TestOmniStore_DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "./data", config.DataDir)
	// KV and Graph configurations are handled by the underlying storage layer
	assert.NotNil(t, config.Files)
	assert.NotNil(t, config.Search)
	assert.NotNil(t, config.Transaction)
	assert.NotNil(t, config.Performance)
	assert.NotNil(t, config.Security)

	// Check file store defaults
	fileConfig := config.Files
	assert.Equal(t, "files", fileConfig.StorageDir)
	assert.False(t, fileConfig.EnableEncryption)
	assert.True(t, fileConfig.EnableCompression)
	assert.Equal(t, 6, fileConfig.CompressionLevel)
	assert.True(t, fileConfig.EnableDeduplication)
	assert.Equal(t, int64(100*1024*1024), fileConfig.MaxFileSize)
	assert.True(t, fileConfig.IndexingEnabled)

	// Check search defaults
	searchConfig := config.Search
	assert.Equal(t, "search_index", searchConfig.IndexDir)
	assert.Equal(t, "standard", searchConfig.DefaultAnalyzer)
	assert.Equal(t, []string{"en"}, searchConfig.Languages)
	assert.Equal(t, 1000, searchConfig.MaxResults)
	assert.Equal(t, 100, searchConfig.IndexBatchSize)

	// Check performance defaults
	perfConfig := config.Performance
	assert.Equal(t, 100, perfConfig.MaxConnections)
	assert.Equal(t, int64(100*1024*1024), perfConfig.CacheSize)
	assert.Equal(t, 10, perfConfig.WorkerPoolSize)
}

func TestOmniStore_FileMetadataOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	fileStore := store.Files()

	testData := []byte("Metadata operations test")
	initialMetadata := map[string]string{
		"title":  "Test Document",
		"author": "Test Author",
	}

	hash, err := fileStore.Store(testData, initialMetadata)
	require.NoError(t, err)

	// Test metadata retrieval
	metadata, err := fileStore.GetMetadata(hash)
	require.NoError(t, err)
	assert.Equal(t, initialMetadata, metadata)

	// Test metadata update
	updates := map[string]string{
		"author":        "Updated Author",
		"last_modified": "2023-01-01",
	}
	err = fileStore.UpdateMetadata(hash, updates)
	require.NoError(t, err)

	updatedMetadata, err := fileStore.GetMetadata(hash)
	require.NoError(t, err)
	assert.Equal(t, "Test Document", updatedMetadata["title"])
	assert.Equal(t, "Updated Author", updatedMetadata["author"])
	assert.Equal(t, "2023-01-01", updatedMetadata["last_modified"])

	// Test metadata replacement
	newMetadata := map[string]string{
		"type": "replacement",
	}
	err = fileStore.SetMetadata(hash, newMetadata)
	require.NoError(t, err)

	finalMetadata, err := fileStore.GetMetadata(hash)
	require.NoError(t, err)
	assert.Equal(t, newMetadata, finalMetadata)
}

func TestOmniStore_CompactOperation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	// Add some data
	kvStore := store.KV()
	for i := 0; i < 100; i++ {
		err := kvStore.Set(fmt.Sprintf("key_%d", i), []byte(fmt.Sprintf("value_%d", i)))
		require.NoError(t, err)
	}

	fileStore := store.Files()
	for i := 0; i < 10; i++ {
		data := []byte(fmt.Sprintf("file_data_%d", i))
		_, err := fileStore.Store(data, map[string]string{"index": fmt.Sprintf("%d", i)})
		require.NoError(t, err)
	}

	// Test compact operation - currently returns "not implemented" error
	err = store.Compact()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compact functionality not yet implemented")

	// Verify data is still accessible even without compaction
	value, err := kvStore.Get("key_50")
	require.NoError(t, err)
	assert.Equal(t, []byte("value_50"), value)
}

// Benchmark tests
func TestOmniStore_ListKVKeys(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "omnistore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(t, err)
	defer store.Close()

	kvStore := store.KV()

	// Store test data with different prefixes
	testData := map[string][]byte{
		"user:alice":   []byte("Alice data"),
		"user:bob":     []byte("Bob data"),
		"user:charlie": []byte("Charlie data"),
		"config:app":   []byte("App config"),
		"config:db":    []byte("DB config"),
	}

	for key, value := range testData {
		err := kvStore.Set(key, value)
		require.NoError(t, err)
	}

	// Test ListKVKeys with "user:" prefix
	userKeys, err := store.ListKVKeys("user:", 10)
	require.NoError(t, err)
	assert.Len(t, userKeys, 3)
	assert.Contains(t, userKeys, "user:alice")
	assert.Contains(t, userKeys, "user:bob")
	assert.Contains(t, userKeys, "user:charlie")

	// Test ListKVKeys with "config:" prefix
	configKeys, err := store.ListKVKeys("config:", 10)
	require.NoError(t, err)
	assert.Len(t, configKeys, 2)
	assert.Contains(t, configKeys, "config:app")
	assert.Contains(t, configKeys, "config:db")

	// Test ListKVKeys with "u" prefix to get all user keys
	allUserKeys, err := store.ListKVKeys("u", 10)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(allUserKeys), 3)

	// Test ListKVKeys with limit
	limitedKeys, err := store.ListKVKeys("user:", 2)
	require.NoError(t, err)
	assert.Len(t, limitedKeys, 2)

	// Test ListKVKeys with non-existent prefix
	noKeys, err := store.ListKVKeys("nonexistent:", 10)
	require.NoError(t, err)
	assert.Len(t, noKeys, 0)
}

func BenchmarkOmniStore_FileOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "omnistore_benchmark")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	store, err := NewOmniStoreWithDefaults(tempDir)
	require.NoError(b, err)
	defer store.Close()

	fileStore := store.Files()
	testData := []byte("Benchmark test data")
	metadata := map[string]string{"benchmark": "true"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash, err := fileStore.Store(testData, metadata)
		require.NoError(b, err)

		_, _, err = fileStore.Retrieve(hash)
		require.NoError(b, err)
	}
}
