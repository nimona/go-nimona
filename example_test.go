package nimona

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/tilde"
)

func TestExample_Graph(t *testing.T) {
	// these are considered well known for the purposes of this test
	var providerAddress PeerAddr
	var peerOneIdentity KeyGraphID

	t.Run("setup server", func(t *testing.T) {
		srvCtx := context.Background()
		srvPub, srvPrv, err := GenerateKey()
		require.NoError(t, err)

		srvTransport := &TransportUTP{}
		srvListener, err := srvTransport.Listen(srvCtx, "127.0.0.1:0")
		require.NoError(t, err)

		srv, err := NewSessionManager(srvTransport, srvListener, srvPub, srvPrv)
		require.NoError(t, err)

		providerAddress = srv.PeerAddr()
		docStore := NewTestDocumentStore(t)

		// register handlers
		HandlePingRequest(srv)
		HandleDocumentRequest(srv, docStore)
		HandleDocumentStoreRequest(srv, docStore)
		HandleDocumentGraphRequest(srv, docStore)
	})

	fakeResolver := &ResolverFake{
		identities: map[string]*IdentityInfo{
			"nimona.dev": {
				Alias: IdentityAlias{
					Hostname: "nimona.dev",
				},
				KeyGraphID: NewTestKeyGraphID(t),
				PeerAddresses: []PeerAddr{
					providerAddress,
				},
			},
		},
	}

	t.Run("setup peer one", func(t *testing.T) {
		// generate client keypair
		cltPeerPub, cltPeerPrv, err := GenerateKey()
		require.NoError(t, err)

		// construct transport and start listening
		cltTransport := &TransportUTP{}
		cltListener, err := cltTransport.Listen(context.Background(), "127.0.0.1:0")
		require.NoError(t, err)

		// construct session manager
		csm, err := NewSessionManager(cltTransport, cltListener, cltPeerPub, cltPeerPrv)
		require.NoError(t, err)

		// replace resolver with a fake one
		csm.resolver = fakeResolver

		// create a new request context
		rctx := NewTestRequestContext(t)
		rctx.KeyGraphID = NewTestKeyGraphID(t)
		peerOneIdentity = rctx.KeyGraphID

		// construct a new document store
		docStore := NewTestDocumentStore(t)

		// create new profile
		profile := &Profile{
			Metadata: Metadata{
				Owner: rctx.KeyGraphID,
			},
		}
		profileDoc := profile.Document()
		profileDocID := NewDocumentID(profileDoc)

		DumpDocument(profileDoc)

		// store profile
		err = docStore.PutDocument(profileDoc)
		require.NoError(t, err)

		// verify that the document is in the store
		gotDoc, err := docStore.GetDocument(profileDocID)
		require.NoError(t, err)
		EqualDocument(t, profileDoc, gotDoc)

		// publish profile
		ctx := context.Background()
		prv := FromAlias(IdentityAlias{Hostname: "nimona.dev"})
		err = PublishDocument(ctx, csm, rctx, profileDoc, prv)
		require.NoError(t, err)

		// create profile patch
		profilePatch, err := docStore.CreatePatch(
			profileDocID,
			"replace",
			"displayName",
			tilde.String("John Doe"),
			SigningContext{
				KeyGraphID: rctx.KeyGraphID,
				PrivateKey: rctx.PrivateKey,
			},
		)
		require.NoError(t, err)

		fmt.Println("profilePatch")
		DumpDocument(profilePatch)

		err = PublishDocument(ctx, csm, rctx, profilePatch, prv)
		require.NoError(t, err)
	})

	t.Run("setup peer two", func(t *testing.T) {
		// generate client keypair
		cltPeerPub, cltPeerPrv, err := GenerateKey()
		require.NoError(t, err)

		// construct transport and start listening
		cltTransport := &TransportUTP{}
		cltListener, err := cltTransport.Listen(context.Background(), "127.0.0.1:0")
		require.NoError(t, err)

		// construct session manager
		csm, err := NewSessionManager(cltTransport, cltListener, cltPeerPub, cltPeerPrv)
		require.NoError(t, err)

		// replace resolver with a fake one
		csm.resolver = fakeResolver

		// create a new request context
		rctx := NewTestRequestContext(t)
		rctx.DocumentStore = NewTestDocumentStore(t)

		// figure out the profile document id to request
		profile := &Profile{
			Metadata: Metadata{
				Owner: peerOneIdentity,
			},
		}
		profileDoc := profile.Document()
		profileDocID := NewDocumentID(profileDoc)

		// get graph
		ctx := context.Background()
		prv := FromAlias(IdentityAlias{Hostname: "nimona.dev"})
		res, err := RequestDocumentGraph(ctx, rctx, csm, profileDocID, prv)
		require.NoError(t, err)
		require.Len(t, res.PatchDocumentIDs, 1)

		// get root document
		gotRootDoc, gotPatches, err := SyncDocumentGraph(ctx, rctx, csm, res.RootDocumentID, prv)
		require.NoError(t, err)

		// print graph
		DumpDocument(gotRootDoc)
		for _, p := range gotPatches {
			DumpDocument(p.Document())
		}

		// create aggregate document
		gotAggregateDoc, err := ApplyDocumentPatch(gotRootDoc, gotPatches...)
		require.NoError(t, err)
		EqualDocument(t, NewDocument(tilde.Map{
			"$type": tilde.String("core/identity/profile"),
			"$metadata": tilde.Map{
				"owner": peerOneIdentity.TildeValue(),
			},
			"displayName": tilde.String("John Doe"),
		}), gotAggregateDoc)

		// check stored documents
		storedRootDoc, err := rctx.DocumentStore.GetDocument(profileDocID)
		require.NoError(t, err)
		EqualDocument(t, gotRootDoc, storedRootDoc)

		storedPatchDocs, err := rctx.DocumentStore.GetDocumentsByRootID(profileDocID)
		require.NoError(t, err)
		require.Len(t, storedPatchDocs, 1)
		EqualDocument(t, gotPatches[0].Document(), storedPatchDocs[0])
	})
}
