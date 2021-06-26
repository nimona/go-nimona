package network

import (
	"time"

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
		connectionInfo         *peer.ConnectionInfo
		waitForResponse        interface{}
		waitForResponseTimeout time.Duration
	}
)

func SendWithConnectionInfo(c *peer.ConnectionInfo) func(*sendOptions) {
	return func(w *sendOptions) {
		w.connectionInfo = c
	}
}

func SendWithResponse(
	v interface{},
	t time.Duration,
) func(*sendOptions) {
	return func(w *sendOptions) {
		w.waitForResponse = v
		if t == 0 {
			t = time.Second
		}
		w.waitForResponseTimeout = t
	}
}
