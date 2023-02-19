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
	  "$type:s": "test/fixture",
	  "bool:b": true,
	  "bytes:d": "YmFy",
	  "int64:i": -42,
	  "map:m": {
	    "$type:s": "test/fixture",
	    "string:s": "foo"
	  },
	  "repeatedbytes:ad": [
	    "Zm9v",
	    "YmFy"
	  ],
	  "repeatedint64:ai": [
	    -1,
	    -2,
	    -3
	  ],
	  "repeatedmap:am": [
	    {
	      "$type:s": "test/fixture",
	      "string:s": "foo"
	    },
	    {
	      "$type:s": "test/fixture",
	      "string:s": "bar"
	    }
	  ],
	  "repeatedstring:as": [
	    "foo",
	    "bar"
	  ],
	  "repeateduint64:au": [
	    1,
	    2,
	    3
	  ],
	  "string:s": "foo",
	  "uint64:u": 42
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
}
