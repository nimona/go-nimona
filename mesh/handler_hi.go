package mesh

import (
	"net"
)

type Hi struct {
}

func (hi *Hi) Initiate(conn net.Conn) (net.Conn, error) {
	// fmt.Println("> HI")
	return conn, nil
}

func (hi *Hi) Handle(conn net.Conn) (net.Conn, error) {
	// fmt.Println("< HI")
	return conn, nil
}
