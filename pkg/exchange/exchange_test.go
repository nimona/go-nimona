package exchange

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/mocks"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

func TestSendSuccess(t *testing.T) {
	// create the stores
	dblite1 := tempSqlite3(t)
	store1, err := sqlobjectstore.New(dblite1)
	assert.NoError(t, err)

	dblite2 := tempSqlite3(t)
	store2, err := sqlobjectstore.New(dblite2)
	assert.NoError(t, err)

	// attach the stores to discovery
	disc1 := discovery.NewPeerStorer(store1)
	disc2 := discovery.NewPeerStorer(store2)

	// create new peers
	kc1, n1, x1, _ := newPeer(t, disc1)
	kc2, n2, x2, _ := newPeer(t, disc2)

	// make peers aware of each other
	disc1.Add(&peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
	}, true)
	disc2.Add(&peer.Peer{
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
	}, true)

	// peer1 looks for peer2
	dr1, err := disc1.Lookup(
		context.Background(),
		peer.LookupByOwner(kc2.GetPrimaryPeerKey().PublicKey()),
	)
	require.NoError(t, err)

	gp := gatherPeers(dr1)
	require.Len(t, gp, 1)
	require.Equal(t, kc2.GetPrimaryPeerKey().PublicKey(), gp[0].PublicKey())
	require.Equal(t, n2.Addresses(), gp[0].Addresses)

	// create test objects
	eo1 := object.Object{}
	eo1 = eo1.Set("body:s", "bar1")
	eo1 = eo1.SetType("test/msg")

	eo2 := object.Object{}
	eo2 = eo2.Set("body:s", "bar1")
	eo2 = eo2.SetType("test/msg")

	wg := sync.WaitGroup{}
	wg.Add(2)

	handled := int32(0)

	sig, err := object.NewSignature(kc2.GetPrimaryPeerKey(), eo1)
	assert.NoError(t, err)
	eo1 = eo1.AddSignature(sig)

	sig, err = object.NewSignature(kc1.GetPrimaryPeerKey(), eo2)
	assert.NoError(t, err)
	eo2 = eo2.AddSignature(sig)

	// add message handlers
	// nolint: dupl
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("test/msg"),
		),
		func(e *Envelope) error {
			o := e.Payload
			assert.Equal(t, eo1.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	// nolint: dupl
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("tes**"),
		),
		func(e *Envelope) error {
			o := e.Payload
			assert.Equal(t, eo2.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	ctx := context.Background()

	errS1 := x2.Send(
		ctx,
		eo1,
		peer.LookupByOwner(kc1.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, errS1)

	time.Sleep(time.Second)

	errS2 := x1.Send(
		ctx,
		eo2,
		peer.LookupByOwner(kc2.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, errS2)

	if errS1 == nil && errS2 == nil {
		wg.Wait()
	}

	assert.Equal(t, int32(2), atomic.LoadInt32(&handled))
}

func TestRequestSuccess(t *testing.T) {
	dblite1 := tempSqlite3(t)
	store1, err := sqlobjectstore.New(dblite1)
	assert.NoError(t, err)

	dblite2 := tempSqlite3(t)
	store2, err := sqlobjectstore.New(dblite2)
	assert.NoError(t, err)

	disc1 := discovery.NewPeerStorer(store1)
	disc2 := discovery.NewPeerStorer(store2)

	kc1, n1, x1, _ := newPeer(t, disc1)
	kc2, n2, _, d2 := newPeer(t, disc2)

	disc1.Add(&peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
	}, true)
	disc2.Add(&peer.Peer{
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
	}, true)

	mp2 := &mocks.Discoverer{}
	err = disc2.AddDiscoverer(mp2)
	assert.NoError(t, err)

	// add an object to n2's store
	eo1 := object.Object{}
	eo1 = eo1.Set("body:s", "bar1")
	eo1 = eo1.SetType("test/msg")
	err = d2.Put(eo1)
	assert.NoError(t, err)

	// handle events in x1 to make sure we received responses
	out := make(chan *Envelope, 1)
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("test/msg"),
		),
		func(e *Envelope) error {
			out <- e
			return nil
		},
	)

	// request object, with req id
	ctx := context.New(context.WithTimeout(time.Second * 3))
	err = x1.Request(
		ctx,
		object.NewHash(eo1),
		peer.LookupByOwner(kc2.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, err)

	// check if we got back the expected obj
	select {
	case <-ctx.Done():
		t.Log("did not receive response in time")
		t.FailNow()
	case o1r := <-out:
		compareObjects(t, eo1, o1r.Payload)
	}
}

func TestSendRelay(t *testing.T) {
	dblite1 := tempSqlite3(t)
	store1, err := sqlobjectstore.New(dblite1)
	assert.NoError(t, err)

	dblite2 := tempSqlite3(t)
	store2, err := sqlobjectstore.New(dblite2)
	assert.NoError(t, err)

	dblite3 := tempSqlite3(t)
	store3, err := sqlobjectstore.New(dblite3)
	assert.NoError(t, err)

	discRel := discovery.NewPeerStorer(store1)
	disc1 := discovery.NewPeerStorer(store2)
	disc2 := discovery.NewPeerStorer(store3)

	// enable binding to local addresses
	net.BindLocal = true

	// relay peer
	rkc, rn, _, _ := newPeer(t, discRel)

	// disable binding to local addresses
	net.BindLocal = false
	kc1, n1, x1, _ := newPeer(t, disc1)
	kc2, n2, x2, _ := newPeer(t, disc2)

	discRel.Add(&peer.Peer{
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
		Relays: []crypto.PublicKey{
			rkc.GetPrimaryPeerKey().PublicKey(),
		},
	}, false)
	discRel.Add(&peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
		Relays: []crypto.PublicKey{
			rkc.GetPrimaryPeerKey().PublicKey(),
		},
	}, false)
	disc1.Add(&peer.Peer{
		Owners:    rkc.ListPublicKeys(keychain.PeerKey),
		Addresses: rn.Addresses(),
	}, false)
	disc1.Add(&peer.Peer{
		Owners:    kc2.ListPublicKeys(keychain.PeerKey),
		Addresses: n2.Addresses(),
		Relays: []crypto.PublicKey{
			rkc.GetPrimaryPeerKey().PublicKey(),
		},
	}, false)
	disc2.Add(&peer.Peer{
		Owners:    rkc.ListPublicKeys(keychain.PeerKey),
		Addresses: rn.Addresses(),
	}, false)
	disc2.Add(&peer.Peer{
		Owners:    kc1.ListPublicKeys(keychain.PeerKey),
		Addresses: n1.Addresses(),
		Relays: []crypto.PublicKey{
			rkc.GetPrimaryPeerKey().PublicKey(),
		},
	}, false)

	// init connection from peer1 to relay
	o1 := object.Object{}
	o1 = o1.SetType("foo")
	o1 = o1.Set("foo:s", "bar")
	err = x1.Send(
		context.Background(),
		o1,
		peer.LookupByOwner(rkc.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, err)

	// init connection from peer2 to relay
	o2 := object.Object{}
	o2 = o2.SetType("foo")
	o2 = o2.Set("foo:s", "bar")
	err = x2.Send(
		context.Background(),
		o2,
		peer.LookupByOwner(rkc.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, err)

	// create the messages
	eo1 := object.Object{}
	eo1 = eo1.Set("body:s", "bar1")
	eo1 = eo1.SetType("test/msg")

	eo2 := object.Object{}
	eo2 = eo2.Set("body:s", "bar1")
	eo2 = eo2.SetType("test/msg")

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1ObjectHandled := false
	w2ObjectHandled := false

	sig, err := object.NewSignature(kc2.GetPrimaryPeerKey(), eo1)
	assert.NoError(t, err)
	eo1 = eo1.AddSignature(sig)

	handled := int32(0)

	// add handlers
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("test/msg"),
		),
		func(e *Envelope) error {
			o := e.Payload
			w1ObjectHandled = true
			assert.Equal(t, eo1.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)

	// nolint: dupl
	go HandleEnvelopeSubscription(
		x1.Subscribe(
			FilterByObjectType("tes**"),
		),
		func(e *Envelope) error {
			o := e.Payload
			w2ObjectHandled = true
			assert.Equal(t, eo2.Get("body:s"), o.Get("body:s"))
			atomic.AddInt32(&handled, 1)
			wg.Done()
			return nil
		},
	)
	assert.NoError(t, err)

	ctx := context.New(context.WithTimeout(time.Second * 5))
	defer ctx.Cancel()

	err = x2.Send(
		ctx,
		eo1,
		peer.LookupByOwner(kc1.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctx2 := context.New(context.WithTimeout(time.Second * 5))
	defer ctx2.Cancel()

	// TODO should be able to send not signed
	err = x1.Send(
		ctx2,
		eo2,
		peer.LookupByOwner(kc2.GetPrimaryPeerKey().PublicKey()),
	)
	assert.NoError(t, err)

	wg.Wait()

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func newPeer(
	t *testing.T,
	discover discovery.PeerStorer,
) (
	keychain.Keychain,
	net.Network,
	*exchange,
	*sqlobjectstore.Store,
) {
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	eb := eventbus.New()

	kc := keychain.New()
	kc.Put(keychain.PrimaryPeerKey, pk)

	ds, err := sqlobjectstore.New(tempSqlite3(t))
	assert.NoError(t, err)

	ctx := context.Background()

	n := net.New(
		net.WithKeychain(kc),
	)
	_, err = n.Listen(ctx, "127.0.0.1:0")
	require.NoError(t, err)

	x, err := New(ctx, eb, kc, n, ds, discover)
	assert.NoError(t, err)

	return kc, n, x.(*exchange), ds
}

func compareObjects(t *testing.T, expected, actual object.Object) {
	assert.Equal(t, jp(expected), jp(actual))
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}

func tempSqlite3(t *testing.T) *sql.DB {
	dirPath, err := ioutil.TempDir("", "nimona-store-sql")
	require.NoError(t, err)
	fmt.Println(path.Join(dirPath, "sqlite3.db"))
	db, err := sql.Open("sqlite3", path.Join(dirPath, "sqlite3.db"))
	require.NoError(t, err)
	return db
}

func gatherPeers(p <-chan *peer.Peer) []*peer.Peer {
	ps := []*peer.Peer{}
	for p := range p {
		p := p
		ps = append(ps, p)
	}
	return peer.Unique(ps)
}
