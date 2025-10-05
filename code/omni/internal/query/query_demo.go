//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/graph"
	"github.com/tenzoki/agen/omni/internal/query"
	"github.com/tenzoki/agen/omni/internal/storage"
)

func main() {
	fmt.Println("🔍 Query Language Complete Demo")
	fmt.Println("==================================================")
	fmt.Println("Gremlin-inspired graph query language for BadgerDB Dual Store")

	// Setup temporary storage
	tmpDir := "/tmp/query-language-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	graphStore := graph.NewGraphStore(store)
	queryExecutor := query.NewQueryExecutor(graphStore)
	parser := query.NewQueryParser()

	fmt.Println("\n🏗️  Setting up demo data...")
	setupDemoData(graphStore)

	// Demo sections
	fmt.Println("\n1. Basic Query Builder API")
	demoQueryBuilder()

	fmt.Println("\n2. Query String Parsing")
	demoQueryParsing(parser)

	fmt.Println("\n3. Vertex Queries")
	demoVertexQueries(parser, queryExecutor)

	fmt.Println("\n4. Edge Queries")
	demoEdgeQueries(parser, queryExecutor)

	fmt.Println("\n5. Graph Traversals")
	demoTraversals(parser, queryExecutor)

	fmt.Println("\n6. Filtering Operations")
	demoFiltering(parser, queryExecutor)

	fmt.Println("\n7. Aggregation Operations")
	demoAggregations(parser, queryExecutor)

	fmt.Println("\n8. Complex Multi-Step Queries")
	demoComplexQueries(parser, queryExecutor)

	fmt.Println("\n9. Performance Demonstration")
	demoPerformance(parser, queryExecutor)

	fmt.Println("\n✅ Query language demo completed successfully!")
}

func setupDemoData(gs graph.GraphStore) {
	// Create a rich social network graph
	users := []*common.Vertex{
		createUser("user:alice", "Alice Johnson", "engineer", "New York", []string{"tech", "music"}),
		createUser("user:bob", "Bob Smith", "designer", "Boston", []string{"design", "art"}),
		createUser("user:charlie", "Charlie Brown", "manager", "Chicago", []string{"management", "sports"}),
		createUser("user:diana", "Diana Prince", "analyst", "Seattle", []string{"data", "science"}),
		createUser("user:eve", "Eve Wilson", "developer", "Austin", []string{"coding", "gaming"}),
	}

	companies := []*common.Vertex{
		createCompany("company:techcorp", "TechCorp Inc", "Technology", 500),
		createCompany("company:designco", "Design Co", "Design", 50),
		createCompany("company:datatech", "DataTech Solutions", "Analytics", 200),
	}

	projects := []*common.Vertex{
		createProject("project:webapp", "Web Application", "active", 1),
		createProject("project:mobile", "Mobile App", "completed", 2),
		createProject("project:analytics", "Analytics Platform", "planning", 3),
	}

	// Add all vertices
	allVertices := append(users, append(companies, projects...)...)
	for _, vertex := range allVertices {
		if err := gs.AddVertex(vertex); err != nil {
			log.Printf("Failed to add vertex %s: %v", vertex.ID, err)
		}
	}

	// Create relationships
	relationships := []*common.Edge{
		// Work relationships
		createEdge("works:alice:techcorp", "works_at", "user:alice", "company:techcorp", map[string]interface{}{"role": "Senior Engineer", "years": 3}),
		createEdge("works:bob:designco", "works_at", "user:bob", "company:designco", map[string]interface{}{"role": "Lead Designer", "years": 2}),
		createEdge("works:charlie:techcorp", "works_at", "user:charlie", "company:techcorp", map[string]interface{}{"role": "Engineering Manager", "years": 5}),
		createEdge("works:diana:datatech", "works_at", "user:diana", "company:datatech", map[string]interface{}{"role": "Data Scientist", "years": 1}),
		createEdge("works:eve:techcorp", "works_at", "user:eve", "company:techcorp", map[string]interface{}{"role": "Full Stack Developer", "years": 2}),

		// Social relationships
		createEdge("follows:alice:bob", "follows", "user:alice", "user:bob", map[string]interface{}{"since": "2023-01-15"}),
		createEdge("follows:alice:diana", "follows", "user:alice", "user:diana", map[string]interface{}{"since": "2023-03-20"}),
		createEdge("follows:bob:charlie", "follows", "user:bob", "user:charlie", map[string]interface{}{"since": "2023-02-10"}),
		createEdge("follows:charlie:eve", "follows", "user:charlie", "user:eve", map[string]interface{}{"since": "2023-01-05"}),
		createEdge("follows:diana:alice", "follows", "user:diana", "user:alice", map[string]interface{}{"since": "2023-04-12"}),

		// Friendship (bidirectional concept)
		createEdge("friend:alice:charlie", "friend", "user:alice", "user:charlie", map[string]interface{}{"closeness": "high", "years": 4}),
		createEdge("friend:bob:diana", "friend", "user:bob", "user:diana", map[string]interface{}{"closeness": "medium", "years": 2}),

		// Project assignments
		createEdge("assigned:alice:webapp", "assigned_to", "user:alice", "project:webapp", map[string]interface{}{"role": "lead", "allocation": 0.8}),
		createEdge("assigned:eve:webapp", "assigned_to", "user:eve", "project:webapp", map[string]interface{}{"role": "developer", "allocation": 1.0}),
		createEdge("assigned:bob:mobile", "assigned_to", "user:bob", "project:mobile", map[string]interface{}{"role": "designer", "allocation": 0.6}),
		createEdge("assigned:diana:analytics", "assigned_to", "user:diana", "project:analytics", map[string]interface{}{"role": "analyst", "allocation": 0.9}),

		// Company ownership of projects
		createEdge("owns:techcorp:webapp", "owns", "company:techcorp", "project:webapp", map[string]interface{}{"investment": 250000}),
		createEdge("owns:designco:mobile", "owns", "company:designco", "project:mobile", map[string]interface{}{"investment": 100000}),
		createEdge("owns:datatech:analytics", "owns", "company:datatech", "project:analytics", map[string]interface{}{"investment": 180000}),
	}

	for _, edge := range relationships {
		if err := gs.AddEdge(edge); err != nil {
			log.Printf("Failed to add edge %s: %v", edge.ID, err)
		}
	}

	fmt.Printf("   ✅ Created %d vertices and %d edges\n", len(allVertices), len(relationships))
}

