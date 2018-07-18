package dht

// import (
// 	"context"
// 	"testing"

// 	"github.com/nimona/go-nimona/net"

// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"
// )

// type dhtTestSuite struct {
// 	suite.Suite
// 	addressBook *net.AddressBook
// 	messenger   *net.MockMessenger
// 	peerID      string
// 	envelopes    chan interface{}
// 	peers       chan interface{}
// 	dht         *DHT
// }

// func (suite *dhtTestSuite) SetupTest() {
// 	suite.envelopes = make(chan interface{}, 10)
// 	suite.peers = make(chan interface{}, 10)
// 	suite.addressBook = net.NewAddressBook()
// 	peer1, _ := suite.addressBook.CreateNewPeer()
// 	suite.addressBook.PutLocalPeerInfo(peer1)
// 	suite.addressBook.PutPeerInfoFromEnvelope(&net.PeerInfo{
// 		ID: "bootstrap",
// 		Addresses: []string{
// 			"localhost",
// 		},
// 	})
// 	suite.messenger = &net.MockMessenger{}
// 	suite.messenger.On("Handle", mock.Anything, mock.Anything).Return(nil)
// 	suite.dht, _ = NewDHT(suite.messenger, suite.addressBook)
// }

// func (suite *dhtTestSuite) TestPutSuccess() {
// 	ctx := context.Background()
// 	key := "a"
// 	value := "b"
// 	payload := EnvelopePutValue{
// 		SenderPeerInfo: suite.addressBook.GetLocalPeerInfo().Envelope(),
// 		Key:            "a",
// 		Value:          "b",
// 	}
// 	to := []string{"bootstrap"}
// 	envelope, err := net.NewEnvelope(PayloadTypePutValue, to, payload)
// 	suite.NoError(err)
// 	suite.messenger.On("Send", PayloadTypePutValue, envelope).Return(nil)
// 	err = suite.dht.PutValue(ctx, key, value)
// 	suite.NoError(err)

// 	suite.messenger.On("Send", PayloadTypeGetValue, mock.Anything).Return(nil)
// 	retValue, err := suite.dht.GetValue(ctx, key)
// 	suite.NoError(err)
// 	suite.Equal(value, retValue)
// }

// func TestDHTTestSuite(t *testing.T) {
// 	suite.Run(t, new(dhtTestSuite))
// }
