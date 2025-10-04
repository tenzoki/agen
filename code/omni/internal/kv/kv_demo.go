package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/agen/omni/internal/kv"
	"github.com/agen/omni/internal/storage"
)

func main() {
	fmt.Println("🔑 KV Store Complete Demo")
	fmt.Println("==================================================")

	// Setup temporary storage
	tmpDir := "/tmp/kv-store-complete-demo"
	defer os.RemoveAll(tmpDir)

	config := storage.DefaultConfig(tmpDir)
	store, err := storage.NewBadgerStore(config)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	kvStore := kv.NewKVStore(store)

	// Run all demo sections
	fmt.Println("\n1. Basic KV Operations")
	demoBasicOperations(kvStore)

	fmt.Println("\n2. Batch Operations")
	demoBatchOperations(kvStore)

	fmt.Println("\n3. Range Query Operations")
	demoRangeQueries(kvStore)

	fmt.Println("\n4. TTL Operations")
	demoTTLOperations(kvStore)

	fmt.Println("\n5. Performance Operations")
	demoPerformanceOperations(kvStore)

	fmt.Println("\n6. Statistics")
	printKVStatistics(kvStore)

	fmt.Println("\n✅ KV Store demo completed successfully!")
}

func demoBasicOperations(kv kv.KVStore) {
	fmt.Println("   Basic CRUD operations...")

	// Set operations
	data := map[string][]byte{
		"user:1001":      []byte("Alice Johnson"),
		"user:1002":      []byte("Bob Smith"),
		"config:timeout": []byte("30s"),
		"config:retries": []byte("5"),
		"session:abc123": []byte("active"),
	}

	fmt.Println("   Setting key-value pairs:")
	for key, value := range data {
		if err := kv.Set(key, value); err != nil {
			log.Printf("   ❌ Failed to set %s: %v", key, err)
		} else {
			fmt.Printf("   ✅ %s = %s\n", key, string(value))
		}
	}

	// Get operations
	fmt.Println("\n   Retrieving values:")
	for key := range data {
		if value, err := kv.Get(key); err != nil {
			log.Printf("   ❌ Failed to get %s: %v", key, err)
		} else {
			fmt.Printf("   📋 %s = %s\n", key, string(value))
		}
	}

	// Exists operations
	fmt.Println("\n   Checking existence:")
	testKeys := []string{"user:1001", "user:9999", "config:timeout"}
	for _, key := range testKeys {
		if exists, err := kv.Exists(key); err != nil {
			log.Printf("   ❌ Failed to check %s: %v", key, err)
		} else if exists {
			fmt.Printf("   ✅ Key exists: %s\n", key)
		} else {
			fmt.Printf("   ❌ Key missing: %s\n", key)
		}
	}

	// Delete operation
	fmt.Println("\n   Deleting a key:")
	if err := kv.Delete("session:abc123"); err != nil {
		log.Printf("   ❌ Failed to delete: %v", err)
	} else {
		fmt.Println("   🗑️  Deleted: session:abc123")
		// Verify deletion
		if exists, _ := kv.Exists("session:abc123"); !exists {
			fmt.Println("   ✅ Confirmed: key no longer exists")
		}
	}
}

func demoBatchOperations(kv kv.KVStore) {
	fmt.Println("   Bulk operations for efficiency...")

	// Batch set operation
	products := map[string][]byte{
		"product:electronics:laptop": []byte("MacBook Pro"),
		"product:electronics:phone":  []byte("iPhone 15"),
		"product:electronics:tablet": []byte("iPad Air"),
		"product:books:programming":  []byte("Clean Code"),
		"product:books:algorithms":   []byte("CLRS Algorithms"),
		"product:clothing:shirt":     []byte("Cotton T-Shirt"),
		"product:clothing:jeans":     []byte("Denim Jeans"),
	}

	fmt.Printf("   Setting %d products in batch:\n", len(products))
	if err := kv.BatchSet(products); err != nil {
		log.Printf("   ❌ Batch set failed: %v", err)
	} else {
		fmt.Printf("   ✅ Batch set %d items successfully\n", len(products))
	}

	// Batch get operation
	productKeys := []string{
		"product:electronics:laptop",
		"product:books:programming",
		"product:clothing:shirt",
		"product:nonexistent:item", // This won't exist
	}

	fmt.Printf("\n   Retrieving %d products in batch:\n", len(productKeys))
	results, err := kv.BatchGet(productKeys)
	if err != nil {
		log.Printf("   ❌ Batch get failed: %v", err)
	} else {
		fmt.Printf("   📦 Retrieved %d out of %d requested items:\n", len(results), len(productKeys))
		for key, value := range results {
			fmt.Printf("     • %s: %s\n", key, string(value))
		}
	}
}

func demoRangeQueries(kv kv.KVStore) {
	fmt.Println("   Prefix-based range scanning...")

	prefixes := []string{
		"user:",
		"config:",
		"product:electronics:",
		"product:books:",
		"product:clothing:",
	}

	for _, prefix := range prefixes {
		fmt.Printf("\n   🔍 Scanning prefix '%s':\n", prefix)
		results, err := kv.Scan(prefix, 10) // Limit to 10 results
		if err != nil {
			log.Printf("     ❌ Scan failed: %v", err)
			continue
		}

		if len(results) == 0 {
			fmt.Printf("     (no matches found)\n")
		} else {
			fmt.Printf("     Found %d matches:\n", len(results))
			for key, value := range results {
				fmt.Printf("       • %s: %s\n", key, string(value))
			}
		}
	}

	// Demonstrate unlimited scan
	fmt.Printf("\n   🔍 Complete scan of all products (unlimited):\n")
	allProducts, err := kv.Scan("product:", -1) // No limit
	if err != nil {
		log.Printf("     ❌ Unlimited scan failed: %v", err)
	} else {
		fmt.Printf("     Found %d total products\n", len(allProducts))
	}
}

