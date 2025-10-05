package transaction

import (
	"fmt"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/graph"
	"github.com/tenzoki/agen/omni/internal/storage"
)

// Vertex Operations (GraphTx interface implementation)

// CreateVertex creates a new vertex within the transaction
func (tx *graphTransaction) CreateVertex(vertex *common.Vertex) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Validate vertex
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex validation failed: %w", err)
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Check if vertex already exists
	exists, err := tx.graphStore.VertexExistsInTx(storageTx, vertex.ID)
	if err != nil {
		return fmt.Errorf("failed to check vertex existence: %w", err)
	}
	if exists {
		return fmt.Errorf("vertex %s already exists", vertex.ID)
	}

	// Create vertex using graph store with transaction
	if err := tx.graphStore.AddVertexInTx(storageTx, vertex); err != nil {
		return fmt.Errorf("failed to create vertex: %w", err)
	}

	// Log operation
	vertexData, _ := vertex.MarshalBinary()
	tx.addOperation(OpCreateVertex, vertex.ID, vertexData, nil)

	return nil
}

// GetVertex retrieves a vertex within the transaction
func (tx *graphTransaction) GetVertex(id string) (*common.Vertex, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	vertex, err := tx.graphStore.GetVertexInTx(storageTx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get vertex: %w", err)
	}

	// Log read operation
	vertexData, _ := vertex.MarshalBinary()
	tx.addOperation(OpCreateVertex, id, vertexData, nil) // Using CreateVertex as read op type

	return vertex, nil
}

// UpdateVertex updates a vertex within the transaction
func (tx *graphTransaction) UpdateVertex(vertex *common.Vertex) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Validate vertex
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex validation failed: %w", err)
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Get old vertex for rollback
	oldVertex, err := tx.graphStore.GetVertexInTx(storageTx, vertex.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing vertex: %w", err)
	}

	// Update vertex using graph store with transaction
	if err := tx.graphStore.UpdateVertexInTx(storageTx, vertex); err != nil {
		return fmt.Errorf("failed to update vertex: %w", err)
	}

	// Log operation
	vertexData, _ := vertex.MarshalBinary()
	oldVertexData, _ := oldVertex.MarshalBinary()
	tx.addOperation(OpUpdateVertex, vertex.ID, vertexData, oldVertexData)

	return nil
}

// DeleteVertex deletes a vertex within the transaction
func (tx *graphTransaction) DeleteVertex(id string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Get vertex for rollback
	vertex, err := tx.graphStore.GetVertexInTx(storageTx, id)
	if err != nil {
		return fmt.Errorf("failed to get vertex for deletion: %w", err)
	}

	// Delete vertex using graph store with transaction
	if err := tx.graphStore.DeleteVertexInTx(storageTx, id); err != nil {
		return fmt.Errorf("failed to delete vertex: %w", err)
	}

	// Log operation
	vertexData, _ := vertex.MarshalBinary()
	tx.addOperation(OpDeleteVertex, id, nil, vertexData)

	return nil
}

// VertexExists checks if a vertex exists within the transaction
func (tx *graphTransaction) VertexExists(id string) (bool, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return false, err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	return tx.graphStore.VertexExistsInTx(storageTx, id)
}

// Edge Operations

// CreateEdge creates a new edge within the transaction
func (tx *graphTransaction) CreateEdge(edge *common.Edge) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Validate edge
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("edge validation failed: %w", err)
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Check if edge already exists
	exists, err := tx.graphStore.EdgeExistsInTx(storageTx, edge.ID)
	if err != nil {
		return fmt.Errorf("failed to check edge existence: %w", err)
	}
	if exists {
		return fmt.Errorf("edge %s already exists", edge.ID)
	}

	// Verify that source and target vertices exist
	if exists, err := tx.graphStore.VertexExistsInTx(storageTx, edge.FromVertex); err != nil || !exists {
		return fmt.Errorf("source vertex %s does not exist", edge.FromVertex)
	}
	if exists, err := tx.graphStore.VertexExistsInTx(storageTx, edge.ToVertex); err != nil || !exists {
		return fmt.Errorf("target vertex %s does not exist", edge.ToVertex)
	}

	// Create edge using graph store with transaction
	if err := tx.graphStore.AddEdgeInTx(storageTx, edge); err != nil {
		return fmt.Errorf("failed to create edge: %w", err)
	}

	// Log operation
	edgeData, _ := edge.MarshalBinary()
	tx.addOperation(OpCreateEdge, edge.ID, edgeData, nil)

	return nil
}

// GetEdge retrieves an edge within the transaction
func (tx *graphTransaction) GetEdge(edgeID string) (*common.Edge, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	edge, err := tx.graphStore.GetEdgeInTx(storageTx, edgeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge: %w", err)
	}

	return edge, nil
}

