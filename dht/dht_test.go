package dht

// import (
// 	"context"
// 	"io/ioutil"
// 	"testing"

// 	"github.com/nimona/go-nimona/net"

// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"
// )

// type dhtTestSuite struct {
// 	suite.Suite
// 	addressBook *net.AddressBook
// 	exchange    *net.MockExchange
// 	peerID      string
// 	blocks      chan interface{}
// 	peers       chan interface{}
// 	dht         *DHT
// }

// func (suite *dhtTestSuite) SetupTest() {
// 	suite.blocks = make(chan interface{}, 10)
// 	suite.peers = make(chan interface{}, 10)
// 	td, _ := ioutil.TempDir("", "nimona-test-dht")
// 	suite.addressBook, _ = net.NewAddressBook(td)
// 	peer1, _ := suite.addressBook.CreateNewPeer()
// 	suite.addressBook.PutLocalPeerInfo(peer1)
// 	bootstrapBlock := net.NewEphemeralBlock("peer.info", &net.PeerInfo{
// 		ID: "bootstrap",
// 		Addresses: []string{
// 			"localhost",
// 		},
// 	})
// 	net.SetSigner(bootstrapBlock, suite.addressBook.GetLocalPeerInfo())
// 	suite.addressBook.PutPeerInfoFromBlock(bootstrapBlock)
// 	suite.exchange = &net.MockExchange{}
// 	suite.exchange.On("Handle", mock.Anything, mock.Anything).Return(nil)
// 	suite.dht, _ = NewDHT(suite.exchange, suite.addressBook)
// }

// func (suite *dhtTestSuite) TestPutSuccess() {
// 	ctx := context.Background()
// 	key := "a"
// 	payload := BlockPutProviders{
// 		Key: "a",
// 		Providers: []*net.Block{
// 			suite.addressBook.GetLocalPeerInfo().Block(),
// 		},
// 	}
// 	to := "bootstrap"
// 	block := net.NewEphemeralBlock(PayloadTypePutValue, payload)
// 	suite.exchange.On("Send", PayloadTypePutValue, block, to).Return(nil)
// 	err := suite.dht.PutProviders(ctx, key)
// 	suite.NoError(err)

// 	suite.exchange.On("Send", PayloadTypeGetValue, mock.Anything).Return(nil)
// 	retValue, err := suite.dht.GetProviders(ctx, key)
// 	suite.NoError(err)
// 	suite.Equal([]string{"bootstrap"}, retValue)
// }

// func TestDHTTestSuite(t *testing.T) {
// 	suite.Run(t, new(dhtTestSuite))
// }
