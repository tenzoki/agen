//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/graph"
	"github.com/tenzoki/agen/omni/internal/storage"
)

func main() {
	fmt.Println("🌐 Graph Store Complete Demo")
	fmt.Println("==================================================")

	// Setup temporary storage
	tmpDir := "/tmp/graph-store-complete-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	graphStore := graph.NewGraphStore(store)
	defer graphStore.Close()

	// Run all demo sections
	fmt.Println("\n1. Vertex Operations")
	demoVertexOperations(graphStore)

	fmt.Println("\n2. Edge Operations")
	demoEdgeOperations(graphStore)

	fmt.Println("\n3. Type-based Queries")
	demoTypeBasedQueries(graphStore)

	fmt.Println("\n4. Batch Operations")
	demoBatchGraphOperations(graphStore)

	fmt.Println("\n5. Basic Traversal")
	demoBasicTraversal(graphStore)

	fmt.Println("\n6. Graph Statistics")
	printGraphStatistics(graphStore)

	fmt.Println("\n✅ Graph Store demo completed successfully!")
}

func demoVertexOperations(gs graph.GraphStore) {
	fmt.Println("   Comprehensive vertex management...")

	// Create diverse vertex types
	vertices := []*common.Vertex{
		createUser("user:alice", "Alice Johnson", 28, "Engineer"),
		createUser("user:bob", "Bob Smith", 32, "Designer"),
		createUser("user:charlie", "Charlie Brown", 25, "Manager"),
		createCompany("company:techcorp", "TechCorp Inc", "Technology", 1200),
		createCompany("company:designco", "Design Co", "Design", 50),
		createProject("project:webapp", "Web Application", "Active"),
		createProject("project:mobile", "Mobile App", "Planning"),
	}

	fmt.Println("   Creating vertices:")
	for _, vertex := range vertices {
		if err := gs.AddVertex(vertex); err != nil {
			log.Printf("   ❌ Failed to add %s: %v", vertex.ID, err)
		} else {
			fmt.Printf("   ✅ Added %s (%s): %s\n",
				vertex.Type, vertex.ID, vertex.Properties["name"])
		}
	}

	// Demonstrate vertex retrieval
	fmt.Println("\n   Retrieving vertices:")
	testVertexIDs := []string{"user:alice", "company:techcorp", "project:webapp"}
	for _, id := range testVertexIDs {
		if vertex, err := gs.GetVertex(id); err != nil {
			log.Printf("   ❌ Failed to get %s: %v", id, err)
		} else {
			fmt.Printf("   📋 %s: %s (created: %s)\n",
				vertex.ID, vertex.Properties["name"], vertex.CreatedAt.Format(time.RFC3339))
		}
	}

	// Demonstrate vertex existence checking
	fmt.Println("\n   Checking vertex existence:")
	checkIDs := []string{"user:alice", "user:nonexistent", "company:techcorp"}
	for _, id := range checkIDs {
		if exists, err := gs.VertexExists(id); err != nil {
			log.Printf("   ❌ Error checking %s: %v", id, err)
		} else if exists {
			fmt.Printf("   ✅ Exists: %s\n", id)
		} else {
			fmt.Printf("   ❌ Missing: %s\n", id)
		}
	}

	// Demonstrate vertex update
	fmt.Println("\n   Updating vertex:")
	if alice, err := gs.GetVertex("user:alice"); err != nil {
		log.Printf("   ❌ Failed to get Alice: %v", err)
	} else {
		alice.Properties["title"] = "Senior Engineer"
		alice.Properties["location"] = "San Francisco"
		if err := gs.UpdateVertex(alice); err != nil {
			log.Printf("   ❌ Failed to update Alice: %v", err)
		} else {
			fmt.Printf("   ✅ Updated %s: added title and location\n", alice.ID)
		}
	}

	fmt.Printf("\n   📊 Created %d vertices of various types\n", len(vertices))
}

