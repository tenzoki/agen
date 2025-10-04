package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/godast/godast/internal/common"
	"github.com/godast/godast/internal/graph"
	"github.com/godast/godast/internal/kv"
	"github.com/godast/godast/internal/storage"
)

type UserProfile struct {
	ID         string
	Name       string
	Username   string
	Bio        string
	Location   string
	Profession string
	Interests  []string
	JoinDate   string
	Verified   bool
}

type Post struct {
	ID       string
	AuthorID string
	Content  string
	Tags     []string
	Likes    int
	Shares   int
	Posted   string
}

func main() {
	fmt.Println("📱 Social Network Platform Demo")
	fmt.Println("==================================================")
	fmt.Println("Complete social media application simulation using BadgerDB Dual Store")

	// Setup storage
	tmpDir := "/tmp/social-network-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Initialize dual stores
	kvStore := kv.NewKVStore(store)
	graphStore := graph.NewGraphStore(store)
	defer graphStore.Close()

	// Run social network simulation
	fmt.Println("\n🏗️  Building Social Network Platform")
	platform := &SocialNetworkPlatform{
		KV:    kvStore,
		Graph: graphStore,
	}

	platform.Initialize()
	platform.CreateUsers()
	platform.EstablishConnections()
	platform.CreateContent()
	platform.SimulateUserActivity()
	platform.AnalyzeSocialGraph()
	platform.ShowRecommendations()
	platform.SimulateViralContent()
	platform.ShowPlatformStatistics()

	fmt.Println("\n✅ Social network platform demo completed!")
}

type SocialNetworkPlatform struct {
	KV    kv.KVStore
	Graph graph.GraphStore
}

func (p *SocialNetworkPlatform) Initialize() {
	fmt.Println("   Initializing platform configuration...")

	// Platform configuration
	config := map[string][]byte{
		"platform:name":                   []byte("SocialConnect"),
		"platform:version":                []byte("1.0.0"),
		"platform:max_posts_per_day":      []byte("50"),
		"platform:max_follows":            []byte("5000"),
		"platform:verification_threshold": []byte("1000"),
		"features:trending_enabled":       []byte("true"),
		"features:recommendations":        []byte("true"),
		"features:analytics":              []byte("true"),
	}

	if err := p.KV.BatchSet(config); err != nil {
		log.Printf("   ❌ Failed to set platform config: %v", err)
	} else {
		fmt.Printf("   ⚙️  Platform configuration initialized\n")
	}

	// Initialize trending topics cache
	trendingTopics := map[string][]byte{
		"trending:hashtags": []byte("#AI,#GraphDB,#Go,#Tech,#Startup"),
		"trending:keywords": []byte("innovation,technology,networking,community"),
	}

	for key, value := range trendingTopics {
		if err := p.KV.SetWithTTL(key, value, 1*time.Hour); err == nil {
			fmt.Printf("   📈 Cached: %s\n", key)
		}
	}
}

