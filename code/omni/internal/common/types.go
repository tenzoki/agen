package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	ErrInvalidID         = errors.New("invalid ID: cannot be empty")
	ErrInvalidType       = errors.New("invalid type: cannot be empty")
	ErrInvalidVertex     = errors.New("invalid vertex")
	ErrInvalidEdge       = errors.New("invalid edge")
	ErrVertexNotFound    = errors.New("vertex not found")
	ErrEdgeNotFound      = errors.New("edge not found")
	ErrDuplicateVertex   = errors.New("vertex already exists")
	ErrDuplicateEdge     = errors.New("edge already exists")
	ErrCircularReference = errors.New("circular reference detected")
	ErrInvalidDirection  = errors.New("invalid traversal direction")
)

type Direction int

const (
	Incoming Direction = iota
	Outgoing
	Both
)

func (d Direction) String() string {
	switch d {
	case Incoming:
		return "incoming"
	case Outgoing:
		return "outgoing"
	case Both:
		return "both"
	default:
		return "unknown"
	}
}

type TraversalStrategy int

const (
	BFS TraversalStrategy = iota
	DFS
)

func (ts TraversalStrategy) String() string {
	switch ts {
	case BFS:
		return "bfs"
	case DFS:
		return "dfs"
	default:
		return "unknown"
	}
}

type Vertex struct {
	ID         string                 `json:"id" msgpack:"id"`
	Type       string                 `json:"type" msgpack:"type"`
	Properties map[string]interface{} `json:"properties" msgpack:"properties"`
	CreatedAt  time.Time              `json:"created_at" msgpack:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" msgpack:"updated_at"`
	Version    uint64                 `json:"version" msgpack:"version"`
}

func NewVertex(id, vertexType string) *Vertex {
	now := time.Now().UTC()
	return &Vertex{
		ID:         id,
		Type:       vertexType,
		Properties: make(map[string]interface{}),
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

func (v *Vertex) SetProperty(key string, value interface{}) {
	if v.Properties == nil {
		v.Properties = make(map[string]interface{})
	}
	v.Properties[key] = value
	v.UpdatedAt = time.Now().UTC()
	v.Version++
}

func (v *Vertex) GetProperty(key string) (interface{}, bool) {
	if v.Properties == nil {
		return nil, false
	}
	value, exists := v.Properties[key]
	return value, exists
}

func (v *Vertex) Validate() error {
	if v.ID == "" {
		return fmt.Errorf("%w: vertex ID", ErrInvalidID)
	}
	if v.Type == "" {
		return fmt.Errorf("%w: vertex type", ErrInvalidType)
	}
	if v.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at cannot be zero", ErrInvalidVertex)
	}
	if v.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: updated_at cannot be zero", ErrInvalidVertex)
	}
	if v.Version == 0 {
		return fmt.Errorf("%w: version must be greater than 0", ErrInvalidVertex)
	}
	return nil
}

func (v *Vertex) MarshalBinary() ([]byte, error) {
	data := struct {
		ID         string                 `msgpack:"id"`
		Type       string                 `msgpack:"type"`
		Properties map[string]interface{} `msgpack:"properties"`
		CreatedAt  time.Time              `msgpack:"created_at"`
		UpdatedAt  time.Time              `msgpack:"updated_at"`
		Version    uint64                 `msgpack:"version"`
	}{
		ID:         v.ID,
		Type:       v.Type,
		Properties: v.Properties,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
		Version:    v.Version,
	}
	return msgpack.Marshal(data)
}

func (v *Vertex) UnmarshalBinary(data []byte) error {
	temp := struct {
		ID         string                 `msgpack:"id"`
		Type       string                 `msgpack:"type"`
		Properties map[string]interface{} `msgpack:"properties"`
		CreatedAt  time.Time              `msgpack:"created_at"`
		UpdatedAt  time.Time              `msgpack:"updated_at"`
		Version    uint64                 `msgpack:"version"`
	}{}

	if err := msgpack.Unmarshal(data, &temp); err != nil {
		return err
	}

	v.ID = temp.ID
	v.Type = temp.Type
	v.Properties = temp.Properties
	v.CreatedAt = temp.CreatedAt
	v.UpdatedAt = temp.UpdatedAt
	v.Version = temp.Version

	return nil
}

