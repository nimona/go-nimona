package net_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
	"nimona.io/go/storage"
)

type exchangeTestSuite struct {
	suite.Suite
}

func (suite *exchangeTestSuite) TestSendSuccess() {
	_, p1, _, w1, r1 := suite.newPeer()
	_, p2, k2, w2, r2 := suite.newPeer()

	err := r2.PutPeerInfo(p1)
	suite.NoError(err)
	err = r1.PutPeerInfo(p2)
	suite.NoError(err)

	time.Sleep(time.Second)

	exPayload1 := &primitives.Block{
		Type: "test.nimona.io/dummy",
		Payload: map[string]interface{}{
			"foo": "bar1",
		},
	}

	exPayload2 := &primitives.Block{
		Type: "test.nimona.io/dummy",
		Payload: map[string]interface{}{
			"foo": "bar2",
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1BlockHandled := false
	w2BlockHandled := false

	w1.Handle("test.nimona.io/dummy", func(block *primitives.Block) error {
		suite.Equal(k2.GetPublicKey().Thumbprint(), block.Signature.Key.Thumbprint())
		block.Signature = nil
		exPayload1.Signature = nil
		suite.Equal(exPayload1, block)
		w1BlockHandled = true
		wg.Done()
		return nil
	})

	w2.Handle("test.nimona.**", func(block *primitives.Block) error {
		suite.Equal(exPayload2, block)
		w2BlockHandled = true
		wg.Done()
		return nil
	})

	ctx := context.Background()

	err = w2.Send(ctx, exPayload1, p1.Signature.Key, primitives.SignWith(k2))
	suite.NoError(err)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	err = w1.Send(ctx, exPayload2, p2.Signature.Key)
	suite.NoError(err)

	wg.Wait()

	suite.True(w1BlockHandled)
	suite.True(w2BlockHandled)
}

// func (suite *exchangeTestSuite) TestRelayedSendSuccess() {
// 	portR, pR, wR, rR := suite.newPeer()
// 	pRs := pR.Block()
// 	pRs.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", portR)}

// 	_, p1, w1, r1 := suite.newPeer()
// 	_, p2, w2, r2 := suite.newPeer()

// 	r1.PutPeerInfoFromBlock(&pRs)
// 	r2.PutPeerInfoFromBlock(&pRs)

// 	p1s := p1.Block()
// 	p1s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromBlock(&p1s)
// 	r2.PutPeerInfoFromBlock(&p1s)

// 	p2s := p2.Block()
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
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w1BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(block *nprimitives.Block) error {
// 		decPayload := map[string]string{}
// 		err := block.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w2BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	block, err := net.NewBlock("foo.bar", []string{p1.ID}, payload)
// 	suite.NoError(err)
// 	err = w2.Send(ctx, block)
// 	suite.NoError(err)

// 	time.Sleep(time.Second)

// 	block, err = net.NewBlock("foo.bar", []string{p2.ID}, payload)
// 	suite.NoError(err)
// 	err = w1.Send(ctx, block)
// 	suite.NoError(err)

// 	wg.Wait()

// 	suite.True(w1BlockHandled)
// 	suite.True(w2BlockHandled)
// }

func (suite *exchangeTestSuite) newPeer() (int, *peers.PeerInfo, *primitives.Key, net.Exchange, *peers.AddressBook) {
	td, _ := ioutil.TempDir("", "nimona-test-net")
	ab, _ := peers.NewAddressBook(td)
	storagePath := path.Join(td, "storage")
	dpr := storage.NewDiskStorage(storagePath)
	wre, err := net.NewExchange(ab, dpr, fmt.Sprintf("0.0.0.0:%d", 0))
	suite.NoError(err)
	return 0, ab.GetLocalPeerInfo(), ab.GetLocalPeerKey(), wre, ab
}

func TestExchangeTestSuite(t *testing.T) {
	suite.Run(t, new(exchangeTestSuite))
}
