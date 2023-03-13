package nimona

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"nimona.io/tilde"
)

func TestExample_Graph(t *testing.T) {
	// these are considered well known for the purposes of this test
	var providerAddress PeerAddr
	var peerOneIdentity *Identity

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
				Identity: Identity{
					KeyGraphID: NewTestRandomDocumentID(t),
				},
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
		peerOneIdentity = rctx.Identity

		// create new profile and publish it
		profile := &Profile{
			Metadata: Metadata{
				Owner: rctx.Identity,
			},
		}
		profileDoc := profile.Document()
		profileDocID := NewDocumentID(profileDoc)
		ctx := context.Background()
		prv := FromAlias(IdentityAlias{Hostname: "nimona.dev"})
		err = RequestDocumentStore(ctx, csm, rctx, profileDoc, prv)
		require.NoError(t, err)

		// create profile patch and publish it
		profilePatch := &DocumentPatch{
			Metadata: Metadata{
				Owner:     rctx.Identity,
				Root:      &profileDocID,
				Parents:   []DocumentID{profileDocID},
				Sequence:  1,
				Timestamp: time.Now().Format(time.RFC3339),
			},
			Operations: []DocumentPatchOperation{{
				Op:    "replace",
				Path:  "displayName",
				Value: tilde.String("John Doe"),
			}},
		}
		profilePatchDoc := profilePatch.Document()
		err = RequestDocumentStore(ctx, csm, rctx, profilePatchDoc, prv)
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

		// create aggregate document
		gotAggregateDoc, err := ApplyDocumentPatch(gotRootDoc, gotPatches...)
		require.NoError(t, err)
		EqualDocument(t, NewDocument(tilde.Map{
			"$type":       tilde.String("core/identity/profile"),
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
