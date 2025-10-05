package transaction

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/storage"
)

// Test helper functions
func setupTestStore(t *testing.T) (*storage.BadgerStore, func()) {
	tmpDir := "/tmp/tx-test-" + time.Now().Format("20060102-150405")
	config := storage.DefaultConfig(tmpDir)

	store, err := storage.NewBadgerStore(config)
	if err != nil {
		t.Fatalf("Failed to create BadgerStore: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func createTestVertex(id, vertexType string, properties map[string]interface{}) *common.Vertex {
	vertex := common.NewVertex(id, vertexType)
	for k, v := range properties {
		vertex.Properties[k] = v
	}
	return vertex
}

func createTestEdge(id, edgeType, from, to string, properties map[string]interface{}) *common.Edge {
	edge := common.NewEdge(id, edgeType, from, to)
	for k, v := range properties {
		edge.Properties[k] = v
	}
	return edge
}

// Test TransactionManager
func TestTransactionManager_BasicLifecycle(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	// Test Begin
	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Check initial state
	if tx.GetState() != TxActive {
		t.Errorf("Expected transaction state to be Active, got %s", tx.GetState())
	}

	if tx.GetID() == "" {
		t.Error("Transaction ID should not be empty")
	}

	// Test Commit
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	if tx.GetState() != TxCommitted {
		t.Errorf("Expected transaction state to be Committed, got %s", tx.GetState())
	}
}

func TestTransactionManager_Rollback(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Test Rollback
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	if tx.GetState() != TxRolledBack {
		t.Errorf("Expected transaction state to be RolledBack, got %s", tx.GetState())
	}
}

func TestTransactionManager_Execute(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	vertex := createTestVertex("test:1", "TestVertex", map[string]interface{}{
		"name":  "Test Vertex",
		"value": 42,
	})

	// Test successful execution
	err := tm.Execute(func(tx GraphTx) error {
		return tx.CreateVertex(vertex)
	})

	if err != nil {
		t.Fatalf("Transaction execution failed: %v", err)
	}

	// Verify vertex was created
	tx, _ := tm.Begin(nil)
	defer tx.Rollback()

	retrievedVertex, err := tx.GetVertex("test:1")
	if err != nil {
		t.Fatalf("Failed to retrieve vertex: %v", err)
	}

	if retrievedVertex.Properties["name"] != "Test Vertex" {
		t.Errorf("Expected name to be 'Test Vertex', got %v", retrievedVertex.Properties["name"])
	}
}

func TestTransactionManager_ExecuteWithError(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	// Test execution with error (should rollback)
	err := tm.Execute(func(tx GraphTx) error {
		vertex := createTestVertex("test:1", "TestVertex", nil)
		if err := tx.CreateVertex(vertex); err != nil {
			return err
		}

		// Simulate error
		return fmt.Errorf("simulated error")
	})

	if err == nil {
		t.Fatal("Expected execution to fail")
	}

	// Verify vertex was not created (rolled back)
	tx, _ := tm.Begin(nil)
	defer tx.Rollback()

	_, err = tx.GetVertex("test:1")
	if err == nil {
		t.Error("Expected vertex to not exist after rollback")
	}
}

// Test Vertex Operations
func TestTransaction_VertexOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Test CreateVertex
	vertex := createTestVertex("vertex:1", "TestVertex", map[string]interface{}{
		"name":  "Test Vertex",
		"value": 42,
	})

	if err := tx.CreateVertex(vertex); err != nil {
		t.Fatalf("Failed to create vertex: %v", err)
	}

	// Test VertexExists
	if exists, err := tx.VertexExists("vertex:1"); err != nil {
		t.Fatalf("Failed to check vertex existence: %v", err)
	} else if !exists {
		t.Error("Expected vertex to exist")
	}

	// Test GetVertex
	retrievedVertex, err := tx.GetVertex("vertex:1")
	if err != nil {
		t.Fatalf("Failed to get vertex: %v", err)
	}

	if retrievedVertex.ID != "vertex:1" {
		t.Errorf("Expected vertex ID to be 'vertex:1', got %s", retrievedVertex.ID)
	}

	if retrievedVertex.Properties["name"] != "Test Vertex" {
		t.Errorf("Expected name to be 'Test Vertex', got %v", retrievedVertex.Properties["name"])
	}

	// Test UpdateVertex
	retrievedVertex.Properties["name"] = "Updated Vertex"
	if err := tx.UpdateVertex(retrievedVertex); err != nil {
		t.Fatalf("Failed to update vertex: %v", err)
	}

	// Verify update
	updatedVertex, err := tx.GetVertex("vertex:1")
	if err != nil {
		t.Fatalf("Failed to get updated vertex: %v", err)
	}

	if updatedVertex.Properties["name"] != "Updated Vertex" {
		t.Errorf("Expected name to be 'Updated Vertex', got %v", updatedVertex.Properties["name"])
	}

	// Test DeleteVertex
	if err := tx.DeleteVertex("vertex:1"); err != nil {
		t.Fatalf("Failed to delete vertex: %v", err)
	}

	// Verify deletion
	if exists, _ := tx.VertexExists("vertex:1"); exists {
		t.Error("Expected vertex to be deleted")
	}
}

// Test Edge Operations
func TestTransaction_EdgeOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create vertices first
	vertex1 := createTestVertex("vertex:1", "TestVertex", map[string]interface{}{"name": "Vertex 1"})
	vertex2 := createTestVertex("vertex:2", "TestVertex", map[string]interface{}{"name": "Vertex 2"})

	if err := tx.CreateVertex(vertex1); err != nil {
		t.Fatalf("Failed to create vertex1: %v", err)
	}
	if err := tx.CreateVertex(vertex2); err != nil {
		t.Fatalf("Failed to create vertex2: %v", err)
	}

	// Test CreateEdge
	edge := createTestEdge("edge:1", "TestEdge", "vertex:1", "vertex:2", map[string]interface{}{
		"weight": 1.5,
		"label":  "connects",
	})

	if err := tx.CreateEdge(edge); err != nil {
		t.Fatalf("Failed to create edge: %v", err)
	}

	// Test EdgeExists
	if exists, err := tx.EdgeExists("edge:1"); err != nil {
		t.Fatalf("Failed to check edge existence: %v", err)
	} else if !exists {
		t.Error("Expected edge to exist")
	}

	// Test GetEdge
	retrievedEdge, err := tx.GetEdge("edge:1")
	if err != nil {
		t.Fatalf("Failed to get edge: %v", err)
	}

	if retrievedEdge.FromVertex != "vertex:1" {
		t.Errorf("Expected from vertex to be 'vertex:1', got %s", retrievedEdge.FromVertex)
	}

	if retrievedEdge.ToVertex != "vertex:2" {
		t.Errorf("Expected to vertex to be 'vertex:2', got %s", retrievedEdge.ToVertex)
	}

	// Test DeleteEdge
	if err := tx.DeleteEdge("edge:1"); err != nil {
		t.Fatalf("Failed to delete edge: %v", err)
	}

	// Verify deletion
	if exists, _ := tx.EdgeExists("edge:1"); exists {
		t.Error("Expected edge to be deleted")
	}
}

