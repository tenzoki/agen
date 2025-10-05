package storage

import (
	"fmt"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
)

type CRUDManager struct {
	store      Store
	keyBuilder *common.KeyBuilder
	keyParser  *common.KeyParser
}

func NewCRUDManager(store Store) *CRUDManager {
	return &CRUDManager{
		store:      store,
		keyBuilder: common.NewKeyBuilder(),
		keyParser:  common.NewKeyParser(),
	}
}

// StoreVertex stores a vertex, creating its own transaction
func (cm *CRUDManager) StoreVertex(vertex *common.Vertex) error {
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex validation failed: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertex.ID)
	data, err := vertex.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal vertex: %w", err)
	}

	return cm.store.Update(func(tx Transaction) error {
		return cm.storeVertexInTx(tx, vertex, key, data)
	})
}

// StoreVertexInTx stores a vertex using the provided transaction
func (cm *CRUDManager) StoreVertexInTx(tx Transaction, vertex *common.Vertex) error {
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex validation failed: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertex.ID)
	data, err := vertex.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal vertex: %w", err)
	}

	return cm.storeVertexInTx(tx, vertex, key, data)
}

// storeVertexInTx is the internal implementation
func (cm *CRUDManager) storeVertexInTx(tx Transaction, vertex *common.Vertex, key []byte, data []byte) error {
	exists, err := tx.Exists(key)
	if err != nil {
		return err
	}
	if exists {
		return common.ErrDuplicateVertex
	}

	if err := tx.Set(key, data); err != nil {
		return err
	}

	if err := cm.createVertexIndices(tx, vertex); err != nil {
		return fmt.Errorf("failed to create vertex indices: %w", err)
	}

	return nil
}

func (cm *CRUDManager) GetVertex(vertexID string) (*common.Vertex, error) {
	if err := common.ValidateKey(vertexID); err != nil {
		return nil, fmt.Errorf("invalid vertex ID: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertexID)
	data, err := cm.store.Get(key)
	if err == ErrKeyNotFound {
		return nil, common.ErrVertexNotFound
	}
	if err != nil {
		return nil, err
	}

	vertex := &common.Vertex{}
	if err := vertex.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vertex: %w", err)
	}

	return vertex, nil
}

// UpdateVertex updates a vertex, creating its own transaction
func (cm *CRUDManager) UpdateVertex(vertex *common.Vertex) error {
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex validation failed: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertex.ID)

	return cm.store.Update(func(tx Transaction) error {
		return cm.updateVertexInTx(tx, vertex, key)
	})
}

// UpdateVertexInTx updates a vertex using the provided transaction
func (cm *CRUDManager) UpdateVertexInTx(tx Transaction, vertex *common.Vertex) error {
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex validation failed: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertex.ID)
	return cm.updateVertexInTx(tx, vertex, key)
}

// updateVertexInTx is the internal implementation
func (cm *CRUDManager) updateVertexInTx(tx Transaction, vertex *common.Vertex, key []byte) error {
	exists, err := tx.Exists(key)
	if err != nil {
		return err
	}
	if !exists {
		return common.ErrVertexNotFound
	}

	oldData, err := tx.Get(key)
	if err != nil {
		return err
	}

	oldVertex := &common.Vertex{}
	if err := oldVertex.UnmarshalBinary(oldData); err != nil {
		return fmt.Errorf("failed to unmarshal old vertex: %w", err)
	}

	if err := cm.deleteVertexIndices(tx, oldVertex); err != nil {
		return fmt.Errorf("failed to delete old vertex indices: %w", err)
	}

	vertex.UpdatedAt = time.Now().UTC()
	vertex.Version = oldVertex.Version + 1

	newData, err := vertex.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal updated vertex: %w", err)
	}

	if err := tx.Set(key, newData); err != nil {
		return err
	}

	if err := cm.createVertexIndices(tx, vertex); err != nil {
		return fmt.Errorf("failed to create new vertex indices: %w", err)
	}

	return nil
}

