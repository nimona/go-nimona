package did

import (
	"strings"

	"nimona.io/pkg/errors"
)

const (
	ErrInvalidDID = errors.Error("invalid DID")
)

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

func (d *DID) Equals(d2 *DID) bool {
	if d == d2 && d != nil {
		return true
	}
	if d == nil || d2 == nil {
		return false
	}
	return d.Method == d2.Method && d.Identity == d2.Identity
}

func (id DID) MarshalString() (string, error) {
	return didPrefix + string(id.Method) + ":" + id.Identity, nil
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
