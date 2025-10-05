package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/tenzoki/agen/omni/public/omnistore"
)

// ValidationResult holds the result of a validation test
type ValidationResult struct {
	TestName string        `json:"test_name"`
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Details  interface{}   `json:"details,omitempty"`
}

// ValidationSuite runs comprehensive validation tests
type ValidationSuite struct {
	omniStore omnistore.OmniStore
	results   []ValidationResult
	dataPath  string
}

func main() {
	fmt.Println("=== Godast Storage Validation Suite ===")

	// Create temporary directory for validation
	tempDir := filepath.Join(os.TempDir(), "gox-storage-validation-"+uuid.New().String())
	defer os.RemoveAll(tempDir)

	fmt.Printf("Using validation directory: %s\n\n", tempDir)

	// Initialize validation suite
	suite := &ValidationSuite{
		dataPath: tempDir,
		results:  make([]ValidationResult, 0),
	}

	// Run all validation tests
	suite.runValidationTests()

	// Print results
	suite.printResults()

	// Generate report
	suite.generateReport()
}

func (vs *ValidationSuite) runValidationTests() {
	fmt.Println("Running validation tests...")

	// Test 1: OmniStore Initialization
	vs.runTest("OmniStore Initialization", vs.testOmniStoreInit)

	// Test 2: KV Operations
	vs.runTest("KV Store Operations", vs.testKVOperations)

	// Test 3: File Operations
	vs.runTest("File Store Operations", vs.testFileOperations)

	// Test 4: Search Operations
	vs.runTest("Search Operations", vs.testSearchOperations)

	// Test 5: Concurrent Operations
	vs.runTest("Concurrent Operations", vs.testConcurrentOperations)

	// Test 6: Large Data Handling
	vs.runTest("Large Data Handling", vs.testLargeDataHandling)

	// Test 7: Error Handling
	vs.runTest("Error Handling", vs.testErrorHandling)

	// Test 8: Performance Benchmarks
	vs.runTest("Performance Benchmarks", vs.testPerformanceBenchmarks)

	// Cleanup
	if vs.omniStore != nil {
		vs.omniStore.Close()
	}
}

func (vs *ValidationSuite) runTest(testName string, testFunc func() (interface{}, error)) {
	fmt.Printf("🔍 %s...", testName)

	start := time.Now()
	details, err := testFunc()
	duration := time.Since(start)

	result := ValidationResult{
		TestName: testName,
		Success:  err == nil,
		Duration: duration,
		Details:  details,
	}

	if err != nil {
		result.Error = err.Error()
		fmt.Printf(" ❌ FAIL (%v)\n", duration)
		fmt.Printf("   Error: %s\n", err.Error())
	} else {
		fmt.Printf(" ✅ PASS (%v)\n", duration)
		if details != nil {
			fmt.Printf("   Details: %+v\n", details)
		}
	}

	vs.results = append(vs.results, result)
}

func (vs *ValidationSuite) testOmniStoreInit() (interface{}, error) {
	// Initialize OmniStore
	omniStore, err := omnistore.NewOmniStoreWithDefaults(vs.dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create OmniStore: %w", err)
	}

	vs.omniStore = omniStore

	// Test basic functionality
	config := omniStore.GetConfig()
	if config == nil {
		return nil, fmt.Errorf("OmniStore config is nil")
	}

	return map[string]interface{}{
		"data_path": vs.dataPath,
		"config":    "loaded",
	}, nil
}

func (vs *ValidationSuite) testKVOperations() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	kvStore := vs.omniStore.KV()
	testKey := "validation:test:kv"
	testValue := []byte(`{"test": "data", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`)

	// Test Set
	err := kvStore.Set(testKey, testValue)
	if err != nil {
		return nil, fmt.Errorf("KV Set failed: %w", err)
	}

	// Test Get
	retrievedValue, err := kvStore.Get(testKey)
	if err != nil {
		return nil, fmt.Errorf("KV Get failed: %w", err)
	}

	if string(retrievedValue) != string(testValue) {
		return nil, fmt.Errorf("retrieved value doesn't match stored value")
	}

	// Test Delete
	err = kvStore.Delete(testKey)
	if err != nil {
		return nil, fmt.Errorf("KV Delete failed: %w", err)
	}

	// Verify deletion
	_, err = kvStore.Get(testKey)
	if err == nil {
		return nil, fmt.Errorf("key should be deleted but still exists")
	}

	return map[string]interface{}{
		"operations": []string{"set", "get", "delete", "verify_deletion"},
		"test_key":   testKey,
		"value_size": len(testValue),
	}, nil
}

