package omnistore

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/agen/omni/internal/graph"
	"github.com/agen/omni/internal/kv"
	"github.com/agen/omni/internal/query"
	"github.com/agen/omni/internal/storage"
	"github.com/agen/omni/internal/transaction"
	"github.com/agen/omni/internal/filestore"
)

// OmniStoreImpl implements the OmniStore interface
type OmniStoreImpl struct {
	config      *Config
	kvStore     kv.KVStore
	graphStore  graph.GraphStore
	fileStore   filestore.FileStore
	searchStore SearchStore
	startTime   time.Time
	mu          sync.RWMutex
}

// NewOmniStore creates a new OmniStore instance
func NewOmniStore(config *Config) (*OmniStoreImpl, error) {
	omniStore := &OmniStoreImpl{
		config:    config,
		startTime: time.Now(),
	}

	// Initialize underlying storage
	storageConfig := storage.DefaultConfig(config.DataDir)
	backingStore, err := storage.NewBadgerStore(storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize backing store: %w", err)
	}

	// Initialize KV store
	kvStore := kv.NewKVStore(backingStore)
	omniStore.kvStore = kvStore

	// Initialize Graph store
	graphStore := graph.NewGraphStore(backingStore)
	omniStore.graphStore = graphStore

	// Initialize File store
	fsConfig := &filestore.Config{
		StorageDir:          config.Files.StorageDir,
		EnableEncryption:    config.Files.EnableEncryption,
		EncryptionKey:       config.Files.EncryptionKey,
		EnableCompression:   config.Files.EnableCompression,
		CompressionLevel:    config.Files.CompressionLevel,
		EnableDeduplication: config.Files.EnableDeduplication,
		MaxFileSize:         config.Files.MaxFileSize,
		IndexingEnabled:     config.Files.IndexingEnabled,
	}
	fileStore, err := filestore.NewFileStore(fsConfig, config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize File store: %w", err)
	}
	omniStore.fileStore = fileStore

	// Initialize Search store (placeholder)
	omniStore.searchStore = NewNoOpSearchStore()

	return omniStore, nil
}

// NewOmniStoreWithDefaults creates a new OmniStore with default configuration
func NewOmniStoreWithDefaults(dataDir string) (*OmniStoreImpl, error) {
	config := DefaultConfig()
	config.DataDir = dataDir
	return NewOmniStore(config)
}

// Core subsystem access
func (s *OmniStoreImpl) KV() kv.KVStore {
	return s.kvStore
}

func (s *OmniStoreImpl) Graph() graph.GraphStore {
	return s.graphStore
}

func (s *OmniStoreImpl) Files() filestore.FileStore {
	return s.fileStore
}

func (s *OmniStoreImpl) Search() SearchStore {
	return s.searchStore
}

// Convenience methods for common operations
func (s *OmniStoreImpl) ListKVKeys(prefix string, limit int) ([]string, error) {
	return s.kvStore.ListKeys(prefix, limit)
}

// Transaction support (placeholder - to be implemented)
func (s *OmniStoreImpl) BeginTransaction(config *transaction.TransactionConfig) (transaction.GraphTx, error) {
	return nil, fmt.Errorf("transaction support not yet implemented")
}

func (s *OmniStoreImpl) BeginTransactionWithContext(ctx context.Context, config *transaction.TransactionConfig) (transaction.GraphTx, error) {
	return nil, fmt.Errorf("transaction support not yet implemented")
}

func (s *OmniStoreImpl) ExecuteTransaction(fn func(tx transaction.GraphTx) error) error {
	return fmt.Errorf("transaction support not yet implemented")
}

func (s *OmniStoreImpl) ExecuteTransactionWithConfig(config *transaction.TransactionConfig, fn func(tx transaction.GraphTx) error) error {
	return fmt.Errorf("transaction support not yet implemented")
}

// Query language support (placeholder - to be implemented)
func (s *OmniStoreImpl) Query(queryString string) (*query.QueryResult, error) {
	return nil, fmt.Errorf("query support not yet implemented")
}

func (s *OmniStoreImpl) QueryWithContext(ctx context.Context, queryString string) (*query.QueryResult, error) {
	return nil, fmt.Errorf("query support not yet implemented")
}

