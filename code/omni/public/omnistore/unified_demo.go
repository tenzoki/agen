package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/godast/godast/internal/common"
	"github.com/godast/godast/internal/graph"
	"github.com/godast/godast/internal/kv"
	"github.com/godast/godast/internal/storage"
	"github.com/godast/godast/pkg/filestore"
)

func main() {
	fmt.Println("🚀 OmniStore Unified Interface - Complete Demo")
	fmt.Println("===============================================")
	fmt.Println("Demonstrating KV Store + Graph Store + File Storage integration")

	// Setup shared storage
	tmpDir := "./demo-data/unified-omnistore"
	os.RemoveAll(tmpDir) // Clean up any previous runs
	defer os.RemoveAll(tmpDir)

	fmt.Println("\n📁 1. Initializing Storage Components...")

	// Create shared BadgerDB storage
	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Initialize KV store
	kvStore := kv.NewKVStore(store)
	fmt.Println("✅ KV Store initialized")

	// Initialize Graph store
	graphStore := graph.NewGraphStore(store)
	defer graphStore.Close()
	fmt.Println("✅ Graph Store initialized")

	// Initialize File store
	fileConfig := filestore.DefaultConfig()
	fileConfig.StorageDir = "files"
	fileConfig.EnableCompression = true
	fileConfig.EnableDeduplication = true
	fileConfig.CompressionLevel = 6

	fileStore, err := filestore.NewFileStore(fileConfig, tmpDir)
	if err != nil {
		log.Fatalf("Failed to create file store: %v", err)
	}
	defer fileStore.Close()
	fmt.Println("✅ File Store initialized")

	fmt.Println("✅ OmniStore components initialized successfully")

	// Demonstrate all components working together
	fmt.Println("\n🎯 Component Integration Demonstrations:")

	// 1. KV Store Operations
	demoKVOperations(kvStore)

	// 2. Graph Store Operations
	demoGraphOperations(graphStore)

	// 3. File Storage Operations (NEW!)
	demoFileStorageOperations(fileStore)

	// 4. Cross-Component Integration
	demoCrossComponentIntegration(kvStore, graphStore, fileStore)

	// 5. Multi-Modal Data Patterns
	demoMultiModalPatterns(kvStore, graphStore, fileStore)

	// 6. Performance and Statistics
	demoStatistics(kvStore, graphStore, fileStore)

	fmt.Println("\n🎉 OmniStore Unified Demo Complete!")
	fmt.Println("All storage paradigms working together seamlessly:")
	fmt.Println("   • 🗄️  Key-Value: Fast lookups and caching")
	fmt.Println("   • 🕸️  Graph: Relationships and traversals")
	fmt.Println("   • 📁 Files: Content-addressable blob storage")
	fmt.Println("   • 🔗 Integration: Cross-component data flows")
}

func demoKVOperations(kvStore kv.KVStore) {
	fmt.Println("\n🗄️  2. Key-Value Store Operations")
	fmt.Println("--------------------------------")

	// Store user profiles
	users := map[string]string{
		"user:alice": `{"name": "Alice Johnson", "email": "alice@example.com", "role": "Engineer", "dept": "R&D"}`,
		"user:bob":   `{"name": "Bob Smith", "email": "bob@example.com", "role": "Designer", "dept": "UX"}`,
		"user:carol": `{"name": "Carol Williams", "email": "carol@example.com", "role": "Manager", "dept": "Engineering"}`,
	}

	fmt.Println("📝 Storing user profiles:")
	for key, value := range users {
		if err := kvStore.Set(key, []byte(value)); err != nil {
			fmt.Printf("❌ Failed to set %s: %v\n", key, err)
		} else {
			fmt.Printf("✅ Stored %s\n", key)
		}
	}

	// Store application configuration
	configs := map[string]string{
		"config:api_endpoint": "https://api.example.com/v2",
		"config:db_host":      "localhost:5432",
		"config:cache_ttl":    "3600",
		"config:max_upload":   "50MB",
	}

	fmt.Println("\n⚙️  Storing application configuration:")
	for key, value := range configs {
		if err := kvStore.Set(key, []byte(value)); err != nil {
			fmt.Printf("❌ Failed to set %s: %v\n", key, err)
		} else {
			fmt.Printf("✅ Set %s = %s\n", key, value)
		}
	}

	// Demonstrate prefix scanning
	fmt.Println("\n🔍 Scanning user profiles:")
	userResults, err := kvStore.Scan("user:", 10)
	if err != nil {
		fmt.Printf("❌ Scan failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d users:\n", len(userResults))
		for key := range userResults {
			fmt.Printf("   • %s\n", key)
		}
	}
}

