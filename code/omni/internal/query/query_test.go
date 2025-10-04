package query

import (
	"os"
	"testing"
	"time"

	"github.com/agen/omni/internal/common"
	"github.com/agen/omni/internal/graph"
	"github.com/agen/omni/internal/storage"
)

// Test helper functions
func setupTestGraphStore(t *testing.T) (graph.GraphStore, func()) {
	tmpDir := "/tmp/query-test-" + time.Now().Format("20060102-150405")
	config := storage.DefaultConfig(tmpDir)

	store, err := storage.NewBadgerStore(config)
	if err != nil {
		t.Fatalf("Failed to create BadgerStore: %v", err)
	}

	graphStore := graph.NewGraphStore(store)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return graphStore, cleanup
}

func createTestData(t *testing.T, gs graph.GraphStore) {
	// Create vertices using the constructor to ensure proper initialization
	users := []*common.Vertex{
		common.NewVertex("user:alice", "User"),
		common.NewVertex("user:bob", "User"),
		common.NewVertex("user:charlie", "User"),
	}

	// Set user properties
	users[0].Properties["name"] = "Alice Johnson"
	users[0].Properties["age"] = 30
	users[0].Properties["city"] = "New York"
	users[0].Properties["interests"] = []interface{}{"tech", "music"}

	users[1].Properties["name"] = "Bob Smith"
	users[1].Properties["age"] = 25
	users[1].Properties["city"] = "Boston"
	users[1].Properties["interests"] = []interface{}{"sports", "tech"}

	users[2].Properties["name"] = "Charlie Brown"
	users[2].Properties["age"] = 35
	users[2].Properties["city"] = "Chicago"
	users[2].Properties["interests"] = []interface{}{"art", "music"}

	companies := []*common.Vertex{
		common.NewVertex("company:techcorp", "Company"),
		common.NewVertex("company:designco", "Company"),
	}

	// Set company properties
	companies[0].Properties["name"] = "TechCorp Inc"
	companies[0].Properties["industry"] = "Technology"
	companies[0].Properties["size"] = 500

	companies[1].Properties["name"] = "Design Co"
	companies[1].Properties["industry"] = "Design"
	companies[1].Properties["size"] = 50

	projects := []*common.Vertex{
		common.NewVertex("project:webapp", "Project"),
		common.NewVertex("project:mobile", "Project"),
	}

	// Set project properties
	projects[0].Properties["name"] = "Web Application"
	projects[0].Properties["status"] = "active"
	projects[0].Properties["priority"] = 1

	projects[1].Properties["name"] = "Mobile App"
	projects[1].Properties["status"] = "completed"
	projects[1].Properties["priority"] = 2

	// Add vertices
	for _, user := range users {
		if err := gs.AddVertex(user); err != nil {
			t.Fatalf("Failed to create user vertex: %v", err)
		}
	}
	for _, company := range companies {
		if err := gs.AddVertex(company); err != nil {
			t.Fatalf("Failed to create company vertex: %v", err)
		}
	}
	for _, project := range projects {
		if err := gs.AddVertex(project); err != nil {
			t.Fatalf("Failed to create project vertex: %v", err)
		}
	}

	// Create edges using the constructor
	edges := []*common.Edge{
		common.NewEdge("works_at:alice:techcorp", "works_at", "user:alice", "company:techcorp"),
		common.NewEdge("works_at:bob:designco", "works_at", "user:bob", "company:designco"),
		common.NewEdge("follows:alice:bob", "follows", "user:alice", "user:bob"),
		common.NewEdge("follows:bob:charlie", "follows", "user:bob", "user:charlie"),
		common.NewEdge("friend:alice:charlie", "friend", "user:alice", "user:charlie"),
		common.NewEdge("assigned:alice:webapp", "assigned_to", "user:alice", "project:webapp"),
		common.NewEdge("assigned:bob:webapp", "assigned_to", "user:bob", "project:webapp"),
		common.NewEdge("owns:techcorp:webapp", "owns", "company:techcorp", "project:webapp"),
	}

	// Set edge properties
	edges[0].Properties["role"] = "Engineer"
	edges[0].Properties["years"] = 3
	edges[1].Properties["role"] = "Designer"
	edges[1].Properties["years"] = 2
	edges[2].Properties["since"] = "2023-01-01"
	edges[3].Properties["since"] = "2023-06-01"
	edges[4].Properties["closeness"] = "high"
	edges[5].Properties["role"] = "lead"
	edges[6].Properties["role"] = "contributor"
	edges[7].Properties["investment"] = 100000

	for _, edge := range edges {
		if err := gs.AddEdge(edge); err != nil {
			t.Fatalf("Failed to create edge: %v", err)
		}
	}
}

