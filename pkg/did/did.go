package did

import (
	"strings"

	"nimona.io/pkg/errors"
)

const (
	ErrInvalidDID = errors.Error("invalid DID")
)

var Empty = DID{}

type Method string

const (
	MethodKey    Method = "key"
	MethodNimona Method = "nimona"
)

const (
	didPrefix = "did:"
)

// DID is a distributed identity structure.
// It does not currently support the full DID spec but should eventually
// be able to be fully compliant.
// TODO: make compatible with the full DID spec
type DID struct {
	Method   Method
	Identity string
}

func (d *DID) Equals(d2 DID) bool {
	return *d == d2
}

func (d *DID) IsEmpty() bool {
	return *d == Empty
}

// MarshalString returns the string representation of the DID.
// Never returns an error.
func (id DID) MarshalString() (string, error) {
	return id.String(), nil
}

// String returns the string representation of the DID.
func (id DID) String() string {
	if id == Empty {
		return ""
	}
	return didPrefix + string(id.Method) + ":" + id.Identity
}

func (id *DID) UnmarshalString(s string) error {
	if !strings.HasPrefix(s, didPrefix) {
		return ErrInvalidDID
	}
	parts := strings.SplitN(s[len(didPrefix):], ":", 2)
	if len(parts) != 2 {
		return ErrInvalidDID
	}
	id.Method = Method(parts[0])
	id.Identity = parts[1]
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
