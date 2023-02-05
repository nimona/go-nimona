package nimona

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentCodegen_WriteToFile(t *testing.T) {
	t.Skip("Used to generate fixture_gen.go")
	err := GenerateDocumentMapMethods(
		"document_codegen_fixture_docgen.go",
		"nimona",
		codegenFixture{},
		codegenFixtureWithType{},
	)
	require.NoError(t, err)
}

func TestDocumentCodegen(t *testing.T) {
	f1 := codegenFixture{
		String:         "string",
		Uint64:         64,
		Int64:          -64,
		Bytes:          []byte("bytes"),
		Bool:           true,
		MapPtr:         &codegenFixture{String: "nested"},
		RepeatedString: []string{"repeated", "string"},
		RepeatedUint64: []uint64{64, 64},
		RepeatedInt64:  []int64{-64, -64},
		RepeatedBytes:  [][]byte{[]byte("repeated"), []byte("bytes")},
		RepeatedBool:   []bool{true, false},
		RepeatedMap:    []codegenFixture{{String: "repeated", Uint64: 64}},
		RepeatedMapPtr: []*codegenFixture{{String: "repeated", Uint64: 64}},
	}

	m1 := f1.DocumentMap()

	require.Equal(t, "foo", m1["stringConst"])

	b1, err := json.MarshalIndent(m1, "", "  ")
	require.NoError(t, err)

	mm := map[string]interface{}{}
	err = json.Unmarshal(b1, &mm)
	require.NoError(t, err)

	f2 := &codegenFixture{}
	f2.FromDocumentMap(m1)
	require.EqualValues(t, f1, *f2)
}

func TestDocumentCodegen_WithType(t *testing.T) {
	f1 := codegenFixtureWithType{
		String: "string",
	}

	m1 := f1.DocumentMap()

	require.Equal(t, "foobar", m1["$type"])

	b1, err := json.MarshalIndent(m1, "", "  ")
	require.NoError(t, err)

	mm := map[string]interface{}{}
	err = json.Unmarshal(b1, &mm)
	require.NoError(t, err)

	f2 := &codegenFixtureWithType{}
	f2.FromDocumentMap(m1)
	require.EqualValues(t, f1, *f2)
}
