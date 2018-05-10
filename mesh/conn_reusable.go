package mesh

import (
	"net"
)

type reusableConn struct {
	net.Conn
	newConnFn    func() (net.Conn, error)
	acceptConnFn func() (net.Conn, error)
}

func NewReusableConn(conn net.Conn, newConnFn func() (net.Conn, error), acceptConnFn func() (net.Conn, error)) net.Conn {
	return &reusableConn{
		conn,
		newConnFn,
		acceptConnFn,
	}
}

func (c *reusableConn) NewConn() (net.Conn, error) {
	return c.newConnFn()
}

func (c *reusableConn) Accepted(accepted chan net.Conn) error {
	for {
		conn, err := c.acceptConnFn()
		if err != nil {
			// TODO should we return?
			continue
		}
		accepted <- conn
	}
}