func (v *Vertex) Clone() *Vertex {
	clone := &Vertex{
		ID:        v.ID,
		Type:      v.Type,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
		Version:   v.Version,
	}

	if v.Properties != nil {
		clone.Properties = make(map[string]interface{})
		for k, val := range v.Properties {
			if val != v {
				clone.Properties[k] = val
			}
		}
	}

	return clone
}

type Edge struct {
	ID         string                 `json:"id" msgpack:"id"`
	Type       string                 `json:"type" msgpack:"type"`
	FromVertex string                 `json:"from_vertex" msgpack:"from_vertex"`
	ToVertex   string                 `json:"to_vertex" msgpack:"to_vertex"`
	Properties map[string]interface{} `json:"properties" msgpack:"properties"`
	Weight     float64                `json:"weight,omitempty" msgpack:"weight"`
	CreatedAt  time.Time              `json:"created_at" msgpack:"created_at"`
	Version    uint64                 `json:"version" msgpack:"version"`
}

func NewEdge(id, edgeType, fromVertex, toVertex string) *Edge {
	return &Edge{
		ID:         id,
		Type:       edgeType,
		FromVertex: fromVertex,
		ToVertex:   toVertex,
		Properties: make(map[string]interface{}),
		Weight:     1.0,
		CreatedAt:  time.Now().UTC(),
		Version:    1,
	}
}

func (e *Edge) SetProperty(key string, value interface{}) {
	if e.Properties == nil {
		e.Properties = make(map[string]interface{})
	}
	e.Properties[key] = value
	e.Version++
}

func (e *Edge) GetProperty(key string) (interface{}, bool) {
	if e.Properties == nil {
		return nil, false
	}
	value, exists := e.Properties[key]
	return value, exists
}

func (e *Edge) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("%w: edge ID", ErrInvalidID)
	}
	if e.Type == "" {
		return fmt.Errorf("%w: edge type", ErrInvalidType)
	}
	if e.FromVertex == "" {
		return fmt.Errorf("%w: from_vertex cannot be empty", ErrInvalidEdge)
	}
	if e.ToVertex == "" {
		return fmt.Errorf("%w: to_vertex cannot be empty", ErrInvalidEdge)
	}
	if e.FromVertex == e.ToVertex {
		return fmt.Errorf("%w: self-loop detected", ErrCircularReference)
	}
	if e.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at cannot be zero", ErrInvalidEdge)
	}
	if e.Version == 0 {
		return fmt.Errorf("%w: version must be greater than 0", ErrInvalidEdge)
	}
	return nil
}

func (e *Edge) MarshalBinary() ([]byte, error) {
	data := struct {
		ID         string                 `msgpack:"id"`
		Type       string                 `msgpack:"type"`
		FromVertex string                 `msgpack:"from_vertex"`
		ToVertex   string                 `msgpack:"to_vertex"`
		Properties map[string]interface{} `msgpack:"properties"`
		Weight     float64                `msgpack:"weight"`
		CreatedAt  time.Time              `msgpack:"created_at"`
		Version    uint64                 `msgpack:"version"`
	}{
		ID:         e.ID,
		Type:       e.Type,
		FromVertex: e.FromVertex,
		ToVertex:   e.ToVertex,
		Properties: e.Properties,
		Weight:     e.Weight,
		CreatedAt:  e.CreatedAt,
		Version:    e.Version,
	}
	return msgpack.Marshal(data)
}

func (e *Edge) UnmarshalBinary(data []byte) error {
	temp := struct {
		ID         string                 `msgpack:"id"`
		Type       string                 `msgpack:"type"`
		FromVertex string                 `msgpack:"from_vertex"`
		ToVertex   string                 `msgpack:"to_vertex"`
		Properties map[string]interface{} `msgpack:"properties"`
		Weight     float64                `msgpack:"weight"`
		CreatedAt  time.Time              `msgpack:"created_at"`
		Version    uint64                 `msgpack:"version"`
	}{}

	if err := msgpack.Unmarshal(data, &temp); err != nil {
		return err
	}

	e.ID = temp.ID
	e.Type = temp.Type
	e.FromVertex = temp.FromVertex
	e.ToVertex = temp.ToVertex
	e.Properties = temp.Properties
	e.Weight = temp.Weight
	e.CreatedAt = temp.CreatedAt
	e.Version = temp.Version

	return nil
}

