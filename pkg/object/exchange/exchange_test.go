package exchange

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/mocks"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
	"nimona.io/pkg/storage"
)

func TestSendSuccess(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	k1, n1, x1 := newPeer(t, "", disc1)
	k2, n2, x2 := newPeer(t, "", disc2)

	disc1.Add(n2.GetPeerInfo())
	disc2.Add(n1.GetPeerInfo())

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

	_, err = x1.Handle("test/msg", func(o *object.Object) error {
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		assert.NotNil(t, eo1.GetSignerKey())
		assert.NotNil(t, o.GetSignerKey())
		assert.Equal(t, eo1.GetSignerKey(), o.GetSignerKey())
		assert.Equal(t, eo1.GetSignerKey().HashBase58(), o.GetSignerKey().HashBase58())
		assert.NotNil(t, eo1.GetSignature())
		assert.NotNil(t, o.GetSignature())
		assert.Equal(t, eo1.GetSignature(), o.GetSignature())
		assert.Equal(t, eo1.GetSignature().HashBase58(), o.GetSignature().HashBase58())
		w1ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	_, err = x2.Handle("tes**", func(o *object.Object) error {
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

	ctx := context.Background()

	errS1 := x2.Send(ctx, eo1, "peer:"+k1.GetPublicKey().HashBase58())
	assert.NoError(t, errS1)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	errS2 := x1.Send(ctx, eo2, "peer:"+k2.GetPublicKey().HashBase58())
	assert.NoError(t, errS2)

	if errS1 == nil && errS2 == nil {
		wg.Wait()
	}

	assert.True(t, w1ObjectHandled)
	assert.True(t, w2ObjectHandled)
}

func TestGetLocalSuccess(t *testing.T) {
	disc1 := discovery.NewDiscoverer()

	k1, _, x1 := newPeer(t, "", disc1)

	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)

	err := crypto.Sign(eo1, k1)
	assert.NoError(t, err)

	eo1b, _ := object.Marshal(eo1)
	err = x1.store.Store(eo1.HashBase58(), eo1b)
	assert.NoError(t, err)

	ctx := context.Background()

	o1, err := x1.Get(ctx, eo1.HashBase58())
	assert.NoError(t, err)
	compareObjects(t, eo1, o1)
}

func TestGetSuccess(t *testing.T) {
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()

	k1, n1, x1 := newPeer(t, "", disc1)
	_, n2, x2 := newPeer(t, "", disc2)

	disc1.Add(n2.GetPeerInfo())
	disc2.Add(n1.GetPeerInfo())

	mp2 := &mocks.Provider{}
	err := disc2.AddProvider(mp2)
	assert.NoError(t, err)

	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := object.FromMap(em1)

	err = crypto.Sign(eo1, k1)
	assert.NoError(t, err)

	eo1b, _ := object.Marshal(eo1)
	err = x1.store.Store(eo1.HashBase58(), eo1b)
	assert.NoError(t, err)

	mp2.On("Discover", &peer.PeerInfoRequest{
		ContentIDs: []string{
			eo1.HashBase58(),
		},
	}).Return([]*peer.PeerInfo{
		n1.GetPeerInfo(),
	}, nil)

	mp2.On("Discover", &peer.PeerInfoRequest{
		SignerKeyHash: n1.GetPeerInfo().HashBase58(),
	}).Return([]*peer.PeerInfo{
		n1.GetPeerInfo(),
	}, nil)

	p1 := n1.GetPeerInfo()
	p1.ContentIDs = []string{
		eo1.HashBase58(),
	}

	disc1.Add(n2.GetPeerInfo())
	disc1.Add(p1)

	ctx := context.Background()
	o1, err := x2.Get(ctx, eo1.HashBase58())
	assert.NoError(t, err)
	compareObjects(t, eo1, o1)
}

func TestSendRelay(t *testing.T) {
	// enable binding to local addresses
	disc1 := discovery.NewDiscoverer()
	disc2 := discovery.NewDiscoverer()
	disc3 := discovery.NewDiscoverer()

	net.BindLocal = true
	k0, n0, _ := newPeer(t, "", disc1)

	// disable binding to local addresses
	net.BindLocal = false
	k1, n1, x1 := newPeer(t, "relay:"+n0.GetPeerInfo().Addresses[0], disc2)
	k2, n2, x2 := newPeer(t, "relay:"+n0.GetPeerInfo().Addresses[0], disc3)

	fmt.Printf("\n\n\n\n-----------------------------\n")
	fmt.Println("k0:", k0.GetPublicKey().HashBase58(), n0.GetPeerInfo().Addresses)
	fmt.Println("k1:", k1.GetPublicKey().HashBase58(), n1.GetPeerInfo().Addresses)
	fmt.Println("k2:", k2.GetPublicKey().HashBase58(), n2.GetPeerInfo().Addresses)
	fmt.Printf("-----------------------------\n\n\n\n")

	disc1.Add(n1.GetPeerInfo())
	disc1.Add(n2.GetPeerInfo())
	disc2.Add(n2.GetPeerInfo())
	disc3.Add(n1.GetPeerInfo())

	// init connection from n1 to n0
	err := x1.Send(
		context.Background(),
		object.FromMap(map[string]interface{}{"foo": "bar"}),
		n0.GetPeerInfo().Addresses[0],
	)
	assert.NoError(t, err)

	// init connection from n2 to n0
	err = x2.Send(
		context.Background(),
		object.FromMap(map[string]interface{}{"foo": "bar"}),
		n0.GetPeerInfo().Addresses[0],
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

	_, err = x1.Handle("test/msg", func(o *object.Object) error {
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		assert.NotNil(t, eo1.GetSignerKey())
		assert.NotNil(t, o.GetSignerKey())
		assert.Equal(t, eo1.GetSignerKey(), o.GetSignerKey())
		assert.Equal(t, eo1.GetSignerKey().HashBase58(), o.GetSignerKey().HashBase58())
		assert.NotNil(t, eo1.GetSignature())
		assert.NotNil(t, o.GetSignature())
		assert.Equal(t, eo1.GetSignature(), o.GetSignature())
		assert.Equal(t, eo1.GetSignature().HashBase58(), o.GetSignature().HashBase58())
		w1ObjectHandled = true
		wg.Done()
		return nil
	})
	assert.NoError(t, err)

	_, err = x2.Handle("tes**", func(o *object.Object) error {
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

func newPeer(t *testing.T, relayAddress string,
	discover discovery.Discoverer) (*crypto.Key, net.Network, *exchange) {
	tp, err := ioutil.TempDir("", "nimona-test-net")
	assert.NoError(t, err)

	sp := filepath.Join(tp, "objects")

	pk, err := crypto.GenerateKey()
	assert.NoError(t, err)

	ds := storage.NewDiskStorage(sp)

	relayAddresses := []string{}
	if relayAddress != "" {
		relayAddresses = append(relayAddresses, relayAddress)
	}
	n, err := net.New(pk, "", relayAddresses, discover)
	assert.NoError(t, err)

	x, err := New(pk, n, ds, discover, fmt.Sprintf("0.0.0.0:%d", 0))
	assert.NoError(t, err)

	return pk, n, x.(*exchange)
}

func compareObjects(t *testing.T, expected, actual *object.Object) {
	for m := range expected.Members {
		assert.Equal(t, expected.GetRaw(m), actual.GetRaw(m))
	}
}
