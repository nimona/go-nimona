package dht

import (
	"github.com/nimona/go-nimona/net"
)

func init() {
	net.RegisterContentType(PayloadTypePing, EnvelopePing{})
	net.RegisterContentType(PayloadTypePong, EnvelopePong{})

	net.RegisterContentType(PayloadTypeGetPeerInfo, EnvelopeGetPeerInfo{})
	net.RegisterContentType(PayloadTypePutPeerInfo, EnvelopePutPeerInfoFromEnvelope{})

	net.RegisterContentType(PayloadTypeGetProviders, EnvelopeGetProviders{})
	net.RegisterContentType(PayloadTypePutProviders, EnvelopePutProviders{})

	net.RegisterContentType(PayloadTypeGetValue, EnvelopeGetValue{})
	net.RegisterContentType(PayloadTypePutValue, EnvelopePutValue{})
}

// Envelope types
const (
	PayloadTypePing string = "dht.ping"
	PayloadTypePong        = "dht.pong"

	PayloadTypeGetPeerInfo = "dht.get-peer-info"
	PayloadTypePutPeerInfo = "dht.put-peer-info"

	PayloadTypeGetProviders = "dht.get-providers"
	PayloadTypePutProviders = "dht.put-providers"

	PayloadTypeGetValue = "dht.get-value"
	PayloadTypePutValue = "dht.put-value"
)

type EnvelopePing struct {
	SenderPeerInfo *net.Envelope `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type EnvelopePong struct {
	SenderPeerInfo *net.Envelope `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type EnvelopeGetPeerInfo struct {
	SenderPeerInfo *net.Envelope `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type EnvelopePutPeerInfoFromEnvelope struct {
	SenderPeerInfo *net.Envelope   `json:"sender_peer_info"`
	RequestID      string         `json:"request_id,omitempty"`
	PeerInfo       *net.Envelope   `json:"peer_info"`
	ClosestPeers   []*net.Envelope `json:"closest_peers"`
}

type EnvelopeGetProviders struct {
	SenderPeerInfo *net.Envelope `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
}

type EnvelopePutProviders struct {
	SenderPeerInfo *net.Envelope   `json:"sender_peer_info"`
	RequestID      string         `json:"request_id,omitempty"`
	Key            string         `json:"key"`
	PeerIDs        []string       `json:"peer_ids"`
	ClosestPeers   []*net.Envelope `json:"closest_peers"`
}

type EnvelopeGetValue struct {
	SenderPeerInfo *net.Envelope `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
}

type EnvelopePutValue struct {
	SenderPeerInfo *net.Envelope   `json:"sender_peer_info"`
	RequestID      string         `json:"request_id,omitempty"`
	Key            string         `json:"key"`
	Value          string         `json:"value"`
	ClosestPeers   []*net.Envelope `json:"closest_peers"`
}