func (vs *ValidationSuite) testFileOperations() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	fileStore := vs.omniStore.Files()
	testContent := []byte("This is test content for file validation: " + time.Now().String())
	metadata := map[string]string{
		"filename":     "validation_test.txt",
		"content_type": "text/plain",
		"created_by":   "validation_suite",
	}

	// Test Store
	hash, err := fileStore.Store(testContent, metadata)
	if err != nil {
		return nil, fmt.Errorf("File Store failed: %w", err)
	}

	if hash == "" {
		return nil, fmt.Errorf("file hash is empty")
	}

	// Test Retrieve
	retrievedContent, retrievedMetadata, err := fileStore.Retrieve(hash)
	if err != nil {
		return nil, fmt.Errorf("File Retrieve failed: %w", err)
	}

	if string(retrievedContent) != string(testContent) {
		return nil, fmt.Errorf("retrieved content doesn't match stored content")
	}

	if retrievedMetadata["filename"] != metadata["filename"] {
		return nil, fmt.Errorf("retrieved metadata doesn't match stored metadata")
	}

	return map[string]interface{}{
		"hash":         hash,
		"content_size": len(testContent),
		"metadata":     retrievedMetadata,
	}, nil
}

func (vs *ValidationSuite) testSearchOperations() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	searchStore := vs.omniStore.Search()

	// Test documents
	documents := []struct {
		ID      string
		Content string
		Meta    map[string]interface{}
	}{
		{
			ID:      "doc1",
			Content: "This is a test document about storage validation and testing procedures",
			Meta: map[string]interface{}{
				"title":    "Storage Validation",
				"category": "testing",
			},
		},
		{
			ID:      "doc2",
			Content: "Advanced search capabilities with full-text indexing and retrieval",
			Meta: map[string]interface{}{
				"title":    "Search Features",
				"category": "features",
			},
		},
	}

	// Index documents
	for _, doc := range documents {
		err := searchStore.IndexDocument(doc.ID, doc.Content, doc.Meta)
		if err != nil {
			return nil, fmt.Errorf("failed to index document %s: %w", doc.ID, err)
		}
	}

	// Refresh index
	err := searchStore.RefreshIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh search index: %w", err)
	}

	// Test search
	results, err := searchStore.Search("storage validation", nil)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if results == nil {
		return nil, fmt.Errorf("search results are nil")
	}

	return map[string]interface{}{
		"indexed_docs": len(documents),
		"search_query": "storage validation",
		"has_results":  results != nil,
	}, nil
}

func (vs *ValidationSuite) testConcurrentOperations() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	kvStore := vs.omniStore.KV()
	numOperations := 100
	numWorkers := 10

	// Channel for results
	resultChan := make(chan error, numOperations)
	workerChan := make(chan int, numOperations)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			for j := range workerChan {
				key := fmt.Sprintf("concurrent:test:%d", j)
				value := []byte(fmt.Sprintf("value_%d_%s", j, time.Now().String()))

				// Set operation
				err := kvStore.Set(key, value)
				if err != nil {
					resultChan <- fmt.Errorf("concurrent set failed for key %s: %w", key, err)
					return
				}

				// Get operation
				_, err = kvStore.Get(key)
				if err != nil {
					resultChan <- fmt.Errorf("concurrent get failed for key %s: %w", key, err)
					return
				}

				resultChan <- nil
			}
		}()
	}

	// Send work
	start := time.Now()
	for i := 0; i < numOperations; i++ {
		workerChan <- i
	}
	close(workerChan)

	// Collect results
	errors := 0
	for i := 0; i < numOperations; i++ {
		if err := <-resultChan; err != nil {
			errors++
		}
	}
	duration := time.Since(start)

	if errors > 0 {
		return nil, fmt.Errorf("%d out of %d concurrent operations failed", errors, numOperations)
	}

	opsPerSecond := float64(numOperations) / duration.Seconds()

	return map[string]interface{}{
		"operations":     numOperations,
		"workers":        numWorkers,
		"duration":       duration,
		"ops_per_second": opsPerSecond,
		"errors":         errors,
	}, nil
}

func (vs *ValidationSuite) testLargeDataHandling() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	// Test with different data sizes
	sizes := []int{1024, 10240, 102400, 1048576} // 1KB, 10KB, 100KB, 1MB
	results := make(map[string]interface{})

	for _, size := range sizes {
		// Create test data
		testData := make([]byte, size)
		for i := range testData {
			testData[i] = byte(i % 256)
		}

		// Test KV operations with large data
		kvStore := vs.omniStore.KV()
		key := fmt.Sprintf("large_data:%d", size)

		start := time.Now()
		err := kvStore.Set(key, testData)
		setDuration := time.Since(start)

		if err != nil {
			return nil, fmt.Errorf("failed to set large data (%d bytes): %w", size, err)
		}

		start = time.Now()
		retrievedData, err := kvStore.Get(key)
		getDuration := time.Since(start)

		if err != nil {
			return nil, fmt.Errorf("failed to get large data (%d bytes): %w", size, err)
		}

		if len(retrievedData) != len(testData) {
			return nil, fmt.Errorf("retrieved data size mismatch for %d bytes: got %d", size, len(retrievedData))
		}

		results[fmt.Sprintf("%d_bytes", size)] = map[string]interface{}{
			"set_duration": setDuration,
			"get_duration": getDuration,
			"success":      true,
		}

		// Cleanup
		kvStore.Delete(key)
	}

	return results, nil
}

