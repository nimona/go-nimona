package objectmanager

import (
	"nimona.io/pkg/network"
	"nimona.io/pkg/localpeer"
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

func WithLocalPeer(k localpeer.LocalPeer) func(*manager) {
	return func(m *manager) {
		m.localpeer = k
	}
}

func WithResolver(k resolver.Resolver) func(*manager) {
	return func(m *manager) {
		m.resolver = k
	}
}
