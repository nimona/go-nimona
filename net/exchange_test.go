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

	blocks "github.com/nimona/go-nimona/blocks"
	nnet "github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/peers"
	storage "github.com/nimona/go-nimona/storage"
)

type exchangeTestSuite struct {
	suite.Suite
}

type DummyPayload struct {
	Foo       string            `nimona:"foo"`
	Signature *blocks.Signature `nimona:",signature"`
}

func (suite *exchangeTestSuite) TestSendSuccess() {
	_, p1, w1, r1 := suite.newPeer()
	_, p2, w2, r2 := suite.newPeer()

	err := r2.PutPeerInfo(p1.GetPeerInfo())
	suite.NoError(err)
	err = r1.PutPeerInfo(p2.GetPeerInfo())
	suite.NoError(err)

	blocks.RegisterContentType("foo", DummyPayload{})

	time.Sleep(time.Second)

	exPayload := &DummyPayload{
		Foo: "bar",
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1BlockHandled := false
	w2BlockHandled := false

	w1.Handle("foo", func(payload interface{}) error {
		suite.Equal(exPayload.Foo, payload.(*DummyPayload).Foo)
		w1BlockHandled = true
		wg.Done()
		return nil
	})

	w2.Handle("foo", func(payload interface{}) error {
		suite.Equal(exPayload.Foo, payload.(*DummyPayload).Foo)
		w2BlockHandled = true
		wg.Done()
		return nil
	})

	ctx := context.Background()

	err = w2.Send(ctx, exPayload, p1.GetPublicKey(), blocks.SignWith(p2.Key))
	suite.NoError(err)

	time.Sleep(time.Second)

	// TODO should be able to send not signed
	err = w1.Send(ctx, exPayload, p2.GetPublicKey(), blocks.SignWith(p1.Key))
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

// 	w1.Handle("foo", func(block *nblocks.Block) error {
// 		decPayload := map[string]string{}
// 		err := block.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w1BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(block *nblocks.Block) error {
// 		decPayload := map[string]string{}
// 		err := block.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w2BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	block, err := nnet.NewBlock("foo.bar", []string{p1.ID}, payload)
// 	suite.NoError(err)
// 	err = w2.Send(ctx, block)
// 	suite.NoError(err)

// 	time.Sleep(time.Second)

// 	block, err = nnet.NewBlock("foo.bar", []string{p2.ID}, payload)
// 	suite.NoError(err)
// 	err = w1.Send(ctx, block)
// 	suite.NoError(err)

// 	wg.Wait()

// 	suite.True(w1BlockHandled)
// 	suite.True(w2BlockHandled)
// }

func (suite *exchangeTestSuite) newPeer() (int, *peers.PrivatePeerInfo, nnet.Exchange, *peers.AddressBook) {
	td, _ := ioutil.TempDir("", "nimona-test-net")
	ab, _ := peers.NewAddressBook(td)
	storagePath := path.Join(td, "storage")
	dpr := storage.NewDiskStorage(storagePath)
	wre, _ := nnet.NewExchange(ab, dpr)
	_, lErr := wre.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", 0))
	suite.NoError(lErr)
	return 0, ab.GetLocalPeerInfo(), wre, ab
}

func TestExchangeTestSuite(t *testing.T) {
	suite.Run(t, new(exchangeTestSuite))
}
