package nimona

import (
	"encoding/json"
	"testing"

	"github.com/jimeh/undent"
	"github.com/stretchr/testify/require"
)

func TestDocumentMap(t *testing.T) {
	fix := &CborFixture{
		String: "foo",
		Uint64: 42,
		Int64:  -42,
		Bytes:  []byte("bar"),
		Bool:   true,
		Map: &CborFixture{
			String: "foo",
		},
		RepeatedString: []string{"foo", "bar"},
		RepeatedUint64: []uint64{1, 2, 3},
		RepeatedInt64:  []int64{-1, -2, -3},
		RepeatedBytes:  [][]byte{[]byte("foo"), []byte("bar")},
		// RepeatedBool:   []bool{true, false},
		RepeatedMap: []*CborFixture{{
			String: "foo",
		}, {
			String: "bar",
		}},
	}

	// "repeatedbool": [
	//     true,
	//     false
	//   ],

	exp := `
	{
	  "$type": "test/fixture",
	  "bool": true,
	  "bytes": "YmFy",
	  "int64": -42,
	  "map": {
	    "$type": "test/fixture",
	    "string": "foo"
	  },
	  "repeatedbytes": [
	    "Zm9v",
	    "YmFy"
	  ],
	  "repeatedint64": [
	    -1,
	    -2,
	    -3
	  ],
	  "repeatedmap": [
	    {
	      "$type": "test/fixture",
	      "string": "foo"
	    },
	    {
	      "$type": "test/fixture",
	      "string": "bar"
	    }
	  ],
	  "repeatedstring": [
	    "foo",
	    "bar"
	  ],
	  "repeateduint64": [
	    1,
	    2,
	    3
	  ],
	  "string": "foo",
	  "uint64": 42
	}`

	fixMap := fix.DocumentMap()

	t.Run("test converting to map", func(t *testing.T) {
		b, err := json.MarshalIndent(fixMap, "", "  ")
		require.NoError(t, err)
		require.Equal(t, undent.String(exp), string(b))
	})

	t.Run("test converting from map", func(t *testing.T) {
		g := &CborFixture{}
		g.FromDocumentMap(fixMap)
		require.Equal(t, fix, g)
	})

	t.Run("test CBOR encoding", func(t *testing.T) {
		b, err := cborEncoder.Marshal(fixMap)
		require.NoError(t, err)

		g := DocumentMap{}
		err = g.UnmarshalCBOR(b)
		require.NoError(t, err)
		require.EqualValues(t, fixMap, g)
	})
}