func demoEdgeOperations(gs graph.GraphStore) {
	fmt.Println("   Comprehensive edge management...")

	// Create diverse edge types representing relationships
	edges := []*common.Edge{
		// Employment relationships
		createEmploymentEdge("emp:alice:techcorp", "user:alice", "company:techcorp",
			"Senior Engineer", "2022-01-15"),
		createEmploymentEdge("emp:bob:designco", "user:bob", "company:designco",
			"Lead Designer", "2023-03-01"),

		// Social relationships
		createSocialEdge("follows:alice:bob", "follows", "user:alice", "user:bob", "2023-06-01"),
		createSocialEdge("follows:bob:charlie", "follows", "user:bob", "user:charlie", "2023-07-15"),
		createSocialEdge("friend:alice:charlie", "friend", "user:alice", "user:charlie", "2023-05-10"),

		// Project assignments
		createProjectEdge("assigned:alice:webapp", "user:alice", "project:webapp",
			"Lead Developer", 80),
		createProjectEdge("assigned:bob:webapp", "user:bob", "project:webapp",
			"UI Designer", 50),
		createProjectEdge("assigned:charlie:mobile", "user:charlie", "project:mobile",
			"Project Manager", 100),

		// Company-project relationships
		createOwnershipEdge("owns:techcorp:webapp", "company:techcorp", "project:webapp", "2023-01-01"),
		createOwnershipEdge("owns:designco:mobile", "company:designco", "project:mobile", "2023-08-01"),
	}

	fmt.Println("   Creating edges:")
	for _, edge := range edges {
		if err := gs.AddEdge(edge); err != nil {
			log.Printf("   ❌ Failed to add %s: %v", edge.ID, err)
		} else {
			fmt.Printf("   ✅ %s: %s → %s (%s)\n",
				edge.Type, edge.FromVertex, edge.ToVertex, edge.ID)
		}
	}

	// Demonstrate edge retrieval
	fmt.Println("\n   Retrieving specific edges:")
	testEdgeIDs := []string{"emp:alice:techcorp", "follows:alice:bob", "assigned:alice:webapp"}
	for _, id := range testEdgeIDs {
		if edge, err := gs.GetEdge(id); err != nil {
			log.Printf("   ❌ Failed to get %s: %v", id, err)
		} else {
			fmt.Printf("   📋 %s: %s → %s (weight: %.2f)\n",
				edge.ID, edge.FromVertex, edge.ToVertex, edge.Weight)
		}
	}

	// Demonstrate edge existence checking
	fmt.Println("\n   Checking edge existence:")
	checkEdgeIDs := []string{"follows:alice:bob", "nonexistent:edge", "friend:alice:charlie"}
	for _, id := range checkEdgeIDs {
		if exists, err := gs.EdgeExists(id); err != nil {
			log.Printf("   ❌ Error checking %s: %v", id, err)
		} else if exists {
			fmt.Printf("   ✅ Exists: %s\n", id)
		} else {
			fmt.Printf("   ❌ Missing: %s\n", id)
		}
	}

	fmt.Printf("\n   📊 Created %d edges representing various relationship types\n", len(edges))
}

func demoTypeBasedQueries(gs graph.GraphStore) {
	fmt.Println("   Type-based filtering and queries...")

	// Query vertices by type
	vertexTypes := []string{"User", "Company", "Project"}
	for _, vType := range vertexTypes {
		fmt.Printf("\n   🔍 Querying vertices of type '%s':\n", vType)
		vertices, err := gs.GetVerticesByType(vType, -1)
		if err != nil {
			log.Printf("     ❌ Query failed: %v", err)
			continue
		}

		if len(vertices) == 0 {
			fmt.Printf("     (no vertices found)\n")
		} else {
			fmt.Printf("     Found %d vertices:\n", len(vertices))
			for _, vertex := range vertices {
				fmt.Printf("       • %s: %s\n", vertex.ID, vertex.Properties["name"])
			}
		}
	}

	// Query edges by type
	edgeTypes := []string{"works_at", "follows", "friend", "assigned_to", "owns"}
	for _, eType := range edgeTypes {
		fmt.Printf("\n   🔍 Querying edges of type '%s':\n", eType)
		edges, err := gs.GetEdgesByType(eType, 5) // Limit to 5 for brevity
		if err != nil {
			log.Printf("     ❌ Query failed: %v", err)
			continue
		}

		if len(edges) == 0 {
			fmt.Printf("     (no edges found)\n")
		} else {
			fmt.Printf("     Found %d edges:\n", len(edges))
			for _, edge := range edges {
				fmt.Printf("       • %s: %s → %s\n", edge.ID, edge.FromVertex, edge.ToVertex)
			}
		}
	}

	// Demonstrate comprehensive queries
	fmt.Println("\n   📊 Summary of all graph data:")
	allVertices, err := gs.GetAllVertices(-1)
	if err != nil {
		log.Printf("   ❌ Failed to get all vertices: %v", err)
	} else {
		fmt.Printf("   Total vertices: %d\n", len(allVertices))
	}

	allEdges, err := gs.GetAllEdges(-1)
	if err != nil {
		log.Printf("   ❌ Failed to get all edges: %v", err)
	} else {
		fmt.Printf("   Total edges: %d\n", len(allEdges))
	}
}

