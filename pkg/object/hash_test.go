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

func TestFromValue(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{{
		name: "5",
		json: `{"something:s":"foo","metadata:m":{}}`,
		want: "QmT1f4wLjS9tbf6WRUMHQAfqt7UT48qTBb1J75tmerc5rm",
	}, {
		name: "6",
		json: `{"something:s":"foo"}`,
		want: "QmT1f4wLjS9tbf6WRUMHQAfqt7UT48qTBb1J75tmerc5rm",
	}, {
		name: "7",
		json: `{"something:d":"Zm9v"}`,
		want: "QmXocPFUysHQxuNifQaFyYLgcndxDvmnu1hs6BCwVHKnUJ",
	}, {
		name: "8",
		json: `{"something:b":false}`,
		want: "Qmaw1rVTP2y7Y4taU1GgHyyZ2BFdmdF3uA2exP4NanHjBj",
	}, {
		name: "9",
		json: `{"something:b":true}`,
		want: "QmaJoWsfScjaB3P2iEEKgWPmR1iRtG9hAswoj2w84W3FhB",
	}, {
		name: "10",
		json: `{"something:i":1234567890}`,
		want: "QmQMB1ajKQNHpBoHmYV4vBxjoyCJn7e7c8aGkTZRoYpHte",
	}, {
		name: "11",
		json: `{"something:f":12345.6789}`,
		want: "Qme79NMzWUYi2sybKZ9j4o6cmni1BAiZkFvRTbtoDPMWNn",
	}, {
		name: "13",
		json: `{"something:as":["foo","bar"]}`,
		want: "QmSKCPJmDJEXR51YVCd2kjR1Kw8gDp2qJMSCiUwaZPnNM7",
	}, {
		name: "14",
		json: `{"something:ai":[123,456]}`,
		want: "QmWk9fcxpHS4wPTprnnmi9mPevmyP21bwhfbWGtjqV2arT",
	}, {
		name: "15",
		json: `{"foo:s":"bar"}`,
		want: "QmZP9BNzNEzxp8QnQYePUyCWEXwqUCvB16W53AoKmGFQhw",
	}, {
		name: "17",
		// nolint: lll
		json: `{"data:m":{"foo:s":"bar","nested:m":{"_sig:s":"should not matter","foo:s":"bar"}}}`,
		want: "QmNNgkh9Yi2qPFaZQQYEkQDVjZkAU81ALg1htJukoD23wm",
	}, {
		name: "18",
		// nolint: lll
		json: `{"data:m":{"foo:s":"bar","nested:m":{"_signature:m":{"foo:s":"bar"},"foo:s":"bar"}}}`,
		want: "QmNNgkh9Yi2qPFaZQQYEkQDVjZkAU81ALg1htJukoD23wm",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Map{}
			assert.NoError(t, json.Unmarshal([]byte(tt.json), &m))
			got, err := fromValue(m)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.B58String())
		})
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
		json: `{"type:s":"foo","data:m":{"foo:h":"` + string(inner.Hash()) + `"}}`,
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
