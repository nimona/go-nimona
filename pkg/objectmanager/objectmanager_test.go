package objectmanager

import (
	"database/sql"
	"io/ioutil"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"

	_ "github.com/mattn/go-sqlite3"
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
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
	}
	p2 := &peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
	}

	// create test objects
	obj := object.Object{}
	obj = obj.Set("body:s", "bar1")
	obj = obj.SetType("test/msg")

	// setup hander
	go exchange.HandleEnvelopeSubscription(
		x2.Subscribe(
			exchange.FilterByObjectType(new(Request).GetType()),
		),
		func(e *exchange.Envelope) error {
			o := e.Payload
			objectHandled = true

			objr := Request{}
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
	assert.Equal(t, obj, *objRecv)

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
	Requester,
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
