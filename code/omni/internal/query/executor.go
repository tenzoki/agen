package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/agen/omni/internal/common"
	"github.com/agen/omni/internal/graph"
)

// QueryExecutor executes graph queries against a graph store
type QueryExecutor struct {
	graphStore graph.GraphStore
}

// NewQueryExecutor creates a new query executor
func NewQueryExecutor(graphStore graph.GraphStore) *QueryExecutor {
	return &QueryExecutor{
		graphStore: graphStore,
	}
}

// Execute executes a query and returns the results
func (qe *QueryExecutor) Execute(query *Query) (*QueryResult, error) {
	traverser := NewTraverser()

	// Execute each step in sequence
	for _, step := range query.Steps {
		var err error
		traverser, err = qe.executeStep(step, traverser)
		if err != nil {
			return nil, fmt.Errorf("error executing step %s: %w", step.String(), err)
		}
	}

	return &QueryResult{
		Vertices: traverser.Vertices,
		Edges:    traverser.Edges,
		Values:   traverser.Values,
		Count:    traverser.Count,
	}, nil
}

// executeStep executes a single query step
func (qe *QueryExecutor) executeStep(step QueryStep, traverser *Traverser) (*Traverser, error) {
	switch s := step.(type) {
	case *VStep:
		return qe.executeVStep(s, traverser)
	case *EStep:
		return qe.executeEStep(s, traverser)
	case *OutStep:
		return qe.executeOutStep(s, traverser)
	case *InStep:
		return qe.executeInStep(s, traverser)
	case *BothStep:
		return qe.executeBothStep(s, traverser)
	case *HasStep:
		return qe.executeHasStep(s, traverser)
	case *HasLabelStep:
		return qe.executeHasLabelStep(s, traverser)
	case *CountStep:
		return qe.executeCountStep(s, traverser)
	case *LimitStep:
		return qe.executeLimitStep(s, traverser)
	case *ValuesStep:
		return qe.executeValuesStep(s, traverser)
	case *WhereStep:
		return qe.executeWhereStep(s, traverser)
	default:
		return nil, fmt.Errorf("unknown step type: %T", step)
	}
}

// executeVStep executes a V() step
func (qe *QueryExecutor) executeVStep(step *VStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	if len(step.IDs) == 0 {
		// Get all vertices
		vertices, err := qe.graphStore.GetAllVertices(-1)
		if err != nil {
			return nil, err
		}
		result.Vertices = vertices
	} else {
		// Get specific vertices
		for _, id := range step.IDs {
			vertex, err := qe.graphStore.GetVertex(id)
			if err == nil {
				result.Vertices = append(result.Vertices, vertex)
			}
		}
	}

	return result, nil
}

// executeEStep executes an E() step
func (qe *QueryExecutor) executeEStep(step *EStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	if len(step.IDs) == 0 {
		// Get all edges
		edges, err := qe.graphStore.GetAllEdges(-1)
		if err != nil {
			return nil, err
		}
		result.Edges = edges
	} else {
		// Get specific edges
		for _, id := range step.IDs {
			edge, err := qe.graphStore.GetEdge(id)
			if err == nil {
				result.Edges = append(result.Edges, edge)
			}
		}
	}

	return result, nil
}

// executeOutStep executes an out() step
func (qe *QueryExecutor) executeOutStep(step *OutStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	for _, vertex := range traverser.Vertices {
		outgoingEdges, err := qe.graphStore.GetOutgoingEdges(vertex.ID)
		if err != nil {
			continue
		}

		// Filter by edge labels if specified
		filteredEdges := qe.filterEdgesByLabels(outgoingEdges, step.EdgeLabels)

		// Get target vertices
		for _, edge := range filteredEdges {
			targetVertex, err := qe.graphStore.GetVertex(edge.ToVertex)
			if err == nil {
				result.Vertices = append(result.Vertices, targetVertex)
			}
		}
	}

	return result, nil
}

