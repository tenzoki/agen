package transaction

import (
	"context"
	"time"

	"github.com/agen/omni/internal/common"
)

// IsolationLevel defines the transaction isolation level
type IsolationLevel int

const (
	// ReadUncommitted allows dirty reads
	ReadUncommitted IsolationLevel = iota
	// ReadCommitted prevents dirty reads
	ReadCommitted
	// RepeatableRead prevents dirty and non-repeatable reads
	RepeatableRead
	// Serializable prevents dirty, non-repeatable, and phantom reads
	Serializable
)

func (il IsolationLevel) String() string {
	switch il {
	case ReadUncommitted:
		return "ReadUncommitted"
	case ReadCommitted:
		return "ReadCommitted"
	case RepeatableRead:
		return "RepeatableRead"
	case Serializable:
		return "Serializable"
	default:
		return "Unknown"
	}
}

// TransactionState represents the current state of a transaction
type TransactionState int

const (
	// TxActive - transaction is active and can perform operations
	TxActive TransactionState = iota
	// TxCommitting - transaction is in the process of committing
	TxCommitting
	// TxCommitted - transaction has been successfully committed
	TxCommitted
	// TxRolledBack - transaction has been rolled back
	TxRolledBack
	// TxAborted - transaction was aborted due to error
	TxAborted
)

func (ts TransactionState) String() string {
	switch ts {
	case TxActive:
		return "Active"
	case TxCommitting:
		return "Committing"
	case TxCommitted:
		return "Committed"
	case TxRolledBack:
		return "RolledBack"
	case TxAborted:
		return "Aborted"
	default:
		return "Unknown"
	}
}

// OperationType represents the type of operation in a transaction
type OperationType int

const (
	OpCreateVertex OperationType = iota
	OpUpdateVertex
	OpDeleteVertex
	OpCreateEdge
	OpDeleteEdge
	OpKVSet
	OpKVDelete
)

func (ot OperationType) String() string {
	switch ot {
	case OpCreateVertex:
		return "CreateVertex"
	case OpUpdateVertex:
		return "UpdateVertex"
	case OpDeleteVertex:
		return "DeleteVertex"
	case OpCreateEdge:
		return "CreateEdge"
	case OpDeleteEdge:
		return "DeleteEdge"
	case OpKVSet:
		return "KVSet"
	case OpKVDelete:
		return "KVDelete"
	default:
		return "Unknown"
	}
}

// Operation represents a single operation within a transaction
type Operation struct {
	Type      OperationType
	Key       string
	Value     []byte
	OldValue  []byte // For rollback purposes
	Timestamp time.Time
	Metadata  map[string]interface{}
}

// TransactionConfig holds configuration for transactions
type TransactionConfig struct {
	IsolationLevel IsolationLevel
	Timeout        time.Duration
	ReadOnly       bool
	MaxRetries     int
}

// DefaultTransactionConfig returns a default transaction configuration
func DefaultTransactionConfig() *TransactionConfig {
	return &TransactionConfig{
		IsolationLevel: ReadCommitted,
		Timeout:        30 * time.Second,
		ReadOnly:       false,
		MaxRetries:     3,
	}
}

// TransactionStats holds statistics about a transaction
type TransactionStats struct {
	ID             string
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	State          TransactionState
	IsolationLevel IsolationLevel
	OperationCount int
	ReadCount      int
	WriteCount     int
	ConflictCount  int
	RetryCount     int
	BytesRead      int64
	BytesWritten   int64
}

// GraphTx defines the interface for graph database transactions
type GraphTx interface {
	// Transaction control
	Commit() error
	Rollback() error
	GetState() TransactionState
	GetID() string
	GetStats() *TransactionStats

	// Vertex operations (transactional)
	CreateVertex(vertex *common.Vertex) error
	GetVertex(id string) (*common.Vertex, error)
	UpdateVertex(vertex *common.Vertex) error
	DeleteVertex(id string) error
	VertexExists(id string) (bool, error)

	// Edge operations (transactional)
	CreateEdge(edge *common.Edge) error
	GetEdge(edgeID string) (*common.Edge, error)
	DeleteEdge(edgeID string) error
	EdgeExists(edgeID string) (bool, error)

	// Query operations (transactional)
	GetVerticesByType(vertexType string, limit int) ([]*common.Vertex, error)
	GetEdgesByType(edgeType string, limit int) ([]*common.Edge, error)
	GetAllVertices(limit int) ([]*common.Vertex, error)
	GetAllEdges(limit int) ([]*common.Edge, error)

	// Traversal operations (transactional)
	GetOutgoingEdges(vertexID string) ([]*common.Edge, error)
	GetIncomingEdges(vertexID string) ([]*common.Edge, error)
	GetNeighbors(vertexID string, direction common.Direction) ([]*common.Vertex, error)

	// KV operations (transactional)
	KVGet(key string) ([]byte, error)
	KVSet(key string, value []byte) error
	KVDelete(key string) error
	KVExists(key string) (bool, error)

	// Batch operations (transactional)
	BatchCreateVertices(vertices []*common.Vertex) error
	BatchCreateEdges(edges []*common.Edge) error
	BatchKVSet(kvPairs map[string][]byte) error

	// Transaction-specific operations
	Savepoint(name string) error
	RollbackToSavepoint(name string) error
	ReleaseSavepoint(name string) error
}