func demoBatchGraphOperations(gs graph.GraphStore) {
	fmt.Println("   Efficient batch operations...")

	// Create additional vertices for batch testing
	batchVertices := []*common.Vertex{
		createUser("user:david", "David Wilson", 30, "DevOps"),
		createUser("user:emma", "Emma Davis", 27, "Data Scientist"),
		createUser("user:frank", "Frank Miller", 35, "Architect"),
		createCompany("company:startup", "Cool Startup", "AI", 25),
		createProject("project:ai", "AI Platform", "Active"),
	}

	fmt.Printf("   Adding %d vertices in batch:\n", len(batchVertices))
	if err := gs.BatchAddVertices(batchVertices); err != nil {
		log.Printf("   ❌ Batch vertex add failed: %v", err)
	} else {
		fmt.Printf("   ✅ Successfully added %d vertices\n", len(batchVertices))
		for _, vertex := range batchVertices {
			fmt.Printf("     • %s: %s\n", vertex.ID, vertex.Properties["name"])
		}
	}

	// Create additional edges for batch testing
	batchEdges := []*common.Edge{
		createEmploymentEdge("emp:david:startup", "user:david", "company:startup", "DevOps Lead", "2023-09-01"),
		createEmploymentEdge("emp:emma:startup", "user:emma", "company:startup", "Data Scientist", "2023-09-15"),
		createSocialEdge("follows:david:emma", "follows", "user:david", "user:emma", "2023-10-01"),
		createProjectEdge("assigned:emma:ai", "user:emma", "project:ai", "AI Researcher", 90),
		createOwnershipEdge("owns:startup:ai", "company:startup", "project:ai", "2023-09-01"),
	}

	fmt.Printf("\n   Adding %d edges in batch:\n", len(batchEdges))
	if err := gs.BatchAddEdges(batchEdges); err != nil {
		log.Printf("   ❌ Batch edge add failed: %v", err)
	} else {
		fmt.Printf("   ✅ Successfully added %d edges\n", len(batchEdges))
		for _, edge := range batchEdges {
			fmt.Printf("     • %s: %s → %s\n", edge.Type, edge.FromVertex, edge.ToVertex)
		}
	}
}

func demoBasicTraversal(gs graph.GraphStore) {
	fmt.Println("   Basic graph traversal operations...")

	testVertices := []string{"user:alice", "company:techcorp", "project:webapp"}

	for _, vertexID := range testVertices {
		fmt.Printf("\n   🔄 Traversal from %s:\n", vertexID)

		// Outgoing edges
		outgoing, err := gs.GetOutgoingEdges(vertexID)
		if err != nil {
			log.Printf("     ❌ Failed to get outgoing edges: %v", err)
		} else {
			fmt.Printf("     Outgoing edges (%d):\n", len(outgoing))
			for _, edge := range outgoing {
				fmt.Printf("       → %s (%s)\n", edge.ToVertex, edge.Type)
			}
		}

		// Incoming edges
		incoming, err := gs.GetIncomingEdges(vertexID)
		if err != nil {
			log.Printf("     ❌ Failed to get incoming edges: %v", err)
		} else {
			fmt.Printf("     Incoming edges (%d):\n", len(incoming))
			for _, edge := range incoming {
				fmt.Printf("       ← %s (%s)\n", edge.FromVertex, edge.Type)
			}
		}

		// All neighbors
		neighbors, err := gs.GetNeighbors(vertexID, graph.DirectionBoth)
		if err != nil {
			log.Printf("     ❌ Failed to get neighbors: %v", err)
		} else {
			fmt.Printf("     All neighbors (%d):\n", len(neighbors))
			for _, neighbor := range neighbors {
				fmt.Printf("       ↔ %s (%s)\n", neighbor.ID, neighbor.Properties["name"])
			}
		}
	}
}

