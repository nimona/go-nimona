package objectmanager

import (
	"sync"
	"testing"

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

func TestManager_Request(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owner: testPeerKey.PublicKey(),
		},
	}
	f00 := &object.Object{
		Data: map[string]interface{}{
			"f00:s": "f00",
		},
	}
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

func TestManager_handleObjectRequest(t *testing.T) {
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
	localPeer.PutPrimaryPeerKey(localPeerKey)
	localPeer.PutPrimaryIdentityKey(localPeerKey)

	f00 := peer1.ToObject()
	f01 := &object.Object{
		Metadata: object.Metadata{},
		Data: map[string]interface{}{
			"f01:s":  "f01",
			"asdf:m": object.Copy(f00),
		},
	}

	unloadedF01 := &object.Object{
		Metadata: object.Metadata{},
		Data: map[string]interface{}{
			"f01:s":  "f01",
			"asdf:r": f00.Hash(),
		},
	}

	type fields struct {
		storeHandler   func(*testing.T) objectstore.Store
		networkHandler func(
			*testing.T,
			context.Context,
			bool, *sync.WaitGroup,
			*object.Object,
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
		want    *object.Object
		wantErr bool
	}{
		{
			name: "object should be unloaded",
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
					excludeNested bool,
					wg *sync.WaitGroup,
					want *object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().LocalPeer().Return(localPeer)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Request{
									RequestID:             "8",
									ObjectHash:            f01.Hash(),
									ExcludedNestedObjects: excludeNested,
								}.ToObject(),
							}},
						},
					)
					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj *object.Object,
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
			want: object.Response{
				Metadata: object.Metadata{
					Owner: localPeerKey.PublicKey(),
				},
				Object:    unloadedF01,
				RequestID: "8",
			}.ToObject(),
		},
		{
			name: "object should NOT be unloaded",
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
					excludeNested bool,
					wg *sync.WaitGroup,
					want *object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().LocalPeer().Return(localPeer)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Request{
									RequestID:             "8",
									ObjectHash:            f01.Hash(),
									ExcludedNestedObjects: excludeNested,
								}.ToObject(),
							}},
						},
					)
					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj *object.Object,
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
					excludeNested bool,
					wg *sync.WaitGroup,
					want *object.Object,
				) network.Network {
					m := networkmock.NewMockNetwork(gomock.NewController(t))
					m.EXPECT().LocalPeer().Return(localPeer)
					m.EXPECT().Subscribe(gomock.Any()).Return(
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Request{
									RequestID:             "8",
									ObjectHash:            f01.Hash(),
									ExcludedNestedObjects: excludeNested,
								}.ToObject(),
							}},
						},
					)
					m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
						DoAndReturn(func(
							ctx context.Context,
							obj *object.Object,
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
			want: object.Response{
				Metadata: object.Metadata{
					Owner: localPeerKey.PublicKey(),
				},
				Object:    nil,
				RequestID: "8",
			}.ToObject(),
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

func TestManager_RequestStream(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owner: testPeerKey.PublicKey(),
		},
	}
	f00 := &object.Object{
		Type:     "foo",
		Metadata: object.Metadata{},
		Data: map[string]interface{}{
			"f00:s": "f00",
		},
	}
	f01 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Stream:  f00.Hash(),
			Parents: []object.Hash{f00.Hash()},
		},
		Data: map[string]interface{}{
			"f01:s": "f01",
		},
	}
	f02 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Stream:  f00.Hash(),
			Parents: []object.Hash{f01.Hash()},
		},
		Data: map[string]interface{}{
			"f02:s": "f02",
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
		peer     *peer.Peer
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

	peer1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: peer1Key.PublicKey(),
		},
	}

	localPeer := localpeer.New()
	localPeer.PutPrimaryPeerKey(localPeerKey)
	localPeer.PutPrimaryIdentityKey(localPeerKey)

	f00 := &object.Object{
		Type:     "foo",
		Metadata: object.Metadata{},
		Data: map[string]interface{}{
			"foo": "bar",
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
		Data: map[string]interface{}{
			"foo": "bar",
		},
	}

	type fields struct {
		storeHandler   func(*testing.T) objectstore.Store
		networkHandler func(
			*testing.T,
			context.Context,
			bool, *sync.WaitGroup,
			*object.Object,
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
				excludeNested bool,
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
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
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
					GetByStream(f00.Hash()).
					Return(nil, objectstore.ErrNotFound)
				return m
			},
			networkHandler: func(
				t *testing.T,
				ctx context.Context,
				excludeNested bool,
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
				m.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(
						ctx context.Context,
						obj *object.Object,
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
	testSubscriberPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owner: testSubscriberPublicKey,
		},
		Addresses: []string{
			"not-important",
		},
	}
	testObjectSimple := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey,
		},
		Data: map[string]interface{}{
			"foo:s": "bar",
		},
	}
	testObjectStreamRoot := &object.Object{
		Type: "fooRoot",
		Metadata: object.Metadata{
			Owner: testOwnPublicKey,
		},
		Data: map[string]interface{}{
			"root:s": "true",
		},
	}
	testObjectWithStream := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner:  testOwnPublicKey,
			Stream: testObjectStreamRoot.Hash(),
		},
		Data: map[string]interface{}{
			"foo:s": "bar",
		},
	}
	bar1 := &object.Object{
		Data: map[string]interface{}{
			"foo:s": "bar1",
		},
	}
	bar2 := &object.Object{
		Data: map[string]interface{}{
			"foo:s": "bar2",
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
		Data: map[string]interface{}{
			"foo:s": "bar",
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
		Data: map[string]interface{}{
			"foo:s":           "bar",
			"nested-simple:m": testObjectSimple,
		},
	}
	testObjectComplexUpdated := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: map[string]interface{}{
			"foo:s":           "bar",
			"nested-simple:r": testObjectSimple.Hash(),
		},
	}
	testObjectComplexReturned := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: map[string]interface{}{
			"foo:s":           "bar",
			"nested-simple:m": testObjectSimple,
		},
	}
	testFeedHash := getFeedRootHash(
		testOwnPrivateKey.PublicKey(),
		getTypeForFeed(testObjectSimple.Type),
	)
	testFeedFirst := feed.Added{
		ObjectHash: []object.Hash{
			testObjectSimple.Hash(),
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
					MaxTimes(2).
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
			Data: map[string]interface{}{
				"foo:s": "bar",
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
