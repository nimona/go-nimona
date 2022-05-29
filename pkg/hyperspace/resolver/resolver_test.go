package resolver

import (
	"database/sql"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/connmanager"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/keystream"
	"nimona.io/pkg/keystreammock"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

func TestResolver_Integration(t *testing.T) {
	// net0 is our provider
	k0, n0 := newPeerConnManager(t)
	pr0 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(k0.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner:     peer.IDFromPublicKey(k0.PublicKey()),
			Addresses: n0.Addresses(),
		},
		PeerCapabilities: []string{"foo", "bar"},
	}

	// net1 is a normal peer
	k1, net1 := newPeer(t)
	pr1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(k1.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner:     peer.IDFromPublicKey(k1.PublicKey()),
			Addresses: net1.GetAddresses(),
		},
		PeerCapabilities: []string{"foo"},
	}

	// construct provider
	prv, err := provider.New(context.New(), n0, k0, nil)
	require.NoError(t, err)

	// net1 announces to provider
	err = net1.Send(
		context.New(),
		object.MustMarshal(pr1),
		pr0.ConnectionInfo.Owner,
		network.SendWithConnectionInfo(pr0.ConnectionInfo),
	)
	require.NoError(t, err)

	p2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	p3, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// add a couple more random peers to the provider's cache
	pr2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(p2.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner: peer.IDFromPublicKey(p2.PublicKey()),
		},
		Digests: []tilde.Digest{"foo", "bar"},
	}
	pr3 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(k0.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner: peer.IDFromPublicKey(p3.PublicKey()),
		},
		Digests: []tilde.Digest{"foo"},
	}
	prv.Put(pr2)
	prv.Put(pr3)

	// construct resolver
	str1 := tempObjectStore(t)
	ksm := keystreammock.NewMockManager(gomock.NewController(t))
	ksm.EXPECT().
		GetController().
		Return(nil, keystream.ErrControllerNotFound).
		AnyTimes()
	ksm.EXPECT().
		WaitForController(
			gomock.Any(),
		).
		Return(nil, fmt.Errorf("something")).
		AnyTimes()
	res := New(
		context.New(),
		net1,
		k1,
		str1,
		ksm,
		WithBoostrapPeers(pr0.ConnectionInfo),
	)

	// lookup by content
	pr, err := res.LookupByContent(context.New(), "bar")
	require.NoError(t, err)
	assert.ElementsMatch(t, []*peer.ConnectionInfo{pr2.ConnectionInfo}, pr)

	// lookup by owner
	pr, err = res.LookupByDID(context.New(), peer.IDFromPublicKey(p2.PublicKey()))
	require.NoError(t, err)
	assert.ElementsMatch(t, []*peer.ConnectionInfo{
		pr2.ConnectionInfo,
	}, pr)

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
		pr, err := res.LookupByContent(context.New(), obj1hash)
		require.NoError(t, err)
		assert.Len(t, pr, 1)
		assert.Equal(t,
			pr1.Metadata.Owner.String(),
			pr[0].Owner.String(),
		)
	})
}

func newPeer(t *testing.T) (
	crypto.PrivateKey,
	network.Network,
) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	ctx := context.New()

	nn := network.New(ctx, k)
	lis, err := nn.Listen(
		ctx,
		"127.0.0.1:0",
		network.ListenOnLocalIPs,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		lis.Close() // nolint: errcheck
	})

	return k, nn
}

func newPeerConnManager(t *testing.T) (
	crypto.PrivateKey,
	connmanager.ConnManager,
) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	ctx := context.New()

	n := connmanager.New(k)
	lis, err := n.Listen(
		ctx,
		"127.0.0.1:0",
		&connmanager.ListenConfig{
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
