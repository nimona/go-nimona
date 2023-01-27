package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStream_ApplyStreamPatch(t *testing.T) {
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

	p, err := CreateStreamPatch(aCbor, bCbor)
	require.NoError(t, err)

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

	aCbor, err := MarshalCBORBytes(a)
	require.NoError(t, err)

	bCbor, err := MarshalCBORBytes(b)
	require.NoError(t, err)

	p, err := CreateStreamPatch(aCbor, bCbor)
	require.NoError(t, err)

	id := NewTestIdentity(t).IdentityID()
	p.Metadata.Owner = id

	p.Dependencies = []DocumentID{{
		DocumentHash: NewRandomHash(t),
	}}
}
