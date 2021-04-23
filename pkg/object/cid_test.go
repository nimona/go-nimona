package object

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkCID(b *testing.B) {
	o := Map{
		"@type":    String("blob"),
		"filename": String("foo"),
		"dummy": Map{
			"@type": String("dummy"),
			"@metadata": Map{
				"owner": String("foo"),
			},
			"foo": String("bar"),
			"data": Data(
				"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14" +
					"\n15\n16\n17\n18\n19\n20",
			),
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
		want    CID
		wantErr bool
	}{{
		name: "1",
		json: `{"@type:s":"foo","foo:s":"bar"}`,
		want: inner.CID(),
	}, {
		name: "2",
		// nolint: lll
		json: `{"@type:s":"foo","foo:o":{"@type:s":"foo","foo:s":"bar"}}`,
		want: parentWithInner.CID(),
	}, {
		name: "3",
		// nolint: lll
		json: `{"@type:s":"foo","foo:r":"` + string(inner.CID()) + `"}`,
		want: parentWithInner.CID(),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Object{}
			assert.NoError(t, json.Unmarshal([]byte(tt.json), o))
			got, err := NewCID(o)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEdgecases_NullStream(t *testing.T) {
	// nolint: lll
	b := `{"@type:s":"stream:poc.nimona.io/conversation","nonce:s":"44273fc3-5bd0-4ed5-a9eb-3abb588f68cd","metadata:m":{"owner:s":"@peer","stream:r":null,"datetime:s":"2021-02-14T20:51:38.989872"}}`
	o := &Object{}
	err := json.Unmarshal([]byte(b), o)
	require.NoError(t, err)
	h := o.CID()
	assert.NotEmpty(t, h)
}
