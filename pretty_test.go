package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettySPrint(t *testing.T) {
	v := &CborFixture{
		String: "foo",
		Int64:  42,
	}

	exp := `cbor: a36524747970656c746573742f6669787475726566737472696e6763666f6f65696e743634182a
json: {"$type":"test/fixture","int64":42,"string":"foo"}
`

	s := PrettySPrint(v)
	require.Equal(t, exp, s)
}
