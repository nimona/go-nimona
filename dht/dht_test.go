package dht

import (
	"context"
	"testing"

	"github.com/nimona/go-nimona/wire"

	"github.com/nimona/go-nimona/mesh"

	"github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

type dhtTestSuite struct {
	suite.Suite
	mockMesh *mesh.MockMesh
	registry mesh.Registry
	wire     *wire.MockWire
	messages chan interface{}
	peers    chan interface{}
	dht      *DHT
}

func (suite *dhtTestSuite) SetupTest() {
	suite.mockMesh = &mesh.MockMesh{}
	suite.messages = make(chan interface{}, 10)
	suite.peers = make(chan interface{}, 10)
	suite.registry, _ = mesh.NewRegisty("local-peer")
	suite.wire = &wire.MockWire{}
	suite.wire.On("HandleExtensionEvents", mock.Anything, mock.Anything).Return(nil)
	suite.dht, _ = NewDHT(suite.wire, suite.registry, "local-peer", false, "bootstrap-address")
}

func (suite *dhtTestSuite) TestPutSuccess() {
	ctx := context.Background()
	key := "a"
	value := "b"
	payload := messagePutValue{
		SenderPeerInfo: mesh.PeerInfo{
			ID:        "local-peer",
			Protocols: map[string][]string{},
		},
		Key:   "a",
		Value: "b",
	}
	to := []string{"bootstrap"}
	suite.wire.On("Send", mock.Anything, "dht", "put-value", payload, to).Return(nil)
	err := suite.dht.PutValue(ctx, key, value)
	suite.NoError(err)

	suite.wire.On("Send", mock.Anything, "dht", "get-value", mock.Anything, to).Return(nil)
	retValue, err := suite.dht.GetValue(ctx, key)
	suite.NoError(err)
	suite.Equal(value, retValue)
}

func TestDHTTestSuite(t *testing.T) {
	suite.Run(t, new(dhtTestSuite))
}