func (vs *ValidationSuite) testErrorHandling() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	kvStore := vs.omniStore.KV()
	testResults := make(map[string]bool)

	// Test 1: Get non-existent key
	_, err := kvStore.Get("non_existent_key")
	testResults["get_non_existent"] = err != nil // Should return error

	// Test 2: Delete non-existent key
	err = kvStore.Delete("non_existent_key")
	testResults["delete_non_existent"] = err != nil // Should return error

	// Test 3: Empty key operations
	err = kvStore.Set("", []byte("test"))
	testResults["set_empty_key"] = err != nil // Should return error

	// Test 4: Large key name
	largeKey := string(make([]byte, 10000)) // Very large key
	err = kvStore.Set(largeKey, []byte("test"))
	testResults["set_large_key"] = err != nil // Should return error or handle gracefully

	// Count successful error handling
	successfulTests := 0
	for _, success := range testResults {
		if success {
			successfulTests++
		}
	}

	return map[string]interface{}{
		"tests":            testResults,
		"successful_tests": successfulTests,
		"total_tests":      len(testResults),
	}, nil
}

func (vs *ValidationSuite) testPerformanceBenchmarks() (interface{}, error) {
	if vs.omniStore == nil {
		return nil, fmt.Errorf("OmniStore not initialized")
	}

	kvStore := vs.omniStore.KV()
	numOperations := 1000

	// Benchmark KV Set operations
	start := time.Now()
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("bench:set:%d", i)
		value := []byte(fmt.Sprintf("benchmark_value_%d", i))
		err := kvStore.Set(key, value)
		if err != nil {
			return nil, fmt.Errorf("benchmark set operation failed: %w", err)
		}
	}
	setDuration := time.Since(start)

	// Benchmark KV Get operations
	start = time.Now()
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("bench:set:%d", i)
		_, err := kvStore.Get(key)
		if err != nil {
			return nil, fmt.Errorf("benchmark get operation failed: %w", err)
		}
	}
	getDuration := time.Since(start)

	// Calculate rates
	setOpsPerSecond := float64(numOperations) / setDuration.Seconds()
	getOpsPerSecond := float64(numOperations) / getDuration.Seconds()

	return map[string]interface{}{
		"operations":      numOperations,
		"set_duration":    setDuration,
		"get_duration":    getDuration,
		"set_ops_per_sec": setOpsPerSecond,
		"get_ops_per_sec": getOpsPerSecond,
	}, nil
}

func (vs *ValidationSuite) printResults() {
	fmt.Println("\n=== Validation Results ===")

	passed := 0
	failed := 0

	for _, result := range vs.results {
		status := "❌ FAIL"
		if result.Success {
			status = "✅ PASS"
			passed++
		} else {
			failed++
		}

		fmt.Printf("%s %s (%v)\n", status, result.TestName, result.Duration)
		if !result.Success {
			fmt.Printf("    Error: %s\n", result.Error)
		}
	}

	fmt.Printf("\nSummary: %d passed, %d failed, %d total\n", passed, failed, len(vs.results))

	if failed == 0 {
		fmt.Println("🎉 All validation tests passed!")
	} else {
		fmt.Printf("⚠️  %d validation tests failed\n", failed)
	}
}

func (vs *ValidationSuite) generateReport() {
	reportPath := filepath.Join(vs.dataPath, "validation_report.json")

	report := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"data_path":    vs.dataPath,
		"total_tests":  len(vs.results),
		"passed_tests": vs.countPassedTests(),
		"failed_tests": vs.countFailedTests(),
		"results":      vs.results,
		"summary": map[string]interface{}{
			"all_passed":   vs.countFailedTests() == 0,
			"success_rate": float64(vs.countPassedTests()) / float64(len(vs.results)),
		},
	}

	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("Failed to generate JSON report: %v", err)
		return
	}

	err = os.WriteFile(reportPath, reportBytes, 0644)
	if err != nil {
		log.Printf("Failed to write report file: %v", err)
		return
	}

	fmt.Printf("\nValidation report written to: %s\n", reportPath)
}

func (vs *ValidationSuite) countPassedTests() int {
	count := 0
	for _, result := range vs.results {
		if result.Success {
			count++
		}
	}
	return count
}

func (vs *ValidationSuite) countFailedTests() int {
	count := 0
	for _, result := range vs.results {
		if !result.Success {
			count++
		}
	}
	return count
}