// executeInStep executes an in() step
func (qe *QueryExecutor) executeInStep(step *InStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	for _, vertex := range traverser.Vertices {
		incomingEdges, err := qe.graphStore.GetIncomingEdges(vertex.ID)
		if err != nil {
			continue
		}

		// Filter by edge labels if specified
		filteredEdges := qe.filterEdgesByLabels(incomingEdges, step.EdgeLabels)

		// Get source vertices
		for _, edge := range filteredEdges {
			sourceVertex, err := qe.graphStore.GetVertex(edge.FromVertex)
			if err == nil {
				result.Vertices = append(result.Vertices, sourceVertex)
			}
		}
	}

	return result, nil
}

// executeBothStep executes a both() step
func (qe *QueryExecutor) executeBothStep(step *BothStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	for _, vertex := range traverser.Vertices {
		// Get both incoming and outgoing neighbors
		neighbors, err := qe.graphStore.GetNeighbors(vertex.ID, graph.DirectionBoth)
		if err != nil {
			continue
		}

		// If edge labels are specified, we need to check each edge
		if len(step.EdgeLabels) > 0 {
			// Get outgoing edges
			outgoingEdges, err := qe.graphStore.GetOutgoingEdges(vertex.ID)
			if err == nil {
				filteredOut := qe.filterEdgesByLabels(outgoingEdges, step.EdgeLabels)
				for _, edge := range filteredOut {
					if targetVertex, err := qe.graphStore.GetVertex(edge.ToVertex); err == nil {
						result.Vertices = append(result.Vertices, targetVertex)
					}
				}
			}

			// Get incoming edges
			incomingEdges, err := qe.graphStore.GetIncomingEdges(vertex.ID)
			if err == nil {
				filteredIn := qe.filterEdgesByLabels(incomingEdges, step.EdgeLabels)
				for _, edge := range filteredIn {
					if sourceVertex, err := qe.graphStore.GetVertex(edge.FromVertex); err == nil {
						result.Vertices = append(result.Vertices, sourceVertex)
					}
				}
			}
		} else {
			// No edge label filtering, use all neighbors
			for _, neighbor := range neighbors {
				result.Vertices = append(result.Vertices, neighbor)
			}
		}
	}

	return result, nil
}

// executeHasStep executes a has() step
func (qe *QueryExecutor) executeHasStep(step *HasStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	// Filter vertices
	for _, vertex := range traverser.Vertices {
		if qe.vertexMatchesHas(vertex, step) {
			result.Vertices = append(result.Vertices, vertex)
		}
	}

	// Filter edges
	for _, edge := range traverser.Edges {
		if qe.edgeMatchesHas(edge, step) {
			result.Edges = append(result.Edges, edge)
		}
	}

	return result, nil
}

// executeHasLabelStep executes a hasLabel() step
func (qe *QueryExecutor) executeHasLabelStep(step *HasLabelStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	// Filter vertices by type
	for _, vertex := range traverser.Vertices {
		for _, label := range step.Labels {
			if vertex.Type == label {
				result.Vertices = append(result.Vertices, vertex)
				break
			}
		}
	}

	// Filter edges by type
	for _, edge := range traverser.Edges {
		for _, label := range step.Labels {
			if edge.Type == label {
				result.Edges = append(result.Edges, edge)
				break
			}
		}
	}

	return result, nil
}

// executeCountStep executes a count() step
func (qe *QueryExecutor) executeCountStep(step *CountStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()
	count := int64(len(traverser.Vertices) + len(traverser.Edges) + len(traverser.Values))
	result.Count = count
	result.Values = []interface{}{count}
	return result, nil
}

// executeLimitStep executes a limit() step
func (qe *QueryExecutor) executeLimitStep(step *LimitStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	// Apply limit to vertices
	if len(traverser.Vertices) > step.Count {
		result.Vertices = traverser.Vertices[:step.Count]
	} else {
		result.Vertices = traverser.Vertices
	}

	// Apply limit to edges (if no vertices)
	if len(result.Vertices) == 0 {
		if len(traverser.Edges) > step.Count {
			result.Edges = traverser.Edges[:step.Count]
		} else {
			result.Edges = traverser.Edges
		}
	}

	// Apply limit to values (if no vertices or edges)
	if len(result.Vertices) == 0 && len(result.Edges) == 0 {
		if len(traverser.Values) > step.Count {
			result.Values = traverser.Values[:step.Count]
		} else {
			result.Values = traverser.Values
		}
	}

	return result, nil
}