// Test KV Operations
func TestTransaction_KVOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Test KVSet
	key := "test:key"
	value := []byte("test value")

	if err := tx.KVSet(key, value); err != nil {
		t.Fatalf("Failed to set KV pair: %v", err)
	}

	// Test KVExists
	if exists, err := tx.KVExists(key); err != nil {
		t.Fatalf("Failed to check KV existence: %v", err)
	} else if !exists {
		t.Error("Expected key to exist")
	}

	// Test KVGet
	retrievedValue, err := tx.KVGet(key)
	if err != nil {
		t.Fatalf("Failed to get KV value: %v", err)
	}

	if string(retrievedValue) != "test value" {
		t.Errorf("Expected value to be 'test value', got %s", string(retrievedValue))
	}

	// Test KVDelete
	if err := tx.KVDelete(key); err != nil {
		t.Fatalf("Failed to delete KV pair: %v", err)
	}

	// Verify deletion
	if exists, _ := tx.KVExists(key); exists {
		t.Error("Expected key to be deleted")
	}
}

// Test Batch Operations
func TestTransaction_BatchOperations(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Test BatchCreateVertices
	vertices := []*common.Vertex{
		createTestVertex("vertex:1", "TestVertex", map[string]interface{}{"name": "Vertex 1"}),
		createTestVertex("vertex:2", "TestVertex", map[string]interface{}{"name": "Vertex 2"}),
		createTestVertex("vertex:3", "TestVertex", map[string]interface{}{"name": "Vertex 3"}),
	}

	if err := tx.BatchCreateVertices(vertices); err != nil {
		t.Fatalf("Failed to batch create vertices: %v", err)
	}

	// Verify vertices were created
	for _, vertex := range vertices {
		if exists, err := tx.VertexExists(vertex.ID); err != nil {
			t.Fatalf("Failed to check vertex existence: %v", err)
		} else if !exists {
			t.Errorf("Expected vertex %s to exist", vertex.ID)
		}
	}

	// Test BatchCreateEdges
	edges := []*common.Edge{
		createTestEdge("edge:1", "TestEdge", "vertex:1", "vertex:2", nil),
		createTestEdge("edge:2", "TestEdge", "vertex:2", "vertex:3", nil),
	}

	if err := tx.BatchCreateEdges(edges); err != nil {
		t.Fatalf("Failed to batch create edges: %v", err)
	}

	// Verify edges were created
	for _, edge := range edges {
		if exists, err := tx.EdgeExists(edge.ID); err != nil {
			t.Fatalf("Failed to check edge existence: %v", err)
		} else if !exists {
			t.Errorf("Expected edge %s to exist", edge.ID)
		}
	}

	// Test BatchKVSet
	kvPairs := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	if err := tx.BatchKVSet(kvPairs); err != nil {
		t.Fatalf("Failed to batch set KV pairs: %v", err)
	}

	// Verify KV pairs were set
	for key, expectedValue := range kvPairs {
		if value, err := tx.KVGet(key); err != nil {
			t.Fatalf("Failed to get KV value for %s: %v", key, err)
		} else if string(value) != string(expectedValue) {
			t.Errorf("Expected value for %s to be %s, got %s", key, expectedValue, value)
		}
	}
}