// Test Query Builder
func TestQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() *QueryBuilder
		expected string
	}{
		{
			name:     "Empty query",
			builder:  func() *QueryBuilder { return G() },
			expected: "g",
		},
		{
			name:     "V() query",
			builder:  func() *QueryBuilder { return G().V() },
			expected: "g.V()",
		},
		{
			name:     "V() with IDs",
			builder:  func() *QueryBuilder { return G().V("user:alice", "user:bob") },
			expected: "g.V('user:alice', 'user:bob')",
		},
		{
			name:     "V().hasLabel()",
			builder:  func() *QueryBuilder { return G().V().HasLabel("User") },
			expected: "g.V().hasLabel('User')",
		},
		{
			name:     "V().out()",
			builder:  func() *QueryBuilder { return G().V().Out() },
			expected: "g.V().out()",
		},
		{
			name:     "V().out() with labels",
			builder:  func() *QueryBuilder { return G().V().Out("follows", "friend") },
			expected: "g.V().out('follows', 'friend')",
		},
		{
			name:     "Complex query",
			builder:  func() *QueryBuilder { return G().V().HasLabel("User").Out("follows").Count() },
			expected: "g.V().hasLabel('User').out('follows').count()",
		},
		{
			name:     "Has with property only",
			builder:  func() *QueryBuilder { return G().V().Has("name") },
			expected: "g.V().has('name')",
		},
		{
			name:     "Has with property and value",
			builder:  func() *QueryBuilder { return G().V().Has("age", 30) },
			expected: "g.V().has('age', 30)",
		},
		{
			name:     "Values query",
			builder:  func() *QueryBuilder { return G().V().Values("name", "age") },
			expected: "g.V().values('name', 'age')",
		},
		{
			name:     "Limit query",
			builder:  func() *QueryBuilder { return G().V().Limit(10) },
			expected: "g.V().limit(10)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.builder()
			if query.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, query.String())
			}
		})
	}
}

// Test Query Parser
func TestQueryParser(t *testing.T) {
	parser := NewQueryParser()

	tests := []struct {
		name      string
		queryStr  string
		expectErr bool
	}{
		{
			name:      "Empty g",
			queryStr:  "g",
			expectErr: false,
		},
		{
			name:      "Simple V()",
			queryStr:  "g.V()",
			expectErr: false,
		},
		{
			name:      "V() with ID",
			queryStr:  "g.V('user:alice')",
			expectErr: false,
		},
		{
			name:      "Complex query",
			queryStr:  "g.V().hasLabel('User').out('follows').count()",
			expectErr: false,
		},
		{
			name:      "Invalid - no g prefix",
			queryStr:  "V().count()",
			expectErr: true,
		},
		{
			name:      "Has with value",
			queryStr:  "g.V().has('age', 30)",
			expectErr: false,
		},
		{
			name:      "Multiple parameters",
			queryStr:  "g.V('user:alice', 'user:bob').hasLabel('User', 'Admin')",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.Parse(tt.queryStr)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if query == nil {
					t.Errorf("Query should not be nil")
				}
			}
		})
	}
}

// Test Query Execution - Basic Traversals
func TestQueryExecution_BasicTraversals(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name           string
		query          string
		expectVertices int
		expectEdges    int
		expectValues   int
	}{
		{
			name:           "Get all vertices",
			query:          "g.V()",
			expectVertices: 7, // 3 users + 2 companies + 2 projects
			expectEdges:    0,
		},
		{
			name:           "Get all edges",
			query:          "g.E()",
			expectVertices: 0,
			expectEdges:    8, // All test edges
		},
		{
			name:           "Get specific vertex",
			query:          "g.V('user:alice')",
			expectVertices: 1,
			expectEdges:    0,
		},
		{
			name:           "Filter by label",
			query:          "g.V().hasLabel('User')",
			expectVertices: 3, // Alice, Bob, Charlie
			expectEdges:    0,
		},
		{
			name:           "Filter multiple labels",
			query:          "g.V().hasLabel('User', 'Company')",
			expectVertices: 5, // 3 users + 2 companies
			expectEdges:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if len(result.Vertices) != tt.expectVertices {
				t.Errorf("Expected %d vertices, got %d", tt.expectVertices, len(result.Vertices))
			}
			if len(result.Edges) != tt.expectEdges {
				t.Errorf("Expected %d edges, got %d", tt.expectEdges, len(result.Edges))
			}
			if tt.expectValues > 0 && len(result.Values) != tt.expectValues {
				t.Errorf("Expected %d values, got %d", tt.expectValues, len(result.Values))
			}
		})
	}
}

