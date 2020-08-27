package objectmanager

import (
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/resolver"
)

func WithStore(st objectstore.Store) func(*manager) {
	return func(m *manager) {
		m.objectstore = st
	}
}

func WithExchange(x exchange.Exchange) func(*manager) {
	return func(m *manager) {
		m.exchange = x
	}
}

func WithKeychain(k keychain.Keychain) func(*manager) {
	return func(m *manager) {
		m.keychain = k
	}
}

func WithResolver(k resolver.Resolver) func(*manager) {
	return func(m *manager) {
		m.resolver = k
	}
}
