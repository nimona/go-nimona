package nimona

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jimeh/undent"
	"github.com/stretchr/testify/require"

	"nimona.io/tilde"
)

func TestDocumentMap(t *testing.T) {
	fix := &documentFixture{
		String: "foo",
		Uint64: 42,
		Int64:  -42,
		Bytes:  []byte("bar"),
		Bool:   true,
		MapPtr: &documentFixture{
			String: "foo",
		},
		RepeatedString: []string{"foo", "bar"},
		RepeatedUint64: []uint64{1, 2, 3},
		RepeatedInt64:  []int64{-1, -2, -3},
		RepeatedBytes:  [][]byte{[]byte("foo"), []byte("bar")},
		// RepeatedBool:   []bool{true, false},
		RepeatedMap: []documentFixture{{
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
	  "mapPtr:m": {
	    "$type:s": "test/fixture",
	    "string:s": "foo",
	    "stringConst:s": "foo"
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
	      "string:s": "foo",
	      "stringConst:s": "foo"
	    },
	    {
	      "$type:s": "test/fixture",
	      "string:s": "bar",
	      "stringConst:s": "foo"
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
	  "stringConst:s": "foo",
	  "uint64:u": 42
	}`

	fixMap := fix.Document()

	t.Run("test converting to map", func(t *testing.T) {
		b, err := json.MarshalIndent(fixMap, "", "  ")
		require.NoError(t, err)
		require.Equal(t, undent.String(exp), string(b))
	})

	t.Run("test converting from map", func(t *testing.T) {
		g := &documentFixture{}
		g.FromDocument(fixMap)
		require.Equal(t, fix, g)
	})
}

func removeUnderscoreKeys(m tilde.Map) {
	for k := range m {
		if k[0] == '_' {
			delete(m, k)
			continue
		}
		if v, ok := m[k].(tilde.Map); ok {
			removeUnderscoreKeys(v)
		}
	}
}

func NewTestDocument(t *testing.T) *Document {
	t.Helper()
	doc := &documentFixture{
		String: uuid.New().String(),
	}
	return doc.Document()
}

func EqualDocument(
	t require.TestingT,
	expected *Document,
	actual *Document,
	msgAndArgs ...interface{},
) {
	expectedMap := tilde.Copy(expected.Map())
	actualMap := tilde.Copy(actual.Map())

	// remove all underscore prefixed keys
	removeUnderscoreKeys(expectedMap)
	removeUnderscoreKeys(actualMap)

	require.EqualValues(t, expectedMap, actualMap, msgAndArgs...)
}
