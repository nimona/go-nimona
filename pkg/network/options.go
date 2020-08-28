package network

import (
	"nimona.io/pkg/localpeer"
)

// WithLocalPeer overrides the default localpeer for the network.
func WithLocalPeer(k localpeer.LocalPeer) func(*network) {
	return func(w *network) {
		w.localpeer = k
	}
}
