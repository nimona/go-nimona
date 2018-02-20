package fabric

// Basic imports
import (
	"errors"
	"net"
	"testing"
	"time"

	mock "github.com/stretchr/testify/mock"
	suite "github.com/stretchr/testify/suite"
)

// ConnTestSuite -
type ConnTestSuite struct {
	suite.Suite
	conn     *conn
	mockConn *MockConn
}

func (suite *ConnTestSuite) SetupTest() {
	suite.mockConn = &MockConn{}
	suite.conn = &conn{
		conn: suite.mockConn,
	}
}

func (suite *ConnTestSuite) TestNewConnWrapper() {
	addr := NewAddress("foo/bar")
	cn := NewConnWrapper(suite.mockConn, addr).(*conn)
	suite.Assert().Equal(addr, cn.address)
	suite.Assert().Equal(suite.mockConn, cn.conn)
}

func (suite *ConnTestSuite) TestRead() {
	bs := []byte{}
	n := 12
	err := errors.New("some-test")
	suite.mockConn.On("Read", bs).Return(n, err)
	retN, retErr := suite.conn.Read(bs)
	suite.Assert().Equal(n, retN)
	suite.Assert().Equal(err, retErr)
	suite.mockConn.AssertCalled(suite.T(), "Read", bs)
}

func (suite *ConnTestSuite) TestWrite() {
	bs := []byte{}
	n := 12
	err := errors.New("some-test")
	suite.mockConn.On("Write", bs).Return(n, err)
	retN, retErr := suite.conn.Write(bs)
	suite.Assert().Equal(n, retN)
	suite.Assert().Equal(err, retErr)
	suite.mockConn.AssertCalled(suite.T(), "Write", bs)
}

func (suite *ConnTestSuite) TestClose() {
	err := errors.New("some-test")
	suite.mockConn.On("Close").Return(err)
	retErr := suite.conn.Close()
	suite.Assert().Equal(err, retErr)
	suite.mockConn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnTestSuite) TestLocalAddr() {
	addr := &net.IPNet{}
	suite.mockConn.On("LocalAddr").Return(addr)
	retAddr := suite.conn.LocalAddr()
	suite.Assert().Equal(addr, retAddr)
	suite.mockConn.AssertCalled(suite.T(), "LocalAddr")
}

func (suite *ConnTestSuite) TestRemoteAddr() {
	addr := &net.IPNet{}
	suite.mockConn.On("RemoteAddr").Return(addr)
	retAddr := suite.conn.RemoteAddr()
	suite.Assert().Equal(addr, retAddr)
	suite.mockConn.AssertCalled(suite.T(), "RemoteAddr")
}

func (suite *ConnTestSuite) TestReadDeadline() {
	dl := time.Now()
	err := errors.New("some-test")
	suite.mockConn.On("SetDeadline", dl).Return(err)
	retErr := suite.conn.SetDeadline(dl)
	suite.Assert().Equal(err, retErr)
	suite.mockConn.AssertCalled(suite.T(), "SetDeadline", dl)
}

func (suite *ConnTestSuite) TestSetReadDeadline() {
	dl := time.Now()
	err := errors.New("some-test")
	suite.mockConn.On("SetReadDeadline", dl).Return(err)
	retErr := suite.conn.SetReadDeadline(dl)
	suite.Assert().Equal(err, retErr)
	suite.mockConn.AssertCalled(suite.T(), "SetReadDeadline", dl)
}

func (suite *ConnTestSuite) TestSetWriteDeadline() {
	dl := time.Now()
	err := errors.New("some-test")
	suite.mockConn.On("SetWriteDeadline", dl).Return(err)
	retErr := suite.conn.SetWriteDeadline(dl)
	suite.Assert().Equal(err, retErr)
	suite.mockConn.AssertCalled(suite.T(), "SetWriteDeadline", dl)
}

func (suite *ConnTestSuite) TestWriteToken() {
	// TODO check what is written
	err := errors.New("some-test")
	suite.mockConn.On("Write", mock.Anything).Return(0, err)
	retErr := suite.conn.WriteToken([]byte(""))
	suite.Assert().Equal(err, retErr)
}

func (suite *ConnTestSuite) TestReadToken() {
	// TODO check what is read
	err := errors.New("some-test")
	suite.mockConn.On("Read", mock.Anything).Return(0, err)
	bs, retErr := suite.conn.ReadToken()
	suite.Assert().Empty(bs)
	suite.Assert().Equal(err, retErr)
}

func (suite *ConnTestSuite) TestGetAddress() {
	addr := NewAddress("")
	suite.conn.address = addr
	retAddr := suite.conn.GetAddress()
	suite.Assert().Equal(addr, retAddr)
}

func TestConnTestSuite(t *testing.T) {
	suite.Run(t, new(ConnTestSuite))
}
