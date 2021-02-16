package object

import (
	"nimona.io/pkg/context"
)

type (
	Getter interface {
		Get(
			context.Context,
			CID,
		) (*Object, error)
	}
	// GetterFunc is an adapter to allow the use of ordinary functions as
	// object.Getter
	GetterFunc func(
		context.Context,
		CID,
	) (*Object, error)
)
