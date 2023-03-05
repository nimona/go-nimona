package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerStream(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	docStore := NewTestDocumentStore(t)

	hnd := &HandlerDocumentGraph{
		DocumentStore: docStore,
	}
	srv.RegisterHandler(
		"core/document/graph.request",
		hnd.HandleDocumentGraphRequest,
	)

	// create documents
	rootDoc := NewTestDocument(t)
	patchDoc := NewTestDocument(t)

	rootDocID := NewDocumentID(rootDoc)
	patchDoc.Metadata.Root = &rootDocID

	patchDocID := NewDocumentID(patchDoc)

	// store documents
	require.NoError(t, docStore.PutDocument(rootDoc))
	require.NoError(t, docStore.PutDocument(patchDoc))

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// ask for stream
	ctx := context.Background()
	res, err := RequestDocumentGraph(ctx, ses, rootDocID)
	require.NoError(t, err)
	require.Len(t, res.PatchDocumentIDs, 1)
	require.Equal(t, rootDocID, res.RootDocumentID)
	require.Equal(t, []DocumentID{patchDocID}, res.PatchDocumentIDs)
}