func (p *SocialNetworkPlatform) CreateUsers() {
	fmt.Println("\n   Creating diverse user base...")

	users := []UserProfile{
		{"user:alice", "Alice Chen", "@alice_codes", "Senior Software Engineer passionate about distributed systems", "San Francisco, CA", "Software Engineer", []string{"tech", "AI", "opensource"}, "2023-01-15", true},
		{"user:bob", "Bob Martinez", "@design_bob", "UX Designer creating beautiful and functional experiences", "New York, NY", "UX Designer", []string{"design", "art", "photography"}, "2023-02-20", true},
		{"user:charlie", "Charlie Kim", "@product_guru", "Product Manager building the future of social commerce", "Austin, TX", "Product Manager", []string{"product", "business", "strategy"}, "2023-03-10", false},
		{"user:diana", "Diana Patel", "@data_diana", "Data Scientist exploring patterns in human behavior", "Seattle, WA", "Data Scientist", []string{"data", "AI", "research"}, "2023-04-05", true},
		{"user:eve", "Eve Johnson", "@devops_eve", "DevOps Engineer automating everything", "Denver, CO", "DevOps Engineer", []string{"tech", "automation", "cloud"}, "2023-05-01", false},
		{"user:frank", "Frank Zhang", "@marketing_frank", "Growth marketer helping startups scale", "Chicago, IL", "Marketing Lead", []string{"marketing", "growth", "startup"}, "2023-06-15", false},
		{"user:grace", "Grace Williams", "@grace_creates", "Creative director and digital artist", "Los Angeles, CA", "Creative Director", []string{"art", "design", "creativity"}, "2023-07-01", true},
		{"user:henry", "Henry Brown", "@tech_writer", "Technology journalist covering emerging trends", "Boston, MA", "Tech Journalist", []string{"tech", "writing", "journalism"}, "2023-08-01", false},
		{"user:iris", "Iris Lee", "@startup_iris", "Serial entrepreneur and investor", "Miami, FL", "Entrepreneur", []string{"startup", "investing", "business"}, "2023-09-01", true},
		{"user:jack", "Jack Wilson", "@research_jack", "AI researcher exploring machine consciousness", "Cambridge, MA", "AI Researcher", []string{"AI", "research", "philosophy"}, "2023-10-01", false},
	}

	// Create user vertices
	vertices := make([]*common.Vertex, len(users))
	for i, user := range users {
		vertex := common.NewVertex(user.ID, "User")
		vertex.Properties["name"] = user.Name
		vertex.Properties["username"] = user.Username
		vertex.Properties["bio"] = user.Bio
		vertex.Properties["location"] = user.Location
		vertex.Properties["profession"] = user.Profession
		vertex.Properties["interests"] = user.Interests
		vertex.Properties["join_date"] = user.JoinDate
		vertex.Properties["verified"] = user.Verified
		vertex.Properties["followers_count"] = 0
		vertex.Properties["following_count"] = 0
		vertex.Properties["posts_count"] = 0
		vertices[i] = vertex
	}

	if err := p.Graph.BatchAddVertices(vertices); err != nil {
		log.Printf("   ❌ Failed to create users: %v", err)
	} else {
		fmt.Printf("   👥 Created %d user profiles\n", len(users))
	}

	// Cache user profiles in KV store for fast lookup
	for _, user := range users {
		profileData := fmt.Sprintf("%s|%s|%s|%s", user.Name, user.Username, user.Profession, user.Location)
		cacheKey := fmt.Sprintf("cache:profile:%s", user.ID)
		if err := p.KV.SetWithTTL(cacheKey, []byte(profileData), 6*time.Hour); err == nil {
			fmt.Printf("   💾 Cached profile: %s\n", user.Username)
		}
	}
}

