package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandler_DocumentRequest(t *testing.T) {
	ctx := context.Background()

	// create new session manager
	srv, clt := newTestSessionManager(t)

	// create new document store
	store := NewTestDocumentStore(t)

	// create new document
	doc := NewTestDocument(t)
	docID := NewDocumentID(doc)

	// Add document to store
	err := store.PutDocument(doc)
	require.NoError(t, err)

	// construct a new HandlerDocument
	HandleDocumentRequest(srv, store)

	// construct new request context
	rctx := NewTestRequestContext(t)

	// request document
	gotDoc, err := RequestDocument(ctx, rctx, clt, docID, FromPeerAddr(srv.PeerAddr()))
	require.NoError(t, err)
	require.Equal(t, doc, gotDoc)
}

func TestHandler_DocumentStoreRequest(t *testing.T) {
	ctx := context.Background()

	// create new session manager
	srv, clt := newTestSessionManager(t)

	// create new document store
	store := NewTestDocumentStore(t)

	// start handling requests
	HandleDocumentStoreRequest(srv, store)
	HandleDocumentRequest(srv, store)

	// construct new request context
	rctx := NewTestRequestContext(t)

	// create new document
	doc := NewTestDocument(t)
	docID := NewDocumentID(doc)

	// request server to store document
	err := RequestDocumentStore(ctx, clt, rctx, doc, FromPeerAddr(srv.PeerAddr()))
	require.NoError(t, err)

	// request document back
	gotDoc, err := RequestDocument(ctx, rctx, clt, docID, FromPeerAddr(srv.PeerAddr()))
	require.NoError(t, err)
	require.Equal(t, doc, gotDoc)

	// verify that the document is in the store
	gotDoc, err = store.GetDocument(docID)
	require.NoError(t, err)
	require.Equal(t, doc, gotDoc)
}