func createUser(id, name, role, city string, interests []string) *common.Vertex {
	user := common.NewVertex(id, "User")
	user.Properties["name"] = name
	user.Properties["role"] = role
	user.Properties["city"] = city
	user.Properties["interests"] = interests
	return user
}

func createCompany(id, name, industry string, size int) *common.Vertex {
	company := common.NewVertex(id, "Company")
	company.Properties["name"] = name
	company.Properties["industry"] = industry
	company.Properties["size"] = size
	return company
}

func createProject(id, name, status string, priority int) *common.Vertex {
	project := common.NewVertex(id, "Project")
	project.Properties["name"] = name
	project.Properties["status"] = status
	project.Properties["priority"] = priority
	return project
}

func createEdge(id, edgeType, from, to string, props map[string]interface{}) *common.Edge {
	edge := common.NewEdge(id, edgeType, from, to)
	for k, v := range props {
		edge.Properties[k] = v
	}
	return edge
}

func demoQueryBuilder() {
	fmt.Println("   Building queries using fluent API...")

	// Basic query building examples
	queries := []struct {
		description string
		builder     func() *query.QueryBuilder
	}{
		{
			"All vertices",
			func() *query.QueryBuilder { return query.G().V() },
		},
		{
			"All users",
			func() *query.QueryBuilder { return query.G().V().HasLabel("User") },
		},
		{
			"User followers",
			func() *query.QueryBuilder { return query.G().V().HasLabel("User").Out("follows") },
		},
		{
			"Count all edges",
			func() *query.QueryBuilder { return query.G().E().Count() },
		},
		{
			"User names only",
			func() *query.QueryBuilder { return query.G().V().HasLabel("User").Values("name") },
		},
		{
			"Limited results",
			func() *query.QueryBuilder { return query.G().V().HasLabel("Company").Limit(2) },
		},
	}

	for _, q := range queries {
		builder := q.builder()
		fmt.Printf("   📝 %s: %s\n", q.description, builder.String())
	}
}

