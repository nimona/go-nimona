package objectmanager

import (
	"database/sql"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/fixtures"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/hyperspace/resolvermock"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
	"nimona.io/pkg/tilde"
)

func TestManager_Request(t *testing.T) {
	testPeerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.ConnectionInfo{
		PublicKey: testPeerKey.PublicKey(),
	}
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
		rootHash tilde.Hash
		peer     *peer.ConnectionInfo
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
			peer:     testPeer,
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
			got, err := m.Request(tt.args.ctx, tt.args.rootHash, tt.args.peer)
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

	peer1Key, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		PublicKey: peer1Key.PublicKey(),
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
		rootHash tilde.Hash
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
					m.EXPECT().GetPeerKey().Return(peerKey)
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
							recipient crypto.PublicKey,
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
						Owner: peerKey.PublicKey().DID(),
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
					m.EXPECT().GetPeerKey().Return(peerKey)
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
							recipient crypto.PublicKey,
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
						Owner: peerKey.PublicKey().DID(),
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

func TestManager_RequestStream(t *testing.T) {
	testPeerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.ConnectionInfo{
		PublicKey: testPeerKey.PublicKey(),
	}
	f00 := &object.Object{
		Type:     "foo",
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"f00": tilde.String("f00"),
		},
	}
	f01 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Root: f00.Hash(),
			Parents: object.Parents{
				"*": []tilde.Hash{f00.Hash()},
			},
		},
		Data: tilde.Map{
			"f01": tilde.String("f01"),
		},
	}
	f02 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Root: f00.Hash(),
			Parents: object.Parents{
				"*": []tilde.Hash{f01.Hash()},
			},
		},
		Data: tilde.Map{
			"f02": tilde.String("f02"),
		},
	}

	type fields struct {
		store   func(*testing.T) objectstore.Store
		network func(*testing.T) network.Network
	}
	type args struct {
		ctx      context.Context
		rootHash tilde.Hash
		peer     *peer.ConnectionInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*object.Object
		wantErr bool
	}{{
		name: "should pass, all missing",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				m.EXPECT().
					Get(f00.Hash()).
					Return(nil, objectstore.ErrNotFound)
				m.EXPECT().
					Get(f01.Hash()).
					Return(nil, objectstore.ErrNotFound)
				m.EXPECT().
					Get(f02.Hash()).
					Return(nil, objectstore.ErrNotFound)
				m.EXPECT().
					Put(f00).
					Return(nil)
				m.EXPECT().
					Put(f01).
					Return(nil)
				m.EXPECT().
					Put(f02).
					Return(nil)
				m.EXPECT().
					GetByStream(f00.Hash()).
					Return(object.NewReadCloserFromObjects([]*object.Object{
						object.Copy(f00),
						object.Copy(f01),
						object.Copy(f02),
					}), err)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					SendCalls: []error{
						nil,
						nil,
						nil,
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.MustMarshal(
									&stream.Response{
										RequestID: "7",
										Leaves: []tilde.Hash{
											f02.Hash(),
										},
									},
								),
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.MustMarshal(
									&object.Response{
										RequestID: "7",
										Object:    object.Copy(f02),
									},
								),
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.MustMarshal(
									&object.Response{
										RequestID: "7",
										Object:    object.Copy(f01),
									},
								),
							}},
						},
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
			peer:     testPeer,
		},
		want: []*object.Object{
			f00,
			f01,
			f02,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				pubsub:      NewObjectPubSub(),
				objectstore: tt.fields.store(t),
				network:     tt.fields.network(t),
				newRequestID: func() string {
					return "7"
				},
			}
			got, err := m.RequestStream(tt.args.ctx, tt.args.rootHash, tt.args.peer)
			if (err != nil) != tt.wantErr {
				t.Errorf("manager.RequestStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				objs := []*object.Object{}
				for {
					obj, err := got.Read()
					if err == object.ErrReaderDone {
						break
					}
					if err != nil {
						break
					}
					objs = append(objs, obj)
				}
				require.Equal(t, len(tt.want), len(objs))
				for i := 0; i < len(tt.want); i++ {
					assert.Equal(
						t,
						tt.want[i],
						objs[i],
						"for index %d", i,
					)
				}
			}
		})
	}
}