func demoGraphOperations(graphStore graph.GraphStore) {
	fmt.Println("\n🕸️  3. Graph Store Operations")
	fmt.Println("-----------------------------")

	fmt.Println("🏗️  Building organizational graph:")

	// Create person vertices
	people := []*common.Vertex{
		{
			ID:        "person:alice",
			Type:      "Person",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Properties: map[string]interface{}{
				"name":       "Alice Johnson",
				"email":      "alice@example.com",
				"role":       "Senior Engineer",
				"department": "R&D",
				"level":      "L5",
			},
		},
		{
			ID:        "person:bob",
			Type:      "Person",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Properties: map[string]interface{}{
				"name":       "Bob Smith",
				"email":      "bob@example.com",
				"role":       "UX Designer",
				"department": "Design",
				"level":      "L4",
			},
		},
		{
			ID:        "person:carol",
			Type:      "Person",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Properties: map[string]interface{}{
				"name":       "Carol Williams",
				"email":      "carol@example.com",
				"role":       "Engineering Manager",
				"department": "Engineering",
				"level":      "L6",
			},
		},
	}

	for _, person := range people {
		if err := graphStore.AddVertex(person); err != nil {
			fmt.Printf("❌ Failed to add person %s: %v\n", person.ID, err)
		} else {
			fmt.Printf("✅ Added %s (%s)\n", person.ID, person.Properties["name"])
		}
	}

	// Create department and project vertices
	entities := []*common.Vertex{
		{
			ID:        "dept:engineering",
			Type:      "Department",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Properties: map[string]interface{}{
				"name":      "Engineering",
				"budget":    5000000,
				"headcount": 50,
			},
		},
		{
			ID:        "project:webapp",
			Type:      "Project",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Properties: map[string]interface{}{
				"name":     "Web Application Platform",
				"status":   "active",
				"priority": "high",
				"budget":   500000,
			},
		},
	}

	for _, entity := range entities {
		if err := graphStore.AddVertex(entity); err != nil {
			fmt.Printf("❌ Failed to add %s: %v\n", entity.ID, err)
		} else {
			fmt.Printf("✅ Added %s (%s)\n", entity.ID, entity.Properties["name"])
		}
	}

	// Create relationships
	fmt.Println("\n🔗 Creating relationships:")
	relationships := []*common.Edge{
		{
			ID:         "reports_to:alice-carol",
			Type:       "reports_to",
			FromVertex: "person:alice",
			ToVertex:   "person:carol",
			CreatedAt:  time.Now(),
			Properties: map[string]interface{}{
				"since":        "2022-03-01",
				"relationship": "direct_report",
			},
		},
		{
			ID:         "works_in:alice-dept",
			Type:       "works_in",
			FromVertex: "person:alice",
			ToVertex:   "dept:engineering",
			CreatedAt:  time.Now(),
			Properties: map[string]interface{}{
				"start_date": "2021-06-15",
				"role":       "Senior Engineer",
			},
		},
		{
			ID:         "assigned_to:alice-webapp",
			Type:       "assigned_to",
			FromVertex: "person:alice",
			ToVertex:   "project:webapp",
			CreatedAt:  time.Now(),
			Properties: map[string]interface{}{
				"role":       "Lead Developer",
				"allocation": 80,
				"start_date": "2023-01-01",
			},
		},
		{
			ID:         "manages:carol-dept",
			Type:       "manages",
			FromVertex: "person:carol",
			ToVertex:   "dept:engineering",
			CreatedAt:  time.Now(),
			Properties: map[string]interface{}{
				"start_date": "2020-01-01",
				"level":      "director",
			},
		},
	}

	for _, edge := range relationships {
		if err := graphStore.AddEdge(edge); err != nil {
			fmt.Printf("❌ Failed to add relationship %s: %v\n", edge.ID, err)
		} else {
			fmt.Printf("✅ Added %s (%s)\n", edge.ID, edge.Type)
		}
	}

	// Query the graph
	fmt.Println("\n🔍 Graph queries:")

	// Get all people
	people_results, err := graphStore.GetVerticesByType("Person", 10)
	if err != nil {
		fmt.Printf("❌ Failed to get people: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d people in the organization\n", len(people_results))
	}

	// Get Alice's connections
	aliceEdges, err := graphStore.GetOutgoingEdges("person:alice")
	if err != nil {
		fmt.Printf("❌ Failed to get Alice's connections: %v\n", err)
	} else {
		fmt.Printf("✅ Alice has %d professional relationships:\n", len(aliceEdges))
		for _, edge := range aliceEdges {
			fmt.Printf("   • %s to %s\n", edge.Type, edge.ToVertex)
		}
	}
}

