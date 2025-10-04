package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/agen/omni/internal/common"
)

// QueryStep represents a single step in a graph traversal query
type QueryStep interface {
	String() string
	Execute(traverser *Traverser) (*Traverser, error)
}

// Traverser holds the current state of query execution
type Traverser struct {
	Vertices []*common.Vertex
	Edges    []*common.Edge
	Values   []interface{} // For computed values
	Count    int64
}

// NewTraverser creates a new traverser
func NewTraverser() *Traverser {
	return &Traverser{
		Vertices: make([]*common.Vertex, 0),
		Edges:    make([]*common.Edge, 0),
		Values:   make([]interface{}, 0),
		Count:    0,
	}
}

// Query represents a complete graph query
type Query struct {
	Steps []QueryStep
}

// NewQuery creates a new empty query
func NewQuery() *Query {
	return &Query{
		Steps: make([]QueryStep, 0),
	}
}

// AddStep adds a step to the query
func (q *Query) AddStep(step QueryStep) *Query {
	q.Steps = append(q.Steps, step)
	return q
}

// String returns the string representation of the query
func (q *Query) String() string {
	if len(q.Steps) == 0 {
		return "g"
	}

	parts := make([]string, len(q.Steps)+1)
	parts[0] = "g"
	for i, step := range q.Steps {
		parts[i+1] = step.String()
	}
	return strings.Join(parts, ".")
}

// VStep - Start traversal from vertices
type VStep struct {
	IDs []string // Optional vertex IDs to start from
}

func (s *VStep) String() string {
	if len(s.IDs) == 0 {
		return "V()"
	}
	ids := make([]string, len(s.IDs))
	for i, id := range s.IDs {
		ids[i] = fmt.Sprintf("'%s'", id)
	}
	return fmt.Sprintf("V(%s)", strings.Join(ids, ", "))
}

func (s *VStep) Execute(traverser *Traverser) (*Traverser, error) {
	// Implementation will be added in execution engine
	return traverser, nil
}

// EStep - Start traversal from edges
type EStep struct {
	IDs []string // Optional edge IDs to start from
}

func (s *EStep) String() string {
	if len(s.IDs) == 0 {
		return "E()"
	}
	ids := make([]string, len(s.IDs))
	for i, id := range s.IDs {
		ids[i] = fmt.Sprintf("'%s'", id)
	}
	return fmt.Sprintf("E(%s)", strings.Join(ids, ", "))
}

