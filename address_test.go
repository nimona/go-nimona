package fabric

// Basic imports
import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// AddressTestSuite -
type AddressTestSuite struct {
	suite.Suite
}

func (suite *AddressTestSuite) SetupTest() {
}

func (suite *AddressTestSuite) TestNewAddressSuccess() {
	addrString := "foo/bar"
	addr := NewAddress(addrString)
	suite.Assert().Equal([]string{
		"foo",
		"bar",
	}, addr.stack)
	suite.Assert().Equal(addrString, addr.original)
	suite.Assert().Equal(0, addr.index)
}

func (suite *AddressTestSuite) TestCurrentSuccess() {
	addrString := "foo/bar"
	addr := NewAddress(addrString)

	part := addr.Pop()
	suite.Assert().Equal("foo", part)
	suite.Assert().Equal(1, addr.index)

	part = addr.Current()
	suite.Assert().Equal("bar", part)
	suite.Assert().Equal(1, addr.index)
}

func (suite *AddressTestSuite) TestCurrentParamsSuccess() {
	addrString := "foo:bar"
	addr := NewAddress(addrString)

	params := addr.CurrentParams()
	suite.Assert().Equal("bar", params)
	suite.Assert().Equal(0, addr.index)
}

func (suite *AddressTestSuite) TestCurrentProtocolSuccess() {
	addrString := "foo:bar"
	addr := NewAddress(addrString)

	protocol := addr.CurrentProtocol()
	suite.Assert().Equal("foo", protocol)
	suite.Assert().Equal(0, addr.index)
}

func (suite *AddressTestSuite) TestPopSuccess() {
	addrString := "foo/bar"
	addr := NewAddress(addrString)

	part := addr.Pop()
	suite.Assert().Equal("foo", part)
	suite.Assert().Equal(1, addr.index)

	part = addr.Pop()
	suite.Assert().Equal("bar", part)
	suite.Assert().Equal(2, addr.index)

	part = addr.Pop()
	suite.Assert().Equal("", part)
	suite.Assert().Equal(2, addr.index)
}

func (suite *AddressTestSuite) TestRemainingSuccess() {
	addrString := "foo/bar/more"
	addr := NewAddress(addrString)

	part := addr.Pop()
	suite.Assert().Equal("foo", part)
	suite.Assert().Equal(1, addr.index)

	remaining := addr.Remaining()
	suite.Assert().Equal([]string{"bar", "more"}, remaining)
	suite.Assert().Equal(1, addr.index)
}

func (suite *AddressTestSuite) TestRemainingStringSuccess() {
	addrString := "foo/bar/more"
	addr := NewAddress(addrString)

	part := addr.Pop()
	suite.Assert().Equal("foo", part)
	suite.Assert().Equal(1, addr.index)

	remaining := addr.RemainingString()
	suite.Assert().Equal("bar/more", remaining)
	suite.Assert().Equal(1, addr.index)
}

func (suite *AddressTestSuite) TestResetSuccess() {
	addrString := "foo/bar"
	addr := NewAddress(addrString)

	part := addr.Pop()
	suite.Assert().Equal("foo", part)
	suite.Assert().Equal(1, addr.index)

	addr.Reset()
	suite.Assert().Equal(0, addr.index)
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(AddressTestSuite))
}
