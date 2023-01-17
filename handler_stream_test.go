package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerStream(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	docStore := NewTestDocumentStore(t)

	hnd := &HandlerStream{
		DocumentStore: docStore,
	}
	srv.RegisterHandler(
		"core/stream/request",
		hnd.HandleStreamRequest,
	)

	// create documents
	rootDoc := NewTestDocument(t)
	patchDoc := NewTestDocument(t)

	rootDocID := NewDocumentID(rootDoc)
	patchDocID := NewDocumentID(patchDoc)

	// store documents
	require.NoError(t, docStore.PutDocument(&DocumentEntry{
		DocumentID:       rootDocID,
		DocumentType:     "test/root",
		DocumentEncoding: "cbor",
		DocumentBytes:    MustMarshal(t, rootDoc),
	}))

	require.NoError(t, docStore.PutDocument(&DocumentEntry{
		DocumentID:       patchDocID,
		DocumentType:     "test/root",
		DocumentEncoding: "cbor",
		DocumentBytes:    MustMarshal(t, patchDoc),
		RootDocumentID:   &rootDocID,
	}))

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// ask for stream
	ctx := context.Background()
	res, err := RequestStream(ctx, ses, rootDocID)
	require.NoError(t, err)
	require.Len(t, res.PatchDocumentIDs, 1)
	require.Equal(t, rootDocID, res.RootDocumentID)
	require.Equal(t, []DocumentID{patchDocID}, res.PatchDocumentIDs)
}
