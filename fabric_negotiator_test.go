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
	protocol := "foo"

	negotiator := &MockNegotiator{}
	negotiator.On("Name").Return(protocol)
	err := suite.fabric.AddNegotiator(negotiator)
	suite.Assert().Nil(err)

	ctx := context.Background()
	addr := NewAddress(protocol)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	negotiator.On("Negotiate", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Equal(ctx, retCtx)
	suite.Assert().Equal(mockConn, retConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(1, addr.index)
	negotiator.AssertCalled(suite.T(), "Negotiate", mock.Anything, mock.Anything)
}

func (suite *FabricNegotiatorTestSuite) TestNegotiateMultipleSuccess() {
	protocolFoo := "foo"
	protocolBar := "bar"

	negotiatorFoo := &MockNegotiator{}
	negotiatorFoo.On("Name").Return(protocolFoo)
	err := suite.fabric.AddNegotiator(negotiatorFoo)
	suite.Assert().Nil(err)

	negotiatorBar := &MockNegotiator{}
	negotiatorBar.On("Name").Return(protocolBar)
	err = suite.fabric.AddNegotiator(negotiatorBar)
	suite.Assert().Nil(err)

	ctx := context.Background()
	addr := NewAddress(protocolFoo + "/" + protocolBar)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	negotiatorFoo.On("Negotiate", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	negotiatorBar.On("Negotiate", mock.Anything, mock.Anything).Return(ctx, mockConn, nil)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Equal(ctx, retCtx)
	suite.Assert().Equal(mockConn, retConn)
	suite.Assert().Nil(retErr)
	suite.Assert().Equal(2, addr.index)
	negotiatorFoo.AssertCalled(suite.T(), "Negotiate", mock.Anything, mock.Anything)
	negotiatorBar.AssertCalled(suite.T(), "Negotiate", mock.Anything, mock.Anything)
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
	protocol := "foo"

	negotiator := &MockNegotiator{}
	negotiator.On("Name").Return(protocol)
	err := suite.fabric.AddNegotiator(negotiator)
	suite.Assert().Nil(err)

	ctx := context.Background()
	addr := NewAddress(protocol)
	mockConn := &MockConn{}
	mockConn.On("GetAddress").Return(&addr)
	someErr := errors.New("some error")
	negotiator.On("Negotiate", mock.Anything, mock.Anything).Return(nil, nil, someErr)
	retCtx, retConn, retErr := suite.fabric.Negotiate(ctx, mockConn)
	suite.Assert().Nil(retCtx)
	suite.Assert().Nil(retConn)
	suite.Assert().Equal(someErr, retErr)
}

func TestFabricNegotiatorTestSuite(t *testing.T) {
	suite.Run(t, new(FabricNegotiatorTestSuite))
}
