package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerDocument(t *testing.T) {
	ctx := context.Background()

	// Create new peer configs
	srvPeerConfig := NewTestPeerConfig(t)
	clnPeerConfig := NewTestPeerConfig(t)

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
	hnd := &HandlerDocument{
		Hostname:      "testing.nimona.io",
		PeerConfig:    srvPeerConfig,
		DocumentStore: store,
	}
	srv.RegisterHandler(
		"core/document.request",
		hnd.HandleDocumentRequest,
	)

	// Dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// Request document
	gotDoc, err := RequestDocument(ctx, ses, clnPeerConfig, docID)
	require.NoError(t, err)
	require.Equal(t, doc, gotDoc)
}
