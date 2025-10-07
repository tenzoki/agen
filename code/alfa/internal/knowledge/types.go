package knowledge

import "time"

// AgentInfo contains extracted agent information
type AgentInfo struct {
	Name         string
	Type         string // kebab-case from pool.yaml
	Directory    string
	Intent       string
	Usage        string
	Capabilities []string
	Config       map[string]string
	Description  string // from pool.yaml
	Operator     string // call/spawn/await
	Binary       string
}

// CellInfo contains extracted cell information
type CellInfo struct {
	ID          string
	Category    string // pipelines/services/analysis/synthesis
	Description string
	Agents      []string
	Purpose     string
	InputOutput string
	UseCases    []string
	FilePath    string
}

// PoolInfo contains pool.yaml information
type PoolInfo struct {
	AgentTypes map[string]PoolAgent
}

// PoolAgent represents an agent type from pool.yaml
type PoolAgent struct {
	Type         string
	Binary       string
	Operator     string
	Capabilities []string
	Description  string
}

// Manifest tracks extraction metadata
type Manifest struct {
	Timestamp   time.Time
	Sources     []SourceFile
	AgentCount  int
	CellCount   int
	GeneratedBy string
}

// SourceFile tracks a source file and its modification time
type SourceFile struct {
	Path    string
	ModTime time.Time
}

// KnowledgeBase is the main interface for knowledge extraction
type KnowledgeBase interface {
	IsStale() bool
	Extract() error
	LoadAgentsDoc() (string, error)
	LoadCellsDoc() (string, error)
	LoadCapabilitiesDoc() (string, error)
}
