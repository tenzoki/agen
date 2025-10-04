package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/agen/omni/internal/common"
	"github.com/agen/omni/internal/graph"
	"github.com/agen/omni/internal/storage"
)

func main() {
	fmt.Println("🌐 Advanced Graph Traversal Demo")
	fmt.Println("==================================================")

	// Setup temporary storage
	tmpDir := "/tmp/graph-traversal-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	graphStore := graph.NewGraphStore(store)
	defer graphStore.Close()

	// Create comprehensive test graph
	fmt.Println("\n1. Building Test Graph")
	buildTestGraph(graphStore)

	fmt.Println("\n2. Basic Traversal Operations")
	demoBasicTraversal(graphStore)

	fmt.Println("\n3. Breadth-First Search (BFS)")
	demoBFSTraversal(graphStore)

	fmt.Println("\n4. Depth-First Search (DFS)")
	demoDFSTraversal(graphStore)

	fmt.Println("\n5. Path Finding")
	demoPathFinding(graphStore)

	fmt.Println("\n6. Advanced Traversal Patterns")
	demoAdvancedTraversal(graphStore)

	fmt.Println("\n✅ Advanced graph traversal demo completed!")
}

func buildTestGraph(gs graph.GraphStore) {
	fmt.Println("   Creating a complex social network graph...")

	// Create users (vertices)
	users := []*common.Vertex{
		createPerson("person:alice", "Alice", "Engineering", "San Francisco"),
		createPerson("person:bob", "Bob", "Design", "New York"),
		createPerson("person:charlie", "Charlie", "Product", "Austin"),
		createPerson("person:diana", "Diana", "Marketing", "Seattle"),
		createPerson("person:eve", "Eve", "Engineering", "San Francisco"),
		createPerson("person:frank", "Frank", "Sales", "Chicago"),
		createPerson("person:grace", "Grace", "Engineering", "San Francisco"),
		createPerson("person:henry", "Henry", "Design", "New York"),
	}

	// Create organizations
	orgs := []*common.Vertex{
		createOrganization("org:techcorp", "TechCorp", "Technology"),
		createOrganization("org:designstudio", "Design Studio", "Creative"),
		createOrganization("org:startup", "Cool Startup", "AI"),
	}

	// Add all vertices
	allVertices := append(users, orgs...)
	for _, vertex := range allVertices {
		if err := gs.AddVertex(vertex); err != nil {
			log.Printf("   ❌ Failed to add vertex %s: %v", vertex.ID, err)
		}
	}

	// Create complex relationship network
	relationships := []*common.Edge{
		// Social connections (follows)
		createConnection("follows:alice:bob", "follows", "person:alice", "person:bob", 0.8),
		createConnection("follows:bob:charlie", "follows", "person:bob", "person:charlie", 0.7),
		createConnection("follows:charlie:diana", "follows", "person:charlie", "person:diana", 0.6),
		createConnection("follows:diana:eve", "follows", "person:diana", "person:eve", 0.9),
		createConnection("follows:eve:frank", "follows", "person:eve", "person:frank", 0.5),
		createConnection("follows:frank:grace", "follows", "person:frank", "person:grace", 0.8),
		createConnection("follows:grace:henry", "follows", "person:grace", "person:henry", 0.7),
		createConnection("follows:henry:alice", "follows", "person:henry", "person:alice", 0.9), // Creates cycle

		// Bidirectional friendships
		createConnection("friend:alice:eve", "friend", "person:alice", "person:eve", 1.0),
		createConnection("friend:eve:alice", "friend", "person:eve", "person:alice", 1.0),
		createConnection("friend:bob:henry", "friend", "person:bob", "person:henry", 0.9),
		createConnection("friend:henry:bob", "friend", "person:henry", "person:bob", 0.9),

		// Work relationships
		createConnection("works:alice:techcorp", "works_at", "person:alice", "org:techcorp", 1.0),
		createConnection("works:eve:techcorp", "works_at", "person:eve", "org:techcorp", 1.0),
		createConnection("works:grace:techcorp", "works_at", "person:grace", "org:techcorp", 1.0),
		createConnection("works:bob:designstudio", "works_at", "person:bob", "org:designstudio", 1.0),
		createConnection("works:henry:designstudio", "works_at", "person:henry", "org:designstudio", 1.0),
		createConnection("works:diana:startup", "works_at", "person:diana", "org:startup", 1.0),
		createConnection("works:frank:startup", "works_at", "person:frank", "org:startup", 1.0),

		// Mentorship (stronger connections)
		createConnection("mentors:grace:alice", "mentors", "person:grace", "person:alice", 1.0),
		createConnection("mentors:charlie:diana", "mentors", "person:charlie", "person:diana", 1.0),
	}

	// Add all edges
	for _, edge := range relationships {
		if err := gs.AddEdge(edge); err != nil {
			log.Printf("   ❌ Failed to add edge %s: %v", edge.ID, err)
		}
	}

	fmt.Printf("   ✅ Created graph with %d vertices and %d edges\n",
		len(allVertices), len(relationships))
	fmt.Println("   Graph structure includes:")
	fmt.Println("     • Social connections (follows, friends)")
	fmt.Println("     • Work relationships (employment)")
	fmt.Println("     • Mentorship connections")
	fmt.Println("     • Cyclic paths for complex traversal")
}

