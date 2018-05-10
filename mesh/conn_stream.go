package mesh

import (
	"io"
	"net"
)

type streamConn struct {
	net.Conn
	stream io.ReadWriteCloser
}

func NewStreamConn(conn net.Conn, stream io.ReadWriteCloser) net.Conn {
	return &streamConn{
		conn,
		stream,
	}
}

// Read implements the Conn Read method.
func (c *streamConn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

// Write implements the Conn Write method.
func (c *streamConn) Write(b []byte) (int, error) {
	return c.stream.Write(b)
}

// Close closes the connection.
func (c *streamConn) Close() error {
	return c.stream.Close()
}
