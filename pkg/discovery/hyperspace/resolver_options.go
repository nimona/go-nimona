package hyperspace

import (
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/peer"
)

// WithKeychain overrides the default keychain for the Discoverer.
func WithKeychain(k keychain.Keychain) func(*Discoverer) {
	return func(w *Discoverer) {
		w.keychain = k
	}
}

// WithEventbus overrides the default eventbus for the Discoverer.
func WithEventbus(eb eventbus.Eventbus) func(*Discoverer) {
	return func(w *Discoverer) {
		w.eventbus = eb
	}
}

// WithExchange overrides the default exchange for the Discoverer.
func WithExchange(x exchange.Exchange) func(*Discoverer) {
	return func(w *Discoverer) {
		w.exchange = x
	}
}

// WithBoostrapPeers overrides the default bootstrap peers for the
// Discoverer.
func WithBoostrapPeers(ps []*peer.Peer) func(*Discoverer) {
	return func(w *Discoverer) {
		w.initialBootstrapPeers = ps
	}
}
