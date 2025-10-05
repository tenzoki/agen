package graph

import (
	"fmt"
	"strings"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/storage"
)

// GraphStore defines the interface for graph database operations
type GraphStore interface {
	// Vertex operations
	AddVertex(vertex *common.Vertex) error
	GetVertex(id string) (*common.Vertex, error)
	UpdateVertex(vertex *common.Vertex) error
	DeleteVertex(id string) error
	VertexExists(id string) (bool, error)

	// Vertex operations with transaction
	AddVertexInTx(tx storage.Transaction, vertex *common.Vertex) error
	GetVertexInTx(tx storage.Transaction, id string) (*common.Vertex, error)
	UpdateVertexInTx(tx storage.Transaction, vertex *common.Vertex) error
	DeleteVertexInTx(tx storage.Transaction, id string) error
	VertexExistsInTx(tx storage.Transaction, id string) (bool, error)

	// Edge operations
	AddEdge(edge *common.Edge) error
	GetEdge(edgeID string) (*common.Edge, error)
	DeleteEdge(edgeID string) error
	EdgeExists(edgeID string) (bool, error)

	// Edge operations with transaction
	AddEdgeInTx(tx storage.Transaction, edge *common.Edge) error
	GetEdgeInTx(tx storage.Transaction, edgeID string) (*common.Edge, error)
	DeleteEdgeInTx(tx storage.Transaction, edgeID string) error
	EdgeExistsInTx(tx storage.Transaction, edgeID string) (bool, error)

	// Query operations
	GetVerticesByType(vertexType string, limit int) ([]*common.Vertex, error)
	GetEdgesByType(edgeType string, limit int) ([]*common.Edge, error)
	GetAllVertices(limit int) ([]*common.Vertex, error)
	GetAllEdges(limit int) ([]*common.Edge, error)

	// Traversal operations
	GetOutgoingEdges(vertexID string) ([]*common.Edge, error)
	GetIncomingEdges(vertexID string) ([]*common.Edge, error)
	GetNeighbors(vertexID string, direction TraversalDirection) ([]*common.Vertex, error)

	// Advanced traversal
	TraverseBFS(startVertexID string, direction TraversalDirection, maxDepth int, visitFn func(*common.Vertex, int) bool) error
	TraverseDFS(startVertexID string, direction TraversalDirection, maxDepth int, visitFn func(*common.Vertex, int) bool) error
	FindPath(fromVertexID, toVertexID string, direction TraversalDirection, maxDepth int) ([]*common.Vertex, error)

	// Batch operations
	BatchAddVertices(vertices []*common.Vertex) error
	BatchAddEdges(edges []*common.Edge) error

	// Statistics and maintenance
	GetStats() (*GraphStats, error)
	Close() error
}

// TraversalDirection defines the direction for graph traversal
type TraversalDirection int

const (
	DirectionOutgoing TraversalDirection = iota
	DirectionIncoming
	DirectionBoth
)

// GraphStats provides statistics about the graph store
type GraphStats struct {
	VertexCount   int64     `json:"vertex_count"`
	EdgeCount     int64     `json:"edge_count"`
	VertexTypes   []string  `json:"vertex_types"`
	EdgeTypes     []string  `json:"edge_types"`
	TotalSize     int64     `json:"total_size"`
	LastAccess    time.Time `json:"last_access"`
	IndexCount    int64     `json:"index_count"`
	AvgVertexSize float64   `json:"avg_vertex_size"`
	AvgEdgeSize   float64   `json:"avg_edge_size"`
}

// graphStore implements the GraphStore interface
type graphStore struct {
	crudManager *storage.CRUDManager
	store       storage.Store
}

// NewGraphStore creates a new graph store instance
func NewGraphStore(store storage.Store) GraphStore {
	return &graphStore{
		crudManager: storage.NewCRUDManager(store),
		store:       store,
	}
}

// AddVertex adds a new vertex to the graph
func (gs *graphStore) AddVertex(vertex *common.Vertex) error {
	return gs.crudManager.StoreVertex(vertex)
}

