package object

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/object/value"
)

type (
	Getter interface {
		Get(
			context.Context,
			value.CID,
		) (*Object, error)
	}
	// GetterFunc is an adapter to allow the use of ordinary functions as
	// object.Getter
	GetterFunc func(
		context.Context,
		value.CID,
	) (*Object, error)
)
