package provider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestProvider_handleAnnouncement(t *testing.T) {
	// net0 is our provider
	net0 := newPeer(t)
	pr0 := &peer.ConnectionInfo{
		PublicKey: net0.LocalPeer().GetPeerKey().PublicKey(),
		Addresses: net0.GetAddresses(),
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
		PeerVector: hyperspace.New("foo", "bar"),
	}

	// construct provider
	prv, err := New(context.New(), net0, nil)
	require.NoError(t, err)

	// net1 announces to provider
	err = net1.Send(
		context.New(),
		object.MustMarshal(pr1),
		pr0.PublicKey,
		network.SendWithConnectionInfo(pr0),
	)
	require.NoError(t, err)

	// wait a bit and check if provder has cached the peer
	time.Sleep(100 * time.Millisecond)
	// the second peer is our own
	assert.Len(t, prv.peerCache.List(), 2)
}

func TestProvider_distributeAnnouncement(t *testing.T) {
	// net0 is our provider
	net0 := newPeer(t)
	pr0 := &peer.ConnectionInfo{
		PublicKey: net0.LocalPeer().GetPeerKey().PublicKey(),
		Addresses: net0.GetAddresses(),
	}

	// net1 is another provider
	net1 := newPeer(t)

	// net 2 is a normal peer
	net2 := newPeer(t)
	pr2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: net2.LocalPeer().GetPeerKey().PublicKey(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: net2.LocalPeer().GetPeerKey().PublicKey(),
			Addresses: net2.GetAddresses(),
		},
		PeerVector:       hyperspace.New("foo", "bar"),
		PeerCapabilities: []string{"foo", "bar"},
	}

	// construct providers
	time.Sleep(250 * time.Millisecond)
	prv0, err := New(
		context.New(),
		net0,
		nil,
	)
	require.NoError(t, err)
	prv1, err := New(
		context.New(),
		net1,
		[]*peer.ConnectionInfo{pr0},
	)
	require.NoError(t, err)

	// net2 announces to provider 0
	time.Sleep(250 * time.Millisecond)
	err = net2.Send(
		context.New(),
		object.MustMarshal(pr2),
		pr0.PublicKey,
		network.SendWithConnectionInfo(pr0),
	)
	require.NoError(t, err)

	// wait a bit and check if both provder have cached the peer
	time.Sleep(250 * time.Millisecond)
	_, existsInPrv1 := prv0.peerCache.Get(pr2.ConnectionInfo.PublicKey)
	assert.NoError(t, existsInPrv1)
	time.Sleep(250 * time.Millisecond)
	_, existsInPrv2 := prv1.peerCache.Get(pr2.ConnectionInfo.PublicKey)
	assert.NoError(t, existsInPrv2)
}

func TestProvider_handlePeerLookup(t *testing.T) {
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
	}

	// net1 is a normal peer
	net1 := newPeer(t)

	// construct provider
	prv, err := New(context.New(), net0, nil)
	require.NoError(t, err)

	// add a couple more random peers to the provider's cache
	pr2k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	pr2 := &hyperspace.Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: pr2k.PublicKey(),
		},
		PeerVector: hyperspace.New("foo", "bar"),
	}
	pr3k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	pr3 := &hyperspace.Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: pr3k.PublicKey(),
		},
		PeerVector: hyperspace.New("foo"),
	}
	prv.Put(pr2)
	prv.Put(pr3)

	// start listening for lookup responses on net1
	sub := net1.Subscribe(
		network.FilterByObjectType(new(hyperspace.LookupResponse).Type()),
	)

	// lookup "foo" as net1
	err = net1.Send(
		context.New(),
		object.MustMarshal(
			&hyperspace.LookupRequest{
				Nonce:       "1",
				QueryVector: hyperspace.New("foo", "bar"),
			},
		),
		pr0.ConnectionInfo.PublicKey,
		network.SendWithConnectionInfo(pr0.ConnectionInfo),
	)
	require.NoError(t, err)

	// wait for response
	env, err := sub.Next()
	require.NoError(t, err)

	// check response
	res := &hyperspace.LookupResponse{}
	err = res.UnmarshalObject(env.Payload)
	require.NoError(t, err)
	assert.Equal(t, "1", res.Nonce)
	assert.ElementsMatch(t, []*hyperspace.Announcement{pr2}, res.Announcements)
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
