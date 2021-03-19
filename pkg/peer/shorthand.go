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

func (s Shorthand) ConnectionInfo() (*ConnectionInfo, error) {
	ps := strings.Split(string(s), "@")
	if len(ps) != 2 {
		return nil, ErrInvalidShorthand
	}
	pk := &crypto.PublicKey{}
	if err := pk.UnmarshalString(ps[0]); err != nil {
		return nil, err
	}
	return &ConnectionInfo{
		PublicKey: pk,
		Addresses: []string{
			ps[1],
		},
	}, nil
}