func demoQueryParsing(parser *query.QueryParser) {
	fmt.Println("   Parsing query strings...")

	queryStrings := []string{
		"g.V()",
		"g.V('user:alice')",
		"g.V().hasLabel('User')",
		"g.V().hasLabel('User').out('follows')",
		"g.V().hasLabel('Company').has('name')",
		"g.E().hasLabel('works_at').count()",
		"g.V().hasLabel('User').values('name').limit(3)",
	}

	for _, queryStr := range queryStrings {
		parsedQuery, err := parser.Parse(queryStr)
		if err != nil {
			fmt.Printf("   ❌ Failed to parse '%s': %v\n", queryStr, err)
		} else {
			fmt.Printf("   ✅ Parsed '%s' → %d steps\n", queryStr, len(parsedQuery.Steps))
		}
	}
}

func demoVertexQueries(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Vertex querying operations...")

	vertexQueries := []struct {
		description string
		queryStr    string
	}{
		{"All vertices", "g.V()"},
		{"Specific user", "g.V('user:alice')"},
		{"All users", "g.V().hasLabel('User')"},
		{"All companies", "g.V().hasLabel('Company')"},
		{"Users with name property", "g.V().hasLabel('User').has('name')"},
	}

	for _, vq := range vertexQueries {
		parsedQuery, err := parser.Parse(vq.queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error for '%s': %v\n", vq.queryStr, err)
			continue
		}

		result, err := executor.Execute(parsedQuery)
		if err != nil {
			fmt.Printf("   ❌ Execution error for '%s': %v\n", vq.queryStr, err)
			continue
		}

		fmt.Printf("   📊 %s: %d vertices\n", vq.description, len(result.Vertices))
	}
}

func demoEdgeQueries(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Edge querying operations...")

	edgeQueries := []struct {
		description string
		queryStr    string
	}{
		{"All edges", "g.E()"},
		{"Work relationships", "g.E().hasLabel('works_at')"},
		{"Social follows", "g.E().hasLabel('follows')"},
		{"Friendships", "g.E().hasLabel('friend')"},
		{"Project assignments", "g.E().hasLabel('assigned_to')"},
	}

	for _, eq := range edgeQueries {
		parsedQuery, err := parser.Parse(eq.queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error for '%s': %v\n", eq.queryStr, err)
			continue
		}

		result, err := executor.Execute(parsedQuery)
		if err != nil {
			fmt.Printf("   ❌ Execution error for '%s': %v\n", eq.queryStr, err)
			continue
		}

		fmt.Printf("   🔗 %s: %d edges\n", eq.description, len(result.Edges))
	}
}

func demoTraversals(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Graph traversal operations...")

	traversalQueries := []struct {
		description string
		queryStr    string
	}{
		{"Alice's outgoing connections", "g.V('user:alice').out()"},
		{"Who Alice follows", "g.V('user:alice').out('follows')"},
		{"Alice's workplace", "g.V('user:alice').out('works_at')"},
		{"Who follows Alice", "g.V('user:alice').in('follows')"},
		{"TechCorp employees", "g.V('company:techcorp').in('works_at')"},
		{"Alice's bidirectional connections", "g.V('user:alice').both()"},
		{"Users following other users", "g.V().hasLabel('User').out('follows')"},
	}

	for _, tq := range traversalQueries {
		parsedQuery, err := parser.Parse(tq.queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error for '%s': %v\n", tq.queryStr, err)
			continue
		}

		result, err := executor.Execute(parsedQuery)
		if err != nil {
			fmt.Printf("   ❌ Execution error for '%s': %v\n", tq.queryStr, err)
			continue
		}

		fmt.Printf("   🔄 %s: %d results\n", tq.description, len(result.Vertices))

		// Show some sample results for interesting queries
		if len(result.Vertices) > 0 && (tq.description == "Alice's outgoing connections" || tq.description == "TechCorp employees") {
			fmt.Printf("       Sample results:\n")
			for i, vertex := range result.Vertices {
				if i >= 3 { // Limit to first 3 results
					fmt.Printf("       ... and %d more\n", len(result.Vertices)-3)
					break
				}
				name := vertex.Properties["name"]
				fmt.Printf("       • %s (%s): %v\n", vertex.ID, vertex.Type, name)
			}
		}
	}
}

