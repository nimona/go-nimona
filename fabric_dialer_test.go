package fabric

// Basic imports
import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// FabricDialerTestSuite -
type FabricDialerTestSuite struct {
	suite.Suite
	fabric *Fabric
}

func (suite *FabricDialerTestSuite) SetupTest() {
	suite.fabric = New()
}

func (suite *FabricDialerTestSuite) TestDialContextSuccess() {
	transport := &MockTransport{}
	err := suite.fabric.AddTransport(transport)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport, suite.fabric.transports[0])

	ctx := context.Background()
	mockConn := &MockConn{}
	addrString := "some-address"
	addr := NewAddress(addrString)
	transport.On("DialContext", mock.Anything, addr).Return(mockConn, nil)
	transport.On("CanDial", addr).Return(true, nil)
	retCtx, retConn, retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().NotNil(retCtx)
	suite.Assert().Equal(mockConn, retConn.(*conn).conn)
	suite.Assert().Nil(retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, addr)
}

func (suite *FabricDialerTestSuite) TestDialTransportSuccess() {
	transport := &MockTransport{}
	err := suite.fabric.AddTransport(transport)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport, suite.fabric.transports[0])

	ctx := context.Background()
	mockConn := &MockConn{}
	addrString := "some-address"
	addr := NewAddress(addrString)
	transport.On("DialContext", mock.Anything, addr).Return(mockConn, nil)
	transport.On("CanDial", addr).Return(true, nil)
	retConn, retErr := suite.fabric.dialTransport(ctx, addr)
	suite.Assert().Equal(mockConn, retConn.(*conn).conn)
	suite.Assert().Nil(retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, addr)
}

func (suite *FabricDialerTestSuite) TestDialTransportFails() {
	transport := &MockTransport{}
	err := suite.fabric.AddTransport(transport)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport, suite.fabric.transports[0])

	ctx := context.Background()
	addrString := "some-address"
	addr := NewAddress(addrString)
	transport.On("DialContext", mock.Anything, addr).Return(nil, errors.New("Random error"))
	transport.On("CanDial", addr).Return(true, nil)
	retConn, retErr := suite.fabric.dialTransport(ctx, addr)
	suite.Assert().Nil(retConn)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, addr)
}

func (suite *FabricDialerTestSuite) TestGetTransportSuccess() {
	transport := &MockTransport{}
	err := suite.fabric.AddTransport(transport)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport, suite.fabric.transports[0])

	addrString := "some-address"
	addr := NewAddress(addrString)
	transport.On("CanDial", addr).Return(true, nil)
	retTransport, retErr := suite.fabric.getTransport(addr)
	suite.Assert().Equal(transport, retTransport)
	suite.Assert().Nil(retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
}

func (suite *FabricDialerTestSuite) TestGetTransportError() {
	transport := &MockTransport{}
	err := suite.fabric.AddTransport(transport)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport, suite.fabric.transports[0])

	addrString := "some-address"
	addr := NewAddress(addrString)
	transport.On("CanDial", addr).Return(false, errors.New("Random error"))
	retTransport, retErr := suite.fabric.getTransport(addr)
	suite.Assert().Nil(retTransport)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
}

func (suite *FabricDialerTestSuite) TestDialContextFails() {
	transport := &MockTransport{}
	err := suite.fabric.AddTransport(transport)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport, suite.fabric.transports[0])

	ctx := context.Background()
	addrString := "some-address"
	addr := NewAddress(addrString)
	transport.On("CanDial", addr).Return(false, nil)
	retCtx, retConn, retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Nil(retCtx)
	suite.Assert().Nil(retConn)
	suite.Assert().NotNil(retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertNumberOfCalls(suite.T(), "DialContext", 0)
}

func TestFabricDialerTestSuite(t *testing.T) {
	suite.Run(t, new(FabricDialerTestSuite))
}