func TestManager_handleStreamRequest(t *testing.T) {
	peerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	peer1Key, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		PublicKey: peer1Key.PublicKey(),
	}

	f00 := &object.Object{
		Type:     "foo",
		Metadata: object.Metadata{},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}

	f01 := &object.Object{
		Type: "foo-child",
		Metadata: object.Metadata{
			Root: f00.Hash(),
			Parents: object.Parents{
				"*": []tilde.Hash{
					f00.Hash(),
				},
			},
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
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
		rootHash tilde.Hash
		peer     *peer.ConnectionInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.Object
		wantErr bool
	}{{
		name: "should pass, stream found",
		fields: fields{
			storeHandler: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				m.EXPECT().
					GetStreamLeaves(f00.Hash()).
					Return(
						[]tilde.Hash{
							f01.Hash(),
						},
						nil,
					)
				return m
			},
			networkHandler: func(
				t *testing.T,
				ctx context.Context,
				wg *sync.WaitGroup,
				want *object.Object,
			) network.Network {
				m := networkmock.NewMockNetwork(gomock.NewController(t))
				m.EXPECT().GetPeerKey().Return(peerKey)
				m.EXPECT().Subscribe(gomock.Any()).Return(
					&networkmock.MockSubscriptionSimple{
						Objects: []*network.Envelope{{
							Payload: object.MustMarshal(
								&stream.Request{
									RequestID: "7",
									RootHash:  f00.Hash(),
								},
							),
						}},
					},
				)
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
						recipient crypto.PublicKey,
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
			&stream.Response{
				Metadata: object.Metadata{
					Owner: peerKey.PublicKey().DID(),
				},
				RequestID: "7",
				RootHash:  f00.Hash(),
				Leaves:    []tilde.Hash{f01.Hash()},
			},
		),
	}, {
		name: "should pass, unknown stream",
		fields: fields{
			storeHandler: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				m.EXPECT().
					GetStreamLeaves(f00.Hash()).
					Return(nil, objectstore.ErrNotFound)
				return m
			},
			networkHandler: func(
				t *testing.T,
				ctx context.Context,
				wg *sync.WaitGroup,
				want *object.Object,
			) network.Network {
				m := networkmock.NewMockNetwork(gomock.NewController(t))
				m.EXPECT().GetPeerKey().Return(peerKey)
				m.EXPECT().Subscribe(gomock.Any()).Return(
					&networkmock.MockSubscriptionSimple{
						Objects: []*network.Envelope{{
							Payload: object.MustMarshal(
								&stream.Request{
									RequestID: "7",
									RootHash:  f00.Hash(),
								},
							),
						}},
					},
				)
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
						recipient crypto.PublicKey,
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
			&stream.Response{
				Metadata: object.Metadata{
					Owner: peerKey.PublicKey().DID(),
				},
				RequestID: "7",
				RootHash:  f00.Hash(),
				Leaves:    []tilde.Hash{},
			},
		),
	}}
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
			Owner: testOwnPublicKey.DID(),
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
			for _, obj := range tt.fields.receivedSubscriptions {
				err := m.(*manager).handleStreamSubscription(
					context.Background(),
					&network.Envelope{
						Payload: obj,
					},
				)
				require.NoError(t, err)
			}
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

func TestManager_Append(t *testing.T) {
	peerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	testOwnPublicKey := peerKey.PublicKey()
	testSubscriberPrivateKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	testSubscriberPublicKey := testSubscriberPrivateKey.PublicKey()
	testObjectSimple := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey.DID(),
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	require.NoError(t, err)
	testObjectStreamRoot := &object.Object{
		Type: "foo-root",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey.DID(),
		},
		Data: tilde.Map{
			"root": tilde.String("true"),
		},
	}
	testObjectWithStream := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey.DID(),
			Root:  testObjectStreamRoot.Hash(),
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	bar1 := &object.Object{
		Metadata: object.Metadata{
			Sequence: 1,
		},
		Data: tilde.Map{
			"foo": tilde.String("bar1"),
		},
	}
	bar2 := &object.Object{
		Metadata: object.Metadata{
			Sequence: 2,
		},
		Data: tilde.Map{
			"foo": tilde.String("bar2"),
		},
	}
	testObjectWithStreamUpdated := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner:    testOwnPublicKey.DID(),
			Root:     testObjectStreamRoot.Hash(),
			Sequence: 3,
			Parents: object.Parents{
				"*": tilde.SortHashes(
					[]tilde.Hash{
						bar1.Hash(),
						bar2.Hash(),
					},
				),
			},
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
		},
	}
	testObjectSubscriptionInline := object.MustMarshal(
		&stream.Subscription{
			Metadata: object.Metadata{
				Owner:    testSubscriberPublicKey.DID(),
				Root:     testObjectStreamRoot.Hash(),
				Sequence: 6,
			},
		},
	)
	testObjectWithStreamInlineUpdated := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner:    testOwnPublicKey.DID(),
			Root:     testObjectStreamRoot.Hash(),
			Sequence: 7,
			Parents: object.Parents{
				"*": tilde.SortHashes(
					[]tilde.Hash{
						bar1.Hash(),
						bar2.Hash(),
						testObjectSubscriptionInline.Hash(),
					},
				),
			},
		},
		Data: tilde.Map{
			"foo": tilde.String("bar"),
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
		name: "should fail, simple object",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
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
		want:    nil,
		wantErr: true,
	}, {
		name: "should pass, stream event",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]*object.Object{
								bar1,
								bar2,
							},
						),
						nil,
					)
				m.EXPECT().
					GetStreamLeaves(testObjectStreamRoot.Hash()).
					Return(
						tilde.SortHashes(
							[]tilde.Hash{
								bar1.Hash(),
								bar2.Hash(),
							},
						),
						nil,
					)
				m.EXPECT().
					Get(bar1.Hash()).
					Return(
						bar1,
						nil,
					)
				m.EXPECT().
					Get(bar2.Hash()).
					Return(
						bar2,
						nil,
					)
				m.EXPECT().
					Put(testObjectWithStreamUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnPeerKey: peerKey,
					SendCalls: []error{
						nil,
					},
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
			o: object.Copy(testObjectWithStream),
		},
		want: testObjectWithStreamUpdated,
	}, {
		name: "should pass, stream event, with subscribers",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetStreamLeaves(testObjectStreamRoot.Hash()).
					Return(
						[]tilde.Hash{
							bar1.Hash(),
							bar2.Hash(),
						},
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]*object.Object{
								bar1,
								bar2,
							},
						),
						nil,
					)
				m.EXPECT().
					Get(bar1.Hash()).
					Return(
						bar1,
						nil,
					)
				m.EXPECT().
					Get(bar2.Hash()).
					Return(
						bar2,
						nil,
					)
				m.EXPECT().
					Put(testObjectWithStreamUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnPeerKey: peerKey,
					SendCalls: []error{
						nil,
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{},
					},
				}
				t.Cleanup(func() {
					// TODO this is flaky, depending on how fast the runner is
					// go m.announceObject might or might not get called
					// assert.Equal(t, 1, m.SendCalled())
				})
				return m
			},
			resolver: func(t *testing.T) resolver.Resolver {
				m := resolvermock.NewMockResolver(
					gomock.NewController(t),
				)
				return m
			},
			receivedSubscriptions: []*object.Object{
				object.MustMarshal(
					&stream.Subscription{
						Metadata: object.Metadata{
							Owner: testSubscriberPublicKey.DID(),
						},
						RootHashes: []tilde.Hash{
							testObjectStreamRoot.Hash(),
						},
					},
				),
			},
		},
		args: args{
			o: object.Copy(testObjectWithStream),
		},
		want: testObjectWithStreamUpdated,
	}, {
		// root
		//  |
		//  o1
		//  |
		//  o2
		//  |
		//  o3 (subscription)
		name: "should pass, stream event, with inline subscribers",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetStreamLeaves(testObjectStreamRoot.Hash()).
					Return(
						[]tilde.Hash{
							bar1.Hash(),
							bar2.Hash(),
							testObjectSubscriptionInline.Hash(),
						},
						nil,
					)
				m.EXPECT().
					Get(bar1.Hash()).
					Return(
						bar1,
						nil,
					)
				m.EXPECT().
					Get(bar2.Hash()).
					Return(
						bar2,
						nil,
					)
				m.EXPECT().
					Get(testObjectSubscriptionInline.Hash()).
					Return(
						testObjectSubscriptionInline,
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]*object.Object{
								bar1,
								bar2,
								testObjectSubscriptionInline,
							},
						),
						nil,
					)
				m.EXPECT().
					Put(testObjectWithStreamInlineUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnPeerKey: peerKey,
					SendCalls: []error{
						nil,
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{},
					},
				}
				t.Cleanup(func() {
					// TODO this is flaky, depending on how fast the runner is
					// go m.announceObject might or might not get called
					// assert.Equal(t, 1, m.SendCalled())
				})
				return m
			},
			resolver: func(t *testing.T) resolver.Resolver {
				m := resolvermock.NewMockResolver(
					gomock.NewController(t),
				)
				return m
			},
			receivedSubscriptions: []*object.Object{},
		},
		args: args{
			o: object.Copy(testObjectWithStream),
		},
		want: &object.Object{
			Type: "foo",
			Metadata: object.Metadata{
				Root:     testObjectStreamRoot.Hash(),
				Owner:    testOwnPublicKey.DID(),
				Sequence: 7,
				Parents: object.Parents{
					"*": tilde.SortHashes(
						[]tilde.Hash{
							bar1.Hash(),
							bar2.Hash(),
							testObjectSubscriptionInline.Hash(),
						},
					),
				},
			},
			Data: tilde.Map{
				"foo": tilde.String("bar"),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(
				context.Background(),
				tt.fields.network(t),
				tt.fields.resolver(t),
				tt.fields.store(t),
			)
			for _, obj := range tt.fields.receivedSubscriptions {
				err := m.(*manager).handleStreamSubscription(
					context.Background(),
					&network.Envelope{
						Payload: obj,
					},
				)
				require.NoError(t, err)
			}
			got, err := m.Append(
				context.Background(),
				tt.args.o,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got)
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
			Owner: p.PublicKey().DID(),
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
			FilterByOwner(p.PublicKey().DID()),
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
			FilterByStreamHash(tilde.Hash("foo")),
			FilterByOwner(p.PublicKey().DID()),
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

func TestManager_Integration_AddStreamSubscription(t *testing.T) {
	peerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	ntw := &networkmock.MockNetworkSimple{
		ReturnPeerKey: peerKey,
		SubscribeCalls: []network.EnvelopeSubscription{
			&networkmock.MockSubscriptionSimple{},
		},
	}
	res := resolvermock.NewMockResolver(gomock.NewController(t))
	str, err := sqlobjectstore.New(tempSqlite3(t))
	require.NoError(t, err)

	man := New(
		context.TODO(),
		ntw,
		res,
		str,
	)

	// create a new stream
	rootObj := object.MustMarshal(
		&fixtures.TestStream{
			Metadata: object.Metadata{
				Owner: peerKey.PublicKey().DID(),
			},
			Nonce: "foo",
		},
	)
	err = man.Put(context.TODO(), rootObj)
	require.NoError(t, err)

	// subscribe to stream
	err = man.AddStreamSubscription(context.TODO(), rootObj.Hash())
	require.NoError(t, err)

	// subscribe to stream
	err = man.AddStreamSubscription(context.TODO(), rootObj.Hash())
	require.NoError(t, err)

	// check if the subscription has been added once
	r, err := str.GetByStream(rootObj.Hash())
	require.NoError(t, err)
	os, err := object.ReadAll(r)
	require.NoError(t, err)
	require.Len(t, os, 2)
}

func tempSqlite3(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", path.Join(t.TempDir(), "sqlite3.db"))
	require.NoError(t, err)
	return db
}
