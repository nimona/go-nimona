package peer

import (
	"strings"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

const (
	ErrInvalidNimonaDID = errors.Error("invalid nimona ID")
)

var EmptyID = ID{}

type (
	Method       string
	IdentityType string
)

const (
	MethodNimona          Method       = "nimona"
	IdentityTypePeer      IdentityType = "peer"
	IdentityTypeKeyStream IdentityType = "keystream"
)

// ID defines a peer's identifier.
// Note(geoah): This was at one point a DID but idn't really make much sense.
// DID terminology is being used still but should be phased out.
type ID struct {
	Method       Method
	IdentityType IdentityType
	Identity     string
}

func (d ID) Equals(d2 ID) bool {
	return d == d2
}

func (d ID) IsEmpty() bool {
	return d == EmptyID
}

// MarshalString returns the string representation of the ID.
// Never returns an error.
func (d ID) MarshalString() (string, error) {
	return d.String(), nil
}

// String returns the string representation of the ID.
func (d ID) String() string {
	if d == EmptyID {
		return ""
	}
	return strings.Join([]string{
		string(d.Method),
		string(d.IdentityType),
		d.Identity,
	}, ":")
}

func (d *ID) UnmarshalString(s string) error {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return ErrInvalidNimonaDID
	}
	if Method(parts[0]) != MethodNimona {
		return ErrInvalidNimonaDID
	}
	switch IdentityType(parts[1]) {
	case IdentityTypePeer:
		d.IdentityType = IdentityTypePeer
	case IdentityTypeKeyStream:
		d.IdentityType = IdentityTypeKeyStream
	default:
		return ErrInvalidNimonaDID
	}
	d.Method = MethodNimona
	d.Identity = parts[2]
	return nil
}

func NewID(s string) (*ID, error) {
	if s == "" {
		return &EmptyID, nil
	}
	id := &ID{}
	if err := id.UnmarshalString(s); err != nil {
		return nil, err
	}
	return id, nil
}

func MustNewID(s string) *ID {
	id, err := NewID(s)
	if err != nil {
		panic(err)
	}
	return id
}

func IDFromPublicKey(k crypto.PublicKey) ID {
	return ID{
		Method:       MethodNimona,
		IdentityType: IdentityTypePeer,
		Identity:     k.String(),
	}
}

func NewIDFromKey(d ID) (*crypto.PublicKey, error) {
	pk := &crypto.PublicKey{}
	err := pk.UnmarshalString(d.Identity)
	if err != nil {
		return nil, err
	}
	return pk, nil
}
