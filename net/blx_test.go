package net_test

// import (
// 	"context"
// 	"fmt"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/suite"

// 	nnet "github.com/nimona/go-nimona/net"
// )

// type messengerTestSuite struct {
// 	suite.Suite
// 	bootstrapPeerInfos []*nnet.Block
// }

// func (suite *messengerTestSuite) SetupTest() {
// 	suite.bootstrapPeerInfos = []*nnet.Block{}
// }

// type DummyPayload struct {
// 	Foo string
// }

// func (suite *messengerTestSuite) TestSendSuccess() {
// 	port1, p1, w1, r1 := suite.newPeer()
// 	_, p2, w2, r2 := suite.newPeer()

// 	p1.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", port1)}
// 	r2.PutPeerInfoFromBlock(p1.Block())
// 	r1.PutPeerInfoFromBlock(p2.Block())

// 	nnet.RegisterContentType("foo.bar", DummyPayload{})

// 	time.Sleep(time.Second)

// 	payload := DummyPayload{
// 		Foo: "bar",
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	w1BlockHandled := false
// 	w2BlockHandled := false

// 	w1.Handle("foo", func(block *nnet.Block) error {
// 		suite.Equal(payload.Foo, block.Payload.(DummyPayload).Foo)
// 		w1BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(block *nnet.Block) error {
// 		suite.Equal(payload.Foo, block.Payload.(DummyPayload).Foo)
// 		w2BlockHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	block := nnet.NewBlock("foo.bar", []string{p1.ID}, payload)
// 	err := w2.Send(ctx, block)
// 	suite.NoError(err)

// 	time.Sleep(time.Second)

// 	block = nnet.NewBlock("foo.bar", []string{p2.ID}, payload)
// 	err = w1.Send(ctx, block)
// 	suite.NoError(err)

// 	wg.Wait()

// 	suite.True(w1BlockHandled)
// 	suite.True(w2BlockHandled)
// }

// // func (suite *messengerTestSuite) TestRelayedSendSuccess() {
// // 	portR, pR, wR, rR := suite.newPeer()
// // 	pRs := pR.Block()
// // 	pRs.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", portR)}

// // 	_, p1, w1, r1 := suite.newPeer()
// // 	_, p2, w2, r2 := suite.newPeer()

// // 	r1.PutPeerInfoFromBlock(&pRs)
// // 	r2.PutPeerInfoFromBlock(&pRs)

// // 	p1s := p1.Block()
// // 	p1s.Addresses = []string{"relay:" + pRs.ID}
// // 	rR.PutPeerInfoFromBlock(&p1s)
// // 	r2.PutPeerInfoFromBlock(&p1s)

// // 	p2s := p2.Block()
// // 	p2s.Addresses = []string{"relay:" + pRs.ID}
// // 	rR.PutPeerInfoFromBlock(&p2s)
// // 	r1.PutPeerInfoFromBlock(&p2s)

// // 	dht.NewDHT(wR, rR)
// // 	dht.NewDHT(w1, r1)
// // 	dht.NewDHT(w2, r1)

// // 	time.Sleep(time.Second)

// // 	payload := map[string]string{
// // 		"foo": "bar",
// // 	}

// // 	wg := sync.WaitGroup{}
// // 	wg.Add(2)

// // 	w1BlockHandled := false
// // 	w2BlockHandled := false

// // 	w1.Handle("foo", func(block *nnet.Block) error {
// // 		decPayload := map[string]string{}
// // 		err := block.DecodePayload(&decPayload)
// // 		suite.NoError(err)
// // 		suite.Equal(payload, decPayload)
// // 		w1BlockHandled = true
// // 		wg.Done()
// // 		return nil
// // 	})

// // 	w2.Handle("foo", func(block *nnet.Block) error {
// // 		decPayload := map[string]string{}
// // 		err := block.DecodePayload(&decPayload)
// // 		suite.NoError(err)
// // 		suite.Equal(payload, decPayload)
// // 		w2BlockHandled = true
// // 		wg.Done()
// // 		return nil
// // 	})

// // 	ctx := context.Background()

// // 	block, err := nnet.NewBlock("foo.bar", []string{p1.ID}, payload)
// // 	suite.NoError(err)
// // 	err = w2.Send(ctx, block)
// // 	suite.NoError(err)

// // 	time.Sleep(time.Second)

// // 	block, err = nnet.NewBlock("foo.bar", []string{p2.ID}, payload)
// // 	suite.NoError(err)
// // 	err = w1.Send(ctx, block)
// // 	suite.NoError(err)

// // 	wg.Wait()

// // 	suite.True(w1BlockHandled)
// // 	suite.True(w2BlockHandled)
// // }

// func (suite *messengerTestSuite) newPeer() (int, *nnet.PrivatePeerInfo, nnet.Messenger, *nnet.AddressBook) {
// 	reg := nnet.NewAddressBook()
// 	spi, _ := reg.CreateNewPeer()
// 	reg.PutLocalPeerInfo(spi)

// 	for _, peerInfo := range suite.bootstrapPeerInfos {
// 		err := reg.PutPeerInfoFromBlock(peerInfo)
// 		suite.NoError(err)
// 	}

// 	wre, _ := nnet.NewMessenger(reg)
// 	listener, lErr := wre.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", 0))
// 	suite.NoError(lErr)
// 	port := listener.Addr().(*TCPAddr).Port
// 	return port, spi, wre, reg
// }

// func TestMessengerTestSuite(t *testing.T) {
// 	suite.Run(t, new(messengerTestSuite))
// }
