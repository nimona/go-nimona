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
	"nimona.io/pkg/feed"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/keychainmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectmanagermock"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstoremock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/subscription"
)

func Test_subscriptionmanager_Subscribe(t *testing.T) {
	testOwners := []crypto.PublicKey{
		"me",
		"also_me",
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
		Metadata: object.Metadata{
			Owners: testOwners,
		},
	}
	testPeer0 := &peer.Peer{
		Metadata: object.Metadata{
			Owners: []crypto.PublicKey{
				"foo",
			},
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
	defaultObjectStoreMock := func(t *testing.T) objectstore.Store {
		c := gomock.NewController(t)
		m := objectstoremock.NewMockStore(c)
		m.EXPECT().Put(
			feed.GetFeedHypotheticalRoot(
				testOwners,
				"nimona.io/subscription.Subscription",
			).ToObject(),
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
	defaultObjectManagerMock := func(t *testing.T) objectmanager.ObjectManager {
		c := gomock.NewController(t)
		m := objectmanagermock.NewMockObjectManager(c)
		m.EXPECT().Put(
			gomock.Any(),
			testSubscription0.ToObject(),
		)
		return m
	}

	type fields struct {
		keychain      func(*testing.T) keychain.Keychain
		exchange      func(*testing.T) exchange.Exchange
		objectstore   func(*testing.T) objectstore.Store
		objectmanager func(*testing.T) objectmanager.ObjectManager
	}
	type args struct {
		ctx       context.Context
		subjects  []crypto.PublicKey
		types     []string
		streams   []object.Hash
		expiry    time.Time
		recipient *peer.Peer
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
			exchange:      defaultExchangeMock,
			objectstore:   defaultObjectStoreMock,
			objectmanager: defaultObjectManagerMock,
		},
		args: args{
			ctx: context.New(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			recipient: testPeer0,
		},
	}, {
		name: "should fail, could not write subscription",
		fields: fields{
			keychain: defaultKeychainMock,
			exchange: func(t *testing.T) exchange.Exchange {
				return nil
			},
			objectstore: defaultObjectStoreMock,
			objectmanager: func(t *testing.T) objectmanager.ObjectManager {
				c := gomock.NewController(t)
				m := objectmanagermock.NewMockObjectManager(c)
				m.EXPECT().Put(
					gomock.Any(),
					testSubscription0.ToObject(),
				).Return(
					object.Object{},
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
			expiry:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			recipient: testPeer0,
		},
		wantErr: true,
	}, {
		name: "should fail, failed to send to recipient",
		fields: fields{
			keychain: defaultKeychainMock,
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
			objectstore:   defaultObjectStoreMock,
			objectmanager: defaultObjectManagerMock,
		},
		args: args{
			ctx: context.Background(),
			subjects: []crypto.PublicKey{
				crypto.PublicKey("foo"),
			},
			types: []string{
				"foo/bar",
			},
			expiry:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			recipient: testPeer0,
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := New(
				context.Background(),
				tt.fields.keychain(t),
				tt.fields.exchange(t),
				tt.fields.objectstore(t),
				tt.fields.objectmanager(t),
			)
			require.NoError(t, err)
			if err := m.Subscribe(
				tt.args.ctx,
				tt.args.subjects,
				tt.args.types,
				tt.args.streams,
				tt.args.expiry,
				tt.args.recipient,
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
		Metadata: object.Metadata{
			Owners: testOwners,
		},
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
		Metadata: object.Metadata{
			Owners: testOwners,
		},
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
		Metadata: object.Metadata{
			Owners: testOwners,
		},
	}
	testFeedSubscriptionReader1 := object.NewReadCloserFromObjects(
		[]object.Object{
			feed.Added{
				Metadata: object.Metadata{
					Owners: testOwners,
				},
				ObjectHash: []object.Hash{
					testSubscription0.ToObject().Hash(),
				},
			}.ToObject(),
			feed.Added{
				Metadata: object.Metadata{
					Owners: testOwners,
				},
				ObjectHash: []object.Hash{
					testSubscription1.ToObject().Hash(),
				},
			}.ToObject(),
			feed.Added{
				Metadata: object.Metadata{
					Owners: testOwners,
				},
				ObjectHash: []object.Hash{
					testSubscription2.ToObject().Hash(),
				},
			}.ToObject(),
			feed.Removed{
				Metadata: object.Metadata{
					Owners: testOwners,
				},
				ObjectHash: []object.Hash{
					testSubscription1.ToObject().Hash(),
				},
			}.ToObject(),
		},
	)
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
	defaultObjectStoreMock := func(t *testing.T) objectstore.Store {
		c := gomock.NewController(t)
		m := objectstoremock.NewMockStore(c)
		m.EXPECT().
			Get(testSubscription0.ToObject().Hash()).
			Return(testSubscription0.ToObject(), nil)
		m.EXPECT().
			Get(testSubscription2.ToObject().Hash()).
			Return(testSubscription2.ToObject(), nil)
		return m
	}
	defaultExchangeMock := func(t *testing.T) exchange.Exchange {
		return nil
	}
	defaultObjectManagerMock := func(t *testing.T) objectmanager.ObjectManager {
		c := gomock.NewController(t)
		m := objectmanagermock.NewMockObjectManager(c)
		m.EXPECT().RequestStream(
			gomock.Any(),
			feed.GetFeedHypotheticalRootHash(
				testOwners,
				subscriptionType,
			),
		).Return(
			testFeedSubscriptionReader1,
			nil,
		)
		return m
	}
	type fields struct {
		keychain      func(*testing.T) keychain.Keychain
		exchange      func(*testing.T) exchange.Exchange
		objectstore   func(*testing.T) objectstore.Store
		objectmanager func(*testing.T) objectmanager.ObjectManager
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
			exchange:      defaultExchangeMock,
			objectstore:   defaultObjectStoreMock,
			objectmanager: defaultObjectManagerMock,
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
				exchange:      tt.fields.exchange(t),
				objectstore:   tt.fields.objectstore(t),
				objectmanager: tt.fields.objectmanager(t),
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
		Metadata: object.Metadata{
			Owners: testOwners,
		},
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
		Metadata: object.Metadata{
			Owners: testOwners,
		},
	}
	testFeedSubscriptionReader1 := object.NewReadCloserFromObjects(
		[]object.Object{
			feed.Added{
				Metadata: object.Metadata{
					Owners: testOwners,
				},
				ObjectHash: []object.Hash{
					testSubscription0.ToObject().Hash(),
				},
			}.ToObject(),
			feed.Added{
				Metadata: object.Metadata{
					Owners: testOwners,
				},
				ObjectHash: []object.Hash{
					testSubscription1.ToObject().Hash(),
				},
			}.ToObject(),
		},
	)
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
	defaultObjectStoreMock := func(t *testing.T) objectstore.Store {
		c := gomock.NewController(t)
		m := objectstoremock.NewMockStore(c)
		m.EXPECT().
			Get(testSubscription0.ToObject().Hash()).
			Return(testSubscription0.ToObject(), nil)
		m.EXPECT().
			Get(testSubscription1.ToObject().Hash()).
			Return(testSubscription1.ToObject(), nil)
		return m
	}
	type fields struct {
		keychain      func(*testing.T) keychain.Keychain
		objectstore   func(*testing.T) objectstore.Store
		objectmanager func(*testing.T) objectmanager.ObjectManager
	}
	type args struct {
		ctx        context.Context
		objectType string
	}
	defaultObjectManagerMock := func(t *testing.T) objectmanager.ObjectManager {
		c := gomock.NewController(t)
		m := objectmanagermock.NewMockObjectManager(c)
		m.EXPECT().RequestStream(
			gomock.Any(),
			feed.GetFeedHypotheticalRootHash(
				testOwners,
				subscriptionType,
			),
		).Return(
			testFeedSubscriptionReader1,
			nil,
		)
		return m
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
			objectstore:   defaultObjectStoreMock,
			objectmanager: defaultObjectManagerMock,
			keychain:      defaultKeychainMock,
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
				keychain:      tt.fields.keychain(t),
				objectmanager: tt.fields.objectmanager(t),
				objectstore:   tt.fields.objectstore(t),
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
