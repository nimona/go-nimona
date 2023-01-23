package nimona

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentMap(t *testing.T) {
	m := &CborFixture{
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

	exp := `{
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

	t.Run("test converting to map", func(t *testing.T) {
		m, err := NewDocumentMap(m)
		require.NoError(t, err)

		b, err := json.MarshalIndent(m, "", "  ")
		fmt.Println(string(b))
		require.NoError(t, err)
		require.Equal(t, exp, string(b))
	})
}
