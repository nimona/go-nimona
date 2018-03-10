package net

// Basic imports
import (
	"context"
	"testing"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

// NetTestSuite -
type NetTestSuite struct {
	suite.Suite
	fnet *nnet
	ctx  context.Context
}

func (suite *NetTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.fnet = New(suite.ctx).(*nnet)
}

func (suite *NetTestSuite) TestAddTransportSuccess() {
	transport1 := &MockTransport{}
	transport1.On("Listen", mock.Anything, mock.Anything).Return(nil)
	err := suite.fnet.AddTransport(transport1, []Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fnet.transports, 1)
	suite.Assert().Equal(transport1, suite.fnet.transports[0].Transport)

	transport2 := &MockTransport{}
	transport2.On("Listen", mock.Anything, mock.Anything).Return(nil)
	err = suite.fnet.AddTransport(transport2, []Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fnet.transports, 2)
	suite.Assert().Equal(transport2, suite.fnet.transports[1].Transport)
}

func (suite *NetTestSuite) TestAddProtocolSuccess() {
	name1 := "protocol1"
	protocol1 := &MockProtocol{}
	protocol1.On("Name").Return(name1)
	err := suite.fnet.AddProtocol(protocol1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fnet.protocols, 1)
	protocol1.AssertCalled(suite.T(), "Name")

	name2 := "protocol2"
	protocol2 := &MockProtocol{}
	protocol2.On("Name").Return(name2)
	err = suite.fnet.AddProtocol(protocol2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fnet.protocols, 2)
	protocol2.AssertCalled(suite.T(), "Name")
}

func (suite *NetTestSuite) TestGetAddressesSuccess() {
	transport1 := &MockTransport{}
	addresses1 := []string{
		"tr1.addr1",
		"tr1.addr2",
	}
	transport1.On("Addresses").Return(addresses1)
	transport1.On("Listen", mock.Anything, mock.Anything).Return(nil)
	err := suite.fnet.AddTransport(transport1, []Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fnet.transports, 1)

	transport2 := &MockTransport{}
	addresses2 := []string{
		"tr2.addr1",
		"tr2.addr2",
	}
	transport2.On("Addresses").Return(addresses2)
	transport2.On("Listen", mock.Anything, mock.Anything).Return(nil)
	err = suite.fnet.AddTransport(transport2, []Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fnet.transports, 2)

	addressesAll := append(addresses1, addresses2...)

	addresses := suite.fnet.GetAddresses()
	suite.Assert().Len(addresses, 4)
	suite.Assert().Equal(addressesAll, addresses)
}

func TestNetTestSuite(t *testing.T) {
	suite.Run(t, new(NetTestSuite))
}
