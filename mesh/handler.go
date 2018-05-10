package mesh

import "net"

type Handler interface {
	Initiate(net.Conn) (net.Conn, error)
	Handle(net.Conn) (net.Conn, error)
}
