package fabric

// Basic imports
import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

// FabricTestSuite -
type FabricTestSuite struct {
	suite.Suite
	fabric *Fabric
}

func (suite *FabricTestSuite) SetupTest() {
	suite.fabric = New()
}

func (suite *FabricTestSuite) TestAddTransportSuccess() {
	transport1 := &MockTransport{}
	err := suite.fabric.AddTransport(transport1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport1, suite.fabric.transports[0])

	transport2 := &MockTransport{}
	err = suite.fabric.AddTransport(transport2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 2)
	suite.Assert().Equal(transport2, suite.fabric.transports[1])
}

func (suite *FabricTestSuite) TestAddMiddlewareSuccess() {
	name1 := "middleware1"
	middleware1 := &MockMiddleware{}
	middleware1.On("Name").Return(name1)
	err := suite.fabric.AddMiddleware(middleware1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.handlers, 1)
	suite.Assert().Len(suite.fabric.negotiators, 1)
	middleware1.AssertCalled(suite.T(), "Name")

	name2 := "middleware2"
	middleware2 := &MockMiddleware{}
	middleware2.On("Name").Return(name2)
	err = suite.fabric.AddMiddleware(middleware2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.handlers, 2)
	suite.Assert().Len(suite.fabric.negotiators, 2)
	middleware2.AssertCalled(suite.T(), "Name")
}

func (suite *FabricTestSuite) TestAddHandlerSuccess() {
	name1 := "handler1"
	handler1 := &MockHandler{}
	handler1.On("Name").Return(name1)
	err := suite.fabric.AddHandler(handler1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.handlers, 1)
	handler1.AssertCalled(suite.T(), "Name")

	name2 := "handler2"
	handler2 := &MockHandler{}
	handler2.On("Name").Return(name2)
	err = suite.fabric.AddHandler(handler2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.handlers, 2)
	handler2.AssertCalled(suite.T(), "Name")
}

func (suite *FabricTestSuite) TestAddNegotiatorSuccess() {
	name1 := "negotiator1"
	negotiator1 := &MockNegotiator{}
	negotiator1.On("Name").Return(name1)
	err := suite.fabric.AddNegotiator(negotiator1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.negotiators, 1)
	negotiator1.AssertCalled(suite.T(), "Name")

	name2 := "negotiator2"
	negotiator2 := &MockNegotiator{}
	negotiator2.On("Name").Return(name2)
	err = suite.fabric.AddNegotiator(negotiator2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.negotiators, 2)
	negotiator2.AssertCalled(suite.T(), "Name")
}

func (suite *FabricTestSuite) TestAddHandlerFuncSuccess() {
	name1 := "handler1"
	handler1 := func(ctx context.Context, conn Conn) (context.Context, Conn, error) {
		return context.Background(), &MockConn{}, nil
	}
	err := suite.fabric.AddHandlerFunc(name1, handler1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.handlers, 1)

	name2 := "handler2"
	handler2 := func(ctx context.Context, conn Conn) (context.Context, Conn, error) {
		return context.Background(), &MockConn{}, nil
	}
	err = suite.fabric.AddHandlerFunc(name2, handler2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.handlers, 2)
}

func (suite *FabricTestSuite) TestAddNegotiatorFuncSuccess() {
	name1 := "negotiator1"
	negotiator1 := func(ctx context.Context, conn Conn) (context.Context, Conn, error) {
		return context.Background(), &MockConn{}, nil
	}
	err := suite.fabric.AddNegotiatorFunc(name1, negotiator1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.negotiators, 1)

	name2 := "negotiator2"
	negotiator2 := func(ctx context.Context, conn Conn) (context.Context, Conn, error) {
		return context.Background(), &MockConn{}, nil
	}
	err = suite.fabric.AddNegotiatorFunc(name2, negotiator2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.negotiators, 2)
}

func (suite *FabricTestSuite) TestGetAddressesSuccess() {
	transport1 := &MockTransport{}
	addresses1 := []string{
		"tr1.addr1",
		"tr1.addr2",
	}
	transport1.On("Addresses").Return(addresses1)
	err := suite.fabric.AddTransport(transport1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport1, suite.fabric.transports[0])

	transport2 := &MockTransport{}
	addresses2 := []string{
		"tr2.addr1",
		"tr2.addr2",
	}
	transport2.On("Addresses").Return(addresses2)
	err = suite.fabric.AddTransport(transport2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 2)
	suite.Assert().Equal(transport2, suite.fabric.transports[1])

	addressesAll := append(addresses1, addresses2...)

	addresses := suite.fabric.GetAddresses()
	suite.Assert().Len(addresses, 4)
	suite.Assert().Equal(addressesAll, addresses)
}

func TestFabricTestSuite(t *testing.T) {
	suite.Run(t, new(FabricTestSuite))
}
