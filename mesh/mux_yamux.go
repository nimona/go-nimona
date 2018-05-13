package mesh

import (
	"net"

	"github.com/hashicorp/yamux"
)

type Yamux struct{}

func (mux *Yamux) Initiate(conn net.Conn) (net.Conn, error) {
	// fmt.Println("> YAMUX")

	session, err := yamux.Client(conn, nil)
	if err != nil {
		panic(err)
	}

	stream, err := session.Open()
	if err != nil {
		return nil, err
	}

	newConnFn := func() (net.Conn, error) {
		newStream, err := session.Open()
		if err != nil {
			return nil, err
		}
		newStreamConn := NewStreamConn(conn, newStream)
		return newStreamConn, nil
	}

	acceptConnFn := func() (net.Conn, error) {
		newStream, err := session.AcceptStream()
		if err != nil {
			return nil, err
		}
		newStreamConn := NewStreamConn(conn, newStream)
		return newStreamConn, nil
	}

	streamConn := NewStreamConn(conn, stream)
	reusableConn := NewReusableConn(streamConn, newConnFn, acceptConnFn)
	return reusableConn, nil
}

func (mux *Yamux) Handle(conn net.Conn) (net.Conn, error) {
	// fmt.Println("< YAMUX")

	session, err := yamux.Server(conn, nil)
	if err != nil {
		return nil, err
	}

	stream, err := session.Accept()
	if err != nil {
		return nil, err
	}

	newConnFn := func() (net.Conn, error) {
		newStream, err := session.Open()
		if err != nil {
			return nil, err
		}
		newStreamConn := NewStreamConn(conn, newStream)
		return newStreamConn, nil
	}

	acceptConnFn := func() (net.Conn, error) {
		newStream, err := session.AcceptStream()
		if err != nil {
			return nil, err
		}
		newStreamConn := NewStreamConn(conn, newStream)
		return newStreamConn, nil
	}

	streamConn := NewStreamConn(conn, stream)
	reusableConn := NewReusableConn(streamConn, newConnFn, acceptConnFn)
	return reusableConn, nil
}
