package network

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/object/value"
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

func FilterByObjectCID(objectCIDs ...value.CID) EnvelopeFilter {
	return func(e *Envelope) bool {
		for _, cid := range objectCIDs {
			if cid == e.Payload.CID() {
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
		rID, ok := rIDVal.(value.String)
		if !ok {
			return false
		}
		if rID == "" {
			return false
		}
		return string(rID) == requestID
	}
}
