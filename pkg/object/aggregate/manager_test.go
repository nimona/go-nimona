package aggregate_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"time"

	"nimona.io/internal/context"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/aggregate"
	"nimona.io/pkg/object/dag"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/object/mutation"
)

//
//    o
//   / \
//  m1  m2
//  |   | \
//  m3  m4 |
//   \ /   |
//    m6   m5
//
// o Ffsa2mABctpZ1rTpguU1N65GaDVMnbMHW3sLvJ3cAVri
// m1 EnFp6PUXJd7UckwkMpzFD9iVPGwRnYJEVc5ADtjHF7rj
// m2 7oPoh9GC5wRt8xXUFCBpySYYS6V5pbz9PzxexPYi33et
// m3 FkAHo36tu1zUqiX1eBwUq8AWUcRDgjRcDtKTPz31YiBa
// m4 AU1qfwJEAmxCZgRSVkW9FX4ZQBzXeE8nEgQtMoXKpUmP
// m5 GXLmHnazQGu6obWWgsi1dgPRwjHQ4xxVj6oYpZTvYk4V
// m6 FMSQefm7qnspxSH13zXkCX2nhvC7CJL2e1BfX5SM3t7S
//

var (
	o = object.FromMap(map[string]interface{}{
		"@ctx:s": "foo",
		"foo:s":  "bar",
		"numbers:ai": []int{
			1, 2, 3,
		},
		"strings:as": []string{
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

	m1 = &mutation.Mutation{
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

	m2 = &mutation.Mutation{
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

	m3 = &mutation.Mutation{
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

	m4 = &mutation.Mutation{
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

	m5 = &mutation.Mutation{
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

	m6 = &mutation.Mutation{
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
)

func Test_manager_Put(t *testing.T) {
	tests := []struct {
		name    string
		object  *object.Object
		want    map[string]interface{}
		wantErr error
	}{
		{
			name:   "put base object",
			object: o,
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "bar",
				"numbers:ai": []int{
					1, 2, 3,
				},
				"strings:as": []string{
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
			},
		},
		{
			name:   "apply m1, assign foo",
			object: m1.ToObject(),
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "not-bar",
				"numbers:ai": []int{
					1, 2, 3,
				},
				"strings:as": []string{
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
			},
		},
		{
			name:   "apply m2, append 4",
			object: m2.ToObject(),
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "not-bar",
				"numbers:ai": []int{
					1, 2, 3, 4,
				},
				"strings:as": []string{
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
			},
		},
		{
			name:   "apply m3, append d",
			object: m3.ToObject(),
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "not-bar",
				"numbers:ai": []int{
					1, 2, 3, 4,
				},
				"strings:as": []string{
					"a", "b", "c", "d",
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
			},
		},
		{
			name:   "apply m4, assign nested foo",
			object: m4.ToObject(),
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "not-bar",
				"numbers:ai": []int{
					1, 2, 3, 4,
				},
				"strings:as": []string{
					"a", "b", "c", "d",
				},
				"map:o": map[string]interface{}{
					"nested-foo:s": "not-nested-bar",
					"nested-numbers:ai": []interface{}{
						1, 2, 3,
					},
					"nested-strings:as": []interface{}{
						"a", "b", "c",
					},
				},
			},
		},
		{
			name:   "apply m5, append nested 9",
			object: m5.ToObject(),
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "not-bar",
				"numbers:ai": []int{
					1, 2, 3, 4,
				},
				"strings:as": []string{
					"a", "b", "c", "d",
				},
				"map:o": map[string]interface{}{
					"nested-foo:s": "not-nested-bar",
					"nested-numbers:ai": []interface{}{
						1, 2, 3, 9,
					},
					"nested-strings:as": []interface{}{
						"a", "b", "c",
					},
				},
			},
		},
		{
			name:   "apply m6, append nested z",
			object: m6.ToObject(),
			want: map[string]interface{}{
				"@ctx:s": "foo",
				"foo:s":  "not-bar",
				"numbers:ai": []int{
					1, 2, 3, 4,
				},
				"strings:as": []string{
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
			},
		},
	}

	x := &exchange.MockExchange{}
	x.On("Handle", mock.Anything, mock.Anything).Return(nil, nil)

	os, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	li, err := net.NewLocalInfo("", pk)
	assert.NoError(t, err)

	d, err := dag.New(os, x, nil, li)
	assert.NoError(t, err)
	assert.NotNil(t, d)

	m, err := aggregate.New(os, x, d)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	a := make(chan *aggregate.AggregateObject, 10)
	m.Subscribe(a)

	for _, tt := range tests {
		err := d.Put(tt.object)
		assert.Equal(t, tt.wantErr, err, tt.name)
		got := <-a
		assert.Equal(t,
			jp(tt.want),
			jp(got.Aggregate.ToMap()),
			tt.name,
		)
	}

	os.(*graph.Cayley).Dump() // nolint
}

func TestAppend(t *testing.T) {
	ctx := context.Background()

	x := &exchange.MockExchange{}
	x.On("Handle", mock.Anything, mock.Anything).Return(nil, nil)

	os, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	li, err := net.NewLocalInfo("", pk)
	assert.NoError(t, err)

	d, err := dag.New(os, x, nil, li)
	assert.NoError(t, err)
	assert.NotNil(t, d)

	m, err := aggregate.New(os, x, d)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	err = d.Put(o)
	assert.NoError(t, err)

	err = d.Put(m1.ToObject())
	assert.NoError(t, err)

	err = d.Put(m2.ToObject())
	assert.NoError(t, err)

	err = d.Put(m3.ToObject())
	assert.NoError(t, err)

	err = d.Put(m4.ToObject())
	assert.NoError(t, err)

	err = d.Put(m5.ToObject())
	assert.NoError(t, err)

	// HACK(geoah) to wait for the dust to settle, this is needed as all Put()s
	// result in the aggregate being generated and published, which takes time.
	// This resulted in getting the aggregate generated by m5 in our channel.
	time.Sleep(time.Millisecond * 100)

	a := make(chan *aggregate.AggregateObject, 10)
	m.Subscribe(a)

	em6 := map[string]interface{}{
		"@ctx:s": "foo",
		"foo:s":  "not-bar",
		"numbers:ai": []int{
			1, 2, 3, 4,
		},
		"strings:as": []string{
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
	}

	err = m.Append(o.HashBase58(), m6.Operations...)
	assert.NoError(t, err)

	g, err := d.Get(ctx, o.HashBase58())
	assert.NoError(t, err)
	assert.Len(t, g.Objects, 7)

	var got *aggregate.AggregateObject
	select {
	case got = <-a:
	case <-time.After(time.Second):
	}
	assert.Equal(t, jp(em6), jp(got.Aggregate.ToMap()))
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}
