package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/tilde"
)

func TestExample_Graph(t *testing.T) {
	peerOneContext := NewTestRequestContext(t)
	peerTwoContext := NewTestRequestContext(t)

	var providerAddress PeerAddr

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

		// create new profile and publish it
		profile := &Profile{
			Metadata: Metadata{
				Owner: peerOneContext.Identity,
			},
		}
		profileDoc := profile.Document()
		profileDocID := NewDocumentID(profileDoc)
		ctx := context.Background()
		prv := FromAlias(IdentityAlias{Hostname: "nimona.dev"})
		err = RequestDocumentStore(ctx, csm, peerOneContext, profileDoc, prv)
		require.NoError(t, err)

		// create profile patch and publish it
		profilePatch := &DocumentPatch{
			Metadata: Metadata{
				Owner: peerOneContext.Identity,
				Root:  &profileDocID,
			},
			Operations: []DocumentPatchOperation{{
				Op:    "replace",
				Path:  "displayName",
				Value: tilde.String("John Doe"),
			}},
		}
		profilePatchDoc := profilePatch.Document()
		err = RequestDocumentStore(ctx, csm, peerOneContext, profilePatchDoc, prv)
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

		// figure out the profile document id to request
		profile := &Profile{
			Metadata: Metadata{
				Owner: peerOneContext.Identity,
			},
		}
		profileDoc := profile.Document()
		profileDocID := NewDocumentID(profileDoc)

		// get graph
		ctx := context.Background()
		prv := FromAlias(IdentityAlias{Hostname: "nimona.dev"})
		res, err := RequestDocumentGraph(ctx, peerTwoContext, csm, profileDocID, prv)
		require.NoError(t, err)
		require.Len(t, res.PatchDocumentIDs, 1)

		// get all documents
		gotRootDoc, err := RequestDocument(ctx, peerTwoContext, csm, res.RootDocumentID, prv)
		require.NoError(t, err)

		// get patch documents
		gotPatches := []*DocumentPatch{}
		gotPatchDocs := []*Document{}
		for _, gotPatchDocID := range res.PatchDocumentIDs {
			gotPatchDoc, err := RequestDocument(ctx, peerTwoContext, csm, gotPatchDocID, prv)
			require.NoError(t, err)
			gotPatch := &DocumentPatch{}
			gotPatch.FromDocument(gotPatchDoc)
			gotPatches = append(gotPatches, gotPatch)
			gotPatchDocs = append(gotPatchDocs, gotPatchDoc)
		}
		require.Len(t, gotPatches, 1)
		require.Len(t, gotPatchDocs, 1)

		// create aggregate document
		gotAggregateDoc, err := ApplyDocumentPatch(gotRootDoc, gotPatches...)
		require.NoError(t, err)
		EqualDocument(t, NewDocument(tilde.Map{
			"$type":       tilde.String("core/identity/profile"),
			"displayName": tilde.String("John Doe"),
		}), gotAggregateDoc)
	})
}
