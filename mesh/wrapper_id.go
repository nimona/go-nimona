package mesh

import (
	"encoding/json"
	"errors"
	"net"
)

type ID struct {
	registry Registry
}

func (id *ID) Initiate(conn net.Conn) (net.Conn, error) {
	// fmt.Println("> ID")
	localPeerInfoBs := id.registry.GetLocalPeerInfo().Marshal()
	if err := WriteToken(conn, localPeerInfoBs); err != nil {
		return nil, err
	}
	remotePeerInfoBs, err := ReadToken(conn)
	if err != nil {
		return nil, err
	}
	remotePeerInfo := &PeerInfo{}
	if err := json.Unmarshal(remotePeerInfoBs, &remotePeerInfo); err != nil {
		// TODO close conn?
		return nil, err
	}
	if !remotePeerInfo.IsValid() {
		return nil, errors.New("invalid peer info")
	}
	localAddress := peerAddress{
		network: conn.LocalAddr().Network(),
		peerID:  conn.LocalAddr().String(),
	}
	remoteAddress := peerAddress{
		network: conn.RemoteAddr().Network(),
		peerID:  remotePeerInfo.ID,
	}
	newConn := NewAddressableConn(conn, localAddress, remoteAddress)
	return newConn, nil
}

func (id *ID) Handle(conn net.Conn) (net.Conn, error) {
	// fmt.Println("< ID")
	remotePeerInfoBs, err := ReadToken(conn)
	if err != nil {
		return nil, err
	}
	remotePeerInfo := &PeerInfo{}
	if err := json.Unmarshal(remotePeerInfoBs, &remotePeerInfo); err != nil {
		// TODO close conn?
		return nil, err
	}
	if !remotePeerInfo.IsValid() {
		return nil, errors.New("invalid peer info")
	}
	localPeerInfoBs := id.registry.GetLocalPeerInfo().Marshal()
	if err := WriteToken(conn, localPeerInfoBs); err != nil {
		return nil, err
	}
	localAddress := peerAddress{
		network: conn.LocalAddr().Network(),
		peerID:  conn.LocalAddr().String(),
	}
	remoteAddress := peerAddress{
		network: conn.RemoteAddr().Network(),
		peerID:  remotePeerInfo.ID,
	}
	newConn := NewAddressableConn(conn, localAddress, remoteAddress)
	return newConn, nil
}