func demoFiltering(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Filtering and property-based queries...")

	filterQueries := []struct {
		description string
		queryStr    string
	}{
		{"Users with name property", "g.V().hasLabel('User').has('name')"},
		{"Companies with size property", "g.V().hasLabel('Company').has('size')"},
		{"Projects with active status", "g.V().hasLabel('Project').has('status', 'active')"},
		{"Users in New York", "g.V().hasLabel('User').has('city', 'New York')"},
		{"High priority projects", "g.V().hasLabel('Project').has('priority', 1)"},
	}

	for _, fq := range filterQueries {
		parsedQuery, err := parser.Parse(fq.queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error for '%s': %v\n", fq.queryStr, err)
			continue
		}

		result, err := executor.Execute(parsedQuery)
		if err != nil {
			fmt.Printf("   ❌ Execution error for '%s': %v\n", fq.queryStr, err)
			continue
		}

		fmt.Printf("   🔍 %s: %d matches\n", fq.description, len(result.Vertices))
	}
}

func demoAggregations(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Aggregation and counting operations...")

	aggQueries := []struct {
		description string
		queryStr    string
	}{
		{"Total vertices", "g.V().count()"},
		{"Total edges", "g.E().count()"},
		{"Total users", "g.V().hasLabel('User').count()"},
		{"Total companies", "g.V().hasLabel('Company').count()"},
		{"Work relationships", "g.E().hasLabel('works_at').count()"},
		{"Follow relationships", "g.E().hasLabel('follows').count()"},
		{"Active projects", "g.V().hasLabel('Project').has('status', 'active').count()"},
	}

	for _, aq := range aggQueries {
		parsedQuery, err := parser.Parse(aq.queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error for '%s': %v\n", aq.queryStr, err)
			continue
		}

		result, err := executor.Execute(parsedQuery)
		if err != nil {
			fmt.Printf("   ❌ Execution error for '%s': %v\n", aq.queryStr, err)
			continue
		}

		fmt.Printf("   📊 %s: %d\n", aq.description, result.Count)
	}
}

func demoComplexQueries(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Complex multi-step query operations...")

	complexQueries := []struct {
		description string
		queryStr    string
		showResults bool
	}{
		{
			"Names of people Alice follows",
			"g.V('user:alice').out('follows').values('name')",
			true,
		},
		{
			"Count of TechCorp projects",
			"g.V('company:techcorp').out('owns').count()",
			false,
		},
		{
			"All user names (limited to 3)",
			"g.V().hasLabel('User').limit(3).values('name')",
			true,
		},
		{
			"Companies that employ users",
			"g.V().hasLabel('User').out('works_at').hasLabel('Company')",
			false,
		},
		{
			"People working on active projects",
			"g.V().hasLabel('Project').has('status', 'active').in('assigned_to').hasLabel('User')",
			false,
		},
	}

	for _, cq := range complexQueries {
		parsedQuery, err := parser.Parse(cq.queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error for '%s': %v\n", cq.queryStr, err)
			continue
		}

		result, err := executor.Execute(parsedQuery)
		if err != nil {
			fmt.Printf("   ❌ Execution error for '%s': %v\n", cq.queryStr, err)
			continue
		}

		if result.Count > 0 {
			fmt.Printf("   🎯 %s: %d\n", cq.description, result.Count)
		} else {
			fmt.Printf("   🎯 %s: %d vertices, %d values\n", cq.description, len(result.Vertices), len(result.Values))
		}

		// Show sample values for interesting queries
		if cq.showResults && len(result.Values) > 0 {
			fmt.Printf("       Results: ")
			for i, value := range result.Values {
				if i >= 5 { // Limit display
					fmt.Printf("... and %d more", len(result.Values)-5)
					break
				}
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%v", value)
			}
			fmt.Printf("\n")
		}
	}
}

func demoPerformance(parser *query.QueryParser, executor *query.QueryExecutor) {
	fmt.Println("   Query performance demonstration...")

	performanceQueries := []string{
		"g.V().count()",
		"g.E().count()",
		"g.V().hasLabel('User')",
		"g.V().hasLabel('User').out('follows')",
		"g.V().hasLabel('User').values('name')",
		"g.V('user:alice').out().count()",
	}

	for _, queryStr := range performanceQueries {
		parsedQuery, err := parser.Parse(queryStr)
		if err != nil {
			fmt.Printf("   ❌ Parse error: %v\n", err)
			continue
		}

		// Measure execution time
		start := time.Now()
		result, err := executor.Execute(parsedQuery)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ Execution error: %v\n", err)
			continue
		}

		resultSize := len(result.Vertices) + len(result.Edges) + len(result.Values)
		if result.Count > 0 {
			resultSize = int(result.Count)
		}

		fmt.Printf("   ⚡ '%s': %d results in %v\n", queryStr, resultSize, duration)
	}
}
