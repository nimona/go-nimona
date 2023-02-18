package nimona

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/tilde"
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

	gotConst, err := m1.m.Get("stringConst")
	require.NoError(t, err)
	require.Equal(t, tilde.String("foo"), gotConst)

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

	gotType, err := m1.m.Get("$type")
	require.NoError(t, err)
	require.Equal(t, tilde.String("foobar"), gotType)

	b1, err := json.MarshalIndent(m1, "", "  ")
	require.NoError(t, err)

	mm := map[string]interface{}{}
	err = json.Unmarshal(b1, &mm)
	require.NoError(t, err)

	f2 := &codegenFixtureWithType{}
	f2.FromDocumentMap(m1)
	require.EqualValues(t, f1, *f2)
}