// DeleteEdge deletes an edge within the transaction
func (tx *graphTransaction) DeleteEdge(edgeID string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Get edge for rollback
	edge, err := tx.graphStore.GetEdgeInTx(storageTx, edgeID)
	if err != nil {
		return fmt.Errorf("failed to get edge for deletion: %w", err)
	}

	// Delete edge using graph store with transaction
	if err := tx.graphStore.DeleteEdgeInTx(storageTx, edgeID); err != nil {
		return fmt.Errorf("failed to delete edge: %w", err)
	}

	// Log operation
	edgeData, _ := edge.MarshalBinary()
	tx.addOperation(OpDeleteEdge, edgeID, nil, edgeData)

	return nil
}

// EdgeExists checks if an edge exists within the transaction
func (tx *graphTransaction) EdgeExists(edgeID string) (bool, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return false, err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	return tx.graphStore.EdgeExistsInTx(storageTx, edgeID)
}

// Query Operations

// GetVerticesByType retrieves vertices by type within the transaction
func (tx *graphTransaction) GetVerticesByType(vertexType string, limit int) ([]*common.Vertex, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	return tx.graphStore.GetVerticesByType(vertexType, limit)
}

// GetEdgesByType retrieves edges by type within the transaction
func (tx *graphTransaction) GetEdgesByType(edgeType string, limit int) ([]*common.Edge, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	return tx.graphStore.GetEdgesByType(edgeType, limit)
}

// GetAllVertices retrieves all vertices within the transaction
func (tx *graphTransaction) GetAllVertices(limit int) ([]*common.Vertex, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	return tx.graphStore.GetAllVertices(limit)
}

// GetAllEdges retrieves all edges within the transaction
func (tx *graphTransaction) GetAllEdges(limit int) ([]*common.Edge, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	return tx.graphStore.GetAllEdges(limit)
}

// Traversal Operations

// GetOutgoingEdges retrieves outgoing edges for a vertex within the transaction
func (tx *graphTransaction) GetOutgoingEdges(vertexID string) ([]*common.Edge, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	return tx.graphStore.GetOutgoingEdges(vertexID)
}

// GetIncomingEdges retrieves incoming edges for a vertex within the transaction
func (tx *graphTransaction) GetIncomingEdges(vertexID string) ([]*common.Edge, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	return tx.graphStore.GetIncomingEdges(vertexID)
}

// GetNeighbors retrieves neighbors for a vertex within the transaction
func (tx *graphTransaction) GetNeighbors(vertexID string, direction common.Direction) ([]*common.Vertex, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	// Convert common.Direction to graph.TraversalDirection
	var graphDirection graph.TraversalDirection
	switch direction {
	case common.Incoming:
		graphDirection = graph.DirectionIncoming
	case common.Outgoing:
		graphDirection = graph.DirectionOutgoing
	case common.Both:
		graphDirection = graph.DirectionBoth
	default:
		return nil, fmt.Errorf("invalid direction: %v", direction)
	}

	return tx.graphStore.GetNeighbors(vertexID, graphDirection)
}

// KV Operations

// KVGet retrieves a value from KV store within the transaction
func (tx *graphTransaction) KVGet(key string) ([]byte, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return nil, err
	}

	value, err := tx.kvStore.Get(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get KV value: %w", err)
	}

	return value, nil
}

// KVSet sets a value in KV store within the transaction
func (tx *graphTransaction) KVSet(key string, value []byte) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Get old value for rollback
	var oldValue []byte
	if exists, _ := tx.kvStore.Exists(key); exists {
		oldValue, _ = tx.kvStore.Get(key)
	}

	// Set value using KV store
	if err := tx.kvStore.Set(key, value); err != nil {
		return fmt.Errorf("failed to set KV value: %w", err)
	}

	// Log operation
	tx.addOperation(OpKVSet, key, value, oldValue)

	return nil
}

// KVDelete deletes a value from KV store within the transaction
func (tx *graphTransaction) KVDelete(key string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Get value for rollback
	oldValue, err := tx.kvStore.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get value for deletion: %w", err)
	}

	// Delete value using KV store
	if err := tx.kvStore.Delete(key); err != nil {
		return fmt.Errorf("failed to delete KV value: %w", err)
	}

	// Log operation
	tx.addOperation(OpKVDelete, key, nil, oldValue)

	return nil
}

// KVExists checks if a key exists in KV store within the transaction
func (tx *graphTransaction) KVExists(key string) (bool, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if err := tx.checkState(); err != nil {
		return false, err
	}

	return tx.kvStore.Exists(key)
}

// Batch Operations

