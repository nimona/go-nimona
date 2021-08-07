package object

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/tilde"
)

type (
	Getter interface {
		Get(
			context.Context,
			tilde.Digest,
		) (*Object, error)
	}
	// GetterFunc is an adapter to allow the use of ordinary functions as
	// object.Getter
	GetterFunc func(
		context.Context,
		tilde.Digest,
	) (*Object, error)
)
