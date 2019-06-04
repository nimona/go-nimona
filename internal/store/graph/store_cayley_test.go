// +build flaky

package graph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/cayleygraph/cayley/graph/kv/bolt"

	"nimona.io/internal/store/graph"
	"nimona.io/pkg/object"
)

// var dependsOn = g.Morphism().In("<dependsOn>")
// g
// 	.V("<GNaN2foWa18FiBx3D5dU6qTkgj6ikJCuxgPBzXxqVSCp>")
// 	.Tag("source")
// 	.FollowRecursive(dependsOn)
// 	.Tag("target")
// 	.All()

var (
	o1 = object.FromMap(map[string]interface{}{
		"@ctx:s":     "message",
		"@display:s": "o1",
	})
	o2 = object.FromMap(map[string]interface{}{
		"@ctx:s":     "mutation",
		"@display:s": "o2",
		"@parents:as": []string{
			o1.HashBase58(),
		},
	})
	o3 = object.FromMap(map[string]interface{}{
		"@ctx:s":     "mutation",
		"@display:s": "o3",
		"@parents:as": []string{
			o2.HashBase58(),
		},
	})
	o4 = object.FromMap(map[string]interface{}{
		"@ctx:s":     "mutation",
		"@display:s": "o4",
		"@parents:as": []string{
			o2.HashBase58(),
		},
	})
	o5 = object.FromMap(map[string]interface{}{
		"@ctx:s":     "mutation",
		"@display:s": "o5",
		"@parents:as": []string{
			o3.HashBase58(),
			o4.HashBase58(),
		},
	})
	o6 = object.FromMap(map[string]interface{}{
		"@ctx:s":        "object",
		"@display:s":    "o6",
		"@parents:as": []string{},
	})
)

func TestCayley_Children(t *testing.T) {
	s, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	aos := []*object.Object{o1, o2, o3, o4, o5, o6}
	for _, ao := range aos {
		err = s.Put(ao)
		assert.NoError(t, err)
	}

	os, err := s.Children(o1.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 5)
}

func TestCayley_Graph(t *testing.T) {
	s, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	aos := []*object.Object{o3, o4, o5}
	for _, ao := range aos {
		err = s.Put(ao)
		assert.NoError(t, err)
	}

	os, err := s.Graph(o3.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 3)

	os, err = s.Graph(o4.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 3)

	os, err = s.Graph(o5.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 3)

	err = s.Put(o1)
	assert.NoError(t, err)

	err = s.Put(o2)
	assert.NoError(t, err)

	os, err = s.Graph(o1.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 5)

	os, err = s.Graph(o2.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 5)

	os, err = s.Graph(o3.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 5)

	os, err = s.Graph(o4.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 5)

	os, err = s.Graph(o5.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 5)
}

func TestCayley_Get(t *testing.T) {
	s, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	eo := object.FromMap(map[string]interface{}{
		"@ctx:s":     "message",
		"@display:s": "eo",
	})

	err = s.Put(eo)
	assert.NoError(t, err)

	o, err := s.Get(eo.HashBase58())
	assert.NoError(t, err)
	assert.Equal(t, eo, o)
}

func TestCayley_Heads(t *testing.T) {
	s, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	ox := object.FromMap(map[string]interface{}{
		"@ctx:s":        "something",
		"@display:s":    "ox",
		"@parents:as": []string{},
	})

	aos := []*object.Object{o1, o2, ox}
	for _, ao := range aos {
		err = s.Put(ao)
		assert.NoError(t, err)
	}

	os, err := s.Heads()
	assert.NoError(t, err)
	assert.Len(t, os, 2)
}

func TestCayley_Tails(t *testing.T) {
	s, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	ox1 := object.FromMap(map[string]interface{}{
		"@ctx:s":        "object",
		"@display:s":    "ox1",
		"@parents:as": []string{},
	})

	ox2 := object.FromMap(map[string]interface{}{
		"@ctx:s":     "mutation",
		"@display:s": "ox2",
		"@parents:as": []string{
			ox1.HashBase58(),
		},
	})

	aos := []*object.Object{o1, o2, o3, o4, ox1, ox2}
	for _, ao := range aos {
		err = s.Put(ao)
		assert.NoError(t, err)
	}

	os, err := s.Tails(o1.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 2)
	assert.Equal(t, o3.HashBase58(), os[0].HashBase58())
	assert.Equal(t, o4.HashBase58(), os[1].HashBase58())

	os, err = s.Tails(ox1.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, os, 1)
	assert.Equal(t, ox2.HashBase58(), os[0].HashBase58())
}

func TestCayley_Head(t *testing.T) {
	s, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	ox1 := object.FromMap(map[string]interface{}{
		"@ctx:s":        "object",
		"@display:s":    "ox1",
		"@parents:as": []string{},
	})

	ox2 := object.FromMap(map[string]interface{}{
		"@ctx:s":     "mutation",
		"@display:s": "ox2",
		"@parents:as": []string{
			ox1.HashBase58(),
		},
	})

	aos := []*object.Object{o1, o2, o3, o4, ox1, ox2}
	for _, ao := range aos {
		err = s.Put(ao)
		assert.NoError(t, err)
	}

	o, err := s.Head(o2.HashBase58())
	assert.NoError(t, err)
	assert.Equal(t, o1.HashBase58(), o.HashBase58())

	o, err = s.Head(ox2.HashBase58())
	assert.NoError(t, err)
	assert.Equal(t, ox1.HashBase58(), o.HashBase58())
}
