package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/godast/godast/pkg/omnistore"
)

func main() {
	fmt.Println("🗄️  File Storage Demo - OmniStore")
	fmt.Println("==================================")

	// Create temporary directory for demo
	tempDir := "./demo_data"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatal("Failed to create demo directory:", err)
	}
	defer os.RemoveAll(tempDir)

	// Configure file store
	config := omnistore.DefaultFileStoreConfig()
	config.StorageDir = "files"
	config.EnableCompression = true
	config.EnableDeduplication = true

	// Create file store
	fileStore, err := omnistore.NewFileStoreImpl(config, tempDir)
	if err != nil {
		log.Fatal("Failed to create file store:", err)
	}
	defer fileStore.Close()

	fmt.Println("\n1. Basic File Storage Operations")
	fmt.Println("================================")

	// Store some test files
	testFiles := []struct {
		content  string
		metadata map[string]string
		key      string
	}{
		{
			content: "Hello, World! This is a test document.",
			metadata: map[string]string{
				"content_type": "text/plain",
				"author":       "demo",
				"category":     "test",
			},
			key: "hello_world.txt",
		},
		{
			content: strings.Repeat("Compress this text! ", 100),
			metadata: map[string]string{
				"content_type": "text/plain",
				"author":       "demo",
				"category":     "compression_test",
			},
			key: "compression_test.txt",
		},
		{
			content: `{"name": "Alice", "age": 30, "city": "New York"}`,
			metadata: map[string]string{
				"content_type": "application/json",
				"author":       "api",
				"category":     "user_data",
			},
			key: "user_profile.json",
		},
	}

	hashes := make([]string, 0, len(testFiles))

	// Store files
	for i, file := range testFiles {
		fmt.Printf("Storing file %d: %s\n", i+1, file.key)

		// Store with content hash
		hash, err := fileStore.Store([]byte(file.content), file.metadata)
		if err != nil {
			log.Printf("Error storing file: %v", err)
			continue
		}
		fmt.Printf("  Content Hash: %s\n", hash)

		// Store with key
		err = fileStore.StoreWithKey(file.key, []byte(file.content), file.metadata)
		if err != nil {
			log.Printf("Error storing with key: %v", err)
			continue
		}
		fmt.Printf("  Stored with key: %s\n", file.key)

		hashes = append(hashes, hash)
	}

	fmt.Println("\n2. File Retrieval Operations")
	fmt.Println("============================")

	// Retrieve files by hash
	for i, hash := range hashes {
		fmt.Printf("Retrieving file by hash: %s...\n", hash[:16])

		data, metadata, err := fileStore.Retrieve(hash)
		if err != nil {
			log.Printf("Error retrieving file: %v", err)
			continue
		}

		fmt.Printf("  Size: %d bytes\n", len(data))
		fmt.Printf("  Content type: %s\n", metadata["content_type"])
		fmt.Printf("  Author: %s\n", metadata["author"])

		// Show first 50 chars of content
		content := string(data)
		if len(content) > 50 {
			content = content[:50] + "..."
		}
		fmt.Printf("  Content: %s\n", content)
	}

	fmt.Println("\n3. Key-based File Operations")
	fmt.Println("============================")

	// Find files by key prefix
	keys, err := fileStore.FindByPrefix("user_")
	if err != nil {
		log.Printf("Error finding by prefix: %v", err)
	} else {
		fmt.Printf("Files with prefix 'user_': %v\n", keys)
	}

	// Find keys by hash (deduplication demo)
	fmt.Println("\n4. Deduplication Demo")
	fmt.Println("====================")

	// Store same content with different keys
	duplicateContent := "This content will be deduplicated!"
	metadata1 := map[string]string{"version": "1"}
	metadata2 := map[string]string{"version": "2"}

	hash1, err := fileStore.Store([]byte(duplicateContent), metadata1)
	if err != nil {
		log.Printf("Error storing duplicate 1: %v", err)
	} else {
		fmt.Printf("First storage hash: %s\n", hash1)
	}

	hash2, err := fileStore.Store([]byte(duplicateContent), metadata2)
	if err != nil {
		log.Printf("Error storing duplicate 2: %v", err)
	} else {
		fmt.Printf("Second storage hash: %s\n", hash2)
		fmt.Printf("Hashes are identical: %v\n", hash1 == hash2)
	}

	// Store with different keys
	err = fileStore.StoreWithKey("doc_v1", []byte(duplicateContent), metadata1)
	if err != nil {
		log.Printf("Error storing with key doc_v1: %v", err)
	}

	err = fileStore.StoreWithKey("doc_v2", []byte(duplicateContent), metadata2)
	if err != nil {
		log.Printf("Error storing with key doc_v2: %v", err)
	}

	// Show keys that point to same hash
	keys, err = fileStore.FindByHash(hash1)
	if err != nil {
		log.Printf("Error finding by hash: %v", err)
	} else {
		fmt.Printf("Keys pointing to hash %s: %v\n", hash1[:16], keys)
	}

	fmt.Println("\n5. Metadata Operations")
	fmt.Println("======================")

	if len(hashes) > 0 {
		hash := hashes[0]

		// Get current metadata
		metadata, err := fileStore.GetMetadata(hash)
		if err != nil {
			log.Printf("Error getting metadata: %v", err)
		} else {
			fmt.Printf("Current metadata for %s:\n", hash[:16])
			for k, v := range metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		// Update metadata
		updates := map[string]string{
			"last_accessed": "2023-12-01",
			"tags":          "demo,test,updated",
		}
		err = fileStore.UpdateMetadata(hash, updates)
		if err != nil {
			log.Printf("Error updating metadata: %v", err)
		} else {
			fmt.Printf("Updated metadata with: %v\n", updates)
		}

		// Show updated metadata
		metadata, err = fileStore.GetMetadata(hash)
		if err != nil {
			log.Printf("Error getting updated metadata: %v", err)
		} else {
			fmt.Printf("Updated metadata:\n")
			for k, v := range metadata {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
	}

	fmt.Println("\n6. Statistics and Deduplication")
	fmt.Println("===============================")

	// Show file store statistics
	stats, err := fileStore.Stats()
	if err != nil {
		log.Printf("Error getting stats: %v", err)
	} else {
		fmt.Printf("File Store Statistics:\n")
		fmt.Printf("  Total files: %d\n", stats.FileCount)
		fmt.Printf("  Total size: %d bytes\n", stats.TotalSize)
		fmt.Printf("  Average file size: %.2f bytes\n", stats.AverageSize)
		fmt.Printf("  Content types:\n")
		for contentType, count := range stats.ContentTypes {
			fmt.Printf("    %s: %d files\n", contentType, count)
		}
		fmt.Printf("  Last access: %s\n", stats.LastAccess.Format("2006-01-02 15:04:05"))
	}

	// Show deduplication statistics
	dedupStats, err := fileStore.GetDeduplicationStats()
	if err != nil {
		log.Printf("Error getting deduplication stats: %v", err)
	} else {
		fmt.Printf("\nDeduplication Statistics:\n")
		fmt.Printf("  Total files: %d\n", dedupStats.TotalFiles)
		fmt.Printf("  Unique files: %d\n", dedupStats.UniqueFiles)
		fmt.Printf("  Duplicate files: %d\n", dedupStats.DuplicateFiles)
		fmt.Printf("  Space saved: %d bytes\n", dedupStats.SpaceSaved)
		fmt.Printf("  Deduplication rate: %.2f%%\n", dedupStats.DeduplicationRate*100)
	}

	fmt.Println("\n7. Stream Operations")
	fmt.Println("===================")

	// Store from stream
	content := "This is content stored via stream operations!"
	reader := strings.NewReader(content)
	streamMetadata := map[string]string{
		"content_type": "text/plain",
		"source":       "stream",
	}

	hash, size, err := fileStore.StoreStream(reader, streamMetadata)
	if err != nil {
		log.Printf("Error storing stream: %v", err)
	} else {
		fmt.Printf("Stored stream content:\n")
		fmt.Printf("  Hash: %s\n", hash)
		fmt.Printf("  Size: %d bytes\n", size)
	}

	// Retrieve as stream
	readCloser, retrievedMetadata, err := fileStore.RetrieveStream(hash)
	if err != nil {
		log.Printf("Error retrieving stream: %v", err)
	} else {
		defer readCloser.Close()

		buffer := make([]byte, 1024)
		n, err := readCloser.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			log.Printf("Error reading stream: %v", err)
		} else {
			fmt.Printf("Retrieved stream content: %s\n", string(buffer[:n]))
			fmt.Printf("Retrieved metadata: %v\n", retrievedMetadata)
		}
	}

	fmt.Println("\n✅ File Storage Demo completed successfully!")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("• Content-addressable storage with SHA-256 hashing")
	fmt.Println("• Automatic deduplication")
	fmt.Println("• Compression support")
	fmt.Println("• Rich metadata management")
	fmt.Println("• Key-based file organization")
	fmt.Println("• Stream-based operations for large files")
	fmt.Println("• Comprehensive statistics and monitoring")
}
