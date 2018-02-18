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
	ctx    context.Context
}

func (suite *FabricDialerTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.fabric = New(suite.ctx)
}

func (suite *FabricDialerTestSuite) TestDialContextSuccess() {
	ctx := context.Background()

	protocol := &MockProtocol{}
	protocol.On("Name").Return("test")
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(protocol)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := NewAddress(addrString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)

	transport := &MockTransport{}
	transport.On("Listen", mock.Anything, mock.Anything).Return(nil)
	transport.On("CanDial", addr).Return(true, nil)
	transport.On("DialContext", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	err := suite.fabric.AddTransport(transport, []Protocol{protocol})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.CallContext(ctx, addrString)
	suite.Assert().Nil(retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func (suite *FabricDialerTestSuite) TestDialTransportCannotDial() {
	ctx := context.Background()

	addrString := "test"
	addr := NewAddress(addrString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)

	transport := &MockTransport{}
	transport.On("Listen", mock.Anything, mock.Anything).Return(nil)
	transport.On("CanDial", addr).Return(false, nil)
	err := suite.fabric.AddTransport(transport, []Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.CallContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
}

func (suite *FabricDialerTestSuite) TestDialTransportError() {
	ctx := context.Background()

	addrString := "test"
	addr := NewAddress(addrString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)

	transport := &MockTransport{}
	transport.On("Listen", mock.Anything, mock.Anything).Return(nil)
	transport.On("CanDial", addr).Return(false, errors.New("error"))
	err := suite.fabric.AddTransport(transport, []Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.CallContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
}

func (suite *FabricDialerTestSuite) TestDialContextFails() {
	ctx := context.Background()

	protocol := &MockProtocol{}
	protocol.On("Name").Return("test")
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(protocol)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := NewAddress(addrString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)

	transport := &MockTransport{}
	transport.On("Listen", mock.Anything, mock.Anything).Return(nil)
	transport.On("CanDial", addr).Return(true, nil)
	transport.On("DialContext", mock.Anything, mock.Anything).Return(nil, nil, errors.New("error"))
	err := suite.fabric.AddTransport(transport, []Protocol{protocol})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.CallContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func (suite *FabricDialerTestSuite) TestNegotiatorFails() {
	ctx := context.Background()

	protocol := &MockProtocol{}
	protocol.On("Name").Return("test")
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		return errors.New("error")
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(protocol)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := NewAddress(addrString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)

	transport := &MockTransport{}
	transport.On("Listen", mock.Anything, mock.Anything).Return(nil)
	transport.On("CanDial", addr).Return(true, nil)
	transport.On("DialContext", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	err := suite.fabric.AddTransport(transport, []Protocol{protocol})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.CallContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func (suite *FabricDialerTestSuite) TestInvalidProtocolFails() {
	ctx := context.Background()

	protocol := &MockProtocol{}
	protocol.On("Name").Return("nope")
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(protocol)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := NewAddress(addrString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)

	transport := &MockTransport{}
	transport.On("Listen", mock.Anything, mock.Anything).Return(nil)
	transport.On("CanDial", addr).Return(true, nil)
	transport.On("DialContext", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	err := suite.fabric.AddTransport(transport, []Protocol{protocol})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.CallContext(ctx, addrString)
	suite.Assert().Equal(ErrInvalidProtocol, retErr)
	transport.AssertCalled(suite.T(), "CanDial", addr)
	transport.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func TestFabricDialerTestSuite(t *testing.T) {
	suite.Run(t, new(FabricDialerTestSuite))
}
