package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/godast/godast/internal/common"
	"github.com/godast/godast/internal/graph"
	"github.com/godast/godast/internal/kv"
	"github.com/godast/godast/internal/storage"
)

func main() {
	fmt.Println("🏗️  Complete BadgerDB Dual Store Integration Demo")
	fmt.Println("============================================================")

	// Setup shared storage
	tmpDir := "/tmp/complete-dual-store-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Create both KV and Graph stores using the same underlying storage
	kvStore := kv.NewKVStore(store)
	graphStore := graph.NewGraphStore(store)
	defer graphStore.Close()

	fmt.Println("\n🔄 This demo shows how KV and Graph stores can coexist and complement each other")
	fmt.Println("   using the same BadgerDB instance with namespace isolation.")

	// Run integrated demo sections
	fmt.Println("\n1. Setting up Application Configuration (KV Store)")
	setupApplicationConfig(kvStore)

	fmt.Println("\n2. Building Social Graph (Graph Store)")
	buildSocialGraph(graphStore)

	fmt.Println("\n3. Cross-Store Integration Patterns")
	demonstrateCrossStoreIntegration(kvStore, graphStore)

	fmt.Println("\n4. Performance Comparison")
	performanceComparison(kvStore, graphStore)

	fmt.Println("\n5. Monitoring and Statistics")
	showComprehensiveStatistics(kvStore, graphStore)

	fmt.Println("\n6. Real-World Use Case Simulation")
	simulateRealWorldUsage(kvStore, graphStore)

	fmt.Println("\n✅ Complete integration demo finished successfully!")
}

func setupApplicationConfig(kv kv.KVStore) {
	fmt.Println("   Using KV store for application configuration and caching...")

	// Application configuration
	config := map[string][]byte{
		"app:name":           []byte("SocialNetworkApp"),
		"app:version":        []byte("2.1.4"),
		"app:environment":    []byte("production"),
		"db:max_connections": []byte("100"),
		"db:timeout":         []byte("30s"),
		"cache:ttl":          []byte("3600"),
		"auth:jwt_secret":    []byte("super-secret-key-2023"),
		"feature:dark_mode":  []byte("enabled"),
		"feature:analytics":  []byte("enabled"),
		"rate:api_limit":     []byte("1000"),
	}

	fmt.Println("   Setting application configuration:")
	for key, value := range config {
		if err := kv.Set(key, value); err != nil {
			log.Printf("   ❌ Failed to set config %s: %v", key, err)
		} else {
			fmt.Printf("   ✅ Config: %s = %s\n", key, string(value))
		}
	}

	// User sessions (with TTL)
	sessions := map[string]struct {
		data []byte
		ttl  time.Duration
	}{
		"session:user123": {[]byte("active:alice:2023-12-01"), 1 * time.Hour},
		"session:user456": {[]byte("active:bob:2023-12-01"), 30 * time.Minute},
		"session:user789": {[]byte("active:charlie:2023-12-01"), 2 * time.Hour},
	}

	fmt.Println("\n   Setting user sessions with TTL:")
	for sessionKey, session := range sessions {
		if err := kv.SetWithTTL(sessionKey, session.data, session.ttl); err != nil {
			log.Printf("   ❌ Failed to set session %s: %v", sessionKey, err)
		} else {
			fmt.Printf("   ⏰ Session: %s (TTL: %v)\n", sessionKey, session.ttl)
		}
	}

	// Cache frequently accessed data
	cacheData := map[string][]byte{
		"cache:popular_posts":    []byte("[1,5,12,23,45]"),
		"cache:trending_topics":  []byte("[\"AI\",\"GraphDB\",\"Go\",\"BadgerDB\"]"),
		"cache:user_stats:alice": []byte("{\"posts\":42,\"followers\":156,\"following\":89}"),
		"cache:user_stats:bob":   []byte("{\"posts\":28,\"followers\":234,\"following\":67}"),
	}

	fmt.Println("\n   Setting cache data:")
	if err := kv.BatchSet(cacheData); err != nil {
		log.Printf("   ❌ Cache batch set failed: %v", err)
	} else {
		fmt.Printf("   ✅ Cached %d items for fast access\n", len(cacheData))
	}
}

