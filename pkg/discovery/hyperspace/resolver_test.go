package hyperspace

import (
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
	"nimona.io/pkg/peer"
)

func TestDiscoverer_BootstrapLookup(t *testing.T) {
	_, k0, _, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	_, k1, _, x1, disc1, l1, ctx1 := newPeer(t, "peer1")

	d0, err := NewDiscoverer(ctx0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetAddresses()

	d1, err := NewDiscoverer(ctx1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctxR1 := context.New(context.WithCorrelationID("req1"))
	peers, err := d1.FindByFingerprint(ctxR1, k0.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l0.GetAddresses(), peers[0].Addresses)

	ctxR2 := context.New(context.WithCorrelationID("req2"))
	peers, err = d0.FindByFingerprint(ctxR2, k1.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetAddresses(), peers[0].Addresses)
}

func newPeer(
	t *testing.T,
	name string,
) (
	*crypto.PrivateKey,
	*crypto.PrivateKey,
	net.Network,
	exchange.Exchange,
	discovery.Discoverer,
	*peer.LocalPeer,
	context.Context,
) {
	ctx := context.New(context.WithCorrelationID(name))

	// owner key
	opk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	// peer key
	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	sig, err := crypto.NewSignature(
		opk,
		crypto.AlgorithmObjectHash,
		pk.ToObject(),
	)
	assert.NoError(t, err)

	pk.PublicKey.Signature = sig

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
