package resolver

import (
	"nimona.io/pkg/peer"
)

// WithBoostrapPeers overrides the default bootstrap peers for the
// resolver.
func WithBoostrapPeers(ps ...*peer.ConnectionInfo) func(*resolver) {
	return func(w *resolver) {
		w.bootstrapPeers = ps
	}
}