func demoBasicTraversal(gs graph.GraphStore) {
	fmt.Println("   Direction-aware neighbor discovery...")

	testPersons := []string{"person:alice", "person:charlie", "person:grace"}

	for _, personID := range testPersons {
		fmt.Printf("\n   👤 %s connections:\n", personID)

		// Get person info
		person, err := gs.GetVertex(personID)
		if err != nil {
			log.Printf("     ❌ Failed to get person: %v", err)
			continue
		}
		fmt.Printf("     %s (%s, %s)\n",
			person.Properties["name"], person.Properties["department"], person.Properties["location"])

		// Outgoing relationships
		outgoing, err := gs.GetNeighbors(personID, graph.DirectionOutgoing)
		if err != nil {
			log.Printf("     ❌ Failed to get outgoing: %v", err)
		} else {
			fmt.Printf("     Outgoing (%d): ", len(outgoing))
			for i, neighbor := range outgoing {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", getDisplayName(neighbor))
			}
			fmt.Println()
		}

		// Incoming relationships
		incoming, err := gs.GetNeighbors(personID, graph.DirectionIncoming)
		if err != nil {
			log.Printf("     ❌ Failed to get incoming: %v", err)
		} else {
			fmt.Printf("     Incoming (%d): ", len(incoming))
			for i, neighbor := range incoming {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s", getDisplayName(neighbor))
			}
			fmt.Println()
		}

		// All connections
		all, err := gs.GetNeighbors(personID, graph.DirectionBoth)
		if err != nil {
			log.Printf("     ❌ Failed to get all neighbors: %v", err)
		} else {
			fmt.Printf("     Total unique connections: %d\n", len(all))
		}
	}
}

func demoBFSTraversal(gs graph.GraphStore) {
	fmt.Println("   Breadth-First Search exploration...")

	startVertex := "person:alice"
	fmt.Printf("   🔍 BFS from %s:\n", startVertex)

	visited := []string{}
	err := gs.TraverseBFS(startVertex, graph.DirectionOutgoing, 4, func(vertex *common.Vertex, depth int) bool {
		indent := strings.Repeat("  ", depth)
		name := getDisplayName(vertex)
		fmt.Printf("     %s[Depth %d] %s (%s)\n", indent, depth, name, vertex.ID)
		visited = append(visited, vertex.ID)
		return true // Continue traversal
	})

	if err != nil {
		log.Printf("   ❌ BFS failed: %v", err)
	} else {
		fmt.Printf("   ✅ BFS visited %d vertices in breadth-first order\n", len(visited))
	}

	// Demonstrate limited depth BFS
	fmt.Printf("\n   🔍 BFS with depth limit 2:\n")
	limitedVisited := []string{}
	err = gs.TraverseBFS(startVertex, graph.DirectionOutgoing, 2, func(vertex *common.Vertex, depth int) bool {
		indent := strings.Repeat("  ", depth)
		name := getDisplayName(vertex)
		fmt.Printf("     %s[Depth %d] %s\n", indent, depth, name)
		limitedVisited = append(limitedVisited, vertex.ID)
		return true
	})

	if err != nil {
		log.Printf("   ❌ Limited BFS failed: %v", err)
	} else {
		fmt.Printf("   ✅ Limited BFS visited %d vertices (depth ≤ 2)\n", len(limitedVisited))
	}

	// Demonstrate early termination
	fmt.Printf("\n   🔍 BFS with early termination (stop after 5 vertices):\n")
	earlyCount := 0
	err = gs.TraverseBFS(startVertex, graph.DirectionBoth, -1, func(vertex *common.Vertex, depth int) bool {
		earlyCount++
		name := getDisplayName(vertex)
		fmt.Printf("     [%d] %s (depth %d)\n", earlyCount, name, depth)
		return earlyCount < 5 // Stop after 5 vertices
	})

	if err != nil {
		log.Printf("   ❌ Early termination BFS failed: %v", err)
	} else {
		fmt.Printf("   ✅ Early termination BFS stopped after %d vertices\n", earlyCount)
	}
}