// Test Query Execution - Graph Traversals
func TestQueryExecution_Traversals(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name           string
		query          string
		expectVertices int
		description    string
	}{
		{
			name:           "Alice's outgoing connections",
			query:          "g.V('user:alice').out()",
			expectVertices: 4, // TechCorp, Bob, Charlie, WebApp
			description:    "Alice works at TechCorp, follows Bob, friends with Charlie, assigned to WebApp",
		},
		{
			name:           "Alice follows who",
			query:          "g.V('user:alice').out('follows')",
			expectVertices: 1, // Bob
			description:    "Alice follows Bob",
		},
		{
			name:           "Who works at companies",
			query:          "g.V().hasLabel('User').out('works_at')",
			expectVertices: 2, // TechCorp, DesignCo
			description:    "Alice works at TechCorp, Bob works at DesignCo",
		},
		{
			name:           "Who follows Bob",
			query:          "g.V('user:bob').in('follows')",
			expectVertices: 1, // Alice
			description:    "Alice follows Bob",
		},
		{
			name:           "TechCorp employees",
			query:          "g.V('company:techcorp').in('works_at')",
			expectVertices: 1, // Alice
			description:    "Alice works at TechCorp",
		},
		{
			name:           "Alice's all connections",
			query:          "g.V('user:alice').both()",
			expectVertices: 4, // Bob, Charlie, TechCorp, WebApp
			description:    "Alice's bidirectional connections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if len(result.Vertices) != tt.expectVertices {
				t.Errorf("Expected %d vertices, got %d for query: %s (%s)",
					tt.expectVertices, len(result.Vertices), tt.query, tt.description)

				// Debug output
				t.Logf("Vertices found:")
				for _, v := range result.Vertices {
					t.Logf("  - %s (%s): %v", v.ID, v.Type, v.Properties["name"])
				}
			}
		})
	}
}

// Test Query Execution - Filtering
func TestQueryExecution_Filtering(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name           string
		query          string
		expectVertices int
	}{
		{
			name:           "Users with name property",
			query:          "g.V().hasLabel('User').has('name')",
			expectVertices: 3, // All users have name
		},
		{
			name:           "Users aged 30",
			query:          "g.V().hasLabel('User').has('age', 30)",
			expectVertices: 1, // Alice
		},
		{
			name:           "Users in New York",
			query:          "g.V().hasLabel('User').has('city', 'New York')",
			expectVertices: 1, // Alice
		},
		{
			name:           "Companies with size property",
			query:          "g.V().hasLabel('Company').has('size')",
			expectVertices: 2, // Both companies have size
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if len(result.Vertices) != tt.expectVertices {
				t.Errorf("Expected %d vertices, got %d", tt.expectVertices, len(result.Vertices))
			}
		})
	}
}

// Test Query Execution - Aggregations
func TestQueryExecution_Aggregations(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name        string
		query       string
		expectCount int64
	}{
		{
			name:        "Count all vertices",
			query:       "g.V().count()",
			expectCount: 7,
		},
		{
			name:        "Count all edges",
			query:       "g.E().count()",
			expectCount: 8,
		},
		{
			name:        "Count users",
			query:       "g.V().hasLabel('User').count()",
			expectCount: 3,
		},
		{
			name:        "Count follows relationships",
			query:       "g.E().hasLabel('follows').count()",
			expectCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if result.Count != tt.expectCount {
				t.Errorf("Expected count %d, got %d", tt.expectCount, result.Count)
			}

			if len(result.Values) != 1 {
				t.Errorf("Expected 1 value, got %d", len(result.Values))
			}

			if result.Values[0] != tt.expectCount {
				t.Errorf("Expected value %d, got %v", tt.expectCount, result.Values[0])
			}
		})
	}
}

