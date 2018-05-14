package dht

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/wire"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type dhtTestSuite struct {
	suite.Suite
	mockMesh *mesh.MockMesh
	registry mesh.Registry
	wire     *wire.MockWire
	peerID   string
	messages chan interface{}
	peers    chan interface{}
	dht      *DHT
}

func (suite *dhtTestSuite) SetupTest() {
	tempDir, err := ioutil.TempDir("", "nimona")
	suite.NoError(err)

	tempFile1 := path.Join(tempDir, "key1")
	key1, err := mesh.LoadOrCreatePrivateKey(tempFile1)
	suite.peerID = mesh.Thumbprint(key1)
	suite.NoError(err)

	suite.mockMesh = &mesh.MockMesh{}
	suite.messages = make(chan interface{}, 10)
	suite.peers = make(chan interface{}, 10)
	suite.registry = mesh.NewRegisty(key1)
	suite.registry.PutPeerInfo(&mesh.PeerInfo{
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
		SenderPeerInfo: mesh.PeerInfo{
			ID:        suite.peerID,
			Addresses: []string{},
		},
		Key:   "a",
		Value: "b",
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