func buildSocialGraph(gs graph.GraphStore) {
	fmt.Println("   Building social network graph with relationships...")

	// Create diverse user profiles
	users := []*common.Vertex{
		createUserProfile("user:alice", "Alice Chen", "Software Engineer", "San Francisco", "tech", 1250, 890),
		createUserProfile("user:bob", "Bob Rodriguez", "UX Designer", "New York", "design", 890, 1340),
		createUserProfile("user:charlie", "Charlie Kim", "Product Manager", "Austin", "product", 560, 670),
		createUserProfile("user:diana", "Diana Patel", "Data Scientist", "Seattle", "data", 2100, 450),
		createUserProfile("user:eve", "Eve Johnson", "DevOps Engineer", "Denver", "tech", 670, 890),
		createUserProfile("user:frank", "Frank Zhang", "Marketing Lead", "Chicago", "marketing", 340, 1100),
	}

	// Create interest communities
	communities := []*common.Vertex{
		createCommunity("community:tech", "Tech Enthusiasts", "Technology discussions", 15420),
		createCommunity("community:design", "Design Collective", "UI/UX and design", 8930),
		createCommunity("community:startup", "Startup Founders", "Entrepreneurship", 5670),
		createCommunity("community:ai", "AI Researchers", "Artificial Intelligence", 12300),
	}

	// Add all vertices
	allVertices := append(users, communities...)
	if err := gs.BatchAddVertices(allVertices); err != nil {
		log.Printf("   ❌ Failed to add vertices: %v", err)
	} else {
		fmt.Printf("   ✅ Added %d users and %d communities\n", len(users), len(communities))
	}

	// Create rich relationship network
	relationships := []*common.Edge{
		// Following relationships (directional)
		createFollowRelation("follows:alice:bob", "user:alice", "user:bob", 0.8, "2023-01-15"),
		createFollowRelation("follows:bob:charlie", "user:bob", "user:charlie", 0.6, "2023-02-20"),
		createFollowRelation("follows:charlie:diana", "user:charlie", "user:diana", 0.9, "2023-03-10"),
		createFollowRelation("follows:diana:eve", "user:diana", "user:eve", 0.7, "2023-04-05"),
		createFollowRelation("follows:eve:alice", "user:eve", "user:alice", 0.8, "2023-04-15"),
		createFollowRelation("follows:frank:alice", "user:frank", "user:alice", 0.5, "2023-05-01"),

		// Mutual friendships (bidirectional)
		createFriendship("friend:alice:diana", "user:alice", "user:diana", 1.0, "college"),
		createFriendship("friend:diana:alice", "user:diana", "user:alice", 1.0, "college"),
		createFriendship("friend:bob:eve", "user:bob", "user:eve", 0.9, "work"),
		createFriendship("friend:eve:bob", "user:eve", "user:bob", 0.9, "work"),

		// Community memberships
		createMembership("member:alice:tech", "user:alice", "community:tech", 1.0, "2023-01-01", "admin"),
		createMembership("member:diana:ai", "user:diana", "community:ai", 1.0, "2023-01-01", "moderator"),
		createMembership("member:bob:design", "user:bob", "community:design", 0.9, "2023-02-01", "active"),
		createMembership("member:charlie:startup", "user:charlie", "community:startup", 0.8, "2023-03-01", "active"),
		createMembership("member:eve:tech", "user:eve", "community:tech", 0.7, "2023-04-01", "active"),
		createMembership("member:frank:startup", "user:frank", "community:startup", 0.6, "2023-05-01", "active"),

		// Cross-community connections
		createMembership("member:alice:ai", "user:alice", "community:ai", 0.6, "2023-06-01", "lurker"),
		createMembership("member:bob:tech", "user:bob", "community:tech", 0.5, "2023-07-01", "lurker"),
	}

	if err := gs.BatchAddEdges(relationships); err != nil {
		log.Printf("   ❌ Failed to add relationships: %v", err)
	} else {
		fmt.Printf("   ✅ Created %d relationships (follows, friends, memberships)\n", len(relationships))
	}

	fmt.Println("   Social graph structure:")
	fmt.Println("     • User profiles with rich metadata")
	fmt.Println("     • Interest-based communities")
	fmt.Println("     • Multiple relationship types")
	fmt.Println("     • Weighted connections for engagement strength")
}

