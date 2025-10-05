package transaction

import (
	"fmt"
	"strings"

	"github.com/tenzoki/agen/omni/internal/common"
)

// DefaultConsistencyChecker implements basic consistency checks
type DefaultConsistencyChecker struct {
	inconsistencies []string
}

// NewDefaultConsistencyChecker creates a new default consistency checker
func NewDefaultConsistencyChecker() ConsistencyChecker {
	return &DefaultConsistencyChecker{
		inconsistencies: make([]string, 0),
	}
}

// CheckConsistency performs comprehensive consistency checks
func (cc *DefaultConsistencyChecker) CheckConsistency(tx GraphTx) error {
	cc.inconsistencies = cc.inconsistencies[:0] // Clear previous inconsistencies

	// Check all vertices
	vertices, err := tx.GetAllVertices(-1)
	if err != nil {
		return fmt.Errorf("failed to get vertices for consistency check: %w", err)
	}

	for _, vertex := range vertices {
		if err := cc.CheckVertexConsistency(tx, vertex); err != nil {
			cc.inconsistencies = append(cc.inconsistencies, err.Error())
		}
	}

	// Check all edges
	edges, err := tx.GetAllEdges(-1)
	if err != nil {
		return fmt.Errorf("failed to get edges for consistency check: %w", err)
	}

	for _, edge := range edges {
		if err := cc.CheckEdgeConsistency(tx, edge); err != nil {
			cc.inconsistencies = append(cc.inconsistencies, err.Error())
		}
	}

	// Check referential integrity
	if err := cc.CheckReferentialIntegrity(tx); err != nil {
		cc.inconsistencies = append(cc.inconsistencies, err.Error())
	}

	if len(cc.inconsistencies) > 0 {
		return fmt.Errorf("consistency check failed with %d issues: %s",
			len(cc.inconsistencies), strings.Join(cc.inconsistencies, "; "))
	}

	return nil
}

// CheckVertexConsistency checks individual vertex consistency
func (cc *DefaultConsistencyChecker) CheckVertexConsistency(tx GraphTx, vertex *common.Vertex) error {
	// Basic validation
	if err := vertex.Validate(); err != nil {
		return fmt.Errorf("vertex %s validation failed: %w", vertex.ID, err)
	}

	// Check ID format
	if strings.TrimSpace(vertex.ID) == "" {
		return fmt.Errorf("vertex ID cannot be empty")
	}

	// Check type format
	if strings.TrimSpace(vertex.Type) == "" {
		return fmt.Errorf("vertex %s type cannot be empty", vertex.ID)
	}

	// Check timestamp consistency
	if vertex.UpdatedAt.Before(vertex.CreatedAt) {
		return fmt.Errorf("vertex %s updated time is before created time", vertex.ID)
	}

	// Check version consistency
	if vertex.Version == 0 {
		return fmt.Errorf("vertex %s version must be greater than 0", vertex.ID)
	}

	return nil
}

// CheckEdgeConsistency checks individual edge consistency
func (cc *DefaultConsistencyChecker) CheckEdgeConsistency(tx GraphTx, edge *common.Edge) error {
	// Basic validation
	if err := edge.Validate(); err != nil {
		return fmt.Errorf("edge %s validation failed: %w", edge.ID, err)
	}

	// Check ID format
	if strings.TrimSpace(edge.ID) == "" {
		return fmt.Errorf("edge ID cannot be empty")
	}

	// Check type format
	if strings.TrimSpace(edge.Type) == "" {
		return fmt.Errorf("edge %s type cannot be empty", edge.ID)
	}

	// Check vertex references
	if strings.TrimSpace(edge.FromVertex) == "" {
		return fmt.Errorf("edge %s from vertex cannot be empty", edge.ID)
	}

	if strings.TrimSpace(edge.ToVertex) == "" {
		return fmt.Errorf("edge %s to vertex cannot be empty", edge.ID)
	}

	// Self-loops check (optional - some graphs allow self-loops)
	if edge.FromVertex == edge.ToVertex {
		// This could be a warning rather than an error depending on requirements
		// return fmt.Errorf("edge %s is a self-loop", edge.ID)
	}

	// Check weight validity
	if edge.Weight < 0 {
		return fmt.Errorf("edge %s has negative weight: %f", edge.ID, edge.Weight)
	}

	return nil
}

