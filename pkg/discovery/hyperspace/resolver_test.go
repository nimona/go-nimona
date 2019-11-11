package hyperspace

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/store/graph"
	"nimona.io/internal/store/kv"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestDiscoverer_TwoPeersCanFindEachOther(t *testing.T) {
	_, k0, _, x0, disc0, l0, ctx0 := newPeer(t, "peer0")

	d0, err := NewDiscoverer(ctx0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetAddresses()

	time.Sleep(time.Second)

	_, k1, _, x1, disc1, l1, ctx1 := newPeer(t, "peer1")

	d1, err := NewDiscoverer(ctx1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctx := context.New(
		context.WithCorrelationID("req1"),
		context.WithTimeout(time.Second),
	)
	peers, err := d1.FindByPublicKey(ctx, k0.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l0.GetAddresses(), peers[0].Addresses)

	ctxR2 := context.New(
		context.WithCorrelationID("req2"),
		context.WithTimeout(time.Second),
	)
	peers, err = d0.FindByPublicKey(ctxR2, k1.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetAddresses(), peers[0].Addresses)
}

func TestDiscoverer_TwoPeersAndOneBootstrapCanFindEachOther(t *testing.T) {
	_, k0, _, x0, disc0, l0, ctx0 := newPeer(t, "peer0")

	// bootstrap node
	d0, err := NewDiscoverer(ctx0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	_, k1, _, x1, disc1, l1, ctx1 := newPeer(t, "peer1")
	_, k2, _, x2, disc2, l2, ctx2 := newPeer(t, "peer2")

	fmt.Println("k0", k0)
	fmt.Println("k1", k1)
	fmt.Println("k2", k2)

	// bootstrap address
	ba := l0.GetAddresses()

	// node 1
	d1, err := NewDiscoverer(ctx1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	// node 2
	d2, err := NewDiscoverer(ctx2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

	// wait for everything to settle
	time.Sleep(time.Second * 5)

	// find bootstrap from node1
	ctx := context.New(
		context.WithCorrelationID("req1"),
		context.WithTimeout(time.Second*2),
	)
	peers, err := d1.FindByPublicKey(ctx, k0.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l0.GetAddresses(), peers[0].Addresses)

	// find node 1 from node 2
	ctx = context.New(
		context.WithCorrelationID("req2"),
		context.WithTimeout(time.Second*2),
	)
	peers, err = d2.FindByPublicKey(ctx, k1.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetAddresses(), peers[0].Addresses)

	// find node 2 from node 1
	ctx = context.New(
		context.WithCorrelationID("req3"),
		context.WithTimeout(time.Second*2),
	)
	peers, err = d1.FindByPublicKey(ctx, k2.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l2.GetAddresses(), peers[0].Addresses)

	// add extra peer
	_, k3, _, x3, disc3, l3, ctx3 := newPeer(t, "peer3")

	// setup node 3
	d3, err := NewDiscoverer(ctx3, x3, l3, ba)
	assert.NoError(t, err)
	err = disc3.AddProvider(d3)
	assert.NoError(t, err)
	assert.NotNil(t, d3)

	// wait for everything to settle
	time.Sleep(time.Second * 5)

	fmt.Println("peer0", k0)
	fmt.Println("peer1", k1)
	fmt.Println("peer2", k2)
	fmt.Println("peer3", k3)

	fmt.Println("-------------------")
	fmt.Println("-------------------")
	fmt.Println("-------------------")
	fmt.Println("-------------------")

	// find node 3 from node 1
	ctx = context.New(
		context.WithCorrelationID("req4"),
		context.WithTimeout(time.Second*2),
	)
	peers, err = d1.FindByPublicKey(ctx, k3.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l3.GetAddresses(), peers[0].Addresses)

	// find node 3 from node 2
	ctx = context.New(
		context.WithCorrelationID("req5"),
		context.WithTimeout(time.Second*2),
	)
	peers, err = d2.FindByPublicKey(ctx, k3.PublicKey())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l3.GetAddresses(), peers[0].Addresses)
}

func TestDiscoverer_TwoPeersAndOneBootstrapCanProvideForEachOther(t *testing.T) {
	_, k0, _, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	_, k1, _, x1, disc1, l1, ctx1 := newPeer(t, "peer1")
	_, k2, _, x2, disc2, l2, ctx2 := newPeer(t, "peer2")

	// make peer 1 a provider
	token := make([]byte, 32)
	rand.Read(token) // nolint: errcheck
	ch := &object.Hash{
		Algorithm: "OH1",
		D:         token,
	}
	l1.AddContentHash(ch)

	// print peer info
	fmt.Println("k0", k0)
	fmt.Println("k1", k1)
	fmt.Println("k2", k2)

	// bootstrap peer
	d0, err := NewDiscoverer(ctx0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	// bootstrap address
	ba := l0.GetAddresses()

	// peer 1
	d1, err := NewDiscoverer(ctx1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	// peer 2
	d2, err := NewDiscoverer(ctx2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

	// wait for everything to settle
	time.Sleep(time.Second * 5)

	// find peer 1 from peer 2
	ctx := context.New(
		context.WithCorrelationID("req1"),
		context.WithTimeout(time.Second),
	)
	providers, err := d2.FindByContent(ctx, ch)
	require.NoError(t, err)
	require.Len(t, providers, 1)
	require.Equal(t, k1.PublicKey(), providers[0])

	// find peer 1 from bootstrap
	ctx = context.New(
		context.WithCorrelationID("req2"),
		context.WithTimeout(time.Second*2),
	)
	providers, err = d0.FindByContent(ctx, ch)
	require.NoError(t, err)
	require.Len(t, providers, 1)
	require.Equal(t, k1.PublicKey(), providers[0])

	// add extra peer
	_, _, _, x3, disc3, l3, ctx3 := newPeer(t, "peer3")

	// setup peer 3
	d3, err := NewDiscoverer(ctx3, x3, l3, ba)
	assert.NoError(t, err)
	err = disc3.AddProvider(d1)
	assert.NoError(t, err)

	// wait for everything to settle
	time.Sleep(time.Second * 5)

	// find peer 1 from peer 3
	ctx = context.New(
		context.WithCorrelationID("req3"),
		context.WithTimeout(time.Second),
	)
	providers, err = d3.FindByContent(ctx, ch)
	require.NoError(t, err)
	require.Len(t, providers, 1)
	require.Equal(t, k1.PublicKey(), providers[0])
}

func newPeer(
	t *testing.T,
	name string,
) (
	crypto.PrivateKey,
	crypto.PrivateKey,
	net.Network,
	exchange.Exchange,
	discovery.Discoverer,
	*peer.LocalPeer,
	context.Context,
) {
	ctx := context.New(context.WithCorrelationID(name))

	// owner key
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	// peer key
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	// sig, err := crypto.NewSignature(
	// 	opk,
	// 	pk.PublicKey().ToObject(),
	// )
	// assert.NoError(t, err)

	// pk.PublicKey.Signature = sig

	disc := discovery.NewDiscoverer()
	ds := graph.New(kv.NewMemory())
	local, err := peer.NewLocalPeer("", pk)
	assert.NoError(t, err)

	n, err := net.New(disc, local)
	assert.NoError(t, err)

	tcp := net.NewTCPTransport(local, "127.0.0.1:0")
	hsm := handshake.New(local, disc)
	n.AddMiddleware(hsm.Handle())
	n.AddTransport("tcps", tcp)

	x, err := exchange.New(
		ctx,
		pk,
		n,
		ds,
		disc,
		local,
	)
	assert.NoError(t, err)

	return opk, pk, n, x, disc, local, ctx
}