func demoFileStorageOperations(fileStore filestore.FileStore) {
	fmt.Println("\n📁 4. File Storage Operations")
	fmt.Println("-----------------------------")

	fmt.Println("📄 Storing various document types:")

	// Store different types of documents
	documents := []struct {
		content     string
		metadata    map[string]string
		key         string
		description string
	}{
		{
			content: `# Engineering Team Handbook

## Code of Conduct
- Be respectful and inclusive
- Collaborate effectively
- Maintain high technical standards

## Development Practices
- Write clean, documented code
- Perform thorough code reviews
- Use automated testing

## Architecture Principles
- Design for scalability
- Prioritize security
- Optimize for maintainability`,
			metadata: map[string]string{
				"content_type": "text/markdown",
				"author":       "carol@example.com",
				"category":     "handbook",
				"department":   "engineering",
				"version":      "2.1",
				"access_level": "internal",
			},
			key:         "docs/engineering_handbook.md",
			description: "Engineering team documentation",
		},
		{
			content: `{
  "project_name": "WebApp Platform",
  "version": "1.5.0",
  "dependencies": {
    "react": "^18.2.0",
    "typescript": "^5.0.0",
    "express": "^4.18.0",
    "badgerdb": "^4.0.0"
  },
  "build_config": {
    "target": "es2020",
    "module": "commonjs",
    "strict": true
  },
  "deployment": {
    "environment": "production",
    "region": "us-west-2",
    "scaling": "auto"
  }
}`,
			metadata: map[string]string{
				"content_type": "application/json",
				"author":       "alice@example.com",
				"category":     "configuration",
				"project":      "webapp",
				"environment":  "production",
			},
			key:         "config/project_config.json",
			description: "Project configuration file",
		},
		{
			content: strings.Repeat("BINARY TEST DATA WITH REPEATED PATTERNS FOR COMPRESSION TESTING. ", 75),
			metadata: map[string]string{
				"content_type": "application/octet-stream",
				"author":       "system",
				"category":     "test_data",
				"purpose":      "compression_testing",
			},
			key:         "test/compression_sample.bin",
			description: "Binary test data for compression",
		},
	}

	var fileHashes []string

	for i, doc := range documents {
		fmt.Printf("\n📄 Storing document %d: %s\n", i+1, doc.description)

		// Store the document
		hash, err := fileStore.Store([]byte(doc.content), doc.metadata)
		if err != nil {
			fmt.Printf("❌ Failed to store document: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Stored successfully\n")
		fmt.Printf("   📊 Content Hash: %s\n", hash)
		fmt.Printf("   📊 Size: %d bytes\n", len(doc.content))

		// Associate with a human-readable key
		err = fileStore.StoreWithKey(doc.key, []byte(doc.content), doc.metadata)
		if err != nil {
			fmt.Printf("❌ Failed to associate key: %v\n", err)
		} else {
			fmt.Printf("   🔑 Key: %s\n", doc.key)
		}

		fileHashes = append(fileHashes, hash)
	}

	// Demonstrate deduplication
	fmt.Println("\n🔄 Testing content deduplication:")
	duplicateContent := "This exact content will be stored multiple times to test deduplication!"

	hash1, _ := fileStore.Store([]byte(duplicateContent), map[string]string{"version": "v1"})
	hash2, _ := fileStore.Store([]byte(duplicateContent), map[string]string{"version": "v2"})

	if hash1 == hash2 {
		fmt.Println("✅ Deduplication working! Same content produces identical hash")
		fmt.Printf("   📊 Hash: %s\n", hash1)
	} else {
		fmt.Println("❌ Deduplication failed - different hashes for same content")
	}

	// Store with different keys but same content
	fileStore.StoreWithKey("docs/terms_v1.txt", []byte(duplicateContent), map[string]string{"version": "v1"})
	fileStore.StoreWithKey("docs/terms_v2.txt", []byte(duplicateContent), map[string]string{"version": "v2"})

	keys, _ := fileStore.FindByHash(hash1)
	fmt.Printf("   🔑 Keys pointing to same content: %v\n", keys)

	// File retrieval demonstration
	fmt.Println("\n🔍 File retrieval test:")
	if len(fileHashes) > 0 {
		hash := fileHashes[0]
		data, metadata, err := fileStore.Retrieve(hash)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve file: %v\n", err)
		} else {
			fmt.Printf("✅ Retrieved file successfully\n")
			fmt.Printf("   📊 Size: %d bytes\n", len(data))
			fmt.Printf("   📋 Type: %s\n", metadata["content_type"])
			fmt.Printf("   📋 Author: %s\n", metadata["author"])
		}
	}

	// Statistics
	fmt.Println("\n📊 File storage statistics:")
	stats, err := fileStore.Stats()
	if err != nil {
		fmt.Printf("❌ Failed to get stats: %v\n", err)
	} else {
		fmt.Printf("   📁 Total files: %d\n", stats.FileCount)
		fmt.Printf("   💾 Total size: %d bytes (%.2f KB)\n", stats.TotalSize, float64(stats.TotalSize)/1024.0)
		fmt.Printf("   📏 Average size: %.0f bytes\n", stats.AverageSize)
	}

	dedupStats, err := fileStore.GetDeduplicationStats()
	if err != nil {
		fmt.Printf("❌ Failed to get dedup stats: %v\n", err)
	} else {
		fmt.Printf("\n🔄 Deduplication statistics:\n")
		fmt.Printf("   📁 Total references: %d\n", dedupStats.TotalFiles)
		fmt.Printf("   🎯 Unique files: %d\n", dedupStats.UniqueFiles)
		fmt.Printf("   📊 Deduplication rate: %.1f%%\n", dedupStats.DeduplicationRate*100)
	}
}

