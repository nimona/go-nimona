package network

import (
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/fixtures"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

func TestNetwork_SimpleConnection(t *testing.T) {
	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	n1 := New(context.Background(), k1)
	n2 := New(context.Background(), k2)

	l1, err := n1.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l1.Close()

	l2, err := n2.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l2.Close()

	testObj := &object.Object{
		Type: "foo",
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	// subscribe to objects of type "foo" coming to n2
	s2 := n2.Subscribe(
		FilterByObjectType("foo"),
	)
	require.NotNil(t, s2)

	// send from p1 to p2
	err = n1.Send(
		context.Background(),
		testObj,
		n2.GetPeerKey().PublicKey().DID(),
		SendWithConnectionInfo(
			&peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: n2.GetPeerKey().PublicKey().DID(),
				},
				Addresses: n2.GetAddresses(),
			},
		),
	)
	require.NoError(t, err)

	// wait for event from n1 to arrive
	env, err := s2.Next()
	require.NoError(t, err)
	assert.Equal(t, testObj, env.Payload)

	// subscribe to all objects coming to n1
	s1 := n1.Subscribe()

	// send from p2 to p1
	err = n2.Send(
		context.Background(),
		testObj,
		n1.GetPeerKey().PublicKey().DID(),
		SendWithConnectionInfo(
			&peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: n1.GetPeerKey().PublicKey().DID(),
				},
				Addresses: n1.GetAddresses(),
			},
		),
	)
	require.NoError(t, err)

	// next object should be our foo
	env, err = s1.Next()
	require.NoError(t, err)
	require.NotNil(t, s1)
	assert.Equal(t, testObj, env.Payload)

	t.Run("re-establish broken connections", func(t *testing.T) {
		// close p2's connection to p1
		c, err := n2.(*network).connmanager.Dial(
			context.New(),
			&peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: n1.GetPeerKey().PublicKey().DID(),
				},
			},
		)
		require.NoError(t, err)
		err = c.Close()
		require.NoError(t, err)
		// try to send something from p1 to p2
		err = n1.Send(
			context.Background(),
			testObj,
			n2.GetPeerKey().PublicKey().DID(),
			SendWithConnectionInfo(
				&peer.ConnectionInfo{
					Metadata: object.Metadata{
						Owner: n2.GetPeerKey().PublicKey().DID(),
					},
					Addresses: n2.GetAddresses(),
				},
			),
		)
		require.NoError(t, err)
		// try to send something from p2 to p1
		err = n2.Send(
			context.Background(),
			testObj,
			k1.PublicKey().DID(),
			SendWithConnectionInfo(
				&peer.ConnectionInfo{
					Metadata: object.Metadata{
						Owner: k1.PublicKey().DID(),
					},
					Addresses: n1.GetAddresses(),
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
			reqo, err := object.Marshal(req)
			require.NoError(t, err)
			err = n1.Send(
				context.Background(),
				reqo,
				n2.GetPeerKey().PublicKey().DID(),
				SendWithConnectionInfo(
					&peer.ConnectionInfo{
						Metadata: object.Metadata{
							Owner: n2.GetPeerKey().PublicKey().DID(),
						},
						Addresses: n2.GetAddresses(),
					},
				),
				SendWithResponse(gotRes, 0),
			)
			require.NoError(t, err)
			sendErr <- err
		}()
		// wait for p2 to get the req
		select {
		case gotReq := <-reqSub.Channel():
			v := string(gotReq.Payload.Data["requestID"].(tilde.String))
			assert.Equal(t, "1", v)
		case <-time.After(time.Second * 2):
			t.Fatal("timed out waiting for request")
		}
		// send response from p2 to p1
		reso, err := object.Marshal(res)
		require.NoError(t, err)
		require.NoError(t, err)
		// nolint: errcheck
		n2.Send(
			context.Background(),
			reso,
			n1.GetPeerKey().PublicKey().DID(),
			SendWithConnectionInfo(
				&peer.ConnectionInfo{
					Metadata: object.Metadata{
						Owner: n1.GetPeerKey().PublicKey().DID(),
					},
					Addresses: n1.GetAddresses(),
				},
			),
		)
		// check response
		select {
		case err := <-sendErr:
			require.NoError(t, err)
			assert.Equal(t, res, gotRes)
		case <-time.After(time.Second * 2):
			t.Fatal("timeout waiting for response")
		}
	})
}

