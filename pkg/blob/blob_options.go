package blob

import (
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/objectmanager"
)

func WithObjectManager(x objectmanager.ObjectManager) func(*manager) {
	return func(r *manager) {
		r.objectmanager = x
	}
}

func WithResolver(res resolver.Resolver) func(*manager) {
	return func(r *manager) {
		r.resolver = res
	}
}

func WithChunkSize(bytes int) func(*manager) {
	return func(r *manager) {
		r.chunkSize = bytes
	}
}
