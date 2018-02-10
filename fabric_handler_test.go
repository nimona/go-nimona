package fabric

// Basic imports
import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// FabricHandlerTestSuite -
type FabricHandlerTestSuite struct {
	suite.Suite
	fabric *Fabric
}

func (suite *FabricHandlerTestSuite) SetupTest() {
	suite.fabric = New()
}

func (suite *FabricHandlerTestSuite) TestHandleSuccess() {
	protocol := "foo"
	suite.fabric.base = []string{protocol}

	handler := &MockHandler{}
	handler.On("Name").Return(protocol)
	err := suite.fabric.AddHandler(handler)
	suite.Assert().Nil(err)

	ctx := context.Background()
	addr := NewAddress(protocol)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	handler.On("Handle", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retErr := suite.fabric.Handle(ctx, mockConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(1, addr.index)
	handler.AssertCalled(suite.T(), "Handle", mock.Anything, mock.Anything)
}

func (suite *FabricHandlerTestSuite) TestHandleMultipleSuccess() {
	protocolFoo := "foo"
	protocolBar := "bar"
	suite.fabric.base = []string{
		protocolFoo,
		protocolBar,
	}

	handlerFoo := &MockHandler{}
	handlerFoo.On("Name").Return(protocolFoo)
	err := suite.fabric.AddHandler(handlerFoo)
	suite.Assert().Nil(err)

	handlerBar := &MockHandler{}
	handlerBar.On("Name").Return(protocolBar)
	err = suite.fabric.AddHandler(handlerBar)
	suite.Assert().Nil(err)

	ctx := context.Background()
	addr := NewAddress(protocolFoo + "/" + protocolBar)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	handlerFoo.On("Handle", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	handlerBar.On("Handle", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retErr := suite.fabric.Handle(ctx, mockConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(2, addr.index)
	handlerFoo.AssertCalled(suite.T(), "Handle", mock.Anything, mock.Anything)
	handlerBar.AssertCalled(suite.T(), "Handle", mock.Anything, mock.Anything)
}

func (suite *FabricHandlerTestSuite) TestHandleFail() {
	protocol := "foo"
	baseProtocol := "not-foo"
	suite.fabric.base = []string{baseProtocol}

	handler := &MockHandler{}
	handler.On("Name").Return(protocol)
	err := suite.fabric.AddHandler(handler)
	suite.Assert().Nil(err)

	ctx := context.Background()
	addr := NewAddress(protocol)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	retErr := suite.fabric.Handle(ctx, mockConn)
	suite.Assert().Equal(ErrInvalidProtocol, retErr)
}

func TestFabricHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FabricHandlerTestSuite))
}
