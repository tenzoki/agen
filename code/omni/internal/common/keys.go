package common

import (
	"fmt"
	"strings"
)

const (
	KVPrefix       = "kv:"
	VertexPrefix   = "v:"
	EdgePrefix     = "e:"
	IndexPrefix    = "idx:"
	MetadataPrefix = "meta:"
)

const (
	OutgoingIndexPrefix = IndexPrefix + "out:"
	IncomingIndexPrefix = IndexPrefix + "in:"
	VertexTypePrefix    = IndexPrefix + "vtype:"
	EdgeTypePrefix      = IndexPrefix + "etype:"
	PropertyIndexPrefix = IndexPrefix + "prop:"
)

const (
	GraphMetadataKey = MetadataPrefix + "graph"
	SchemaKey        = MetadataPrefix + "schema"
)

type KeyBuilder struct{}

func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

func (kb *KeyBuilder) KVKey(key string) []byte {
	return []byte(KVPrefix + key)
}

func (kb *KeyBuilder) VertexKey(vertexID string) []byte {
	return []byte(VertexPrefix + vertexID)
}

func (kb *KeyBuilder) EdgeKey(edgeID string) []byte {
	return []byte(EdgePrefix + edgeID)
}

func (kb *KeyBuilder) OutgoingEdgeIndexKey(vertexID, edgeID string) []byte {
	return []byte(OutgoingIndexPrefix + vertexID + ":" + edgeID)
}

func (kb *KeyBuilder) IncomingEdgeIndexKey(vertexID, edgeID string) []byte {
	return []byte(IncomingIndexPrefix + vertexID + ":" + edgeID)
}

func (kb *KeyBuilder) VertexTypeIndexKey(vertexType, vertexID string) []byte {
	return []byte(VertexTypePrefix + vertexType + ":" + vertexID)
}

func (kb *KeyBuilder) EdgeTypeIndexKey(edgeType, edgeID string) []byte {
	return []byte(EdgeTypePrefix + edgeType + ":" + edgeID)
}

func (kb *KeyBuilder) PropertyIndexKey(propertyName, value, entityID string) []byte {
	return []byte(PropertyIndexPrefix + propertyName + ":" + value + ":" + entityID)
}

func (kb *KeyBuilder) GraphMetadataKey() []byte {
	return []byte(GraphMetadataKey)
}

func (kb *KeyBuilder) SchemaKey() []byte {
	return []byte(SchemaKey)
}

func (kb *KeyBuilder) OutgoingEdgePrefix(vertexID string) []byte {
	return []byte(OutgoingIndexPrefix + vertexID + ":")
}

func (kb *KeyBuilder) IncomingEdgePrefix(vertexID string) []byte {
	return []byte(IncomingIndexPrefix + vertexID + ":")
}

func (kb *KeyBuilder) VertexTypePrefix(vertexType string) []byte {
	return []byte(VertexTypePrefix + vertexType + ":")
}

func (kb *KeyBuilder) EdgeTypePrefix(edgeType string) []byte {
	return []byte(EdgeTypePrefix + edgeType + ":")
}

func (kb *KeyBuilder) PropertyPrefix(propertyName, value string) []byte {
	return []byte(PropertyIndexPrefix + propertyName + ":" + value + ":")
}

func (kb *KeyBuilder) KVPrefix() []byte {
	return []byte(KVPrefix)
}

func (kb *KeyBuilder) AllVerticesPrefix() []byte {
	return []byte(VertexPrefix)
}

func (kb *KeyBuilder) AllEdgesPrefix() []byte {
	return []byte(EdgePrefix)
}

type KeyParser struct{}

func NewKeyParser() *KeyParser {
	return &KeyParser{}
}

func (kp *KeyParser) ParseKVKey(key []byte) (string, bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, KVPrefix) {
		return "", false
	}
	return strings.TrimPrefix(keyStr, KVPrefix), true
}

func (kp *KeyParser) ParseVertexKey(key []byte) (string, bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, VertexPrefix) {
		return "", false
	}
	return strings.TrimPrefix(keyStr, VertexPrefix), true
}

func (kp *KeyParser) ParseEdgeKey(key []byte) (string, bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, EdgePrefix) {
		return "", false
	}
	return strings.TrimPrefix(keyStr, EdgePrefix), true
}

func (kp *KeyParser) ParseOutgoingIndexKey(key []byte) (vertexID, edgeID string, ok bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, OutgoingIndexPrefix) {
		return "", "", false
	}

	suffix := strings.TrimPrefix(keyStr, OutgoingIndexPrefix)
	parts := strings.SplitN(suffix, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func (kp *KeyParser) ParseIncomingIndexKey(key []byte) (vertexID, edgeID string, ok bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, IncomingIndexPrefix) {
		return "", "", false
	}

	suffix := strings.TrimPrefix(keyStr, IncomingIndexPrefix)
	parts := strings.SplitN(suffix, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func (kp *KeyParser) ParseVertexTypeIndexKey(key []byte) (vertexType, vertexID string, ok bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, VertexTypePrefix) {
		return "", "", false
	}

	suffix := strings.TrimPrefix(keyStr, VertexTypePrefix)
	parts := strings.SplitN(suffix, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func (kp *KeyParser) ParseEdgeTypeIndexKey(key []byte) (edgeType, edgeID string, ok bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, EdgeTypePrefix) {
		return "", "", false
	}

	suffix := strings.TrimPrefix(keyStr, EdgeTypePrefix)
	parts := strings.SplitN(suffix, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func (kp *KeyParser) ParsePropertyIndexKey(key []byte) (propertyName, value, entityID string, ok bool) {
	keyStr := string(key)
	if !strings.HasPrefix(keyStr, PropertyIndexPrefix) {
		return "", "", "", false
	}

	suffix := strings.TrimPrefix(keyStr, PropertyIndexPrefix)
	parts := strings.SplitN(suffix, ":", 3)
	if len(parts) != 3 {
		return "", "", "", false
	}

	return parts[0], parts[1], parts[2], true
}

func (kp *KeyParser) GetKeyType(key []byte) string {
	keyStr := string(key)

	switch {
	case strings.HasPrefix(keyStr, KVPrefix):
		return "kv"
	case strings.HasPrefix(keyStr, VertexPrefix):
		return "vertex"
	case strings.HasPrefix(keyStr, EdgePrefix):
		return "edge"
	case strings.HasPrefix(keyStr, OutgoingIndexPrefix):
		return "outgoing_index"
	case strings.HasPrefix(keyStr, IncomingIndexPrefix):
		return "incoming_index"
	case strings.HasPrefix(keyStr, VertexTypePrefix):
		return "vertex_type_index"
	case strings.HasPrefix(keyStr, EdgeTypePrefix):
		return "edge_type_index"
	case strings.HasPrefix(keyStr, PropertyIndexPrefix):
		return "property_index"
	case strings.HasPrefix(keyStr, MetadataPrefix):
		return "metadata"
	default:
		return "unknown"
	}
}

func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if len(key) > 1024 {
		return fmt.Errorf("key length cannot exceed 1024 characters")
	}

	return nil
}
