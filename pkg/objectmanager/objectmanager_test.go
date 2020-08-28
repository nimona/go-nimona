package objectmanager

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/gomockutil"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/localpeermock"
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
		store    func(*testing.T) objectstore.Store
		exchange func(*testing.T) exchange.Exchange
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
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{
						nil,
					},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{
							Objects: []*exchange.Envelope{{
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
				exchange:    tt.fields.exchange(t),
				newNonce: func() string {
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
		store    func(*testing.T) objectstore.Store
		exchange func(*testing.T) exchange.Exchange
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
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{
						nil,
						nil,
						nil,
						nil,
					},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{
							Objects: []*exchange.Envelope{{
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
						&exchangemock.MockSubscriptionSimple{
							Objects: []*exchange.Envelope{{
								Payload: f00,
							}},
						},
						&exchangemock.MockSubscriptionSimple{
							Objects: []*exchange.Envelope{{
								Payload: f01,
							}},
						},
						&exchangemock.MockSubscriptionSimple{
							Objects: []*exchange.Envelope{{
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
				exchange:    tt.fields.exchange(t),
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
		localpeer              func(*testing.T) localpeer.LocalPeer
		exchange              func(*testing.T) exchange.Exchange
		resolver              func(*testing.T) resolver.Resolver
		receivedSubscriptions []object.Object
		registeredTypes       []string
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
			localpeer: func(t *testing.T) localpeer.LocalPeer {
				m := localpeermock.NewMockLocalPeer(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPrimaryIdentityKey().
					Return(testOwnPrivateKey)
				return m
			},
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{},
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
			localpeer: func(t *testing.T) localpeer.LocalPeer {
				m := localpeermock.NewMockLocalPeer(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPrimaryIdentityKey().
					Return(testOwnPrivateKey)
				return m
			},
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{},
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
					Put(testObjectWithStreamUpdated)
				return m
			},
			localpeer: func(t *testing.T) localpeer.LocalPeer {
				m := localpeermock.NewMockLocalPeer(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPrimaryIdentityKey().
					Return(testOwnPrivateKey)
				return m
			},
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{},
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
						time.Hour,
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
			localpeer: func(t *testing.T) localpeer.LocalPeer {
				m := localpeermock.NewMockLocalPeer(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPrimaryIdentityKey().
					MaxTimes(2).
					Return(testOwnPrivateKey)
				return m
			},
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{},
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
			registeredTypes: []string{
				"foo",
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
					Put(testObjectWithStreamUpdated)
				return m
			},
			localpeer: func(t *testing.T) localpeer.LocalPeer {
				m := localpeermock.NewMockLocalPeer(
					gomock.NewController(t),
				)
				m.EXPECT().
					GetPrimaryIdentityKey().
					Return(testOwnPrivateKey)
				m.EXPECT().
					GetPrimaryPeerKey().
					Return(testOwnPrivateKey)
				return m
			},
			exchange: func(t *testing.T) exchange.Exchange {
				m := &exchangemock.MockExchangeSimple{
					SendCalls: []error{
						nil,
					},
					SubscribeCalls: []exchange.EnvelopeSubscription{
						&exchangemock.MockSubscriptionSimple{},
					},
				}
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
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(
				context.Background(),
				WithStore(tt.fields.store(t)),
				WithLocalPeer(tt.fields.localpeer(t)),
				WithExchange(tt.fields.exchange(t)),
				WithResolver(tt.fields.resolver(t)),
			)
			for _, obj := range tt.fields.receivedSubscriptions {
				err := m.(*manager).handleStreamSubscription(
					context.Background(),
					&exchange.Envelope{
						Payload: obj,
					},
				)
				require.NoError(t, err)
			}
			for _, t := range tt.fields.registeredTypes {
				m.RegisterType(t, time.Hour)
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
