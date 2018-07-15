package net_test

// import (
// 	"context"
// 	"fmt"
// 	"net"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/suite"

// 	"github.com/nimona/go-nimona/dht"
// 	"github.com/nimona/go-nimona/net"
// 	"github.com/nimona/go-nimona/net"
// )

// type messengerTestSuite struct {
// 	suite.Suite
// 	bootstrapPeerInfos []PeerInfo
// }

// func (suite *messengerTestSuite) SetupTest() {
// 	suite.bootstrapPeerInfos = []PeerInfo{}
// }

// func (suite *messengerTestSuite) TestSendSuccess() {
// 	port1, p1, w1, r1 := suite.newPeer()
// 	_, p2, w2, r2 := suite.newPeer()

// 	p1s := p1.Message()
// 	p1s.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", port1)}
// 	r2.PutPeerInfoFromMessage(&p1s)

// 	p2s := p2.Message()
// 	// p2s.Addresses = []string{"tcp:127.0.0.1:32012"}
// 	r1.PutPeerInfoFromMessage(&p2s)

// 	time.Sleep(time.Second)

// 	payload := map[string]string{
// 		"foo": "bar",
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	w1MessageHandled := false
// 	w2MessageHandled := false

// 	w1.Handle("foo", func(message *net.Message) error {
// 		decPayload := map[string]string{}
// 		err := message.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w1MessageHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(message *net.Message) error {
// 		decPayload := map[string]string{}
// 		err := message.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w2MessageHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	err := w2.Send(ctx, "foo", "bar", payload, []string{p1.ID})
// 	suite.NoError(err)

// 	time.Sleep(time.Second)

// 	err = w1.Send(ctx, "foo", "bar", payload, []string{p2.ID})
// 	suite.NoError(err)

// 	wg.Wait()

// 	suite.True(w1MessageHandled)
// 	suite.True(w2MessageHandled)
// }

// func (suite *messengerTestSuite) TestRelayedSendSuccess() {
// 	portR, pR, wR, rR := suite.newPeer()
// 	pRs := pR.Message()
// 	pRs.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", portR)}

// 	_, p1, w1, r1 := suite.newPeer()
// 	_, p2, w2, r2 := suite.newPeer()

// 	r1.PutPeerInfoFromMessage(&pRs)
// 	r2.PutPeerInfoFromMessage(&pRs)

// 	p1s := p1.Message()
// 	p1s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromMessage(&p1s)
// 	r2.PutPeerInfoFromMessage(&p1s)

// 	p2s := p2.Message()
// 	p2s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromMessage(&p2s)
// 	r1.PutPeerInfoFromMessage(&p2s)

// 	dht.NewDHT(wR, rR)
// 	dht.NewDHT(w1, r1)
// 	dht.NewDHT(w2, r1)

// 	time.Sleep(time.Second)

// 	payload := map[string]string{
// 		"foo": "bar",
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	w1MessageHandled := false
// 	w2MessageHandled := false

// 	w1.Handle("foo", func(message *net.Message) error {
// 		decPayload := map[string]string{}
// 		err := message.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w1MessageHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(message *net.Message) error {
// 		decPayload := map[string]string{}
// 		err := message.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w2MessageHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	err := w2.Send(ctx, "foo", "bar", payload, []string{p1.ID})
// 	suite.NoError(err)

// 	time.Sleep(time.Second)

// 	err = w1.Send(ctx, "foo", "bar", payload, []string{p2.ID})
// 	suite.NoError(err)

// 	wg.Wait()

// 	suite.True(w1MessageHandled)
// 	suite.True(w2MessageHandled)
// }

// func (suite *messengerTestSuite) newPeer() (int, *SecretPeerInfo, net.Messenger, AddressBook) {
// 	reg := NewAddressBook()
// 	spi, _ := reg.CreateNewPeer()
// 	reg.PutLocalPeerInfo(spi)

// 	for _, peerInfo := range suite.bootstrapPeerInfos {
// 		err := reg.PutPeerInfoFromMessage(&peerInfo)
// 		suite.NoError(err)
// 	}

// 	wre, _ := net.NewMessenger(reg)
// 	listener, _, lErr := wre.Listen(fmt.Sprintf("0.0.0.0:%d", 0))
// 	suite.NoError(lErr)
// 	port := listener.Addr().(*net.TCPAddr).Port
// 	return port, spi, wre, reg
// }

// func TestMessengerTestSuite(t *testing.T) {
// 	suite.Run(t, new(messengerTestSuite))
// }
