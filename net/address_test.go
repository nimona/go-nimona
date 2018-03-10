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

func (suite *AddressTestSuite) TestCurrentParamsEmptySuccess() {
	addrString := "foo"
	addr := NewAddress(addrString)

	params := addr.CurrentParams()
	suite.Assert().Equal("", params)
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

func (suite *AddressTestSuite) TestAppendSuccess() {
	addrString := "foo"
	addr := NewAddress(addrString)
	suite.Assert().Len(addr.stack, 1)
	suite.Assert().Equal(addr.stack[0], "foo")

	addr.Append("bar")
	suite.Assert().Len(addr.stack, 2)
	suite.Assert().Equal(addr.stack[1], "bar")
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

func (suite *AddressTestSuite) TestProcessedSuccess() {
	addrString := "foo/bar/more"
	addr := NewAddress(addrString)

	part := addr.Pop()
	suite.Assert().Equal("foo", part)
	suite.Assert().Equal(1, addr.index)

	processed := addr.Processed()
	suite.Assert().Equal([]string{"foo"}, processed)
	suite.Assert().Equal(1, addr.index)

	addr.Pop()
	processed = addr.Processed()
	suite.Assert().Equal([]string{"foo", "bar"}, processed)
	suite.Assert().Equal(2, addr.index)
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

func (suite *AddressTestSuite) TestRemainingProtocolsSuccess() {
	addrString := "foo:param/bar:param/more:param"
	addr := NewAddress(addrString)

	remaining := addr.RemainingProtocols()
	suite.Assert().Equal([]string{"foo", "bar", "more"}, remaining)
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

func (suite *AddressTestSuite) TestStringSuccess() {
	addrString := "foo/bar"
	addr := NewAddress(addrString)

	retAddr := addr.String()
	suite.Assert().Equal(addrString, retAddr)
}

func TestAddressTestSuite(t *testing.T) {
	suite.Run(t, new(AddressTestSuite))
}
