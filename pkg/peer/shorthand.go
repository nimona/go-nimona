package peer

import (
	"strings"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

const (
	ErrInvalidShorthand = errors.Error("invalid shorthand")
)

// Shorthand has the form of `<public-key>@<address>`.
// They are mostly used for configuration and bootstrapping.
type Shorthand string

func (s Shorthand) IsValid() bool {
	// TODO validate key and address
	return len(strings.Split(string(s), "@")) == 2
}

func (s Shorthand) PublicKey() crypto.PublicKey {
	// TODO validate key
	ps := strings.Split(string(s), "@")
	if len(ps) != 2 {
		return ""
	}
	return crypto.PublicKey(ps[0])
}

func (s Shorthand) Addresses() []string {
	// TODO validate address
	ps := strings.Split(string(s), "@")
	if len(ps) != 2 {
		return nil
	}
	return []string{
		ps[1],
	}
}

func (s Shorthand) Peer() (*ConnectionInfo, error) {
	ps := strings.Split(string(s), "@")
	if len(ps) != 2 {
		return nil, ErrInvalidShorthand
	}
	return &ConnectionInfo{
		PublicKey: crypto.PublicKey(ps[0]),
		Addresses: []string{
			ps[1],
		},
	}, nil
}
