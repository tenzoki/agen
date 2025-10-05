package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

type Config struct {
	Dir                  string
	ValueDir             string
	SyncWrites           bool
	NumVersionsToKeep    int
	ReadOnly             bool
	ValueLogFileSize     int64
	BlockCacheSize       int64
	IndexCacheSize       int64
	NumGoroutines        int
	NumMemtables         int
	NumLevelZeroTables   int
	Compression          options.CompressionType
	ZSTDCompressionLevel int
}

func DefaultConfig(dir string) *Config {
	return &Config{
		Dir:                  dir,
		ValueDir:             "",
		SyncWrites:           false,
		NumVersionsToKeep:    1,
		ReadOnly:             false,
		ValueLogFileSize:     1 << 28,   // 256MB
		BlockCacheSize:       256 << 20, // 256MB
		IndexCacheSize:       0,
		NumGoroutines:        8,
		NumMemtables:         5,
		NumLevelZeroTables:   5,
		Compression:          options.Snappy,
		ZSTDCompressionLevel: 1,
	}
}

type BadgerStore struct {
	db     *badger.DB
	config *Config
	mu     sync.RWMutex
	closed bool
}

func NewBadgerStore(config *Config) (*BadgerStore, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if err := os.MkdirAll(config.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	opts := badger.DefaultOptions(config.Dir)

	if config.ValueDir != "" {
		opts.ValueDir = config.ValueDir
	}

	opts.SyncWrites = config.SyncWrites
	opts.NumVersionsToKeep = config.NumVersionsToKeep
	opts.ReadOnly = config.ReadOnly
	opts.ValueLogFileSize = config.ValueLogFileSize
	opts.BlockCacheSize = config.BlockCacheSize
	opts.IndexCacheSize = config.IndexCacheSize
	opts.NumGoroutines = config.NumGoroutines
	opts.NumMemtables = config.NumMemtables
	opts.NumLevelZeroTables = config.NumLevelZeroTables
	opts.Compression = config.Compression
	opts.ZSTDCompressionLevel = config.ZSTDCompressionLevel

	opts.Logger = &badgerLogger{}

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	store := &BadgerStore{
		db:     db,
		config: config,
	}

	return store, nil
}

func (bs *BadgerStore) Close() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.closed {
		return nil
	}

	bs.closed = true
	return bs.db.Close()
}

func (bs *BadgerStore) isClosed() bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.closed
}

// GetDB returns the underlying BadgerDB instance
func (bs *BadgerStore) GetDB() *badger.DB {
	return bs.db
}

func (bs *BadgerStore) Get(key []byte) ([]byte, error) {
	if bs.isClosed() {
		return nil, fmt.Errorf("store is closed")
	}

	var value []byte
	err := bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		return err
	})

	if err == badger.ErrKeyNotFound {
		return nil, ErrKeyNotFound
	}

	return value, err
}

func (bs *BadgerStore) Set(key, value []byte) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (bs *BadgerStore) SetWithTTL(key, value []byte, ttl time.Duration) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry(key, value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

func (bs *BadgerStore) Delete(key []byte) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (bs *BadgerStore) Exists(key []byte) (bool, error) {
	if bs.isClosed() {
		return false, fmt.Errorf("store is closed")
	}

	var exists bool
	err := bs.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			exists = false
			return nil
		}
		if err != nil {
			return err
		}
		exists = true
		return nil
	})

	return exists, err
}

func (bs *BadgerStore) BatchSet(items map[string][]byte) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		for k, v := range items {
			if err := txn.Set([]byte(k), v); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bs *BadgerStore) BatchGet(keys [][]byte) (map[string][]byte, error) {
	if bs.isClosed() {
		return nil, fmt.Errorf("store is closed")
	}

	result := make(map[string][]byte)

	err := bs.db.View(func(txn *badger.Txn) error {
		for _, key := range keys {
			item, err := txn.Get(key)
			if err == badger.ErrKeyNotFound {
				continue
			}
			if err != nil {
				return err
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			result[string(key)] = value
		}
		return nil
	})

	return result, err
}

func (bs *BadgerStore) Scan(prefix []byte, limit int) (map[string][]byte, error) {
	if bs.isClosed() {
		return nil, fmt.Errorf("store is closed")
	}

	result := make(map[string][]byte)
	count := 0

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix) && (limit <= 0 || count < limit); it.Next() {
			item := it.Item()
			key := item.Key()

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}

			result[string(key)] = value
			count++
		}
		return nil
	})

	return result, err
}

