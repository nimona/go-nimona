package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	IdentityAlias struct {
		Metadata Metadata     `nimona:"$metadata,omitempty,type=core/identity.alias"`
		Network  NetworkAlias `nimona:"network,omitempty"`
		Handle   string       `nimona:"handle,omitempty"`
	}
	KeyGraph struct {
		Metadata Metadata  `nimona:"$metadata,omitempty,type=core/identity"`
		Keys     PublicKey `nimona:"keys"`
		Next     PublicKey `nimona:"next"`
	}
	Identity struct {
		Metadata   Metadata   `nimona:"$metadata,omitempty,type=core/identity/id"`
		KeyGraphID DocumentID `nimona:"keyGraphID"`
	}
	IdentityIdentifier struct {
		IdentityAlias *IdentityAlias
		Identity      *Identity
	}
)

func (i *IdentityAlias) String() string {
	return string(ShorthandIdentityAlias) + i.Network.Hostname + "/" + i.Handle
}

func (i *IdentityAlias) IdentityIdentifier() IdentityIdentifier {
	return IdentityIdentifier{
		IdentityAlias: i,
	}
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

func (i *KeyGraph) Identity() *Identity {
	if i == nil {
		return nil
	}
	return &Identity{
		KeyGraphID: NewDocumentID(i.DocumentMap()),
	}
}

func (i *Identity) String() string {
	return string(ShorthandIdentity) + i.KeyGraphID.DocumentHash.String()
}

func (i *Identity) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i *Identity) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if idString, ok := value.(string); ok {
		id, err := ParseIdentity(idString)
		if err != nil {
			return fmt.Errorf("unable to scan into IdentityID: %w", err)
		}
		i.KeyGraphID = id.KeyGraphID
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityID")
}

func ParseIdentity(id string) (*Identity, error) {
	t := string(ShorthandIdentity)
	if !strings.HasPrefix(id, t) {
		return nil, fmt.Errorf("invalid resource id")
	}

	id = strings.TrimPrefix(id, t)
	dh, err := ParseDocumentHash(id)
	if err != nil {
		return nil, fmt.Errorf("unable to parse identity id: %w", err)
	}
	return &Identity{
		KeyGraphID: DocumentID{
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

// func (i IdentityIdentifier) MarshalCBOR(w io.Writer) error {
// 	if i.IdentityAlias != nil {
// 		return i.IdentityAlias.MarshalCBOR(w)
// 	}
// 	if i.Identity != nil {
// 		return i.Identity.MarshalCBOR(w)
// 	}
// 	return fmt.Errorf("unable to marshal identity identifier")
// }

// func (i IdentityIdentifier) UnmarshalCBOR(r io.Reader) (err error) {
// 	doc := &DocumentBase{}
// 	err = doc.UnmarshalCBOR(r)
// 	if err != nil {
// 		return fmt.Errorf("unable to unmarshal network identifier into doc: %w", err)
// 	}

// 	switch doc.Type {
// 	case "core/identity.alias":
// 		i.IdentityAlias = &IdentityAlias{}
// 		return UnmarshalJSON(doc.DocumentBytes, i.IdentityAlias)
// 	case "core/identity":
// 		i.Identity = &Identity{}
// 		return UnmarshalJSON(doc.DocumentBytes, i.Identity)
// 	default:
// 		return fmt.Errorf("unknown identity identifier type: %s", doc.Type)
// 	}
// }

func (i *IdentityIdentifier) DocumentMap() *DocumentMap {
	if i.IdentityAlias != nil {
		return i.IdentityAlias.DocumentMap()
	}
	if i.Identity != nil {
		return i.Identity.DocumentMap()
	}
	return nil
}

func (i *IdentityIdentifier) FromDocumentMap(m *DocumentMap) error {
	switch m.Type() {
	case "core/identity.alias":
		i.IdentityAlias = &IdentityAlias{}
		err := i.IdentityAlias.FromDocumentMap(m)
		if err != nil {
			return fmt.Errorf("unable to unmarshal identity alias: %w", err)
		}
		return nil
	case "core/identity":
		i.Identity = &Identity{}
		err := i.Identity.FromDocumentMap(m)
		if err != nil {
			return fmt.Errorf("unable to unmarshal identity: %w", err)
		}
		return nil
	}
	return fmt.Errorf("unknown identity identifier type: %s", m.Type())
}
