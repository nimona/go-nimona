package resolver

import (
	"sync"

	"nimona.io/pkg/context"
	"nimona.io/pkg/did"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

const ErrNotFound errors.Error = "not found"

//go:generate mockgen -destination=../resolvermock/resolvermock_generated.go -package=resolvermock -source=resolver.go

type Resolver interface {
	LookupByDID(
		ctx context.Context,
		id did.DID,
	) ([]*peer.ConnectionInfo, error)
	LookupByContent(
		ctx context.Context,
		cid tilde.Digest,
	) ([]*peer.ConnectionInfo, error)
}

type CompositeResolver struct {
	Resolvers []Resolver
	mutex     sync.RWMutex
}

// New creates a new composite resolver.
func New(resolvers ...Resolver) *CompositeResolver {
	return &CompositeResolver{
		Resolvers: resolvers,
	}
}

func (r *CompositeResolver) LookupByDID(
	ctx context.Context,
	id did.DID,
) ([]*peer.ConnectionInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, r := range r.Resolvers {
		if c, err := r.LookupByDID(ctx, id); err == nil {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (r *CompositeResolver) LookupByContent(
	ctx context.Context,
	cid tilde.Digest,
) ([]*peer.ConnectionInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	for _, r := range r.Resolvers {
		if c, err := r.LookupByContent(ctx, cid); err == nil {
			return c, nil
		}
	}
	return nil, ErrNotFound
}

func (r *CompositeResolver) RegisterResolver(resolver Resolver) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.Resolvers = append(r.Resolvers, resolver)
}
