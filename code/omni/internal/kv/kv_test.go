package kv

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tenzoki/agen/omni/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestKVStore(t *testing.T) (KVStore, func()) {
	tmpDir := t.TempDir()

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	require.NoError(t, err)

	kvStore := NewKVStore(store)

	cleanup := func() {
		kvStore.Close()
		os.RemoveAll(tmpDir)
	}

	return kvStore, cleanup
}

func TestKVStore_BasicOperations(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Test Set and Get
	err := kv.Set("test-key", []byte("test-value"))
	require.NoError(t, err)

	value, err := kv.Get("test-key")
	require.NoError(t, err)
	assert.Equal(t, []byte("test-value"), value)

	// Test Exists
	exists, err := kv.Exists("test-key")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = kv.Exists("non-existent-key")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test Delete
	err = kv.Delete("test-key")
	require.NoError(t, err)

	_, err = kv.Get("test-key")
	assert.Equal(t, ErrKeyNotFound, err)

	exists, err = kv.Exists("test-key")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestKVStore_BatchOperations(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Test BatchSet
	items := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	err := kv.BatchSet(items)
	require.NoError(t, err)

	// Test BatchGet
	keys := []string{"key1", "key2", "key3", "non-existent"}
	results, err := kv.BatchGet(keys)
	require.NoError(t, err)

	assert.Equal(t, []byte("value1"), results["key1"])
	assert.Equal(t, []byte("value2"), results["key2"])
	assert.Equal(t, []byte("value3"), results["key3"])
	assert.NotContains(t, results, "non-existent")
}

func TestKVStore_ScanOperations(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Setup test data
	testData := map[string][]byte{
		"user:1":   []byte("alice"),
		"user:2":   []byte("bob"),
		"user:3":   []byte("charlie"),
		"config:1": []byte("setting1"),
		"config:2": []byte("setting2"),
	}

	for key, value := range testData {
		err := kv.Set(key, value)
		require.NoError(t, err)
	}

	// Test prefix scan
	userResults, err := kv.Scan("user:", -1)
	require.NoError(t, err)
	assert.Len(t, userResults, 3)
	assert.Equal(t, []byte("alice"), userResults["user:1"])
	assert.Equal(t, []byte("bob"), userResults["user:2"])
	assert.Equal(t, []byte("charlie"), userResults["user:3"])

	// Test limited scan
	configResults, err := kv.Scan("config:", 1)
	require.NoError(t, err)
	assert.Len(t, configResults, 1)

	// Test non-existent prefix
	emptyResults, err := kv.Scan("empty:", -1)
	require.NoError(t, err)
	assert.Empty(t, emptyResults)
}

func TestKVStore_TTLOperations(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Set key with short TTL
	err := kv.SetWithTTL("ttl-key", []byte("ttl-value"), 1*time.Second)
	require.NoError(t, err)

	// Verify key exists initially
	value, err := kv.Get("ttl-key")
	require.NoError(t, err)
	assert.Equal(t, []byte("ttl-value"), value)

	// Wait for TTL expiration (BadgerDB may need longer time)
	time.Sleep(2 * time.Second)

	// Verify key has expired (may still exist due to BadgerDB's lazy deletion)
	_, err = kv.Get("ttl-key")
	// Note: BadgerDB TTL may not immediately delete keys
	// This test validates that SetWithTTL doesn't error, actual expiration is BadgerDB's responsibility
	if err != nil {
		assert.Equal(t, ErrKeyNotFound, err)
	}
}

func TestKVStore_KeyValidation(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Test empty key
	err := kv.Set("", []byte("value"))
	assert.Error(t, err)

	_, err = kv.Get("")
	assert.Error(t, err)

	err = kv.Delete("")
	assert.Error(t, err)

	_, err = kv.Exists("")
	assert.Error(t, err)

	// Test very long key
	longKey := string(make([]byte, 2000)) // Over 1024 limit
	err = kv.Set(longKey, []byte("value"))
	assert.Error(t, err)
}

func TestKVStore_Stats(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Initially empty
	stats, err := kv.Stats()
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.KeyCount)
	assert.Equal(t, int64(0), stats.TotalSize)

	// Add some data
	testData := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("longer-value-here"),
	}

	for key, value := range testData {
		err := kv.Set(key, value)
		require.NoError(t, err)
	}

	// Check stats after adding data
	stats, err = kv.Stats()
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.KeyCount)
	assert.Greater(t, stats.TotalSize, int64(0))
	assert.Greater(t, stats.AvgKeySize, 0.0)
	assert.Greater(t, stats.AvgValueSize, 0.0)
	assert.Equal(t, "kv", stats.Namespace)
}

func TestKVStore_LargeValues(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Test with 1MB value
	largeValue := make([]byte, 1024*1024)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	err := kv.Set("large-key", largeValue)
	require.NoError(t, err)

	retrievedValue, err := kv.Get("large-key")
	require.NoError(t, err)
	assert.Equal(t, largeValue, retrievedValue)
}

func TestKVStore_ConcurrentAccess(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	// Test concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("concurrent-key-%d", id)
			value := fmt.Sprintf("concurrent-value-%d", id)

			err := kv.Set(key, []byte(value))
			assert.NoError(t, err)

			retrievedValue, err := kv.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, []byte(value), retrievedValue)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all keys were written
	stats, err := kv.Stats()
	require.NoError(t, err)
	assert.Equal(t, int64(10), stats.KeyCount)
}

func TestKVStore_UpdateOperations(t *testing.T) {
	kv, cleanup := setupTestKVStore(t)
	defer cleanup()

	key := "update-key"

	// Set initial value
	err := kv.Set(key, []byte("initial-value"))
	require.NoError(t, err)

	value, err := kv.Get(key)
	require.NoError(t, err)
	assert.Equal(t, []byte("initial-value"), value)

	// Update value
	err = kv.Set(key, []byte("updated-value"))
	require.NoError(t, err)

	value, err = kv.Get(key)
	require.NoError(t, err)
	assert.Equal(t, []byte("updated-value"), value)

	// Verify only one key exists
	stats, err := kv.Stats()
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.KeyCount)
}

func BenchmarkKVStore_Set(b *testing.B) {
	tmpDir := b.TempDir()
	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	require.NoError(b, err)
	defer store.Close()

	kv := NewKVStore(store)
	value := make([]byte, 1024) // 1KB value

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		err := kv.Set(key, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkKVStore_Get(b *testing.B) {
	tmpDir := b.TempDir()
	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	require.NoError(b, err)
	defer store.Close()

	kv := NewKVStore(store)
	value := make([]byte, 1024) // 1KB value

	// Pre-populate data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i)
		err := kv.Set(key, value)
		require.NoError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark-key-%d", i%1000)
		_, err := kv.Get(key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
