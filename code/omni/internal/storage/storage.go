package storage

import (
	"errors"
	"io"
	"time"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrClosed      = errors.New("storage is closed")
	ErrReadOnly    = errors.New("storage is read-only")
)

type Store interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	SetWithTTL(key, value []byte, ttl time.Duration) error
	Delete(key []byte) error
	Exists(key []byte) (bool, error)

	BatchSet(items map[string][]byte) error
	BatchGet(keys [][]byte) (map[string][]byte, error)
	Scan(prefix []byte, limit int) (map[string][]byte, error)

	NewTransaction(update bool) Transaction
	Update(fn func(Transaction) error) error
	View(fn func(Transaction) error) error

	Backup(w io.Writer, since uint64) error
	Load(r io.Reader, maxPendingWrites int) error

	Close() error
	Size() (int64, error)
	Info() map[string]interface{}
}

type Transaction interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	SetWithTTL(key, value []byte, ttl time.Duration) error
	Delete(key []byte) error
	Exists(key []byte) (bool, error)
	Scan(prefix []byte, limit int) (map[string][]byte, error)

	Commit() error
	Discard()
}

type Operation struct {
	Type  OpType
	Key   []byte
	Value []byte
	TTL   time.Duration
}

type OpType int

const (
	OpSet OpType = iota
	OpSetWithTTL
	OpDelete
)

func (ot OpType) String() string {
	switch ot {
	case OpSet:
		return "set"
	case OpSetWithTTL:
		return "set_with_ttl"
	case OpDelete:
		return "delete"
	default:
		return "unknown"
	}
}

type BatchOptions struct {
	MaxBatchSize   int
	MaxBatchDelay  time.Duration
	MaxConcurrency int
}

type Stats struct {
	KeyCount       int64         `json:"key_count"`
	TotalSize      int64         `json:"total_size"`
	LSMSize        int64         `json:"lsm_size"`
	VLogSize       int64         `json:"vlog_size"`
	ReadCount      int64         `json:"read_count"`
	WriteCount     int64         `json:"write_count"`
	ReadLatency    time.Duration `json:"read_latency"`
	WriteLatency   time.Duration `json:"write_latency"`
	LastCompaction time.Time     `json:"last_compaction"`
	GCRuns         int64         `json:"gc_runs"`
}