// executeValuesStep executes a values() step
func (qe *QueryExecutor) executeValuesStep(step *ValuesStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	// Extract values from vertices
	for _, vertex := range traverser.Vertices {
		for _, prop := range step.Properties {
			if value, exists := vertex.Properties[prop]; exists {
				result.Values = append(result.Values, value)
			}
		}
	}

	// Extract values from edges
	for _, edge := range traverser.Edges {
		for _, prop := range step.Properties {
			if value, exists := edge.Properties[prop]; exists {
				result.Values = append(result.Values, value)
			}
		}
	}

	return result, nil
}

// executeWhereStep executes a where() step
func (qe *QueryExecutor) executeWhereStep(step *WhereStep, traverser *Traverser) (*Traverser, error) {
	result := NewTraverser()

	// Filter vertices
	for _, vertex := range traverser.Vertices {
		matches, err := step.Predicate.Evaluate(vertex)
		if err != nil {
			return nil, err
		}
		if matches {
			result.Vertices = append(result.Vertices, vertex)
		}
	}

	// Filter edges
	for _, edge := range traverser.Edges {
		matches, err := step.Predicate.Evaluate(edge)
		if err != nil {
			return nil, err
		}
		if matches {
			result.Edges = append(result.Edges, edge)
		}
	}

	return result, nil
}

// Helper functions

// filterEdgesByLabels filters edges by their labels/types
func (qe *QueryExecutor) filterEdgesByLabels(edges []*common.Edge, labels []string) []*common.Edge {
	if len(labels) == 0 {
		return edges
	}

	filtered := make([]*common.Edge, 0)
	labelSet := make(map[string]bool)
	for _, label := range labels {
		labelSet[label] = true
	}

	for _, edge := range edges {
		if labelSet[edge.Type] {
			filtered = append(filtered, edge)
		}
	}

	return filtered
}

// vertexMatchesHas checks if a vertex matches a has() condition
func (qe *QueryExecutor) vertexMatchesHas(vertex *common.Vertex, step *HasStep) bool {
	value, exists := vertex.Properties[step.Property]

	if step.Value == nil {
		// Just checking for property existence
		return exists
	}

	if !exists {
		return false
	}

	// Compare values - with enhanced robustness for different types
	return qe.compareValuesRobust(value, step.Value)
}

// edgeMatchesHas checks if an edge matches a has() condition
func (qe *QueryExecutor) edgeMatchesHas(edge *common.Edge, step *HasStep) bool {
	value, exists := edge.Properties[step.Property]

	if step.Value == nil {
		// Just checking for property existence
		return exists
	}

	if !exists {
		return false
	}

	// Compare values
	return qe.compareValues(value, step.Value)
}

// compareValues compares two values for equality
func (qe *QueryExecutor) compareValues(a, b interface{}) bool {
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

	// For numeric values, always try numeric conversion first
	// This handles int vs int comparisons more reliably
	if qe.compareNumeric(a, b) {
		return true
	}

	// String comparison
	if aStr, aOk := a.(string); aOk {
		if bStr, bOk := b.(string); bOk {
			return aStr == bStr
		}
	}

	// Use reflection for deeper comparison
	return reflect.DeepEqual(a, b)
}

// compareNumeric handles numeric type conversions
func (qe *QueryExecutor) compareNumeric(a, b interface{}) bool {
	// Convert both to float64 for comparison if they're numeric
	aFloat, aOk := qe.toFloat64(a)
	bFloat, bOk := qe.toFloat64(b)

	if aOk && bOk {
		return aFloat == bFloat
	}

	return false
}

