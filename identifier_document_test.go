package nimona

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDocumentID_Marshal(t *testing.T) {
	doc := &CborFixture{
		String: "foo",
	}

	hash, err := NewDocumentHash(doc)
	require.NoError(t, err)

	id := &DocumentID{
		DocumentHash: hash,
	}

	PrettyPrint(id)
}

func TestDocumentHash_NewTestRandomDocumentID(t *testing.T) {
	t.Run("test random document id", func(t *testing.T) {
		h1 := NewTestRandomDocumentID(t)
		h2 := NewTestRandomDocumentID(t)
		require.NotEqual(t, h1, h2)
	})
}

// NewTestRandomDocumentID returns a random document ID for testing purposes.
func NewTestRandomDocumentID(t *testing.T) DocumentID {
	t.Helper()

	doc := &CborFixture{
		String: uuid.New().String(),
	}

	return NewDocumentID(doc)
}
