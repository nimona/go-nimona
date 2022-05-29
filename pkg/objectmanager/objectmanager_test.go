package objectmanager

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/resolvermock"
	"nimona.io/pkg/tilde"
)

func TestManager_Request(t *testing.T) {
	testPeerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	testDID := peer.IDFromPublicKey(testPeerKey.PublicKey())
	f00 := &object.Object{
		Type: "foo",
		Data: tilde.Map{
			"f00": tilde.String("f00"),
		},
	}
	type fields struct {
		store   func(*testing.T) objectstore.Store
		network func(*testing.T) network.Network
	}
	type args struct {
		ctx      context.Context
		rootHash tilde.Digest
		id       peer.ID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.Object
		wantErr bool
	}{{
		name: "should pass",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					SendCalls: []error{
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.MustMarshal(
									&object.Response{
										RequestID: "7",
										Object:    object.Copy(f00),
									},
								),
							}},
						},
					},
				}
				return m
			},
		},
		args: args{
			ctx:      context.Background(),
			rootHash: f00.Hash(),
			id:       testDID,
		},
		want: f00,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				objectstore: tt.fields.store(t),
				network:     tt.fields.network(t),
				newRequestID: func() string {
					return "7"
				},
			}
			got, err := m.Request(tt.args.ctx, tt.args.rootHash, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(
				t,
				tt.want,
				got,
			)
		})
	}
}

func TestManager_handleObjectRequest(t *testing.T) {
	peerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	peerID := peer.IDFromPublicKey(peerKey.PublicKey())

	peer1Key, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		Owner: peer.IDFromPublicKey(peer1Key.PublicKey()),
	}

	f00 := object.MustMarshal(peer1)
	f00m, err := f00.MarshalMap()
	require.NoError(t, err)
	f01 := &object.Object{
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"f01":  tilde.String("f01"),
			"asdf": f00m,
		},
	}

	type fields struct {
		storeHandler   func(*testing.T) objectstore.Store
		networkHandler func(
			*testing.T,
			context.Context,
			*sync.WaitGroup,
			*object.Object,
		) network.Network
		resolver func(*testing.T) resolver.Resolver
	}
	type args struct {
		ctx      context.Context
		rootHash tilde.Digest
		peer     *peer.ConnectionInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.Object
		wantErr bool
	}{
		{
			name: "object returned",
			fields: fields{
				storeHandler: func(t *testing.T) objectstore.Store {
					m := objectstoremock.NewMockStore(gomock.NewController(t))
					m.EXPECT().Get(f01.Hash()).Return(object.Copy(f01), nil).MaxTimes(2)
					return m
				},
				networkHandler: func(
					t *testing.T,
					ctx context.Context,
					wg *sync.WaitGroup,
					want *object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().GetPeerID().Return(peerID)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.MustMarshal(
									&object.Request{
										RequestID:  "8",
										ObjectHash: f01.Hash(),
									},
								),
							}},
						},
					)
					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj *object.Object,
							id peer.ID,
							opts ...network.SendOption,
						) error {
							assert.Equal(t, want, obj)
							wg.Done()
							return nil
						})
					return m
				},
				resolver: func(t *testing.T) resolver.Resolver {
					m := resolvermock.NewMockResolver(
						gomock.NewController(t),
					)
					return m
				},
			},
			args: args{
				ctx:      context.Background(),
				rootHash: f00.Hash(),
				peer:     peer1,
			},
			want: object.MustMarshal(
				&object.Response{
					Metadata: object.Metadata{
						Owner: peer.IDFromPublicKey(peerKey.PublicKey()),
					},
					Object:    f01,
					RequestID: "8",
				},
			),
		},
		{
			name: "object missing, return empty response",
			fields: fields{
				storeHandler: func(t *testing.T) objectstore.Store {
					m := objectstoremock.NewMockStore(gomock.NewController(t))
					m.EXPECT().Get(f01.Hash()).Return(nil, objectstore.ErrNotFound).MaxTimes(2)
					return m
				},
				networkHandler: func(
					t *testing.T,
					ctx context.Context,
					wg *sync.WaitGroup,
					want *object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().GetPeerID().Return(peerID)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.MustMarshal(
									&object.Request{
										RequestID:  "8",
										ObjectHash: f01.Hash(),
									},
								),
							}},
						},
					)
					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj *object.Object,
							id peer.ID,
							opts ...network.SendOption,
						) error {
							assert.Equal(t, want, obj)
							wg.Done()
							return nil
						})
					return m
				},
				resolver: func(t *testing.T) resolver.Resolver {
					m := resolvermock.NewMockResolver(
						gomock.NewController(t),
					)
					return m
				},
			},
			args: args{
				ctx:      context.Background(),
				rootHash: f00.Hash(),
				peer:     peer1,
			},
			want: object.MustMarshal(
				&object.Response{
					Metadata: object.Metadata{
						Owner: peer.IDFromPublicKey(peerKey.PublicKey()),
					},
					Object:    nil,
					RequestID: "8",
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup

			wg.Add(1)
			mgr := New(
				tt.args.ctx,
				tt.fields.networkHandler(
					t,
					tt.args.ctx,
					&wg,
					tt.want,
				),
				tt.fields.resolver(t),
				tt.fields.storeHandler(t),
			)
			assert.NotNil(t, mgr)
			wg.Wait()
		})
	}
}