// DeleteVertex deletes a vertex, creating its own transaction
func (cm *CRUDManager) DeleteVertex(vertexID string) error {
	if err := common.ValidateKey(vertexID); err != nil {
		return fmt.Errorf("invalid vertex ID: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertexID)

	return cm.store.Update(func(tx Transaction) error {
		return cm.deleteVertexInTx(tx, vertexID, key)
	})
}

// DeleteVertexInTx deletes a vertex using the provided transaction
func (cm *CRUDManager) DeleteVertexInTx(tx Transaction, vertexID string) error {
	if err := common.ValidateKey(vertexID); err != nil {
		return fmt.Errorf("invalid vertex ID: %w", err)
	}

	key := cm.keyBuilder.VertexKey(vertexID)
	return cm.deleteVertexInTx(tx, vertexID, key)
}

// deleteVertexInTx is the internal implementation
func (cm *CRUDManager) deleteVertexInTx(tx Transaction, vertexID string, key []byte) error {
	data, err := tx.Get(key)
	if err == ErrKeyNotFound {
		return common.ErrVertexNotFound
	}
	if err != nil {
		return err
	}

	vertex := &common.Vertex{}
	if err := vertex.UnmarshalBinary(data); err != nil {
		return fmt.Errorf("failed to unmarshal vertex: %w", err)
	}

	if err := cm.checkVertexReferences(tx, vertexID); err != nil {
		return err
	}

	if err := cm.deleteVertexIndices(tx, vertex); err != nil {
		return fmt.Errorf("failed to delete vertex indices: %w", err)
	}

	return tx.Delete(key)
}

// StoreEdge stores an edge, creating its own transaction
func (cm *CRUDManager) StoreEdge(edge *common.Edge) error {
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("edge validation failed: %w", err)
	}

	key := cm.keyBuilder.EdgeKey(edge.ID)
	data, err := edge.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal edge: %w", err)
	}

	return cm.store.Update(func(tx Transaction) error {
		return cm.storeEdgeInTx(tx, edge, key, data)
	})
}

// StoreEdgeInTx stores an edge using the provided transaction
func (cm *CRUDManager) StoreEdgeInTx(tx Transaction, edge *common.Edge) error {
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("edge validation failed: %w", err)
	}

	key := cm.keyBuilder.EdgeKey(edge.ID)
	data, err := edge.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal edge: %w", err)
	}

	return cm.storeEdgeInTx(tx, edge, key, data)
}

// storeEdgeInTx is the internal implementation
func (cm *CRUDManager) storeEdgeInTx(tx Transaction, edge *common.Edge, key []byte, data []byte) error {
	exists, err := tx.Exists(key)
	if err != nil {
		return err
	}
	if exists {
		return common.ErrDuplicateEdge
	}

	fromExists, err := tx.Exists(cm.keyBuilder.VertexKey(edge.FromVertex))
	if err != nil {
		return err
	}
	if !fromExists {
		return fmt.Errorf("from vertex %s does not exist", edge.FromVertex)
	}

	toExists, err := tx.Exists(cm.keyBuilder.VertexKey(edge.ToVertex))
	if err != nil {
		return err
	}
	if !toExists {
		return fmt.Errorf("to vertex %s does not exist", edge.ToVertex)
	}

	if err := tx.Set(key, data); err != nil {
		return err
	}

	if err := cm.createEdgeIndices(tx, edge); err != nil {
		return fmt.Errorf("failed to create edge indices: %w", err)
	}

	return nil
}

func (cm *CRUDManager) GetEdge(edgeID string) (*common.Edge, error) {
	if err := common.ValidateKey(edgeID); err != nil {
		return nil, fmt.Errorf("invalid edge ID: %w", err)
	}

	key := cm.keyBuilder.EdgeKey(edgeID)
	data, err := cm.store.Get(key)
	if err == ErrKeyNotFound {
		return nil, common.ErrEdgeNotFound
	}
	if err != nil {
		return nil, err
	}

	edge := &common.Edge{}
	if err := edge.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edge: %w", err)
	}

	return edge, nil
}

// DeleteEdge deletes an edge, creating its own transaction
func (cm *CRUDManager) DeleteEdge(edgeID string) error {
	if err := common.ValidateKey(edgeID); err != nil {
		return fmt.Errorf("invalid edge ID: %w", err)
	}

	key := cm.keyBuilder.EdgeKey(edgeID)

	return cm.store.Update(func(tx Transaction) error {
		return cm.deleteEdgeInTx(tx, edgeID, key)
	})
}

