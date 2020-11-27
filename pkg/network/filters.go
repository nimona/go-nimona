package network

import (
	"github.com/gobwas/glob"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

func FilterBySender(keys ...crypto.PublicKey) EnvelopeFilter {
	return func(e *Envelope) bool {
		for _, key := range keys {
			if e.Sender.Equals(key) {
				return true
			}
		}
		return false
	}
}

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

func FilterByNonce(nonce string) EnvelopeFilter {
	return func(e *Envelope) bool {
		return e.Payload.Data["nonce:s"].(string) == nonce
	}
}

func FilterByRequestID(requestID string) EnvelopeFilter {
	return func(e *Envelope) bool {
		return e.Payload.Data["requestID:s"].(string) == requestID
	}
}
