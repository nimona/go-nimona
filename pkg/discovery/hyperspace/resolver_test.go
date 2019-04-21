package hyperspace

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
	"nimona.io/pkg/middleware/handshake"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/exchange"
)

func TestDiscoverer(t *testing.T) {
	k0, n0, x0, disc0, l0 := newPeer(t)
	k1, n1, x1, disc1, l1 := newPeer(t)
	k2, n2, x2, disc2, l2 := newPeer(t)

	fmt.Println("k0:", k0.GetPublicKey().HashBase58(), l0.GetPeerInfo().Addresses)
	fmt.Println("k1:", k1.GetPublicKey().HashBase58(), l1.GetPeerInfo().Addresses)
	fmt.Println("k2:", k2.GetPublicKey().HashBase58(), l2.GetPeerInfo().Addresses)
	fmt.Printf("-----------------------------\n\n\n\n")

	d0, err := NewDiscoverer(n0, x0, l0, []string{})
	assert.NoError(t, err)
	err = disc0.AddProvider(d0)
	assert.NoError(t, err)

	ba := l0.GetPeerInfo().Addresses

	d1, err := NewDiscoverer(n1, x1, l1, ba)
	assert.NoError(t, err)
	err = disc1.AddProvider(d1)
	assert.NoError(t, err)

	d2, err := NewDiscoverer(n2, x2, l2, ba)
	assert.NoError(t, err)
	err = disc2.AddProvider(d2)
	assert.NoError(t, err)

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

	_, err = x1.Handle("test/msg", func(e *exchange.Envelope) error {
		o := e.Payload
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		assert.NotNil(t, eo1.GetSignerKey())
		assert.NotNil(t, o.GetSignerKey())
		assert.Equal(t, jp(eo1.GetSignerKey().ToMap()), jp(o.GetSignerKey().ToMap()))
		assert.Equal(t, eo1.GetSignerKey().HashBase58(), o.GetSignerKey().HashBase58())
		assert.NotNil(t, eo1.GetSignature())
		assert.NotNil(t, o.GetSignature())
		assert.Equal(t, jp(eo1.GetSignature().ToMap()), jp(o.GetSignature().ToMap()))
		assert.Equal(t, eo1.GetSignature().HashBase58(), o.GetSignature().HashBase58())
		w1ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	_, err = x2.Handle("tes**", func(e *exchange.Envelope) error {
		o := e.Payload
		assert.Equal(t, eo2.GetRaw("body"), o.GetRaw("body"))
		assert.Nil(t, eo2.GetSignature())
		assert.Nil(t, o.GetSignature())
		assert.Nil(t, eo2.GetSignerKey())
		assert.Nil(t, o.GetSignerKey())

		w2ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	ctx, cf := context.WithTimeout(context.Background(), time.Second*5)
	defer cf()

	err = x2.Send(ctx, eo1, "peer:"+k1.GetPublicKey().HashBase58())
	assert.NoError(t, err)

	time.Sleep(time.Second)

	ctx2, cf2 := context.WithTimeout(context.Background(), time.Second*5)
	defer cf2()

	// TODO should be able to send not signed
	err = x1.Send(ctx2, eo2, "peer:"+k2.GetPublicKey().HashBase58())
	assert.NoError(t, err)

	wg.Wait()

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func newPeer(t *testing.T) (*crypto.Key, net.Network, exchange.Exchange,
	discovery.Discoverer, *net.LocalInfo) {

	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	disc := discovery.NewDiscoverer()
	ds, err := graph.NewCayleyWithTempStore()
	local, err := net.NewLocalInfo("host", pk)
	assert.NoError(t, err)

	n, err := net.New(disc, local)
	assert.NoError(t, err)

	tcp := net.NewTCPTransport(local, []string{})
	hsm := handshake.New(local, disc)
	n.AddMiddleware(hsm.Handle())
	n.AddTransport("tcps", tcp)

	x, err := exchange.New(pk, n, ds, disc, local, fmt.Sprintf("0.0.0.0:%d", 0))
	assert.NoError(t, err)

	return pk, n, x, disc, local
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}
