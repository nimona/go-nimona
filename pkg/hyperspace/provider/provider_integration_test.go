package provider

import (
	"fmt"
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
	"nimona.io/pkg/tilde"
)

func TestProvider_handleAnnouncement(t *testing.T) {
	// net0 is our provider
	net0, k0 := newPeer(context.New(context.WithCorrelationID("prv0")), t)
	pr0 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: k0.PublicKey().DID(),
		},
		Addresses: net0.Addresses(),
	}

	// net1 is a normal peer
	net1, k1 := newPeer(context.New(), t)
	pr1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k1.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: k1.PublicKey().DID(),
			},
			Addresses: net1.Addresses(),
		},
		Digests: []tilde.Digest{"foo", "bar"},
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
	net0, k0 := newPeer(context.New(context.WithCorrelationID("net0")), t)
	pr0 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: k0.PublicKey().DID(),
		},
		Addresses: net0.Addresses(),
	}

	// net1 is another provider
	net1, k1 := newPeer(context.New(context.WithCorrelationID("net1")), t)
	pr1 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: k1.PublicKey().DID(),
		},
		Addresses: net1.Addresses(),
	}

	// net2 is a normal peer
	net2, k2 := newPeer(context.New(context.WithCorrelationID("net2")), t)
	pr2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k2.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: k2.PublicKey().DID(),
			},
			Addresses: net2.Addresses(),
		},
		Digests:          []tilde.Digest{"foo", "bar"},
		PeerCapabilities: []string{"foo", "bar"},
	}

	time.Sleep(time.Second * 2)

	// construct providers
	prv0, err := New(
		context.New(
			context.WithCorrelationID("prv0"),
		),
		net0,
		k0,
		[]*peer.ConnectionInfo{pr1},
	)
	require.NoError(t, err)
	prv1, err := New(
		context.New(
			context.WithCorrelationID("prv1"),
		),
		net1,
		k1,
		[]*peer.ConnectionInfo{pr0},
	)
	require.NoError(t, err)

	time.Sleep(time.Second * 2)

	t.Run("net2 announces to prv0, should propagate", func(t *testing.T) {
		// net2 announces to provider 0
		c2, err := net2.Dial(
			context.New(
				context.WithCorrelationID("net2dial"),
			),
			pr0,
		)
		require.NoError(t, err)
		time.Sleep(time.Second)

		err = c2.Write(
			context.New(
				context.WithCorrelationID("net2dial/write"),
			),
			object.MustMarshal(pr2),
		)
		require.NoError(t, err)
		time.Sleep(time.Second)

		// wait a bit and check if both providers have cached the peer
		var existsInPrv1 error
		for i := 0; i < 10; i++ {
			_, existsInPrv1 = prv0.peerCache.Get(
				pr2.ConnectionInfo.Metadata.Owner,
			)
			if existsInPrv1 != nil {
				time.Sleep(time.Second)
				fmt.Println("existsInPrv1 failed", i+1, "times")
				continue
			}
		}
		assert.NoError(t, existsInPrv1)
		var existsInPrv2 error
		for i := 0; i < 10; i++ {
			_, existsInPrv2 = prv1.peerCache.Get(
				pr2.ConnectionInfo.Metadata.Owner,
			)
			if existsInPrv2 != nil {
				time.Sleep(time.Second)
				fmt.Println("existsInPrv2 failed", i+1, "times")
				continue
			}
		}
		assert.NoError(t, existsInPrv2)
	})
}

func TestProvider_handlePeerLookup(t *testing.T) {
	// net0 is our provider
	net0, k0 := newPeer(context.New(context.WithCorrelationID("prv0")), t)
	pr0 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: k0.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: k0.PublicKey().DID(),
			},
			Addresses: net0.Addresses(),
		},
	}

	// net1 is a normal peer
	net1, k1 := newPeer(context.New(), t)

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

	// add a couple more random peers to the provider's cache
	pr2k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	pr2 := &hyperspace.Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: pr2k.PublicKey().DID(),
			},
		},
		Digests: []tilde.Digest{"foo", "bar"},
	}
	pr3k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	pr3 := &hyperspace.Announcement{
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: pr3k.PublicKey().DID(),
			},
		},
		Digests: []tilde.Digest{"not-foo"},
	}
	prv.Put(pr2)
	prv.Put(pr3)

	// lookup "foo" as net1
	ctx := context.New(context.WithTimeout(time.Second))
	c0, err := net1.Dial(ctx, pr0.ConnectionInfo)
	require.NoError(t, err)

	err = c0.Write(
		context.New(),
		object.MustMarshal(
			&hyperspace.LookupByDigestRequest{
				Nonce:  "1",
				Digest: tilde.Digest("foo"),
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

func newPeer(ctx context.Context, t *testing.T) (
	net.Network,
	crypto.PrivateKey,
) {
	k, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

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

	return n, k
}
