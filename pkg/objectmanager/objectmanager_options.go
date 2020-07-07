package objectmanager

import (
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/sqlobjectstore"
)

func WithStore(st *sqlobjectstore.Store) func(*manager) {
	return func(m *manager) {
		m.store = st
	}
}

func WithExchange(x exchange.Exchange) func(*manager) {
	return func(m *manager) {
		m.exchange = x
	}
}
