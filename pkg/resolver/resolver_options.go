package resolver

import (
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/peer"
)

// WithKeychain overrides the default keychain for the resolver.
func WithKeychain(k keychain.Keychain) func(*resolver) {
	return func(w *resolver) {
		w.keychain = k
	}
}

// WithEventbus overrides the default eventbus for the resolver.
func WithEventbus(eb eventbus.Eventbus) func(*resolver) {
	return func(w *resolver) {
		w.eventbus = eb
	}
}

// WithExchange overrides the default exchange for the resolver.
func WithExchange(x exchange.Exchange) func(*resolver) {
	return func(w *resolver) {
		w.exchange = x
	}
}

// WithBoostrapPeers overrides the default bootstrap peers for the
// resolver.
func WithBoostrapPeers(ps []*peer.Peer) func(*resolver) {
	return func(w *resolver) {
		w.initialBootstrapPeers = ps
	}
}
