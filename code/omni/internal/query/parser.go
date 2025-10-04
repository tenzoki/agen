package query

import (
	"fmt"
	"strconv"
	"strings"
)

// QueryBuilder provides a fluent API for building graph queries
type QueryBuilder struct {
	query *Query
}

// NewQueryBuilder creates a new query builder starting with 'g'
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: NewQuery(),
	}
}

// V starts traversal from vertices
func (qb *QueryBuilder) V(ids ...string) *QueryBuilder {
	qb.query.AddStep(&VStep{IDs: ids})
	return qb
}

// E starts traversal from edges
func (qb *QueryBuilder) E(ids ...string) *QueryBuilder {
	qb.query.AddStep(&EStep{IDs: ids})
	return qb
}

// Out traverses outgoing edges
func (qb *QueryBuilder) Out(labels ...string) *QueryBuilder {
	qb.query.AddStep(&OutStep{EdgeLabels: labels})
	return qb
}

// In traverses incoming edges
func (qb *QueryBuilder) In(labels ...string) *QueryBuilder {
	qb.query.AddStep(&InStep{EdgeLabels: labels})
	return qb
}

// Both traverses both incoming and outgoing edges
func (qb *QueryBuilder) Both(labels ...string) *QueryBuilder {
	qb.query.AddStep(&BothStep{EdgeLabels: labels})
	return qb
}

// Has filters by property existence or value
func (qb *QueryBuilder) Has(property string, value ...interface{}) *QueryBuilder {
	var val interface{}
	if len(value) > 0 {
		val = value[0]
	}
	qb.query.AddStep(&HasStep{Property: property, Value: val})
	return qb
}

// HasLabel filters by vertex/edge label (type)
func (qb *QueryBuilder) HasLabel(labels ...string) *QueryBuilder {
	qb.query.AddStep(&HasLabelStep{Labels: labels})
	return qb
}

// Count counts the current elements
func (qb *QueryBuilder) Count() *QueryBuilder {
	qb.query.AddStep(&CountStep{})
	return qb
}

// Limit limits the number of results
func (qb *QueryBuilder) Limit(count int) *QueryBuilder {
	qb.query.AddStep(&LimitStep{Count: count})
	return qb
}

// Values extracts property values
func (qb *QueryBuilder) Values(properties ...string) *QueryBuilder {
	qb.query.AddStep(&ValuesStep{Properties: properties})
	return qb
}

// Where applies complex filtering with predicates
func (qb *QueryBuilder) Where(predicate Predicate) *QueryBuilder {
	qb.query.AddStep(&WhereStep{Predicate: predicate})
	return qb
}

// Build returns the completed query
func (qb *QueryBuilder) Build() *Query {
	return qb.query
}

// String returns the string representation
func (qb *QueryBuilder) String() string {
	return qb.query.String()
}

// QueryParser parses string queries into Query objects
type QueryParser struct{}

// NewQueryParser creates a new query parser
func NewQueryParser() *QueryParser {
	return &QueryParser{}
}

// Parse parses a query string like "g.V().hasLabel('User').out('follows').count()"
func (p *QueryParser) Parse(queryStr string) (*Query, error) {
	queryStr = strings.TrimSpace(queryStr)

	// Must start with 'g'
	if !strings.HasPrefix(queryStr, "g") {
		return nil, fmt.Errorf("query must start with 'g'")
	}

	// Handle just "g" case
	if queryStr == "g" {
		return NewQuery(), nil
	}

	// Split by dots and parse each step
	if !strings.HasPrefix(queryStr, "g.") {
		return nil, fmt.Errorf("invalid query format, expected 'g.' prefix")
	}

	stepStr := queryStr[2:] // Remove "g."
	steps, err := p.parseSteps(stepStr)
	if err != nil {
		return nil, err
	}

	query := NewQuery()
	for _, step := range steps {
		query.AddStep(step)
	}

	return query, nil
}

// parseSteps parses the step portion of a query
func (p *QueryParser) parseSteps(stepStr string) ([]QueryStep, error) {
	steps := make([]QueryStep, 0)

	// Simple tokenization - split by dots, but be careful about dots in parameters
	tokens := p.tokenize(stepStr)

	for _, token := range tokens {
		step, err := p.parseStep(token)
		if err != nil {
			return nil, fmt.Errorf("error parsing step '%s': %w", token, err)
		}
		steps = append(steps, step)
	}

	return steps, nil
}

