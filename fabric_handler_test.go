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
	protocolString := "foo"

	protocol := &MockProtocol{}
	protocol.On("Name").Return(protocolString)

	suite.fabric = New(protocol)

	ctx := context.Background()
	addr := NewAddress(protocolString)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	protocol.On("Handle", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retErr := suite.fabric.Handle(ctx, mockConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(1, addr.index)
	protocol.AssertCalled(suite.T(), "Handle", mock.Anything, mock.Anything)
}

func (suite *FabricHandlerTestSuite) TestHandleMultipleSuccess() {
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
	protocolFoo.On("Handle", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	protocolBar.On("Handle", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retErr := suite.fabric.Handle(ctx, mockConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(2, addr.index)
	protocolFoo.AssertCalled(suite.T(), "Handle", mock.Anything, mock.Anything)
	protocolBar.AssertCalled(suite.T(), "Handle", mock.Anything, mock.Anything)
}

func TestFabricHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(FabricHandlerTestSuite))
}
