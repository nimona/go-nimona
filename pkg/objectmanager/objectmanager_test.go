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
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/hyperspace/resolvermock"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/stream"
)

func TestManager_Request(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.ConnectionInfo{
		PublicKey: testPeerKey.PublicKey(),
	}
	f00 := &object.Object{
		Data: object.Map{
			"f00": object.String("f00"),
		},
	}
	type fields struct {
		store   func(*testing.T) objectstore.Store
		network func(*testing.T) network.Network
	}
	type args struct {
		ctx      context.Context
		rootHash object.Hash
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
								Payload: object.Response{
									RequestID: "7",
									Object:    object.Copy(f00),
								}.ToObject(),
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
				tt.want.ToMap(),
				got.ToMap(),
			)
		})
	}
}

func TestManager_handleObjectRequest(t *testing.T) {
	localPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	peer1Key, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		PublicKey: peer1Key.PublicKey(),
	}

	localPeer := localpeer.New()
	localPeer.PutPrimaryPeerKey(localPeerKey)
	localPeer.PutPrimaryIdentityKey(localPeerKey)

	f00 := peer1.ToObject()
	f01 := &object.Object{
		Metadata: object.Metadata{},
		Data: object.Map{
			"f01":  object.String("f01"),
			"asdf": object.Copy(f00),
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
		rootHash object.Hash
		peer     *peer.ConnectionInfo
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.Object
		wantErr bool
	}{{
		name: "object returned",
		fields: fields{
			storeHandler: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				m.EXPECT().Get(f01.Hash()).Return(object.Copy(f01), nil).MaxTimes(2)
				m.EXPECT().GetPinned().Return(nil, nil)
				return m
			},
			networkHandler: func(
				t *testing.T,
				ctx context.Context,
				wg *sync.WaitGroup,
				want *object.Object,
			) network.Network {
				m := networkmock.NewMockNetwork(gomock.NewController(t))
				m.EXPECT().LocalPeer().Return(localPeer)
				m.EXPECT().Subscribe(gomock.Any()).Return(
					&networkmock.MockSubscriptionSimple{
						Objects: []*network.Envelope{{
							Payload: object.Request{
								RequestID:  "8",
								ObjectHash: f01.Hash(),
							}.ToObject(),
						}},
					},
				)
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
						recipient crypto.PublicKey,
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
		want: object.Response{
			Metadata: object.Metadata{
				Owner: localPeerKey.PublicKey(),
			},
			Object:    f01,
			RequestID: "8",
		}.ToObject(),
	},
		{
			name: "object missing, return empty response",
			fields: fields{
				storeHandler: func(t *testing.T) objectstore.Store {
					m := objectstoremock.NewMockStore(gomock.NewController(t))
					m.EXPECT().Get(f01.Hash()).Return(nil, objectstore.ErrNotFound).MaxTimes(2)
					m.EXPECT().GetPinned().Return(nil, nil)
					return m
				},
				networkHandler: func(
					t *testing.T,
					ctx context.Context,
					wg *sync.WaitGroup,
					want *object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().LocalPeer().Return(localPeer)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Request{
									RequestID:  "8",
									ObjectHash: f01.Hash(),
								}.ToObject(),
							}},
						},
					)
					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj *object.Object,
							recipient crypto.PublicKey,
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
			want: object.Response{
				Metadata: object.Metadata{
					Owner: localPeerKey.PublicKey(),
				},
				Object:    nil,
				RequestID: "8",
			}.ToObject(),
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

func TestManager_RequestStream(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.ConnectionInfo{
		PublicKey: testPeerKey.PublicKey(),
	}
	f00 := &object.Object{
		Type:     "foo",
		Metadata: object.Metadata{},
		Data: object.Map{
			"f00": object.String("f00"),
		},
	}
	f01 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Stream:  f00.Hash(),
			Parents: []object.Hash{f00.Hash()},
		},
		Data: object.Map{
			"f01": object.String("f01"),
		},
	}
	f02 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Stream:  f00.Hash(),
			Parents: []object.Hash{f01.Hash()},
		},
		Data: object.Map{
			"f02": object.String("f02"),
		},
	}

	type fields struct {
		store   func(*testing.T) objectstore.Store
		network func(*testing.T) network.Network
		local   func(*testing.T) localpeer.LocalPeer
	}
	type args struct {
		ctx      context.Context
		rootHash object.Hash
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
			local: func(t *testing.T) localpeer.LocalPeer {
				l := localpeer.New()
				l.PutContentTypes("foo")
				return l
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
									RequestID: "7",
									Leaves: []object.Hash{
										f02.Hash(),
									},
								}.ToObject(),
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: f02,
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Copy(f01),
							}},
						},
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
		want: []*object.Object{
			f00,
			f01,
			f02,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				localpeer:   tt.fields.local(t),
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
						tt.want[i].ToMap(),
						objs[i].ToMap(),
						"for index %d", i,
					)
				}
			}
		})
	}
}

