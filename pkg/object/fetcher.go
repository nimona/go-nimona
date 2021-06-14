package object

import (
	"nimona.io/pkg/chore"
	"nimona.io/pkg/context"
)

type (
	Getter interface {
		Get(
			context.Context,
			chore.CID,
		) (*Object, error)
	}
	// GetterFunc is an adapter to allow the use of ordinary functions as
	// object.Getter
	GetterFunc func(
		context.Context,
		chore.CID,
	) (*Object, error)
)
