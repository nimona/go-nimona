package dht

import (
	suite "github.com/stretchr/testify/suite"
)

type dhtTestSuite struct {
	suite.Suite
	node1 *DHT
	node2 *DHT
	node3 *DHT
	node4 *DHT
	node5 *DHT
}

// func TestExampleTestSuite(t *testing.T) {
// 	logrus.SetLevel(logrus.DebugLevel)

// 	peer1 := map[string][]string{
// 		"a1": []string{},
// 	}
// 	peer2 := map[string][]string{
// 		"a2": []string{},
// 	}
// 	// peer3 := net.Peer{ID: "a3", Addresses: []string{}}
// 	// peer4 := net.Peer{ID: "a4", Addresses: []string{}}
// 	peer5 := map[string][]string{
// 		"a5": []string{},
// 	}

// 	net1, err := fabric.New(context.Background())
// 	assert.Nil(t, err)

// 	net2, err := fabric.New(context.Background())
// 	assert.Nil(t, err)

// 	// net3, err := net.NewNetwork(&peer3, 0)
// 	// assert.Nil(t, err)

// 	// net4, err := net.NewNetwork(&peer4, 0)
// 	// assert.Nil(t, err)

// 	net5, err := fabric.New(context.Background())
// 	assert.Nil(t, err)

// 	node1, _ := NewDHT(map[string][]string{}, "a1", net1)
// 	node2, _ := NewDHT([]net.Peer{peer1}, peer2, net2)
// 	// node3, _ := NewDHT([]net.Peer{peer1}, peer3, net3)
// 	// node4, _ := NewDHT([]net.Peer{peer1}, peer4, net4)
// 	node5, _ := NewDHT([]net.Peer{peer1}, peer5, net5)

// 	dt := &dhtTestSuite{
// 		node1: node1,
// 		node2: node2,
// 		// node3: node3,
// 		// node4: node4,
// 		node5: node5,
// 	}

// 	time.Sleep(time.Second * 5)

// 	suite.Run(t, dt)
// }

// func (suite *dhtTestSuite) TestFindSuccess() {
// 	ctx := context.Background()
// 	id := "a5"
// 	peer, err := suite.node2.GetPeer(ctx, id)
// 	suite.Nil(err)
// 	suite.Equal(id, peer.ID)
// }

// func (suite *dhtTestSuite) TestFindNodeLocalSuccess() {
// 	ctx := context.Background()
// 	id := "a1"

// 	peer, err := suite.node2.GetPeer(ctx, id)
// 	suite.Nil(err)
// 	suite.Equal(id, peer.ID)
// }

// func (suite *dhtTestSuite) TestFindNodeTimeout() {
// 	// swallow cancelation  function to make sure we test timeout
// 	ctx, _ := context.WithTimeout(
// 		context.Background(),
// 		time.Second,
// 	)

// 	id := "does-not-exist"

// 	peer, err := suite.node2.GetPeer(ctx, id)
// 	suite.Equal(ErrPeerNotFound, err)
// 	suite.Empty(peer.ID)
// }

// func (suite *dhtTestSuite) TestFindNodeCancelation() {
// 	// swallow cancelation  function to make sure we test timeout
// 	ctx, cf := context.WithCancel(
// 		context.Background(),
// 	)

// 	go func() {
// 		time.Sleep(time.Second)
// 		cf()
// 	}()

// 	id := "does-not-exist"

// 	peer, err := suite.node2.GetPeer(ctx, id)
// 	suite.Equal(ErrPeerNotFound, err)
// 	suite.Empty(peer.ID)
// }

// func (suite *dhtTestSuite) TestFindKeySuccess() {
// 	key := "some-key"
// 	value := "some-value"
// 	err := suite.node2.Put(context.Background(), key, value)
// 	suite.Nil(err)

// 	res, err := suite.node5.Get(context.Background(), key)
// 	resValue := ""
// 	select {
// 	case <-time.After(time.Second * 5):
// 	case v := <-res:
// 		resValue = v
// 	}
// 	suite.Nil(err)
// 	suite.Equal(value, resValue)
// }