func demoCrossComponentIntegration(kvStore kv.KVStore, graphStore graph.GraphStore, fileStore filestore.FileStore) {
	fmt.Println("\n🔗 5. Cross-Component Integration")
	fmt.Println("---------------------------------")

	fmt.Println("🔗 Linking data across storage paradigms:")

	// Scenario: Link a document in file storage to a person in the graph,
	// with metadata cached in KV store

	// 1. Store a document
	documentContent := `## Alice's Project Notes

Key achievements this quarter:
- Architected new microservices platform
- Improved API response times by 40%
- Led team code review process

Next quarter goals:
- Implement caching layer
- Optimize database queries
- Mentor junior developers`

	docMetadata := map[string]string{
		"content_type": "text/markdown",
		"author":       "alice@example.com",
		"category":     "project_notes",
		"quarter":      "Q3_2024",
	}

	docHash, err := fileStore.Store([]byte(documentContent), docMetadata)
	if err != nil {
		fmt.Printf("❌ Failed to store document: %v\n", err)
		return
	}
	fmt.Printf("✅ Stored project notes document (hash: %s)\n", docHash[:12]+"...")

	// 2. Create a document vertex in the graph that references the file
	docVertex := &common.Vertex{
		ID:        "document:alice_q3_notes",
		Type:      "Document",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Properties: map[string]interface{}{
			"title":      "Alice's Q3 Project Notes",
			"file_hash":  docHash,
			"author":     "alice@example.com",
			"created_at": time.Now().Format(time.RFC3339),
			"category":   "project_notes",
			"quarter":    "Q3_2024",
		},
	}

	if err := graphStore.AddVertex(docVertex); err != nil {
		fmt.Printf("❌ Failed to create document vertex: %v\n", err)
	} else {
		fmt.Println("✅ Created document vertex in graph")
	}

	// 3. Link the document to the person who authored it
	authorshipEdge := &common.Edge{
		ID:         "authored:alice-q3notes",
		Type:       "authored",
		FromVertex: "person:alice",
		ToVertex:   "document:alice_q3_notes",
		CreatedAt:  time.Now(),
		Properties: map[string]interface{}{
			"created_date": time.Now().Format("2006-01-02"),
			"role":         "author",
		},
	}

	if err := graphStore.AddEdge(authorshipEdge); err != nil {
		fmt.Printf("❌ Failed to create authorship edge: %v\n", err)
	} else {
		fmt.Println("✅ Linked document to author in graph")
	}

	// 4. Cache frequently accessed document metadata in KV store
	cacheKey := fmt.Sprintf("doc_cache:%s", docHash)
	cacheValue := fmt.Sprintf(`{
		"title": "Alice's Q3 Project Notes",
		"author": "alice@example.com",
		"size": %d,
		"created": "%s",
		"access_count": 1
	}`, len(documentContent), time.Now().Format(time.RFC3339))

	if err := kvStore.Set(cacheKey, []byte(cacheValue)); err != nil {
		fmt.Printf("❌ Failed to cache metadata: %v\n", err)
	} else {
		fmt.Println("✅ Cached document metadata in KV store")
	}

	fmt.Println("\n🎯 Cross-component query example:")
	fmt.Println("   1. Query graph for Alice's documents")
	fmt.Println("   2. Retrieve file hashes from graph vertices")
	fmt.Println("   3. Get cached metadata from KV store")
	fmt.Println("   4. Access full content from file storage")

	// Demonstrate the cross-component lookup
	aliceDocEdges, err := graphStore.GetOutgoingEdges("person:alice")
	if err != nil {
		fmt.Printf("❌ Failed to get Alice's connections: %v\n", err)
	} else {
		for _, edge := range aliceDocEdges {
			if edge.Type == "authored" {
				fmt.Printf("✅ Found authored document: %s\n", edge.ToVertex)

				// Get the document vertex to find the file hash
				docVertex, err := graphStore.GetVertex(edge.ToVertex)
				if err == nil && docVertex.Type == "Document" {
					if fileHash, ok := docVertex.Properties["file_hash"].(string); ok {
						// Check KV cache first
						cacheKey := fmt.Sprintf("doc_cache:%s", fileHash)
						if cachedData, err := kvStore.Get(cacheKey); err == nil {
							fmt.Printf("   📋 Cached metadata: %s\n", string(cachedData))
						}
					}
				}
			}
		}
	}
}

