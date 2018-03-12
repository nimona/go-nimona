package net

// Basic imports
import (
	"context"
	"errors"
	"testing"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

// ProtocolIdentityTestSuite -
type ProtocolIdentityTestSuite struct {
	suite.Suite
}

func (suite *ProtocolIdentityTestSuite) SetupTest() {}

func (suite *ProtocolIdentityTestSuite) TestName() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	name := identity.Name()
	suite.Assert().Equal("identity", name)
}

func (suite *ProtocolIdentityTestSuite) TestHandleSuccess() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := NewAddress("identity")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("remote"), nil)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(nil)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Handle(protocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	protocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().True(handlerCalled)
}

func (suite *ProtocolIdentityTestSuite) TestHandleReadTokenFails() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	retErr := errors.New("error")
	addr := NewAddress("identity")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("remote"), retErr)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Handle(protocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Equal(retErr, err)
	protocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(handlerCalled)
}

func (suite *ProtocolIdentityTestSuite) TestHandleWriteTokenFails() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := NewAddress("identity")
	mockConn := &MockConn{}
	retError := errors.New("error")
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("remote"), nil)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(retError)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Handle(protocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Equal(retError, err)
	protocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(handlerCalled)
}

func (suite *ProtocolIdentityTestSuite) TestNegotiateSuccess() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := NewAddress("identity")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(nil)
	mockConn.On("ReadToken").Return([]byte("remote"), nil)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Negotiate(protocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().True(negotiatorCalled)
}

func (suite *ProtocolIdentityTestSuite) TestNegotiateReadTokenFails() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	retErr := errors.New("error")
	addr := NewAddress("identity")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(nil)
	mockConn.On("ReadToken").Return([]byte("remote"), retErr)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Negotiate(protocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Equal(retErr, err)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().False(handlerCalled)
}

func (suite *ProtocolIdentityTestSuite) TestNegotiateWriteTokenFails() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := NewAddress("identity")
	mockConn := &MockConn{}
	retError := errors.New("error")
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(retError)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Negotiate(protocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Equal(retError, err)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().False(handlerCalled)
}

func (suite *ProtocolIdentityTestSuite) TestNegotiateCheckRemote() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := NewAddress("identity:remote")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(nil)
	mockConn.On("ReadToken").Return([]byte("remote"), nil)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Negotiate(protocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().True(negotiatorCalled)
}

func (suite *ProtocolIdentityTestSuite) TestNegotiateUnexpectedRemote() {
	identity := &IdentityProtocol{
		Local: "local",
	}

	protocol := &MockProtocol{}
	protocol.On("Name").Return("identity")
	handlerCalled := false
	negotiatorCalled := false
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		negotiatorCalled = true
		return nil
	}
	protocol.On("Handle", mock.Anything).Return(handler)
	protocol.On("Negotiate", mock.Anything).Return(negotiator)

	addr := NewAddress("identity:remote")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte(identity.Local)).Return(nil)
	mockConn.On("ReadToken").Return([]byte("not-remote"), nil)
	suite.Assert().Equal("identity", addr.CurrentProtocol())

	ctx := context.Background()
	err := identity.Negotiate(protocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Equal(ErrUnexpectedRemote, err)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().False(negotiatorCalled)
}

func TestProtocolIdentityTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolIdentityTestSuite))
}
