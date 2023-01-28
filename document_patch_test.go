package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentPatch_ApplyDocumentPatch(t *testing.T) {
	a := &CborFixture{
		String: "foo",
	}

	b := &CborFixture{
		String: "bar",
		Int64:  42,
	}

	aCbor, err := MarshalCBORBytes(a)
	require.NoError(t, err)

	bCbor, err := MarshalCBORBytes(b)
	require.NoError(t, err)

	p, err := CreateDocumentPatch(aCbor, bCbor)
	require.NoError(t, err)

	err = ApplyDocumentPatch(a, p)
	require.NoError(t, err)
	require.Equal(t, b, a)
}

func TestDocumentPatch_CreateDocumentPatch(t *testing.T) {
	a := &CborFixture{
		String: "foo",
	}

	b := &CborFixture{
		String: "bar",
		Int64:  42,
	}

	aCbor, err := MarshalCBORBytes(a)
	require.NoError(t, err)

	bCbor, err := MarshalCBORBytes(b)
	require.NoError(t, err)

	p, err := CreateDocumentPatch(aCbor, bCbor)
	require.NoError(t, err)

	id := NewTestIdentity(t)
	p.Metadata.Owner = id

	p.Dependencies = []DocumentID{{
		DocumentHash: NewRandomHash(t),
	}}
}