// CheckReferentialIntegrity ensures all edge references point to existing vertices
func (cc *DefaultConsistencyChecker) CheckReferentialIntegrity(tx GraphTx) error {
	edges, err := tx.GetAllEdges(-1)
	if err != nil {
		return fmt.Errorf("failed to get edges for referential integrity check: %w", err)
	}

	for _, edge := range edges {
		// Check if from vertex exists
		if exists, err := tx.VertexExists(edge.FromVertex); err != nil {
			return fmt.Errorf("failed to check from vertex %s for edge %s: %w",
				edge.FromVertex, edge.ID, err)
		} else if !exists {
			return fmt.Errorf("edge %s references non-existent from vertex %s",
				edge.ID, edge.FromVertex)
		}

		// Check if to vertex exists
		if exists, err := tx.VertexExists(edge.ToVertex); err != nil {
			return fmt.Errorf("failed to check to vertex %s for edge %s: %w",
				edge.ToVertex, edge.ID, err)
		} else if !exists {
			return fmt.Errorf("edge %s references non-existent to vertex %s",
				edge.ID, edge.ToVertex)
		}
	}

	return nil
}

// GetInconsistencies returns all detected inconsistencies
func (cc *DefaultConsistencyChecker) GetInconsistencies() []string {
	return cc.inconsistencies
}

// Validation Rules

// RequiredPropertyRule ensures specific properties are present
type RequiredPropertyRule struct {
	name       string
	property   string
	vertexType string
	enabled    bool
}

// NewRequiredPropertyRule creates a new required property rule
func NewRequiredPropertyRule(name, property, vertexType string) ValidationRule {
	return &RequiredPropertyRule{
		name:       name,
		property:   property,
		vertexType: vertexType,
		enabled:    true,
	}
}

// Validate checks if the required property is present
func (r *RequiredPropertyRule) Validate(tx GraphTx, operation *Operation) error {
	if !r.enabled {
		return nil
	}

	if operation.Type != OpCreateVertex && operation.Type != OpUpdateVertex {
		return nil
	}

	// Unmarshal vertex data
	var vertex common.Vertex
	if err := vertex.UnmarshalBinary(operation.Value); err != nil {
		return fmt.Errorf("failed to unmarshal vertex data: %w", err)
	}

	// Check if this rule applies to this vertex type
	if r.vertexType != "" && vertex.Type != r.vertexType {
		return nil
	}

	// Check if required property exists
	if _, exists := vertex.Properties[r.property]; !exists {
		return fmt.Errorf("required property '%s' missing from vertex %s", r.property, vertex.ID)
	}

	return nil
}

// GetRuleName returns the rule name
func (r *RequiredPropertyRule) GetRuleName() string {
	return r.name
}

// IsEnabled returns if the rule is enabled
func (r *RequiredPropertyRule) IsEnabled() bool {
	return r.enabled
}

// UniquePropertyRule ensures property values are unique within a vertex type
type UniquePropertyRule struct {
	name       string
	property   string
	vertexType string
	enabled    bool
}

// NewUniquePropertyRule creates a new unique property rule
func NewUniquePropertyRule(name, property, vertexType string) ValidationRule {
	return &UniquePropertyRule{
		name:       name,
		property:   property,
		vertexType: vertexType,
		enabled:    true,
	}
}

// Validate checks if the property value is unique
func (r *UniquePropertyRule) Validate(tx GraphTx, operation *Operation) error {
	if !r.enabled {
		return nil
	}

	if operation.Type != OpCreateVertex && operation.Type != OpUpdateVertex {
		return nil
	}

	// Unmarshal vertex data
	var vertex common.Vertex
	if err := vertex.UnmarshalBinary(operation.Value); err != nil {
		return fmt.Errorf("failed to unmarshal vertex data: %w", err)
	}

	// Check if this rule applies to this vertex type
	if vertex.Type != r.vertexType {
		return nil
	}

	// Get the property value to check
	propertyValue, exists := vertex.Properties[r.property]
	if !exists {
		return nil // Property doesn't exist, uniqueness doesn't apply
	}

	// Get all vertices of the same type
	vertices, err := tx.GetVerticesByType(r.vertexType, -1)
	if err != nil {
		return fmt.Errorf("failed to get vertices for uniqueness check: %w", err)
	}

	// Check for duplicates
	for _, existingVertex := range vertices {
		// Skip the vertex being updated
		if existingVertex.ID == vertex.ID {
			continue
		}

		if existingValue, exists := existingVertex.Properties[r.property]; exists {
			if fmt.Sprintf("%v", existingValue) == fmt.Sprintf("%v", propertyValue) {
				return fmt.Errorf("property '%s' value '%v' already exists in vertex %s",
					r.property, propertyValue, existingVertex.ID)
			}
		}
	}

	return nil
}

