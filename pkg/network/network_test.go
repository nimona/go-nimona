package network

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/fixtures"
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
		Data: object.Map{
			"foo": object.String("bar"),
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
		n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		SendWithConnectionInfo(
			&peer.ConnectionInfo{
				PublicKey: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
				Addresses: n2.LocalPeer().GetAddresses(),
			},
		),
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
		n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		SendWithConnectionInfo(
			&peer.ConnectionInfo{
				PublicKey: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
				Addresses: n1.LocalPeer().GetAddresses(),
			},
		),
	)
	require.NoError(t, err)

	// next object should be our foo
	env, err = sub.Next()
	require.NoError(t, err)
	require.NotNil(t, sub)
	assert.Equal(t, testObj.ToMap(), env.Payload.ToMap())

	t.Run("re-establish broken connections", func(t *testing.T) {
		// close p2's connection to p1
		c, err := n2.(*network).connmgr.GetConnection(context.
			New(),
			&peer.ConnectionInfo{
				PublicKey: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
			},
		)
		require.NoError(t, err)
		err = c.Close()
		require.NoError(t, err)
		// try to send something from p1 to p2
		err = n1.Send(
			context.Background(),
			testObj,
			n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
			SendWithConnectionInfo(
				&peer.ConnectionInfo{
					PublicKey: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
					Addresses: n2.LocalPeer().GetAddresses(),
				},
			),
		)
		require.NoError(t, err)
	})

	t.Run("wait for response", func(t *testing.T) {
		req := &fixtures.TestRequest{
			RequestID: "1",
			Foo:       "bar",
		}
		res := &fixtures.TestResponse{
			RequestID: "1",
			Foo:       "bar",
		}
		// sub for p2 based on rID
		gotRes := &fixtures.TestResponse{}
		reqSub := n2.Subscribe(
			FilterByRequestID("1"),
		)
		// send request from p1 to p2 in a go routine
		sendErr := make(chan error)
		go func() {
			sendErr <- n1.Send(
				context.Background(),
				req.ToObject(),
				n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
				SendWithResponse(gotRes, 0),
			)
		}()
		// wait for p2 to get the req
		gotReq := <-reqSub.Channel()
		assert.Equal(t, "1", string(gotReq.Payload.Data["requestID"].(object.String)))
		// send response from p2 to p1
		// nolint: errcheck
		n2.Send(
			context.Background(),
			res.ToObject(),
			n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
			SendWithConnectionInfo(
				&peer.ConnectionInfo{
					PublicKey: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
					Addresses: n1.LocalPeer().GetAddresses(),
				},
			),
		)
		// check response
		err = <-sendErr
		require.NoError(t, err)
		assert.Equal(t, res, gotRes)
	})
}

func TestNetwork_Relay(t *testing.T) {
	n0 := New(context.Background())
	n1 := New(context.Background())
	n2 := New(context.Background())

	l0, err := n0.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l0.Close()

	p0 := &peer.ConnectionInfo{
		PublicKey: n0.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		Addresses: n0.LocalPeer().GetAddresses(),
	}

	p1 := &peer.ConnectionInfo{
		PublicKey: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		Addresses: n1.LocalPeer().GetAddresses(),
		Relays: []*peer.ConnectionInfo{
			p0,
		},
	}

	p2 := &peer.ConnectionInfo{
		PublicKey: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		Addresses: n2.LocalPeer().GetAddresses(),
		Relays: []*peer.ConnectionInfo{
			p0,
		},
	}

	testObj := &object.Object{
		Type: "foo",
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}

	testObjFromP1 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: n1.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}

	testObjFromP2 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: n2.LocalPeer().GetPrimaryPeerKey().PublicKey(),
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}

	// send from p1 to p0
	err = n1.Send(
		context.Background(),
		testObj,
		p0.PublicKey,
		SendWithConnectionInfo(p0),
	)
	require.NoError(t, err)

	// send from p2 to p0
	err = n2.Send(
		context.Background(),
		testObj,
		p0.PublicKey,
		SendWithConnectionInfo(p0),
	)
	require.NoError(t, err)

	// now we should be able to send from p1 to p2
	sub := n2.Subscribe(FilterByObjectType("foo"))
	err = n1.Send(
		context.Background(),
		testObjFromP1,
		p2.PublicKey,
		SendWithConnectionInfo(p2),
	)
	require.NoError(t, err)

	env, err := sub.Next()
	require.NoError(t, err)

	require.NotNil(t, sub)
	assert.Equal(t,
		testObjFromP1.Metadata.Signature,
		env.Payload.Metadata.Signature,
	)

	// send from p2 to p1
	sub = n1.Subscribe(FilterByObjectType("foo"))

	err = n2.Send(
		context.Background(),
		testObjFromP2,
		p1.PublicKey,
		SendWithConnectionInfo(p1),
	)
	require.NoError(t, err)

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
			Data: object.Map{
				"foo": object.String("bar"),
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
			Data: object.Map{
				"foo": object.String("bar"),
			},
		}
		o := &object.Object{
			Type: "foo",
			Data: object.Map{
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

func Test_network_lookup(t *testing.T) {
	fooConnInfo := &peer.ConnectionInfo{
		Version:   1,
		PublicKey: "foo",
		Addresses: []string{"a", "b"},
	}
	type fields struct {
		resolvers []Resolver
	}
	type args struct {
		publicKey crypto.PublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *peer.ConnectionInfo
		wantErr bool
	}{{
		name: "one resolver, returns, should pass",
		fields: fields{
			resolvers: []Resolver{
				&testResolver{
					peers: map[crypto.PublicKey]*peer.ConnectionInfo{
						fooConnInfo.PublicKey: fooConnInfo,
					},
				},
			},
		},
		args: args{
			publicKey: fooConnInfo.PublicKey,
		},
		want: fooConnInfo,
	}, {
		name: "two resolver, second returns, should pass",
		fields: fields{
			resolvers: []Resolver{
				&testResolver{
					peers: map[crypto.PublicKey]*peer.ConnectionInfo{},
				},
				&testResolver{
					peers: map[crypto.PublicKey]*peer.ConnectionInfo{
						fooConnInfo.PublicKey: fooConnInfo,
					},
				},
			},
		},
		args: args{
			publicKey: fooConnInfo.PublicKey,
		},
		want: fooConnInfo,
	}, {
		name: "two resolver, none returns, should fail",
		fields: fields{
			resolvers: []Resolver{
				&testResolver{
					peers: map[crypto.PublicKey]*peer.ConnectionInfo{},
				},
				&testResolver{
					peers: map[crypto.PublicKey]*peer.ConnectionInfo{},
				},
			},
		},
		args: args{
			publicKey: fooConnInfo.PublicKey,
		},
		want:    nil,
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := New(
				context.Background(),
			).(*network)
			for _, r := range tt.fields.resolvers {
				w.RegisterResolver(r)
			}
			got, err := w.lookup(context.Background(), tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

type testResolver struct {
	peers map[crypto.PublicKey]*peer.ConnectionInfo
}

func (r *testResolver) LookupPeer(
	ctx context.Context,
	publicKey crypto.PublicKey,
) (*peer.ConnectionInfo, error) {
	c, ok := r.peers[publicKey]
	if !ok || c == nil {
		return nil, errors.New("not found")
	}
	return c, nil
}
