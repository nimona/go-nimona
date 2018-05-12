package mesh

import (
	"context"
	"errors"
	"io"
	"net"
)

type Relay struct {
	net *Net
}

func (mux *Relay) Initiate(conn net.Conn) (net.Conn, error) {
	remotePeerID := conn.RemoteAddr().String()
	if err := WriteToken(conn, []byte(remotePeerID)); err != nil {
		return nil, err
	}
	resp, err := ReadToken(conn)
	if err != nil {
		return nil, err
	}
	if string(resp) != "ok" {
		return nil, errors.New("could not establish relay")
	}
	return conn, nil
}

func (mux *Relay) Handle(conn net.Conn) (net.Conn, error) {
	remotePeerID, err := ReadToken(conn)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	relayConn, err := mux.net.Dial(ctx, string(remotePeerID))
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
	finalConn := NewAddressableConn(relayConn, localAddress, remoteAddress)
	if err := WriteToken(conn, []byte("ok")); err != nil {
		return nil, err
	}
	if err := pipe(conn, relayConn); err != nil {
		return nil, err
	}
	return finalConn, nil
}

func pipe(a, b io.ReadWriteCloser) error {
	done := make(chan error, 1)
	cp := func(r, w io.ReadWriteCloser) {
		_, err := io.Copy(r, w)
		done <- err
	}
	go cp(a, b)
	go cp(b, a)
	return <-done
}