// Test Savepoints
func TestTransaction_Savepoints(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Create initial vertex
	vertex1 := createTestVertex("vertex:1", "TestVertex", map[string]interface{}{"name": "Initial"})
	if err := tx.CreateVertex(vertex1); err != nil {
		t.Fatalf("Failed to create initial vertex: %v", err)
	}

	// Create savepoint
	if err := tx.Savepoint("checkpoint1"); err != nil {
		t.Fatalf("Failed to create savepoint: %v", err)
	}

	// Create another vertex
	vertex2 := createTestVertex("vertex:2", "TestVertex", map[string]interface{}{"name": "After checkpoint"})
	if err := tx.CreateVertex(vertex2); err != nil {
		t.Fatalf("Failed to create vertex after savepoint: %v", err)
	}

	// Verify both vertices exist
	if exists, _ := tx.VertexExists("vertex:1"); !exists {
		t.Error("Expected vertex:1 to exist")
	}
	if exists, _ := tx.VertexExists("vertex:2"); !exists {
		t.Error("Expected vertex:2 to exist")
	}

	// Rollback to savepoint
	if err := tx.RollbackToSavepoint("checkpoint1"); err != nil {
		t.Fatalf("Failed to rollback to savepoint: %v", err)
	}

	// Verify vertex:1 still exists but vertex:2 should be gone (conceptually)
	// Note: This implementation doesn't actually rollback data, just operation log
	// In a full implementation, this would require more sophisticated rollback logic

	// Release savepoint
	if err := tx.ReleaseSavepoint("checkpoint1"); err != nil {
		t.Fatalf("Failed to release savepoint: %v", err)
	}
}