func demonstrateCrossStoreIntegration(kv kv.KVStore, gs graph.GraphStore) {
	fmt.Println("   Showing how KV and Graph stores complement each other...")

	// Use case 1: Graph-derived data cached in KV store
	fmt.Println("\n   📊 Use Case 1: Caching graph analytics in KV store")

	// Calculate user influence and cache it
	users, err := gs.GetVerticesByType("User", -1)
	if err != nil {
		log.Printf("   ❌ Failed to get users: %v", err)
	} else {
		fmt.Println("   Computing user influence scores...")
		for _, user := range users {
			// Calculate influence based on followers
			incoming, err := gs.GetIncomingEdges(user.ID)
			if err != nil {
				continue
			}

			followerCount := 0
			totalEngagement := 0.0
			for _, edge := range incoming {
				if edge.Type == "follows" {
					followerCount++
					totalEngagement += edge.Weight
				}
			}

			influence := float64(followerCount) * (totalEngagement / float64(max(followerCount, 1)))
			influenceData := fmt.Sprintf("%.2f", influence)

			cacheKey := fmt.Sprintf("analytics:influence:%s", user.ID)
			if err := kv.Set(cacheKey, []byte(influenceData)); err != nil {
				log.Printf("   ❌ Failed to cache influence for %s: %v", user.ID, err)
			} else {
				fmt.Printf("   📈 Cached influence for %s: %s\n",
					user.Properties["name"], influenceData)
			}
		}
	}

	// Use case 2: KV store for session management, Graph for social features
	fmt.Println("\n   🔐 Use Case 2: Session + Social recommendations")

	// Get active session
	sessionData, err := kv.Get("session:user123")
	if err != nil {
		log.Printf("   ❌ Failed to get session: %v", err)
	} else {
		fmt.Printf("   ✅ Active session: %s\n", string(sessionData))

		// Use session to get user and their social connections
		userID := "user:alice" // Extracted from session
		recommendations, err := gs.GetNeighbors(userID, graph.DirectionOutgoing)
		if err != nil {
			log.Printf("   ❌ Failed to get recommendations: %v", err)
		} else {
			fmt.Printf("   🤝 Social recommendations for Alice (%d):\n", len(recommendations))
			for _, rec := range recommendations[:min(3, len(recommendations))] {
				fmt.Printf("     • %s (%s)\n", rec.Properties["name"], rec.Properties["profession"])
			}
		}
	}

	// Use case 3: Feature flags from KV + Graph-based personalization
	fmt.Println("\n   🎛️  Use Case 3: Feature flags + Personalized experience")

	darkModeEnabled, err := kv.Get("feature:dark_mode")
	if err == nil && string(darkModeEnabled) == "enabled" {
		fmt.Println("   🌙 Dark mode is enabled globally")

		// Use graph to personalize based on user's community memberships
		userCommunities, err := gs.GetOutgoingEdges("user:alice")
		if err == nil {
			fmt.Println("   🎨 Personalizing UI based on community interests:")
			for _, edge := range userCommunities {
				if edge.Type == "member_of" {
					if community, err := gs.GetVertex(edge.ToVertex); err == nil {
						fmt.Printf("     • Theme variation for %s community\n",
							community.Properties["name"])
					}
				}
			}
		}
	}

	// Use case 4: Cross-store data consistency
	fmt.Println("\n   ⚖️  Use Case 4: Data consistency across stores")

	// Update user profile in graph and invalidate related cache
	user, err := gs.GetVertex("user:alice")
	if err == nil {
		fmt.Println("   Updating user profile...")
		user.Properties["last_active"] = time.Now().Format(time.RFC3339)
		if err := gs.UpdateVertex(user); err != nil {
			log.Printf("   ❌ Failed to update user: %v", err)
		} else {
			fmt.Println("   ✅ Updated user profile in graph store")

			// Invalidate related cache entries
			cacheKeys := []string{
				"cache:user_stats:alice",
				"analytics:influence:user:alice",
			}

			for _, key := range cacheKeys {
				if err := kv.Delete(key); err != nil {
					log.Printf("   ❌ Failed to invalidate cache %s: %v", key, err)
				} else {
					fmt.Printf("   🗑️  Invalidated cache: %s\n", key)
				}
			}
		}
	}
}

