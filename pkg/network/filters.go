package network

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/tilde"
)

func FilterByObjectType(typePatterns ...string) EnvelopeFilter {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(fmt.Errorf("invalid pattern: %w", err))
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

func FilterByObjectHash(objectHashes ...tilde.Hash) EnvelopeFilter {
	return func(e *Envelope) bool {
		for _, h := range objectHashes {
			if e.Payload.Hash().Equal(h) {
				return true
			}
		}
		return false
	}
}

func FilterByRequestID(requestID string) EnvelopeFilter {
	return func(e *Envelope) bool {
		rIDVal, ok := e.Payload.Data["requestID"]
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