// TransactionManager defines the interface for managing transactions
type TransactionManager interface {
	// Transaction lifecycle
	Begin(config *TransactionConfig) (GraphTx, error)
	BeginWithContext(ctx context.Context, config *TransactionConfig) (GraphTx, error)

	// Transaction execution helpers
	Execute(fn func(tx GraphTx) error) error
	ExecuteWithConfig(config *TransactionConfig, fn func(tx GraphTx) error) error
	ExecuteReadOnly(fn func(tx GraphTx) error) error

	// Transaction management
	GetActiveTransactions() []string
	GetTransactionStats(txID string) (*TransactionStats, error)
	AbortTransaction(txID string) error

	// Configuration and cleanup
	SetDefaultConfig(config *TransactionConfig)
	Close() error
}

// ConflictResolver defines how to handle transaction conflicts
type ConflictResolver interface {
	// ResolveConflict attempts to resolve a conflict between transactions
	ResolveConflict(conflictingTx GraphTx, operation *Operation) error

	// CanResolve returns true if the resolver can handle this type of conflict
	CanResolve(operation *Operation) bool

	// GetResolutionStrategy returns the strategy used by this resolver
	GetResolutionStrategy() string
}

// SavepointManager manages savepoints within a transaction
type SavepointManager interface {
	CreateSavepoint(name string) error
	RollbackToSavepoint(name string) error
	ReleaseSavepoint(name string) error
	GetSavepoints() []string
}

// LockManager manages locks for transaction isolation
type LockManager interface {
	AcquireReadLock(key string, txID string) error
	AcquireWriteLock(key string, txID string) error
	ReleaseLock(key string, txID string) error
	ReleaseAllLocks(txID string) error
	GetLockInfo(key string) (*LockInfo, error)
	DetectDeadlock() ([]string, error)
}

// LockInfo contains information about a lock
type LockInfo struct {
	Key        string
	LockType   LockType
	TxID       string
	AcquiredAt time.Time
	WaitingTxs []string
}

// LockType defines the type of lock
type LockType int

const (
	ReadLock LockType = iota
	WriteLock
)

func (lt LockType) String() string {
	switch lt {
	case ReadLock:
		return "ReadLock"
	case WriteLock:
		return "WriteLock"
	default:
		return "Unknown"
	}
}

// ValidationRule defines a rule for transaction validation
type ValidationRule interface {
	Validate(tx GraphTx, operation *Operation) error
	GetRuleName() string
	IsEnabled() bool
}

// ConsistencyChecker defines an interface for checking data consistency
type ConsistencyChecker interface {
	CheckConsistency(tx GraphTx) error
	CheckVertexConsistency(tx GraphTx, vertex *common.Vertex) error
	CheckEdgeConsistency(tx GraphTx, edge *common.Edge) error
	CheckReferentialIntegrity(tx GraphTx) error
	GetInconsistencies() []string
}

// TransactionEvent represents an event that occurs during transaction execution
type TransactionEvent struct {
	TxID      string
	EventType TransactionEventType
	Operation *Operation
	Timestamp time.Time
	Message   string
	Error     error
}

// TransactionEventType defines types of transaction events
type TransactionEventType int

const (
	TxEventBegin TransactionEventType = iota
	TxEventOperation
	TxEventCommit
	TxEventRollback
	TxEventConflict
	TxEventError
	TxEventSavepoint
)

func (tet TransactionEventType) String() string {
	switch tet {
	case TxEventBegin:
		return "Begin"
	case TxEventOperation:
		return "Operation"
	case TxEventCommit:
		return "Commit"
	case TxEventRollback:
		return "Rollback"
	case TxEventConflict:
		return "Conflict"
	case TxEventError:
		return "Error"
	case TxEventSavepoint:
		return "Savepoint"
	default:
		return "Unknown"
	}
}

// TransactionListener defines an interface for listening to transaction events
type TransactionListener interface {
	OnTransactionEvent(event *TransactionEvent)
	GetListenerName() string
	IsEnabled() bool
}
