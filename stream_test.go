package nimona

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/prettycbor"
)

func TestStream_ApplyStreamPatch(t *testing.T) {
	a := &CborFixture{
		String: "foo",
	}

	b := &CborFixture{
		String: "bar",
		Int64:  42,
	}

	aCbor, err := a.MarshalCBORBytes()
	require.NoError(t, err)

	bCbor, err := b.MarshalCBORBytes()
	require.NoError(t, err)

	p, err := CreateStreamPatch(aCbor, bCbor)
	require.NoError(t, err)

	pCbor, err := p.MarshalCBORBytes()
	require.NoError(t, err)

	m, err := NewDocumentMap(p)
	require.NoError(t, err)
	mb, err := json.MarshalIndent(m, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(mb))
	fmt.Println(prettycbor.Dump(pCbor))

	err = ApplyStreamPatch(a, p)
	require.NoError(t, err)
	require.Equal(t, b, a)
}

func TestStream_CreateStreamPatch(t *testing.T) {
	a := &CborFixture{
		String: "foo",
	}

	b := &CborFixture{
		String: "bar",
		Int64:  42,
	}

	aCbor, err := a.MarshalCBORBytes()
	require.NoError(t, err)

	bCbor, err := b.MarshalCBORBytes()
	require.NoError(t, err)

	p, err := CreateStreamPatch(aCbor, bCbor)
	require.NoError(t, err)

	p.Metadata.Owner = "foo"

	p.Dependencies = []DocumentID{{
		DocumentHash: NewRandomHash(t),
	}}

	PrettyPrint(p)
}
