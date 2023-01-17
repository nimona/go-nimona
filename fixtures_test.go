package nimona

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestDocument(t *testing.T) Cborer {
	t.Helper()
	doc := &CborFixture{
		String: uuid.New().String(),
	}
	return doc
}

func MustMarshal(t *testing.T, v Cborer) []byte {
	t.Helper()
	b, err := v.MarshalCBORBytes()
	require.NoError(t, err)
	return b
}
