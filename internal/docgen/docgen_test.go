package docgen

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestWriteToFile(t *testing.T) {
	t.Skip("Used to generate fixture_gen.go")
	err := GenerateDocumentMapMethods("fixture_gen.go", "docgen", Fixture{})
	require.NoError(t, err)
}

func Test(t *testing.T) {
	f1 := Fixture{
		String:         "string",
		Uint64:         64,
		Int64:          -64,
		Bytes:          []byte("bytes"),
		Bool:           true,
		MapPtr:         &Fixture{String: "nested"},
		RepeatedString: []string{"repeated", "string"},
		RepeatedUint64: []uint64{64, 64},
		RepeatedInt64:  []int64{-64, -64},
		RepeatedBytes:  [][]byte{[]byte("repeated"), []byte("bytes")},
		RepeatedBool:   []bool{true, false},
		RepeatedMap:    []Fixture{{String: "repeated", Uint64: 64}},
		RepeatedMapPtr: []*Fixture{{String: "repeated", Uint64: 64}},
	}

	m1 := f1.DocumentMap()

	require.Equal(t, "foo", m1["stringConst"])

	b1, err := json.MarshalIndent(m1, "", "  ")
	require.NoError(t, err)
	mm := map[string]interface{}{}
	err = json.Unmarshal(b1, &mm)
	require.NoError(t, err)
	spew.Dump(mm)

	f2 := &Fixture{}
	f2.FromDocumentMap(m1)
	require.EqualValues(t, f1, *f2)
}
