package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nimona.io/internal/store/graph"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/mocks"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
)

func TestSendSuccess(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	k1, _, x1, _, l1 := newPeer(t, "", disc1, true, false)
	k2, _, x2, _, l2 := newPeer(t, "", disc2, true, false)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)

	em2 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo2 := object.FromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1ObjectHandled := false
	w2ObjectHandled := false

	err := crypto.Sign(eo1, k2)
	assert.NoError(t, err)

	// nolint: dupl
	_, err = x1.Handle("test/msg", func(e *Envelope) error {
		o := e.Payload
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		w1ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	_, err = x2.Handle("tes**", func(e *Envelope) error {
		o := e.Payload
		assert.Equal(t, eo2.GetRaw("body"), o.GetRaw("body"))
		w2ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	ctx := context.Background()

	errS1 := x2.Send(ctx, eo1, "peer:"+k1.PublicKey.Fingerprint().String())
	assert.NoError(t, errS1)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	errS2 := x1.Send(ctx, eo2, "peer:"+k2.PublicKey.Fingerprint().String())
	assert.NoError(t, errS2)

	if errS1 == nil && errS2 == nil {
		wg.Wait()
	}

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func TestSendWithResponseSuccess(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	k1, _, x1, _, l1 := newPeer(t, "", disc1, true, false)
	k2, _, x2, _, l2 := newPeer(t, "", disc2, true, false)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	mp2 := &mocks.Provider{}
	err := disc2.AddProvider(mp2)
	assert.NoError(t, err)

	// send object with request id
	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)
	err = crypto.Sign(eo1, k1)
	assert.NoError(t, err)

	out := make(chan *Envelope, 1)

	ctx, _ := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*3),
	)
	err = x2.Send(
		ctx,
		eo1,
		l1.GetPeerInfo().Addresses[0],
		WithResponse("foo", out),
	)
	assert.NoError(t, err)

	// send object in response with the same request id

	em2 := map[string]interface{}{
		"@ctx":          "test/msg",
		"body":          "bar2",
		ObjectRequestID: "foo",
	}
	eo2 := object.FromMap(em2)
	err = crypto.Sign(eo2, k2)
	assert.NoError(t, err)

	err = x1.Send(
		ctx,
		eo2,
		l2.GetPeerInfo().Addresses[0],
	)
	assert.NoError(t, err)

	select {
	case <-ctx.Done():
		t.Log("did not receive response in time")
		t.FailNow()
	case o2r := <-out:
		compareObjects(t, eo2, o2r.Payload)
	}
}

func TestSendWithResponseSuccessHTTP(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	k1, _, x1, _, l1 := newPeer(t, "", disc1, false, true)
	k2, _, x2, _, l2 := newPeer(t, "", disc2, false, true)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	mp2 := &mocks.Provider{}
	err := disc2.AddProvider(mp2)
	assert.NoError(t, err)

	// send object with request id
	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)
	err = crypto.Sign(eo1, k1)
	assert.NoError(t, err)

	out := make(chan *Envelope, 1)

	ctx, _ := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*3),
	)
	err = x2.Send(
		ctx,
		eo1,
		l1.GetPeerInfo().Addresses[0],
		WithResponse("foo", out),
	)
	assert.NoError(t, err)

	// send object in response with the same request id

	em2 := map[string]interface{}{
		"@ctx":          "test/msg",
		"body":          "bar2",
		ObjectRequestID: "foo",
	}
	eo2 := object.FromMap(em2)
	err = crypto.Sign(eo2, k2)
	assert.NoError(t, err)

	err = x1.Send(
		ctx,
		eo2,
		l2.GetPeerInfo().Addresses[0],
	)
	assert.NoError(t, err)

	select {
	case <-ctx.Done():
		t.Log("did not receive response in time")
		t.FailNow()
	case o2r := <-out:
		compareObjects(t, eo2, o2r.Payload)
	}
}

