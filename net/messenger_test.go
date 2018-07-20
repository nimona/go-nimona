package net_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	nnet "github.com/nimona/go-nimona/net"
)

type messengerTestSuite struct {
	suite.Suite
	bootstrapPeerInfos []*nnet.Envelope
}

func (suite *messengerTestSuite) SetupTest() {
	suite.bootstrapPeerInfos = []*nnet.Envelope{}
}

type DummyPayload struct {
	Foo string
}

func (suite *messengerTestSuite) TestSendSuccess() {
	port1, p1, w1, r1 := suite.newPeer()
	_, p2, w2, r2 := suite.newPeer()

	p1.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", port1)}
	r2.PutPeerInfoFromEnvelope(p1.Envelope())
	r1.PutPeerInfoFromEnvelope(p2.Envelope())

	nnet.RegisterContentType("foo.bar", DummyPayload{})

	time.Sleep(time.Second)

	payload := DummyPayload{
		Foo: "bar",
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	w1EnvelopeHandled := false
	w2EnvelopeHandled := false

	w1.Handle("foo", func(envelope *nnet.Envelope) error {
		suite.Equal(payload.Foo, envelope.Payload.(DummyPayload).Foo)
		w1EnvelopeHandled = true
		wg.Done()
		return nil
	})

	w2.Handle("foo", func(envelope *nnet.Envelope) error {
		suite.Equal(payload.Foo, envelope.Payload.(DummyPayload).Foo)
		w2EnvelopeHandled = true
		wg.Done()
		return nil
	})

	ctx := context.Background()

	envelope := nnet.NewEnvelope("foo.bar", []string{p1.ID}, payload)
	err = w2.Send(ctx, envelope)
	suite.NoError(err)

	time.Sleep(time.Second)

	envelope = nnet.NewEnvelope("foo.bar", []string{p2.ID}, payload)
	err = w1.Send(ctx, envelope)
	suite.NoError(err)

	wg.Wait()

	suite.True(w1EnvelopeHandled)
	suite.True(w2EnvelopeHandled)
}

// func (suite *messengerTestSuite) TestRelayedSendSuccess() {
// 	portR, pR, wR, rR := suite.newPeer()
// 	pRs := pR.Envelope()
// 	pRs.Addresses = []string{fmt.Sprintf("tcp:127.0.0.1:%d", portR)}

// 	_, p1, w1, r1 := suite.newPeer()
// 	_, p2, w2, r2 := suite.newPeer()

// 	r1.PutPeerInfoFromEnvelope(&pRs)
// 	r2.PutPeerInfoFromEnvelope(&pRs)

// 	p1s := p1.Envelope()
// 	p1s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromEnvelope(&p1s)
// 	r2.PutPeerInfoFromEnvelope(&p1s)

// 	p2s := p2.Envelope()
// 	p2s.Addresses = []string{"relay:" + pRs.ID}
// 	rR.PutPeerInfoFromEnvelope(&p2s)
// 	r1.PutPeerInfoFromEnvelope(&p2s)

// 	dht.NewDHT(wR, rR)
// 	dht.NewDHT(w1, r1)
// 	dht.NewDHT(w2, r1)

// 	time.Sleep(time.Second)

// 	payload := map[string]string{
// 		"foo": "bar",
// 	}

// 	wg := sync.WaitGroup{}
// 	wg.Add(2)

// 	w1EnvelopeHandled := false
// 	w2EnvelopeHandled := false

// 	w1.Handle("foo", func(envelope *nnet.Envelope) error {
// 		decPayload := map[string]string{}
// 		err := envelope.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w1EnvelopeHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	w2.Handle("foo", func(envelope *nnet.Envelope) error {
// 		decPayload := map[string]string{}
// 		err := envelope.DecodePayload(&decPayload)
// 		suite.NoError(err)
// 		suite.Equal(payload, decPayload)
// 		w2EnvelopeHandled = true
// 		wg.Done()
// 		return nil
// 	})

// 	ctx := context.Background()

// 	envelope, err := nnet.NewEnvelope("foo.bar", []string{p1.ID}, payload)
// 	suite.NoError(err)
// 	err = w2.Send(ctx, envelope)
// 	suite.NoError(err)

// 	time.Sleep(time.Second)

// 	envelope, err = nnet.NewEnvelope("foo.bar", []string{p2.ID}, payload)
// 	suite.NoError(err)
// 	err = w1.Send(ctx, envelope)
// 	suite.NoError(err)

// 	wg.Wait()

// 	suite.True(w1EnvelopeHandled)
// 	suite.True(w2EnvelopeHandled)
// }

func (suite *messengerTestSuite) newPeer() (int, *nnet.PrivatePeerInfo, nnet.Messenger, *nnet.AddressBook) {
	reg := nnet.NewAddressBook()
	spi, _ := reg.CreateNewPeer()
	reg.PutLocalPeerInfo(spi)

	for _, peerInfo := range suite.bootstrapPeerInfos {
		err := reg.PutPeerInfoFromEnvelope(peerInfo)
		suite.NoError(err)
	}

	wre, _ := nnet.NewMessenger(reg)
	listener, lErr := wre.Listen(context.Background(), fmt.Sprintf("0.0.0.0:%d", 0))
	suite.NoError(lErr)
	port := listener.Addr().(*net.TCPAddr).Port
	return port, spi, wre, reg
}

func TestMessengerTestSuite(t *testing.T) {
	suite.Run(t, new(messengerTestSuite))
}