func (p *SocialNetworkPlatform) EstablishConnections() {
	fmt.Println("\n   Building social connections...")

	// Define realistic social connections
	connections := []struct {
		from, to string
		strength float64
		context  string
	}{
		// Tech community connections
		{"user:alice", "user:diana", 0.9, "Same tech conference"},
		{"user:alice", "user:eve", 0.8, "Former colleagues"},
		{"user:diana", "user:jack", 1.0, "AI research collaboration"},
		{"user:eve", "user:alice", 0.8, "Mutual follow"},

		// Design and creative connections
		{"user:bob", "user:grace", 0.9, "Design community"},
		{"user:grace", "user:bob", 0.9, "Mutual admiration"},

		// Business and entrepreneurship
		{"user:charlie", "user:frank", 0.7, "Product marketing collaboration"},
		{"user:frank", "user:iris", 0.8, "Startup ecosystem"},
		{"user:iris", "user:charlie", 0.6, "Product advice"},

		// Cross-domain connections
		{"user:henry", "user:alice", 0.6, "Tech journalism interview"},
		{"user:henry", "user:jack", 0.7, "AI trend coverage"},
		{"user:bob", "user:charlie", 0.5, "Design-product collaboration"},
		{"user:diana", "user:frank", 0.4, "Data-driven marketing"},
		{"user:grace", "user:iris", 0.5, "Creative entrepreneurship"},

		// Popular follows (high-influence users)
		{"user:jack", "user:alice", 0.6, "Follows tech leader"},
		{"user:henry", "user:diana", 0.7, "Expert source"},
		{"user:eve", "user:iris", 0.5, "Startup tech inspiration"},
	}

	// Create follow relationships
	edges := make([]*common.Edge, len(connections))
	for i, conn := range connections {
		edgeID := fmt.Sprintf("follows:%s:%s", conn.from[5:], conn.to[5:]) // Remove "user:" prefix
		edge := common.NewEdge(edgeID, "follows", conn.from, conn.to)
		edge.Weight = conn.strength
		edge.Properties["context"] = conn.context
		edge.Properties["followed_date"] = time.Now().AddDate(0, -rand.Intn(6), -rand.Intn(28)).Format("2006-01-02")
		edges[i] = edge
	}

	if err := p.Graph.BatchAddEdges(edges); err != nil {
		log.Printf("   ❌ Failed to create connections: %v", err)
	} else {
		fmt.Printf("   🤝 Established %d social connections\n", len(connections))
	}

	// Update follower/following counts
	p.updateConnectionCounts()
}

func (p *SocialNetworkPlatform) CreateContent() {
	fmt.Println("\n   Generating user-generated content...")

	posts := []Post{
		{"post:001", "user:alice", "Just shipped our new distributed caching system! The performance improvements are incredible 🚀 #distributedSystems #performance", []string{"tech", "performance"}, 45, 12, "2023-12-01T10:30:00Z"},
		{"post:002", "user:bob", "The psychology behind great UX design: it's not just about making things pretty, it's about understanding human behavior 🧠 #UXDesign #psychology", []string{"design", "psychology"}, 78, 23, "2023-12-01T14:15:00Z"},
		{"post:003", "user:diana", "Machine learning models are only as good as the data they're trained on. Garbage in, garbage out remains true! 📊 #MachineLearning #DataScience", []string{"AI", "data"}, 92, 31, "2023-12-01T16:45:00Z"},
		{"post:004", "user:charlie", "Product-market fit isn't a destination, it's an ongoing journey of understanding your users better every day 📈 #ProductManagement", []string{"product", "strategy"}, 34, 8, "2023-12-02T09:20:00Z"},
		{"post:005", "user:iris", "Raised our Series A! 🎉 Thank you to everyone who believed in our vision. This is just the beginning! #startup #funding", []string{"startup", "milestone"}, 156, 45, "2023-12-02T11:00:00Z"},
		{"post:006", "user:jack", "The future of AI isn't just about smarter algorithms, it's about creating systems that truly understand context and nuance 🤖 #AI #research", []string{"AI", "future"}, 203, 67, "2023-12-02T15:30:00Z"},
		{"post:007", "user:grace", "Color theory in digital design: warm colors advance, cool colors recede. Use this to guide user attention 🎨 #design #colorTheory", []string{"design", "theory"}, 67, 19, "2023-12-03T10:45:00Z"},
		{"post:008", "user:henry", "The rise of graph databases: why relationships matter more than ever in our connected world 🌐 #GraphDB #technology", []string{"tech", "databases"}, 89, 25, "2023-12-03T13:15:00Z"},
	}

	// Store posts in KV store for fast retrieval
	fmt.Println("   📝 Storing posts:")
	for _, post := range posts {
		postData := fmt.Sprintf("%s|%s|%d|%d|%s", post.AuthorID, post.Content, post.Likes, post.Shares, post.Posted)
		postKey := fmt.Sprintf("post:%s", post.ID)

		if err := p.KV.Set(postKey, []byte(postData)); err != nil {
			log.Printf("   ❌ Failed to store post %s: %v", post.ID, err)
		} else {
			// Get author name for display
			if author, err := p.Graph.GetVertex(post.AuthorID); err == nil {
				authorName := author.Properties["name"].(string)
				fmt.Printf("   📄 %s: \"%s...\" (%d likes, %d shares)\n",
					authorName, truncateText(post.Content, 50), post.Likes, post.Shares)
			}
		}

		// Create author-post relationship in graph
		postVertex := common.NewVertex(post.ID, "Post")
		postVertex.Properties["content"] = post.Content
		postVertex.Properties["likes"] = post.Likes
		postVertex.Properties["shares"] = post.Shares
		postVertex.Properties["posted"] = post.Posted
		postVertex.Properties["tags"] = post.Tags

		if err := p.Graph.AddVertex(postVertex); err == nil {
			// Link author to post
			authorPostEdge := common.NewEdge(
				fmt.Sprintf("authored:%s:%s", post.AuthorID, post.ID),
				"authored", post.AuthorID, post.ID)
			authorPostEdge.Properties["posted_date"] = post.Posted
			p.Graph.AddEdge(authorPostEdge)
		}
	}

	// Create hashtag trending data
	hashtags := map[string]int{
		"#AI": 5, "#tech": 4, "#design": 3, "#startup": 2, "#performance": 1,
	}

	trendingData := ""
	for tag, count := range hashtags {
		if trendingData != "" {
			trendingData += ","
		}
		trendingData += fmt.Sprintf("%s:%d", tag, count)
	}

	if err := p.KV.SetWithTTL("trending:hashtags:live", []byte(trendingData), 30*time.Minute); err == nil {
		fmt.Printf("   📈 Updated trending hashtags: %s\n", trendingData)
	}
}

