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
	"nimona.io/pkg/hash"
	"nimona.io/pkg/middleware/handshake"
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
	k1, _, x1, _, l1 := newPeer(t, disc1, true, false)
	k2, _, x2, _, l2 := newPeer(t, disc2, true, false)

	// make peers aware of each other
	disc1.Add(l2.GetSignedPeer(), true)
	disc2.Add(l1.GetSignedPeer(), true)

	// peer1 looks for peer2
	dr1, err := disc1.Lookup(
		context.Background(),
		peer.LookupByOwner(l2.GetPeerPublicKey()),
	)
	require.NoError(t, err)
	require.Len(t, dr1, 1)
	require.Equal(t, l2.GetIdentityPublicKey(), gatherPeers(dr1)[0].PublicKey())

	// create test objects
	em1 := map[string]interface{}{
		"@type:s": "test/msg",
		"body:s":  "bar1",
	}
	eo1 := object.FromMap(em1)

	em2 := map[string]interface{}{
		"@type:s": "test/msg",
		"body:s":  "bar1",
	}
	eo2 := object.FromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	handled := int32(0)

	err = crypto.Sign(eo1, k2)
	assert.NoError(t, err)

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

	errS1 := x2.Send(ctx, eo1, peer.LookupByOwner(k1.PublicKey()))
	assert.NoError(t, errS1)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	errS2 := x1.Send(ctx, eo2, peer.LookupByOwner(k2.PublicKey()))
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

	_, _, x1, _, l1 := newPeer(t, disc1, true, false)
	_, _, _, d2, l2 := newPeer(t, disc2, true, false)

	disc1.Add(l2.GetSignedPeer(), true)
	disc2.Add(l1.GetSignedPeer(), true)

	mp2 := &mocks.Discoverer{}
	err = disc2.AddDiscoverer(mp2)
	assert.NoError(t, err)

	// add an object to n2's store
	em1 := map[string]interface{}{
		"@type:s": "test/msg",
		"body:s":  "bar1",
	}
	eo1 := object.FromMap(em1)
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
		hash.New(eo1),
		peer.LookupByOwner(l2.GetPeerPublicKey()),
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
	relayPeerKey, _, _, _, relayPeer := newPeer(t, discRel, true, false)

	// disable binding to local addresses
	net.BindLocal = false
	k1, _, x1, _, l1 := newPeer(t, disc1, true, false)
	k2, _, x2, _, l2 := newPeer(t, disc2, true, false)

	l1.AddRelays(relayPeerKey.PublicKey())
	l2.AddRelays(relayPeerKey.PublicKey())

	fmt.Printf("\n\n\n\n-----------------------------\n")
	fmt.Println("k0:",
		relayPeerKey.PublicKey(),
		relayPeer.GetAddresses(),
	)
	fmt.Println("k1:",
		k1.PublicKey(),
		l1.GetAddresses(),
	)
	fmt.Println("k2:",
		k2.PublicKey(),
		l2.GetAddresses(),
	)
	fmt.Printf("-----------------------------\n\n\n\n")

	discRel.Add(l1.GetSignedPeer(), false)
	discRel.Add(l2.GetSignedPeer(), false)
	disc1.Add(relayPeer.GetSignedPeer(), false)
	disc1.Add(l2.GetSignedPeer(), false)
	disc2.Add(relayPeer.GetSignedPeer(), false)
	disc2.Add(l1.GetSignedPeer(), false)

	// init connection from peer1 to relay
	err = x1.Send(
		context.Background(),
		object.FromMap(map[string]interface{}{"foo": "bar"}),
		peer.LookupByOwner(relayPeer.GetPeerPublicKey()),
	)
	assert.NoError(t, err)

	// init connection from peer2 to relay
	err = x2.Send(
		context.Background(),
		object.FromMap(map[string]interface{}{"foo": "bar"}),
		peer.LookupByOwner(relayPeer.GetPeerPublicKey()),
	)
	assert.NoError(t, err)

	// create the messages
	em1 := map[string]interface{}{
		"@type:s": "test/msg",
		"body:s":  "bar1",
	}
	eo1 := object.FromMap(em1)

	em2 := map[string]interface{}{
		"@type:s": "test/msg",
		"body:s":  "bar1",
	}
	eo2 := object.FromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1ObjectHandled := false
	w2ObjectHandled := false

	err = crypto.Sign(eo1, k2)
	assert.NoError(t, err)

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

	err = x2.Send(ctx, eo1, peer.LookupByOwner(k1.PublicKey()))
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctx2 := context.New(context.WithTimeout(time.Second * 5))
	defer ctx2.Cancel()

	// TODO should be able to send not signed
	err = x1.Send(ctx2, eo2, peer.LookupByOwner(k2.PublicKey()))
	assert.NoError(t, err)

	wg.Wait()

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func newPeer(
	t *testing.T,
	discover discovery.PeerStorer,
	listenTCP bool,
	listenHTTP bool,
) (
	crypto.PrivateKey,
	net.Network,
	*exchange,
	*sqlobjectstore.Store,
	*peer.LocalPeer,
) {
	pk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	ds, err := sqlobjectstore.New(tempSqlite3(t))
	assert.NoError(t, err)

	li, err := peer.NewLocalPeer("", pk)
	assert.NoError(t, err)

	n, err := net.New(discover, li)
	assert.NoError(t, err)

	if listenTCP {
		tcp := net.NewTCPTransport(li, "0.0.0.0:0")
		n.AddTransport("tcps", tcp)
	}

	hsm := handshake.New(li, discover)
	n.AddMiddleware(hsm.Handle())

	ctx := context.Background()

	x, err := New(ctx, pk, n, ds, discover, li)
	assert.NoError(t, err)

	return pk, n, x.(*exchange), ds, li
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
