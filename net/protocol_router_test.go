package net

// Basic imports
import (
	"context"
	"errors"
	"testing"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

// ProtocolRouterTestSuite -
type ProtocolRouterTestSuite struct {
	suite.Suite
	router           *RouterProtocol
	nextProtocol     *MockProtocol
	handlerCalled    bool
	negotiatorCalled bool
}

func (suite *ProtocolRouterTestSuite) SetupTest() {
	suite.nextProtocol = &MockProtocol{}
	suite.nextProtocol.On("Name").Return("test")
	var handler HandlerFunc = func(ctx context.Context, c Conn) error {
		suite.handlerCalled = true
		return nil
	}
	var negotiator NegotiatorFunc = func(ctx context.Context, c Conn) error {
		suite.negotiatorCalled = true
		return nil
	}
	suite.nextProtocol.On("Handle", mock.Anything).Return(handler)
	suite.nextProtocol.On("Negotiate", mock.Anything).Return(negotiator)

	suite.handlerCalled = false
	suite.negotiatorCalled = false
	suite.router = &RouterProtocol{
		routes: map[string][]Protocol{
			"test": []Protocol{
				suite.nextProtocol,
			},
		},
	}
}

func (suite *ProtocolRouterTestSuite) TestName() {
	name := suite.router.Name()
	suite.Assert().Equal("router", name)
}

func (suite *ProtocolRouterTestSuite) TestNew() {
	router := NewRouter()
	suite.Assert().Equal(&RouterProtocol{routes: map[string][]Protocol{}}, router)
}

func (suite *ProtocolRouterTestSuite) TestAddRoute() {
	protocol1 := &MockProtocol{}
	protocol1.On("Name").Return("test1")

	protocol2 := &MockProtocol{}
	protocol2.On("Name").Return("test2")

	router := NewRouter()
	router.AddRoute(protocol1)
	suite.Assert().Len(router.routes, 1)
	router.AddRoute(protocol1)
	suite.Assert().Len(router.routes, 1)
	router.AddRoute(protocol2)
	suite.Assert().Len(router.routes, 2)
	router.AddRoute(protocol1, protocol2)
	suite.Assert().Len(router.routes, 3)
	router.AddRoute(protocol2, protocol1)
	suite.Assert().Len(router.routes, 4)
}

func (suite *ProtocolRouterTestSuite) TestHandleSuccess() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("SEL test"), nil)
	mockConn.On("WriteToken", []byte("ACK test")).Return(nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Handle(suite.nextProtocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	suite.nextProtocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().True(suite.handlerCalled)
}

func (suite *ProtocolRouterTestSuite) TestHandleReadTokenError() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	retErr := errors.New("Error")
	mockConn.On("ReadToken").Return([]byte(""), retErr)
	mockConn.On("WriteToken", []byte("ACK test")).Return(nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Handle(suite.nextProtocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Equal(retErr, err)
	suite.nextProtocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(suite.handlerCalled)
}

func (suite *ProtocolRouterTestSuite) TestHandleWriteTokenError() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	retErr := errors.New("Error")
	mockConn.On("ReadToken").Return([]byte("SEL test"), nil)
	mockConn.On("WriteToken", []byte("ACK test")).Return(retErr)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Handle(suite.nextProtocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Equal(retErr, err)
	suite.nextProtocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(suite.handlerCalled)
}

func (suite *ProtocolRouterTestSuite) TestHandleInvalidCommandJunk() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("asdf"), nil)
	mockConn.On("Close").Return(nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Handle(suite.nextProtocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Equal(ErrInvalidCommand, err)
	suite.nextProtocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(suite.handlerCalled)
}

func (suite *ProtocolRouterTestSuite) TestHandleInvalidCommand() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("ASD something"), nil)
	mockConn.On("Close").Return(nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Handle(suite.nextProtocol.Handle(nil))(ctx, mockConn)
	suite.Assert().Equal(ErrInvalidCommand, err)
	suite.nextProtocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(suite.handlerCalled)
}

func (suite *ProtocolRouterTestSuite) TestHandleInvalidRoute() {
	addr := NewAddress("router/not-test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("ReadToken").Return([]byte("SEL not-test"), nil)
	mockConn.On("WriteToken", []byte("ACK test")).Return(nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Handle(suite.nextProtocol.Handle(nil))(ctx, mockConn)
	suite.Assert().NotNil(err)
	suite.nextProtocol.AssertCalled(suite.T(), "Handle", mock.Anything)
	suite.Assert().False(suite.handlerCalled)
}

func (suite *ProtocolRouterTestSuite) TestNegotiateSuccess() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte("SEL test")).Return(nil)
	mockConn.On("ReadToken").Return([]byte("ACK test"), nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Negotiate(suite.nextProtocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Nil(err)
	suite.nextProtocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().True(suite.negotiatorCalled)
}

func (suite *ProtocolRouterTestSuite) TestNegotiateWriteTokenFails() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	retErr := errors.New("error")
	mockConn.On("WriteToken", []byte("SEL test")).Return(retErr)
	mockConn.On("ReadToken").Return([]byte("ACK test"), nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Negotiate(suite.nextProtocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Equal(retErr, err)
	suite.nextProtocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().False(suite.negotiatorCalled)
}

func (suite *ProtocolRouterTestSuite) TestNegotiateReadTokenFails() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	retErr := errors.New("error")
	mockConn.On("WriteToken", []byte("SEL test")).Return(nil)
	mockConn.On("ReadToken").Return([]byte(""), retErr)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Negotiate(suite.nextProtocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().Equal(retErr, err)
	suite.nextProtocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().False(suite.negotiatorCalled)
}

func (suite *ProtocolRouterTestSuite) TestNegotiateInvalidResponseFails() {
	addr := NewAddress("router/test")
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(addr)
	mockConn.On("WriteToken", []byte("SEL test")).Return(nil)
	mockConn.On("ReadToken").Return([]byte("ACK asdf"), nil)
	suite.Assert().Equal("router", addr.CurrentProtocol())

	ctx := context.Background()
	err := suite.router.Negotiate(suite.nextProtocol.Negotiate(nil))(ctx, mockConn)
	suite.Assert().NotNil(err)
	suite.nextProtocol.AssertCalled(suite.T(), "Negotiate", mock.Anything)
	suite.Assert().False(suite.negotiatorCalled)
}

func TestProtocolRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolRouterTestSuite))
}