// DeleteEdgeInTx deletes an edge using the provided transaction
func (cm *CRUDManager) DeleteEdgeInTx(tx Transaction, edgeID string) error {
	if err := common.ValidateKey(edgeID); err != nil {
		return fmt.Errorf("invalid edge ID: %w", err)
	}

	key := cm.keyBuilder.EdgeKey(edgeID)
	return cm.deleteEdgeInTx(tx, edgeID, key)
}

// deleteEdgeInTx is the internal implementation
func (cm *CRUDManager) deleteEdgeInTx(tx Transaction, edgeID string, key []byte) error {
	data, err := tx.Get(key)
	if err == ErrKeyNotFound {
		return common.ErrEdgeNotFound
	}
	if err != nil {
		return err
	}

	edge := &common.Edge{}
	if err := edge.UnmarshalBinary(data); err != nil {
		return fmt.Errorf("failed to unmarshal edge: %w", err)
	}

	if err := cm.deleteEdgeIndices(tx, edge); err != nil {
		return fmt.Errorf("failed to delete edge indices: %w", err)
	}

	return tx.Delete(key)
}

func (cm *CRUDManager) GetVerticesByType(vertexType string, limit int) ([]*common.Vertex, error) {
	prefix := cm.keyBuilder.VertexTypePrefix(vertexType)

	var vertices []*common.Vertex
	err := cm.store.View(func(tx Transaction) error {
		indexData, err := tx.Scan(prefix, limit)
		if err != nil {
			return err
		}

		for indexKey := range indexData {
			_, vertexID, ok := cm.keyParser.ParseVertexTypeIndexKey([]byte(indexKey))
			if !ok {
				continue
			}

			vertex, err := cm.getVertexInTx(tx, vertexID)
			if err != nil {
				continue
			}

			vertices = append(vertices, vertex)
		}
		return nil
	})

	return vertices, err
}

func (cm *CRUDManager) GetEdgesByType(edgeType string, limit int) ([]*common.Edge, error) {
	prefix := cm.keyBuilder.EdgeTypePrefix(edgeType)

	var edges []*common.Edge
	err := cm.store.View(func(tx Transaction) error {
		indexData, err := tx.Scan(prefix, limit)
		if err != nil {
			return err
		}

		for indexKey := range indexData {
			_, edgeID, ok := cm.keyParser.ParseEdgeTypeIndexKey([]byte(indexKey))
			if !ok {
				continue
			}

			edge, err := cm.getEdgeInTx(tx, edgeID)
			if err != nil {
				continue
			}

			edges = append(edges, edge)
		}
		return nil
	})

	return edges, err
}

func (cm *CRUDManager) GetAllVertices(limit int) ([]*common.Vertex, error) {
	prefix := cm.keyBuilder.AllVerticesPrefix()

	var vertices []*common.Vertex
	err := cm.store.View(func(tx Transaction) error {
		data, err := tx.Scan(prefix, limit)
		if err != nil {
			return err
		}

		for _, vertexData := range data {
			vertex := &common.Vertex{}
			if err := vertex.UnmarshalBinary(vertexData); err != nil {
				continue
			}
			vertices = append(vertices, vertex)
		}
		return nil
	})

	return vertices, err
}

func (cm *CRUDManager) GetAllEdges(limit int) ([]*common.Edge, error) {
	prefix := cm.keyBuilder.AllEdgesPrefix()

	var edges []*common.Edge
	err := cm.store.View(func(tx Transaction) error {
		data, err := tx.Scan(prefix, limit)
		if err != nil {
			return err
		}

		for _, edgeData := range data {
			edge := &common.Edge{}
			if err := edge.UnmarshalBinary(edgeData); err != nil {
				continue
			}
			edges = append(edges, edge)
		}
		return nil
	})

	return edges, err
}

func (cm *CRUDManager) createVertexIndices(tx Transaction, vertex *common.Vertex) error {
	typeIndexKey := cm.keyBuilder.VertexTypeIndexKey(vertex.Type, vertex.ID)
	if err := tx.Set(typeIndexKey, []byte{}); err != nil {
		return err
	}

	for propName, propValue := range vertex.Properties {
		propValueStr := fmt.Sprintf("%v", propValue)
		propIndexKey := cm.keyBuilder.PropertyIndexKey(propName, propValueStr, vertex.ID)
		if err := tx.Set(propIndexKey, []byte{}); err != nil {
			return err
		}
	}

	return nil
}

