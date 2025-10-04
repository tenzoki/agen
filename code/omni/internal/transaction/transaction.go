package transaction

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"

	"github.com/agen/omni/internal/graph"
	"github.com/agen/omni/internal/kv"
	"github.com/agen/omni/internal/storage"
)

// graphTransaction implements the GraphTx interface
type graphTransaction struct {
	id         string
	config     *TransactionConfig
	store      *storage.BadgerStore
	badgerTx   *badger.Txn
	graphStore graph.GraphStore
	kvStore    kv.KVStore

	state     TransactionState
	startTime time.Time
	endTime   time.Time

	operations []*Operation
	savepoints map[string][]*Operation

	stats *TransactionStats

	mu        sync.RWMutex
	listeners []TransactionListener

	ctx    context.Context
	cancel context.CancelFunc
}

// transactionManager implements the TransactionManager interface
type transactionManager struct {
	store              *storage.BadgerStore
	activeTransactions map[string]*graphTransaction
	defaultConfig      *TransactionConfig
	listeners          []TransactionListener

	mu sync.RWMutex
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(store *storage.BadgerStore) TransactionManager {
	return &transactionManager{
		store:              store,
		activeTransactions: make(map[string]*graphTransaction),
		defaultConfig:      DefaultTransactionConfig(),
		listeners:          make([]TransactionListener, 0),
	}
}

// Begin starts a new transaction with the given configuration
func (tm *transactionManager) Begin(config *TransactionConfig) (GraphTx, error) {
	return tm.BeginWithContext(context.Background(), config)
}

// BeginWithContext starts a new transaction with context and configuration
func (tm *transactionManager) BeginWithContext(ctx context.Context, config *TransactionConfig) (GraphTx, error) {
	if config == nil {
		config = tm.defaultConfig
	}

	// Create context with timeout
	txCtx, cancel := context.WithTimeout(ctx, config.Timeout)

	// Start BadgerDB transaction
	var badgerTx *badger.Txn
	if config.ReadOnly {
		badgerTx = tm.store.GetDB().NewTransaction(false)
	} else {
		badgerTx = tm.store.GetDB().NewTransaction(true)
	}

	// Create transaction ID
	txID := uuid.New().String()

	// Create graph and KV stores for this transaction
	graphStore := graph.NewGraphStore(tm.store)
	kvStore := kv.NewKVStore(tm.store)

	// Create transaction
	tx := &graphTransaction{
		id:         txID,
		config:     config,
		store:      tm.store,
		badgerTx:   badgerTx,
		graphStore: graphStore,
		kvStore:    kvStore,
		state:      TxActive,
		startTime:  time.Now(),
		operations: make([]*Operation, 0),
		savepoints: make(map[string][]*Operation),
		stats: &TransactionStats{
			ID:             txID,
			StartTime:      time.Now(),
			State:          TxActive,
			IsolationLevel: config.IsolationLevel,
		},
		listeners: tm.listeners,
		ctx:       txCtx,
		cancel:    cancel,
	}

	// Register active transaction
	tm.mu.Lock()
	tm.activeTransactions[txID] = tx
	tm.mu.Unlock()

	// Notify listeners
	tx.notifyEvent(TxEventBegin, nil, "Transaction started", nil)

	return tx, nil
}

// Execute runs a function within a transaction
func (tm *transactionManager) Execute(fn func(tx GraphTx) error) error {
	return tm.ExecuteWithConfig(tm.defaultConfig, fn)
}

// ExecuteWithConfig runs a function within a transaction with specific config
func (tm *transactionManager) ExecuteWithConfig(config *TransactionConfig, fn func(tx GraphTx) error) error {
	tx, err := tm.Begin(config)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if tx.GetState() == TxActive {
			_ = tx.Rollback()
		}
		tm.cleanupTransaction(tx.GetID())
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("execution failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("transaction execution failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteReadOnly runs a read-only function within a transaction
func (tm *transactionManager) ExecuteReadOnly(fn func(tx GraphTx) error) error {
	config := &TransactionConfig{
		IsolationLevel: ReadCommitted,
		Timeout:        30 * time.Second,
		ReadOnly:       true,
		MaxRetries:     1,
	}
	return tm.ExecuteWithConfig(config, fn)
}

// GetActiveTransactions returns IDs of all active transactions
func (tm *transactionManager) GetActiveTransactions() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	txIDs := make([]string, 0, len(tm.activeTransactions))
	for txID := range tm.activeTransactions {
		txIDs = append(txIDs, txID)
	}
	return txIDs
}

// GetTransactionStats returns statistics for a specific transaction
func (tm *transactionManager) GetTransactionStats(txID string) (*TransactionStats, error) {
	tm.mu.RLock()
	tx, exists := tm.activeTransactions[txID]
	tm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("transaction %s not found", txID)
	}

	tx.mu.RLock()
	defer tx.mu.RUnlock()

	// Create a copy of stats
	stats := *tx.stats
	stats.Duration = time.Since(tx.startTime)
	stats.OperationCount = len(tx.operations)

	return &stats, nil
}

// AbortTransaction forcibly aborts a transaction
func (tm *transactionManager) AbortTransaction(txID string) error {
	tm.mu.RLock()
	tx, exists := tm.activeTransactions[txID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("transaction %s not found", txID)
	}

	tx.mu.Lock()
	if tx.state != TxActive {
		tx.mu.Unlock()
		return fmt.Errorf("transaction %s is not active (state: %s)", txID, tx.state)
	}

	tx.state = TxAborted
	tx.endTime = time.Now()
	tx.stats.State = TxAborted
	tx.stats.EndTime = tx.endTime
	tx.stats.Duration = tx.endTime.Sub(tx.startTime)
	tx.mu.Unlock()

	// Cancel context and cleanup
	tx.cancel()
	tx.badgerTx.Discard()

	tm.cleanupTransaction(txID)

	// Notify listeners
	tx.notifyEvent(TxEventError, nil, "Transaction aborted", fmt.Errorf("transaction aborted by manager"))

	return nil
}


// SetDefaultConfig sets the default transaction configuration
func (tm *transactionManager) SetDefaultConfig(config *TransactionConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.defaultConfig = config
}

// Close closes the transaction manager
func (tm *transactionManager) Close() error {
	// Get a copy of active transaction IDs to avoid holding lock during abort
	tm.mu.Lock()
	txIDs := make([]string, 0, len(tm.activeTransactions))
	for txID := range tm.activeTransactions {
		txIDs = append(txIDs, txID)
	}
	tm.mu.Unlock()

	// Abort all active transactions without holding the manager lock
	for _, txID := range txIDs {
		_ = tm.AbortTransaction(txID)
	}

	return nil
}

// cleanupTransaction removes a transaction from active tracking
func (tm *transactionManager) cleanupTransaction(txID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.activeTransactions, txID)
}

