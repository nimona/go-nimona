package object

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/did"
	"nimona.io/pkg/tilde"
)

type ObjectFilter func(*Object) bool

func FilterByType(typePatterns ...string) ObjectFilter {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(fmt.Errorf("invalid pattern: %w", err))
		}
		patterns[i] = g
	}
	return func(o *Object) bool {
		for _, pattern := range patterns {
			if pattern.Match(o.Type) {
				return true
			}
		}
		return false
	}
}

func FilterByObjectHash(objectHashes ...tilde.Digest) ObjectFilter {
	return func(o *Object) bool {
		for _, h := range objectHashes {
			if o.Hash().Equal(h) {
				return true
			}
		}
		return false
	}
}

func FilterByRootHash(rootHashes ...tilde.Digest) ObjectFilter {
	return func(o *Object) bool {
		for _, h := range rootHashes {
			r := o.Metadata.Root
			if !r.IsEmpty() && r.Equal(h) {
				return true
			}
		}
		return false
	}
}

func FilterByOwner(ownerDID ...did.DID) ObjectFilter {
	return func(o *Object) bool {
		for _, h := range ownerDID {
			o := o.Metadata.Owner
			if !o.IsEmpty() && o.Equals(h) {
				return true
			}
		}
		return false
	}
}

func FilterByRequestID(requestID string) ObjectFilter {
	return func(o *Object) bool {
		rIDVal, ok := o.Data["requestID"]
		if !ok {
			return false
		}
		rID, ok := rIDVal.(tilde.String)
		if !ok {
			return false
		}
		if rID == "" {
			return false
		}
		return string(rID) == requestID
	}
}
