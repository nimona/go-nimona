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
