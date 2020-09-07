package objectmanager

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/gomockutil"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/resolvermock"
	"nimona.io/pkg/stream"
)

func Test_manager_Request(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owner: testPeerKey.PublicKey(),
		},
	}
	f00 := object.Object{}.
		Set("f00:s", "f00")
	type fields struct {
		store   func(*testing.T) objectstore.Store
		network func(*testing.T) network.Network
	}
	type args struct {
		ctx      context.Context
		rootHash object.Hash
		peer     *peer.Peer
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
								Payload: f00,
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
		want: &f00,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				objectstore: tt.fields.store(t),
				network:     tt.fields.network(t),
				newNonce: func() string {
					return "7"
				},
			}
			got, err := m.Request(tt.args.ctx, tt.args.rootHash, tt.args.peer, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(
				t,
				tt.want.ToMap(),
				got.ToMap(),
			)
		})
	}
}

func Test_manager_handle_request(t *testing.T) {
	localPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	peer1Key, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	peer1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: peer1Key.PublicKey(),
		},
	}

	localPeer := localpeer.New()
	localPeer.PutPrimaryIdentityKey(localPeerKey)

	f00 := peer1.ToObject()
	f01 := object.Object{}.
		Set("f01:s", "f01").
		Set("asdf:m", f00.Raw())

	unloadedF01 := f01.Set("asdf:m", nil)
	unloadedF01 = unloadedF01.Set("asdf:r", object.Ref(f00.Hash()))

	type fields struct {
		storeHandler   func(*testing.T) objectstore.Store
		networkHandler func(
			*testing.T,
			context.Context,
			bool, *sync.WaitGroup,
			object.Object,
		) network.Network
		resolver func(*testing.T) resolver.Resolver
	}
	type args struct {
		ctx           context.Context
		rootHash      object.Hash
		peer          *peer.Peer
		excludeNested bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    object.Object
		wantErr bool
	}{
		{
			name: "object should be unloaded",
			fields: fields{
				storeHandler: func(t *testing.T) objectstore.Store {
					m := objectstoremock.NewMockStore(gomock.NewController(t))
					m.EXPECT().Get(f01.Hash()).Return(f01, nil).MaxTimes(2)
					return m
				},
				networkHandler: func(
					t *testing.T,
					ctx context.Context,
					excludeNested bool,
					wg *sync.WaitGroup,
					want object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().LocalPeer().Return(localPeer)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Request{
									ObjectHash:            f01.Hash(),
									ExcludedNestedObjects: excludeNested,
								}.ToObject(),
							}},
						},
					)

					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj object.Object,
							recipient *peer.Peer,
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
				ctx:           context.Background(),
				rootHash:      f00.Hash(),
				peer:          peer1,
				excludeNested: true,
			},
			want: unloadedF01,
		},
		{
			name: "object should NOT be unloaded",
			fields: fields{
				storeHandler: func(t *testing.T) objectstore.Store {
					m := objectstoremock.NewMockStore(gomock.NewController(t))
					m.EXPECT().Get(f01.Hash()).Return(f01, nil).MaxTimes(2)
					return m
				},
				networkHandler: func(
					t *testing.T,
					ctx context.Context,
					excludeNested bool,
					wg *sync.WaitGroup,
					want object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().LocalPeer().Return(localPeer)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Request{
									ObjectHash:            f01.Hash(),
									ExcludedNestedObjects: excludeNested,
								}.ToObject(),
							}},
						},
					)

					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj object.Object,
							recipient *peer.Peer,
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
				ctx:           context.Background(),
				rootHash:      f00.Hash(),
				peer:          peer1,
				excludeNested: false,
			},
			want: f01,
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
					tt.args.excludeNested,
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

func Test_manager_RequestStream(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owner: testPeerKey.PublicKey(),
		},
	}
	f00 := object.Object{}.
		Set("f00:s", "f00")
	f01 := object.Object{}.
		Set("f01:s", "f01")
	f02 := object.Object{}.
		Set("f02:s", "f02")

	type fields struct {
		store   func(*testing.T) objectstore.Store
		network func(*testing.T) network.Network
	}
	type args struct {
		ctx      context.Context
		rootHash object.Hash
		peer     *peer.Peer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []object.Object
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
						nil,
						nil,
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: stream.Response{
									Nonce: "7",
									Children: []object.Hash{
										f00.Hash(),
										f01.Hash(),
										f02.Hash(),
									},
								}.ToObject(),
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: f00,
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: f01,
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: f02,
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
		want: []object.Object{
			f00,
			f01,
			f02,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				objectstore: tt.fields.store(t),
				network:     tt.fields.network(t),
				newNonce: func() string {
					return "7"
				},
			}
			got, err := m.RequestStream(tt.args.ctx, tt.args.rootHash, tt.args.peer)
			if (err != nil) != tt.wantErr {
				t.Errorf("manager.RequestStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				objs := []object.Object{}
				for {
					obj, err := got.Read()
					if err == object.ErrReaderDone {
						break
					}
					if err != nil {
						break
					}
					objs = append(objs, *obj)
				}
				require.Equal(t, len(tt.want), len(objs))
				for i := 0; i < len(tt.want); i++ {
					assert.Equal(
						t,
						tt.want[i].ToMap(),
						objs[i].ToMap(),
						"for index %d", i,
					)
				}
			}
		})
	}
}