func (p *SocialNetworkPlatform) SimulateUserActivity() {
	fmt.Println("\n   Simulating real-time user activity...")

	// Simulate user sessions
	sessions := []struct {
		userID   string
		duration time.Duration
		activity string
	}{
		{"user:alice", 45 * time.Minute, "active_posting"},
		{"user:bob", 30 * time.Minute, "browsing_feed"},
		{"user:diana", 60 * time.Minute, "engaging_comments"},
		{"user:charlie", 20 * time.Minute, "quick_check"},
	}

	fmt.Println("   👤 Active user sessions:")
	for _, session := range sessions {
		sessionID := fmt.Sprintf("session:%s:%d", session.userID, time.Now().Unix())
		sessionData := fmt.Sprintf("%s|%s|%s", session.userID, session.activity, time.Now().Format(time.RFC3339))

		if err := p.KV.SetWithTTL(sessionID, []byte(sessionData), session.duration); err == nil {
			if user, err := p.Graph.GetVertex(session.userID); err == nil {
				fmt.Printf("   🟢 %s: %s (%v session)\n",
					user.Properties["username"], session.activity, session.duration)
			}
		}
	}

	// Simulate engagement events
	engagements := []struct {
		userID string
		postID string
		action string
	}{
		{"user:bob", "post:001", "like"},
		{"user:diana", "post:001", "share"},
		{"user:charlie", "post:002", "like"},
		{"user:alice", "post:003", "like"},
		{"user:iris", "post:003", "share"},
		{"user:jack", "post:005", "like"},
	}

	fmt.Println("\n   💫 User engagement events:")
	for _, eng := range engagements {
		// Update engagement counters
		engagementKey := fmt.Sprintf("engagement:%s:%s", eng.postID, eng.action)
		// Get current count or default to 0
		if _, err := p.KV.Get(engagementKey); err != nil {
			// Key doesn't exist, initialize to 0
		}

		// Increment counter (simplified)
		newCount := "1" // In real implementation, would parse and increment
		if err := p.KV.Set(engagementKey, []byte(newCount)); err == nil {
			if user, err := p.Graph.GetVertex(eng.userID); err == nil {
				fmt.Printf("   👍 %s %sd %s\n", user.Properties["username"], eng.action, eng.postID)
			}
		}

		// Create engagement relationship in graph
		engagementEdgeID := fmt.Sprintf("%s:%s:%s", eng.action, eng.userID, eng.postID)
		engagementEdge := common.NewEdge(engagementEdgeID, eng.action, eng.userID, eng.postID)
		engagementEdge.Properties["timestamp"] = time.Now().Format(time.RFC3339)
		p.Graph.AddEdge(engagementEdge)
	}
}

