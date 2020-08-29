package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestNetwork_SimpleConnection(t *testing.T) {
	n1 := New(context.Background())
	n2 := New(context.Background())

	l1, err := n1.Listen(context.Background(), "0.0.0.0:0")
	require.NoError(t, err)
	defer l1.Close()

	l2, err := n2.Listen(context.Background(), "0.0.0.0:0")
	require.NoError(t, err)
	defer l2.Close()

	testObj := new(object.Object).
		SetType("foo").
		Set("foo:s", object.String("bar"))

	// send from p1 to p2
	err = n1.Send(
		context.Background(),
		testObj,
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
			},
			Addresses: n2.LocalPeer().GetAddresses(),
		},
	)
	require.NoError(t, err)

	sub := n2.Subscribe(
		FilterByObjectType("foo"),
	)
	env, err := sub.Next()
	require.NoError(t, err)

	require.NotNil(t, sub)
	assert.Equal(t, testObj.ToMap(), env.Payload.ToMap())

	// send from p2 to p1
	err = n2.Send(
		context.Background(),
		testObj,
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
			},
			Addresses: n1.LocalPeer().GetAddresses(),
		},
	)
	require.NoError(t, err)

	sub = n1.Subscribe()
	env, err = sub.Next()
	require.NoError(t, err)

	require.NotNil(t, sub)
	assert.Equal(t, testObj.ToMap(), env.Payload.ToMap())
}

func TestNetwork_Relay(t *testing.T) {
	n0 := New(context.Background())
	n1 := New(context.Background())
	n2 := New(context.Background())

	l0, err := n0.Listen(context.Background(), "0.0.0.0:0")
	require.NoError(t, err)
	defer l0.Close()

	p0 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: n0.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n0.LocalPeer().GetAddresses(),
	}

	p1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n1.LocalPeer().GetAddresses(),
		Relays: []*peer.Peer{
			p0,
		},
	}

	p2 := &peer.Peer{

		Metadata: object.Metadata{
			Owner: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Addresses: n2.LocalPeer().GetAddresses(),
		Relays: []*peer.Peer{
			p0,
		},
	}

	testObj := new(object.Object).
		SetType("foo").
		Set("foo:s", object.String("bar"))

	testObjFromP1 := testObj.
		SetOwner(n1.LocalPeer().GetPrimaryPeerKey().PublicKey())
	testObjFromP2 := testObj.
		SetOwner(n2.LocalPeer().GetPrimaryPeerKey().PublicKey())

	// send from p1 to p0
	err = n1.Send(context.Background(), testObj, p0)
	require.NoError(t, err)

	// send from p2 to p0
	err = n2.Send(context.Background(), testObj, p0)
	require.NoError(t, err)

	// now we should be able to send from p1 to p2
	err = n1.Send(context.Background(), testObjFromP1, p2)
	require.NoError(t, err)

	sub := n2.Subscribe(FilterByObjectType("foo"))
	env, err := sub.Next()
	require.NoError(t, err)

	require.NotNil(t, sub)
	assert.Equal(t,
		testObjFromP1.SetSignature(object.Signature{}).ToMap(),
		env.Payload.SetSignature(object.Signature{}).ToMap(),
	)

	// send from p2 to p1
	err = n2.Send(context.Background(), testObjFromP2, p1)
	require.NoError(t, err)

	sub = n1.Subscribe(FilterByObjectType("foo"))
	env, err = sub.Next()
	require.NoError(t, err)

	require.NotNil(t, sub)
	assert.Equal(t,
		testObjFromP2.SetSignature(object.Signature{}).ToMap(),
		env.Payload.SetSignature(object.Signature{}).ToMap(),
	)
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