func TestManager_Put(t *testing.T) {
	peerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	testOwnPublicKey := peerKey.PublicKey()
	require.NoError(t, err)
	testObjectSimple := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(testOwnPublicKey),
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	testObjectSimpleMap, err := testObjectSimple.MarshalMap()
	require.NoError(t, err)
	testObjectComplex := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"foo":           tilde.String("bar"),
			"nested-simple": testObjectSimpleMap,
		},
	}

	type fields struct {
		store                 func(*testing.T) objectstore.Store
		network               func(*testing.T) network.Network
		resolver              func(*testing.T) resolver.Resolver
		receivedSubscriptions []*object.Object
	}
	type args struct {
		o *object.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.Object
		wantErr bool
	}{{
		name: "should pass, simple object",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					Put(testObjectSimple).
					Return(nil)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnPeerKey: peerKey,
					SendCalls:     []error{},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{},
					},
				}
				return m
			},
			resolver: func(t *testing.T) resolver.Resolver {
				m := resolvermock.NewMockResolver(
					gomock.NewController(t),
				)
				return m
			},
		},
		args: args{
			o: object.Copy(testObjectSimple),
		},
		want: testObjectSimple,
	}, {
		name: "should pass, complex object",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					Put(testObjectComplex)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnPeerKey: peerKey,
					SendCalls:     []error{},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{},
					},
				}
				return m
			},
			resolver: func(t *testing.T) resolver.Resolver {
				m := resolvermock.NewMockResolver(
					gomock.NewController(t),
				)
				return m
			},
		},
		args: args{
			o: object.Copy(testObjectComplex),
		},
		want: testObjectComplex,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(
				context.Background(),
				tt.fields.network(t),
				tt.fields.resolver(t),
				tt.fields.store(t),
			)
			err := m.Put(
				context.Background(),
				tt.args.o,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_manager_Subscribe(t *testing.T) {
	p, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	o0 := &object.Object{
		Type: "foo",
	}
	o1 := &object.Object{
		Type: "not-bar",
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(p.PublicKey()),
		},
		Data: tilde.Map{
			"foo": tilde.String("not-bar"),
		},
	}
	o2 := &object.Object{
		Type: "bar",
		Metadata: object.Metadata{
			Root: o0.Hash(),
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	tests := []struct {
		name          string
		lookupOptions []LookupOption
		publish       []*object.Object
		want          []*object.Object
	}{{
		name: "subscribe by hash",
		lookupOptions: []LookupOption{
			FilterByHash(o2.Hash()),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}, {
		name: "subscribe by owner",
		lookupOptions: []LookupOption{
			FilterByOwner(peer.IDFromPublicKey(p.PublicKey())),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o1},
	}, {
		name: "subscribe by type",
		lookupOptions: []LookupOption{
			FilterByObjectType("bar"),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}, {
		name: "subscribe by stream",
		lookupOptions: []LookupOption{
			FilterByStreamHash(o0.Hash()),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}, {
		name: "subscribe by stream and owner",
		lookupOptions: []LookupOption{
			FilterByStreamHash(tilde.Digest("foo")),
			FilterByOwner(peer.IDFromPublicKey(p.PublicKey())),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{},
	}, {
		name: "subscribe by hash and type",
		lookupOptions: []LookupOption{
			FilterByHash(o2.Hash()),
			FilterByObjectType("bar"),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				pubsub:        NewObjectPubSub(),
				subscriptions: &SubscriptionsMap{},
			}
			sub := m.Subscribe(tt.lookupOptions...)
			for _, o := range tt.publish {
				m.pubsub.Publish(o)
			}
			time.Sleep(100 * time.Millisecond)
			sub.Close()
			os := []*object.Object{}
			for {
				o, err := sub.Read()
				if err != nil || o == nil {
					break
				}
				os = append(os, o)
			}
			assert.ElementsMatch(t, tt.want, os)
		})
	}
}
