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

func TestProvider_handlePeer(t *testing.T) {
	// net0 is our provider
	net0 := newPeer(t)
	pr0 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: net0.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: net0.LocalPeer().GetAddresses(),
	}

	// net1 is a normal peer
	net1 := newPeer(t)
	pr1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: net1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		QueryVector: hyperspace.New("foo", "bar"),
		Addresses:   net1.LocalPeer().GetAddresses(),
	}

	// construct provider
	prv, err := New(context.New(), net0)
	require.NoError(t, err)

	// net1 announces to provider
	err = net1.Send(
		context.New(),
		pr1.ToObject(),
		pr0,
	)
	require.NoError(t, err)

	// wait a bit and check if provder has cached the peer
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, prv.peerCache.List(), 1)
}

func TestProvider_handlePeerLookup(t *testing.T) {
	// net0 is our provider
	net0 := newPeer(t)
	pr0 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: net0.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: net0.LocalPeer().GetAddresses(),
	}

	// net1 is a normal peer
	net1 := newPeer(t)

	// construct provider
	prv, err := New(context.New(), net0)
	require.NoError(t, err)

	// add a couple more random peers to the provider's cache
	pr2 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: "a",
		},
		QueryVector: hyperspace.New("foo", "bar"),
	}
	pr3 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: "b",
		},
		QueryVector: hyperspace.New("foo"),
	}
	prv.Put(pr2)
	prv.Put(pr3)

	// start listening for lookup responses on net1
	sub := net1.Subscribe(
		network.FilterByObjectType(new(peer.LookupResponse).Type()),
	)

	// lookup "foo" as net1
	err = net1.Send(
		context.New(),
		peer.LookupRequest{
			Nonce:       "1",
			QueryVector: hyperspace.New("foo", "bar"),
		}.ToObject(),
		pr0,
	)
	require.NoError(t, err)

	// wait for response
	env, err := sub.Next()
	require.NoError(t, err)

	// check response
	res := &peer.LookupResponse{}
	err = res.FromObject(env.Payload)
	require.NoError(t, err)
	assert.Equal(t, "1", res.Nonce)
	assert.ElementsMatch(t, []*peer.Peer{pr2}, res.Peers)
}

func newPeer(t *testing.T) network.Network {
	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	ctx := context.New()

	local := localpeer.New()
	local.PutPrimaryPeerKey(k)

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