func (e *Edge) Clone() *Edge {
	clone := &Edge{
		ID:         e.ID,
		Type:       e.Type,
		FromVertex: e.FromVertex,
		ToVertex:   e.ToVertex,
		Weight:     e.Weight,
		CreatedAt:  e.CreatedAt,
		Version:    e.Version,
	}

	if e.Properties != nil {
		clone.Properties = make(map[string]interface{})
		for k, val := range e.Properties {
			if val != e {
				clone.Properties[k] = val
			}
		}
	}

	return clone
}

type GraphMetadata struct {
	Name        string    `json:"name" msgpack:"name"`
	VertexCount int64     `json:"vertex_count" msgpack:"vertex_count"`
	EdgeCount   int64     `json:"edge_count" msgpack:"edge_count"`
	CreatedAt   time.Time `json:"created_at" msgpack:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" msgpack:"updated_at"`
	Schema      *Schema   `json:"schema,omitempty" msgpack:"schema"`
}

func NewGraphMetadata(name string) *GraphMetadata {
	now := time.Now().UTC()
	return &GraphMetadata{
		Name:        name,
		VertexCount: 0,
		EdgeCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (gm *GraphMetadata) IncrementVertexCount() {
	gm.VertexCount++
	gm.UpdatedAt = time.Now().UTC()
}

func (gm *GraphMetadata) DecrementVertexCount() {
	if gm.VertexCount > 0 {
		gm.VertexCount--
	}
	gm.UpdatedAt = time.Now().UTC()
}

func (gm *GraphMetadata) IncrementEdgeCount() {
	gm.EdgeCount++
	gm.UpdatedAt = time.Now().UTC()
}

func (gm *GraphMetadata) DecrementEdgeCount() {
	if gm.EdgeCount > 0 {
		gm.EdgeCount--
	}
	gm.UpdatedAt = time.Now().UTC()
}

func (gm *GraphMetadata) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(gm)
}

func (gm *GraphMetadata) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, gm)
}

type Schema struct {
	VertexTypes map[string]*VertexType `json:"vertex_types" msgpack:"vertex_types"`
	EdgeTypes   map[string]*EdgeType   `json:"edge_types" msgpack:"edge_types"`
	Constraints []Constraint           `json:"constraints" msgpack:"constraints"`
}

type VertexType struct {
	Name       string                  `json:"name" msgpack:"name"`
	Properties map[string]PropertyType `json:"properties" msgpack:"properties"`
	Required   []string                `json:"required" msgpack:"required"`
}

type EdgeType struct {
	Name        string                  `json:"name" msgpack:"name"`
	FromTypes   []string                `json:"from_types" msgpack:"from_types"`
	ToTypes     []string                `json:"to_types" msgpack:"to_types"`
	Properties  map[string]PropertyType `json:"properties" msgpack:"properties"`
	Cardinality Cardinality             `json:"cardinality" msgpack:"cardinality"`
}

type PropertyType struct {
	Type         string      `json:"type" msgpack:"type"`
	Required     bool        `json:"required" msgpack:"required"`
	DefaultValue interface{} `json:"default_value,omitempty" msgpack:"default_value"`
}

type Cardinality int

const (
	OneToOne Cardinality = iota
	OneToMany
	ManyToOne
	ManyToMany
)

func (c Cardinality) String() string {
	switch c {
	case OneToOne:
		return "one_to_one"
	case OneToMany:
		return "one_to_many"
	case ManyToOne:
		return "many_to_one"
	case ManyToMany:
		return "many_to_many"
	default:
		return "unknown"
	}
}

type Constraint struct {
	Type        string      `json:"type" msgpack:"type"`
	Description string      `json:"description" msgpack:"description"`
	Parameters  interface{} `json:"parameters" msgpack:"parameters"`
}
