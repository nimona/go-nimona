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
	IdentityID struct {
		_              string     `cborgen:"$type,const=core/identity/id"`
		IdentityRootID DocumentID `cborgen:"identityStreamID"`
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
	h, err := NewDocumentHash(i)
	if err != nil {
		panic(fmt.Errorf("unable to get hash of identity: %w", err))
	}
	return string(ShorthandIdentity) + h.String()
}

func (i *Identity) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *Identity) Scan(value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (i *Identity) IdentityID() IdentityID {
	return IdentityID{
		IdentityRootID: NewDocumentID(i),
	}
}

func (i *IdentityID) String() string {
	return string(ShorthandIdentity) + i.IdentityRootID.DocumentHash.String()
}

func (i *IdentityID) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *IdentityID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if idString, ok := value.(string); ok {
		id, err := ParseIdentityID(idString)
		if err != nil {
			return fmt.Errorf("unable to scan into IdentityID: %w", err)
		}
		i.IdentityRootID = id.IdentityRootID
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityID")
}

func ParseIdentityID(id string) (*IdentityID, error) {
	t := string(ShorthandIdentity)
	if !strings.HasPrefix(id, t) {
		return nil, fmt.Errorf("invalid resource id")
	}

	id = strings.TrimPrefix(id, t)
	dh, err := ParseDocumentHash(id)
	if err != nil {
		return nil, fmt.Errorf("unable to parse identity id: %w", err)
	}
	return &IdentityID{
		IdentityRootID: DocumentID{
			DocumentHash: dh,
		},
	}, nil
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
