package dht

import (
	"context"
	"testing"
	"time"

	logrus "github.com/sirupsen/logrus"
	suite "github.com/stretchr/testify/suite"

	net "github.com/nimona/go-nimona/net"
)

type dhtTestSuite struct {
	suite.Suite
	node1 *DHT
	node2 *DHT
	node3 *DHT
	node4 *DHT
	node5 *DHT
}

func TestExampleTestSuite(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	net1 := net.New(context.Background())
	node1, _ := NewDHT(map[string][]string{}, "a1", net1)
	net1.AddTransport(net.NewTransportTCP("0.0.0.0", 0), node1)

	net2 := net.New(context.Background())
	node2, _ := NewDHT(map[string][]string{"a1": net1.GetAddresses()}, "a2", net2)
	net2.AddTransport(net.NewTransportTCP("0.0.0.0", 0), node2)

	net5 := net.New(context.Background())
	node5, _ := NewDHT(map[string][]string{"a1": net1.GetAddresses()}, "a5", net5)
	net5.AddTransport(net.NewTransportTCP("0.0.0.0", 0), node5)

	dt := &dhtTestSuite{
		node1: node1,
		node2: node2,
		// node3: node3,
		// node4: node4,
		node5: node5,
	}

	time.Sleep(time.Second * 5)
	suite.Run(t, dt)
}

func (suite *dhtTestSuite) TestFindSuccess() {
	ctx := context.Background()
	id := "a5"
	addresses, err := suite.node2.GetPeer(ctx, id)
	suite.Nil(err)
	suite.NotEmpty(addresses)
}

func (suite *dhtTestSuite) TestFindNodeLocalSuccess() {
	ctx := context.Background()
	id := "a1"

	addresses, err := suite.node2.GetPeer(ctx, id)
	suite.Nil(err)
	suite.NotEmpty(addresses)
}

func (suite *dhtTestSuite) TestFindNodeTimeout() {
	// swallow cancelation  function to make sure we test timeout
	ctx, _ := context.WithTimeout(
		context.Background(),
		time.Second,
	)

	id := "does-not-exist"

	addresses, err := suite.node2.GetPeer(ctx, id)
	suite.Equal(ErrPeerNotFound, err)
	suite.Empty(addresses)
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

	addresses, err := suite.node2.GetPeer(ctx, id)
	suite.Equal(ErrPeerNotFound, err)
	suite.Empty(addresses)
}

func (suite *dhtTestSuite) TestFindKeySuccess() {
	key := "some-key"
	value := "some-value"
	err := suite.node1.Put(context.Background(), key, value)
	suite.Nil(err)

	res, err := suite.node5.Get(context.Background(), key)
	resValue := ""
	select {
	case <-time.After(time.Second * 5):
	case v := <-res:
		resValue = v
	}
	suite.Nil(err)
	suite.Equal(value, resValue)
}
