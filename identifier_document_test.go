package nimona

import (
	"testing"
)

func TestDocumentID_Marshal(t *testing.T) {
	id := &DocumentID{
		DocumentHash: []byte("foo"),
	}
	PrettyPrint(id)
}
