package sql

import (
	"github.com/gobwas/glob"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

func FilterByStreamHash(h object.Hash) SqlStoreFilter {
	return func(o object.Object) bool {
		os := o.Get("stream:s")
		switch oh := os.(type) {
		case object.Hash:
			return h.IsEqual(oh)
		case string:
			return h.String() == os
		default:
			return false
		}
	}
}

func FilterByObjectType(typePatterns ...string) SqlStoreFilter {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(errors.Wrap(err, errors.New("invalid pattern")))
		}
		patterns[i] = g
	}
	return func(o object.Object) bool {
		for _, pattern := range patterns {
			if pattern.Match(o.GetType()) {
				return true
			}
		}
		return false
	}
}
