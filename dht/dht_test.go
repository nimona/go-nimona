package dht

import (
	"context"
	"testing"

	"github.com/nimona/go-nimona/net"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type dhtTestSuite struct {
	suite.Suite
	addressBook net.AddressBook
	messenger   *net.MockWire
	peerID      string
	messages    chan interface{}
	peers       chan interface{}
	dht         *DHT
}

func (suite *dhtTestSuite) SetupTest() {
	suite.messages = make(chan interface{}, 10)
	suite.peers = make(chan interface{}, 10)
	suite.addressBook = net.NewRegisty()
	peer1, _ := suite.addressBook.CreateNewPeer()
	suite.addressBook.PutLocalPeerInfo(peer1)
	suite.addressBook.PutPeerInfo(&net.PeerInfo{
		ID: "bootstrap",
		Addresses: []string{
			"localhost",
		},
	})
	suite.messenger = &net.MockWire{}
	suite.net.On("HandleExtensionEvents", mock.Anything, mock.Anything).Return(nil)
	suite.dht, _ = NewDHT(suite.messenger, suite.addressBook)
}

func (suite *dhtTestSuite) TestPutSuccess() {
	ctx := context.Background()
	key := "a"
	value := "b"
	payload := messagePutValue{
		SenderPeerInfo: suite.addressBook.GetLocalPeerInfo().ToPeerInfo(),
		Key:            "a",
		Value:          "b",
	}
	to := []string{"bootstrap"}
	suite.net.On("Send", mock.Anything, "dht", PayloadTypePutValue, payload, to).Return(nil)
	err := suite.dht.PutValue(ctx, key, value)
	suite.NoError(err)

	suite.net.On("Send", mock.Anything, "dht", PayloadTypeGetValue, mock.Anything, to).Return(nil)
	retValue, err := suite.dht.GetValue(ctx, key)
	suite.NoError(err)
	suite.Equal(value, retValue)
}

func TestDHTTestSuite(t *testing.T) {
	suite.Run(t, new(dhtTestSuite))
}