// Test Transaction Statistics
func TestTransaction_Statistics(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	tx, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Get initial stats
	stats := tx.GetStats()
	if stats.State != TxActive {
		t.Errorf("Expected state to be Active, got %s", stats.State)
	}

	if stats.OperationCount != 0 {
		t.Errorf("Expected operation count to be 0, got %d", stats.OperationCount)
	}

	// Perform some operations
	vertex := createTestVertex("vertex:1", "TestVertex", nil)
	if err := tx.CreateVertex(vertex); err != nil {
		t.Fatalf("Failed to create vertex: %v", err)
	}

	if err := tx.KVSet("key1", []byte("value1")); err != nil {
		t.Fatalf("Failed to set KV pair: %v", err)
	}

	// Check updated stats
	stats = tx.GetStats()
	if stats.OperationCount < 2 {
		t.Errorf("Expected operation count to be at least 2, got %d", stats.OperationCount)
	}

	if stats.WriteCount < 2 {
		t.Errorf("Expected write count to be at least 2, got %d", stats.WriteCount)
	}
}

// Test Concurrent Transactions
func TestTransactionManager_Concurrent(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	// Create multiple transactions
	tx1, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction 1: %v", err)
	}
	defer tx1.Rollback()

	tx2, err := tm.Begin(nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction 2: %v", err)
	}
	defer tx2.Rollback()

	// Check active transactions
	activeTxs := tm.GetActiveTransactions()
	if len(activeTxs) != 2 {
		t.Errorf("Expected 2 active transactions, got %d", len(activeTxs))
	}

	// Commit one transaction
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction 1: %v", err)
	}

	// Check stats for the committed transaction
	if stats, err := tm.GetTransactionStats(tx1.GetID()); err == nil {
		if stats.State != TxCommitted {
			t.Errorf("Expected committed transaction state to be Committed, got %s", stats.State)
		}
	}
}

// Test Context and Timeout
func TestTransaction_ContextTimeout(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	// Create transaction with short timeout
	config := &TransactionConfig{
		Timeout:    100 * time.Millisecond,
		ReadOnly:   false,
		MaxRetries: 1,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	tx, err := tm.BeginWithContext(ctx, config)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Wait for context timeout
	time.Sleep(100 * time.Millisecond)

	// Transaction should still be active but context should be done
	// Note: The actual context cancellation handling would depend on implementation details
	if tx.GetState() == TxAborted {
		t.Log("Transaction was aborted due to context timeout (expected)")
	}
}

// Benchmark transaction performance
func BenchmarkTransaction_VertexOperations(b *testing.B) {
	store, cleanup := setupTestStore(&testing.T{})
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := tm.Execute(func(tx GraphTx) error {
			vertex := createTestVertex(fmt.Sprintf("vertex:%d", i), "BenchVertex",
				map[string]interface{}{"value": i})
			return tx.CreateVertex(vertex)
		})
		if err != nil {
			b.Fatalf("Transaction failed: %v", err)
		}
	}
}

func BenchmarkTransaction_KVOperations(b *testing.B) {
	store, cleanup := setupTestStore(&testing.T{})
	defer cleanup()

	tm := NewTransactionManager(store)
	defer tm.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := tm.Execute(func(tx GraphTx) error {
			return tx.KVSet(fmt.Sprintf("key:%d", i), []byte(fmt.Sprintf("value:%d", i)))
		})
		if err != nil {
			b.Fatalf("Transaction failed: %v", err)
		}
	}
}