func Test_manager_Put(t *testing.T) {
	testOwnPrivateKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testOwnPublicKey := testOwnPrivateKey.PublicKey()
	testSubscriberPrivateKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testLocalPeer := localpeer.New()
	testLocalPeer.PutPrimaryPeerKey(testOwnPrivateKey)
	testLocalPeer.PutPrimaryIdentityKey(testOwnPrivateKey)
	testSubscriberPublicKey := testSubscriberPrivateKey.PublicKey()
	testSubscriberPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owner: testSubscriberPublicKey,
		},
		Addresses: []string{
			"not-important",
		},
	}
	testObjectSimple := object.Object{}.
		SetType("foo").
		Set("foo:s", "bar")
	testObjectSimpleUpdated := object.Object{}.
		SetType("foo").
		Set("foo:s", "bar").
		SetOwner(testOwnPublicKey)
	testObjectStreamRoot := object.Object{}.
		SetType("fooRoot").
		Set("root:s", "true")
	testObjectWithStream := object.Object{}.
		SetType("foo").
		SetStream(testObjectStreamRoot.Hash()).
		Set("foo:s", "bar")
	testObjectWithStreamUpdated := object.Object{}.
		SetType("foo").
		SetStream(testObjectStreamRoot.Hash()).
		Set("foo:s", "bar").
		SetOwner(testOwnPublicKey).
		SetParents(
			[]object.Hash{
				object.Object{}.Set("foo:s", "bar1").Hash(),
				object.Object{}.Set("foo:s", "bar2").Hash(),
			},
		)
	testObjectSubscriptionInline := stream.Subscription{
		Metadata: object.Metadata{
			Owner:  testSubscriberPublicKey,
			Stream: testObjectStreamRoot.Hash(),
		},
	}.ToObject()
	testObjectComplex := object.Object{}.
		SetType("foo-complex").
		Set("foo:s", "bar").
		Set("nested-simple:m", testObjectSimple.Raw())
	testObjectComplexUpdated := object.Object{}.
		SetType("foo-complex").
		Set("foo:s", "bar").
		SetOwner(testOwnPublicKey).
		Set("nested-simple:r", object.Ref(testObjectSimple.Hash()))
	testObjectComplexReturned := object.Object{}.
		SetType("foo-complex").
		Set("foo:s", "bar").
		SetOwner(testOwnPublicKey).
		Set("nested-simple:m", testObjectSimple.Raw())
	testFeedHash := getFeedRootHash(
		testOwnPrivateKey.PublicKey(),
		getTypeForFeed(testObjectSimple.GetType()),
	)
	testFeedFirst := feed.Added{
		ObjectHash: []object.Hash{
			testObjectSimpleUpdated.Hash(),
		},
		Metadata: object.Metadata{
			Stream: testFeedHash,
			Parents: []object.Hash{
				testFeedHash,
			},
		},
	}.ToObject()
	type fields struct {
		store                 func(*testing.T) objectstore.Store
		network               func(*testing.T) network.Network
		resolver              func(*testing.T) resolver.Resolver
		receivedSubscriptions []object.Object
	}
	type args struct {
		o object.Object
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    object.Object
		wantErr bool
	}{{
		name: "should pass, simple object",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					Put(testObjectSimpleUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnLocalPeer: testLocalPeer,
					SendCalls:       []error{},
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
			o: testObjectSimple,
		},
		want: testObjectSimpleUpdated,
	}, {
		name: "should pass, complex object",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					Put(testObjectSimple)
				m.EXPECT().
					Put(testObjectComplexUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnLocalPeer: testLocalPeer,
					SendCalls:       []error{},
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
			o: testObjectComplex,
		},
		want: testObjectComplexReturned,
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
							[]object.Object{
								object.Object{}.Set("foo:s", "bar1"),
								object.Object{}.Set("foo:s", "bar2"),
							},
						),
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]object.Object{
								object.Object{}.Set("foo:s", "bar1"),
								object.Object{}.Set("foo:s", "bar2"),
							},
						),
						nil,
					)
				m.EXPECT().
					Put(testObjectWithStreamUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnLocalPeer: testLocalPeer,
					SendCalls:       []error{},
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
			o: testObjectWithStream,
		},
		want: testObjectWithStreamUpdated,
	}, {
		name: "should pass, simple object, registered, first item in feed",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					PutWithTimeout(
						testObjectSimpleUpdated,
						time.Duration(0),
					)
				m.EXPECT().
					GetByStream(testFeedHash).
					Return(nil, objectstore.ErrNotFound)
				m.EXPECT().
					Put(
						gomockutil.ObjectEq(testFeedFirst),
					)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnLocalPeer: func() localpeer.LocalPeer {
						tmpLocalPeer := localpeer.New()
						tmpLocalPeer.PutContentTypes("foo")
						tmpLocalPeer.PutPrimaryPeerKey(testOwnPrivateKey)
						tmpLocalPeer.PutPrimaryIdentityKey(testOwnPrivateKey)
						return tmpLocalPeer
					}(),
					SendCalls: []error{},
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
			o: testObjectSimple,
		},
		want: testObjectSimpleUpdated,
	}, {
		name: "should pass, stream event, with subscribers",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]object.Object{
								object.Object{}.Set("foo:s", "bar1"),
								object.Object{}.Set("foo:s", "bar2"),
							},
						),
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]object.Object{
								object.Object{}.Set("foo:s", "bar1"),
								object.Object{}.Set("foo:s", "bar2"),
							},
						),
						nil,
					)
				m.EXPECT().
					Put(testObjectWithStreamUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnLocalPeer: testLocalPeer,
					SendCalls: []error{
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{},
					},
				}
				t.Cleanup(func() {
					assert.Equal(t, 1, m.SendCalled)
				})
				return m
			},
			resolver: func(t *testing.T) resolver.Resolver {
				m := resolvermock.NewMockResolver(
					gomock.NewController(t),
				)
				r := make(chan *peer.Peer, 10)
				r <- testSubscriberPeer
				close(r)
				m.EXPECT().Lookup(
					gomock.Any(),
					// TODO we need a custom matcher for options
					gomock.Any(),
				).Return(r, nil)
				return m
			},
			receivedSubscriptions: []object.Object{
				stream.Subscription{
					Metadata: object.Metadata{
						Owner: testSubscriberPublicKey,
					},
					RootHashes: []object.Hash{
						testObjectStreamRoot.Hash(),
					},
				}.ToObject(),
			},
		},
		args: args{
			o: testObjectWithStream,
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
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]object.Object{
								object.Object{}.Set("foo:s", "bar1"),
								object.Object{}.Set("foo:s", "bar2"),
								testObjectSubscriptionInline,
							},
						),
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]object.Object{
								object.Object{}.Set("foo:s", "bar1"),
								object.Object{}.Set("foo:s", "bar2"),
								testObjectSubscriptionInline,
							},
						),
						nil,
					)
				m.EXPECT().
					Put(testObjectWithStreamUpdated)
				return m
			},
			network: func(t *testing.T) network.Network {
				m := &networkmock.MockNetworkSimple{
					ReturnLocalPeer: testLocalPeer,
					SendCalls: []error{
						nil,
					},
					SubscribeCalls: []network.EnvelopeSubscription{
						&networkmock.MockSubscriptionSimple{},
					},
				}
				t.Cleanup(func() {
					assert.Equal(t, 1, m.SendCalled)
				})
				return m
			},
			resolver: func(t *testing.T) resolver.Resolver {
				m := resolvermock.NewMockResolver(
					gomock.NewController(t),
				)
				r := make(chan *peer.Peer, 10)
				r <- testSubscriberPeer
				close(r)
				m.EXPECT().Lookup(
					gomock.Any(),
					// TODO we need a custom matcher for options
					gomock.Any(),
				).Return(r, nil)
				return m
			},
			receivedSubscriptions: []object.Object{},
		},
		args: args{
			o: testObjectWithStream,
		},
		want: object.Object{}.
			SetType("foo").
			SetStream(testObjectStreamRoot.Hash()).
			Set("foo:s", "bar").
			SetOwner(testOwnPublicKey).
			SetParents(
				[]object.Hash{
					object.Object{}.Set("foo:s", "bar1").Hash(),
					object.Object{}.Set("foo:s", "bar2").Hash(),
					testObjectSubscriptionInline.Hash(),
				},
			),
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
			got, err := m.Put(
				context.Background(),
				tt.args.o,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.ToMap(), got.ToMap())
		})
	}
}