func (p *SocialNetworkPlatform) AnalyzeSocialGraph() {
	fmt.Println("\n   🔍 Social graph analysis...")

	// Find most connected users
	fmt.Println("   📊 User influence analysis:")

	users, err := p.Graph.GetVerticesByType("User", -1)
	if err != nil {
		log.Printf("   ❌ Failed to get users: %v", err)
		return
	}

	type UserInfluence struct {
		User      *common.Vertex
		Followers int
		Following int
		Posts     int
	}

	influences := make([]UserInfluence, 0, len(users))

	for _, user := range users {
		// Count followers (incoming follows)
		incoming, err := p.Graph.GetIncomingEdges(user.ID)
		followers := 0
		if err == nil {
			for _, edge := range incoming {
				if edge.Type == "follows" {
					followers++
				}
			}
		}

		// Count following (outgoing follows)
		outgoing, err := p.Graph.GetOutgoingEdges(user.ID)
		following := 0
		posts := 0
		if err == nil {
			for _, edge := range outgoing {
				if edge.Type == "follows" {
					following++
				} else if edge.Type == "authored" {
					posts++
				}
			}
		}

		influences = append(influences, UserInfluence{
			User:      user,
			Followers: followers,
			Following: following,
			Posts:     posts,
		})
	}

	// Sort by follower count
	sort.Slice(influences, func(i, j int) bool {
		return influences[i].Followers > influences[j].Followers
	})

	fmt.Println("   🏆 Top influencers by follower count:")
	for i, inf := range influences[:min(5, len(influences))] {
		username := inf.User.Properties["username"].(string)
		verified := ""
		if inf.User.Properties["verified"].(bool) {
			verified = " ✓"
		}
		fmt.Printf("     %d. %s%s - %d followers, %d following, %d posts\n",
			i+1, username, verified, inf.Followers, inf.Following, inf.Posts)
	}

	// Analyze communities
	fmt.Println("\n   🌐 Community analysis:")

	// Find users in same professions
	professionGroups := make(map[string][]*common.Vertex)
	for _, user := range users {
		profession := user.Properties["profession"].(string)
		professionGroups[profession] = append(professionGroups[profession], user)
	}

	fmt.Println("   Professional communities:")
	for profession, members := range professionGroups {
		if len(members) > 1 {
			fmt.Printf("     %s: %d members\n", profession, len(members))
		}
	}
}

