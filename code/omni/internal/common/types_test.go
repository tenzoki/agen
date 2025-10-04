package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVertexCreation(t *testing.T) {
	vertex := NewVertex("test:1", "TestType")

	assert.Equal(t, "test:1", vertex.ID)
	assert.Equal(t, "TestType", vertex.Type)
	assert.Equal(t, uint64(1), vertex.Version)
	assert.NotNil(t, vertex.Properties)
	assert.False(t, vertex.CreatedAt.IsZero())
	assert.False(t, vertex.UpdatedAt.IsZero())
}

func TestVertexValidation(t *testing.T) {
	tests := []struct {
		name    string
		vertex  *Vertex
		wantErr bool
	}{
		{
			name:   "valid vertex",
			vertex: NewVertex("test:1", "TestType"),
		},
		{
			name: "empty ID",
			vertex: &Vertex{
				Type:      "TestType",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			wantErr: true,
		},
		{
			name: "empty type",
			vertex: &Vertex{
				ID:        "test:1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   1,
			},
			wantErr: true,
		},
		{
			name: "zero version",
			vertex: &Vertex{
				ID:        "test:1",
				Type:      "TestType",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Version:   0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vertex.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVertexProperties(t *testing.T) {
	vertex := NewVertex("test:1", "TestType")

	vertex.SetProperty("name", "Test Name")
	vertex.SetProperty("age", 25)
	vertex.SetProperty("active", true)

	name, exists := vertex.GetProperty("name")
	assert.True(t, exists)
	assert.Equal(t, "Test Name", name)

	age, exists := vertex.GetProperty("age")
	assert.True(t, exists)
	assert.Equal(t, 25, age)

	active, exists := vertex.GetProperty("active")
	assert.True(t, exists)
	assert.Equal(t, true, active)

	nonExistent, exists := vertex.GetProperty("nonexistent")
	assert.False(t, exists)
	assert.Nil(t, nonExistent)
}

func TestVertexSerialization(t *testing.T) {
	original := NewVertex("test:1", "TestType")
	original.SetProperty("name", "Test Name")
	original.SetProperty("age", 25)

	data, err := original.MarshalBinary()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored := &Vertex{}
	err = restored.UnmarshalBinary(data)
	require.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.Version, restored.Version)
	assert.Equal(t, original.CreatedAt.Unix(), restored.CreatedAt.Unix())
	assert.Equal(t, original.UpdatedAt.Unix(), restored.UpdatedAt.Unix())

	name, exists := restored.GetProperty("name")
	assert.True(t, exists)
	assert.Equal(t, "Test Name", name)

	age, exists := restored.GetProperty("age")
	assert.True(t, exists)
	assert.Equal(t, int8(25), age)
}

func TestEdgeCreation(t *testing.T) {
	edge := NewEdge("test:1", "TestEdgeType", "from:1", "to:1")

	assert.Equal(t, "test:1", edge.ID)
	assert.Equal(t, "TestEdgeType", edge.Type)
	assert.Equal(t, "from:1", edge.FromVertex)
	assert.Equal(t, "to:1", edge.ToVertex)
	assert.Equal(t, uint64(1), edge.Version)
	assert.Equal(t, 1.0, edge.Weight)
	assert.NotNil(t, edge.Properties)
	assert.False(t, edge.CreatedAt.IsZero())
}

func TestEdgeValidation(t *testing.T) {
	tests := []struct {
		name    string
		edge    *Edge
		wantErr bool
	}{
		{
			name: "valid edge",
			edge: NewEdge("test:1", "TestType", "from:1", "to:1"),
		},
		{
			name: "empty ID",
			edge: &Edge{
				Type:       "TestType",
				FromVertex: "from:1",
				ToVertex:   "to:1",
				CreatedAt:  time.Now(),
				Version:    1,
			},
			wantErr: true,
		},
		{
			name: "empty type",
			edge: &Edge{
				ID:         "test:1",
				FromVertex: "from:1",
				ToVertex:   "to:1",
				CreatedAt:  time.Now(),
				Version:    1,
			},
			wantErr: true,
		},
		{
			name: "empty from vertex",
			edge: &Edge{
				ID:        "test:1",
				Type:      "TestType",
				ToVertex:  "to:1",
				CreatedAt: time.Now(),
				Version:   1,
			},
			wantErr: true,
		},
		{
			name: "self loop",
			edge: &Edge{
				ID:         "test:1",
				Type:       "TestType",
				FromVertex: "vertex:1",
				ToVertex:   "vertex:1",
				CreatedAt:  time.Now(),
				Version:    1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.edge.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEdgeSerialization(t *testing.T) {
	original := NewEdge("test:1", "TestType", "from:1", "to:1")
	original.SetProperty("strength", "strong")
	original.Weight = 0.8

	data, err := original.MarshalBinary()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	restored := &Edge{}
	err = restored.UnmarshalBinary(data)
	require.NoError(t, err)

	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Type, restored.Type)
	assert.Equal(t, original.FromVertex, restored.FromVertex)
	assert.Equal(t, original.ToVertex, restored.ToVertex)
	assert.Equal(t, original.Weight, restored.Weight)
	assert.Equal(t, original.Version, restored.Version)

	strength, exists := restored.GetProperty("strength")
	assert.True(t, exists)
	assert.Equal(t, "strong", strength)
}

func TestDirectionString(t *testing.T) {
	assert.Equal(t, "incoming", Incoming.String())
	assert.Equal(t, "outgoing", Outgoing.String())
	assert.Equal(t, "both", Both.String())
	assert.Equal(t, "unknown", Direction(999).String())
}

func TestTraversalStrategyString(t *testing.T) {
	assert.Equal(t, "bfs", BFS.String())
	assert.Equal(t, "dfs", DFS.String())
	assert.Equal(t, "unknown", TraversalStrategy(999).String())
}

func TestCardinalityString(t *testing.T) {
	assert.Equal(t, "one_to_one", OneToOne.String())
	assert.Equal(t, "one_to_many", OneToMany.String())
	assert.Equal(t, "many_to_one", ManyToOne.String())
	assert.Equal(t, "many_to_many", ManyToMany.String())
	assert.Equal(t, "unknown", Cardinality(999).String())
}
