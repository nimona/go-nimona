package mesh

import (
	"fmt"
	"net"
)

type ID struct {
}

func (id *ID) Initiate(conn net.Conn) (net.Conn, error) {
	fmt.Println("> ID")
	localPeerID := conn.LocalAddr().String()
	if err := WriteToken(conn, []byte(localPeerID)); err != nil {
		return nil, err
	}
	remotePeerID, err := ReadToken(conn)
	if err != nil {
		return nil, err
	}
	localAddress := peerAddress{
		network: conn.LocalAddr().Network(),
		peerID:  conn.LocalAddr().String(),
	}
	remoteAddress := peerAddress{
		network: conn.RemoteAddr().Network(),
		peerID:  string(remotePeerID),
	}
	newConn := NewAddressableConn(conn, localAddress, remoteAddress)
	return newConn, nil
}

func (id *ID) Handle(conn net.Conn) (net.Conn, error) {
	fmt.Println("< ID")
	remotePeerID, err := ReadToken(conn)
	if err != nil {
		return nil, err
	}
	localPeerID := conn.LocalAddr().String()
	if err := WriteToken(conn, []byte(localPeerID)); err != nil {
		return nil, err
	}
	localAddress := peerAddress{
		network: conn.LocalAddr().Network(),
		peerID:  conn.LocalAddr().String(),
	}
	remoteAddress := peerAddress{
		network: conn.RemoteAddr().Network(),
		peerID:  string(remotePeerID),
	}
	newConn := NewAddressableConn(conn, localAddress, remoteAddress)
	return newConn, nil
}
