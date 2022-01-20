package provider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestProvider_handleAnnouncement(t *testing.T) {
	// net0 is our provider
	net0, k0 := newPeer(t)
	pr0 := &peer.ConnectionInfo{
		PublicKey: k0.PublicKey(),
		Addresses: net0.Addresses(),
	}

	// net1 is a normal peer
	net1, k1 := newPeer(t)
	pr1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k1.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: k1.PublicKey(),
			Addresses: net1.Addresses(),
		},
		PeerVector: hyperspace.New("foo", "bar"),
	}

	// construct provider
	prv, err := New(context.New(), net0, k0, nil)
	require.NoError(t, err)

	// net1 announces to provider
	c0, err := net1.Dial(context.New(), pr0)
	require.NoError(t, err)
	err = c0.Write(
		context.New(),
		object.MustMarshal(pr1),
	)
	require.NoError(t, err)

	// wait a bit and check if provder has cached the peer
	time.Sleep(100 * time.Millisecond)
	// the second peer is our own
	assert.Len(t, prv.peerCache.List(), 2)
}

func TestProvider_distributeAnnouncement(t *testing.T) {
	// net0 is our provider
	net0, k0 := newPeer(t)
	pr0 := &peer.ConnectionInfo{
		PublicKey: k0.PublicKey(),
		Addresses: net0.Addresses(),
	}

	// net1 is another provider
	net1, k1 := newPeer(t)
	pr1 := &peer.ConnectionInfo{
		PublicKey: k1.PublicKey(),
		Addresses: net1.Addresses(),
	}

	// net 2 is a normal peer
	net2, k2 := newPeer(t)
	pr2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k2.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: k2.PublicKey(),
			Addresses: net2.Addresses(),
		},
		PeerVector:       hyperspace.New("foo", "bar"),
		PeerCapabilities: []string{"foo", "bar"},
	}

	// construct providers
	time.Sleep(time.Second)
	prv0, err := New(
		context.New(),
		net0,
		k0,
		[]*peer.ConnectionInfo{pr1},
	)
	require.NoError(t, err)
	prv1, err := New(
		context.New(),
		net1,
		k1,
		[]*peer.ConnectionInfo{pr0},
	)
	require.NoError(t, err)

	// net2 announces to provider 0
	time.Sleep(time.Second)
	c2, err := net2.Dial(context.New(), pr0)
	require.NoError(t, err)
	err = c2.Write(
		context.New(),
		object.MustMarshal(pr2),
	)
	require.NoError(t, err)

	// wait a bit and check if both providers have cached the peer
	time.Sleep(time.Second)
	_, existsInPrv1 := prv0.peerCache.Get(pr2.ConnectionInfo.PublicKey)
	assert.NoError(t, existsInPrv1)
	time.Sleep(time.Second)
	_, existsInPrv2 := prv1.peerCache.Get(pr2.ConnectionInfo.PublicKey)
	assert.NoError(t, existsInPrv2)
}

func TestProvider_handlePeerLookup(t *testing.T) {
	// net0 is our provider
	net0, k0 := newPeer(t)
	pr0 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k0.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: k0.PublicKey(),
			Addresses: net0.Addresses(),
		},
	}

	time.Sleep(time.Second)

	// net1 is a normal peer
	net1, k1 := newPeer(t)

	// construct provider
	prv, err := New(context.New(), net0, k1, nil)
	require.NoError(t, err)

	// start listening for lookup responses on net1
	resp := make(chan *object.Object)
	net1.RegisterConnectionHandler(
		func(c net.Connection) {
			go func() {
				or := c.Read(context.New())
				for {
					o, err := or.Read()
					if err != nil {
						return
					}
					if o.Type == hyperspace.LookupResponseType {
						resp <- o
						return
					}
				}
			}()
		},
	)

	time.Sleep(time.Second)

	// add a couple more random peers to the provider's cache
	pr2k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	pr2 := &hyperspace.Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: pr2k.PublicKey(),
		},
		PeerVector: hyperspace.New("foo", "bar"),
	}
	pr3k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	pr3 := &hyperspace.Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: pr3k.PublicKey(),
		},
		PeerVector: hyperspace.New("foo"),
	}
	prv.Put(pr2)
	prv.Put(pr3)

	time.Sleep(time.Second)

	// lookup "foo" as net1
	ctx := context.New(context.WithTimeout(time.Second))
	c0, err := net1.Dial(ctx, pr0.ConnectionInfo)
	require.NoError(t, err)

	err = c0.Write(
		context.New(),
		object.MustMarshal(
			&hyperspace.LookupRequest{
				Nonce:       "1",
				QueryVector: hyperspace.New("foo", "bar"),
			},
		),
	)
	require.NoError(t, err)

	// wait for response
	respObj := <-resp

	// check response
	res := &hyperspace.LookupResponse{}
	err = object.Unmarshal(respObj, res)
	require.NoError(t, err)
	assert.Equal(t, "1", res.Nonce)
	assert.ElementsMatch(t, []*hyperspace.Announcement{pr2}, res.Announcements)
}

func newPeer(t *testing.T) (net.Network, crypto.PrivateKey) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	n := net.New(k)
	lis, err := n.Listen(
		context.New(),
		"127.0.0.1:0",
		&net.ListenConfig{
			BindLocal: true,
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		lis.Close() // nolint: errcheck
	})

	return n, k
}
