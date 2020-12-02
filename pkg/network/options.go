package network

import (
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/peer"
)

// WithLocalPeer overrides the default localpeer for the network.
func WithLocalPeer(k localpeer.LocalPeer) func(*network) {
	return func(w *network) {
		w.localpeer = k
	}
}

type (
	sendOptions struct {
		connectionInfo *peer.ConnectionInfo
	}
)

func SendWithConnectionInfo(c *peer.ConnectionInfo) func(*sendOptions) {
	return func(w *sendOptions) {
		w.connectionInfo = c
	}
}