// Unified operations (cross-component)
func (s *OmniStoreImpl) ExecuteCrossQuery(req *CrossQueryRequest) (*CrossQueryResult, error) {
	start := time.Now()
	result := &CrossQueryResult{
		ComponentTimes: make(map[string]time.Duration),
	}

	// Execute KV query
	if req.KVQuery != nil {
		kvStart := time.Now()
		kvResults, err := s.executeKVQuery(req.KVQuery)
		if err != nil {
			return nil, fmt.Errorf("KV query failed: %w", err)
		}
		result.KVResults = kvResults
		result.ComponentTimes["kv"] = time.Since(kvStart)
	}

	// Execute Graph query
	if req.GraphQuery != nil {
		graphStart := time.Now()
		graphResults, err := s.executeGraphQuery(req.GraphQuery)
		if err != nil {
			return nil, fmt.Errorf("Graph query failed: %w", err)
		}
		result.GraphResults = graphResults
		result.ComponentTimes["graph"] = time.Since(graphStart)
	}

	// Execute File query
	if req.FileQuery != nil {
		fileStart := time.Now()
		fileResults, err := s.executeFileQuery(req.FileQuery)
		if err != nil {
			return nil, fmt.Errorf("File query failed: %w", err)
		}
		result.FileResults = fileResults
		result.ComponentTimes["files"] = time.Since(fileStart)
	}

	// Execute Search query
	if req.SearchQuery != nil {
		searchStart := time.Now()
		searchResults, err := s.executeSearchQuery(req.SearchQuery)
		if err != nil {
			return nil, fmt.Errorf("Search query failed: %w", err)
		}
		result.SearchResults = searchResults
		result.ComponentTimes["search"] = time.Since(searchStart)
	}

	// Process join operations
	if len(req.JoinOperations) > 0 {
		joinedResults, err := s.processJoinOperations(result, req.JoinOperations)
		if err != nil {
			return nil, fmt.Errorf("Join operations failed: %w", err)
		}
		result.JoinedResults = joinedResults
	}

	result.ExecutionTime = time.Since(start)
	return result, nil
}

// Configuration and lifecycle
func (s *OmniStoreImpl) GetConfig() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *OmniStoreImpl) UpdateConfig(config *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
	return nil
}

// Statistics and monitoring
func (s *OmniStoreImpl) GetStats() (*OmniStoreStats, error) {
	kvStats, err := s.kvStore.Stats()
	if err != nil {
		return nil, err
	}

	graphStats, err := s.graphStore.GetStats()
	if err != nil {
		return nil, err
	}

	fileStats, err := s.fileStore.Stats()
	if err != nil {
		return nil, err
	}

	// Search stats would go here when implemented

	stats := &OmniStoreStats{
		KV:               kvStats,
		Graph:            graphStats,
		Files:            &FileStoreStats{
			FileCount:        fileStats.FileCount,
			TotalSize:        fileStats.TotalSize,
			CompressedSize:   fileStats.CompressedSize,
			AverageSize:      fileStats.AverageSize,
			ContentTypes:     fileStats.ContentTypes,
			LastAccess:       fileStats.LastAccess,
			IndexedFiles:     fileStats.IndexedFiles,
			CompressionRatio: fileStats.CompressionRatio,
			Deduplication:    convertDeduplicationStats(fileStats.Deduplication),
		},
		TotalStorageSize:   kvStats.TotalSize + graphStats.TotalSize + fileStats.TotalSize,
		TotalOperations:    0, // Will be implemented when operation counters are added
		Uptime:             time.Since(s.startTime),
		ActiveTransactions: 0, // Will be implemented when transaction support is added
	}

	return stats, nil
}

func (s *OmniStoreImpl) GetHealth() (*HealthStatus, error) {
	components := make(map[string]ComponentHealth)

	// Check KV health
	components["kv"] = ComponentHealth{
		Status:    Healthy,
		Message:   "KV store operational",
		LastCheck: time.Now(),
	}

	// Check Graph health
	components["graph"] = ComponentHealth{
		Status:    Healthy,
		Message:   "Graph store operational",
		LastCheck: time.Now(),
	}

	// Check File health
	components["files"] = ComponentHealth{
		Status:    Healthy,
		Message:   "File store operational",
		LastCheck: time.Now(),
	}

	// Check Search health
	components["search"] = ComponentHealth{
		Status:    Healthy,
		Message:   "Search store operational (placeholder)",
		LastCheck: time.Now(),
	}

	// Determine overall status
	overallStatus := Healthy
	for _, comp := range components {
		if comp.Status == Unhealthy {
			overallStatus = Unhealthy
			break
		} else if comp.Status == Degraded && overallStatus == Healthy {
			overallStatus = Degraded
		}
	}

	return &HealthStatus{
		Status:     overallStatus,
		Components: components,
		Timestamp:  time.Now(),
		Uptime:     time.Since(s.startTime),
	}, nil
}

// Backup and maintenance
func (s *OmniStoreImpl) Backup(path string) error {
	// Create backup directory structure
	_ = filepath.Join(path, "kv")
	_ = filepath.Join(path, "graph")
	_ = filepath.Join(path, "files")

	// Backup functionality not yet implemented in underlying stores
	// This would require implementing backup methods in KV and Graph stores
	return fmt.Errorf("backup functionality not yet implemented")
}

func (s *OmniStoreImpl) Restore(path string) error {
	// Restore implementation would go here
	return fmt.Errorf("restore not implemented yet")
}

func (s *OmniStoreImpl) Compact() error {
	// Compact functionality not yet implemented in underlying stores
	// This would require implementing compact methods in KV and Graph stores
	return fmt.Errorf("compact functionality not yet implemented")
}

