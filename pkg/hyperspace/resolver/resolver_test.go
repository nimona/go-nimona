package resolver

import (
	"database/sql"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

func TestResolver_Integration(t *testing.T) {
	// net0 is our provider
	k0, net0 := newPeer(t)
	pr0 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k0.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: k0.PublicKey(),
			Addresses: net0.Addresses(),
		},
		PeerCapabilities: []string{"foo", "bar"},
	}

	// net1 is a normal peer
	k1, net1 := newPeer(t)
	pr1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k1.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: k1.PublicKey(),
			Addresses: net1.Addresses(),
		},
		PeerCapabilities: []string{"foo"},
	}

	// construct provider
	prv, err := provider.New(context.New(), net0, k0, nil)
	require.NoError(t, err)

	// net1 announces to provider
	c0, err := net1.Dial(context.New(), pr0.ConnectionInfo)
	require.NoError(t, err)

	err = c0.Write(
		context.New(),
		object.MustMarshal(pr1),
	)
	require.NoError(t, err)

	p2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	p3, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// add a couple more random peers to the provider's cache
	pr2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: p2.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: p2.PublicKey(),
		},
		PeerVector: hyperspace.New("foo", "bar"),
	}
	pr3 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k0.PublicKey().DID(),
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
		k1,
		str1,
		WithBoostrapPeers(pr0.ConnectionInfo),
	)

	// lookup by content
	pr, err := res.Lookup(context.New(), LookupByHash(tilde.Digest("bar")))
	require.NoError(t, err)
	assert.ElementsMatch(t, []*peer.ConnectionInfo{pr2.ConnectionInfo}, pr)

	t.Run("object added", func(t *testing.T) {
		// add new object to pr1 store
		obj1 := &object.Object{
			Data: tilde.Map{
				"foo": tilde.String("bar"),
			},
		}
		obj1hash := obj1.Hash()
		err = str1.Put(obj1)
		require.NoError(t, err)

		time.Sleep(250 * time.Millisecond)

		// lookup by hash
		pr, err := res.Lookup(context.New(), LookupByHash(obj1hash))
		require.NoError(t, err)
		assert.Len(t, pr, 1)
		assert.Equal(t,
			pr1.Metadata.Signature.Signer.String(),
			pr[0].Metadata.Signature.Signer.String(),
		)
	})
}

func newPeer(t *testing.T) (crypto.PrivateKey, net.Network) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	ctx := context.New()

	n := net.New(k)
	lis, err := n.Listen(
		ctx,
		"127.0.0.1:0",
		&net.ListenConfig{
			BindLocal: true,
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		lis.Close() // nolint: errcheck
	})

	return k, n
}

func tempObjectStore(t *testing.T) *sqlobjectstore.Store {
	t.Helper()
	db, err := sql.Open("sqlite", path.Join(t.TempDir(), "sqlite3.db"))
	require.NoError(t, err)
	str, err := sqlobjectstore.New(db)
	require.NoError(t, err)
	return str
}
