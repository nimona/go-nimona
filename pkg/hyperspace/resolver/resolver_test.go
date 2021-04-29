package resolver

import (
	"testing"

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
		pr1.ToObject(),
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
	res := New(context.New(), net1, WithBoostrapPeers(pr0.ConnectionInfo))

	// lookup by content
	pr, err := res.Lookup(context.New(), LookupByCID("bar"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []*peer.ConnectionInfo{pr2.ConnectionInfo}, pr)
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
