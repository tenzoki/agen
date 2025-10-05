package omnistore

import (
	"context"
	"time"

	"github.com/tenzoki/agen/omni/internal/common"
	"github.com/tenzoki/agen/omni/internal/graph"
	"github.com/tenzoki/agen/omni/internal/kv"
	"github.com/tenzoki/agen/omni/internal/query"
	"github.com/tenzoki/agen/omni/internal/transaction"
	"github.com/tenzoki/agen/omni/internal/filestore"
)

// OmniStore provides a unified interface for KV, Graph, Files, and Search operations
type OmniStore interface {
	// Core subsystem access
	KV() kv.KVStore
	Graph() graph.GraphStore
	Files() filestore.FileStore
	Search() SearchStore

	// Convenience methods for common operations
	ListKVKeys(prefix string, limit int) ([]string, error)

	// Transaction support
	BeginTransaction(config *transaction.TransactionConfig) (transaction.GraphTx, error)
	BeginTransactionWithContext(ctx context.Context, config *transaction.TransactionConfig) (transaction.GraphTx, error)
	ExecuteTransaction(fn func(tx transaction.GraphTx) error) error
	ExecuteTransactionWithConfig(config *transaction.TransactionConfig, fn func(tx transaction.GraphTx) error) error

	// Query language support
	Query(queryString string) (*query.QueryResult, error)
	QueryWithContext(ctx context.Context, queryString string) (*query.QueryResult, error)

	// Unified operations (cross-component)
	ExecuteCrossQuery(req *CrossQueryRequest) (*CrossQueryResult, error)

	// Configuration and lifecycle
	GetConfig() *Config
	UpdateConfig(config *Config) error

	// Statistics and monitoring
	GetStats() (*OmniStoreStats, error)
	GetHealth() (*HealthStatus, error)

	// Backup and maintenance
	Backup(path string) error
	Restore(path string) error
	Compact() error

	// Graceful shutdown
	Close() error
}

// Note: FileStore interface is defined in pkg/filestore package

// SearchStore defines the interface for full-text search operations
type SearchStore interface {
	// Document indexing
	IndexDocument(id string, content string, metadata map[string]interface{}) error
	UpdateDocument(id string, content string, metadata map[string]interface{}) error
	DeleteDocument(id string) error
	BulkIndex(documents []SearchDocument) error

	// Search operations
	Search(query string, options *SearchOptions) (*SearchResult, error)
	SearchWithContext(ctx context.Context, query string, options *SearchOptions) (*SearchResult, error)

	// Aggregations and analytics
	Aggregate(query string, aggregation *AggregationRequest) (*AggregationResult, error)
	GetTermFrequency(term string) (int64, error)

	// Suggestions and recommendations
	Suggest(prefix string, count int) ([]string, error)
	MoreLikeThis(documentID string, count int) (*SearchResult, error)

	// Index management
	RefreshIndex() error
	OptimizeIndex() error
	GetIndexStats() (*IndexStats, error)

	Close() error
}

// Configuration types
type Config struct {
	DataDir string `json:"data_dir"`

	// Component configurations are handled by the underlying storage layer
	Files  *FileStoreConfig `json:"files"`
	Search *SearchConfig    `json:"search"`

	// Transaction settings
	Transaction *transaction.TransactionConfig `json:"transaction"`

	// Performance settings
	Performance *PerformanceConfig `json:"performance"`

	// Security settings
	Security *SecurityConfig `json:"security"`
}

type FileStoreConfig struct {
	StorageDir          string `json:"storage_dir"`
	EnableEncryption    bool   `json:"enable_encryption"`
	EncryptionKey       string `json:"encryption_key,omitempty"`
	EnableCompression   bool   `json:"enable_compression"`
	CompressionLevel    int    `json:"compression_level"`
	EnableDeduplication bool   `json:"enable_deduplication"`
	MaxFileSize         int64  `json:"max_file_size"`
	IndexingEnabled     bool   `json:"indexing_enabled"`
}

type SearchConfig struct {
	IndexDir        string   `json:"index_dir"`
	DefaultAnalyzer string   `json:"default_analyzer"`
	Languages       []string `json:"languages"`
	MaxResults      int      `json:"max_results"`
	IndexBatchSize  int      `json:"index_batch_size"`
}

type PerformanceConfig struct {
	MaxConnections    int           `json:"max_connections"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	QueryTimeout      time.Duration `json:"query_timeout"`
	CacheSize         int64         `json:"cache_size"`
	WorkerPoolSize    int           `json:"worker_pool_size"`
}

type SecurityConfig struct {
	EnableAuth    bool     `json:"enable_auth"`
	AuthProviders []string `json:"auth_providers"`
	TLSEnabled    bool     `json:"tls_enabled"`
	CertFile      string   `json:"cert_file,omitempty"`
	KeyFile       string   `json:"key_file,omitempty"`
}

// File storage types
type FileSearchOptions struct {
	ContentTypes []string               `json:"content_types,omitempty"`
	SizeRange    *SizeRange             `json:"size_range,omitempty"`
	DateRange    *DateRange             `json:"date_range,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	HashPrefix   string                 `json:"hash_prefix,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
}