func performanceComparison(kv kv.KVStore, gs graph.GraphStore) {
	fmt.Println("   Comparing performance characteristics...")

	// KV Store performance test
	fmt.Println("\n   ⚡ KV Store Performance:")

	kvTestData := make(map[string][]byte)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("perf:kv:item:%04d", i)
		value := fmt.Sprintf("KV performance test data item #%d", i)
		kvTestData[key] = []byte(value)
	}

	start := time.Now()
	if err := kv.BatchSet(kvTestData); err != nil {
		log.Printf("   ❌ KV batch set failed: %v", err)
	} else {
		duration := time.Since(start)
		rate := float64(len(kvTestData)) / duration.Seconds()
		fmt.Printf("   📈 Batch set %d items: %v (%.0f ops/sec)\n",
			len(kvTestData), duration, rate)
	}

	// Graph Store performance test
	fmt.Println("\n   🌐 Graph Store Performance:")

	graphTestVertices := make([]*common.Vertex, 100)
	for i := 0; i < 100; i++ {
		vertex := common.NewVertex(fmt.Sprintf("perf:vertex:%04d", i), "TestNode")
		vertex.Properties["index"] = i
		vertex.Properties["data"] = fmt.Sprintf("Graph test data #%d", i)
		graphTestVertices[i] = vertex
	}

	start = time.Now()
	if err := gs.BatchAddVertices(graphTestVertices); err != nil {
		log.Printf("   ❌ Graph batch add failed: %v", err)
	} else {
		duration := time.Since(start)
		rate := float64(len(graphTestVertices)) / duration.Seconds()
		fmt.Printf("   📈 Batch add %d vertices: %v (%.0f ops/sec)\n",
			len(graphTestVertices), duration, rate)
	}

	// Traversal performance
	start = time.Now()
	count := 0
	err := gs.TraverseBFS("user:alice", graph.DirectionBoth, 3, func(vertex *common.Vertex, depth int) bool {
		count++
		return true
	})
	if err == nil {
		duration := time.Since(start)
		fmt.Printf("   📈 BFS traversal (%d vertices): %v\n", count, duration)
	}

	// Range query performance
	start = time.Now()
	results, err := kv.Scan("perf:kv:", -1)
	if err == nil {
		duration := time.Since(start)
		fmt.Printf("   📈 KV range scan (%d items): %v\n", len(results), duration)
	}
}

