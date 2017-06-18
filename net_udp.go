package dht

import (
	"encoding/json"
	"net"
)

type UDPNet struct {
}

func (n *UDPNet) StartServer(cb func(net.Conn)) error {
	l, err := net.Listen("tcp", ":8889")
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go cb(c)
	}
}

func (n *UDPNet) SendMessage(msg Message, addr string) (int, error) {
	srv, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return 0, err
	}

	conn, err := net.DialUDP("udp", nil, srv)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	msgm, err := json.Marshal(msg)
	if err != nil {
		return 0, err
	}
	conn.Write([]byte(msgm))
	return len(msgm), nil
}
