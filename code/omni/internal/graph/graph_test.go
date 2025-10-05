package graph

import (
	"os"
	"testing"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestGraphStore(t *testing.T) (GraphStore, func()) {
	tmpDir := t.TempDir()

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	require.NoError(t, err)

	graphStore := NewGraphStore(store)

	cleanup := func() {
		graphStore.Close()
		os.RemoveAll(tmpDir)
	}

	return graphStore, cleanup
}

func TestGraphStore_VertexOperations(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create test vertex
	vertex := common.NewVertex("user:1", "User")
	vertex.Properties["name"] = "Alice"
	vertex.Properties["age"] = 25

	// Test AddVertex
	err := gs.AddVertex(vertex)
	require.NoError(t, err)

	// Test VertexExists
	exists, err := gs.VertexExists("user:1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = gs.VertexExists("user:999")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test GetVertex
	retrieved, err := gs.GetVertex("user:1")
	require.NoError(t, err)
	assert.Equal(t, vertex.ID, retrieved.ID)
	assert.Equal(t, vertex.Type, retrieved.Type)
	assert.Equal(t, "Alice", retrieved.Properties["name"])
	assert.Equal(t, int8(25), retrieved.Properties["age"])

	// Test UpdateVertex
	vertex.Properties["age"] = 26
	err = gs.UpdateVertex(vertex)
	require.NoError(t, err)

	updated, err := gs.GetVertex("user:1")
	require.NoError(t, err)
	assert.Equal(t, int8(26), updated.Properties["age"])
	assert.Equal(t, uint64(2), updated.Version) // UpdateVertex automatically increments version

	// Test DeleteVertex
	err = gs.DeleteVertex("user:1")
	require.NoError(t, err)

	_, err = gs.GetVertex("user:1")
	assert.Equal(t, common.ErrVertexNotFound, err)

	exists, err = gs.VertexExists("user:1")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGraphStore_EdgeOperations(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create test vertices
	alice := common.NewVertex("user:alice", "User")
	bob := common.NewVertex("user:bob", "User")

	err := gs.AddVertex(alice)
	require.NoError(t, err)
	err = gs.AddVertex(bob)
	require.NoError(t, err)

	// Create test edge
	edge := common.NewEdge("follows:alice:bob", "follows", "user:alice", "user:bob")
	edge.Properties["since"] = "2023"

	// Test AddEdge
	err = gs.AddEdge(edge)
	require.NoError(t, err)

	// Test EdgeExists
	exists, err := gs.EdgeExists("follows:alice:bob")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = gs.EdgeExists("follows:bob:alice")
	require.NoError(t, err)
	assert.False(t, exists)

	// Test GetEdge
	retrieved, err := gs.GetEdge("follows:alice:bob")
	require.NoError(t, err)
	assert.Equal(t, edge.ID, retrieved.ID)
	assert.Equal(t, edge.Type, retrieved.Type)
	assert.Equal(t, edge.FromVertex, retrieved.FromVertex)
	assert.Equal(t, edge.ToVertex, retrieved.ToVertex)
	assert.Equal(t, "2023", retrieved.Properties["since"])

	// Test DeleteEdge
	err = gs.DeleteEdge("follows:alice:bob")
	require.NoError(t, err)

	_, err = gs.GetEdge("follows:alice:bob")
	assert.Equal(t, common.ErrEdgeNotFound, err)

	exists, err = gs.EdgeExists("follows:alice:bob")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGraphStore_QueryOperations(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create test data
	users := []*common.Vertex{
		common.NewVertex("user:1", "User"),
		common.NewVertex("user:2", "User"),
		common.NewVertex("group:1", "Group"),
	}

	for _, user := range users {
		err := gs.AddVertex(user)
		require.NoError(t, err)
	}

	edges := []*common.Edge{
		common.NewEdge("follows:1:2", "follows", "user:1", "user:2"),
		common.NewEdge("member:1:group", "member", "user:1", "group:1"),
	}

	for _, edge := range edges {
		err := gs.AddEdge(edge)
		require.NoError(t, err)
	}

	// Test GetVerticesByType
	userVertices, err := gs.GetVerticesByType("User", -1)
	require.NoError(t, err)
	assert.Len(t, userVertices, 2)

	groupVertices, err := gs.GetVerticesByType("Group", -1)
	require.NoError(t, err)
	assert.Len(t, groupVertices, 1)

	// Test GetEdgesByType
	followsEdges, err := gs.GetEdgesByType("follows", -1)
	require.NoError(t, err)
	assert.Len(t, followsEdges, 1)

	memberEdges, err := gs.GetEdgesByType("member", -1)
	require.NoError(t, err)
	assert.Len(t, memberEdges, 1)

	// Test GetAllVertices
	allVertices, err := gs.GetAllVertices(-1)
	require.NoError(t, err)
	assert.Len(t, allVertices, 3)

	// Test GetAllEdges
	allEdges, err := gs.GetAllEdges(-1)
	require.NoError(t, err)
	assert.Len(t, allEdges, 2)

	// Test with limit
	limitedVertices, err := gs.GetVerticesByType("User", 1)
	require.NoError(t, err)
	assert.Len(t, limitedVertices, 1)
}

func TestGraphStore_BatchOperations(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Test BatchAddVertices
	vertices := []*common.Vertex{
		common.NewVertex("user:batch1", "User"),
		common.NewVertex("user:batch2", "User"),
		common.NewVertex("user:batch3", "User"),
	}

	err := gs.BatchAddVertices(vertices)
	require.NoError(t, err)

	// Verify all vertices were added
	for _, vertex := range vertices {
		exists, err := gs.VertexExists(vertex.ID)
		require.NoError(t, err)
		assert.True(t, exists)
	}

	// Test BatchAddEdges
	edges := []*common.Edge{
		common.NewEdge("edge:batch1", "follows", "user:batch1", "user:batch2"),
		common.NewEdge("edge:batch2", "follows", "user:batch2", "user:batch3"),
	}

	err = gs.BatchAddEdges(edges)
	require.NoError(t, err)

	// Verify all edges were added
	for _, edge := range edges {
		exists, err := gs.EdgeExists(edge.ID)
		require.NoError(t, err)
		assert.True(t, exists)
	}
}

func TestGraphStore_Statistics(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Initially empty
	stats, err := gs.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(0), stats.VertexCount)
	assert.Equal(t, int64(0), stats.EdgeCount)
	assert.Equal(t, int64(0), stats.TotalSize)
	assert.Empty(t, stats.VertexTypes)
	assert.Empty(t, stats.EdgeTypes)

	// Add some test data
	vertices := []*common.Vertex{
		common.NewVertex("user:1", "User"),
		common.NewVertex("user:2", "User"),
		common.NewVertex("group:1", "Group"),
	}

	for _, vertex := range vertices {
		vertex.Properties["test"] = "data"
		err := gs.AddVertex(vertex)
		require.NoError(t, err)
	}

	edges := []*common.Edge{
		common.NewEdge("follows:1:2", "follows", "user:1", "user:2"),
		common.NewEdge("member:1:group", "member", "user:1", "group:1"),
	}

	for _, edge := range edges {
		edge.Properties["weight"] = 1.0
		err := gs.AddEdge(edge)
		require.NoError(t, err)
	}

	// Check updated stats
	stats, err = gs.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(3), stats.VertexCount)
	assert.Equal(t, int64(2), stats.EdgeCount)
	assert.Greater(t, stats.TotalSize, int64(0))
	assert.Greater(t, stats.AvgVertexSize, 0.0)
	assert.Greater(t, stats.AvgEdgeSize, 0.0)
	assert.Contains(t, stats.VertexTypes, "User")
	assert.Contains(t, stats.VertexTypes, "Group")
	assert.Contains(t, stats.EdgeTypes, "follows")
	assert.Contains(t, stats.EdgeTypes, "member")
	assert.True(t, stats.LastAccess.After(time.Now().Add(-time.Minute)))
}

func TestGraphStore_ErrorHandling(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Test getting non-existent vertex
	_, err := gs.GetVertex("non-existent")
	assert.Equal(t, common.ErrVertexNotFound, err)

	// Test getting non-existent edge
	_, err = gs.GetEdge("non-existent")
	assert.Equal(t, common.ErrEdgeNotFound, err)

	// Test deleting non-existent vertex
	err = gs.DeleteVertex("non-existent")
	assert.Equal(t, common.ErrVertexNotFound, err)

	// Test deleting non-existent edge
	err = gs.DeleteEdge("non-existent")
	assert.Equal(t, common.ErrEdgeNotFound, err)

	// Test invalid vertex
	invalidVertex := &common.Vertex{} // Empty vertex
	err = gs.AddVertex(invalidVertex)
	assert.Error(t, err)

	// Test invalid edge
	invalidEdge := &common.Edge{} // Empty edge
	err = gs.AddEdge(invalidEdge)
	assert.Error(t, err)
}

func TestGraphStore_DuplicateHandling(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create and add a vertex
	vertex := common.NewVertex("user:duplicate", "User")
	err := gs.AddVertex(vertex)
	require.NoError(t, err)

	// Try to add the same vertex again
	err = gs.AddVertex(vertex)
	assert.Equal(t, common.ErrDuplicateVertex, err)

	// Create vertices for edge test
	alice := common.NewVertex("user:alice", "User")
	bob := common.NewVertex("user:bob", "User")
	err = gs.AddVertex(alice)
	require.NoError(t, err)
	err = gs.AddVertex(bob)
	require.NoError(t, err)

	// Create and add an edge
	edge := common.NewEdge("follows:duplicate", "follows", "user:alice", "user:bob")
	err = gs.AddEdge(edge)
	require.NoError(t, err)

	// Try to add the same edge again
	err = gs.AddEdge(edge)
	assert.Equal(t, common.ErrDuplicateEdge, err)
}

func TestGraphStore_TraversalOperations(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create test vertices
	alice := common.NewVertex("user:alice", "User")
	bob := common.NewVertex("user:bob", "User")
	charlie := common.NewVertex("user:charlie", "User")
	dave := common.NewVertex("user:dave", "User")

	vertices := []*common.Vertex{alice, bob, charlie, dave}
	for _, vertex := range vertices {
		err := gs.AddVertex(vertex)
		require.NoError(t, err)
	}

	// Create edges: alice -> bob -> charlie, alice -> dave, dave -> alice (circular)
	edges := []*common.Edge{
		common.NewEdge("follows:alice:bob", "follows", "user:alice", "user:bob"),
		common.NewEdge("follows:bob:charlie", "follows", "user:bob", "user:charlie"),
		common.NewEdge("friend:alice:dave", "friend", "user:alice", "user:dave"),
		common.NewEdge("follows:dave:alice", "follows", "user:dave", "user:alice"),
	}

	for _, edge := range edges {
		err := gs.AddEdge(edge)
		require.NoError(t, err)
	}

	// Test GetOutgoingEdges
	outgoingAlice, err := gs.GetOutgoingEdges("user:alice")
	require.NoError(t, err)
	assert.Len(t, outgoingAlice, 2)

	edgeTypes := make(map[string]bool)
	targetVertices := make(map[string]bool)
	for _, edge := range outgoingAlice {
		assert.Equal(t, "user:alice", edge.FromVertex)
		edgeTypes[edge.Type] = true
		targetVertices[edge.ToVertex] = true
	}
	assert.True(t, edgeTypes["follows"])
	assert.True(t, edgeTypes["friend"])
	assert.True(t, targetVertices["user:bob"])
	assert.True(t, targetVertices["user:dave"])

	outgoingCharlie, err := gs.GetOutgoingEdges("user:charlie")
	require.NoError(t, err)
	assert.Len(t, outgoingCharlie, 0) // Charlie has no outgoing edges

	// Test GetIncomingEdges
	incomingAlice, err := gs.GetIncomingEdges("user:alice")
	require.NoError(t, err)
	assert.Len(t, incomingAlice, 1) // Only dave -> alice
	assert.Equal(t, "user:dave", incomingAlice[0].FromVertex)
	assert.Equal(t, "user:alice", incomingAlice[0].ToVertex)

	incomingBob, err := gs.GetIncomingEdges("user:bob")
	require.NoError(t, err)
	assert.Len(t, incomingBob, 1) // Only alice -> bob
	assert.Equal(t, "user:alice", incomingBob[0].FromVertex)

	// Test GetNeighbors - Outgoing direction
	neighborsOutgoing, err := gs.GetNeighbors("user:alice", DirectionOutgoing)
	require.NoError(t, err)
	assert.Len(t, neighborsOutgoing, 2) // bob and dave

	neighborIDs := make(map[string]bool)
	for _, neighbor := range neighborsOutgoing {
		neighborIDs[neighbor.ID] = true
	}
	assert.True(t, neighborIDs["user:bob"])
	assert.True(t, neighborIDs["user:dave"])

	// Test GetNeighbors - Incoming direction
	neighborsIncoming, err := gs.GetNeighbors("user:alice", DirectionIncoming)
	require.NoError(t, err)
	assert.Len(t, neighborsIncoming, 1) // only dave
	assert.Equal(t, "user:dave", neighborsIncoming[0].ID)

	// Test GetNeighbors - Both directions
	neighborsBoth, err := gs.GetNeighbors("user:alice", DirectionBoth)
	require.NoError(t, err)
	assert.Len(t, neighborsBoth, 2) // bob, dave (dave appears in both directions but should be unique)

	neighborIDsBoth := make(map[string]bool)
	for _, neighbor := range neighborsBoth {
		neighborIDsBoth[neighbor.ID] = true
	}
	assert.True(t, neighborIDsBoth["user:bob"])
	assert.True(t, neighborIDsBoth["user:dave"])

	// Test with vertex that has only incoming edges
	charlieNeighbors, err := gs.GetNeighbors("user:charlie", DirectionBoth)
	require.NoError(t, err)
	assert.Len(t, charlieNeighbors, 1) // charlie has incoming edge from bob
	assert.Equal(t, "user:bob", charlieNeighbors[0].ID)

	// Test invalid direction
	_, err = gs.GetNeighbors("user:alice", TraversalDirection(999))
	assert.Equal(t, common.ErrInvalidDirection, err)
}

func TestGraphStore_TraversalEdgeCases(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Test traversal with non-existent vertex
	edges, err := gs.GetOutgoingEdges("non-existent")
	require.NoError(t, err)
	assert.Len(t, edges, 0)

	neighbors, err := gs.GetNeighbors("non-existent", DirectionBoth)
	require.NoError(t, err)
	assert.Len(t, neighbors, 0)

	// Test with empty key
	_, err = gs.GetOutgoingEdges("")
	assert.Error(t, err)

	_, err = gs.GetIncomingEdges("")
	assert.Error(t, err)

	_, err = gs.GetNeighbors("", DirectionBoth)
	assert.Error(t, err)
}

func TestGraphStore_AdvancedTraversal(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create a more complex graph for traversal testing
	// Structure: alice -> bob -> charlie -> dave -> eve
	//           alice -> frank
	vertices := []*common.Vertex{
		common.NewVertex("user:alice", "User"),
		common.NewVertex("user:bob", "User"),
		common.NewVertex("user:charlie", "User"),
		common.NewVertex("user:dave", "User"),
		common.NewVertex("user:eve", "User"),
		common.NewVertex("user:frank", "User"),
	}

	for _, vertex := range vertices {
		err := gs.AddVertex(vertex)
		require.NoError(t, err)
	}

	edges := []*common.Edge{
		common.NewEdge("alice:bob", "follows", "user:alice", "user:bob"),
		common.NewEdge("bob:charlie", "follows", "user:bob", "user:charlie"),
		common.NewEdge("charlie:dave", "follows", "user:charlie", "user:dave"),
		common.NewEdge("dave:eve", "follows", "user:dave", "user:eve"),
		common.NewEdge("alice:frank", "friend", "user:alice", "user:frank"),
	}

	for _, edge := range edges {
		err := gs.AddEdge(edge)
		require.NoError(t, err)
	}

	// Test BFS traversal
	visitedBFS := []string{}
	depthsBFS := []int{}
	err := gs.TraverseBFS("user:alice", DirectionOutgoing, -1, func(vertex *common.Vertex, depth int) bool {
		visitedBFS = append(visitedBFS, vertex.ID)
		depthsBFS = append(depthsBFS, depth)
		return true
	})
	require.NoError(t, err)

	// BFS should visit in breadth-first order: alice(0), bob(1), frank(1), charlie(2), dave(3), eve(4)
	assert.Equal(t, "user:alice", visitedBFS[0])
	assert.Equal(t, 0, depthsBFS[0])

	// Depth 1 vertices should come next (order may vary)
	depth1Vertices := make(map[string]bool)
	for i := 1; i < len(visitedBFS); i++ {
		if depthsBFS[i] == 1 {
			depth1Vertices[visitedBFS[i]] = true
		}
	}
	assert.True(t, depth1Vertices["user:bob"])
	assert.True(t, depth1Vertices["user:frank"])

	// Test DFS traversal
	visitedDFS := []string{}
	depthsDFS := []int{}
	err = gs.TraverseDFS("user:alice", DirectionOutgoing, -1, func(vertex *common.Vertex, depth int) bool {
		visitedDFS = append(visitedDFS, vertex.ID)
		depthsDFS = append(depthsDFS, depth)
		return true
	})
	require.NoError(t, err)

	// DFS should visit alice first, then go deep
	assert.Equal(t, "user:alice", visitedDFS[0])
	assert.Equal(t, 0, depthsDFS[0])
	assert.Len(t, visitedDFS, 6) // All vertices reachable from alice

	// Test limited depth traversal
	visitedLimited := []string{}
	err = gs.TraverseBFS("user:alice", DirectionOutgoing, 2, func(vertex *common.Vertex, depth int) bool {
		visitedLimited = append(visitedLimited, vertex.ID)
		return true
	})
	require.NoError(t, err)

	// Should only visit vertices up to depth 2
	assert.Contains(t, visitedLimited, "user:alice")   // depth 0
	assert.Contains(t, visitedLimited, "user:bob")     // depth 1
	assert.Contains(t, visitedLimited, "user:frank")   // depth 1
	assert.Contains(t, visitedLimited, "user:charlie") // depth 2
	assert.NotContains(t, visitedLimited, "user:dave") // depth 3
	assert.NotContains(t, visitedLimited, "user:eve")  // depth 4

	// Test early termination
	visitedEarly := []string{}
	err = gs.TraverseBFS("user:alice", DirectionOutgoing, -1, func(vertex *common.Vertex, depth int) bool {
		visitedEarly = append(visitedEarly, vertex.ID)
		return len(visitedEarly) < 3 // Stop after 3 vertices
	})
	require.NoError(t, err)
	assert.Len(t, visitedEarly, 3)
}

func TestGraphStore_PathFinding(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Create a graph: alice -> bob -> charlie
	//                 alice -> dave -> eve
	vertices := []*common.Vertex{
		common.NewVertex("user:alice", "User"),
		common.NewVertex("user:bob", "User"),
		common.NewVertex("user:charlie", "User"),
		common.NewVertex("user:dave", "User"),
		common.NewVertex("user:eve", "User"),
	}

	for _, vertex := range vertices {
		err := gs.AddVertex(vertex)
		require.NoError(t, err)
	}

	edges := []*common.Edge{
		common.NewEdge("alice:bob", "follows", "user:alice", "user:bob"),
		common.NewEdge("bob:charlie", "follows", "user:bob", "user:charlie"),
		common.NewEdge("alice:dave", "friend", "user:alice", "user:dave"),
		common.NewEdge("dave:eve", "follows", "user:dave", "user:eve"),
	}

	for _, edge := range edges {
		err := gs.AddEdge(edge)
		require.NoError(t, err)
	}

	// Test path from alice to charlie
	path, err := gs.FindPath("user:alice", "user:charlie", DirectionOutgoing, -1)
	require.NoError(t, err)
	require.NotNil(t, path)
	assert.Len(t, path, 3) // alice -> bob -> charlie
	assert.Equal(t, "user:alice", path[0].ID)
	assert.Equal(t, "user:bob", path[1].ID)
	assert.Equal(t, "user:charlie", path[2].ID)

	// Test path from alice to eve
	path, err = gs.FindPath("user:alice", "user:eve", DirectionOutgoing, -1)
	require.NoError(t, err)
	require.NotNil(t, path)
	assert.Len(t, path, 3) // alice -> dave -> eve
	assert.Equal(t, "user:alice", path[0].ID)
	assert.Equal(t, "user:dave", path[1].ID)
	assert.Equal(t, "user:eve", path[2].ID)

	// Test path to self
	path, err = gs.FindPath("user:alice", "user:alice", DirectionOutgoing, -1)
	require.NoError(t, err)
	require.NotNil(t, path)
	assert.Len(t, path, 1)
	assert.Equal(t, "user:alice", path[0].ID)

	// Test no path (charlie to alice with outgoing direction)
	path, err = gs.FindPath("user:charlie", "user:alice", DirectionOutgoing, -1)
	require.NoError(t, err)
	assert.Nil(t, path) // No path exists

	// Test path with depth limit that prevents finding target
	path, err = gs.FindPath("user:alice", "user:charlie", DirectionOutgoing, 1)
	require.NoError(t, err)
	assert.Nil(t, path) // Path exists but is longer than depth limit

	// Test path with sufficient depth limit
	path, err = gs.FindPath("user:alice", "user:charlie", DirectionOutgoing, 2)
	require.NoError(t, err)
	require.NotNil(t, path)
	assert.Len(t, path, 3)

	// Test non-existent start vertex (should return error)
	_, err = gs.FindPath("user:nonexistent", "user:alice", DirectionOutgoing, -1)
	assert.Error(t, err)

	// Test non-existent target vertex (should return nil path, no error)
	path, err = gs.FindPath("user:alice", "user:nonexistent", DirectionOutgoing, -1)
	require.NoError(t, err)
	assert.Nil(t, path)
}

func TestGraphStore_Close(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	// Add some data
	vertex := common.NewVertex("user:test", "User")
	err := gs.AddVertex(vertex)
	require.NoError(t, err)

	// Close should work without error
	err = gs.Close()
	assert.NoError(t, err)
}