// toFloat64 converts various numeric types to float64
func (qe *QueryExecutor) toFloat64(v interface{}) (float64, bool) {
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

// compareValuesRobust is an enhanced version that handles edge cases better
func (qe *QueryExecutor) compareValuesRobust(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Direct equality check first
	if a == b {
		return true
	}

	// Convert to string and compare (handles many edge cases)
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	if aStr == bStr {
		return true
	}

	// Numeric comparison
	if qe.compareNumeric(a, b) {
		return true
	}

	// Reflection comparison
	return reflect.DeepEqual(a, b)
}

// QueryResult represents the result of executing a query
type QueryResult struct {
	Vertices []*common.Vertex
	Edges    []*common.Edge
	Values   []interface{}
	Count    int64
}

// NewQueryResult creates a new empty QueryResult
func NewQueryResult() *QueryResult {
	return &QueryResult{
		Vertices: make([]*common.Vertex, 0),
		Edges:    make([]*common.Edge, 0),
		Values:   make([]interface{}, 0),
		Count:    0,
	}
}

// IsEmpty returns true if the result contains no data
func (qr *QueryResult) IsEmpty() bool {
	return len(qr.Vertices) == 0 && len(qr.Edges) == 0 && len(qr.Values) == 0 && qr.Count == 0
}

// Size returns the total number of elements in the result
func (qr *QueryResult) Size() int {
	if qr.Count > 0 {
		return int(qr.Count)
	}
	return len(qr.Vertices) + len(qr.Edges) + len(qr.Values)
}

// String returns a string representation of the result
func (qr *QueryResult) String() string {
	var parts []string

	if len(qr.Vertices) > 0 {
		parts = append(parts, fmt.Sprintf("Vertices: %d", len(qr.Vertices)))
	}
	if len(qr.Edges) > 0 {
		parts = append(parts, fmt.Sprintf("Edges: %d", len(qr.Edges)))
	}
	if len(qr.Values) > 0 {
		parts = append(parts, fmt.Sprintf("Values: %d", len(qr.Values)))
	}
	if qr.Count > 0 {
		parts = append(parts, fmt.Sprintf("Count: %d", qr.Count))
	}

	if len(parts) == 0 {
		return "Empty result"
	}

	return strings.Join(parts, ", ")
}

// Implement predicate evaluation

// Evaluate evaluates an equals predicate
func (p *EqualsPredicate) Evaluate(element interface{}) (bool, error) {
	switch e := element.(type) {
	case *common.Vertex:
		if value, exists := e.Properties[p.Property]; exists {
			return reflect.DeepEqual(value, p.Value), nil
		}
		return false, nil
	case *common.Edge:
		if value, exists := e.Properties[p.Property]; exists {
			return reflect.DeepEqual(value, p.Value), nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported element type: %T", element)
	}
}

// Evaluate evaluates a contains predicate
func (p *ContainsPredicate) Evaluate(element interface{}) (bool, error) {
	switch e := element.(type) {
	case *common.Vertex:
		if value, exists := e.Properties[p.Property]; exists {
			return p.containsValue(value, p.Value), nil
		}
		return false, nil
	case *common.Edge:
		if value, exists := e.Properties[p.Property]; exists {
			return p.containsValue(value, p.Value), nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported element type: %T", element)
	}
}

// containsValue checks if a value contains another value
func (p *ContainsPredicate) containsValue(container, value interface{}) bool {
	switch c := container.(type) {
	case string:
		if s, ok := value.(string); ok {
			return strings.Contains(c, s)
		}
	case []interface{}:
		for _, item := range c {
			if reflect.DeepEqual(item, value) {
				return true
			}
		}
	case []string:
		if s, ok := value.(string); ok {
			for _, item := range c {
				if item == s {
					return true
				}
			}
		}
	}
	return false
}

// Evaluate evaluates a range predicate
func (p *RangePredicate) Evaluate(element interface{}) (bool, error) {
	switch e := element.(type) {
	case *common.Vertex:
		if value, exists := e.Properties[p.Property]; exists {
			return p.inRange(value), nil
		}
		return false, nil
	case *common.Edge:
		if value, exists := e.Properties[p.Property]; exists {
			return p.inRange(value), nil
		}
		return false, nil
	default:
		return false, fmt.Errorf("unsupported element type: %T", element)
	}
}

// inRange checks if a value is within the specified range
func (p *RangePredicate) inRange(value interface{}) bool {
	// This is a simplified implementation
	// In a real implementation, you'd want more sophisticated comparison
	switch v := value.(type) {
	case int:
		if min, ok := p.Min.(int); ok {
			if max, ok := p.Max.(int); ok {
				return v >= min && v <= max
			}
		}
	case float64:
		if min, ok := p.Min.(float64); ok {
			if max, ok := p.Max.(float64); ok {
				return v >= min && v <= max
			}
		}
	case string:
		if min, ok := p.Min.(string); ok {
			if max, ok := p.Max.(string); ok {
				return v >= min && v <= max
			}
		}
	}
	return false
}
