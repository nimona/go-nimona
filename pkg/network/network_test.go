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

	l1, err := n1.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l1.Close()

	l2, err := n2.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l2.Close()

	testObj := &object.Object{
		Type: "foo",
		Data: map[string]interface{}{
			"foo:s": "bar",
		},
	}

	// subscribe to objects of type "foo" coming to n2
	sub := n2.Subscribe(
		FilterByObjectType("foo"),
	)
	require.NotNil(t, sub)

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

	// wait for event from n1 to arrive
	env, err := sub.Next()
	require.NoError(t, err)
	assert.Equal(t, testObj.ToMap(), env.Payload.ToMap())

	// subscribe to all objects coming to n1
	sub = n1.Subscribe()

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

	// next object should be our foo
	env, err = sub.Next()
	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, testObj.ToMap(), env.Payload.ToMap())
}

func TestNetwork_Relay(t *testing.T) {
	n0 := New(context.Background())
	n1 := New(context.Background())
	n2 := New(context.Background())

	l0, err := n0.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
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

	testObj := &object.Object{
		Type: "foo",
		Data: map[string]interface{}{
			"foo:s": "bar",
		},
	}

	testObjFromP1 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Data: map[string]interface{}{
			"foo:s": "bar",
		},
	}

	testObjFromP2 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Data: map[string]interface{}{
			"foo:s": "bar",
		},
	}

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
		testObjFromP1.Metadata.Signature,
		env.Payload.Metadata.Signature,
	)

	// send from p2 to p1
	err = n2.Send(context.Background(), testObjFromP2, p1)
	require.NoError(t, err)

	sub = n1.Subscribe(FilterByObjectType("foo"))
	env, err = sub.Next()
	require.NoError(t, err)

	require.NotNil(t, sub)
	assert.Equal(t,
		testObjFromP2.Metadata.Signature,
		env.Payload.Metadata.Signature,
	)
}

func Test_exchange_signAll(t *testing.T) {
	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	t.Run("should pass, sign root object", func(t *testing.T) {
		o := &object.Object{
			Type: "foo",
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
			},
			Data: map[string]interface{}{
				"foo:s": "bar",
			},
		}

		g, err := signAll(k, o)
		assert.NoError(t, err)

		assert.NotNil(t, g.Metadata.Signature)
		assert.False(t, g.Metadata.Signature.IsEmpty())
		assert.False(t, g.Metadata.Signature.Signer.IsEmpty())
	})

	t.Run("should pass, sign nested object", func(t *testing.T) {
		n := &object.Object{
			Type: "foo",
			Metadata: object.Metadata{
				Owner: k.PublicKey(),
			},
			Data: map[string]interface{}{
				"foo:s": "bar",
			},
		}
		o := &object.Object{
			Type: "foo",
			Data: map[string]interface{}{
				"foo:m": n,
			},
		}

		g, err := signAll(k, o)
		assert.NoError(t, err)

		assert.True(t, g.Metadata.Signature.IsEmpty())
		assert.True(t, g.Metadata.Signature.Signer.IsEmpty())

		gn := g.Data["foo:m"].(*object.Object)
		assert.False(t, gn.Metadata.Signature.IsEmpty())
		assert.False(t, gn.Metadata.Signature.Signer.IsEmpty())
	})
}
