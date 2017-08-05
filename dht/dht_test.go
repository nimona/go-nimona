package dht

import (
	"context"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
	suite "github.com/stretchr/testify/suite"

	net "github.com/nimona/go-nimona-net"
)

type dhtTestSuite struct {
	suite.Suite
	node1 *DHTNode
	node2 *DHTNode
	node3 *DHTNode
}

func TestExampleTestSuite(t *testing.T) {
	peer1 := net.Peer{ID: "a1", Addresses: []string{"0.0.0.0:21600"}}
	peer2 := net.Peer{ID: "a2", Addresses: []string{"0.0.0.0:21601"}}
	peer3 := net.Peer{ID: "a3", Addresses: []string{"0.0.0.0:21602"}}
	peer4 := net.Peer{ID: "a4", Addresses: []string{"127.0.0.1:8821"}}
	peer5 := net.Peer{ID: "a5", Addresses: []string{"0.0.0.0:21603"}}

	net1, err := net.NewTCPNetwork(&peer1)
	assert.Nil(t, err)

	net2, err := net.NewTCPNetwork(&peer2)
	assert.Nil(t, err)

	net3, err := net.NewTCPNetwork(&peer3)
	assert.Nil(t, err)

	rt1 := NewSimpleRoutingTable()
	rt2 := NewSimpleRoutingTable()
	rt3 := NewSimpleRoutingTable()

	node1, _ := NewDHTNode([]net.Peer{peer2}, peer1, rt1, net1)
	node2, _ := NewDHTNode([]net.Peer{peer1, peer3}, peer2, rt2, net2)
	node3, _ := NewDHTNode([]net.Peer{peer2, peer4, peer5}, peer3, rt3, net3)

	dt := &dhtTestSuite{
		node1: node1,
		node2: node2,
		node3: node3,
	}

	suite.Run(t, dt)
}

func (suite *dhtTestSuite) TestFindSuccess() {
	ctx := context.Background()
	id := "a5"

	peer, err := suite.node1.Find(ctx, id)
	suite.Nil(err)
	suite.Equal(id, peer.ID)

	peer, err = suite.node1.Find(ctx, id)
	suite.Nil(err)
	suite.Equal(id, peer.ID)

	peer, err = suite.node1.Find(ctx, id)
	suite.Nil(err)
	suite.Equal(id, peer.ID)

	peer, err = suite.node1.Find(ctx, id)
	suite.Nil(err)
	suite.Equal(id, peer.ID)
}

func (suite *dhtTestSuite) TestFindNodeLocalSuccess() {
	ctx := context.Background()
	id := "a1"

	peer, err := suite.node2.Find(ctx, id)
	suite.Nil(err)
	suite.Equal(id, peer.ID)
}

func (suite *dhtTestSuite) TestFindNodeTimeout() {
	// swallow cancelation  function to make sure we test timeout
	ctx, _ := context.WithTimeout(
		context.Background(),
		time.Second,
	)

	id := "does-not-exist"

	peer, err := suite.node2.Find(ctx, id)
	suite.Equal(ErrPeerNotFound, err)
	suite.Empty(peer.ID)
}

func (suite *dhtTestSuite) TestFindNodeCancelation() {
	// swallow cancelation  function to make sure we test timeout
	ctx, cf := context.WithCancel(
		context.Background(),
	)

	go func() {
		time.Sleep(time.Second)
		cf()
	}()

	id := "does-not-exist"

	peer, err := suite.node2.Find(ctx, id)
	suite.Equal(ErrPeerNotFound, err)
	suite.Empty(peer.ID)
}
