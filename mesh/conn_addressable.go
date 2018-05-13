package mesh

import (
	"net"
)

type addressableConn struct {
	net.Conn
	localPeerAddress  peerAddress
	remotePeerAddress peerAddress
}

func NewAddressableConn(conn net.Conn, local, remote peerAddress) net.Conn {
	return &addressableConn{
		conn,
		local,
		remote,
	}
}

type peerAddress struct {
	network string
	peerID  string
}

func (a peerAddress) Network() string {
	return a.network
}

func (a peerAddress) String() string {
	return a.peerID
}

// LocalAddr returns the local network
// The Addr returned is shared by all invocations of LocalAddr, so
// do not modify it.
func (c *addressableConn) LocalAddr() net.Addr {
	return c.localPeerAddress
}

// RemoteAddr returns the remote network
// The Addr returned is shared by all invocations of RemoteAddr, so
// do not modify it.
func (c *addressableConn) RemoteAddr() net.Addr {
	return c.remotePeerAddress
}
