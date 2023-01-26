package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	IdentityAlias struct {
		_       string       `cborgen:"$type,const=core/identity.alias"`
		Network NetworkAlias `cborgen:"network,omitempty"`
		Handle  string       `cborgen:"handle,omitempty"`
	}
	Identity struct {
		_        string    `cborgen:"$type,const=core/identity"`
		Metadata Metadata  `cborgen:"$metadata,omitempty"`
		Keys     PublicKey `cborgen:"keys"`
		Next     PublicKey `cborgen:"next"`
	}
	IdentityIdentifier struct {
		IdentityAlias *IdentityAlias
		Identity      *Identity
	}
)

func (i *IdentityAlias) String() string {
	return string(ShorthandIdentityAlias) + i.Network.Hostname + "/" + i.Handle
}

func ParseIdentityAlias(alias string) (*IdentityAlias, error) {
	t := string(ShorthandIdentityAlias)
	if !strings.HasPrefix(alias, t) {
		return nil, fmt.Errorf("invalid resource id")
	}

	alias = strings.TrimPrefix(alias, t)
	hostname, handle, _ := strings.Cut(alias, "/")
	return &IdentityAlias{
		Network: NetworkAlias{
			Hostname: hostname,
		},
		Handle: handle,
	}, nil
}

func (i *IdentityAlias) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *IdentityAlias) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if idString, ok := value.(string); ok {
		id, err := ParseIdentityAlias(idString)
		if err != nil {
			return fmt.Errorf("unable to scan into IdentityAlias: %w", err)
		}
		i.Handle = id.Handle
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityAlias")
}

func (i *Identity) String() string {
	return string(ShorthandIdentity) + NewDocumentID(i).String()
}

func (i *Identity) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *Identity) Scan(value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (i IdentityIdentifier) String() string {
	if i.IdentityAlias != nil {
		return i.IdentityAlias.String()
	}
	if i.Identity != nil {
		return i.Identity.String()
	}
	return ""
}

func (i IdentityIdentifier) MarshalCBORBytes() ([]byte, error) {
	if i.IdentityAlias != nil {
		return i.IdentityAlias.MarshalCBORBytes()
	}
	if i.Identity != nil {
		return i.Identity.MarshalCBORBytes()
	}
	return nil, fmt.Errorf("unable to marshal identity identifier")
}

func (i IdentityIdentifier) UnmarshalCBORBytes(b []byte) error {
	t, err := GetDocumentTypeFromCbor(b)
	if err != nil {
		return fmt.Errorf("unable to find type for identity identifier: %w", err)
	}
	switch t {
	case "core/identity.alias":
		i.IdentityAlias = &IdentityAlias{}
		return i.IdentityAlias.UnmarshalCBORBytes(b)
	case "core/identity":
		i.Identity = &Identity{}
		return i.Identity.UnmarshalCBORBytes(b)
	default:
		return fmt.Errorf("unknown identity identifier type: %s", t)
	}
}
