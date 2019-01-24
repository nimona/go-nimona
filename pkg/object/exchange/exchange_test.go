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
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/storage"
)

func TestSendSuccess(t *testing.T) {
	k1, n1, x1 := newPeer(t, "")
	k2, n2, x2 := newPeer(t, "")

	n1.Discoverer().Add(n2.GetPeerInfo())
	n2.Discoverer().Add(n1.GetPeerInfo())

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

	w1BlockHandled := false
	w2BlockHandled := false

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
		w1BlockHandled = true
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

		w2BlockHandled = true
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

	assert.True(t, w1BlockHandled)
	assert.True(t, w2BlockHandled)
}

func TestSendRelay(t *testing.T) {
	// enable binding to local addresses
	net.BindLocal = true
	k0, n0, _ := newPeer(t, "")

	// disable binding to local addresses
	net.BindLocal = false
	k1, n1, x1 := newPeer(t, "relay:"+n0.GetPeerInfo().Addresses[0])
	k2, n2, x2 := newPeer(t, "relay:"+n0.GetPeerInfo().Addresses[0])

	fmt.Printf("\n\n\n\n-----------------------------\n")
	fmt.Println("k0:", k0.GetPublicKey().HashBase58(), n0.GetPeerInfo().Addresses)
	fmt.Println("k1:", k1.GetPublicKey().HashBase58(), n1.GetPeerInfo().Addresses)
	fmt.Println("k2:", k2.GetPublicKey().HashBase58(), n2.GetPeerInfo().Addresses)
	fmt.Printf("-----------------------------\n\n\n\n")

	n0.Discoverer().Add(n1.GetPeerInfo())
	n0.Discoverer().Add(n2.GetPeerInfo())
	n1.Discoverer().Add(n2.GetPeerInfo())
	n2.Discoverer().Add(n1.GetPeerInfo())

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

	w1BlockHandled := false
	w2BlockHandled := false

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
		w1BlockHandled = true
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

		w2BlockHandled = true
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

	assert.True(t, w1BlockHandled)
	assert.True(t, w2BlockHandled)
}

func newPeer(t *testing.T, relayAddress string) (*crypto.Key, net.Network, *exchange) {
	tp, err := ioutil.TempDir("", "nimona-test-net")
	assert.NoError(t, err)

	kp := filepath.Join(tp, "key.cbor")
	sp := filepath.Join(tp, "objects")

	pk, err := crypto.LoadKey(kp)
	assert.NoError(t, err)

	ds := storage.NewDiskStorage(sp)

	relayAddresses := []string{}
	if relayAddress != "" {
		relayAddresses = append(relayAddresses, relayAddress)
	}
	n, err := net.New(pk, "", relayAddresses)
	assert.NoError(t, err)

	x, err := New(pk, n, ds, fmt.Sprintf("0.0.0.0:%d", 0))
	assert.NoError(t, err)

	return pk, n, x.(*exchange)
}
