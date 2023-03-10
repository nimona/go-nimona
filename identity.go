package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	IdentityAlias struct {
		_        string `nimona:"$type,type=core/identity.alias"`
		Hostname string `nimona:"hostname,omitempty"`
		Path     string `nimona:"path,omitempty"`
	}
	IdentityInfo struct {
		_             Metadata      `nimona:"$metadata,omitempty,type=core/identity/info"`
		Alias         IdentityAlias `nimona:"alias,omitempty"`
		Identity      Identity      `nimona:"identity,omitempty"`
		PeerAddresses []PeerAddr    `nimona:"peerAddresses"`
	}
	KeyGraph struct {
		Metadata Metadata  `nimona:"$metadata,omitempty,type=core/identity"`
		Keys     PublicKey `nimona:"keys"`
		Next     PublicKey `nimona:"next"`
	}
	Identity struct {
		_          string     `nimona:"$type,type=core/identity/id"`
		Use        string     `nimona:"type,omitempty"` // provider, user, etc
		KeyGraphID DocumentID `nimona:"keyGraphID"`
	}
)

func (i *IdentityAlias) String() string {
	r := string(ShorthandIdentityAlias) + i.Hostname
	if i.Path != "" {
		r += "/" + i.Path
	}
	return r
}

func ParseIdentityAlias(alias string) (*IdentityAlias, error) {
	t := string(ShorthandIdentityAlias)
	if !strings.HasPrefix(alias, t) {
		return nil, fmt.Errorf("invalid resource id")
	}

	alias = strings.TrimPrefix(alias, t)
	hostname, path, _ := strings.Cut(alias, "/")
	return &IdentityAlias{
		Hostname: hostname,
		Path:     path,
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
		i.Hostname = id.Hostname
		i.Path = id.Path
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityAlias")
}

func (i *KeyGraph) Identity() *Identity {
	if i == nil {
		return nil
	}
	return &Identity{
		KeyGraphID: NewDocumentID(i.Document()),
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
