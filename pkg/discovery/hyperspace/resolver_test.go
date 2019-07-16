// +build flaky

package hyperspace

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/context"
	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
)

func TestDiscoverer_BootstrapLookup(t *testing.T) {
	_, k0, n0, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	_, k1, n1, x1, disc1, l1, ctx1 := newPeer(t, "peer1")

	d0, err := NewDiscoverer(ctx0, n0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetPeerInfo().Addresses

	d1, err := NewDiscoverer(ctx1, n1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctxR1 := context.New(context.WithCorrelationID("req1"))
	peers, err := d1.FindByFingerprint(ctxR1, k0.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l0.GetPeerInfo().Addresses, peers[0].Addresses)

	ctxR2 := context.New(context.WithCorrelationID("req2"))
	peers, err = d0.FindByFingerprint(ctxR2, k1.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetPeerInfo().Addresses, peers[0].Addresses)
}

func TestDiscoverer_FindBothSides(t *testing.T) {
	_, k0, n0, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	_, k1, n1, x1, disc1, l1, ctx1 := newPeer(t, "peer1")
	_, k2, n2, x2, disc2, l2, ctx2 := newPeer(t, "peer2")

	fmt.Println("k0", k0.Fingerprint())
	fmt.Println("k1", k1.Fingerprint())
	fmt.Println("k2", k2.Fingerprint())

	d0, err := NewDiscoverer(ctx0, n0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetPeerInfo().Addresses

	d1, err := NewDiscoverer(ctx1, n1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	d2, err := NewDiscoverer(ctx2, n2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctxR1 := context.New(context.WithCorrelationID("req1"))
	peers, err := d1.FindByFingerprint(ctxR1, k2.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l2.GetPeerInfo().Addresses, peers[0].Addresses)

	time.Sleep(time.Second)

	ctxR2 := context.New(context.WithCorrelationID("req2"))
	peers, err = d2.FindByFingerprint(ctxR2, k1.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetPeerInfo().Addresses, peers[0].Addresses)
}

func TestDiscoverer_FindBothSides_SubKeys(t *testing.T) {
	_, k0, n0, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	ok1, k1, n1, x1, disc1, l1, ctx1 := newPeer(t, "peer1")
	ok2, k2, n2, x2, disc2, l2, ctx2 := newPeer(t, "peer2")

	fmt.Println("k0", k0.Fingerprint())
	fmt.Println("k1", k1.Fingerprint())
	fmt.Println("k2", k2.Fingerprint())

	d0, err := NewDiscoverer(ctx0, n0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetPeerInfo().Addresses

	d1, err := NewDiscoverer(ctx1, n1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	d2, err := NewDiscoverer(ctx2, n2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctxR1 := context.New(context.WithCorrelationID("req1"))
	peers, err := d1.FindByFingerprint(ctxR1, ok2.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l2.GetPeerInfo().Addresses, peers[0].Addresses)

	time.Sleep(time.Second)

	ctxR2 := context.New(context.WithCorrelationID("req2"))
	peers, err = d2.FindByFingerprint(ctxR2, ok1.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetPeerInfo().Addresses, peers[0].Addresses)
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
	*peer.Peer,
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
	ds, err := graph.NewCayleyWithTempStore()
	local, err := peer.NewPeer("", pk)
	assert.NoError(t, err)

	n, err := net.New(disc, local)
	assert.NoError(t, err)

	tcp := net.NewTCPTransport(local, []string{})
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
		fmt.Sprintf("0.0.0.0:%d", 0),
	)
	assert.NoError(t, err)

	return opk, pk, n, x, disc, local, ctx
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}
