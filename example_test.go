package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/tilde"
)

func TestExample_Graph(t *testing.T) {
	var srvAddr string

	// create test documents
	origProfile := Profile{DisplayName: "John"}
	origRootDoc := origProfile.Document()
	origRootDocID := NewDocumentID(origRootDoc)

	origProfilePatch := DocumentPatch{
		Metadata: Metadata{
			Root: &origRootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:    "replace",
			Path:  "displayName",
			Value: tilde.String("John Doe"),
		}},
	}
	origPatchDoc := origProfilePatch.Document()

	DumpDocument(origRootDoc)
	DumpDocument(origPatchDoc)

	t.Run("setup server", func(t *testing.T) {
		srvCtx := context.Background()
		srvPub, srvPrv, err := GenerateKey()
		require.NoError(t, err)

		srvTransport := &TransportUTP{}
		srvListener, err := srvTransport.Listen(srvCtx, "127.0.0.1:0")
		require.NoError(t, err)

		srv, err := NewSessionManager(srvTransport, srvListener, srvPub, srvPrv)
		require.NoError(t, err)

		srvAddr = srv.PeerAddr().String()
		docStore := NewTestDocumentStore(t)

		// register handlers
		HandleDocumentRequest(srv, docStore)
		HandleDocumentGraphRequest(srv, docStore)

		// store documents
		require.NoError(t, docStore.PutDocument(origRootDoc))
		require.NoError(t, docStore.PutDocument(origPatchDoc))
	})

	srvPeerAddr, err := ParsePeerAddr(srvAddr)
	require.NoError(t, err)

	// setup client
	cltPub, cltPrv, err := GenerateKey()
	require.NoError(t, err)

	cltTransport := &TransportUTP{}
	cltListener, err := cltTransport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)

	csm, err := NewSessionManager(cltTransport, cltListener, cltPub, cltPrv)
	require.NoError(t, err)

	// dial the server
	ses, err := csm.Dial(context.Background(), *srvPeerAddr)
	require.NoError(t, err)

	// construct request context
	rctx := RequestContext{
		PublicKey:  cltPub,
		PrivateKey: cltPrv,
	}

	// get graph
	ctx := context.Background()
	res, err := RequestDocumentGraph(ctx, origRootDocID, ses)
	require.NoError(t, err)
	require.Len(t, res.PatchDocumentIDs, 1)
	require.Equal(t, origRootDocID, res.RootDocumentID)

	// get all documents
	gotRootDoc, err := RequestDocument(ctx, rctx, origRootDocID, ses)
	require.NoError(t, err)
	require.Equal(t, origRootDoc, gotRootDoc)
	require.Equal(t, origRootDocID, NewDocumentID(gotRootDoc))

	// get patch documents
	gotPatches := []*DocumentPatch{}
	gotPatchDocs := []*Document{}
	for _, gotPatchDocID := range res.PatchDocumentIDs {
		gotPatchDoc, err := RequestDocument(ctx, rctx, gotPatchDocID, ses)
		require.NoError(t, err)
		gotPatch := &DocumentPatch{}
		gotPatch.FromDocument(gotPatchDoc)
		gotPatches = append(gotPatches, gotPatch)
		gotPatchDocs = append(gotPatchDocs, gotPatchDoc)
	}
	require.Len(t, gotPatches, 1)
	require.Len(t, gotPatchDocs, 1)
	EqualDocument(t, origPatchDoc, gotPatchDocs[0])

	// create aggregate document
	gotAggregateDoc, err := ApplyDocumentPatch(gotRootDoc, gotPatches...)
	require.NoError(t, err)
	require.Equal(t, NewDocument(tilde.Map{
		"$type":       tilde.String("core/identity/profile"),
		"displayName": tilde.String("John Doe"),
	}), gotAggregateDoc)
}
