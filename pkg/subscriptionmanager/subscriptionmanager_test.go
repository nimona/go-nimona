package subscriptionmanager

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/exchangemock"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/keychainmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/resolvermock"
	"nimona.io/pkg/streammanager"
	"nimona.io/pkg/streammanagermock"
	"nimona.io/pkg/subscription"
)

func Test_subscriptionmanager_Subscribe(t *testing.T) {
	testOwners := []crypto.PublicKey{
		"me",
		"also_me",
	}
	testChain0 := subscription.SubscriptionStreamRoot{
		Owners: testOwners,
	}
	testSubscription0 := subscription.Subscription{
		Subjects: []crypto.PublicKey{
			crypto.PublicKey("foo"),
		},
		Types: []string{
			"foo/bar",
		},
		Expiry: time.
			Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).
			Format(time.RFC3339),
		Owners: testOwners,
	}
	testPeer0 := &peer.Peer{
		Owners: []crypto.PublicKey{
			"foo",
		},
	}
	defaultKeychainMock := func(t *testing.T) keychain.Keychain {
		c := gomock.NewController(t)
		m := keychainmock.NewMockKeychain(c)
		m.EXPECT().ListPublicKeys(
			keychain.IdentityKey,
		).AnyTimes().Return(
			testOwners,
		)
		return m
	}
	defaultResolverMock := func(t *testing.T) resolver.Resolver {
		r := make(chan *peer.Peer)
		go func() {
			r <- testPeer0
			close(r)
		}()
		c := gomock.NewController(t)
		m := resolvermock.NewMockResolver(c)
		m.EXPECT().Lookup(
			gomock.Any(),
			gomock.Any(),
		).DoAndReturn(
			func(
				ctx context.Context,
				opts ...resolver.LookupOption,
			) (<-chan *peer.Peer, error) {
				o := &resolver.LookupOptions{}
				opts[0](o)
				require.Equal(t, "foo", o.Lookups[0])
				return r, nil
			},
		)
		return m
	}
	defaultObjectStoreMock := func(t *testing.T) objectstore.Store {
		c := gomock.NewController(t)
		m := objectstoremock.NewMockStore(c)
		m.EXPECT().Put(
			testChain0.ToObject(),
		)
		m.EXPECT().Put(
			testSubscription0.ToObject(),
		)
		m.EXPECT().Put(
			subscription.SubscriptionAdded{
				Owners:       testOwners,
				Subscription: testSubscription0.ToObject().Hash(),
			}.ToObject(),
		)
		return m
	}
	defaultExchangeMock := func(t *testing.T) exchange.Exchange {
		c := gomock.NewController(t)
		m := exchangemock.NewMockExchange(c)
		m.EXPECT().Send(
			gomock.Any(),
			testSubscription0.ToObject(),
			testPeer0,
		).Return(nil)
		return m
	}
	defaultStreamManagerMock := func(t *testing.T) streammanager.StreamManager {
		c := gomock.NewController(t)
		m := streammanagermock.NewMockStreamManager(c)
		m.EXPECT().Put(
			subscription.SubscriptionAdded{
				Stream:       testChain0.ToObject().Hash(),
				Subscription: testSubscription0.ToObject().Hash(),
				Owners:       testOwners,
			}.ToObject(),
		).Return(nil)
		return m
	}
	type fields struct {
		keychain      func(*testing.T) keychain.Keychain
		resolver      func(*testing.T) resolver.Resolver
		exchange      func(*testing.T) exchange.Exchange
		objectstore   func(*testing.T) objectstore.Store
		streammanager func(*testing.T) streammanager.StreamManager
	}
	type args struct {
		ctx      context.Context
		subjects []crypto.PublicKey
		types    []string
		streams  []object.Hash
		expiry   time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{{
		name: "should pass, subscribe on a new chain",
		fields: fields{
			keychain:      defaultKeychainMock,
			resolver:      defaultResolverMock,
			exchange:      defaultExchangeMock,
			objectstore:   defaultObjectStoreMock,
			streammanager: defaultStreamManagerMock,
		},
		args: args{
			ctx: context.New(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}, {
		name: "should fail, could not write subscription",
		fields: fields{
			keychain: defaultKeychainMock,
			resolver: func(t *testing.T) resolver.Resolver {
				return nil
			},
			exchange: func(t *testing.T) exchange.Exchange {
				return nil
			},
			objectstore: func(t *testing.T) objectstore.Store {
				c := gomock.NewController(t)
				m := objectstoremock.NewMockStore(c)
				m.EXPECT().Put(
					testChain0.ToObject(),
				)
				m.EXPECT().Put(
					testSubscription0.ToObject(),
				).Return(
					errors.New("something bad"),
				)
				return m
			},
			streammanager: func(t *testing.T) streammanager.StreamManager {
				return nil
			},
		},
		args: args{
			ctx: context.Background(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		wantErr: true,
	}, {
		name: "should fail, failed to lookup peers",
		fields: fields{
			keychain: defaultKeychainMock,
			resolver: func(t *testing.T) resolver.Resolver {
				c := gomock.NewController(t)
				m := resolvermock.NewMockResolver(c)
				m.EXPECT().Lookup(
					gomock.Any(),
					gomock.Any(),
				).DoAndReturn(
					func(
						ctx context.Context,
						opts ...resolver.LookupOption,
					) (<-chan *peer.Peer, error) {
						return nil, errors.New("something bad")
					},
				)
				return m
			},
			exchange: func(t *testing.T) exchange.Exchange {
				return nil
			},
			objectstore: func(t *testing.T) objectstore.Store {
				c := gomock.NewController(t)
				m := objectstoremock.NewMockStore(c)
				m.EXPECT().Put(
					testChain0.ToObject(),
				)
				m.EXPECT().Put(
					testSubscription0.ToObject(),
				)
				return m
			},
			streammanager: func(t *testing.T) streammanager.StreamManager {
				return nil
			},
		},
		args: args{
			ctx: context.Background(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		wantErr: true,
	}, {
		name: "should fail, failed to send to all peers",
		fields: fields{
			keychain: defaultKeychainMock,
			resolver: defaultResolverMock,
			exchange: func(t *testing.T) exchange.Exchange {
				c := gomock.NewController(t)
				m := exchangemock.NewMockExchange(c)
				m.EXPECT().Send(
					gomock.Any(),
					testSubscription0.ToObject(),
					testPeer0,
				).Return(
					errors.New("something bad"),
				)
				return m
			},
			objectstore: func(t *testing.T) objectstore.Store {
				c := gomock.NewController(t)
				m := objectstoremock.NewMockStore(c)
				m.EXPECT().Put(
					testChain0.ToObject(),
				)
				m.EXPECT().Put(
					testSubscription0.ToObject(),
				)
				return m
			},
			streammanager: func(t *testing.T) streammanager.StreamManager {
				return nil
			},
		},
		args: args{
			ctx: context.Background(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		wantErr: true,
	}, {
		name: "should fail, failed to put on stream manager",
		fields: fields{
			keychain:    defaultKeychainMock,
			resolver:    defaultResolverMock,
			exchange:    defaultExchangeMock,
			objectstore: defaultObjectStoreMock,
			streammanager: func(t *testing.T) streammanager.StreamManager {
				c := gomock.NewController(t)
				m := streammanagermock.NewMockStreamManager(c)
				m.EXPECT().Put(
					subscription.SubscriptionAdded{
						Stream:       testChain0.ToObject().Hash(),
						Subscription: testSubscription0.ToObject().Hash(),
						Owners:       testOwners,
					}.ToObject(),
				).Return(
					errors.New("something bad"),
				)
				return m
			},
		},
		args: args{
			ctx: context.Background(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(
				context.Background(),
				tt.fields.keychain(t),
				tt.fields.resolver(t),
				tt.fields.exchange(t),
				tt.fields.objectstore(t),
				tt.fields.streammanager(t),
			)
			require.NoError(t, err)
			if err := m.Subscribe(
				tt.args.ctx,
				tt.args.subjects,
				tt.args.types,
				tt.args.streams,
				tt.args.expiry,
			); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_subscriptionmanager_GetOwnSubscriptions(t *testing.T) {
	testOwners := []crypto.PublicKey{
		"me",
		"also_me",
	}
	testChain0 := subscription.SubscriptionStreamRoot{
		Owners: testOwners,
	}
	testSubscription0 := subscription.Subscription{
		Subjects: []crypto.PublicKey{
			crypto.PublicKey("foo"),
		},
		Types: []string{
			"foo/bar/0",
		},
		Expiry: time.
			Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).
			Format(time.RFC3339),
		Owners: testOwners,
	}
	testSubscription0Added := subscription.SubscriptionAdded{
		Owners:       testOwners,
		Subscription: testSubscription0.ToObject().Hash(),
	}
	testSubscription1 := subscription.Subscription{
		Subjects: []crypto.PublicKey{
			crypto.PublicKey("foo"),
		},
		Types: []string{
			"foo/bar/1",
		},
		Expiry: time.
			Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).
			Format(time.RFC3339),
		Owners: testOwners,
	}
	testSubscription1Added := subscription.SubscriptionAdded{
		Owners:       testOwners,
		Subscription: testSubscription1.ToObject().Hash(),
	}
	testSubscription1Removed := subscription.SubscriptionRemoved{
		Owners:       testOwners,
		Subscription: testSubscription1.ToObject().Hash(),
	}
	testSubscriptionAddedInvalid := new(object.Object).
		SetType(
			new(subscription.SubscriptionAdded).GetType(),
		)
	testSubscriptionRemovedInvalid := new(object.Object).
		SetType(
			new(subscription.SubscriptionRemoved).GetType(),
		)
	testSubscriptionInvalid := new(object.Object).
		SetType(
			new(subscription.Subscription).GetType(),
		)
	testSubscriptionErrAdded := subscription.SubscriptionAdded{
		Owners:       testOwners,
		Subscription: testSubscriptionInvalid.ToObject().Hash(),
	}
	testSubscriptionMisAdded := subscription.SubscriptionAdded{
		Owners:       testOwners,
		Subscription: object.Hash("missing"),
	}
	testSubscription2 := subscription.Subscription{
		Subjects: []crypto.PublicKey{
			crypto.PublicKey("foo"),
		},
		Types: []string{
			"foo/bar/2",
		},
		Expiry: time.
			Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).
			Format(time.RFC3339),
		Owners: testOwners,
	}
	testSubscription2Added := subscription.SubscriptionAdded{
		Owners:       testOwners,
		Subscription: testSubscription2.ToObject().Hash(),
	}
	defaultKeychainMock := func(t *testing.T) keychain.Keychain {
		c := gomock.NewController(t)
		m := keychainmock.NewMockKeychain(c)
		m.EXPECT().ListPublicKeys(
			keychain.IdentityKey,
		).AnyTimes().Return(
			testOwners,
		)
		return m
	}
	defaultResolverMock := func(t *testing.T) resolver.Resolver {
		return nil
	}
	defaultObjectStoreMock := func(t *testing.T) objectstore.Store {
		c := gomock.NewController(t)
		m := objectstoremock.NewMockStore(c)
		m.EXPECT().
			Get(testSubscription0.ToObject().Hash()).
			Return(testSubscription0.ToObject(), nil)
		m.EXPECT().
			Get(testSubscription2.ToObject().Hash()).
			Return(testSubscription2.ToObject(), nil)
		m.EXPECT().
			Get(testSubscriptionInvalid.ToObject().Hash()).
			Return(testSubscriptionInvalid.ToObject(), nil)
		m.EXPECT().
			Get(object.Hash("missing")).
			Return(object.Object{}, errors.New("missing"))
		return m
	}
	defaultExchangeMock := func(t *testing.T) exchange.Exchange {
		return nil
	}
	defaultStreamManagerMock := func(t *testing.T) streammanager.StreamManager {
		c := gomock.NewController(t)
		m := streammanagermock.NewMockStreamManager(c)
		m.EXPECT().Get(
			gomock.Any(),
			testChain0.ToObject().Hash(),
		).Return(
			&streammanager.Graph{
				Objects: []object.Object{
					testChain0.ToObject(),
					testSubscription0Added.ToObject(),
					testSubscriptionAddedInvalid.ToObject(),
					testSubscription1Added.ToObject(),
					testSubscriptionRemovedInvalid.ToObject(),
					testSubscription1Removed.ToObject(),
					testSubscription2Added.ToObject(),
					testSubscriptionMisAdded.ToObject(),
					testSubscriptionErrAdded.ToObject(),
				},
			},
			nil,
		)
		return m
	}
	type fields struct {
		keychain      func(*testing.T) keychain.Keychain
		resolver      func(*testing.T) resolver.Resolver
		exchange      func(*testing.T) exchange.Exchange
		objectstore   func(*testing.T) objectstore.Store
		streammanager func(*testing.T) streammanager.StreamManager
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []subscription.Subscription
		wantErr bool
	}{{
		name: "should pass, got subs",
		fields: fields{
			keychain:      defaultKeychainMock,
			resolver:      defaultResolverMock,
			exchange:      defaultExchangeMock,
			objectstore:   defaultObjectStoreMock,
			streammanager: defaultStreamManagerMock,
		},
		want: []subscription.Subscription{
			testSubscription0,
			testSubscription2,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &subscriptionmanager{
				keychain:      tt.fields.keychain(t),
				resolver:      tt.fields.resolver(t),
				exchange:      tt.fields.exchange(t),
				objectstore:   tt.fields.objectstore(t),
				streammanager: tt.fields.streammanager(t),
			}
			got, err := m.GetOwnSubscriptions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, len(tt.want), len(got))
			for i, v := range tt.want {
				assert.Equal(t, v.ToObject().ToMap(), got[i].ToObject().ToMap())
			}
		})
	}
}

func Test_subscriptionmanager_GetSubscriptions(t *testing.T) {
	testOwners := []crypto.PublicKey{
		"me",
		"also_me",
	}
	testSubscription0 := subscription.Subscription{
		Subjects: []crypto.PublicKey{
			crypto.PublicKey("foo0"),
		},
		Types: []string{
			"foo/bar/0",
		},
		Expiry: time.
			Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).
			Format(time.RFC3339),
		Owners: testOwners,
	}
	testSubscription1 := subscription.Subscription{
		Subjects: []crypto.PublicKey{
			crypto.PublicKey("foo1"),
		},
		Types: []string{
			"foo/bar/1",
		},
		Expiry: time.
			Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).
			Format(time.RFC3339),
		Owners: testOwners,
	}
	testSubscriptionInvalid := new(object.Object).
		SetType(
			new(subscription.Subscription).GetType(),
		)
	defaultObjectStoreMock := func(t *testing.T) objectstore.Store {
		c := gomock.NewController(t)
		m := objectstoremock.NewMockStore(c)
		m.EXPECT().
			GetByType("nimona.io/subscription.Subscription").
			Return(
				[]object.Object{
					testSubscription0.ToObject(),
					testSubscription1.ToObject(),
					testSubscriptionInvalid.ToObject(),
				},
				nil,
			)
		return m
	}
	type fields struct {
		objectstore func(*testing.T) objectstore.Store
	}
	type args struct {
		ctx        context.Context
		objectType string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []subscription.Subscription
		wantErr bool
	}{{
		name: "should pass, found subs",
		fields: fields{
			objectstore: defaultObjectStoreMock,
		},
		args: args{
			ctx:        context.New(),
			objectType: "foo/bar/1",
		},
		want: []subscription.Subscription{
			testSubscription1,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &subscriptionmanager{
				objectstore: tt.fields.objectstore(t),
			}
			got, err := m.GetSubscriptionsByType(
				tt.args.ctx,
				tt.args.objectType,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, len(tt.want), len(got))
			for i, v := range tt.want {
				assert.Equal(t, v.ToObject().ToMap(), got[i].ToObject().ToMap())
			}
		})
	}
}
