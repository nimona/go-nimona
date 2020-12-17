package network

import (
	"github.com/gobwas/glob"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

func FilterByObjectType(typePatterns ...string) EnvelopeFilter {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(errors.Wrap(err, errors.New("invalid pattern")))
		}
		patterns[i] = g
	}
	return func(e *Envelope) bool {
		for _, pattern := range patterns {
			if pattern.Match(e.Payload.Type) {
				return true
			}
		}
		return false
	}
}

func FilterByObjectHash(objectHashes ...object.Hash) EnvelopeFilter {
	return func(e *Envelope) bool {
		for _, hash := range objectHashes {
			if hash == e.Payload.Hash() {
				return true
			}
		}
		return false
	}
}

func FilterByRequestID(requestID string) EnvelopeFilter {
	return func(e *Envelope) bool {
		rIDVal, ok := e.Payload.Data["requestID:s"]
		if !ok {
			return false
		}
		return rIDVal.(string) == requestID
	}
}
