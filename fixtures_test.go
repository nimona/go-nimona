package nimona

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func NewTestDocument(t *testing.T) DocumentMap {
	t.Helper()
	doc := &CborFixture{
		String: uuid.New().String(),
	}
	return doc.DocumentMap()
}

func MustMarshal(t *testing.T, v DocumentMapper) []byte {
	t.Helper()
	b, err := MarshalCBORBytes(v)
	require.NoError(t, err)
	return b
}
