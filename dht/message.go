package dht

import (
	net "github.com/nimona/go-nimona-net"
)

// Message types
const (
	MessageTypePing string = "PING"
	MessageTypePut         = "PUT"
	MessageTypeGet         = "GET"
)

// Key prefixes
const (
	KeyPrefixPeer         string = "nimona/peer/"
	KeyPrefixProvider            = "nimona/provider/"
	KeyPrefixKeyValuePair        = "nimona/kv/"
)

type messageGet struct {
	OriginPeer net.Peer `json:"p"`
	QueryID    string   `json:"q"`
	Key        string   `json:"k"`
}

type messagePut struct {
	OriginPeer net.Peer `json:"p"`
	QueryID    string   `json:"q"`
	Key        string   `json:"k"`
	Values     []string `json:"v"`
}
