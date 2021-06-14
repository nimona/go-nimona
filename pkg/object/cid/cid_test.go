package cid

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"nimona.io/pkg/chore"
)

func BenchmarkCID(b *testing.B) {
	o := chore.Map{
		"@type":    chore.String("blob"),
		"filename": chore.String("foo"),
		"dummy": chore.Map{
			"@type": chore.String("dummy"),
			"@metadata": chore.Map{
				"owner": chore.String("foo"),
			},
			"foo": chore.String("bar"),
			"data": chore.Data(
				"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14" +
					"\n15\n16\n17\n18\n19\n20",
			),
		},
	}
	for n := 0; n < b.N; n++ {
		FromValue(o) // nolint: errcheck
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

func TestMapRedaction(t *testing.T) {
	inner := chore.Map{
		"@type": chore.String("foo"),
		"foo":   chore.String("bar"),
	}
	parentWithInner := chore.Map{
		"@type": chore.String("foo"),
		"foo":   inner,
	}
	parentWithInnerRedacted := chore.Map{
		"@type": chore.String("foo"),
		"foo":   Must(New(inner)),
	}
	tests := []struct {
		name      string
		json      string
		wantValue chore.Value
		wantCID   chore.CID
		wantErr   bool
	}{{
		name:      "1",
		json:      `{"@type:s":"foo","foo:s":"bar"}`,
		wantValue: inner,
		wantCID:   Must(New(inner)),
	}, {
		name: "2",
		// nolint: lll
		json:      `{"@type:s":"foo","foo:m":{"@type:s":"foo","foo:s":"bar"}}`,
		wantValue: parentWithInner,
		wantCID:   Must(New(parentWithInner)),
	}, {
		name: "3",
		// nolint: lll
		json:      `{"@type:s":"foo","foo:r":"` + string(Must(New(inner))) + `"}`,
		wantValue: parentWithInnerRedacted,
		wantCID:   Must(New(parentWithInner)),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := chore.Map{}
			assert.NoError(t, json.Unmarshal([]byte(tt.json), &m))
			assert.Equal(t, tt.wantValue, m)
			fmt.Println(tt.json)
			got, err := New(m)
			require.NoError(t, err)
			assert.Equal(t, tt.wantCID, got)
		})
	}
}

func TestEdgecases_NullStream(t *testing.T) {
	// nolint: lll
	b := `{"@type:s":"stream:poc.nimona.io/conversation","nonce:s":"44273fc3-5bd0-4ed5-a9eb-3abb588f68cd","metadata:m":{"owner:s":"@peer","stream:r":null,"datetime:s":"2021-02-14T20:51:38.989872"}}`
	o := chore.Map{}
	err := json.Unmarshal([]byte(b), &o)
	require.NoError(t, err)
	h := Must(New(o))
	assert.NotEmpty(t, h)
}
