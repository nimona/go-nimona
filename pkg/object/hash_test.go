package object

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkHash(b *testing.B) {
	o := Map{
		"type": String("blob"),
		"data": Map{
			"filename": String("foo"),
			"dummy": Map{
				"type": String("dummy"),
				"metadata": Map{
					"owner": String("foo"),
				},
				"data": Map{
					"foo": String("bar"),
					"data": Data(
						"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14" +
							"\n15\n16\n17\n18\n19\n20",
					),
				},
			},
		},
	}
	for n := 0; n < b.N; n++ {
		fromValue(o) // nolint: errcheck
	}
}

func TestRaw(t *testing.T) {
	r := []byte("aa")
	h, _ := mhFromBytes("t", r)
	c := mhToCid(h)
	fmt.Println(c)
	g, err := mhFromCid(c)
	require.NoError(t, err)
	require.Equal(t, h, g)
}

func TestObjectReplace(t *testing.T) {
	inner := &Object{
		Type: "foo",
		Data: Map{
			"foo": String("bar"),
		},
	}
	parentWithInner := &Object{
		Type: "foo",
		Data: Map{
			"foo": inner,
		},
	}
	tests := []struct {
		name    string
		json    string
		want    Hash
		wantErr bool
	}{{
		name: "1",
		json: `{"type:s":"foo","data:m":{"foo:s":"bar"}}`,
		want: inner.Hash(),
	}, {
		name: "2",
		// nolint: lll
		json: `{"type:s":"foo","data:m":{"foo:o":{"type:s":"foo","data:m":{"foo:s":"bar"}}}}`,
		want: parentWithInner.Hash(),
	}, {
		name: "3",
		// nolint: lll
		json: `{"type:s":"foo","data:m":{"foo:r":"` + string(inner.Hash()) + `"}}`,
		want: parentWithInner.Hash(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Object{}
			assert.NoError(t, json.Unmarshal([]byte(tt.json), o))
			got, err := NewHash(o)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
