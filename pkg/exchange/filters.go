package exchange

import (
	"github.com/gobwas/glob"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
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
			if pattern.Match(e.Payload.GetType()) {
				return true
			}
		}
		return false
	}
}
