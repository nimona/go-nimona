package fabric

// Basic imports
import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TransportTCPTestSuite -
type TransportTCPTestSuite struct {
	suite.Suite
}

func (suite *TransportTCPTestSuite) SetupTest() {

}

func (suite *TransportTCPTestSuite) TestCanDial() {
	tcp := NewTransportTCP("127.0.0.1", 0)

	addrsValid := map[string]bool{
		"http:something.com:80/ping": false,
		"tcp:some-dns.com:90/ping":   true,
		"tcp:1.1.1.1:100/ping":       true,
		"/tcp:1.1.1.1:100/ping":      false,
	}

	for addr, can := range addrsValid {
		res, err := tcp.CanDial(NewAddress(addr))
		suite.Assert().Equal(can, res, "Addr %s should return %b", addr, can)
		suite.Assert().Nil(err, "Addr %s should not return error", addr)
	}

	addrsInvalid := map[string]bool{
		"tcp:some-dns.com/ping":         false,
		"tcp/some-dns.com/ping":         false,
		"tcp::/some-dns.com/ping":       false,
		"tcp:some-dns.com:9999990/ping": false,
		"tcp:some-dns.com:0/ping":       false,
		"tcp:some-dns.com:/ping":        false,
	}

	for addr, can := range addrsInvalid {
		res, err := tcp.CanDial(NewAddress(addr))
		suite.Assert().Equal(can, res, "Addr %s should return %b", addr, can)
		suite.Assert().NotNil(err, "Addr %s should not return error", addr)
	}
}

func (suite *TransportTCPTestSuite) TestListenSuccess() {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	handled := 0

	handler := func(context.Context, net.Conn) error {
		handled++
		wg.Done()
		return nil
	}

	ctx := context.Background()
	tcps := NewTransportTCP("0.0.0.0", 0)
	err := tcps.Listen(ctx, handler)
	suite.Assert().Nil(err)

	addrs := tcps.Addresses()
	suite.Assert().NotEmpty(addrs)

	tcpc := NewTransportTCP("0.0.0.0", 0)
	conn, err := tcpc.DialContext(ctx, NewAddress(addrs[0]))

	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
	}

	suite.Assert().Nil(err)
	suite.Assert().NotNil(conn)
	suite.Assert().Equal(1, handled)
}

func (suite *TransportTCPTestSuite) TestListenMultipleSuccess() {
	wg := &sync.WaitGroup{}
	wg.Add(4)

	handled := 0

	handler := func(context.Context, net.Conn) error {
		handled++
		wg.Done()
		return nil
	}

	ctx := context.Background()
	tcps := NewTransportTCP("0.0.0.0", 0)
	err := tcps.Listen(ctx, handler)
	suite.Assert().Nil(err)

	addrs := tcps.Addresses()
	suite.Assert().NotEmpty(addrs)

	tcpc := NewTransportTCP("0.0.0.0", 0)
	conn1, err1 := tcpc.DialContext(ctx, NewAddress(addrs[0]))
	conn2, err2 := tcpc.DialContext(ctx, NewAddress(addrs[0]))
	conn3, err3 := tcpc.DialContext(ctx, NewAddress(addrs[0]))
	conn4, err4 := tcpc.DialContext(ctx, NewAddress(addrs[0]))

	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 2):
	}

	suite.Assert().Nil(err1)
	suite.Assert().NotNil(conn1)
	suite.Assert().Nil(err2)
	suite.Assert().NotNil(conn2)
	suite.Assert().Nil(err3)
	suite.Assert().NotNil(conn3)
	suite.Assert().Nil(err4)
	suite.Assert().NotNil(conn4)
	suite.Assert().Equal(4, handled)
}

func TestTransportTCPTestSuite(t *testing.T) {
	suite.Run(t, new(TransportTCPTestSuite))
}