func (s *EStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// OutStep - Traverse outgoing edges
type OutStep struct {
	EdgeLabels []string // Optional edge labels to filter by
}

func (s *OutStep) String() string {
	if len(s.EdgeLabels) == 0 {
		return "out()"
	}
	labels := make([]string, len(s.EdgeLabels))
	for i, label := range s.EdgeLabels {
		labels[i] = fmt.Sprintf("'%s'", label)
	}
	return fmt.Sprintf("out(%s)", strings.Join(labels, ", "))
}

func (s *OutStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// InStep - Traverse incoming edges
type InStep struct {
	EdgeLabels []string
}

func (s *InStep) String() string {
	if len(s.EdgeLabels) == 0 {
		return "in()"
	}
	labels := make([]string, len(s.EdgeLabels))
	for i, label := range s.EdgeLabels {
		labels[i] = fmt.Sprintf("'%s'", label)
	}
	return fmt.Sprintf("in(%s)", strings.Join(labels, ", "))
}

func (s *InStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// BothStep - Traverse both incoming and outgoing edges
type BothStep struct {
	EdgeLabels []string
}

func (s *BothStep) String() string {
	if len(s.EdgeLabels) == 0 {
		return "both()"
	}
	labels := make([]string, len(s.EdgeLabels))
	for i, label := range s.EdgeLabels {
		labels[i] = fmt.Sprintf("'%s'", label)
	}
	return fmt.Sprintf("both(%s)", strings.Join(labels, ", "))
}

func (s *BothStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// HasStep - Filter by property existence or value
type HasStep struct {
	Property string
	Value    interface{} // nil means check existence only
}

func (s *HasStep) String() string {
	if s.Value == nil {
		return fmt.Sprintf("has('%s')", s.Property)
	}
	return fmt.Sprintf("has('%s', %v)", s.Property, s.formatValue(s.Value))
}

func (s *HasStep) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", v)
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (s *HasStep) Execute(traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	// Filter vertices
	for _, vertex := range traverser.Vertices {
		if s.vertexMatches(vertex) {
			result.Vertices = append(result.Vertices, vertex)
		}
	}

	// Filter edges
	for _, edge := range traverser.Edges {
		if s.edgeMatches(edge) {
			result.Edges = append(result.Edges, edge)
		}
	}

	return result, nil
}

// vertexMatches checks if a vertex matches this HasStep
func (s *HasStep) vertexMatches(vertex *common.Vertex) bool {
	value, exists := vertex.Properties[s.Property]

	if s.Value == nil {
		// Just checking for property existence
		return exists
	}

	if !exists {
		return false
	}

	// Compare values - handle different numeric types
	if s.compareValues(value, s.Value) {
		return true
	}

	return false
}

// edgeMatches checks if an edge matches this HasStep
func (s *HasStep) edgeMatches(edge *common.Edge) bool {
	value, exists := edge.Properties[s.Property]

	if s.Value == nil {
		// Just checking for property existence
		return exists
	}

	if !exists {
		return false
	}

	// Compare values
	return s.compareValues(value, s.Value)
}

// compareValues compares two values for equality
func (s *HasStep) compareValues(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Direct equality check
	if a == b {
		return true
	}

	// Try numeric conversions for common type mismatches
	if s.compareNumeric(a, b) {
		return true
	}

	// Use reflection for deeper comparison
	return reflect.DeepEqual(a, b)
}

// compareNumeric handles numeric type conversions
func (s *HasStep) compareNumeric(a, b interface{}) bool {
	// Convert both to float64 for comparison if they're numeric
	aFloat, aOk := s.toFloat64(a)
	bFloat, bOk := s.toFloat64(b)
	if aOk && bOk {
		return aFloat == bFloat
	}
	return false
}

// toFloat64 converts various numeric types to float64
func (s *HasStep) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// HasLabelStep - Filter by vertex/edge label (type)
type HasLabelStep struct {
	Labels []string
}

func (s *HasLabelStep) String() string {
	labels := make([]string, len(s.Labels))
	for i, label := range s.Labels {
		labels[i] = fmt.Sprintf("'%s'", label)
	}
	return fmt.Sprintf("hasLabel(%s)", strings.Join(labels, ", "))
}

func (s *HasLabelStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// CountStep - Count the current elements
type CountStep struct{}

func (s *CountStep) String() string {
	return "count()"
}

func (s *CountStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// LimitStep - Limit the number of results
type LimitStep struct {
	Count int
}

func (s *LimitStep) String() string {
	return fmt.Sprintf("limit(%d)", s.Count)
}

func (s *LimitStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// ValuesStep - Extract property values
type ValuesStep struct {
	Properties []string
}

func (s *ValuesStep) String() string {
	if len(s.Properties) == 1 {
		return fmt.Sprintf("values('%s')", s.Properties[0])
	}
	props := make([]string, len(s.Properties))
	for i, prop := range s.Properties {
		props[i] = fmt.Sprintf("'%s'", prop)
	}
	return fmt.Sprintf("values(%s)", strings.Join(props, ", "))
}

func (s *ValuesStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// WhereStep - Complex filtering with predicates
type WhereStep struct {
	Predicate Predicate
}

func (s *WhereStep) String() string {
	return fmt.Sprintf("where(%s)", s.Predicate.String())
}

func (s *WhereStep) Execute(traverser *Traverser) (*Traverser, error) {
	return traverser, nil
}

// Predicate represents filtering conditions
type Predicate interface {
	String() string
	Evaluate(element interface{}) (bool, error)
}

// EqualsPredicate - Property equals value
type EqualsPredicate struct {
	Property string
	Value    interface{}
}

func (p *EqualsPredicate) String() string {
	return fmt.Sprintf("%s == %v", p.Property, p.Value)
}

// ContainsPredicate - Property contains value (for strings/arrays)
type ContainsPredicate struct {
	Property string
	Value    interface{}
}

func (p *ContainsPredicate) String() string {
	return fmt.Sprintf("%s.contains(%v)", p.Property, p.Value)
}

// RangePredicate - Property within range
type RangePredicate struct {
	Property string
	Min      interface{}
	Max      interface{}
}

func (p *RangePredicate) String() string {
	return fmt.Sprintf("%s.between(%v, %v)", p.Property, p.Min, p.Max)
}
