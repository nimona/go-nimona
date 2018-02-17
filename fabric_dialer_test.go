package fabric

// Basic imports
import (
	"context"
	"errors"
	"testing"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"

	address "github.com/nimona/go-nimona-fabric/address"
	conn "github.com/nimona/go-nimona-fabric/connection"
	protocol "github.com/nimona/go-nimona-fabric/protocol"
	transport "github.com/nimona/go-nimona-fabric/transport"
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

	prot := &protocol.MockProtocol{}
	prot.On("Name").Return("test")
	var handler protocol.HandlerFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	var negotiator protocol.NegotiatorFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	prot.On("Handle", mock.Anything).Return(handler)
	prot.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(prot)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := address.NewAddress(addrString)
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)

	tran := &transport.MockTransport{}
	tran.On("Listen", mock.Anything, mock.Anything).Return(nil)
	tran.On("CanDial", addr).Return(true, nil)
	tran.On("DialContext", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	err := suite.fabric.AddTransport(tran, []protocol.Protocol{prot})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Nil(retErr)
	tran.AssertCalled(suite.T(), "CanDial", addr)
	tran.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func (suite *FabricDialerTestSuite) TestDialTransportCannotDial() {
	ctx := context.Background()

	addrString := "test"
	addr := address.NewAddress(addrString)
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)

	tran := &transport.MockTransport{}
	tran.On("Listen", mock.Anything, mock.Anything).Return(nil)
	tran.On("CanDial", addr).Return(false, nil)
	err := suite.fabric.AddTransport(tran, []protocol.Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	tran.AssertCalled(suite.T(), "CanDial", addr)
}

func (suite *FabricDialerTestSuite) TestDialTransportError() {
	ctx := context.Background()

	addrString := "test"
	addr := address.NewAddress(addrString)
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)

	tran := &transport.MockTransport{}
	tran.On("Listen", mock.Anything, mock.Anything).Return(nil)
	tran.On("CanDial", addr).Return(false, errors.New("error"))
	err := suite.fabric.AddTransport(tran, []protocol.Protocol{})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	tran.AssertCalled(suite.T(), "CanDial", addr)
}

func (suite *FabricDialerTestSuite) TestDialContextFails() {
	ctx := context.Background()

	prot := &protocol.MockProtocol{}
	prot.On("Name").Return("test")
	var handler protocol.HandlerFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	var negotiator protocol.NegotiatorFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	prot.On("Handle", mock.Anything).Return(handler)
	prot.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(prot)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := address.NewAddress(addrString)
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)

	tran := &transport.MockTransport{}
	tran.On("Listen", mock.Anything, mock.Anything).Return(nil)
	tran.On("CanDial", addr).Return(true, nil)
	tran.On("DialContext", mock.Anything, mock.Anything).Return(nil, nil, errors.New("error"))
	err := suite.fabric.AddTransport(tran, []protocol.Protocol{prot})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	tran.AssertCalled(suite.T(), "CanDial", addr)
	tran.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func (suite *FabricDialerTestSuite) TestNegotiatorFails() {
	ctx := context.Background()

	prot := &protocol.MockProtocol{}
	prot.On("Name").Return("test")
	var handler protocol.HandlerFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	var negotiator protocol.NegotiatorFunc = func(ctx context.Context, c conn.Conn) error {
		return errors.New("error")
	}
	prot.On("Handle", mock.Anything).Return(handler)
	prot.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(prot)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := address.NewAddress(addrString)
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)

	tran := &transport.MockTransport{}
	tran.On("Listen", mock.Anything, mock.Anything).Return(nil)
	tran.On("CanDial", addr).Return(true, nil)
	tran.On("DialContext", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	err := suite.fabric.AddTransport(tran, []protocol.Protocol{prot})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Equal(ErrCouldNotDial, retErr)
	tran.AssertCalled(suite.T(), "CanDial", addr)
	tran.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func (suite *FabricDialerTestSuite) TestInvalidProtocolFails() {
	ctx := context.Background()

	prot := &protocol.MockProtocol{}
	prot.On("Name").Return("nope")
	var handler protocol.HandlerFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	var negotiator protocol.NegotiatorFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	prot.On("Handle", mock.Anything).Return(handler)
	prot.On("Negotiate", mock.Anything).Return(negotiator)
	protErr := suite.fabric.AddProtocol(prot)
	suite.Assert().Nil(protErr)
	suite.Assert().Len(suite.fabric.protocols, 1)

	addrString := "test"
	addr := address.NewAddress(addrString)
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)

	tran := &transport.MockTransport{}
	tran.On("Listen", mock.Anything, mock.Anything).Return(nil)
	tran.On("CanDial", addr).Return(true, nil)
	tran.On("DialContext", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	err := suite.fabric.AddTransport(tran, []protocol.Protocol{prot})
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)

	retErr := suite.fabric.DialContext(ctx, addrString)
	suite.Assert().Equal(ErrInvalidProtocol, retErr)
	tran.AssertCalled(suite.T(), "CanDial", addr)
	tran.AssertCalled(suite.T(), "DialContext", mock.Anything, mock.Anything)
}

func TestFabricDialerTestSuite(t *testing.T) {
	suite.Run(t, new(FabricDialerTestSuite))
}
