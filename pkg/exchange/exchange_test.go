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
		Metadata: object.Metadata{
			Owner: kc1.GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n1.Addresses(),
	}
	p2 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: kc2.GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n2.Addresses(),
	}

	// create test objects from p2 to p1
	eo1 := object.Object{}.
		Set("body:s", "bar1").
		SetType("test/msg")

	// and from p1 to p2
	eo2 := object.Object{}.
		Set("body:s", "bar2").
		SetType("test/msg")

	wg := sync.WaitGroup{}
	wg.Add(2)

	handled := int32(0)

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
		x2.Subscribe(
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
		Metadata: object.Metadata{
			Owner: rkc.GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: rn.Addresses(),
	}
	p1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: kc1.GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n1.Addresses(),
		Relays: []*peer.Peer{
			pR,
		},
	}
	p2 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: kc2.GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n2.Addresses(),
		Relays: []*peer.Peer{
			pR,
		},
	}

	// init connection from peer1 to relay
	o1 := object.Object{}.
		SetOwner(p1.PublicKey()).
		SetType("foo").
		Set("foo:s", "bar")
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
	o2 := object.Object{}.
		SetOwner(p2.PublicKey()).
		SetType("foo").
		Set("foo:s", "bar")
	err = x2.Send(
		context.New(
			context.WithCorrelationID("pre1"),
			context.WithTimeout(time.Second),
		),
		o2,
		pR,
	)
	assert.NoError(t, err)

	// create the message from p2 to p1
	eo1 := object.Object{}.
		SetOwner(p2.PublicKey()).
		Set("body:s", "bar1").
		SetType("test/msg")

		// and from p1 to p2
	eo2 := object.Object{}.
		SetOwner(p1.PublicKey()).
		Set("body:s", "bar2").
		SetType("test/msg")

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1ObjectHandled := false
	w2ObjectHandled := false

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
		x2.Subscribe(
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

func Test_exchange_signAll(t *testing.T) {
	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	t.Run("should pass, sign root object", func(t *testing.T) {
		o := object.Object{}.
			SetType("foo").
			SetOwner(k.PublicKey()).
			Set("foo:s", "data")

		g, err := signAll(k, o)
		assert.NoError(t, err)

		assert.NotNil(t, g.GetSignature())
		assert.False(t, g.GetSignature().IsEmpty())
		assert.False(t, g.GetSignature().Signer.IsEmpty())
	})

	t.Run("should pass, sign nested object", func(t *testing.T) {
		n := object.Object{}.
			SetType("foo").
			SetOwner(k.PublicKey()).
			Set("foo:s", "data")
		o := object.Object{}.
			SetType("foo").
			Set("foo:m", n.Raw())

		g, err := signAll(k, o)
		assert.NoError(t, err)

		assert.True(t, g.GetSignature().IsEmpty())
		assert.True(t, g.GetSignature().Signer.IsEmpty())

		gn := object.FromMap(g.Get("foo:m").(map[string]interface{}))
		assert.False(t, gn.GetSignature().IsEmpty())
		assert.False(t, gn.GetSignature().Signer.IsEmpty())
	})
}
