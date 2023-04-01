package nimona

import (
	"testing"

	"github.com/google/uuid"
)

// NewTestRandomDocumentID returns a random document ID for testing purposes.
func NewTestRandomDocumentID(t *testing.T) DocumentID {
	return DocumentID{
		DocumentHash: NewTestRandomDocumentHash(t),
	}
}

func NewTestRandomDocumentHash(t *testing.T) DocumentHash {
	t.Helper()

	doc := &documentFixture{
		String: uuid.New().String(),
	}

	return NewDocumentHash(doc.Document())
}
