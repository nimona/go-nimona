package net

import (
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/keychain"
)

// WithKeychain overrides the default keychain for the network.
func WithKeychain(k keychain.Keychain) func(*network) {
	return func(n *network) {
		n.keychain = k
	}
}

// WithEventBus overrides the default eventbus for the network.
func WithEventBus(k eventbus.Eventbus) func(*network) {
	return func(n *network) {
		n.eventbus = k
	}
}