// GetRuleName returns the rule name
func (r *UniquePropertyRule) GetRuleName() string {
	return r.name
}

// IsEnabled returns if the rule is enabled
func (r *UniquePropertyRule) IsEnabled() bool {
	return r.enabled
}

// EdgeCountLimitRule limits the number of edges per vertex
type EdgeCountLimitRule struct {
	name      string
	maxEdges  int
	direction common.Direction
	enabled   bool
}

// NewEdgeCountLimitRule creates a new edge count limit rule
func NewEdgeCountLimitRule(name string, maxEdges int, direction common.Direction) ValidationRule {
	return &EdgeCountLimitRule{
		name:      name,
		maxEdges:  maxEdges,
		direction: direction,
		enabled:   true,
	}
}

// Validate checks if the edge count limit is exceeded
func (r *EdgeCountLimitRule) Validate(tx GraphTx, operation *Operation) error {
	if !r.enabled {
		return nil
	}

	if operation.Type != OpCreateEdge {
		return nil
	}

	// Unmarshal edge data
	var edge common.Edge
	if err := edge.UnmarshalBinary(operation.Value); err != nil {
		return fmt.Errorf("failed to unmarshal edge data: %w", err)
	}

	// Check limits based on direction
	switch r.direction {
	case common.Outgoing:
		edges, err := tx.GetOutgoingEdges(edge.FromVertex)
		if err != nil {
			return fmt.Errorf("failed to get outgoing edges: %w", err)
		}
		if len(edges) >= r.maxEdges {
			return fmt.Errorf("vertex %s exceeds maximum outgoing edge limit of %d",
				edge.FromVertex, r.maxEdges)
		}

	case common.Incoming:
		edges, err := tx.GetIncomingEdges(edge.ToVertex)
		if err != nil {
			return fmt.Errorf("failed to get incoming edges: %w", err)
		}
		if len(edges) >= r.maxEdges {
			return fmt.Errorf("vertex %s exceeds maximum incoming edge limit of %d",
				edge.ToVertex, r.maxEdges)
		}

	case common.Both:
		// Check both vertices
		outEdges, err := tx.GetOutgoingEdges(edge.FromVertex)
		if err != nil {
			return fmt.Errorf("failed to get outgoing edges: %w", err)
		}
		inEdges, err := tx.GetIncomingEdges(edge.FromVertex)
		if err != nil {
			return fmt.Errorf("failed to get incoming edges: %w", err)
		}
		if len(outEdges)+len(inEdges) >= r.maxEdges {
			return fmt.Errorf("vertex %s exceeds maximum total edge limit of %d",
				edge.FromVertex, r.maxEdges)
		}
	}

	return nil
}

// GetRuleName returns the rule name
func (r *EdgeCountLimitRule) GetRuleName() string {
	return r.name
}

// IsEnabled returns if the rule is enabled
func (r *EdgeCountLimitRule) IsEnabled() bool {
	return r.enabled
}

// ValidationEngine manages and applies validation rules
type ValidationEngine struct {
	rules   []ValidationRule
	enabled bool
}

// NewValidationEngine creates a new validation engine
func NewValidationEngine() *ValidationEngine {
	return &ValidationEngine{
		rules:   make([]ValidationRule, 0),
		enabled: true,
	}
}

// AddRule adds a validation rule to the engine
func (ve *ValidationEngine) AddRule(rule ValidationRule) {
	ve.rules = append(ve.rules, rule)
}

// RemoveRule removes a validation rule by name
func (ve *ValidationEngine) RemoveRule(ruleName string) {
	for i, rule := range ve.rules {
		if rule.GetRuleName() == ruleName {
			ve.rules = append(ve.rules[:i], ve.rules[i+1:]...)
			break
		}
	}
}

// ValidateOperation validates an operation against all rules
func (ve *ValidationEngine) ValidateOperation(tx GraphTx, operation *Operation) error {
	if !ve.enabled {
		return nil
	}

	for _, rule := range ve.rules {
		if rule.IsEnabled() {
			if err := rule.Validate(tx, operation); err != nil {
				return fmt.Errorf("validation rule '%s' failed: %w", rule.GetRuleName(), err)
			}
		}
	}

	return nil
}

// GetRules returns all validation rules
func (ve *ValidationEngine) GetRules() []ValidationRule {
	return ve.rules
}

// SetEnabled enables or disables the validation engine
func (ve *ValidationEngine) SetEnabled(enabled bool) {
	ve.enabled = enabled
}

// IsEnabled returns if the validation engine is enabled
func (ve *ValidationEngine) IsEnabled() bool {
	return ve.enabled
}
