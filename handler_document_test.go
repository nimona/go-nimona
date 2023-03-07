package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerDocument(t *testing.T) {
	ctx := context.Background()

	// Create new session manager
	srv, clt := newTestSessionManager(t)

	// Create new document store
	store := NewTestDocumentStore(t)

	// Create new document
	doc := NewTestDocument(t)
	docID := NewDocumentID(doc)

	// Add document to store
	err := store.PutDocument(doc)
	require.NoError(t, err)

	// Construct a new HandlerDocument
	HandleDocumentRequest(srv, store)

	// Dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// construct new request context
	rctx := NewTestRequestContext(t)

	// Request document
	gotDoc, err := RequestDocument(ctx, rctx, docID, ses)
	require.NoError(t, err)
	require.Equal(t, doc, gotDoc)
}