func showComprehensiveStatistics(kv kv.KVStore, gs graph.GraphStore) {
	fmt.Println("   Comprehensive system statistics...")

	// KV Store statistics
	fmt.Println("\n   📊 KV Store Statistics:")
	kvStats, err := kv.Stats()
	if err != nil {
		log.Printf("   ❌ Failed to get KV stats: %v", err)
	} else {
		fmt.Printf("     • Keys: %d\n", kvStats.KeyCount)
		fmt.Printf("     • Total Size: %d bytes (%.2f MB)\n",
			kvStats.TotalSize, float64(kvStats.TotalSize)/1024/1024)
		fmt.Printf("     • Avg Key Size: %.2f bytes\n", kvStats.AvgKeySize)
		fmt.Printf("     • Avg Value Size: %.2f bytes\n", kvStats.AvgValueSize)
		fmt.Printf("     • Namespace: %s\n", kvStats.Namespace)
	}

	// Graph Store statistics
	fmt.Println("\n   📊 Graph Store Statistics:")
	graphStats, err := gs.GetStats()
	if err != nil {
		log.Printf("   ❌ Failed to get graph stats: %v", err)
	} else {
		fmt.Printf("     • Vertices: %d\n", graphStats.VertexCount)
		fmt.Printf("     • Edges: %d\n", graphStats.EdgeCount)
		fmt.Printf("     • Total Size: %d bytes (%.2f MB)\n",
			graphStats.TotalSize, float64(graphStats.TotalSize)/1024/1024)
		fmt.Printf("     • Index Count: %d\n", graphStats.IndexCount)
		fmt.Printf("     • Avg Vertex Size: %.2f bytes\n", graphStats.AvgVertexSize)
		fmt.Printf("     • Avg Edge Size: %.2f bytes\n", graphStats.AvgEdgeSize)
		fmt.Printf("     • Vertex Types: %v\n", graphStats.VertexTypes)
		fmt.Printf("     • Edge Types: %v\n", graphStats.EdgeTypes)
	}

	// Combined statistics
	fmt.Println("\n   📊 Combined System Statistics:")
	totalSize := kvStats.TotalSize + graphStats.TotalSize
	totalItems := kvStats.KeyCount + graphStats.VertexCount + graphStats.EdgeCount
	fmt.Printf("     • Total Storage: %.2f MB\n", float64(totalSize)/1024/1024)
	fmt.Printf("     • Total Items: %d\n", totalItems)
	fmt.Printf("     • KV Store Ratio: %.1f%% of total size\n",
		float64(kvStats.TotalSize)/float64(totalSize)*100)
	fmt.Printf("     • Graph Store Ratio: %.1f%% of total size\n",
		float64(graphStats.TotalSize)/float64(totalSize)*100)
	fmt.Printf("     • Index Efficiency: %.2f indices per graph entity\n",
		float64(graphStats.IndexCount)/float64(graphStats.VertexCount+graphStats.EdgeCount))
}