func demoMultiModalPatterns(kvStore kv.KVStore, graphStore graph.GraphStore, fileStore filestore.FileStore) {
	fmt.Println("\n🎯 6. Multi-Modal Data Patterns")
	fmt.Println("-------------------------------")

	fmt.Println("🔄 Demonstrating common integration patterns:")

	// Pattern 1: Configuration + Graph + Files
	fmt.Println("\n📋 Pattern 1: Configuration-driven document management")

	// Store configuration in KV
	configKey := "config:document_retention"
	configValue := `{"retention_days": 365, "auto_archive": true, "backup_enabled": true}`
	kvStore.Set(configKey, []byte(configValue))
	fmt.Println("   ✅ Stored document retention policy in KV")

	// Pattern 2: Graph-based file organization
	fmt.Println("\n🗂️  Pattern 2: Graph-based file organization")

	// Create a project hierarchy in the graph
	projectVertex := &common.Vertex{
		ID:        "project:documentation",
		Type:      "Project",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Properties: map[string]interface{}{
			"name":       "Documentation System",
			"status":     "active",
			"file_count": 0,
		},
	}
	graphStore.AddVertex(projectVertex)

	folderVertex := &common.Vertex{
		ID:        "folder:engineering_docs",
		Type:      "Folder",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Properties: map[string]interface{}{
			"name":       "Engineering Documentation",
			"path":       "/docs/engineering/",
			"file_count": 0,
		},
	}
	graphStore.AddVertex(folderVertex)

	// Link folder to project
	projectFolderEdge := &common.Edge{
		ID:         "contains:project-folder",
		Type:       "contains",
		FromVertex: "project:documentation",
		ToVertex:   "folder:engineering_docs",
		CreatedAt:  time.Now(),
		Properties: map[string]interface{}{
			"created": time.Now().Format(time.RFC3339),
		},
	}
	graphStore.AddEdge(projectFolderEdge)
	fmt.Println("   ✅ Created project/folder hierarchy in graph")

	// Pattern 3: Event-driven updates
	fmt.Println("\n⚡ Pattern 3: Event-driven updates across components")

	// Simulate a file update that triggers updates across all components
	updatedContent := "# Updated Engineering Handbook\n\nThis handbook has been updated for Q4 2024..."
	updatedMetadata := map[string]string{
		"content_type": "text/markdown",
		"author":       "carol@example.com",
		"category":     "handbook",
		"version":      "2.2",
		"last_update":  time.Now().Format(time.RFC3339),
	}

	// Update file storage
	newHash, _ := fileStore.Store([]byte(updatedContent), updatedMetadata)
	fmt.Printf("   ✅ Updated file in storage (new hash: %s)\n", newHash[:12]+"...")

	// Update graph with new version
	versionVertex := &common.Vertex{
		ID:        "document:handbook_v2.2",
		Type:      "Document",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Properties: map[string]interface{}{
			"title":      "Engineering Handbook v2.2",
			"file_hash":  newHash,
			"version":    "2.2",
			"updated_at": time.Now().Format(time.RFC3339),
		},
	}
	graphStore.AddVertex(versionVertex)
	fmt.Println("   ✅ Created new version vertex in graph")

	// Update KV cache with latest version info
	latestKey := "latest:engineering_handbook"
	latestValue := fmt.Sprintf(`{"version": "2.2", "hash": "%s", "updated": "%s"}`,
		newHash, time.Now().Format(time.RFC3339))
	kvStore.Set(latestKey, []byte(latestValue))
	fmt.Println("   ✅ Updated latest version info in KV cache")

	fmt.Println("\n🎯 Benefits of multi-modal integration:")
	fmt.Println("   • Fast lookups with KV caching")
	fmt.Println("   • Rich relationships via graph structure")
	fmt.Println("   • Efficient content storage with deduplication")
	fmt.Println("   • Flexible querying across paradigms")
}

