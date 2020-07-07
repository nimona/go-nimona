package exchange

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestSendSuccess(t *testing.T) {
	// enable binding to local addresses
	net.BindLocal = true

	// create new peers
	kc1, n1, x1 := newPeer(t)
	kc2, n2, x2 := newPeer(t)

	// make up the peers
	p1 := &peer.Peer{
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
	}
	p2 := &peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
	}

	// create test objects
	eo1 := object.Object{}
	eo1 = eo1.Set("body:s", "bar1")
	eo1 = eo1.SetType("test/msg")

	eo2 := object.Object{}
	eo2 = eo2.Set("body:s", "bar1")
	eo2 = eo2.SetType("test/msg")

	wg := sync.WaitGroup{}
	wg.Add(2)

	handled := int32(0)

	sig, err := object.NewSignature(kc2.GetPrimaryPeerKey(), eo1)
	assert.NoError(t, err)
	eo1 = eo1.AddSignature(sig)

	sig, err = object.NewSignature(kc1.GetPrimaryPeerKey(), eo2)
	assert.NoError(t, err)
	eo2 = eo2.AddSignature(sig)

	// add message handlers
	// nolint: dupl
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("test/msg"),
		),
		func(e *Envelope) error {
			o := e.Payload
			assert.Equal(t, eo1.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	// nolint: dupl
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("tes**"),
		),
		func(e *Envelope) error {
			o := e.Payload
			assert.Equal(t, eo2.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	ctx := context.Background()

	errS1 := x2.Send(
		ctx,
		eo1,
		p1,
	)
	assert.NoError(t, errS1)

	errS1b := x2.Send(
		ctx,
		eo1,
		p1,
	)
	assert.Equal(t, ErrAlreadySentDuringContext, errS1b)

	time.Sleep(time.Second)

	errS2 := x1.Send(
		ctx,
		eo2,
		p2,
	)
	assert.NoError(t, errS2)

	if errS1 == nil && errS2 == nil {
		wg.Wait()
	}

	assert.Equal(t, int32(2), atomic.LoadInt32(&handled))
}

func TestSendRelay(t *testing.T) {
	// enable binding to local addresses
	net.BindLocal = true

	// relay peer
	rkc, rn, _ := newPeer(t)

	// disable binding to local addresses
	net.BindLocal = false
	kc1, n1, x1 := newPeer(t)
	kc2, n2, x2 := newPeer(t)

	// make up the peers
	pR := &peer.Peer{
		Owners:    rkc.ListPublicKeys(keychain.PeerKey),
		Addresses: rn.Addresses(),
	}
	p1 := &peer.Peer{
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
		Relays: []*peer.Peer{
			pR,
		},
	}
	p2 := &peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
		Relays: []*peer.Peer{
			pR,
		},
	}

	// init connection from peer1 to relay
	o1 := object.Object{}
	o1 = o1.SetType("foo")
	o1 = o1.Set("foo:s", "bar")
	err := x1.Send(
		context.New(
			context.WithCorrelationID("pre0"),
			context.WithTimeout(time.Second),
		),
		o1,
		pR,
	)
	assert.NoError(t, err)

	// init connection from peer2 to relay
	o2 := object.Object{}
	o2 = o2.SetType("foo")
	o2 = o2.Set("foo:s", "bar")
	err = x2.Send(
		context.New(
			context.WithCorrelationID("pre1"),
			context.WithTimeout(time.Second),
		),
		o2,
		pR,
	)
	assert.NoError(t, err)

	// create the messages
	eo1 := object.Object{}
	eo1 = eo1.Set("body:s", "bar1")
	eo1 = eo1.SetType("test/msg")

	eo2 := object.Object{}
	eo2 = eo2.Set("body:s", "bar1")
	eo2 = eo2.SetType("test/msg")

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1ObjectHandled := false
	w2ObjectHandled := false

	sig, err := object.NewSignature(kc2.GetPrimaryPeerKey(), eo1)
	assert.NoError(t, err)
	eo1 = eo1.AddSignature(sig)

	handled := int32(0)

	// add handlers
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("test/msg"),
		),
		func(e *Envelope) error {
			o := e.Payload
			w1ObjectHandled = true
			assert.Equal(t, eo1.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	// nolint: dupl
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("tes**"),
		),
		func(e *Envelope) error {
			o := e.Payload
			w2ObjectHandled = true
			assert.Equal(t, eo2.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	err = x2.Send(
		context.New(
			context.WithCorrelationID("req0"),
			context.WithTimeout(time.Second),
		),
		eo1,
		p1,
	)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	err = x1.Send(
		context.New(
			context.WithCorrelationID("req1"),
			context.WithTimeout(time.Second),
		),
		eo2,
		p2,
	)
	assert.NoError(t, err)

	wg.Wait()

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func newPeer(
	t *testing.T,
) (
	keychain.Keychain,
	net.Network,
	*exchange,
) {
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	eb := eventbus.New()

	kc := keychain.New()
	kc.Put(keychain.PrimaryPeerKey, pk)

	ctx := context.Background()

	n := net.New(
		net.WithKeychain(kc),
	)
	_, err = n.Listen(ctx, "127.0.0.1:0")
	require.NoError(t, err)

	x := New(
		ctx,
		WithNet(n),
		WithKeychain(kc),
		WithEventbus(eb),
	)

	return kc, n, x.(*exchange)
}
