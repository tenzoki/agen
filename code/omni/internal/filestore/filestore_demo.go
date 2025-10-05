//go:build ignore

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/tenzoki/agen/omni/internal/filestore"
)

func main() {
	fmt.Println("🗄️  Working File Storage Demo")
	fmt.Println("=============================")

	// Create temporary directory for demo
	tempDir := "./demo_filestore_working"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatal("Failed to create demo directory:", err)
	}
	defer os.RemoveAll(tempDir)

	// Configure file store
	config := filestore.DefaultConfig()
	config.StorageDir = "files"
	config.EnableCompression = true
	config.EnableDeduplication = true

	fmt.Printf("File Store Configuration:\n")
	fmt.Printf("  Storage Directory: %s\n", config.StorageDir)
	fmt.Printf("  Compression: %v (level %d)\n", config.EnableCompression, config.CompressionLevel)
	fmt.Printf("  Deduplication: %v\n", config.EnableDeduplication)
	fmt.Printf("  Max File Size: %d bytes\n", config.MaxFileSize)

	// Create file store
	store, err := filestore.NewFileStore(config, tempDir)
	if err != nil {
		log.Fatal("Failed to create file store:", err)
	}
	defer store.Close()

	fmt.Println("\n1. Basic File Storage Operations")
	fmt.Println("================================")

	// Store some test files
	testFiles := []struct {
		content  string
		metadata map[string]string
		key      string
	}{
		{
			content: "Hello, World! This is a test document with some content.",
			metadata: map[string]string{
				"content_type": "text/plain",
				"author":       "demo_user",
				"category":     "test",
				"description":  "Simple test document",
			},
			key: "hello_world.txt",
		},
		{
			content: strings.Repeat("This text will compress very well! ", 200),
			metadata: map[string]string{
				"content_type": "text/plain",
				"author":       "compression_test",
				"category":     "compression",
				"description":  "Repetitive text for compression testing",
			},
			key: "compression_test.txt",
		},
		{
			content: `{
  "name": "Alice Johnson",
  "age": 30,
  "city": "New York",
  "occupation": "Software Engineer",
  "hobbies": ["reading", "hiking", "coding"],
  "contact": {
    "email": "alice@example.com",
    "phone": "+1-555-0123"
  }
}`,
			metadata: map[string]string{
				"content_type": "application/json",
				"author":       "api_system",
				"category":     "user_data",
				"description":  "User profile JSON data",
			},
			key: "user_alice.json",
		},
	}

	var storedHashes []string

	// Store each file
	for i, file := range testFiles {
		fmt.Printf("\n📄 Storing file %d: %s\n", i+1, file.key)

		// Calculate and display content hash first
		hash := store.GetContentHash([]byte(file.content))
		fmt.Printf("   Content Hash: %s\n", hash)

		// Store the file content
		actualHash, err := store.Store([]byte(file.content), file.metadata)
		if err != nil {
			log.Printf("❌ Error storing file: %v", err)
			continue
		}

		if hash != actualHash {
			log.Printf("⚠️  Warning: Hash mismatch! Expected %s, got %s", hash, actualHash)
		}

		fmt.Printf("   ✅ Stored successfully\n")
		fmt.Printf("   📊 Size: %d bytes\n", len(file.content))

		// Store with custom key for easy access
		err = store.StoreWithKey(file.key, []byte(file.content), file.metadata)
		if err != nil {
			log.Printf("❌ Error storing with key: %v", err)
			continue
		}
		fmt.Printf("   🔑 Key association: %s\n", file.key)

		storedHashes = append(storedHashes, actualHash)
	}

	fmt.Println("\n2. File Retrieval and Content Verification")
	fmt.Println("==========================================")

	// Retrieve and verify each file
	for i, hash := range storedHashes {
		fmt.Printf("\n🔍 Retrieving file %d by hash: %s...\n", i+1, hash[:12]+"...")

		data, metadata, err := store.Retrieve(hash)
		if err != nil {
			log.Printf("❌ Error retrieving file: %v", err)
			continue
		}

		fmt.Printf("   ✅ Retrieved successfully\n")
		fmt.Printf("   📊 Size: %d bytes\n", len(data))
		fmt.Printf("   📋 Metadata:\n")
		for k, v := range metadata {
			fmt.Printf("      %s: %s\n", k, v)
		}

		// Show content preview
		content := string(data)
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		fmt.Printf("   📝 Content preview: %s\n", strings.ReplaceAll(content, "\n", "\\n"))

		// Verify content integrity
		originalContent := testFiles[i].content
		if string(data) == originalContent {
			fmt.Printf("   ✅ Content integrity verified\n")
		} else {
			fmt.Printf("   ❌ Content integrity check failed!\n")
		}
	}

	fmt.Println("\n3. Key-Based Operations and Search")
	fmt.Println("==================================")

	// Test key-based operations
	fmt.Printf("🔍 Finding files by key prefix 'user_':\n")
	keys, err := store.FindByPrefix("user_")
	if err != nil {
		log.Printf("❌ Error finding by prefix: %v", err)
	} else {
		if len(keys) > 0 {
			for _, key := range keys {
				fmt.Printf("   📄 Found: %s\n", key)
			}
		} else {
			fmt.Printf("   📭 No files found with prefix 'user_'\n")
		}
	}

	fmt.Printf("\n🔍 Finding files by key prefix 'hello':\n")
	keys, err = store.FindByPrefix("hello")
	if err != nil {
		log.Printf("❌ Error finding by prefix: %v", err)
	} else {
		if len(keys) > 0 {
			for _, key := range keys {
				fmt.Printf("   📄 Found: %s\n", key)
			}
		} else {
			fmt.Printf("   📭 No files found with prefix 'hello'\n")
		}
	}

	fmt.Println("\n4. Deduplication Demonstration")
	fmt.Println("==============================")

	// Store identical content multiple times
	duplicateContent := "This is identical content that should be deduplicated!"

	fmt.Printf("🔄 Storing identical content multiple times:\n")
	fmt.Printf("   Content: %s\n", duplicateContent)

	// Store the same content with different metadata
	metadata1 := map[string]string{"version": "1.0", "author": "user1"}
	metadata2 := map[string]string{"version": "2.0", "author": "user2"}

	hash1, err := store.Store([]byte(duplicateContent), metadata1)
	if err != nil {
		log.Printf("❌ Error storing duplicate 1: %v", err)
	} else {
		fmt.Printf("   Hash 1: %s\n", hash1)
	}

	hash2, err := store.Store([]byte(duplicateContent), metadata2)
	if err != nil {
		log.Printf("❌ Error storing duplicate 2: %v", err)
	} else {
		fmt.Printf("   Hash 2: %s\n", hash2)
	}

	if hash1 == hash2 {
		fmt.Printf("   ✅ Deduplication working! Same hash for identical content\n")
	} else {
		fmt.Printf("   ❌ Deduplication failed! Different hashes for identical content\n")
	}

	// Store with different keys pointing to same content
	err = store.StoreWithKey("document_v1", []byte(duplicateContent), metadata1)
	if err != nil {
		log.Printf("❌ Error storing with key document_v1: %v", err)
	}

	err = store.StoreWithKey("document_v2", []byte(duplicateContent), metadata2)
	if err != nil {
		log.Printf("❌ Error storing with key document_v2: %v", err)
	}

	// Show all keys that point to the same content hash
	keys, err = store.FindByHash(hash1)
	if err != nil {
		log.Printf("❌ Error finding keys by hash: %v", err)
	} else {
		fmt.Printf("   🔑 Keys pointing to content hash %s: %v\n", hash1[:12]+"...", keys)
	}

	fmt.Println("\n5. Metadata Management")
	fmt.Println("======================")

	if len(storedHashes) > 0 {
		testHash := storedHashes[0]
		fmt.Printf("🏷️  Testing metadata operations on: %s...\n", testHash[:12])

		// Get current metadata
		metadata, err := store.GetMetadata(testHash)
		if err != nil {
			log.Printf("❌ Error getting metadata: %v", err)
		} else {
			fmt.Printf("   📋 Current metadata:\n")
			for k, v := range metadata {
				fmt.Printf("      %s: %s\n", k, v)
			}
		}

		// Update metadata with new fields
		updates := map[string]string{
			"last_accessed": "2024-01-15T10:30:00Z",
			"tags":          "demo,test,updated,featured",
			"version":       "1.1",
		}

		fmt.Printf("\n   🔄 Updating metadata with: %v\n", updates)
		err = store.UpdateMetadata(testHash, updates)
		if err != nil {
			log.Printf("❌ Error updating metadata: %v", err)
		} else {
			fmt.Printf("   ✅ Metadata updated successfully\n")
		}

		// Show updated metadata
		metadata, err = store.GetMetadata(testHash)
		if err != nil {
			log.Printf("❌ Error getting updated metadata: %v", err)
		} else {
			fmt.Printf("   📋 Updated metadata:\n")
			for k, v := range metadata {
				fmt.Printf("      %s: %s\n", k, v)
			}
		}
	}

	fmt.Println("\n6. Stream Operations for Large Content")
	fmt.Println("=====================================")

	// Create larger content for stream testing
	largeContent := strings.Repeat("This is a line of content for stream testing.\n", 100)

	fmt.Printf("📁 Testing stream operations with large content (%d bytes)\n", len(largeContent))

	// Store via stream
	reader := strings.NewReader(largeContent)
	streamMetadata := map[string]string{
		"content_type":   "text/plain",
		"source":         "stream_demo",
		"operation_type": "stream_store",
		"size":           fmt.Sprintf("%d", len(largeContent)),
	}

	hash, size, err := store.StoreStream(reader, streamMetadata)
	if err != nil {
		log.Printf("❌ Error storing via stream: %v", err)
	} else {
		fmt.Printf("   ✅ Stream storage successful\n")
		fmt.Printf("   📊 Hash: %s\n", hash)
		fmt.Printf("   📊 Size: %d bytes\n", size)
	}

	// Retrieve via stream
	fmt.Printf("\n🔄 Retrieving content via stream:\n")
	readCloser, retrievedMetadata, err := store.RetrieveStream(hash)
	if err != nil {
		log.Printf("❌ Error retrieving via stream: %v", err)
	} else {
		defer readCloser.Close()

		// Read first chunk to verify
		buffer := make([]byte, 200)
		n, err := readCloser.Read(buffer)
		if err != nil && err != io.EOF {
			log.Printf("❌ Error reading stream: %v", err)
		} else {
			fmt.Printf("   ✅ Stream retrieval successful\n")
			fmt.Printf("   📊 Read %d bytes\n", n)
			fmt.Printf("   📋 Metadata: %v\n", retrievedMetadata)
			fmt.Printf("   📝 Content preview: %s...\n", string(buffer[:min(n, 80)]))
		}
	}

	fmt.Println("\n7. Statistics and Performance Metrics")
	fmt.Println("====================================")

	// Get file store statistics
	stats, err := store.Stats()
	if err != nil {
		log.Printf("❌ Error getting stats: %v", err)
	} else {
		fmt.Printf("📊 File Store Statistics:\n")
		fmt.Printf("   📁 Total files: %d\n", stats.FileCount)
		fmt.Printf("   💾 Total size: %d bytes (%.2f KB)\n", stats.TotalSize, float64(stats.TotalSize)/1024.0)
		fmt.Printf("   📏 Average file size: %.2f bytes\n", stats.AverageSize)
		fmt.Printf("   🕐 Last access: %s\n", stats.LastAccess.Format("2006-01-02 15:04:05"))
		fmt.Printf("   📋 Content types:\n")
		for contentType, count := range stats.ContentTypes {
			fmt.Printf("      %s: %d files\n", contentType, count)
		}
	}

	// Get deduplication statistics
	dedupStats, err := store.GetDeduplicationStats()
	if err != nil {
		log.Printf("❌ Error getting deduplication stats: %v", err)
	} else {
		fmt.Printf("\n🔄 Deduplication Statistics:\n")
		fmt.Printf("   📁 Total file references: %d\n", dedupStats.TotalFiles)
		fmt.Printf("   🎯 Unique files stored: %d\n", dedupStats.UniqueFiles)
		fmt.Printf("   📄 Duplicate references: %d\n", dedupStats.DuplicateFiles)
		fmt.Printf("   💾 Space saved: %d bytes\n", dedupStats.SpaceSaved)
		fmt.Printf("   📊 Deduplication rate: %.1f%%\n", dedupStats.DeduplicationRate*100)
	}

	fmt.Println("\n✅ File Storage Demo Completed Successfully!")
	fmt.Println("\n📚 Summary of Features Demonstrated:")
	fmt.Println("   • Content-addressable storage using SHA-256 hashing")
	fmt.Println("   • Automatic content deduplication")
	fmt.Println("   • Configurable compression support")
	fmt.Println("   • Rich metadata management and updates")
	fmt.Println("   • Key-based file organization and retrieval")
	fmt.Println("   • Stream-based operations for efficient large file handling")
	fmt.Println("   • Comprehensive statistics and performance monitoring")
	fmt.Println("   • Content integrity verification")
	fmt.Println("   • Prefix-based key searching")
	fmt.Println()
	fmt.Println("🎯 This demonstrates the power of content-addressable storage")
	fmt.Println("   for building efficient, deduplicating file storage systems!")
}

// Helper function for Go versions that don't have min builtin
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
