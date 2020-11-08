package blob

import (
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/sqlobjectstore"
)

func WithStore(st *sqlobjectstore.Store) func(*requester) {
	return func(r *requester) {
		r.store = st
	}
}

func WithObjectManager(x objectmanager.ObjectManager) func(*requester) {
	return func(r *requester) {
		r.objmgr = x
	}
}

func WithResolver(res resolver.Resolver) func(*requester) {
	return func(r *requester) {
		r.resolver = res
	}
}
