package fabric

import (
	"net"
	"time"
)

func newConnWrapper(c net.Conn, addr *Address) Conn {
	// TODO copy address
	return &conn{
		conn:    c,
		address: addr,
	}
}

type conn struct {
	conn    net.Conn
	address *Address
}

// Conn is a generic stream-oriented network connection.
//
// Multiple goroutines may invoke methods on a Conn simultaneously.
type Conn interface {
	// Read reads data from the connection.
	// Read can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetReadDeadline.
	Read(b []byte) (n int, err error)

	// Write writes data to the connection.
	// Write can be made to time out and return an Error with Timeout() == true
	// after a fixed time limit; see SetDeadline and SetWriteDeadline.
	Write(b []byte) (n int, err error)

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// LocalAddr returns the local network address.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr

	// SetDeadline sets the read and write deadlines associated
	// with the connection. It is equivalent to calling both
	// SetReadDeadline and SetWriteDeadline.
	//
	// A deadline is an absolute time after which I/O operations
	// fail with a timeout (see type Error) instead of
	// blocking. The deadline applies to all future and pending
	// I/O, not just the immediately following call to Read or
	// Write. After a deadline has been exceeded, the connection
	// can be refreshed by setting a deadline in the future.
	//
	// An idle timeout can be implemented by repeatedly extending
	// the deadline after successful Read or Write calls.
	//
	// A zero value for t means I/O operations will not time out.
	SetDeadline(t time.Time) error

	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error

	GetAddress() *Address
}

// Read implements the Conn Read method.
func (c *conn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

// Write implements the Conn Write method.
func (c *conn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

// Close closes the connection.
func (c *conn) Close() error {
	return c.conn.Close()
}

// LocalAddr returns the local network address.
// The Addr returned is shared by all invocations of LocalAddr, so
// do not modify it.
func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
// The Addr returned is shared by all invocations of RemoteAddr, so
// do not modify it.
func (c *conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline implements the Conn SetDeadline method.
func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *conn) GetAddress() *Address {
	return c.address
}
