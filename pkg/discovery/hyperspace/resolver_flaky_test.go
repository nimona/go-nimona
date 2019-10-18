package hyperspace

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
)

func TestDiscoverer_FindBothSides_SubKeys(t *testing.T) {
	ok0, k0, _, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	ok1, k1, _, x1, disc1, l1, ctx1 := newPeer(t, "peer1")
	ok2, k2, _, x2, disc2, l2, ctx2 := newPeer(t, "peer2")

	fmt.Println("ok0", ok0.Fingerprint(), "k0", k0.Fingerprint())
	fmt.Println("ok1", ok1.Fingerprint(), "k1", k1.Fingerprint())
	fmt.Println("ok2", ok2.Fingerprint(), "k2", k2.Fingerprint())

	d0, err := NewDiscoverer(ctx0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetAddresses()

	d1, err := NewDiscoverer(ctx1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	d2, err := NewDiscoverer(ctx2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 50)
	ctxR1 := context.New(context.WithCorrelationID("req1"))
	peers, err := d1.FindByFingerprint(ctxR1, ok2.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l2.GetAddresses(), peers[0].Addresses)

	ctxR2 := context.New(context.WithCorrelationID("req2"))
	peers, err = d2.FindByFingerprint(ctxR2, ok1.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetAddresses(), peers[0].Addresses)
}

func TestDiscoverer_FindBothSides(t *testing.T) {
	_, k0, _, x0, disc0, l0, ctx0 := newPeer(t, "peer0")
	_, k1, _, x1, disc1, l1, ctx1 := newPeer(t, "peer1")
	_, k2, _, x2, disc2, l2, ctx2 := newPeer(t, "peer2")

	fmt.Println("k0", k0.Fingerprint())
	fmt.Println("k1", k1.Fingerprint())
	fmt.Println("k2", k2.Fingerprint())

	d0, err := NewDiscoverer(ctx0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetAddresses()

	d1, err := NewDiscoverer(ctx1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	d2, err := NewDiscoverer(ctx2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 50)

	ctxR1 := context.New(context.WithCorrelationID("req1"))
	peers, err := d1.FindByFingerprint(ctxR1, k2.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l2.GetAddresses(), peers[0].Addresses)

	ctxR2 := context.New(context.WithCorrelationID("req2"))
	peers, err = d2.FindByFingerprint(ctxR2, k1.Fingerprint())
	require.NoError(t, err)
	require.Len(t, peers, 1)
	require.Equal(t, l1.GetAddresses(), peers[0].Addresses)
}