func (p *SocialNetworkPlatform) ShowRecommendations() {
	fmt.Println("\n   🎯 Personalized recommendations...")

	targetUser := "user:alice"
	if user, err := p.Graph.GetVertex(targetUser); err == nil {
		username := user.Properties["username"].(string)
		fmt.Printf("   Recommendations for %s:\n", username)

		// Find friends of friends (2nd degree connections)
		fmt.Println("\n   👥 People you may know:")
		err := p.Graph.TraverseBFS(targetUser, graph.DirectionOutgoing, 2, func(vertex *common.Vertex, depth int) bool {
			if depth == 2 && vertex.Type == "User" && vertex.ID != targetUser {
				// Check if already following
				path, err := p.Graph.FindPath(targetUser, vertex.ID, graph.DirectionOutgoing, 1)
				if err == nil && path == nil { // Not directly connected
					username := vertex.Properties["username"].(string)
					profession := vertex.Properties["profession"].(string)
					fmt.Printf("     • %s (%s) - 2nd degree connection\n", username, profession)
				}
			}
			return true
		})
		if err != nil {
			log.Printf("   ❌ Failed to find recommendations: %v", err)
		}

		// Content recommendations based on interests
		fmt.Println("\n   📄 Content recommendations:")
		interestsInterface := user.Properties["interests"].([]interface{})
		userInterests := make([]string, len(interestsInterface))
		for i, v := range interestsInterface {
			userInterests[i] = v.(string)
		}

		posts, err := p.Graph.GetVerticesByType("Post", -1)
		if err == nil {
			for _, post := range posts[:min(3, len(posts))] {
				if post.Properties["tags"] != nil {
					tagsInterface := post.Properties["tags"].([]interface{})
					postTags := make([]string, len(tagsInterface))
					for i, v := range tagsInterface {
						postTags[i] = v.(string)
					}
					// Simple interest matching
					for _, userInt := range userInterests {
						for _, postTag := range postTags {
							if userInt == postTag {
								if authorID, ok := post.Properties["content"]; ok {
									content := authorID.(string)
									fmt.Printf("     • \"%s...\" #%s\n", truncateText(content, 40), postTag)
								}
								break
							}
						}
					}
				}
			}
		}
	}
}

func (p *SocialNetworkPlatform) SimulateViralContent() {
	fmt.Println("\n   🔥 Viral content simulation...")

	// Simulate a trending post
	viralPost := Post{
		ID:       "post:viral001",
		AuthorID: "user:jack",
		Content:  "BREAKTHROUGH: Our AI model just achieved human-level performance on reasoning tasks! This changes everything 🤯 #AI #breakthrough #research",
		Tags:     []string{"AI", "breakthrough", "research"},
		Likes:    1250,
		Shares:   340,
		Posted:   time.Now().Format(time.RFC3339),
	}

	// Add viral post to graph
	viralVertex := common.NewVertex(viralPost.ID, "Post")
	viralVertex.Properties["content"] = viralPost.Content
	viralVertex.Properties["likes"] = viralPost.Likes
	viralVertex.Properties["shares"] = viralPost.Shares
	viralVertex.Properties["posted"] = viralPost.Posted
	viralVertex.Properties["tags"] = viralPost.Tags
	viralVertex.Properties["viral"] = true

	if err := p.Graph.AddVertex(viralVertex); err == nil {
		// Author relationship
		authorEdge := common.NewEdge("authored:jack:viral001", "authored", viralPost.AuthorID, viralPost.ID)
		p.Graph.AddEdge(authorEdge)

		fmt.Printf("   🚀 Viral post created by @tech_writer:\n")
		fmt.Printf("   📄 \"%s...\"\n", truncateText(viralPost.Content, 80))
		fmt.Printf("   📊 %d likes, %d shares\n", viralPost.Likes, viralPost.Shares)
	}

	// Simulate viral spread through social graph
	fmt.Println("\n   📈 Viral spread simulation:")

	// Get Jack's followers
	followers, err := p.Graph.GetIncomingEdges("user:jack")
	if err == nil {
		fmt.Printf("   🌊 Spreading through %d followers:\n", len(followers))

		for _, followerEdge := range followers {
			if followerEdge.Type == "follows" {
				// Simulate engagement
				engagementEdge := common.NewEdge(
					fmt.Sprintf("viral_like:%s:%s", followerEdge.FromVertex, viralPost.ID),
					"like", followerEdge.FromVertex, viralPost.ID)
				engagementEdge.Properties["viral_boost"] = true
				engagementEdge.Properties["timestamp"] = time.Now().Format(time.RFC3339)

				if err := p.Graph.AddEdge(engagementEdge); err == nil {
					if user, err := p.Graph.GetVertex(followerEdge.FromVertex); err == nil {
						username := user.Properties["username"].(string)
						fmt.Printf("     👍 %s engaged with viral content\n", username)
					}
				}
			}
		}
	}

	// Update trending cache
	trendingKey := "trending:posts:viral"
	trendingData := fmt.Sprintf("%s:%d:%d", viralPost.ID, viralPost.Likes, viralPost.Shares)
	if err := p.KV.SetWithTTL(trendingKey, []byte(trendingData), 2*time.Hour); err == nil {
		fmt.Printf("   🔥 Added to trending cache: %s\n", viralPost.ID)
	}
}

