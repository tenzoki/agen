package kv

import (
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/storage"
)

type KVStore interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	Exists(key string) (bool, error)

	BatchSet(items map[string][]byte) error
	BatchGet(keys []string) (map[string][]byte, error)

	Scan(prefix string, limit int) (map[string][]byte, error)
	ListKeys(prefix string, limit int) ([]string, error)

	SetWithTTL(key string, value []byte, ttl time.Duration) error

	Close() error
	Stats() (*KVStats, error)
}

type KVStats struct {
	KeyCount     int64     `json:"key_count"`
	TotalSize    int64     `json:"total_size"`
	Namespace    string    `json:"namespace"`
	LastAccess   time.Time `json:"last_access"`
	AvgKeySize   float64   `json:"avg_key_size"`
	AvgValueSize float64   `json:"avg_value_size"`
}

type kvStore struct {
	store      storage.Store
	keyBuilder *common.KeyBuilder
	keyParser  *common.KeyParser
}

func NewKVStore(store storage.Store) KVStore {
	return &kvStore{
		store:      store,
		keyBuilder: common.NewKeyBuilder(),
		keyParser:  common.NewKeyParser(),
	}
}

func (kv *kvStore) Get(key string) ([]byte, error) {
	if err := common.ValidateKey(key); err != nil {
		return nil, err
	}

	storageKey := kv.keyBuilder.KVKey(key)
	data, err := kv.store.Get(storageKey)
	if err == storage.ErrKeyNotFound {
		return nil, ErrKeyNotFound
	}
	return data, err
}

func (kv *kvStore) Set(key string, value []byte) error {
	if err := common.ValidateKey(key); err != nil {
		return err
	}

	storageKey := kv.keyBuilder.KVKey(key)
	return kv.store.Set(storageKey, value)
}

func (kv *kvStore) Delete(key string) error {
	if err := common.ValidateKey(key); err != nil {
		return err
	}

	storageKey := kv.keyBuilder.KVKey(key)
	return kv.store.Delete(storageKey)
}

func (kv *kvStore) Exists(key string) (bool, error) {
	if err := common.ValidateKey(key); err != nil {
		return false, err
	}

	storageKey := kv.keyBuilder.KVKey(key)
	return kv.store.Exists(storageKey)
}

func (kv *kvStore) BatchSet(items map[string][]byte) error {
	kvItems := make(map[string][]byte)

	for key, value := range items {
		if err := common.ValidateKey(key); err != nil {
			return err
		}
		storageKey := string(kv.keyBuilder.KVKey(key))
		kvItems[storageKey] = value
	}

	return kv.store.BatchSet(kvItems)
}

func (kv *kvStore) BatchGet(keys []string) (map[string][]byte, error) {
	storageKeys := make([][]byte, len(keys))
	keyMap := make(map[string]string) // storageKey -> originalKey

	for i, key := range keys {
		if err := common.ValidateKey(key); err != nil {
			return nil, err
		}
		storageKey := kv.keyBuilder.KVKey(key)
		storageKeys[i] = storageKey
		keyMap[string(storageKey)] = key
	}

	result, err := kv.store.BatchGet(storageKeys)
	if err != nil {
		return nil, err
	}

	// Convert back to original keys
	userResult := make(map[string][]byte)
	for storageKey, value := range result {
		if originalKey, exists := keyMap[storageKey]; exists {
			userResult[originalKey] = value
		}
	}

	return userResult, nil
}

func (kv *kvStore) Scan(prefix string, limit int) (map[string][]byte, error) {
	if err := common.ValidateKey(prefix); err != nil {
		return nil, err
	}

	storagePrefix := kv.keyBuilder.KVKey(prefix)
	result, err := kv.store.Scan(storagePrefix, limit)
	if err != nil {
		return nil, err
	}

	// Convert storage keys back to user keys
	userResult := make(map[string][]byte)
	for storageKey, value := range result {
		if userKey, ok := kv.keyParser.ParseKVKey([]byte(storageKey)); ok {
			userResult[userKey] = value
		}
	}

	return userResult, nil
}

func (kv *kvStore) ListKeys(prefix string, limit int) ([]string, error) {
	if err := common.ValidateKey(prefix); err != nil {
		return nil, err
	}

	storagePrefix := kv.keyBuilder.KVKey(prefix)
	result, err := kv.store.Scan(storagePrefix, limit)
	if err != nil {
		return nil, err
	}

	// Convert storage keys back to user keys
	keys := make([]string, 0, len(result))
	for storageKey := range result {
		if userKey, ok := kv.keyParser.ParseKVKey([]byte(storageKey)); ok {
			keys = append(keys, userKey)
		}
	}

	return keys, nil
}

func (kv *kvStore) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	if err := common.ValidateKey(key); err != nil {
		return err
	}

	storageKey := kv.keyBuilder.KVKey(key)
	return kv.store.SetWithTTL(storageKey, value, ttl)
}

func (kv *kvStore) Close() error {
	return kv.store.Close()
}

func (kv *kvStore) Stats() (*KVStats, error) {
	prefix := kv.keyBuilder.KVPrefix()

	var keyCount int64
	var totalSize int64
	var totalKeySize int64
	var totalValueSize int64

	// Scan all KV keys to calculate statistics
	err := kv.store.View(func(tx storage.Transaction) error {
		data, err := tx.Scan(prefix, -1) // Get all KV keys
		if err != nil {
			return err
		}

		for storageKey, value := range data {
			keyCount++

			// Parse back to user key for size calculation
			if userKey, ok := kv.keyParser.ParseKVKey([]byte(storageKey)); ok {
				totalKeySize += int64(len(userKey))
			}
			totalValueSize += int64(len(value))
		}

		totalSize = totalKeySize + totalValueSize
		return nil
	})

	if err != nil {
		return nil, err
	}

	stats := &KVStats{
		KeyCount:   keyCount,
		TotalSize:  totalSize,
		Namespace:  "kv",
		LastAccess: time.Now(),
	}

	if keyCount > 0 {
		stats.AvgKeySize = float64(totalKeySize) / float64(keyCount)
		stats.AvgValueSize = float64(totalValueSize) / float64(keyCount)
	}

	return stats, nil
}
