package did

import (
	"strings"

	"nimona.io/pkg/errors"
)

const (
	ErrInvalidNimonaDID = errors.Error("invalid nimona DID")
)

var Empty = DID{}

type (
	Method       string
	IdentityType string
)

const (
	MethodNimona          Method       = "nimona"
	IdentityTypePeer      IdentityType = "peer"
	IdentityTypeKeyStream IdentityType = "keystream"
)

const (
	didPrefix = "did"
)

// DID is a distributed identity structure.
// It does not currently support the full DID spec but should eventually
// be able to be fully compliant.
// TODO: make compatible with the full DID spec
type DID struct {
	Method       Method
	IdentityType IdentityType
	Identity     string
}

func (d DID) Equals(d2 DID) bool {
	return d == d2
}

func (d DID) IsEmpty() bool {
	return d == Empty
}

// MarshalString returns the string representation of the DID.
// Never returns an error.
func (d DID) MarshalString() (string, error) {
	return d.String(), nil
}

// String returns the string representation of the DID.
func (d DID) String() string {
	if d == Empty {
		return ""
	}
	return strings.Join([]string{
		didPrefix,
		string(d.Method),
		string(d.IdentityType),
		d.Identity,
	}, ":")
}

func (d *DID) UnmarshalString(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 4 {
		return ErrInvalidNimonaDID
	}
	if parts[0] != didPrefix {
		return ErrInvalidNimonaDID
	}
	if Method(parts[1]) != MethodNimona {
		return ErrInvalidNimonaDID
	}
	switch IdentityType(parts[2]) {
	case IdentityTypePeer:
		d.IdentityType = IdentityTypePeer
	case IdentityTypeKeyStream:
		d.IdentityType = IdentityTypeKeyStream
	default:
		return ErrInvalidNimonaDID
	}
	d.Method = MethodNimona
	d.Identity = parts[3]
	return nil
}

func Parse(s string) (*DID, error) {
	if s == "" {
		return &Empty, nil
	}
	did := &DID{}
	if err := did.UnmarshalString(s); err != nil {
		return nil, err
	}
	return did, nil
}

func MustParse(s string) *DID {
	did, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return did
}
