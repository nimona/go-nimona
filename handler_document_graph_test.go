package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerStream(t *testing.T) {
	// construct server and client
	srv, clt := newTestSessionManager(t)
	srvDocStore := NewTestDocumentStore(t)

	// handle requests
	HandleDocumentGraphRequest(srv, srvDocStore)

	// create documents
	rootDoc := NewTestDocument(t)
	patchDoc := NewTestDocument(t)

	rootDocID := NewDocumentID(rootDoc)
	patchDoc.Metadata.Root = &rootDocID

	patchDocID := NewDocumentID(patchDoc)

	// store documents
	require.NoError(t, srvDocStore.PutDocument(rootDoc))
	require.NoError(t, srvDocStore.PutDocument(patchDoc))

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// ask for stream
	ctx := context.Background()
	res, err := RequestDocumentGraph(ctx, rootDocID, ses)
	require.NoError(t, err)
	require.Len(t, res.PatchDocumentIDs, 1)
	require.Equal(t, rootDocID, res.RootDocumentID)
	require.Equal(t, []DocumentID{patchDocID}, res.PatchDocumentIDs)
}
