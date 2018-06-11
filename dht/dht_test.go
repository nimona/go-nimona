package dht

import (
	"context"
	"testing"

	"github.com/nimona/go-nimona/peer"
	"github.com/nimona/go-nimona/wire"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type dhtTestSuite struct {
	suite.Suite
	registry peer.Registry
	wire     *wire.MockWire
	peerID   string
	messages chan interface{}
	peers    chan interface{}
	dht      *DHT
}

func (suite *dhtTestSuite) SetupTest() {
	suite.messages = make(chan interface{}, 10)
	suite.peers = make(chan interface{}, 10)
	suite.registry = peer.NewRegisty()
	peer1, _ := suite.registry.CreateNewPeer()
	suite.registry.PutLocalPeerInfo(peer1)
	suite.registry.PutPeerInfo(&peer.PeerInfo{
		ID: "bootstrap",
		Addresses: []string{
			"localhost",
		},
	})
	suite.wire = &wire.MockWire{}
	suite.wire.On("HandleExtensionEvents", mock.Anything, mock.Anything).Return(nil)
	suite.dht, _ = NewDHT(suite.wire, suite.registry)
}

func (suite *dhtTestSuite) TestPutSuccess() {
	ctx := context.Background()
	key := "a"
	value := "b"
	payload := messagePutValue{
		SenderPeerInfo: suite.registry.GetLocalPeerInfo().ToPeerInfo(),
		Key:            "a",
		Value:          "b",
	}
	to := []string{"bootstrap"}
	suite.wire.On("Send", mock.Anything, "dht", PayloadTypePutValue, payload, to).Return(nil)
	err := suite.dht.PutValue(ctx, key, value)
	suite.NoError(err)

	suite.wire.On("Send", mock.Anything, "dht", PayloadTypeGetValue, mock.Anything, to).Return(nil)
	retValue, err := suite.dht.GetValue(ctx, key)
	suite.NoError(err)
	suite.Equal(value, retValue)
}

func TestDHTTestSuite(t *testing.T) {
	suite.Run(t, new(dhtTestSuite))
}