func TestNetwork_Relay(t *testing.T) {
	k0, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	n0 := New(context.Background(), k0)
	time.Sleep(time.Millisecond * 500)
	n1 := New(context.Background(), k1)
	time.Sleep(time.Millisecond * 500)
	n2 := New(context.Background(), k2)
	time.Sleep(time.Millisecond * 500)

	l0, err := n0.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l0.Close()

	p0 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: n0.GetPeerKey().PublicKey().DID(),
		},
		Addresses: n0.GetAddresses(),
	}

	p1 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: n1.GetPeerKey().PublicKey().DID(),
		},
		Addresses: n1.GetAddresses(),
		Relays: []*peer.ConnectionInfo{
			p0,
		},
	}

	p2 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: n2.GetPeerKey().PublicKey().DID(),
		},
		Addresses: n2.GetAddresses(),
		Relays: []*peer.ConnectionInfo{
			p0,
		},
	}

	testObj := &object.Object{
		Type: "foo",
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	testObjFromP1 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: n1.GetPeerKey().PublicKey().DID(),
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	testObjFromP2 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: n2.GetPeerKey().PublicKey().DID(),
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	// send from p1 to p0
	err = n1.Send(
		context.Background(),
		testObj,
		p0.Metadata.Owner,
		SendWithConnectionInfo(p0),
	)
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	// send from p2 to p0
	err = n2.Send(
		context.Background(),
		testObj,
		p0.Metadata.Owner,
		SendWithConnectionInfo(p0),
	)
	require.NoError(t, err)
	time.Sleep(time.Millisecond * 500)

	// now we should be able to send from p1 to p2
	sub := n2.Subscribe(FilterByObjectType("foo"))
	err = n1.Send(
		context.Background(),
		testObjFromP1,
		p2.Metadata.Owner,
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

	time.Sleep(time.Millisecond * 500)

	// send from p2 to p1
	sub = n1.Subscribe(FilterByObjectType("foo"))

	err = n2.Send(
		context.Background(),
		testObjFromP2,
		p1.Metadata.Owner,
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

func Test_network_lookup(t *testing.T) {
	p0, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	fooConnInfo := &peer.ConnectionInfo{
		Version: 1,
		Metadata: object.Metadata{
			Owner: p0.PublicKey().DID(),
		},
		Addresses: []string{"a", "b"},
	}
	type fields struct {
		resolvers []Resolver
	}
	type args struct {
		lookupDID did.DID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*peer.ConnectionInfo
		wantErr bool
	}{{
		name: "one resolver, returns, should pass",
		fields: fields{
			resolvers: []Resolver{
				&testResolver{
					peers: map[string][]*peer.ConnectionInfo{
						p0.PublicKey().DID().String(): {
							fooConnInfo,
						},
					},
				},
			},
		},
		args: args{
			lookupDID: fooConnInfo.Metadata.Owner,
		},
		want: []*peer.ConnectionInfo{
			fooConnInfo,
		},
	}, {
		name: "two resolver, second returns, should pass",
		fields: fields{
			resolvers: []Resolver{
				&testResolver{
					peers: map[string][]*peer.ConnectionInfo{},
				},
				&testResolver{
					peers: map[string][]*peer.ConnectionInfo{
						p0.PublicKey().DID().String(): {
							fooConnInfo,
						},
					},
				},
			},
		},
		args: args{
			lookupDID: fooConnInfo.Metadata.Owner,
		},
		want: []*peer.ConnectionInfo{
			fooConnInfo,
		},
	}, {
		name: "two resolver, none returns, should fail",
		fields: fields{
			resolvers: []Resolver{
				&testResolver{
					peers: map[string][]*peer.ConnectionInfo{},
				},
				&testResolver{
					peers: map[string][]*peer.ConnectionInfo{},
				},
			},
		},
		args: args{
			lookupDID: fooConnInfo.Metadata.Owner,
		},
		want:    nil,
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := crypto.NewEd25519PrivateKey()
			require.NoError(t, err)
			w := New(
				context.Background(),
				k,
			).(*network)
			for _, r := range tt.fields.resolvers {
				w.RegisterResolver(r)
			}
			got, err := w.lookup(context.Background(), tt.args.lookupDID)
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

func BenchmarkNetworkSendToSinglePeer(b *testing.B) {
	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(b, err)
	n1 := New(context.Background(), k1).(*network)

	l1, err := n1.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(b, err)
	defer l1.Close()

	n1s := n1.Subscribe(FilterByObjectType("foo")).Channel()

	for n := 0; n < b.N; n++ {
		k2, err := crypto.NewEd25519PrivateKey()
		require.NoError(b, err)
		n2 := New(context.Background(), k2).(*network)
		err = n2.Send(
			context.Background(),
			&object.Object{
				Type: "foo",
				Data: tilde.Map{
					"foo": tilde.String("bar"),
				},
			},
			k1.PublicKey().DID(),
			SendWithConnectionInfo(
				&peer.ConnectionInfo{
					Metadata: object.Metadata{
						Owner: k1.PublicKey().DID(),
					},
					Addresses: n1.GetAddresses(),
				},
			),
		)
		require.NoError(b, err)
		select {
		case env := <-n1s:
			require.NotNil(b, env)
		case <-time.After(time.Second * 2):
			b.Fatal("timeout")
		}
		err = n2.Close()
		require.NoError(b, err)
	}
}

func TestNetwork_RequestRespond(t *testing.T) {
	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	s1 := objectstoremock.NewMockStore(gomock.NewController(t))
	s2 := objectstoremock.NewMockStore(gomock.NewController(t))

	n1 := New(context.Background(), k1, s1)
	n2 := New(context.Background(), k2, s2)

	l1, err := n1.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l1.Close()

	l2, err := n2.Listen(context.Background(), "127.0.0.1:0", ListenOnLocalIPs)
	require.NoError(t, err)
	defer l2.Close()

	testObj := &object.Object{
		Type: "foo",
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	s1.EXPECT().Get(testObj.Hash()).Return(testObj, nil)

	res := &object.Response{}
	err = n2.Send(
		context.New(),
		object.MustMarshal(&object.Request{
			Metadata:   object.Metadata{},
			RequestID:  "foo",
			ObjectHash: testObj.Hash(),
		}),
		n1.GetConnectionInfo().Metadata.Owner,
		SendWithConnectionInfo(n1.GetConnectionInfo()),
		SendWithResponse(res, time.Second),
	)
	require.NoError(t, err)
	require.Equal(t, testObj, res.Object)
	require.Equal(t, true, res.Found)
}

type testResolver struct {
	peers map[string][]*peer.ConnectionInfo
}

func (r *testResolver) LookupByDID(
	ctx context.Context,
	id did.DID,
) ([]*peer.ConnectionInfo, error) {
	c, ok := r.peers[id.String()]
	if !ok || c == nil {
		return nil, errors.Error("not found")
	}
	return c, nil
}