func TestRequestSuccess(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	_, _, x1, _, l1 := newPeer(t, "", disc1, true, false)
	_, _, _, d2, l2 := newPeer(t, "", disc2, true, false)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	mp2 := &mocks.Provider{}
	err := disc2.AddProvider(mp2)
	assert.NoError(t, err)

	// add an object to n2's store
	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)
	err = d2.Put(eo1)
	assert.NoError(t, err)

	// request object, with req id
	out := make(chan *Envelope, 1)
	ctx, _ := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*3),
	)
	err = x1.Request(
		ctx,
		eo1.HashBase58(),
		l2.GetPeerInfo().Addresses[0],
		WithResponse("foo", out),
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

func TestRequestSuccessHTTP(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	_, _, x1, _, l1 := newPeer(t, "", disc1, false, true)
	_, _, _, d2, l2 := newPeer(t, "", disc2, false, true)

	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l1.GetPeerInfo())

	mp2 := &mocks.Provider{}
	err := disc2.AddProvider(mp2)
	assert.NoError(t, err)

	// add an object to n2's store
	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)
	err = d2.Put(eo1)
	assert.NoError(t, err)

	// request object, with req id
	out := make(chan *Envelope, 1)
	ctx, _ := context.WithDeadline(
		context.Background(),
		time.Now().Add(time.Second*3),
	)
	err = x1.Request(
		ctx,
		eo1.HashBase58(),
		l2.GetPeerInfo().Addresses[0],
		WithResponse("foo", out),
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
	// enable binding to local addresses
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()
	disc3 := discovery.NewDiscoverer()

	net.BindLocal = true
	k0, _, _, _, l0 := newPeer(t, "", disc1, true, false)

	// disable binding to local addresses
	net.BindLocal = false
	k1, _, x1, _, l1 := newPeer(t, "relay:"+l0.GetPeerInfo().Addresses[0], disc2, true, false)
	k2, _, x2, _, l2 := newPeer(t, "relay:"+l0.GetPeerInfo().Addresses[0], disc3, true, false)

	fmt.Printf("\n\n\n\n-----------------------------\n")
	fmt.Println("k0:",
		k0.PublicKey.Fingerprint(),
		l0.GetPeerInfo().Addresses,
	)
	fmt.Println("k1:",
		k1.PublicKey.Fingerprint(),
		l1.GetPeerInfo().Addresses,
	)
	fmt.Println("k2:",
		k2.PublicKey.Fingerprint(),
		l2.GetPeerInfo().Addresses,
	)
	fmt.Printf("-----------------------------\n\n\n\n")

	disc1.Add(l1.GetPeerInfo())
	disc1.Add(l2.GetPeerInfo())
	disc2.Add(l2.GetPeerInfo())
	disc3.Add(l1.GetPeerInfo())

	// init connection from n1 to n0
	err := x1.Send(
		context.Background(),
		object.FromMap(map[string]interface{}{"foo": "bar"}),
		l0.GetPeerInfo().Addresses[0],
	)
	assert.NoError(t, err)

	// init connection from n2 to n0
	err = x2.Send(
		context.Background(),
		object.FromMap(map[string]interface{}{"foo": "bar"}),
		l0.GetPeerInfo().Addresses[0],
	)
	assert.NoError(t, err)

	// now we should be able to relay objects between n1 and n2
	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)

	em2 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo2 := object.FromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1ObjectHandled := false
	w2ObjectHandled := false

	err = crypto.Sign(eo1, k2)
	assert.NoError(t, err)

	// nolint: dupl
	_, err = x1.Handle("test/msg", func(e *Envelope) error {
		o := e.Payload
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		w1ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	_, err = x2.Handle("tes**", func(e *Envelope) error {
		o := e.Payload
		assert.Equal(t, eo2.GetRaw("body"), o.GetRaw("body"))
		w2ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
	defer cf()

	err = x2.Send(ctx, eo1, "peer:"+k1.PublicKey.Fingerprint().String())
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctx2, cf2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cf2()

	// TODO should be able to send not signed
	err = x1.Send(ctx2, eo2, "peer:"+k2.PublicKey.Fingerprint().String())
	assert.NoError(t, err)

	wg.Wait()

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func newPeer(
	t *testing.T,
	relayAddress string,
	discover discovery.Discoverer,
	listenTCP bool,
	listenHTTP bool,
) (
	*crypto.PrivateKey,
	net.Network,
	*exchange,
	graph.Store,
	*peer.Peer,
) {
	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	ds, err := graph.NewCayleyWithTempStore()
	assert.NoError(t, err)

	li, err := peer.NewPeer("", pk)
	assert.NoError(t, err)

	if relayAddress != "" {
		li.AddAddress(relayAddress)
	}

	n, err := net.New(discover, li)
	assert.NoError(t, err)

	if listenTCP {
		tcp := net.NewTCPTransport(li, "0.0.0.0:0")
		n.AddTransport("tcps", tcp)
	}

	if listenHTTP {
		http := net.NewHTTPTransport(li, "0.0.0.0:0")
		n.AddTransport("https", http)
	}

	hsm := handshake.New(li, discover)
	n.AddMiddleware(hsm.Handle())

	ctx := context.Background()

	x, err := New(ctx, pk, n, ds, discover, li)
	assert.NoError(t, err)

	return pk, n, x.(*exchange), ds, li
}

func compareObjects(t *testing.T, expected, actual *object.Object) {
	for m := range expected.Members {
		assert.Equal(t, jp(expected.GetRaw(m)), jp(actual.GetRaw(m)))
	}
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}
