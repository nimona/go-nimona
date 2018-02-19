package protocol

// Basic imports
import (
	"context"
	"testing"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"

	address "github.com/nimona/go-nimona-fabric/address"
	conn "github.com/nimona/go-nimona-fabric/connection"
)

// ProtocolTransportTestSuite -
type ProtocolTransportTestSuite struct {
	suite.Suite
}

func (suite *ProtocolTransportTestSuite) SetupTest() {}

func (suite *ProtocolTransportTestSuite) TestName() {
	wrp := &TransportWrapper{
		protocolNames: []string{},
	}

	name := wrp.Name()
	suite.Assert().Equal("", name)
}

func (suite *ProtocolTransportTestSuite) TestHandleSuccess() {
	wrp := &TransportWrapper{
		protocolNames: []string{},
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("test")
	var handler HandlerFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := address.NewAddress("test")
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)
	suite.Assert().Equal("test", addr.CurrentProtocol())

	ctx := context.Background()
	err := wrp.Handle(protocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	protocol.AssertCalled(suite.T(), "Handle", mock.Anything)
}

func (suite *ProtocolTransportTestSuite) TestNegotiateSuccess() {
	wrp := &TransportWrapper{
		protocolNames: []string{},
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("test")
	var handler HandlerFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c conn.Conn) error {
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := address.NewAddress("test")
	mockConn := &conn.MockConn{}
	mockConn.On("GetAddress").Return(addr)
	suite.Assert().Equal("test", addr.CurrentProtocol())

	ctx := context.Background()
	err := wrp.Negotiate(protocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
}

func TestProtocolTransportTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolTransportTestSuite))
}
