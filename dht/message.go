package dht

import (
	peer "github.com/nimona/go-nimona-fabric/peer"
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
	OriginPeer peer.Peer `json:"p"`
	QueryID    string    `json:"q"`
	Key        string    `json:"k"`
}

type messagePut struct {
	OriginPeer peer.Peer `json:"p"`
	QueryID    string    `json:"q"`
	Key        string    `json:"k"`
	Values     []string  `json:"v"`
}