func (bs *BadgerStore) NewTransaction(update bool) Transaction {
	txn := bs.db.NewTransaction(update)
	return &BadgerTransaction{txn: txn}
}

func (bs *BadgerStore) Update(fn func(Transaction) error) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Update(func(txn *badger.Txn) error {
		tx := &BadgerTransaction{txn: txn}
		return fn(tx)
	})
}

func (bs *BadgerStore) View(fn func(Transaction) error) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.View(func(txn *badger.Txn) error {
		tx := &BadgerTransaction{txn: txn}
		return fn(tx)
	})
}

func (bs *BadgerStore) Backup(w io.Writer, since uint64) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	_, err := bs.db.Backup(w, since)
	return err
}

func (bs *BadgerStore) Load(r io.Reader, maxPendingWrites int) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Load(r, maxPendingWrites)
}

func (bs *BadgerStore) RunValueLogGC(discardRatio float64) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	for {
		err := bs.db.RunValueLogGC(discardRatio)
		if err != nil {
			if err == badger.ErrNoRewrite {
				break
			}
			return err
		}
	}
	return nil
}

func (bs *BadgerStore) Flatten(workers int) error {
	if bs.isClosed() {
		return fmt.Errorf("store is closed")
	}

	return bs.db.Flatten(workers)
}

func (bs *BadgerStore) Size() (int64, error) {
	if bs.isClosed() {
		return 0, fmt.Errorf("store is closed")
	}

	lsm, vlog := bs.db.Size()
	return lsm + vlog, nil
}

func (bs *BadgerStore) Info() map[string]interface{} {
	if bs.isClosed() {
		return map[string]interface{}{"status": "closed"}
	}

	lsm, vlog := bs.db.Size()

	return map[string]interface{}{
		"status":      "open",
		"dir":         bs.config.Dir,
		"lsm_size":    lsm,
		"vlog_size":   vlog,
		"total_size":  lsm + vlog,
		"sync_writes": bs.config.SyncWrites,
		"read_only":   bs.config.ReadOnly,
		"compression": fmt.Sprintf("%v", bs.config.Compression),
	}
}

func (bs *BadgerStore) StartGarbageCollector(ctx context.Context, interval time.Duration, discardRatio float64) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := bs.RunValueLogGC(discardRatio); err != nil {
				continue
			}
		}
	}
}

type BadgerTransaction struct {
	txn *badger.Txn
}

// NewBadgerTransaction wraps a badger.Txn to implement the Transaction interface
func NewBadgerTransaction(txn *badger.Txn) *BadgerTransaction {
	return &BadgerTransaction{txn: txn}
}

func (bt *BadgerTransaction) Get(key []byte) ([]byte, error) {
	item, err := bt.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	return item.ValueCopy(nil)
}

func (bt *BadgerTransaction) Set(key, value []byte) error {
	return bt.txn.Set(key, value)
}

func (bt *BadgerTransaction) SetWithTTL(key, value []byte, ttl time.Duration) error {
	entry := badger.NewEntry(key, value).WithTTL(ttl)
	return bt.txn.SetEntry(entry)
}

func (bt *BadgerTransaction) Delete(key []byte) error {
	return bt.txn.Delete(key)
}

func (bt *BadgerTransaction) Exists(key []byte) (bool, error) {
	_, err := bt.txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (bt *BadgerTransaction) Scan(prefix []byte, limit int) (map[string][]byte, error) {
	result := make(map[string][]byte)
	count := 0

	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	it := bt.txn.NewIterator(opts)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix) && (limit <= 0 || count < limit); it.Next() {
		item := it.Item()
		key := item.Key()

		value, err := item.ValueCopy(nil)
		if err != nil {
			return nil, err
		}

		result[string(key)] = value
		count++
	}

	return result, nil
}

func (bt *BadgerTransaction) Commit() error {
	return bt.txn.Commit()
}

func (bt *BadgerTransaction) Discard() {
	bt.txn.Discard()
}

type badgerLogger struct{}

func (bl *badgerLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("BADGER ERROR: "+format+"\n", args...)
}

func (bl *badgerLogger) Warningf(format string, args ...interface{}) {
	fmt.Printf("BADGER WARNING: "+format+"\n", args...)
}

func (bl *badgerLogger) Infof(format string, args ...interface{}) {
}

func (bl *badgerLogger) Debugf(format string, args ...interface{}) {
}
