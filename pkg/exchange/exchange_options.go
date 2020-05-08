package exchange

import (
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/net"
)

// WithNet overrides the default network for the exchange.
func WithNet(n net.Network) func(*exchange) {
	return func(w *exchange) {
		w.net = n
	}
}

// WithKeychain overrides the default keychain for the exchange.
func WithKeychain(k keychain.Keychain) func(*exchange) {
	return func(w *exchange) {
		w.keychain = k
	}
}

// WithEventbus overrides the default eventbusfor the exchange.
func WithEventbus(eb eventbus.Eventbus) func(*exchange) {
	return func(w *exchange) {
		w.eventbus = eb
	}
}
