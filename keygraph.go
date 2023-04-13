package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"nimona.io/tilde"
)

type (
	KeyGraphID DocumentHash
	KeyGraph   struct {
		Metadata Metadata  `nimona:"$metadata,omitempty,type=core/identity"`
		Keys     PublicKey `nimona:"keys"`
		Next     PublicKey `nimona:"next"`
	}
)

func NewKeyGraph(current, next PublicKey) *KeyGraph {
	// TODO(@geoah) add metadata.permissions, etc
	return &KeyGraph{
		Keys: current,
		Next: next,
	}
}

func ParseKeyGraphNRI(nri string) (KeyGraphID, error) {
	t := string(ShorthandIdentity)
	if !strings.HasPrefix(nri, t) {
		return KeyGraphID{}, fmt.Errorf("invalid keygraph nri")
	}

	nri = strings.TrimPrefix(nri, t)
	dh, err := ParseDocumentHash(nri)
	if err != nil {
		return KeyGraphID{}, fmt.Errorf("unable to parse keygraph nri: %w", err)
	}
	return KeyGraphID(dh), nil
}

func ParseKeyGraphID(s string) (KeyGraphID, error) {
	dh, err := ParseDocumentHash(s)
	if err != nil {
		return KeyGraphID{}, fmt.Errorf("unable to parse keygraph id: %w", err)
	}
	return KeyGraphID(dh), nil
}

func (k *KeyGraph) ID() KeyGraphID {
	return KeyGraphID(NewDocumentHash(k.Document()))
}

func (k KeyGraphID) DocumentHash() DocumentHash {
	return DocumentHash(k)
}

func (k KeyGraphID) NRI() string {
	return ShorthandIdentity.String() + k.String()
}

func (k KeyGraphID) DocumentID() DocumentID {
	return DocumentID{
		DocumentHash: k.DocumentHash(),
	}
}

func (k KeyGraphID) String() string {
	return DocumentHash(k).String()
}

func (k KeyGraphID) IsEmpty() bool {
	return DocumentHash(k).IsEmpty()
}

func (k KeyGraphID) TildeValue() tilde.Value {
	return DocumentHash(k).TildeValue()
}

func (k KeyGraphID) Value() (driver.Value, error) {
	return k.String(), nil
}

func (k *KeyGraphID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if s, ok := value.(string); ok {
		id, err := ParseDocumentHash(s)
		if err != nil {
			return fmt.Errorf("unable to scan into IdentityAlias: %w", err)
		}
		*k = KeyGraphID(id)
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityAlias")
}