// AddVertexInTx adds a new vertex using the provided transaction
func (gs *graphStore) AddVertexInTx(tx storage.Transaction, vertex *common.Vertex) error {
	return gs.crudManager.StoreVertexInTx(tx, vertex)
}

// GetVertex retrieves a vertex by its ID
func (gs *graphStore) GetVertex(id string) (*common.Vertex, error) {
	return gs.crudManager.GetVertex(id)
}

// GetVertexInTx retrieves a vertex using the provided transaction
func (gs *graphStore) GetVertexInTx(tx storage.Transaction, id string) (*common.Vertex, error) {
	if err := common.ValidateKey(id); err != nil {
		return nil, fmt.Errorf("invalid vertex ID: %w", err)
	}

	key := common.NewKeyBuilder().VertexKey(id)
	data, err := tx.Get(key)
	if err == storage.ErrKeyNotFound {
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

// UpdateVertex updates an existing vertex
func (gs *graphStore) UpdateVertex(vertex *common.Vertex) error {
	return gs.crudManager.UpdateVertex(vertex)
}

// UpdateVertexInTx updates an existing vertex using the provided transaction
func (gs *graphStore) UpdateVertexInTx(tx storage.Transaction, vertex *common.Vertex) error {
	return gs.crudManager.UpdateVertexInTx(tx, vertex)
}

// DeleteVertex removes a vertex from the graph
func (gs *graphStore) DeleteVertex(id string) error {
	return gs.crudManager.DeleteVertex(id)
}

// DeleteVertexInTx removes a vertex using the provided transaction
func (gs *graphStore) DeleteVertexInTx(tx storage.Transaction, id string) error {
	return gs.crudManager.DeleteVertexInTx(tx, id)
}

// VertexExists checks if a vertex exists
func (gs *graphStore) VertexExists(id string) (bool, error) {
	_, err := gs.crudManager.GetVertex(id)
	if err == common.ErrVertexNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// VertexExistsInTx checks if a vertex exists using the provided transaction
func (gs *graphStore) VertexExistsInTx(tx storage.Transaction, id string) (bool, error) {
	_, err := gs.GetVertexInTx(tx, id)
	if err == common.ErrVertexNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddEdge adds a new edge to the graph
func (gs *graphStore) AddEdge(edge *common.Edge) error {
	return gs.crudManager.StoreEdge(edge)
}

// AddEdgeInTx adds a new edge using the provided transaction
func (gs *graphStore) AddEdgeInTx(tx storage.Transaction, edge *common.Edge) error {
	return gs.crudManager.StoreEdgeInTx(tx, edge)
}

// GetEdge retrieves an edge by its ID
func (gs *graphStore) GetEdge(edgeID string) (*common.Edge, error) {
	return gs.crudManager.GetEdge(edgeID)
}

// GetEdgeInTx retrieves an edge using the provided transaction
func (gs *graphStore) GetEdgeInTx(tx storage.Transaction, edgeID string) (*common.Edge, error) {
	if err := common.ValidateKey(edgeID); err != nil {
		return nil, fmt.Errorf("invalid edge ID: %w", err)
	}

	key := common.NewKeyBuilder().EdgeKey(edgeID)
	data, err := tx.Get(key)
	if err == storage.ErrKeyNotFound {
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

// DeleteEdge removes an edge from the graph
func (gs *graphStore) DeleteEdge(edgeID string) error {
	return gs.crudManager.DeleteEdge(edgeID)
}

// DeleteEdgeInTx removes an edge using the provided transaction
func (gs *graphStore) DeleteEdgeInTx(tx storage.Transaction, edgeID string) error {
	return gs.crudManager.DeleteEdgeInTx(tx, edgeID)
}

// EdgeExists checks if an edge exists
func (gs *graphStore) EdgeExists(edgeID string) (bool, error) {
	_, err := gs.crudManager.GetEdge(edgeID)
	if err == common.ErrEdgeNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// EdgeExistsInTx checks if an edge exists using the provided transaction
func (gs *graphStore) EdgeExistsInTx(tx storage.Transaction, edgeID string) (bool, error) {
	_, err := gs.GetEdgeInTx(tx, edgeID)
	if err == common.ErrEdgeNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetVerticesByType retrieves all vertices of a specific type
func (gs *graphStore) GetVerticesByType(vertexType string, limit int) ([]*common.Vertex, error) {
	return gs.crudManager.GetVerticesByType(vertexType, limit)
}

// GetEdgesByType retrieves all edges of a specific type
func (gs *graphStore) GetEdgesByType(edgeType string, limit int) ([]*common.Edge, error) {
	return gs.crudManager.GetEdgesByType(edgeType, limit)
}

// GetAllVertices retrieves all vertices with optional limit
func (gs *graphStore) GetAllVertices(limit int) ([]*common.Vertex, error) {
	return gs.crudManager.GetAllVertices(limit)
}

// GetAllEdges retrieves all edges with optional limit
func (gs *graphStore) GetAllEdges(limit int) ([]*common.Edge, error) {
	return gs.crudManager.GetAllEdges(limit)
}

// GetOutgoingEdges retrieves all edges outgoing from a vertex
func (gs *graphStore) GetOutgoingEdges(vertexID string) ([]*common.Edge, error) {
	if err := common.ValidateKey(vertexID); err != nil {
		return nil, err
	}

	keyBuilder := common.NewKeyBuilder()
	prefix := keyBuilder.OutgoingEdgePrefix(vertexID)

	var edges []*common.Edge
	err := gs.store.View(func(tx storage.Transaction) error {
		indexData, err := tx.Scan(prefix, -1)
		if err != nil {
			return err
		}

		for indexKey := range indexData {
			// Remove the prefix to get just the edge ID
			edgeID := strings.TrimPrefix(indexKey, string(prefix))
			if edgeID == "" {
				continue
			}

			edge, err := gs.crudManager.GetEdge(edgeID)
			if err == common.ErrEdgeNotFound {
				continue // Index might be stale
			}
			if err != nil {
				return err
			}

			edges = append(edges, edge)
		}
		return nil
	})

	return edges, err
}

// GetIncomingEdges retrieves all edges incoming to a vertex
func (gs *graphStore) GetIncomingEdges(vertexID string) ([]*common.Edge, error) {
	if err := common.ValidateKey(vertexID); err != nil {
		return nil, err
	}

	keyBuilder := common.NewKeyBuilder()
	prefix := keyBuilder.IncomingEdgePrefix(vertexID)

	var edges []*common.Edge
	err := gs.store.View(func(tx storage.Transaction) error {
		indexData, err := tx.Scan(prefix, -1)
		if err != nil {
			return err
		}

		for indexKey := range indexData {
			// Remove the prefix to get just the edge ID
			edgeID := strings.TrimPrefix(indexKey, string(prefix))
			if edgeID == "" {
				continue
			}

			edge, err := gs.crudManager.GetEdge(edgeID)
			if err == common.ErrEdgeNotFound {
				continue // Index might be stale
			}
			if err != nil {
				return err
			}

			edges = append(edges, edge)
		}
		return nil
	})

	return edges, err
}

// GetNeighbors retrieves neighboring vertices in the specified direction
func (gs *graphStore) GetNeighbors(vertexID string, direction TraversalDirection) ([]*common.Vertex, error) {
	var edges []*common.Edge
	var err error

	switch direction {
	case DirectionOutgoing:
		edges, err = gs.GetOutgoingEdges(vertexID)
	case DirectionIncoming:
		edges, err = gs.GetIncomingEdges(vertexID)
	case DirectionBoth:
		outgoing, err1 := gs.GetOutgoingEdges(vertexID)
		if err1 != nil {
			return nil, err1
		}
		incoming, err2 := gs.GetIncomingEdges(vertexID)
		if err2 != nil {
			return nil, err2
		}
		edges = append(outgoing, incoming...)
	default:
		return nil, common.ErrInvalidDirection
	}

	if err != nil {
		return nil, err
	}

	// Get unique neighbor vertices
	neighborMap := make(map[string]*common.Vertex)
	for _, edge := range edges {
		var neighborID string
		if edge.FromVertex == vertexID {
			neighborID = edge.ToVertex
		} else {
			neighborID = edge.FromVertex
		}

		if _, exists := neighborMap[neighborID]; !exists {
			vertex, err := gs.crudManager.GetVertex(neighborID)
			if err == common.ErrVertexNotFound {
				continue // Vertex might have been deleted
			}
			if err != nil {
				return nil, err
			}
			neighborMap[neighborID] = vertex
		}
	}

	// Convert map to slice
	neighbors := make([]*common.Vertex, 0, len(neighborMap))
	for _, vertex := range neighborMap {
		neighbors = append(neighbors, vertex)
	}

	return neighbors, nil
}

// TraverseBFS performs breadth-first traversal starting from the given vertex
// visitFn is called for each vertex with the vertex and its depth
// If visitFn returns false, traversal stops
func (gs *graphStore) TraverseBFS(startVertexID string, direction TraversalDirection, maxDepth int, visitFn func(*common.Vertex, int) bool) error {
	if err := common.ValidateKey(startVertexID); err != nil {
		return err
	}

	// Check if start vertex exists
	startVertex, err := gs.crudManager.GetVertex(startVertexID)
	if err != nil {
		return err
	}

	visited := make(map[string]bool)
	queue := []*vertexWithDepth{{vertex: startVertex, depth: 0}}
	visited[startVertexID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Visit current vertex
		if !visitFn(current.vertex, current.depth) {
			break
		}

		// Stop if we've reached max depth
		if maxDepth >= 0 && current.depth >= maxDepth {
			continue
		}

		// Add neighbors to queue
		neighbors, err := gs.GetNeighbors(current.vertex.ID, direction)
		if err != nil {
			return err
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor.ID] {
				visited[neighbor.ID] = true
				queue = append(queue, &vertexWithDepth{
					vertex: neighbor,
					depth:  current.depth + 1,
				})
			}
		}
	}

	return nil
}

// TraverseDFS performs depth-first traversal starting from the given vertex
// visitFn is called for each vertex with the vertex and its depth
// If visitFn returns false, traversal stops
func (gs *graphStore) TraverseDFS(startVertexID string, direction TraversalDirection, maxDepth int, visitFn func(*common.Vertex, int) bool) error {
	if err := common.ValidateKey(startVertexID); err != nil {
		return err
	}

	// Check if start vertex exists
	startVertex, err := gs.crudManager.GetVertex(startVertexID)
	if err != nil {
		return err
	}

	visited := make(map[string]bool)

	var dfsRecursive func(*common.Vertex, int) bool
	dfsRecursive = func(vertex *common.Vertex, depth int) bool {
		// Visit current vertex
		if !visitFn(vertex, depth) {
			return false
		}

		visited[vertex.ID] = true

		// Stop if we've reached max depth
		if maxDepth >= 0 && depth >= maxDepth {
			return true
		}

		// Recursively visit neighbors
		neighbors, err := gs.GetNeighbors(vertex.ID, direction)
		if err != nil {
			return false
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor.ID] {
				if !dfsRecursive(neighbor, depth+1) {
					return false
				}
			}
		}

		return true
	}

	dfsRecursive(startVertex, 0)
	return nil
}

// FindPath finds a path between two vertices using BFS
// Returns the path as a slice of vertices, or nil if no path exists
func (gs *graphStore) FindPath(fromVertexID, toVertexID string, direction TraversalDirection, maxDepth int) ([]*common.Vertex, error) {
	if err := common.ValidateKey(fromVertexID); err != nil {
		return nil, err
	}
	if err := common.ValidateKey(toVertexID); err != nil {
		return nil, err
	}

	if fromVertexID == toVertexID {
		vertex, err := gs.crudManager.GetVertex(fromVertexID)
		if err != nil {
			return nil, err
		}
		return []*common.Vertex{vertex}, nil
	}

	// Check if start vertex exists
	startVertex, err := gs.crudManager.GetVertex(fromVertexID)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	parent := make(map[string]*common.Vertex)
	queue := []*vertexWithDepth{{vertex: startVertex, depth: 0}}
	visited[fromVertexID] = true

	var targetVertex *common.Vertex

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Check if we found the target
		if current.vertex.ID == toVertexID {
			targetVertex = current.vertex
			break
		}

		// Stop if we've reached max depth
		if maxDepth >= 0 && current.depth >= maxDepth {
			continue
		}

		// Add neighbors to queue
		neighbors, err := gs.GetNeighbors(current.vertex.ID, direction)
		if err != nil {
			return nil, err
		}

		for _, neighbor := range neighbors {
			if !visited[neighbor.ID] {
				visited[neighbor.ID] = true
				parent[neighbor.ID] = current.vertex
				queue = append(queue, &vertexWithDepth{
					vertex: neighbor,
					depth:  current.depth + 1,
				})
			}
		}
	}

	// If target not found, return nil
	if targetVertex == nil {
		return nil, nil
	}

	// Reconstruct path
	path := []*common.Vertex{}
	current := targetVertex

	for current != nil {
		path = append([]*common.Vertex{current}, path...)
		current = parent[current.ID]
	}

	return path, nil
}

// vertexWithDepth is a helper struct for traversal algorithms
type vertexWithDepth struct {
	vertex *common.Vertex
	depth  int
}

// BatchAddVertices adds multiple vertices sequentially
func (gs *graphStore) BatchAddVertices(vertices []*common.Vertex) error {
	for _, vertex := range vertices {
		if err := gs.crudManager.StoreVertex(vertex); err != nil {
			return err
		}
	}
	return nil
}

// BatchAddEdges adds multiple edges sequentially
func (gs *graphStore) BatchAddEdges(edges []*common.Edge) error {
	for _, edge := range edges {
		if err := gs.crudManager.StoreEdge(edge); err != nil {
			return err
		}
	}
	return nil
}

// GetStats returns comprehensive statistics about the graph
func (gs *graphStore) GetStats() (*GraphStats, error) {
	var stats GraphStats
	var vertexCount, edgeCount, indexCount int64
	var totalSize int64
	var totalVertexSize, totalEdgeSize int64
	vertexTypes := make(map[string]bool)
	edgeTypes := make(map[string]bool)

	err := gs.store.View(func(tx storage.Transaction) error {
		// Count vertices and collect types
		vertices, err := tx.Scan([]byte("v:"), -1)
		if err != nil {
			return err
		}

		for _, data := range vertices {
			vertexCount++
			totalVertexSize += int64(len(data))

			// Parse vertex to get type
			var vertex common.Vertex
			if err := vertex.UnmarshalBinary(data); err == nil {
				vertexTypes[vertex.Type] = true
			}
		}

		// Count edges and collect types
		edges, err := tx.Scan([]byte("e:"), -1)
		if err != nil {
			return err
		}

		for _, data := range edges {
			edgeCount++
			totalEdgeSize += int64(len(data))

			// Parse edge to get type
			var edge common.Edge
			if err := edge.UnmarshalBinary(data); err == nil {
				edgeTypes[edge.Type] = true
			}
		}

		// Count indices
		indices, err := tx.Scan([]byte("idx:"), -1)
		if err != nil {
			return err
		}
		indexCount = int64(len(indices))

		totalSize = totalVertexSize + totalEdgeSize
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert type maps to slices
	stats.VertexTypes = make([]string, 0, len(vertexTypes))
	for vType := range vertexTypes {
		stats.VertexTypes = append(stats.VertexTypes, vType)
	}

	stats.EdgeTypes = make([]string, 0, len(edgeTypes))
	for eType := range edgeTypes {
		stats.EdgeTypes = append(stats.EdgeTypes, eType)
	}

	stats.VertexCount = vertexCount
	stats.EdgeCount = edgeCount
	stats.IndexCount = indexCount
	stats.TotalSize = totalSize
	stats.LastAccess = time.Now()

	if vertexCount > 0 {
		stats.AvgVertexSize = float64(totalVertexSize) / float64(vertexCount)
	}
	if edgeCount > 0 {
		stats.AvgEdgeSize = float64(totalEdgeSize) / float64(edgeCount)
	}

	return &stats, nil
}

// Close closes the graph store
func (gs *graphStore) Close() error {
	return gs.store.Close()
}