type FileSearchResult struct {
	Files         []FileResult  `json:"files"`
	TotalCount    int64         `json:"total_count"`
	ExecutionTime time.Duration `json:"execution_time"`
}

type FileResult struct {
	Hash        string            `json:"hash"`
	Key         string            `json:"key,omitempty"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Compressed  bool              `json:"compressed"`
	Encrypted   bool              `json:"encrypted"`
}

type DeduplicationStats struct {
	TotalFiles        int64   `json:"total_files"`
	UniqueFiles       int64   `json:"unique_files"`
	DuplicateFiles    int64   `json:"duplicate_files"`
	SpaceSaved        int64   `json:"space_saved"`
	DeduplicationRate float64 `json:"deduplication_rate"`
}

// Search types
type SearchDocument struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

type SearchOptions struct {
	From      int                    `json:"from,omitempty"`
	Size      int                    `json:"size,omitempty"`
	Sort      []string               `json:"sort,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Facets    []string               `json:"facets,omitempty"`
	Highlight *HighlightOptions      `json:"highlight,omitempty"`
	MinScore  float64                `json:"min_score,omitempty"`
	Analyzer  string                 `json:"analyzer,omitempty"`
}

type HighlightOptions struct {
	Fields    []string `json:"fields"`
	PreTag    string   `json:"pre_tag,omitempty"`
	PostTag   string   `json:"post_tag,omitempty"`
	MaxLength int      `json:"max_length,omitempty"`
}

type SearchResult struct {
	Hits          []SearchHit        `json:"hits"`
	TotalHits     int64              `json:"total_hits"`
	MaxScore      float64            `json:"max_score"`
	Facets        map[string][]Facet `json:"facets,omitempty"`
	Suggestions   []string           `json:"suggestions,omitempty"`
	ExecutionTime time.Duration      `json:"execution_time"`
}

type SearchHit struct {
	ID         string                 `json:"id"`
	Score      float64                `json:"score"`
	Source     map[string]interface{} `json:"source"`
	Highlights map[string][]string    `json:"highlights,omitempty"`
}

