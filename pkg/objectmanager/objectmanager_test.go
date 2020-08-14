package objectmanager

import (
	"database/sql"
	"io/ioutil"
	"path"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/gomockutil"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/exchangemock"
	"nimona.io/pkg/feed"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/keychainmock"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func TestObjectRequest(t *testing.T) {
	// enable binding to local addresses
	net.BindLocal = true
	wg := sync.WaitGroup{}
	wg.Add(2)

	objectHandled := false
	objectReceived := false

	// create new peers
	kc1, n1, x1, _, mgr := newPeer(t)
	kc2, n2, x2, st2, _ := newPeer(t)

	// make up the peers
	_ = &peer.Peer{
		Metadata: object.Metadata{
			Owners: kc1.ListPublicKeys(keychain.PeerKey),
		},
		Addresses: n1.Addresses(),
	}
	p2 := &peer.Peer{
		Metadata: object.Metadata{
			Owners: kc2.ListPublicKeys(keychain.PeerKey),
		},
		Addresses: n2.Addresses(),
	}

	// create test objects
	obj := object.Object{}
	obj = obj.Set("body:s", "bar1")
	obj = obj.SetType("test/msg")

	// setup hander
	go exchange.HandleEnvelopeSubscription(
		x2.Subscribe(
			exchange.FilterByObjectType(new(object.Request).GetType()),
		),
		func(e *exchange.Envelope) error {
			o := e.Payload
			objectHandled = true

			objr := object.Request{}
			err := objr.FromObject(o)
			require.NoError(t, err)

			assert.Equal(t, obj.Hash().String(), objr.ObjectHash.String())
			wg.Done()
			return nil
		},
	)

	go exchange.HandleEnvelopeSubscription(
		x1.Subscribe(
			exchange.FilterBySender(p2.PublicKey()),
		),
		func(e *exchange.Envelope) error {
			o := e.Payload
			objectReceived = true

			assert.Equal(t, obj.Get("body:s"), o.Get("body:s"))
			wg.Done()
			return nil
		},
	)
	err := st2.Put(obj)
	assert.NoError(t, err)

	ctx := context.Background()

	objRecv, err := mgr.Request(ctx, obj.Hash(), p2)
	assert.NoError(t, err)
	assert.Equal(t, obj.ToMap(), objRecv.ToMap())

	wg.Wait()

	assert.True(t, objectHandled)
	assert.True(t, objectReceived)
}

func newPeer(
	t *testing.T,
) (
	keychain.Keychain,
	net.Network,
	exchange.Exchange,
	*sqlobjectstore.Store,
	ObjectManager,
) {
	dblite := tempSqlite3(t)
	store, err := sqlobjectstore.New(dblite)
	assert.NoError(t, err)

	ctx := context.Background()

	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	eb := eventbus.New()

	kc := keychain.New()
	kc.Put(keychain.PrimaryPeerKey, pk)

	n := net.New(
		net.WithKeychain(kc),
	)
	_, err = n.Listen(ctx, "127.0.0.1:0")
	require.NoError(t, err)

	x := exchange.New(
		ctx,
		exchange.WithNet(n),
		exchange.WithKeychain(kc),
		exchange.WithEventbus(eb),
	)

	mgr := New(
		ctx,
		WithExchange(x),
		WithStore(store),
	)

	return kc, n, x, store, mgr
}

func tempSqlite3(t *testing.T) *sql.DB {
	dirPath, err := ioutil.TempDir("", "nimona-store-sql")
	require.NoError(t, err)
	db, err := sql.Open("sqlite3", path.Join(dirPath, "sqlite3.db"))
	require.NoError(t, err)
	return db
}

func Test_manager_RequestStream(t *testing.T) {
	testPeerKey, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)
	testPeer := &peer.Peer{
		Metadata: object.Metadata{
			Owners: []crypto.PublicKey{
				testPeerKey.PublicKey(),
			},
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
				store:    tt.fields.store(t),
				exchange: tt.fields.exchange(t),
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
	testOwnPublicKeys := []crypto.PublicKey{
		testOwnPrivateKey.PublicKey(),
	}
	testObjectSimple := object.Object{}.
		SetType("foo").
		Set("foo:s", "bar")
	testObjectSimpleUpdated := object.Object{}.
		SetType("foo").
		Set("foo:s", "bar").
		SetOwners(testOwnPublicKeys)
	testObjectWithStream := object.Object{}.
		SetType("foo").
		SetStream("streamRoot").
		Set("foo:s", "bar")
	testObjectWithStreamUpdated := object.Object{}.
		SetType("foo").
		SetStream("streamRoot").
		Set("foo:s", "bar").
		SetOwners(testOwnPublicKeys).
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
		SetOwners(testOwnPublicKeys).
		Set("nested-simple:r", object.Ref(testObjectSimple.Hash()))
	testObjectComplexReturned := object.Object{}.
		SetType("foo-complex").
		Set("foo:s", "bar").
		SetOwners(testOwnPublicKeys).
		Set("nested-simple:m", testObjectSimple.Raw())
	testFeedHash := getFeedRootHash(
		[]crypto.PublicKey{
			testOwnPrivateKey.PublicKey(),
		},
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
		store           func(*testing.T) objectstore.Store
		keychain        func(*testing.T) keychain.Keychain
		pubsub          func(*testing.T) ObjectPubSub
		registeredTypes []string
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
			keychain: func(t *testing.T) keychain.Keychain {
				m := keychainmock.NewMockKeychain(
					gomock.NewController(t),
				)
				m.EXPECT().
					ListPublicKeys(keychain.IdentityKey).
					Return(testOwnPublicKeys)
				return m
			},
			pubsub: func(t *testing.T) ObjectPubSub {
				return NewObjectPubSub()
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
			keychain: func(t *testing.T) keychain.Keychain {
				m := keychainmock.NewMockKeychain(
					gomock.NewController(t),
				)
				m.EXPECT().
					ListPublicKeys(keychain.IdentityKey).
					Return(testOwnPublicKeys)
				return m
			},
			pubsub: func(t *testing.T) ObjectPubSub {
				return NewObjectPubSub()
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
					GetByStream(object.Hash("streamRoot")).
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
			keychain: func(t *testing.T) keychain.Keychain {
				m := keychainmock.NewMockKeychain(
					gomock.NewController(t),
				)
				m.EXPECT().
					ListPublicKeys(keychain.IdentityKey).
					Return(testOwnPublicKeys)
				return m
			},
			pubsub: func(t *testing.T) ObjectPubSub {
				return NewObjectPubSub()
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
			keychain: func(t *testing.T) keychain.Keychain {
				m := keychainmock.NewMockKeychain(
					gomock.NewController(t),
				)
				m.EXPECT().
					ListPublicKeys(keychain.IdentityKey).
					MaxTimes(2).
					Return(testOwnPublicKeys)
				return m
			},
			pubsub: func(t *testing.T) ObjectPubSub {
				return NewObjectPubSub()
			},
			registeredTypes: []string{
				"foo",
			},
		},
		args: args{
			o: testObjectSimple,
		},
		want: testObjectSimpleUpdated,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				store:    tt.fields.store(t),
				keychain: tt.fields.keychain(t),
				pubsub:   tt.fields.pubsub(t),
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
