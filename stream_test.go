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

	p, err := CreateStreamPatch(a, b)
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

	p, err := CreateStreamPatch(a, b)
	require.NoError(t, err)

	p.Dependencies = []DocumentID{{
		DocumentHash: []byte("foo"),
	}}

	PrettyPrint(&p)
}