func demoDFSTraversal(gs graph.GraphStore) {
	fmt.Println("   Depth-First Search exploration...")

	startVertex := "person:alice"
	fmt.Printf("   🕳️  DFS from %s:\n", startVertex)

	visited := []string{}
	maxDepth := 0
	err := gs.TraverseDFS(startVertex, graph.DirectionOutgoing, 4, func(vertex *common.Vertex, depth int) bool {
		if depth > maxDepth {
			maxDepth = depth
		}
		indent := strings.Repeat("  ", depth)
		name := getDisplayName(vertex)
		fmt.Printf("     %s[Depth %d] %s (%s)\n", indent, depth, name, vertex.ID)
		visited = append(visited, vertex.ID)
		return true // Continue traversal
	})

	if err != nil {
		log.Printf("   ❌ DFS failed: %v", err)
	} else {
		fmt.Printf("   ✅ DFS visited %d vertices, max depth: %d\n", len(visited), maxDepth)
	}

	// Compare BFS vs DFS visiting order
	fmt.Printf("\n   📊 Comparing BFS vs DFS traversal patterns:\n")

	bfsOrder := []string{}
	gs.TraverseBFS(startVertex, graph.DirectionOutgoing, 3, func(vertex *common.Vertex, depth int) bool {
		bfsOrder = append(bfsOrder, getDisplayName(vertex))
		return true
	})

	dfsOrder := []string{}
	gs.TraverseDFS(startVertex, graph.DirectionOutgoing, 3, func(vertex *common.Vertex, depth int) bool {
		dfsOrder = append(dfsOrder, getDisplayName(vertex))
		return true
	})

	fmt.Printf("     BFS order: %s\n", strings.Join(bfsOrder, " → "))
	fmt.Printf("     DFS order: %s\n", strings.Join(dfsOrder, " → "))
}

func demoPathFinding(gs graph.GraphStore) {
	fmt.Println("   Shortest path discovery...")

	pathTests := []struct {
		from, to  string
		direction graph.TraversalDirection
		desc      string
	}{
		{"person:alice", "person:diana", graph.DirectionOutgoing, "Alice to Diana (following chain)"},
		{"person:grace", "person:charlie", graph.DirectionOutgoing, "Grace to Charlie (via mentorship)"},
		{"person:bob", "person:eve", graph.DirectionBoth, "Bob to Eve (any direction)"},
		{"person:alice", "org:techcorp", graph.DirectionBoth, "Alice to TechCorp (work relationship)"},
		{"person:alice", "person:frank", graph.DirectionOutgoing, "Alice to Frank (long path)"},
		{"person:charlie", "person:alice", graph.DirectionOutgoing, "Charlie to Alice (no outgoing path)"},
	}

	for _, test := range pathTests {
		fmt.Printf("\n   🛤️  %s:\n", test.desc)

		path, err := gs.FindPath(test.from, test.to, test.direction, -1)
		if err != nil {
			log.Printf("     ❌ Path finding failed: %v", err)
			continue
		}

		if path == nil {
			fmt.Printf("     ❌ No path found\n")
		} else {
			fmt.Printf("     ✅ Path found (%d steps):\n", len(path))
			for i, vertex := range path {
				name := getDisplayName(vertex)
				if i == 0 {
					fmt.Printf("       🏁 %s", name)
				} else if i == len(path)-1 {
					fmt.Printf(" → 🎯 %s", name)
				} else {
					fmt.Printf(" → %s", name)
				}
			}
			fmt.Println()
		}
	}

	// Demonstrate path finding with depth limits
	fmt.Printf("\n   🛤️  Path finding with depth limits:\n")
	depthLimits := []int{1, 2, 3, -1}

	for _, limit := range depthLimits {
		limitStr := "unlimited"
		if limit > 0 {
			limitStr = fmt.Sprintf("%d", limit)
		}

		path, err := gs.FindPath("person:alice", "person:diana", graph.DirectionOutgoing, limit)
		if err != nil {
			log.Printf("     ❌ Depth %s failed: %v", limitStr, err)
		} else if path == nil {
			fmt.Printf("     ❌ Depth %s: No path found\n", limitStr)
		} else {
			fmt.Printf("     ✅ Depth %s: Path with %d steps\n", limitStr, len(path))
		}
	}
}

