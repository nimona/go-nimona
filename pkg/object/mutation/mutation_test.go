package mutation

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/object"
)

func TestMutation(t *testing.T) {
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

	m := &Mutation{
		Operations: []*Operation{
			{
				Operation: OpAssign,
				Cursor:    []string{"foo:s"},
				Value:     "not-bar",
			},
			{
				Operation: OpAppend,
				Cursor:    []string{"numbers:ai"},
				Value:     4,
			},
			{
				Operation: OpAppend,
				Cursor:    []string{"strings:as"},
				Value:     "d",
			},
			{
				Operation: OpAssign,
				Cursor:    []string{"map:o", "nested-foo:s"},
				Value:     "not-nested-bar",
			},
			{
				Operation: OpAppend,
				Cursor:    []string{"map:o", "nested-numbers:ai"},
				Value:     9,
			},
			{
				Operation: OpAppend,
				Cursor:    []string{"map:o", "nested-strings:as"},
				Value:     "z",
			},
		},
	}

	om := m.ToObject().ToMap()
	assert.Equal(t, "/object.mutation", om["@ctx:s"])

	err := m.Mutate(o)
	assert.NoError(t, err)
	assert.Equal(t, eo, o)
	b, _ := json.MarshalIndent(o.ToMap(), "", "  ")
	fmt.Println(string(b))
}