// tokenize splits the step string by dots while handling parameters
func (p *QueryParser) tokenize(stepStr string) []string {
	tokens := make([]string, 0)
	current := ""
	parenDepth := 0

	for _, char := range stepStr {
		switch char {
		case '(':
			parenDepth++
			current += string(char)
		case ')':
			parenDepth--
			current += string(char)
		case '.':
			if parenDepth == 0 {
				if current != "" {
					tokens = append(tokens, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		tokens = append(tokens, current)
	}

	return tokens
}

// parseStep parses a single step like "V('user:1')" or "hasLabel('User')"
func (p *QueryParser) parseStep(stepStr string) (QueryStep, error) {
	stepStr = strings.TrimSpace(stepStr)

	// Find step name and parameters
	parenIndex := strings.Index(stepStr, "(")
	if parenIndex == -1 {
		return nil, fmt.Errorf("step must have parentheses: %s", stepStr)
	}

	stepName := stepStr[:parenIndex]
	paramStr := stepStr[parenIndex+1 : len(stepStr)-1] // Remove '(' and ')'

	params, err := p.parseParameters(paramStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing parameters '%s': %w", paramStr, err)
	}

	// Parse based on step name
	switch strings.ToLower(stepName) {
	case "v":
		ids := make([]string, len(params))
		for i, param := range params {
			ids[i] = p.paramToString(param)
		}
		return &VStep{IDs: ids}, nil

	case "e":
		ids := make([]string, len(params))
		for i, param := range params {
			ids[i] = p.paramToString(param)
		}
		return &EStep{IDs: ids}, nil

	case "out":
		labels := make([]string, len(params))
		for i, param := range params {
			labels[i] = p.paramToString(param)
		}
		return &OutStep{EdgeLabels: labels}, nil

	case "in":
		labels := make([]string, len(params))
		for i, param := range params {
			labels[i] = p.paramToString(param)
		}
		return &InStep{EdgeLabels: labels}, nil

	case "both":
		labels := make([]string, len(params))
		for i, param := range params {
			labels[i] = p.paramToString(param)
		}
		return &BothStep{EdgeLabels: labels}, nil

	case "has":
		if len(params) == 0 {
			return nil, fmt.Errorf("has() requires at least one parameter")
		}
		prop := p.paramToString(params[0])
		var value interface{}
		if len(params) > 1 {
			value = params[1]
		}
		return &HasStep{Property: prop, Value: value}, nil

	case "haslabel":
		labels := make([]string, len(params))
		for i, param := range params {
			labels[i] = p.paramToString(param)
		}
		return &HasLabelStep{Labels: labels}, nil

	case "count":
		return &CountStep{}, nil

	case "limit":
		if len(params) != 1 {
			return nil, fmt.Errorf("limit() requires exactly one parameter")
		}
		count, ok := params[0].(int)
		if !ok {
			return nil, fmt.Errorf("limit() parameter must be an integer")
		}
		return &LimitStep{Count: count}, nil

	case "values":
		props := make([]string, len(params))
		for i, param := range params {
			props[i] = p.paramToString(param)
		}
		return &ValuesStep{Properties: props}, nil

	default:
		return nil, fmt.Errorf("unknown step: %s", stepName)
	}
}

// parseParameters parses parameter list like "'user:1', 'user:2'" or "name, 'Alice'"
func (p *QueryParser) parseParameters(paramStr string) ([]interface{}, error) {
	if paramStr == "" {
		return []interface{}{}, nil
	}

	params := make([]interface{}, 0)
	current := ""
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(paramStr); i++ {
		char := paramStr[i]

		switch char {
		case '\'', '"':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current += string(char)
			}
		case ',':
			if !inQuotes {
				param, err := p.parseParameter(strings.TrimSpace(current))
				if err != nil {
					return nil, err
				}
				params = append(params, param)
				current = ""
			} else {
				current += string(char)
			}
		default:
			current += string(char)
		}
	}

	// Add the last parameter
	if current != "" {
		param, err := p.parseParameter(strings.TrimSpace(current))
		if err != nil {
			return nil, err
		}
		params = append(params, param)
	}

	return params, nil
}

// parseParameter parses a single parameter value
func (p *QueryParser) parseParameter(paramStr string) (interface{}, error) {
	paramStr = strings.TrimSpace(paramStr)

	if paramStr == "" {
		return "", nil
	}

	// Try to parse as integer
	if intVal, err := strconv.Atoi(paramStr); err == nil {
		return intVal, nil
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(paramStr, 64); err == nil {
		return floatVal, nil
	}

	// Try to parse as boolean
	if boolVal, err := strconv.ParseBool(paramStr); err == nil {
		return boolVal, nil
	}

	// Default to string
	return paramStr, nil
}

// paramToString converts a parameter to string, removing quotes if present
func (p *QueryParser) paramToString(param interface{}) string {
	str := fmt.Sprintf("%v", param)
	// Remove surrounding quotes if present
	if len(str) >= 2 && ((str[0] == '\'' && str[len(str)-1] == '\'') ||
		(str[0] == '"' && str[len(str)-1] == '"')) {
		return str[1 : len(str)-1]
	}
	return str
}

// Convenience functions for creating common queries

// G creates a new query builder (equivalent to 'g')
func G() *QueryBuilder {
	return NewQueryBuilder()
}

// Eq creates an equals predicate
func Eq(property string, value interface{}) Predicate {
	return &EqualsPredicate{Property: property, Value: value}
}

// Contains creates a contains predicate
func Contains(property string, value interface{}) Predicate {
	return &ContainsPredicate{Property: property, Value: value}
}

// Between creates a range predicate
func Between(property string, min, max interface{}) Predicate {
	return &RangePredicate{Property: property, Min: min, Max: max}
}
