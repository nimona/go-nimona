package dht

import (
	"net"
)

type Net interface {
	StartServer(address string, callback func(net.Conn)) error
	SendMessage(msg Message, address string) (int, error)
}

type CallType int

const (
	PING CallType = iota
	STORE
	FIND_NODE
	FIND_VALUE
)

type Message struct {
	Type        CallType `json:"t"`
	Nonce       string   `json:"n"`
	OriginPeer  Peer     `json:"op"`
	QueryPeerID ID       `json:"qp"`
	Peers       []Peer   `json:"rp"`
}
