package resolver

import (
	"nimona.io/pkg/peer"
)

// WithBoostrapPeers overrides the default bootstrap peers for the
// resolver.
func WithBoostrapPeers(ps []*peer.Peer) func(*resolver) {
	return func(w *resolver) {
		w.initialBootstrapPeers = ps
	}
}