// Test Query Execution - Value Extraction
func TestQueryExecution_Values(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name         string
		query        string
		expectValues int
	}{
		{
			name:         "Extract all names",
			query:        "g.V().values('name')",
			expectValues: 7, // All vertices have names
		},
		{
			name:         "Extract user ages",
			query:        "g.V().hasLabel('User').values('age')",
			expectValues: 3, // All users have ages
		},
		{
			name:         "Extract multiple properties",
			query:        "g.V().hasLabel('User').values('name', 'age')",
			expectValues: 6, // 3 users × 2 properties
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if len(result.Values) != tt.expectValues {
				t.Errorf("Expected %d values, got %d", tt.expectValues, len(result.Values))
				t.Logf("Values: %v", result.Values)
			}
		})
	}
}

// Test Query Execution - Limit
func TestQueryExecution_Limit(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name           string
		query          string
		expectVertices int
	}{
		{
			name:           "Limit vertices to 3",
			query:          "g.V().limit(3)",
			expectVertices: 3,
		},
		{
			name:           "Limit users to 2",
			query:          "g.V().hasLabel('User').limit(2)",
			expectVertices: 2,
		},
		{
			name:           "Limit more than available",
			query:          "g.V().hasLabel('Project').limit(10)",
			expectVertices: 2, // Only 2 projects available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if len(result.Vertices) != tt.expectVertices {
				t.Errorf("Expected %d vertices, got %d", tt.expectVertices, len(result.Vertices))
			}
		})
	}
}

// Test Complex Queries
func TestQueryExecution_ComplexQueries(t *testing.T) {
	gs, cleanup := setupTestGraphStore(t)
	defer cleanup()

	createTestData(t, gs)
	executor := NewQueryExecutor(gs)

	tests := []struct {
		name           string
		query          string
		expectVertices int
		expectEdges    int
		expectValues   int
		expectCount    int64
		description    string
	}{
		{
			name:           "Users who follow others",
			query:          "g.V().hasLabel('User').out('follows')",
			expectVertices: 2, // Bob, Charlie (followed by Alice, Bob respectively)
			description:    "Find users who are followed by other users",
		},
		{
			name:        "Count users' follows",
			query:       "g.V().hasLabel('User').out('follows').count()",
			expectCount: 2, // Alice->Bob, Bob->Charlie
			description: "Count how many follow relationships exist from users",
		},
		{
			name:         "Company names",
			query:        "g.V().hasLabel('Company').values('name')",
			expectValues: 2, // TechCorp Inc, Design Co
			description:  "Get all company names",
		},
		{
			name:         "Limited user names",
			query:        "g.V().hasLabel('User').limit(2).values('name')",
			expectValues: 2, // First 2 user names
			description:  "Get names of first 2 users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser()
			query, err := parser.Parse(tt.query)
			if err != nil {
				t.Fatalf("Failed to parse query: %v", err)
			}

			result, err := executor.Execute(query)
			if err != nil {
				t.Fatalf("Failed to execute query: %v", err)
			}

			if tt.expectVertices > 0 && len(result.Vertices) != tt.expectVertices {
				t.Errorf("Expected %d vertices, got %d", tt.expectVertices, len(result.Vertices))
			}
			if tt.expectEdges > 0 && len(result.Edges) != tt.expectEdges {
				t.Errorf("Expected %d edges, got %d", tt.expectEdges, len(result.Edges))
			}
			if tt.expectValues > 0 && len(result.Values) != tt.expectValues {
				t.Errorf("Expected %d values, got %d", tt.expectValues, len(result.Values))
			}
			if tt.expectCount > 0 && result.Count != tt.expectCount {
				t.Errorf("Expected count %d, got %d", tt.expectCount, result.Count)
			}
		})
	}
}

// Benchmark tests
func BenchmarkQueryExecution(b *testing.B) {
	gs, cleanup := setupTestGraphStore(&testing.T{})
	defer cleanup()

	createTestData(&testing.T{}, gs)
	executor := NewQueryExecutor(gs)
	parser := NewQueryParser()

	queries := []string{
		"g.V().count()",
		"g.V().hasLabel('User')",
		"g.V().hasLabel('User').out('follows')",
		"g.V().hasLabel('User').values('name')",
	}

	b.ResetTimer()

	for _, queryStr := range queries {
		b.Run(queryStr, func(b *testing.B) {
			query, _ := parser.Parse(queryStr)
			for i := 0; i < b.N; i++ {
				_, err := executor.Execute(query)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