// Transaction Implementation (GraphTx interface)

// Commit commits the transaction
func (tx *graphTransaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.state != TxActive {
		return fmt.Errorf("transaction is not active (state: %s)", tx.state)
	}

	tx.state = TxCommitting
	tx.stats.State = TxCommitting

	// Notify listeners
	tx.notifyEvent(TxEventCommit, nil, "Committing transaction", nil)

	// Commit BadgerDB transaction
	if err := tx.badgerTx.Commit(); err != nil {
		tx.state = TxAborted
		tx.stats.State = TxAborted
		tx.notifyEvent(TxEventError, nil, "Commit failed", err)
		return fmt.Errorf("failed to commit badger transaction: %w", err)
	}

	// Update state
	tx.state = TxCommitted
	tx.endTime = time.Now()
	tx.stats.State = TxCommitted
	tx.stats.EndTime = tx.endTime
	tx.stats.Duration = tx.endTime.Sub(tx.startTime)

	// Cancel context
	tx.cancel()

	return nil
}

// Rollback rolls back the transaction
func (tx *graphTransaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.state != TxActive && tx.state != TxCommitting {
		return fmt.Errorf("cannot rollback transaction in state: %s", tx.state)
	}

	tx.state = TxRolledBack
	tx.endTime = time.Now()
	tx.stats.State = TxRolledBack
	tx.stats.EndTime = tx.endTime
	tx.stats.Duration = tx.endTime.Sub(tx.startTime)

	// Discard BadgerDB transaction
	tx.badgerTx.Discard()

	// Cancel context
	tx.cancel()

	// Notify listeners
	tx.notifyEvent(TxEventRollback, nil, "Transaction rolled back", nil)

	return nil
}

// GetState returns the current transaction state
func (tx *graphTransaction) GetState() TransactionState {
	tx.mu.RLock()
	defer tx.mu.RUnlock()
	return tx.state
}

// GetID returns the transaction ID
func (tx *graphTransaction) GetID() string {
	return tx.id
}

// GetStats returns transaction statistics
func (tx *graphTransaction) GetStats() *TransactionStats {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	stats := *tx.stats
	stats.Duration = time.Since(tx.startTime)
	stats.OperationCount = len(tx.operations)

	return &stats
}

// notifyEvent notifies all listeners of a transaction event
func (tx *graphTransaction) notifyEvent(eventType TransactionEventType, operation *Operation, message string, err error) {
	event := &TransactionEvent{
		TxID:      tx.id,
		EventType: eventType,
		Operation: operation,
		Timestamp: time.Now(),
		Message:   message,
		Error:     err,
	}

	for _, listener := range tx.listeners {
		if listener.IsEnabled() {
			listener.OnTransactionEvent(event)
		}
	}
}

// addOperation adds an operation to the transaction log
func (tx *graphTransaction) addOperation(opType OperationType, key string, value []byte, oldValue []byte) {
	op := &Operation{
		Type:      opType,
		Key:       key,
		Value:     value,
		OldValue:  oldValue,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	tx.operations = append(tx.operations, op)
	tx.stats.OperationCount++

	if opType == OpKVSet || opType == OpCreateVertex || opType == OpUpdateVertex || opType == OpCreateEdge {
		tx.stats.WriteCount++
		tx.stats.BytesWritten += int64(len(value))
	} else {
		tx.stats.ReadCount++
		tx.stats.BytesRead += int64(len(value))
	}

	tx.notifyEvent(TxEventOperation, op, fmt.Sprintf("Operation: %s", opType), nil)
}

// checkState verifies the transaction is in an active state
func (tx *graphTransaction) checkState() error {
	if tx.state != TxActive {
		return fmt.Errorf("transaction is not active (state: %s)", tx.state)
	}
	return nil
}
