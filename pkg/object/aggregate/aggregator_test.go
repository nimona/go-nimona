package aggregate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/object"
	"nimona.io/pkg/object/mutation"
)

func TestAggregator(t *testing.T) {
	a := NewAggregator()

	o := object.FromMap(map[string]interface{}{
		"foo:s": "bar",
		"numbers:ai": []interface{}{
			1, 2, 3,
		},
		"strings:as": []interface{}{
			"a", "b", "c",
		},
		"map:o": map[string]interface{}{
			"nested-foo:s": "bar",
			"nested-numbers:ai": []interface{}{
				1, 2, 3,
			},
			"nested-strings:as": []interface{}{
				"a", "b", "c",
			},
		},
	})

	m1 := &mutation.Mutation{
		Parents: []string{
			o.HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAssign,
				Cursor:    []string{"foo:s"},
				Value:     "not-bar",
			},
		},
	}

	m2 := &mutation.Mutation{
		Parents: []string{
			o.HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"numbers:ai"},
				Value:     4,
			},
		},
	}

	m3 := &mutation.Mutation{
		Parents: []string{
			m1.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"strings:as"},
				Value:     "d",
			},
		},
	}

	m4 := &mutation.Mutation{
		Parents: []string{
			m2.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAssign,
				Cursor:    []string{"map:o", "nested-foo:s"},
				Value:     "not-nested-bar",
			},
		},
	}

	m5 := &mutation.Mutation{
		Parents: []string{
			m2.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"map:o", "nested-numbers:ai"},
				Value:     9,
			},
		},
	}

	m6 := &mutation.Mutation{
		Parents: []string{
			m3.ToObject().HashBase58(),
			m4.ToObject().HashBase58(),
		},
		Operations: []*mutation.Operation{
			{
				Operation: mutation.OpAppend,
				Cursor:    []string{"map:o", "nested-strings:as"},
				Value:     "z",
			},
		},
	}

	eo := object.FromMap(map[string]interface{}{
		"foo:s": "not-bar",
		"numbers:ai": []interface{}{
			1, 2, 3, 4,
		},
		"strings:as": []interface{}{
			"a", "b", "c", "d",
		},
		"map:o": map[string]interface{}{
			"nested-foo:s": "not-nested-bar",
			"nested-numbers:ai": []interface{}{
				1, 2, 3, 9,
			},
			"nested-strings:as": []interface{}{
				"a", "b", "c", "z",
			},
		},
	})

	ao, err := a.Aggregate(o, []*mutation.Mutation{m1, m2, m3, m4, m5, m6})
	assert.NoError(t, err)
	assert.Equal(t, eo.ToMap(), ao.Aggregate.ToMap())

	// b, _ := json.MarshalIndent(ao.aggregate.ToMap(), "", "  ")
	// fmt.Println(string(b))
}