// BatchCreateVertices creates multiple vertices within the transaction
func (tx *graphTransaction) BatchCreateVertices(vertices []*common.Vertex) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Validate all vertices first
	for _, vertex := range vertices {
		if err := vertex.Validate(); err != nil {
			return fmt.Errorf("vertex validation failed for %s: %w", vertex.ID, err)
		}

		if exists, err := tx.graphStore.VertexExistsInTx(storageTx, vertex.ID); err != nil {
			return fmt.Errorf("failed to check existence for vertex %s: %w", vertex.ID, err)
		} else if exists {
			return fmt.Errorf("vertex %s already exists", vertex.ID)
		}
	}

	// Create all vertices using transactional method
	for _, vertex := range vertices {
		if err := tx.graphStore.AddVertexInTx(storageTx, vertex); err != nil {
			return fmt.Errorf("failed to create vertex %s: %w", vertex.ID, err)
		}
	}

	// Log operations
	for _, vertex := range vertices {
		vertexData, _ := vertex.MarshalBinary()
		tx.addOperation(OpCreateVertex, vertex.ID, vertexData, nil)
	}

	return nil
}

// BatchCreateEdges creates multiple edges within the transaction
func (tx *graphTransaction) BatchCreateEdges(edges []*common.Edge) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Wrap badger transaction for use with graph store
	storageTx := storage.NewBadgerTransaction(tx.badgerTx)

	// Validate all edges first
	for _, edge := range edges {
		if err := edge.Validate(); err != nil {
			return fmt.Errorf("edge validation failed for %s: %w", edge.ID, err)
		}

		if exists, err := tx.graphStore.EdgeExistsInTx(storageTx, edge.ID); err != nil {
			return fmt.Errorf("failed to check existence for edge %s: %w", edge.ID, err)
		} else if exists {
			return fmt.Errorf("edge %s already exists", edge.ID)
		}

		// Verify vertices exist
		if exists, err := tx.graphStore.VertexExistsInTx(storageTx, edge.FromVertex); err != nil || !exists {
			return fmt.Errorf("source vertex %s does not exist for edge %s", edge.FromVertex, edge.ID)
		}
		if exists, err := tx.graphStore.VertexExistsInTx(storageTx, edge.ToVertex); err != nil || !exists {
			return fmt.Errorf("target vertex %s does not exist for edge %s", edge.ToVertex, edge.ID)
		}
	}

	// Create all edges using transactional method
	for _, edge := range edges {
		if err := tx.graphStore.AddEdgeInTx(storageTx, edge); err != nil {
			return fmt.Errorf("failed to create edge %s: %w", edge.ID, err)
		}
	}

	// Log operations
	for _, edge := range edges {
		edgeData, _ := edge.MarshalBinary()
		tx.addOperation(OpCreateEdge, edge.ID, edgeData, nil)
	}

	return nil
}

// BatchKVSet sets multiple key-value pairs within the transaction
func (tx *graphTransaction) BatchKVSet(kvPairs map[string][]byte) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Get old values for rollback
	oldValues := make(map[string][]byte)
	for key := range kvPairs {
		if exists, _ := tx.kvStore.Exists(key); exists {
			if oldValue, err := tx.kvStore.Get(key); err == nil {
				oldValues[key] = oldValue
			}
		}
	}

	// Set all values
	if err := tx.kvStore.BatchSet(kvPairs); err != nil {
		return fmt.Errorf("failed to batch set KV pairs: %w", err)
	}

	// Log operations
	for key, value := range kvPairs {
		oldValue := oldValues[key]
		tx.addOperation(OpKVSet, key, value, oldValue)
	}

	return nil
}

// Savepoint Operations

// Savepoint creates a savepoint with the given name
func (tx *graphTransaction) Savepoint(name string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	// Copy current operations
	savepoint := make([]*Operation, len(tx.operations))
	copy(savepoint, tx.operations)

	tx.savepoints[name] = savepoint

	tx.notifyEvent(TxEventSavepoint, nil, fmt.Sprintf("Savepoint created: %s", name), nil)

	return nil
}

// RollbackToSavepoint rolls back to a specific savepoint
func (tx *graphTransaction) RollbackToSavepoint(name string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	savepoint, exists := tx.savepoints[name]
	if !exists {
		return fmt.Errorf("savepoint %s does not exist", name)
	}

	// Restore operations to savepoint state
	tx.operations = make([]*Operation, len(savepoint))
	copy(tx.operations, savepoint)

	tx.notifyEvent(TxEventSavepoint, nil, fmt.Sprintf("Rolled back to savepoint: %s", name), nil)

	return nil
}

// ReleaseSavepoint releases a savepoint
func (tx *graphTransaction) ReleaseSavepoint(name string) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if err := tx.checkState(); err != nil {
		return err
	}

	if _, exists := tx.savepoints[name]; !exists {
		return fmt.Errorf("savepoint %s does not exist", name)
	}

	delete(tx.savepoints, name)

	tx.notifyEvent(TxEventSavepoint, nil, fmt.Sprintf("Savepoint released: %s", name), nil)

	return nil
}