func demoStatistics(kvStore kv.KVStore, graphStore graph.GraphStore, fileStore filestore.FileStore) {
	fmt.Println("\n📊 7. Performance and Statistics")
	fmt.Println("--------------------------------")

	// KV Store stats
	fmt.Println("🗄️  KV Store Statistics:")
	fmt.Printf("   📊 KV Store: Operational (stats not implemented)\n")

	// Graph Store stats
	fmt.Println("\n🕸️  Graph Store Statistics:")
	graphStats, err := graphStore.GetStats()
	if err != nil {
		fmt.Printf("❌ Failed to get graph stats: %v\n", err)
	} else {
		fmt.Printf("   🔵 Vertices: %d\n", graphStats.VertexCount)
		fmt.Printf("   🔗 Edges: %d\n", graphStats.EdgeCount)
		fmt.Printf("   💾 Storage size: %d bytes\n", graphStats.TotalSize)
	}

	// File Store stats
	fmt.Println("\n📁 File Store Statistics:")
	fileStats, err := fileStore.Stats()
	if err != nil {
		fmt.Printf("❌ Failed to get file stats: %v\n", err)
	} else {
		fmt.Printf("   📄 Files: %d\n", fileStats.FileCount)
		fmt.Printf("   💾 Total size: %d bytes (%.2f KB)\n", fileStats.TotalSize, float64(fileStats.TotalSize)/1024.0)
		fmt.Printf("   📏 Average size: %.0f bytes\n", fileStats.AverageSize)
	}

	dedupStats, err := fileStore.GetDeduplicationStats()
	if err != nil {
		fmt.Printf("❌ Failed to get dedup stats: %v\n", err)
	} else {
		fmt.Printf("   🔄 Deduplication rate: %.1f%%\n", dedupStats.DeduplicationRate*100)
		fmt.Printf("   💾 Space saved: %d bytes\n", dedupStats.SpaceSaved)
	}

	// Combined statistics
	fmt.Println("\n🎯 Combined OmniStore Statistics:")
	totalSize := int64(0)
	if graphStats != nil {
		totalSize += graphStats.TotalSize
	}
	if fileStats != nil {
		totalSize += fileStats.TotalSize
	}

	fmt.Printf("   💾 Total storage across paradigms: %d bytes (%.2f KB)\n",
		totalSize, float64(totalSize)/1024.0)
	if totalSize > 0 && graphStats != nil && fileStats != nil {
		fmt.Printf("   🎯 Storage distribution: Graph (%d%%) + Files (%d%%)\n",
			int(float64(graphStats.TotalSize)/float64(totalSize)*100),
			int(float64(fileStats.TotalSize)/float64(totalSize)*100))
	}
}

// Helper function for Go versions that don't have min builtin
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
