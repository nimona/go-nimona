package dht

import (
	"encoding/json"
	"net"

	log "github.com/sirupsen/logrus"
)

type UDPNet struct {
}

func (n *UDPNet) StartServer(addr string, cb func(net.Conn)) error {
	srv, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.WithError(err).Error("Could not resolve server address")
		return err
	}
	l, err := net.ListenUDP("udp", srv)
	if err != nil {
		log.WithError(err).Error("Failed to started listening")
		return err
	}

	defer l.Close()

	for {
		cb(l)
	}
	return nil
}

func (n *UDPNet) SendMessage(msg Message, addr string) (int, error) {
	saddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.WithError(err).Error("Could not resolve server address")
		return 0, err
	}

	conn, err := net.DialUDP("udp", nil, saddr)
	if err != nil {
		log.WithError(err).Error("Could not dial server")
		return 0, err
	}
	defer conn.Close()

	msgm, err := json.Marshal(msg)
	if err != nil {
		log.WithError(err).Error("Could not marshall json")
		return 0, err
	}
	i, err := conn.Write([]byte(msgm))
	if err != nil {
		log.WithError(err).Error("Could not write to conn")
		return 0, err
	}
	return i, nil
}
