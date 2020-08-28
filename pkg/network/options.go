package network

import (
	"nimona.io/pkg/keychain"
)

// WithKeychain overrides the default keychain for the network.
func WithKeychain(k keychain.Keychain) func(*network) {
	return func(w *network) {
		w.keychain = k
	}
}