type Facet struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// Cross-component query types
type CrossQueryRequest struct {
	// Query components
	KVQuery     *KVQueryComponent     `json:"kv_query,omitempty"`
	GraphQuery  *GraphQueryComponent  `json:"graph_query,omitempty"`
	FileQuery   *FileQueryComponent   `json:"file_query,omitempty"`
	SearchQuery *SearchQueryComponent `json:"search_query,omitempty"`

	// Cross-component operations
	JoinOperations []JoinOperation `json:"join_operations,omitempty"`

	// Query options
	Limit   int                    `json:"limit,omitempty"`
	Offset  int                    `json:"offset,omitempty"`
	OrderBy []string               `json:"order_by,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

type KVQueryComponent struct {
	Prefix    string            `json:"prefix,omitempty"`
	Keys      []string          `json:"keys,omitempty"`
	ScanLimit int               `json:"scan_limit,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

type GraphQueryComponent struct {
	QueryString string                 `json:"query_string,omitempty"`
	StartVertex string                 `json:"start_vertex,omitempty"`
	VertexTypes []string               `json:"vertex_types,omitempty"`
	EdgeTypes   []string               `json:"edge_types,omitempty"`
	MaxDepth    int                    `json:"max_depth,omitempty"`
	Direction   common.Direction       `json:"direction,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
}

type FileQueryComponent struct {
	ContentTypes []string               `json:"content_types,omitempty"`
	SizeRange    *SizeRange             `json:"size_range,omitempty"`
	DateRange    *DateRange             `json:"date_range,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	HashPrefix   string                 `json:"hash_prefix,omitempty"`
	Hashes       []string               `json:"hashes,omitempty"`
}

type SearchQueryComponent struct {
	Query        string                 `json:"query"`
	Fields       []string               `json:"fields,omitempty"`
	Boost        map[string]float64     `json:"boost,omitempty"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Facets       []string               `json:"facets,omitempty"`
	Highlighting bool                   `json:"highlighting,omitempty"`
}

// Join operations for cross-component queries
type JoinOperation struct {
	Type        JoinType `json:"type"`
	LeftSource  string   `json:"left_source"` // kv, graph, files, search
	RightSource string   `json:"right_source"`
	LeftField   string   `json:"left_field"`
	RightField  string   `json:"right_field"`
}

type JoinType int

const (
	InnerJoin JoinType = iota
	LeftJoin
	RightJoin
	FullJoin
)

// Range types
type SizeRange struct {
	Min int64 `json:"min,omitempty"`
	Max int64 `json:"max,omitempty"`
}

type DateRange struct {
	Start time.Time `json:"start,omitempty"`
	End   time.Time `json:"end,omitempty"`
}

// CrossQueryResult contains results from multiple components
type CrossQueryResult struct {
	KVResults     map[string][]byte  `json:"kv_results,omitempty"`
	GraphResults  *query.QueryResult `json:"graph_results,omitempty"`
	FileResults   []FileResult       `json:"file_results,omitempty"`
	SearchResults *SearchResult      `json:"search_results,omitempty"`

	// Aggregated results
	JoinedResults []map[string]interface{} `json:"joined_results,omitempty"`

	// Metadata
	ExecutionTime  time.Duration            `json:"execution_time"`
	ComponentTimes map[string]time.Duration `json:"component_times"`
	TotalResults   int                      `json:"total_results"`
}

// Statistics types
type OmniStoreStats struct {
	KV     *kv.KVStats       `json:"kv_stats"`
	Graph  *graph.GraphStats `json:"graph_stats"`
	Files  *FileStoreStats   `json:"file_stats"`
	Search *IndexStats       `json:"search_stats"`

	// Overall stats
	TotalStorageSize   int64         `json:"total_storage_size"`
	TotalOperations    int64         `json:"total_operations"`
	Uptime             time.Duration `json:"uptime"`
	ActiveTransactions int           `json:"active_transactions"`
}

type FileStoreStats struct {
	FileCount        int64               `json:"file_count"`
	TotalSize        int64               `json:"total_size"`
	CompressedSize   int64               `json:"compressed_size"`
	AverageSize      float64             `json:"average_size"`
	ContentTypes     map[string]int64    `json:"content_types"`
	LastAccess       time.Time           `json:"last_access"`
	IndexedFiles     int64               `json:"indexed_files"`
	CompressionRatio float64             `json:"compression_ratio"`
	Deduplication    *DeduplicationStats `json:"deduplication,omitempty"`
}

type IndexStats struct {
	DocumentCount    int64     `json:"document_count"`
	IndexSize        int64     `json:"index_size"`
	TermCount        int64     `json:"term_count"`
	AverageDocSize   float64   `json:"average_doc_size"`
	LastIndexUpdate  time.Time `json:"last_index_update"`
	QueriesPerSecond float64   `json:"queries_per_second"`
}

// Aggregation types
type AggregationRequest struct {
	Name     string               `json:"name"`
	Type     string               `json:"type"` // terms, range, date_histogram, etc.
	Field    string               `json:"field"`
	Size     int                  `json:"size,omitempty"`
	Ranges   []AggregationRange   `json:"ranges,omitempty"`
	Interval string               `json:"interval,omitempty"`
	Nested   []AggregationRequest `json:"nested,omitempty"`
}

type AggregationRange struct {
	From interface{} `json:"from,omitempty"`
	To   interface{} `json:"to,omitempty"`
	Key  string      `json:"key,omitempty"`
}

type AggregationResult struct {
	Name    string                       `json:"name"`
	Buckets []AggregationBucket          `json:"buckets,omitempty"`
	Value   interface{}                  `json:"value,omitempty"`
	Nested  map[string]AggregationResult `json:"nested,omitempty"`
}

type AggregationBucket struct {
	Key      interface{}                  `json:"key"`
	DocCount int64                        `json:"doc_count"`
	Nested   map[string]AggregationResult `json:"nested,omitempty"`
}

// Health and monitoring
type HealthStatus struct {
	Status     HealthState                `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     time.Duration              `json:"uptime"`
}

type HealthState int

const (
	Healthy HealthState = iota
	Degraded
	Unhealthy
)

func (hs HealthState) String() string {
	switch hs {
	case Healthy:
		return "healthy"
	case Degraded:
		return "degraded"
	case Unhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

type ComponentHealth struct {
	Status    HealthState            `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
	LastCheck time.Time              `json:"last_check"`
}

// Default configuration functions
func DefaultConfig() *Config {
	return &Config{
		DataDir: "./data",
		Files:   DefaultFileStoreConfig(),
		Search:  DefaultSearchConfig(),
		Transaction: &transaction.TransactionConfig{
			IsolationLevel: transaction.ReadCommitted,
			Timeout:        30 * time.Second,
		},
		Performance: &PerformanceConfig{
			MaxConnections:    100,
			ConnectionTimeout: 10 * time.Second,
			QueryTimeout:      30 * time.Second,
			CacheSize:         100 * 1024 * 1024, // 100MB
			WorkerPoolSize:    10,
		},
		Security: &SecurityConfig{
			EnableAuth:    false,
			AuthProviders: []string{},
			TLSEnabled:    false,
		},
	}
}

func DefaultFileStoreConfig() *FileStoreConfig {
	return &FileStoreConfig{
		StorageDir:          "files",
		EnableEncryption:    false,
		EnableCompression:   true,
		CompressionLevel:    6,
		EnableDeduplication: true,
		MaxFileSize:         100 * 1024 * 1024, // 100MB
		IndexingEnabled:     true,
	}
}

func DefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		IndexDir:        "search_index",
		DefaultAnalyzer: "standard",
		Languages:       []string{"en"},
		MaxResults:      1000,
		IndexBatchSize:  100,
	}
}
