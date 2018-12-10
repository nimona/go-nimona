package dht

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/net"
	"nimona.io/go/storage"
)

func TestSendSuccess(t *testing.T) {
	os.Setenv("BIND_LOCAL", "true")
	os.Setenv("UPNP", "false")

	k0, n0, x0 := newPeer(t)
	k1, n1, x1 := newPeer(t)
	k2, n2, x2 := newPeer(t)

	fmt.Printf("\n\n\n\n-----------------------------\n")
	fmt.Println("k0:", k0.GetPublicKey().HashBase58())
	fmt.Println("k1:", k1.GetPublicKey().HashBase58())
	fmt.Println("k2:", k2.GetPublicKey().HashBase58())
	fmt.Printf("-----------------------------\n\n\n\n")

	d0, err := NewDHT(k0, n0, x0, []string{})
	assert.NoError(t, err)

	err = n0.Resolver().AddProvider(d0)
	assert.NoError(t, err)

	ba := []string{}
	ba = append(ba, n1.GetPeerInfo().Addresses...)
	ba = append(ba, n2.GetPeerInfo().Addresses...)

	d1, err := NewDHT(k1, n1, x1, ba)
	assert.NoError(t, err)

	err = n1.Resolver().AddProvider(d1)
	assert.NoError(t, err)

	d2, err := NewDHT(k2, n2, x2, ba)
	assert.NoError(t, err)

	err = n2.Resolver().AddProvider(d2)
	assert.NoError(t, err)

	em1 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo1 := encoding.NewObjectFromMap(em1)

	em2 := map[string]interface{}{
		"@ctx": "test/msg",
		"body": "bar1",
	}
	eo2 := encoding.NewObjectFromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1BlockHandled := false
	w2BlockHandled := false

	err = crypto.Sign(eo1, k2)
	assert.NoError(t, err)

	_, err = x1.Handle("test/msg", func(o *encoding.Object) error {
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

	_, err = x2.Handle("tes**", func(o *encoding.Object) error {
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

func newPeer(t *testing.T) (*crypto.Key, net.Network, net.Exchange) {
	tp, err := ioutil.TempDir("", "nimona-test-dht")
	assert.NoError(t, err)

	kp := filepath.Join(tp, "key.cbor")
	sp := filepath.Join(tp, "objects")

	pk, err := crypto.LoadKey(kp)
	assert.NoError(t, err)

	ds := storage.NewDiskStorage(sp)

	n, err := net.NewNetwork(pk, "")
	assert.NoError(t, err)

	x, err := net.NewExchange(pk, n, ds, fmt.Sprintf("127.0.0.1:%d", 0))
	assert.NoError(t, err)

	return pk, n, x
}
