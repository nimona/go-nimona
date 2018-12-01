package net_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/storage"
)

func TestSendSuccess(t *testing.T) {

	_, p1, _, w1, r1 := newPeer(t)
	_, p2, k2, w2, r2 := newPeer(t)

	err := r2.PutPeerInfo(p1)
	assert.NoError(t, err)
	err = r1.PutPeerInfo(p2)
	assert.NoError(t, err)

	time.Sleep(time.Second)

	em1 := map[string]interface{}{
		"@ctx:s": "test/msg",
		"body:s": "bar1",
	}
	eo1 := encoding.NewObjectFromMap(em1)

	em2 := map[string]interface{}{
		"@ctx:s": "test/msg",
		"body:s": "bar1",
	}
	eo2 := encoding.NewObjectFromMap(em2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1BlockHandled := false
	w2BlockHandled := false

	seo1, err := crypto.Sign(eo1, k2)
	assert.NoError(t, err)

	w1.Handle("test/msg", func(o *encoding.Object) error {
		// assert.Equal(t, k2.GetPublicKey().Thumbprint(), block.Signature.Key.Thumbprint())
		assert.Equal(t, eo1.GetRaw("body"), o.GetRaw("body"))
		// assert.Equal(t, exPayload1.Signature.Signature, block.(*TestMessage).Signature.Signature)
		// assert.Equal(t, exPayload1.Signature.Key.Curve, block.(*TestMessage).Signature.Key.Curve)
		w1BlockHandled = true
		wg.Done()
		return nil
	})

	w2.Handle("tes**", func(o *encoding.Object) error {
		assert.Equal(t, eo2.GetRaw("body"), o.GetRaw("body"))
		// assert.Equal(t, exPayload2.Body, block.(*TestMessage).Body)
		// assert.Nil(t, block.(*TestMessage).Signature.Signature)

		w2BlockHandled = true
		wg.Done()
		return nil
	})

	ctx := context.Background()

	err = w2.Send(ctx, seo1, "peer:"+p1.SignerKey.HashBase58())
	assert.NoError(t, err)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	err = w1.Send(ctx, eo2, "peer:"+p2.SignerKey.HashBase58())
	assert.NoError(t, err)

	wg.Wait()

	assert.True(t, w1BlockHandled)
	assert.True(t, w2BlockHandled)
}

// func (assert *exchangeTestSuite)t,  TestRelayedSendSuccess() {
// 	portR, pR, wR, rR := assert.newPeer(t, )
// 	pRs := pR
// 	pRs.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", portR)}

// 	_, p1, w1, r1 := assert.newPeer(t, )
// 	_, p2, w2, r2 := assert.newPeer(t, )

// 	r1.PutPeerInfoFromBlock(&pRs)
// 	r2.PutPeerInfoFromBlock(&pRs)

// 	p1s := p1
// 	p1s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromBlock(&p1s)
// 	r2.PutPeerInfoFromBlock(&p1s)

// 	p2s := p2
// 	p2s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromBlock(&p2s)
// 	r1.PutPeerInfoFromBlock(&p2s)

// 	dht.NewDHT(wR, rR)
// 	dht.NewDHT(w1, r1)
// 	dht.NewDHT(w2, r1)

// 	time.Sleep(time.Second)

// 	payload := map[string]string{
// 		"foo": "bar",
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	w1BlockHandled := false
// 	w2BlockHandled := false

// 	w1.Handle("foo", func(block *nprimitives.Block) error {
// 		decPayload := map[string]string{}
// 		err := block.DecodePayload(&decPayload)
// 		assert.NoError(t, err)
// 		assert.Equal(t, payload, decPayload)
// 		w1BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(block *nprimitives.Block) error {
// 		decPayload := map[string]string{}
// 		err := block.DecodePayload(&decPayload)
// 		assert.NoError(t, err)
// 		assert.Equal(t, payload, decPayload)
// 		w2BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	block, err := net.NewBlock("foo.bar", []string{p1.ID}, payload)
// 	assert.NoError(t, err)
// 	err = w2.Send(ctx, block)
// 	assert.NoError(t, err)

// 	time.Sleep(time.Second)

// 	block, err = net.NewBlock("foo.bar", []string{p2.ID}, payload)
// 	assert.NoError(t, err)
// 	err = w1.Send(ctx, block)
// 	assert.NoError(t, err)

// 	wg.Wait()

// 	assert.True(t, w1BlockHandled)
// 	assert.True(t, w2BlockHandled)
// }

func newPeer(t *testing.T) (int, *peers.PeerInfo, *crypto.Key, net.Exchange, *peers.AddressBook) {
	td, _ := ioutil.TempDir("", "nimona-test-net")
	ab, _ := peers.NewAddressBook(td)
	storagePath := path.Join(td, "storage")
	dpr := storage.NewDiskStorage(storagePath)
	wre, err := net.NewExchange(ab, dpr, fmt.Sprintf("0.0.0.0:%d", 0))
	assert.NoError(t, err)
	return 0, ab.GetLocalPeerInfo(), ab.GetLocalPeerKey(), wre, ab
}
