package fabric

// Basic imports
import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// FabricTestSuite -
type FabricTestSuite struct {
	suite.Suite
	fabric *Fabric
}

func (suite *FabricTestSuite) SetupTest() {
	suite.fabric = New()
}

func (suite *FabricTestSuite) TestAddTransportSuccess() {
	transport1 := &MockTransport{}
	err := suite.fabric.AddTransport(transport1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport1, suite.fabric.transports[0])

	transport2 := &MockTransport{}
	err = suite.fabric.AddTransport(transport2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 2)
	suite.Assert().Equal(transport2, suite.fabric.transports[1])
}

func (suite *FabricTestSuite) TestAddProtocolSuccess() {
	name1 := "protocol1"
	protocol1 := &MockProtocol{}
	protocol1.On("Name").Return(name1)
	err := suite.fabric.AddProtocol(protocol1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.protocols, 1)
	protocol1.AssertCalled(suite.T(), "Name")

	name2 := "protocol2"
	protocol2 := &MockProtocol{}
	protocol2.On("Name").Return(name2)
	err = suite.fabric.AddProtocol(protocol2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.protocols, 2)
	protocol2.AssertCalled(suite.T(), "Name")
}

func (suite *FabricTestSuite) TestGetAddressesSuccess() {
	transport1 := &MockTransport{}
	addresses1 := []string{
		"tr1.addr1",
		"tr1.addr2",
	}
	transport1.On("Addresses").Return(addresses1)
	err := suite.fabric.AddTransport(transport1)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 1)
	suite.Assert().Equal(transport1, suite.fabric.transports[0])

	transport2 := &MockTransport{}
	addresses2 := []string{
		"tr2.addr1",
		"tr2.addr2",
	}
	transport2.On("Addresses").Return(addresses2)
	err = suite.fabric.AddTransport(transport2)
	suite.Assert().Nil(err)
	suite.Assert().Len(suite.fabric.transports, 2)
	suite.Assert().Equal(transport2, suite.fabric.transports[1])

	addressesAll := append(addresses1, addresses2...)

	addresses := suite.fabric.GetAddresses()
	suite.Assert().Len(addresses, 4)
	suite.Assert().Equal(addressesAll, addresses)
}

func TestFabricTestSuite(t *testing.T) {
	suite.Run(t, new(FabricTestSuite))
}