func demoAdvancedTraversal(gs graph.GraphStore) {
	fmt.Println("   Advanced traversal patterns and analytics...")

	// Find all people in the same organization
	fmt.Printf("\n   🏢 Finding colleagues (same organization analysis):\n")

	// Get all work relationships
	workEdges, err := gs.GetEdgesByType("works_at", -1)
	if err != nil {
		log.Printf("     ❌ Failed to get work edges: %v", err)
	} else {
		orgEmployees := make(map[string][]string)
		for _, edge := range workEdges {
			orgEmployees[edge.ToVertex] = append(orgEmployees[edge.ToVertex], edge.FromVertex)
		}

		for orgID, employees := range orgEmployees {
			if org, err := gs.GetVertex(orgID); err == nil {
				fmt.Printf("     %s employees:\n", org.Properties["name"])
				for _, empID := range employees {
					if emp, err := gs.GetVertex(empID); err == nil {
						fmt.Printf("       • %s (%s)\n", emp.Properties["name"], emp.Properties["department"])
					}
				}
			}
		}
	}

	// Analyze connection strength
	fmt.Printf("\n   💪 Connection strength analysis:\n")
	analyzeConnections := []string{"person:alice", "person:bob", "person:charlie"}

	for _, personID := range analyzeConnections {
		if person, err := gs.GetVertex(personID); err == nil {
			// Get all outgoing edges to analyze weights
			edges, err := gs.GetOutgoingEdges(personID)
			if err != nil {
				continue
			}

			totalWeight := 0.0
			strongConnections := 0
			for _, edge := range edges {
				totalWeight += edge.Weight
				if edge.Weight > 0.8 {
					strongConnections++
				}
			}

			avgStrength := totalWeight / float64(len(edges))
			fmt.Printf("     %s: %d connections, avg strength %.2f, %d strong (>0.8)\n",
				person.Properties["name"], len(edges), avgStrength, strongConnections)
		}
	}

	// Find mutual connections
	fmt.Printf("\n   🤝 Mutual connection analysis:\n")

	person1, person2 := "person:alice", "person:bob"
	neighbors1, err1 := gs.GetNeighbors(person1, graph.DirectionBoth)
	neighbors2, err2 := gs.GetNeighbors(person2, graph.DirectionBoth)

	if err1 == nil && err2 == nil {
		// Find mutual connections
		connections1 := make(map[string]*common.Vertex)
		for _, neighbor := range neighbors1 {
			connections1[neighbor.ID] = neighbor
		}

		mutualConnections := []*common.Vertex{}
		for _, neighbor := range neighbors2 {
			if _, exists := connections1[neighbor.ID]; exists {
				mutualConnections = append(mutualConnections, neighbor)
			}
		}

		if p1, err := gs.GetVertex(person1); err == nil {
			if p2, err := gs.GetVertex(person2); err == nil {
				fmt.Printf("     %s and %s have %d mutual connections:\n",
					p1.Properties["name"], p2.Properties["name"], len(mutualConnections))
				for _, mutual := range mutualConnections {
					fmt.Printf("       • %s\n", getDisplayName(mutual))
				}
			}
		}
	}

	// Social network analysis: find influencers (most incoming connections)
	fmt.Printf("\n   📈 Influence analysis (most followed people):\n")

	people, err := gs.GetVerticesByType("Person", -1)
	if err != nil {
		log.Printf("     ❌ Failed to get people: %v", err)
	} else {
		influenceMap := make(map[string]int)
		for _, person := range people {
			incoming, err := gs.GetIncomingEdges(person.ID)
			if err == nil {
				// Count only social connections for influence
				socialIncoming := 0
				for _, edge := range incoming {
					if edge.Type == "follows" || edge.Type == "friend" {
						socialIncoming++
					}
				}
				if socialIncoming > 0 {
					influenceMap[person.ID] = socialIncoming
				}
			}
		}

		// Sort and display top influencers
		fmt.Printf("     Top influencers by incoming social connections:\n")
		for personID, count := range influenceMap {
			if person, err := gs.GetVertex(personID); err == nil {
				fmt.Printf("       • %s: %d followers\n", person.Properties["name"], count)
			}
		}
	}
}

// Helper functions
func createPerson(id, name, department, location string) *common.Vertex {
	person := common.NewVertex(id, "Person")
	person.Properties["name"] = name
	person.Properties["department"] = department
	person.Properties["location"] = location
	person.Properties["type"] = "person"
	return person
}

func createOrganization(id, name, industry string) *common.Vertex {
	org := common.NewVertex(id, "Organization")
	org.Properties["name"] = name
	org.Properties["industry"] = industry
	org.Properties["type"] = "organization"
	return org
}

func createConnection(id, edgeType, from, to string, weight float64) *common.Edge {
	edge := common.NewEdge(id, edgeType, from, to)
	edge.Weight = weight
	edge.Properties["created"] = time.Now().Format("2006-01-02")
	return edge
}

func getDisplayName(vertex *common.Vertex) string {
	if name, ok := vertex.Properties["name"].(string); ok {
		return name
	}
	return vertex.ID
}
