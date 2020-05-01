package exchange

import (
	"nimona.io/pkg/keychain"
)

// WithKeychain overrides the default keychain for the exchange.
func WithKeychain(k keychain.Keychain) func(*exchange) {
	return func(w *exchange) {
		w.keychain = k
	}
}
