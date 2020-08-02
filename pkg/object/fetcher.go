package object

import (
	"nimona.io/pkg/context"
)

type (
	FetcherFunc func(
		context.Context,
		Hash,
	) (*Object, error)
	Fetcher interface {
		Fetch(
			context.Context,
			Hash,
		) (*Object, error)
	}
)