func TestManager_handleStreamRequest(t *testing.T) {
	localPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	peer1Key, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		PublicKey: peer1Key.PublicKey(),
	}

	localPeer := localpeer.New()
	localPeer.PutPrimaryPeerKey(localPeerKey)
	localPeer.PutPrimaryIdentityKey(localPeerKey)

	f00 := &object.Object{
		Type:     "foo",
		Metadata: object.Metadata{},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}

	f01 := &object.Object{
		Type: "foo-child",
		Metadata: object.Metadata{
			Stream: f00.Hash(),
			Parents: []object.Hash{
				f00.Hash(),
			},
		},
		Data: object.Map{
			"foo": object.String("bar"),
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
		rootHash object.Hash
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
					GetPinned().
					Return(nil, nil)
				m.EXPECT().
					GetStreamLeaves(f00.Hash()).
					Return(
						[]object.Hash{
							f01.Hash(),
						},
						nil,
					)
				m.EXPECT().
					GetByStream(f00.Hash()).
					Return(
						object.NewReadCloserFromObjects(
							[]*object.Object{
								f00,
								f01,
							},
						),
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
				m.EXPECT().LocalPeer().Return(localPeer)
				m.EXPECT().Subscribe(gomock.Any()).Return(
					&networkmock.MockSubscriptionSimple{
						Objects: []*network.Envelope{{
							Payload: stream.Request{
								RequestID: "7",
								RootHash:  f00.Hash(),
							}.ToObject(),
						}},
					},
				)
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
						recipient crypto.PublicKey,
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
		want: stream.Response{
			Metadata: object.Metadata{
				Owner: localPeerKey.PublicKey(),
			},
			RequestID: "7",
			RootHash:  f00.Hash(),
			Leaves:    []object.Hash{f01.Hash()},
		}.ToObject(),
	}, {
		name: "should pass, unknown stream",
		fields: fields{
			storeHandler: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				m.EXPECT().
					GetPinned().
					Return(nil, nil)
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
				m.EXPECT().LocalPeer().Return(localPeer)
				m.EXPECT().Subscribe(gomock.Any()).Return(
					&networkmock.MockSubscriptionSimple{
						Objects: []*network.Envelope{{
							Payload: stream.Request{
								RequestID: "7",
								RootHash:  f00.Hash(),
							}.ToObject(),
						}},
					},
				)
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
						recipient crypto.PublicKey,
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
		want: stream.Response{
			Metadata: object.Metadata{
				Owner: localPeerKey.PublicKey(),
			},
			RequestID: "7",
			RootHash:  f00.Hash(),
			Leaves:    []object.Hash{},
		}.ToObject(),
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
	testOwnPrivateKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testOwnPublicKey := testOwnPrivateKey.PublicKey()
	testSubscriberPrivateKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testLocalPeer := localpeer.New()
	testLocalPeer.PutPrimaryPeerKey(testOwnPrivateKey)
	testLocalPeer.PutPrimaryIdentityKey(testOwnPrivateKey)
	testSubscriberPublicKey := testSubscriberPrivateKey.PublicKey()
	testSubscriberPeer := &peer.ConnectionInfo{
		PublicKey: testSubscriberPublicKey,
		Addresses: []string{
			"not-important",
		},
	}
	testObjectSimple := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey,
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}
	testObjectStreamRoot := &object.Object{
		Type: "fooRoot",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey,
		},
		Data: object.Map{
			"root": object.String("true"),
		},
	}
	testObjectWithStream := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner:  testOwnPublicKey,
			Stream: testObjectStreamRoot.Hash(),
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}
	bar1 := &object.Object{
		Data: object.Map{
			"foo": object.String("bar1"),
		},
	}
	bar2 := &object.Object{
		Data: object.Map{
			"foo": object.String("bar2"),
		},
	}
	testObjectWithStreamUpdated := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner:  testOwnPublicKey,
			Stream: testObjectStreamRoot.Hash(),
			Parents: []object.Hash{
				bar1.Hash(),
				bar2.Hash(),
			},
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}
	testObjectSubscriptionInline := stream.Subscription{
		Metadata: object.Metadata{
			Owner:  testSubscriberPublicKey,
			Stream: testObjectStreamRoot.Hash(),
		},
	}.ToObject()
	testObjectComplex := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: object.Map{
			"foo":           object.String("bar"),
			"nested-simple": testObjectSimple,
		},
	}
	testObjectComplexUpdated := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: object.Map{
			"foo":             object.String("bar"),
			"nested-simple:r": testObjectSimple.Hash(),
		},
	}
	testObjectComplexReturned := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: object.Map{
			"foo":           object.String("bar"),
			"nested-simple": testObjectSimple,
		},
	}
	testFeedRoot := getFeedRoot(
		testOwnPrivateKey.PublicKey(),
		getTypeForFeed(testObjectSimple.Type),
	).ToObject()
	testFeedRootHash := testFeedRoot.Hash()
	testFeedFirst := feed.Added{
		ObjectHash: []object.Hash{
			testObjectSimple.Hash(),
		},
		Metadata: object.Metadata{
			Stream: testFeedRootHash,
			Parents: []object.Hash{
				testFeedRootHash,
			},
		},
	}.ToObject()
	testFeedSecond := feed.Added{
		ObjectHash: []object.Hash{
			testObjectSimple.Hash(),
		},
		Metadata: object.Metadata{
			Stream: testFeedRootHash,
			Parents: []object.Hash{
				testFeedFirst.Hash(),
			},
		},
	}.ToObject()
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
					GetPinned().
					Return(nil, nil)
				m.EXPECT().
					Put(testObjectSimple).
					Return(nil)
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
					GetPinned().
					Return(nil, nil)
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
			o: object.Copy(testObjectComplex),
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
					GetPinned().
					Return(nil, nil)
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
						[]object.Hash{
							bar1.Hash(),
							bar2.Hash(),
						},
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
			o: object.Copy(testObjectWithStream),
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
					GetPinned().
					Return(nil, nil)
				m.EXPECT().
					Put(testObjectSimple)
				m.EXPECT().
					Get(testFeedRootHash).
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
			o: object.Copy(testObjectSimple),
		},
		want: testObjectSimple,
	}, {
		name: "should pass, simple object, registered, second item in feed",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPinned().
					Return(nil, nil)
				m.EXPECT().
					Put(testObjectSimple)
				m.EXPECT().
					Get(testFeedRootHash).
					Return(testFeedRoot, nil)
				m.EXPECT().
					Put(gomockutil.ObjectEq(testFeedSecond))
				m.EXPECT().
					GetStreamLeaves(testFeedRootHash).
					Return(
						[]object.Hash{
							testFeedFirst.Hash(),
						},
						nil,
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
			o: object.Copy(testObjectSimple),
		},
		want: testObjectSimple,
	}, {
		name: "should pass, stream event, with subscribers",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPinned().
					Return(nil, nil)
				m.EXPECT().
					GetStreamLeaves(testObjectStreamRoot.Hash()).
					Return(
						[]object.Hash{
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
				m.EXPECT().Lookup(
					gomock.Any(),
					gomock.Any(),
				).Return([]*peer.ConnectionInfo{testSubscriberPeer}, nil)
				return m
			},
			receivedSubscriptions: []*object.Object{
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
					GetPinned().
					Return(nil, nil)
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
					GetStreamLeaves(testObjectStreamRoot.Hash()).
					Return(
						[]object.Hash{
							bar1.Hash(),
							bar2.Hash(),
							testObjectSubscriptionInline.Hash(),
						},
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
				m.EXPECT().Lookup(
					gomock.Any(),
					// TODO we need a custom matcher for options
					gomock.Any(),
				).Return([]*peer.ConnectionInfo{testSubscriberPeer}, nil)
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
				Stream: testObjectStreamRoot.Hash(),
				Owner:  testOwnPublicKey,
				Parents: object.SortHashes(
					[]object.Hash{
						bar1.Hash(),
						bar2.Hash(),
						testObjectSubscriptionInline.Hash(),
					},
				),
			},
			Data: object.Map{
				"foo": object.String("bar"),
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
			got, err := m.Put(
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
	o1 := &object.Object{
		Type: "not-bar",
		Metadata: object.Metadata{
			Owner: "foo",
		},
		Data: object.Map{
			"foo": object.String("not-bar"),
		},
	}
	o2 := &object.Object{
		Type: "bar",
		Metadata: object.Metadata{
			Stream: "foo",
		},
		Data: object.Map{
			"foo": object.String("bar"),
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
			FilterByOwner("foo"),
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
			FilterByStreamHash("foo"),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}, {
		name: "subscribe by stream and owner",
		lookupOptions: []LookupOption{
			FilterByStreamHash("foo"),
			FilterByOwner("foo"),
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