func simulateRealWorldUsage(kv kv.KVStore, gs graph.GraphStore) {
	fmt.Println("   Simulating real-world application scenarios...")

	// Scenario 1: User login and recommendations
	fmt.Println("\n   🔐 Scenario 1: User Login & Recommendations")

	userID := "user:alice"
	sessionID := "session:user123"

	// Check session
	sessionData, err := kv.Get(sessionID)
	if err == nil {
		fmt.Printf("   ✅ Valid session found: %s\n", string(sessionData))

		// Get user profile
		user, err := gs.GetVertex(userID)
		if err == nil {
			fmt.Printf("   👤 User: %s (%s)\n", user.Properties["name"], user.Properties["profession"])

			// Get personalized recommendations based on social graph
			friends, err := gs.GetNeighbors(userID, graph.DirectionBoth)
			if err == nil {
				fmt.Printf("   🤝 Social connections: %d\n", len(friends))

				// Cache recommendations
				recList := ""
				for i, friend := range friends[:min(3, len(friends))] {
					if i > 0 {
						recList += ","
					}
					recList += friend.ID
				}

				cacheKey := fmt.Sprintf("cache:recommendations:%s", userID)
				if err := kv.SetWithTTL(cacheKey, []byte(recList), 1*time.Hour); err == nil {
					fmt.Printf("   💾 Cached recommendations: %s\n", recList)
				}
			}
		}
	}

	// Scenario 2: Content discovery through communities
	fmt.Println("\n   🔍 Scenario 2: Content Discovery")

	// Find communities user is part of
	memberships, err := gs.GetOutgoingEdges(userID)
	if err == nil {
		fmt.Printf("   Finding content based on community memberships...\n")
		communityTypes := make(map[string]bool)

		for _, edge := range memberships {
			if edge.Type == "member_of" {
				community, err := gs.GetVertex(edge.ToVertex)
				if err == nil {
					communityName := community.Properties["name"].(string)
					communityTypes[communityName] = true
					fmt.Printf("   📚 Member of: %s\n", communityName)
				}
			}
		}

		// Cache user interests for content personalization
		interests := ""
		for community := range communityTypes {
			if interests != "" {
				interests += ","
			}
			interests += community
		}

		interestKey := fmt.Sprintf("cache:interests:%s", userID)
		if err := kv.SetWithTTL(interestKey, []byte(interests), 24*time.Hour); err == nil {
			fmt.Printf("   💡 Cached user interests: %s\n", interests)
		}
	}

	// Scenario 3: Real-time notifications
	fmt.Println("\n   🔔 Scenario 3: Real-time Notifications")

	// Simulate new follower
	newFollowerEdge := createFollowRelation("follows:frank:alice", "user:frank", userID, 0.6, time.Now().Format("2006-01-02"))
	if err := gs.AddEdge(newFollowerEdge); err == nil {
		fmt.Printf("   ✅ New follower relationship created\n")

		// Update notification queue in KV store
		notificationKey := fmt.Sprintf("notifications:%s", userID)
		notification := fmt.Sprintf("New follower: user:frank at %s", time.Now().Format(time.RFC3339))

		if err := kv.SetWithTTL(notificationKey, []byte(notification), 7*24*time.Hour); err == nil {
			fmt.Printf("   🔔 Notification queued: %s\n", notification)
		}

		// Invalidate cached follower count
		followerCacheKey := fmt.Sprintf("cache:followers:%s", userID)
		kv.Delete(followerCacheKey)
		fmt.Printf("   🗑️  Invalidated follower count cache\n")
	}

	fmt.Println("\n   🎯 Real-world simulation demonstrates:")
	fmt.Println("     • Session management with KV store")
	fmt.Println("     • Social graph analysis for recommendations")
	fmt.Println("     • Community-based content discovery")
	fmt.Println("     • Real-time notification handling")
	fmt.Println("     • Cache invalidation strategies")
	fmt.Println("     • Cross-store data consistency")
}

// Helper functions
func createUserProfile(id, name, profession, location, industry string, followers, following int) *common.Vertex {
	user := common.NewVertex(id, "User")
	user.Properties["name"] = name
	user.Properties["profession"] = profession
	user.Properties["location"] = location
	user.Properties["industry"] = industry
	user.Properties["followers"] = followers
	user.Properties["following"] = following
	user.Properties["verified"] = followers > 1000
	return user
}

func createCommunity(id, name, description string, members int) *common.Vertex {
	community := common.NewVertex(id, "Community")
	community.Properties["name"] = name
	community.Properties["description"] = description
	community.Properties["members"] = members
	community.Properties["type"] = "interest_group"
	return community
}

func createFollowRelation(id, fromID, toID string, strength float64, since string) *common.Edge {
	edge := common.NewEdge(id, "follows", fromID, toID)
	edge.Weight = strength
	edge.Properties["since"] = since
	edge.Properties["type"] = "social"
	return edge
}

func createFriendship(id, fromID, toID string, strength float64, context string) *common.Edge {
	edge := common.NewEdge(id, "friend", fromID, toID)
	edge.Weight = strength
	edge.Properties["context"] = context
	edge.Properties["type"] = "social"
	return edge
}

func createMembership(id, userID, communityID string, engagement float64, joined string, role string) *common.Edge {
	edge := common.NewEdge(id, "member_of", userID, communityID)
	edge.Weight = engagement
	edge.Properties["joined"] = joined
	edge.Properties["role"] = role
	edge.Properties["type"] = "membership"
	return edge
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