func (cm *CRUDManager) deleteVertexIndices(tx Transaction, vertex *common.Vertex) error {
	typeIndexKey := cm.keyBuilder.VertexTypeIndexKey(vertex.Type, vertex.ID)
	if err := tx.Delete(typeIndexKey); err != nil {
		return err
	}

	for propName, propValue := range vertex.Properties {
		propValueStr := fmt.Sprintf("%v", propValue)
		propIndexKey := cm.keyBuilder.PropertyIndexKey(propName, propValueStr, vertex.ID)
		if err := tx.Delete(propIndexKey); err != nil {
			return err
		}
	}

	return nil
}

func (cm *CRUDManager) createEdgeIndices(tx Transaction, edge *common.Edge) error {
	typeIndexKey := cm.keyBuilder.EdgeTypeIndexKey(edge.Type, edge.ID)
	if err := tx.Set(typeIndexKey, []byte{}); err != nil {
		return err
	}

	outgoingKey := cm.keyBuilder.OutgoingEdgeIndexKey(edge.FromVertex, edge.ID)
	if err := tx.Set(outgoingKey, []byte(edge.Type)); err != nil {
		return err
	}

	incomingKey := cm.keyBuilder.IncomingEdgeIndexKey(edge.ToVertex, edge.ID)
	if err := tx.Set(incomingKey, []byte(edge.Type)); err != nil {
		return err
	}

	for propName, propValue := range edge.Properties {
		propValueStr := fmt.Sprintf("%v", propValue)
		propIndexKey := cm.keyBuilder.PropertyIndexKey(propName, propValueStr, edge.ID)
		if err := tx.Set(propIndexKey, []byte{}); err != nil {
			return err
		}
	}

	return nil
}

func (cm *CRUDManager) deleteEdgeIndices(tx Transaction, edge *common.Edge) error {
	typeIndexKey := cm.keyBuilder.EdgeTypeIndexKey(edge.Type, edge.ID)
	if err := tx.Delete(typeIndexKey); err != nil {
		return err
	}

	outgoingKey := cm.keyBuilder.OutgoingEdgeIndexKey(edge.FromVertex, edge.ID)
	if err := tx.Delete(outgoingKey); err != nil {
		return err
	}

	incomingKey := cm.keyBuilder.IncomingEdgeIndexKey(edge.ToVertex, edge.ID)
	if err := tx.Delete(incomingKey); err != nil {
		return err
	}

	for propName, propValue := range edge.Properties {
		propValueStr := fmt.Sprintf("%v", propValue)
		propIndexKey := cm.keyBuilder.PropertyIndexKey(propName, propValueStr, edge.ID)
		if err := tx.Delete(propIndexKey); err != nil {
			return err
		}
	}

	return nil
}

func (cm *CRUDManager) checkVertexReferences(tx Transaction, vertexID string) error {
	outgoingPrefix := cm.keyBuilder.OutgoingEdgePrefix(vertexID)
	outgoingEdges, err := tx.Scan(outgoingPrefix, 1)
	if err != nil {
		return err
	}
	if len(outgoingEdges) > 0 {
		return fmt.Errorf("cannot delete vertex %s: has outgoing edges", vertexID)
	}

	incomingPrefix := cm.keyBuilder.IncomingEdgePrefix(vertexID)
	incomingEdges, err := tx.Scan(incomingPrefix, 1)
	if err != nil {
		return err
	}
	if len(incomingEdges) > 0 {
		return fmt.Errorf("cannot delete vertex %s: has incoming edges", vertexID)
	}

	return nil
}

func (cm *CRUDManager) getVertexInTx(tx Transaction, vertexID string) (*common.Vertex, error) {
	key := cm.keyBuilder.VertexKey(vertexID)
	data, err := tx.Get(key)
	if err != nil {
		return nil, err
	}

	vertex := &common.Vertex{}
	if err := vertex.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return vertex, nil
}

func (cm *CRUDManager) getEdgeInTx(tx Transaction, edgeID string) (*common.Edge, error) {
	key := cm.keyBuilder.EdgeKey(edgeID)
	data, err := tx.Get(key)
	if err != nil {
		return nil, err
	}

	edge := &common.Edge{}
	if err := edge.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return edge, nil
}