func demoTTLOperations(kv kv.KVStore) {
	fmt.Println("   Time-To-Live (TTL) operations...")

	// Set keys with different TTL values
	ttlData := map[string]struct {
		value []byte
		ttl   time.Duration
	}{
		"cache:short":  {[]byte("Short-lived data"), 2 * time.Second},
		"cache:medium": {[]byte("Medium-lived data"), 5 * time.Second},
		"cache:long":   {[]byte("Long-lived data"), 10 * time.Second},
	}

	fmt.Println("   Setting keys with TTL:")
	for key, data := range ttlData {
		if err := kv.SetWithTTL(key, data.value, data.ttl); err != nil {
			log.Printf("   ❌ Failed to set TTL key %s: %v", key, err)
		} else {
			fmt.Printf("   ⏰ %s (TTL: %v) = %s\n", key, data.ttl, string(data.value))
		}
	}

	// Check initial existence
	fmt.Println("\n   Checking initial existence:")
	for key := range ttlData {
		if value, err := kv.Get(key); err != nil {
			log.Printf("   ❌ Failed to get %s: %v", key, err)
		} else {
			fmt.Printf("   ✅ %s = %s\n", key, string(value))
		}
	}

	// Wait and check expiration
	fmt.Println("\n   ⏳ Waiting 3 seconds to observe TTL expiration...")
	time.Sleep(3 * time.Second)

	fmt.Println("   Checking after 3 seconds:")
	for key := range ttlData {
		if value, err := kv.Get(key); err != nil {
			if err.Error() == "key not found" {
				fmt.Printf("   💀 %s has expired\n", key)
			} else {
				log.Printf("   ❌ Error checking %s: %v", key, err)
			}
		} else {
			fmt.Printf("   ✅ %s still exists = %s\n", key, string(value))
		}
	}
}

func demoPerformanceOperations(kv kv.KVStore) {
	fmt.Println("   Performance demonstration...")

	// Bulk insert for performance testing
	bulkData := make(map[string][]byte)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("perf:test:%04d", i)
		value := fmt.Sprintf("Performance test data item #%d with some content", i)
		bulkData[key] = []byte(value)
	}

	fmt.Printf("   📈 Bulk inserting %d items...\n", len(bulkData))
	start := time.Now()
	if err := kv.BatchSet(bulkData); err != nil {
		log.Printf("   ❌ Bulk insert failed: %v", err)
	} else {
		duration := time.Since(start)
		opsPerSec := float64(len(bulkData)) / duration.Seconds()
		fmt.Printf("   ✅ Bulk insert completed in %v (%.0f ops/sec)\n", duration, opsPerSec)
	}

	// Bulk read for performance testing
	keys := make([]string, 500) // Read half of the data
	for i := range keys {
		keys[i] = fmt.Sprintf("perf:test:%04d", i)
	}

	fmt.Printf("   📈 Bulk reading %d items...\n", len(keys))
	start = time.Now()
	results, err := kv.BatchGet(keys)
	if err != nil {
		log.Printf("   ❌ Bulk read failed: %v", err)
	} else {
		duration := time.Since(start)
		opsPerSec := float64(len(results)) / duration.Seconds()
		fmt.Printf("   ✅ Bulk read completed in %v (%.0f ops/sec, %d items)\n",
			duration, opsPerSec, len(results))
	}

	// Range scan performance
	fmt.Println("   📈 Range scan performance...")
	start = time.Now()
	scanResults, err := kv.Scan("perf:test:", -1)
	if err != nil {
		log.Printf("   ❌ Range scan failed: %v", err)
	} else {
		duration := time.Since(start)
		fmt.Printf("   ✅ Scanned %d items in %v\n", len(scanResults), duration)
	}
}

func printKVStatistics(kv kv.KVStore) {
	fmt.Println("   Comprehensive statistics...")

	stats, err := kv.Stats()
	if err != nil {
		log.Printf("   ❌ Failed to get statistics: %v", err)
		return
	}

	fmt.Printf("   📊 KV Store Statistics:\n")
	fmt.Printf("     • Total Keys: %d\n", stats.KeyCount)
	fmt.Printf("     • Total Size: %d bytes (%.2f MB)\n",
		stats.TotalSize, float64(stats.TotalSize)/1024/1024)
	fmt.Printf("     • Namespace: %s\n", stats.Namespace)
	fmt.Printf("     • Average Key Size: %.2f bytes\n", stats.AvgKeySize)
	fmt.Printf("     • Average Value Size: %.2f bytes\n", stats.AvgValueSize)
	fmt.Printf("     • Last Access: %s\n", stats.LastAccess.Format(time.RFC3339))

	// Calculate efficiency metrics
	if stats.KeyCount > 0 {
		overhead := float64(stats.TotalSize) / float64(stats.KeyCount)
		fmt.Printf("     • Storage Overhead: %.2f bytes per item\n", overhead)

		efficiency := (stats.AvgValueSize / (stats.AvgKeySize + stats.AvgValueSize)) * 100
		fmt.Printf("     • Storage Efficiency: %.1f%% (value/total ratio)\n", efficiency)
	}
}
