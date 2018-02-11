package fabric

// Basic imports
import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// FabricNegotiatorTestSuite -
type FabricNegotiatorTestSuite struct {
	suite.Suite
	fabric *Fabric
}

func (suite *FabricNegotiatorTestSuite) SetupTest() {
	suite.fabric = New()
}

func (suite *FabricNegotiatorTestSuite) TestNegotiateSuccess() {
	protocolString := "foo"

	protocol := &MockProtocol{}
	protocol.On("Name").Return(protocolString)

	suite.fabric = New(protocol)

	ctx := context.Background()
	addr := NewAddress(protocolString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	protocol.On("Negotiate", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Equal(ctx, retCtx)
	suite.Assert().Equal(mockConn, retConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(1, addr.index)
	protocol.AssertCalled(suite.T(), "Negotiate", mock.Anything, mock.Anything)
}

func (suite *FabricNegotiatorTestSuite) TestNegotiateMultipleSuccess() {
	protocolFooString := "foo"
	protocolBarString := "bar"

	protocolFoo := &MockProtocol{}
	protocolFoo.On("Name").Return(protocolFooString)

	protocolBar := &MockProtocol{}
	protocolBar.On("Name").Return(protocolBarString)

	suite.fabric = New(protocolFoo, protocolBar)

	ctx := context.Background()
	addr := NewAddress(protocolFooString + "/" + protocolBarString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	protocolFoo.On("Negotiate", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	protocolBar.On("Negotiate", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Equal(ctx, retCtx)
	suite.Assert().Equal(mockConn, retConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(2, addr.index)
	protocolFoo.AssertCalled(suite.T(), "Negotiate", mock.Anything, mock.Anything)
	protocolBar.AssertCalled(suite.T(), "Negotiate", mock.Anything, mock.Anything)
}

func (suite *FabricNegotiatorTestSuite) TestNegotiateEmptyFail() {
	protocol := "foo"
	ctx := context.Background()
	addr := NewAddress(protocol)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Equal(ctx, retCtx)
	suite.Assert().Equal(mockConn, retConn)
	suite.Assert().Equal(errNoMoreProtocols, retErr)
}

func (suite *FabricNegotiatorTestSuite) TestNegotiateFail() {
	protocolString := "foo"

	protocol := &MockProtocol{}
	protocol.On("Name").Return(protocolString)

	suite.fabric = New(protocol)

	ctx := context.Background()
	addr := NewAddress(protocolString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	someErr := errors.New("some error")
	protocol.On("Negotiate", mock.Anything, mock.Anything).Return(nil, nil, someErr)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Nil(retCtx)
	suite.Assert().Nil(retConn)
	suite.Assert().Equal(someErr, retErr)
}

func TestFabricNegotiatorTestSuite(t *testing.T) {
	suite.Run(t, new(FabricNegotiatorTestSuite))
}