func (p *SocialNetworkPlatform) ShowPlatformStatistics() {
	fmt.Println("\n   📊 Platform statistics...")

	// KV Store stats
	kvStats, err := p.KV.Stats()
	if err == nil {
		fmt.Printf("   KV Store (Cache & Config):\n")
		fmt.Printf("     • Keys: %d\n", kvStats.KeyCount)
		fmt.Printf("     • Size: %.2f MB\n", float64(kvStats.TotalSize)/1024/1024)
	}

	// Graph Store stats
	graphStats, err := p.Graph.GetStats()
	if err == nil {
		fmt.Printf("   Graph Store (Social Network):\n")
		fmt.Printf("     • Vertices: %d\n", graphStats.VertexCount)
		fmt.Printf("     • Edges: %d\n", graphStats.EdgeCount)
		fmt.Printf("     • Size: %.2f MB\n", float64(graphStats.TotalSize)/1024/1024)
		fmt.Printf("     • Indices: %d\n", graphStats.IndexCount)
		fmt.Printf("     • Vertex Types: %v\n", graphStats.VertexTypes)
		fmt.Printf("     • Edge Types: %v\n", graphStats.EdgeTypes)
	}

	// Platform metrics
	fmt.Printf("   Platform Metrics:\n")

	users, err := p.Graph.GetVerticesByType("User", -1)
	if err == nil {
		fmt.Printf("     • Total Users: %d\n", len(users))

		verified := 0
		for _, user := range users {
			if user.Properties["verified"].(bool) {
				verified++
			}
		}
		fmt.Printf("     • Verified Users: %d (%.1f%%)\n", verified, float64(verified)/float64(len(users))*100)
	}

	posts, err := p.Graph.GetVerticesByType("Post", -1)
	if err == nil {
		fmt.Printf("     • Total Posts: %d\n", len(posts))

		totalLikes := 0
		for _, post := range posts {
			if likes, ok := post.Properties["likes"].(int); ok {
				totalLikes += likes
			}
		}
		fmt.Printf("     • Total Likes: %d\n", totalLikes)

		if len(posts) > 0 {
			avgLikes := float64(totalLikes) / float64(len(posts))
			fmt.Printf("     • Average Likes per Post: %.1f\n", avgLikes)
		}
	}

	follows, err := p.Graph.GetEdgesByType("follows", -1)
	if err == nil {
		fmt.Printf("     • Social Connections: %d\n", len(follows))
	}
}

func (p *SocialNetworkPlatform) updateConnectionCounts() {
	// Update follower/following counts for each user
	users, err := p.Graph.GetVerticesByType("User", -1)
	if err != nil {
		return
	}

	for _, user := range users {
		// Count followers
		incoming, _ := p.Graph.GetIncomingEdges(user.ID)
		followers := 0
		for _, edge := range incoming {
			if edge.Type == "follows" {
				followers++
			}
		}

		// Count following
		outgoing, _ := p.Graph.GetOutgoingEdges(user.ID)
		following := 0
		for _, edge := range outgoing {
			if edge.Type == "follows" {
				following++
			}
		}

		// Update user properties
		user.Properties["followers_count"] = followers
		user.Properties["following_count"] = following
		p.Graph.UpdateVertex(user)
	}
}

// Helper functions
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