func printGraphStatistics(gs graph.GraphStore) {
	fmt.Println("   Comprehensive graph analytics...")

	stats, err := gs.GetStats()
	if err != nil {
		log.Printf("   ❌ Failed to get statistics: %v", err)
		return
	}

	fmt.Printf("   📊 Graph Statistics:\n")
	fmt.Printf("     • Total Vertices: %d\n", stats.VertexCount)
	fmt.Printf("     • Total Edges: %d\n", stats.EdgeCount)
	fmt.Printf("     • Total Size: %d bytes (%.2f MB)\n",
		stats.TotalSize, float64(stats.TotalSize)/1024/1024)
	fmt.Printf("     • Index Count: %d\n", stats.IndexCount)
	fmt.Printf("     • Average Vertex Size: %.2f bytes\n", stats.AvgVertexSize)
	fmt.Printf("     • Average Edge Size: %.2f bytes\n", stats.AvgEdgeSize)
	fmt.Printf("     • Last Access: %s\n", stats.LastAccess.Format(time.RFC3339))

	fmt.Printf("     • Vertex Types: %v\n", stats.VertexTypes)
	fmt.Printf("     • Edge Types: %v\n", stats.EdgeTypes)

	// Calculate graph density and other metrics
	if stats.VertexCount > 1 {
		maxPossibleEdges := stats.VertexCount * (stats.VertexCount - 1)
		density := (float64(stats.EdgeCount) / float64(maxPossibleEdges)) * 100
		fmt.Printf("     • Graph Density: %.2f%%\n", density)

		avgDegree := float64(stats.EdgeCount*2) / float64(stats.VertexCount)
		fmt.Printf("     • Average Degree: %.2f\n", avgDegree)
	}

	if stats.VertexCount > 0 {
		indexEfficiency := float64(stats.IndexCount) / float64(stats.VertexCount+stats.EdgeCount)
		fmt.Printf("     • Index Efficiency: %.2f indices per entity\n", indexEfficiency)
	}
}

// Helper functions to create different types of vertices
func createUser(id, name string, age int, title string) *common.Vertex {
	user := common.NewVertex(id, "User")
	user.Properties["name"] = name
	user.Properties["age"] = age
	user.Properties["title"] = title
	user.Properties["type"] = "person"
	return user
}

func createCompany(id, name, industry string, size int) *common.Vertex {
	company := common.NewVertex(id, "Company")
	company.Properties["name"] = name
	company.Properties["industry"] = industry
	company.Properties["size"] = size
	company.Properties["type"] = "organization"
	return company
}

func createProject(id, name, status string) *common.Vertex {
	project := common.NewVertex(id, "Project")
	project.Properties["name"] = name
	project.Properties["status"] = status
	project.Properties["type"] = "project"
	return project
}

// Helper functions to create different types of edges
func createEmploymentEdge(id, userID, companyID, role, startDate string) *common.Edge {
	edge := common.NewEdge(id, "works_at", userID, companyID)
	edge.Properties["role"] = role
	edge.Properties["start_date"] = startDate
	edge.Properties["relationship"] = "employment"
	edge.Weight = 1.0
	return edge
}

func createSocialEdge(id, edgeType, fromID, toID, since string) *common.Edge {
	edge := common.NewEdge(id, edgeType, fromID, toID)
	edge.Properties["since"] = since
	edge.Properties["relationship"] = "social"
	edge.Weight = 0.8
	return edge
}

func createProjectEdge(id, userID, projectID, role string, allocation int) *common.Edge {
	edge := common.NewEdge(id, "assigned_to", userID, projectID)
	edge.Properties["role"] = role
	edge.Properties["allocation"] = allocation
	edge.Properties["relationship"] = "assignment"
	edge.Weight = float64(allocation) / 100.0
	return edge
}

func createOwnershipEdge(id, companyID, projectID, since string) *common.Edge {
	edge := common.NewEdge(id, "owns", companyID, projectID)
	edge.Properties["since"] = since
	edge.Properties["relationship"] = "ownership"
	edge.Weight = 1.0
	return edge
}
