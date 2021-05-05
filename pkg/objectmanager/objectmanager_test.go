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
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func TestManager_Request(t *testing.T) {
	testPeerKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
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
		ctx     context.Context
		rootCID object.CID
		peer    *peer.ConnectionInfo
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
			ctx:     context.Background(),
			rootCID: f00.CID(),
			peer:    testPeer,
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
			got, err := m.Request(tt.args.ctx, tt.args.rootCID, tt.args.peer)
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
	localPeerKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	peer1Key, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		PublicKey: peer1Key.PublicKey(),
	}

	localPeer := localpeer.New()
	localPeer.SetPeerKey(localPeerKey)

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
		ctx     context.Context
		rootCID object.CID
		peer    *peer.ConnectionInfo
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
					m.EXPECT().Get(f01.CID()).Return(object.Copy(f01), nil).MaxTimes(2)
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
									RequestID: "8",
									ObjectCID: f01.CID(),
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
				ctx:     context.Background(),
				rootCID: f00.CID(),
				peer:    peer1,
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
					m.EXPECT().Get(f01.CID()).Return(nil, objectstore.ErrNotFound).MaxTimes(2)
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
									RequestID: "8",
									ObjectCID: f01.CID(),
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
				ctx:     context.Background(),
				rootCID: f00.CID(),
				peer:    peer1,
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
	testPeerKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
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
			Stream: f00.CID(),
			Parents: object.Parents{
				"*": []object.CID{f00.CID()},
			},
		},
		Data: object.Map{
			"f01": object.String("f01"),
		},
	}
	f02 := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Stream: f00.CID(),
			Parents: object.Parents{
				"*": []object.CID{f01.CID()},
			},
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
		ctx     context.Context
		rootCID object.CID
		peer    *peer.ConnectionInfo
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
					Get(f00.CID()).
					Return(nil, objectstore.ErrNotFound)
				m.EXPECT().
					Get(f01.CID()).
					Return(nil, objectstore.ErrNotFound)
				m.EXPECT().
					Get(f02.CID()).
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
					GetByStream(f00.CID()).
					Return(object.NewReadCloserFromObjects([]*object.Object{
						object.Copy(f00),
						object.Copy(f01),
						object.Copy(f02),
					}), err)
				return m
			},
			local: func(t *testing.T) localpeer.LocalPeer {
				l := localpeer.New()
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
									Leaves: []object.CID{
										f02.CID(),
									},
								}.ToObject(),
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Response{
									RequestID: "7",
									Object:    object.Copy(f02),
								}.ToObject(),
							}},
						},
						&networkmock.MockSubscriptionSimple{
							Objects: []*network.Envelope{{
								Payload: object.Response{
									RequestID: "7",
									Object:    object.Copy(f01),
								}.ToObject(),
							}},
						},
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
			ctx:     context.Background(),
			rootCID: f00.CID(),
			peer:    testPeer,
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
				pubsub:      NewObjectPubSub(),
				objectstore: tt.fields.store(t),
				network:     tt.fields.network(t),
				newRequestID: func() string {
					return "7"
				},
			}
			got, err := m.RequestStream(tt.args.ctx, tt.args.rootCID, tt.args.peer)
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
	localPeerKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	peer1Key, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	peer1 := &peer.ConnectionInfo{
		PublicKey: peer1Key.PublicKey(),
	}

	localPeer := localpeer.New()
	localPeer.SetPeerKey(localPeerKey)

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
			Stream: f00.CID(),
			Parents: object.Parents{
				"*": []object.CID{
					f00.CID(),
				},
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
		ctx     context.Context
		rootCID object.CID
		peer    *peer.ConnectionInfo
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
					GetStreamLeaves(f00.CID()).
					Return(
						[]object.CID{
							f01.CID(),
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
				m.EXPECT().LocalPeer().Return(localPeer)
				m.EXPECT().Subscribe(gomock.Any()).Return(
					&networkmock.MockSubscriptionSimple{
						Objects: []*network.Envelope{{
							Payload: stream.Request{
								RequestID: "7",
								RootCID:   f00.CID(),
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
			ctx:     context.Background(),
			rootCID: f00.CID(),
			peer:    peer1,
		},
		want: stream.Response{
			Metadata: object.Metadata{
				Owner: localPeerKey.PublicKey(),
			},
			RequestID: "7",
			RootCID:   f00.CID(),
			Leaves:    []object.CID{f01.CID()},
		}.ToObject(),
	}, {
		name: "should pass, unknown stream",
		fields: fields{
			storeHandler: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(gomock.NewController(t))
				m.EXPECT().
					GetStreamLeaves(f00.CID()).
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
								RootCID:   f00.CID(),
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
			ctx:     context.Background(),
			rootCID: f00.CID(),
			peer:    peer1,
		},
		want: stream.Response{
			Metadata: object.Metadata{
				Owner: localPeerKey.PublicKey(),
			},
			RequestID: "7",
			RootCID:   f00.CID(),
			Leaves:    []object.CID{},
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
	testOwnPrivateKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	testOwnPublicKey := testOwnPrivateKey.PublicKey()
	testSubscriberPrivateKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)
	testLocalPeer := localpeer.New()
	testLocalPeer.SetPeerKey(testOwnPrivateKey)
	testSubscriberPublicKey := testSubscriberPrivateKey.PublicKey()
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
		Type: "foo-root",
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
			Stream: testObjectStreamRoot.CID(),
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
			Stream: testObjectStreamRoot.CID(),
			Parents: object.Parents{
				"*": object.SortCIDs(
					[]object.CID{
						bar1.CID(),
						bar2.CID(),
					},
				),
			},
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}
	testObjectSubscriptionInline := stream.Subscription{
		Metadata: object.Metadata{
			Owner:  testSubscriberPublicKey,
			Stream: testObjectStreamRoot.CID(),
		},
	}.ToObject()
	testObjectWithStreamInlineUpdated := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Owner:  testOwnPublicKey,
			Stream: testObjectStreamRoot.CID(),
			Parents: object.Parents{
				"*": object.SortCIDs(
					[]object.CID{
						bar1.CID(),
						bar2.CID(),
						testObjectSubscriptionInline.CID(),
					},
				),
			},
		},
		Data: object.Map{
			"foo": object.String("bar"),
		},
	}
	testObjectComplex := &object.Object{
		Type:     "foo-complex",
		Metadata: object.Metadata{},
		Data: object.Map{
			"foo":           object.String("bar"),
			"nested-simple": testObjectSimple,
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
					Put(testObjectComplex)
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
		want: testObjectComplex,
	}, {
		name: "should pass, stream event",
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.CID()).
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
					GetStreamLeaves(testObjectStreamRoot.CID()).
					Return(
						object.SortCIDs(
							[]object.CID{
								bar1.CID(),
								bar2.CID(),
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
					GetStreamLeaves(testObjectStreamRoot.CID()).
					Return(
						[]object.CID{
							bar1.CID(),
							bar2.CID(),
						},
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.CID()).
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
				stream.Subscription{
					Metadata: object.Metadata{
						Owner: testSubscriberPublicKey,
					},
					RootCIDs: []object.CID{
						testObjectStreamRoot.CID(),
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
					GetStreamLeaves(testObjectStreamRoot.CID()).
					Return(
						[]object.CID{
							bar1.CID(),
							bar2.CID(),
							testObjectSubscriptionInline.CID(),
						},
						nil,
					)
				m.EXPECT().
					GetByStream(testObjectStreamRoot.CID()).
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
					ReturnLocalPeer: testLocalPeer,
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
				Stream: testObjectStreamRoot.CID(),
				Owner:  testOwnPublicKey,
				Parents: object.Parents{
					"*": object.SortCIDs(
						[]object.CID{
							bar1.CID(),
							bar2.CID(),
							testObjectSubscriptionInline.CID(),
						},
					),
				},
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
	p, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	o0 := &object.Object{
		Type: "foo",
	}
	o1 := &object.Object{
		Type: "not-bar",
		Metadata: object.Metadata{
			Owner: p.PublicKey(),
		},
		Data: object.Map{
			"foo": object.String("not-bar"),
		},
	}
	o2 := &object.Object{
		Type: "bar",
		Metadata: object.Metadata{
			Stream: o0.CID(),
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
		name: "subscribe by cid",
		lookupOptions: []LookupOption{
			FilterByCID(o2.CID()),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}, {
		name: "subscribe by owner",
		lookupOptions: []LookupOption{
			FilterByOwner(p.PublicKey()),
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
			FilterByStreamCID(o0.CID()),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{o2},
	}, {
		name: "subscribe by stream and owner",
		lookupOptions: []LookupOption{
			FilterByStreamCID("foo"),
			FilterByOwner(p.PublicKey()),
		},
		publish: []*object.Object{o1, o2},
		want:    []*object.Object{},
	}, {
		name: "subscribe by cid and type",
		lookupOptions: []LookupOption{
			FilterByCID(o2.CID()),
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
	prv0, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	lpr := localpeer.New()
	lpr.SetPeerKey(prv0)

	ntw := &networkmock.MockNetworkSimple{
		ReturnLocalPeer: lpr,
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
	rootObj := fixtures.TestStream{
		Metadata: object.Metadata{
			Owner: prv0.PublicKey(),
		},
		Nonce: "foo",
	}.ToObject()
	rootObj, err = man.Put(context.TODO(), rootObj)
	require.NoError(t, err)

	// subscribe to stream
	err = man.AddStreamSubscription(context.TODO(), rootObj.CID())
	require.NoError(t, err)

	// subscribe to stream
	err = man.AddStreamSubscription(context.TODO(), rootObj.CID())
	require.NoError(t, err)

	// check if the subscription has been added once
	r, err := str.GetByStream(rootObj.CID())
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
