package resolver

import (
	"database/sql"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

func TestResolver_Integration(t *testing.T) {
	// net0 is our provider
	net0 := newPeer(t)
	pr0 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: net0.LocalPeer().GetPeerKey().PublicKey(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: net0.LocalPeer().GetPeerKey().PublicKey(),
			Addresses: net0.GetAddresses(),
		},
		PeerCapabilities: []string{"foo", "bar"},
	}

	// net1 is a normal peer
	net1 := newPeer(t)
	pr1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: net1.LocalPeer().GetPeerKey().PublicKey(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: net1.LocalPeer().GetPeerKey().PublicKey(),
			Addresses: net1.GetAddresses(),
		},
		PeerCapabilities: []string{"foo"},
	}

	// construct provider
	prv, err := provider.New(context.New(), net0, nil)
	require.NoError(t, err)

	// net1 announces to provider
	err = net1.Send(
		context.New(),
		object.MustMarshal(pr1),
		pr0.ConnectionInfo.PublicKey,
		network.SendWithConnectionInfo(pr0.ConnectionInfo),
	)
	require.NoError(t, err)

	p2, err := crypto.NewEd25519PrivateKey(crypto.IdentityKey)
	require.NoError(t, err)

	p3, err := crypto.NewEd25519PrivateKey(crypto.IdentityKey)
	require.NoError(t, err)

	// add a couple more random peers to the provider's cache
	pr2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: p2.PublicKey(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: p2.PublicKey(),
		},
		PeerVector: hyperspace.New("foo", "bar"),
	}
	pr3 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: net0.LocalPeer().GetPeerKey().PublicKey(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: p3.PublicKey(),
		},
		PeerVector: hyperspace.New("foo"),
	}
	prv.Put(pr2)
	prv.Put(pr3)

	// construct resolver
	str1 := tempObjectStore(t)
	res := New(
		context.New(),
		net1,
		str1,
		WithBoostrapPeers(pr0.ConnectionInfo),
	)

	// lookup by content
	pr, err := res.Lookup(context.New(), LookupByCID("bar"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []*peer.ConnectionInfo{pr2.ConnectionInfo}, pr)

	t.Run("object added", func(t *testing.T) {
		// add new object to pr1 store
		obj1 := &object.Object{
			Data: object.Map{
				"foo": object.String("bar"),
			},
		}
		obj1cid := obj1.CID()
		err = str1.Put(obj1)
		require.NoError(t, err)

		time.Sleep(250 * time.Millisecond)

		// lookup by cid
		pr, err := res.Lookup(context.New(), LookupByCID(obj1cid))
		require.NoError(t, err)
		assert.Len(t, pr, 1)
		assert.Equal(t,
			pr1.Metadata.Signature.Signer.String(),
			pr[0].Metadata.Signature.Signer.String(),
		)
	})
}

func newPeer(t *testing.T) network.Network {
	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	ctx := context.New()

	local := localpeer.New()
	local.SetPeerKey(k)

	net := network.New(
		ctx,
		network.WithLocalPeer(local),
	)

	lis, err := net.Listen(
		ctx,
		"127.0.0.1:0",
		network.ListenOnLocalIPs,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		lis.Close() // nolint: errcheck
	})

	return net
}

func tempObjectStore(t *testing.T) *sqlobjectstore.Store {
	t.Helper()
	db, err := sql.Open("sqlite3", path.Join(t.TempDir(), "sqlite3.db"))
	require.NoError(t, err)
	str, err := sqlobjectstore.New(db)
	require.NoError(t, err)
	return str
}