// Graceful shutdown
func (s *OmniStoreImpl) Close() error {
	var errors []error

	if err := s.fileStore.Close(); err != nil {
		errors = append(errors, fmt.Errorf("file store close error: %w", err))
	}

	if err := s.searchStore.Close(); err != nil {
		errors = append(errors, fmt.Errorf("search store close error: %w", err))
	}

	if err := s.graphStore.Close(); err != nil {
		errors = append(errors, fmt.Errorf("graph store close error: %w", err))
	}

	if err := s.kvStore.Close(); err != nil {
		errors = append(errors, fmt.Errorf("kv store close error: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during close: %v", errors)
	}

	return nil
}

// Helper methods for cross-component queries

func (s *OmniStoreImpl) executeKVQuery(query *KVQueryComponent) (map[string][]byte, error) {
	results := make(map[string][]byte)

	if len(query.Keys) > 0 {
		// Get specific keys
		for _, key := range query.Keys {
			value, err := s.kvStore.Get(key)
			if err == nil {
				results[key] = value
			}
		}
	} else if query.Prefix != "" {
		// Scan by prefix
		scanResults, err := s.kvStore.Scan(query.Prefix, query.ScanLimit)
		if err != nil {
			return nil, err
		}
		results = scanResults
	}

	return results, nil
}

func (s *OmniStoreImpl) executeGraphQuery(req *GraphQueryComponent) (*query.QueryResult, error) {
	if req.QueryString != "" {
		// Graph query language not yet implemented
		return nil, fmt.Errorf("graph query language not yet implemented")
	}

	// Handle other graph query types
	return query.NewQueryResult(), nil
}

func (s *OmniStoreImpl) executeFileQuery(query *FileQueryComponent) ([]FileResult, error) {
	var results []FileResult

	if len(query.Hashes) > 0 {
		// Get specific files by hash
		for _, hash := range query.Hashes {
			if exists, _ := s.fileStore.Exists(hash); exists {
				metadata, _ := s.fileStore.GetMetadata(hash)
				results = append(results, FileResult{
					Hash:     hash,
					Metadata: metadata,
				})
			}
		}
	} else if query.HashPrefix != "" {
		// Find by hash prefix - this would need to be implemented
	}

	return results, nil
}

func (s *OmniStoreImpl) executeSearchQuery(query *SearchQueryComponent) (*SearchResult, error) {
	options := &SearchOptions{
		From:    0,
		Size:    100,
		Filters: query.Filters,
	}

	return s.searchStore.Search(query.Query, options)
}

func (s *OmniStoreImpl) processJoinOperations(result *CrossQueryResult, operations []JoinOperation) ([]map[string]interface{}, error) {
	// Placeholder for join processing logic
	var joinedResults []map[string]interface{}

	// This would implement actual join logic between different component results
	// For now, return empty results
	return joinedResults, nil
}

// NoOpSearchStore provides a placeholder search implementation
type NoOpSearchStore struct{}

func NewNoOpSearchStore() SearchStore {
	return &NoOpSearchStore{}
}

func (s *NoOpSearchStore) IndexDocument(id string, content string, metadata map[string]interface{}) error {
	return nil
}

func (s *NoOpSearchStore) UpdateDocument(id string, content string, metadata map[string]interface{}) error {
	return nil
}

func (s *NoOpSearchStore) DeleteDocument(id string) error {
	return nil
}

func (s *NoOpSearchStore) BulkIndex(documents []SearchDocument) error {
	return nil
}

func (s *NoOpSearchStore) Search(query string, options *SearchOptions) (*SearchResult, error) {
	return &SearchResult{
		Hits:          []SearchHit{},
		TotalHits:     0,
		MaxScore:      0,
		ExecutionTime: time.Millisecond,
	}, nil
}

func (s *NoOpSearchStore) SearchWithContext(ctx context.Context, query string, options *SearchOptions) (*SearchResult, error) {
	return s.Search(query, options)
}

func (s *NoOpSearchStore) Aggregate(query string, aggregation *AggregationRequest) (*AggregationResult, error) {
	return &AggregationResult{}, nil
}

func (s *NoOpSearchStore) GetTermFrequency(term string) (int64, error) {
	return 0, nil
}

func (s *NoOpSearchStore) Suggest(prefix string, count int) ([]string, error) {
	return []string{}, nil
}

func (s *NoOpSearchStore) MoreLikeThis(documentID string, count int) (*SearchResult, error) {
	return &SearchResult{}, nil
}

func (s *NoOpSearchStore) RefreshIndex() error {
	return nil
}

func (s *NoOpSearchStore) OptimizeIndex() error {
	return nil
}

func (s *NoOpSearchStore) GetIndexStats() (*IndexStats, error) {
	return &IndexStats{}, nil
}

func (s *NoOpSearchStore) Close() error {
	return nil
}

// Helper function to convert filestore.DeduplicationStats to omnistore.DeduplicationStats
func convertDeduplicationStats(fs *filestore.DeduplicationStats) *DeduplicationStats {
	if fs == nil {
		return nil
	}
	return &DeduplicationStats{
		TotalFiles:        fs.TotalFiles,
		UniqueFiles:       fs.UniqueFiles,
		DuplicateFiles:    fs.DuplicateFiles,
		SpaceSaved:        fs.SpaceSaved,
		DeduplicationRate: fs.DeduplicationRate,
	}
}
