package resolver

import (
	"nimona.io/pkg/peer"
)

// WithBoostrapPeers overrides the default bootstrap peers for the
// resolver.
func WithBoostrapPeers(ps ...*peer.ConnectionInfo) func(*Resolver) {
	return func(w *Resolver) {
		w.bootstrapPeers = ps
	}
}
