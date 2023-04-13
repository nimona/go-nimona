package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"nimona.io/tilde"
)

type (
	KeygraphID DocumentHash
	Keygraph   struct {
		Metadata Metadata  `nimona:"$metadata,omitempty,type=core/identity"`
		Keys     PublicKey `nimona:"keys"`
		Next     PublicKey `nimona:"next"`
	}
)

func NewKeygraph(current, next PublicKey) *Keygraph {
	// TODO(@geoah) add metadata.permissions, etc
	return &Keygraph{
		Keys: current,
		Next: next,
	}
}

func ParseKeygraphNRI(nri string) (KeygraphID, error) {
	t := string(ShorthandIdentity)
	if !strings.HasPrefix(nri, t) {
		return KeygraphID{}, fmt.Errorf("invalid keygraph nri")
	}

	nri = strings.TrimPrefix(nri, t)
	dh, err := ParseDocumentHash(nri)
	if err != nil {
		return KeygraphID{}, fmt.Errorf("unable to parse keygraph nri: %w", err)
	}
	return KeygraphID(dh), nil
}

func ParseKeygraphID(s string) (KeygraphID, error) {
	dh, err := ParseDocumentHash(s)
	if err != nil {
		return KeygraphID{}, fmt.Errorf("unable to parse keygraph id: %w", err)
	}
	return KeygraphID(dh), nil
}

func (k *Keygraph) ID() KeygraphID {
	return KeygraphID(NewDocumentHash(k.Document()))
}

func (k KeygraphID) DocumentHash() DocumentHash {
	return DocumentHash(k)
}

func (k KeygraphID) NRI() string {
	return ShorthandIdentity.String() + k.String()
}

func (k KeygraphID) DocumentID() DocumentID {
	return DocumentID{
		DocumentHash: k.DocumentHash(),
	}
}

func (k KeygraphID) String() string {
	return DocumentHash(k).String()
}

func (k KeygraphID) IsEmpty() bool {
	return DocumentHash(k).IsEmpty()
}

func (k KeygraphID) TildeValue() tilde.Value {
	return DocumentHash(k).TildeValue()
}

func (k KeygraphID) Value() (driver.Value, error) {
	return k.String(), nil
}

func (k *KeygraphID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if s, ok := value.(string); ok {
		id, err := ParseDocumentHash(s)
		if err != nil {
			return fmt.Errorf("unable to scan into IdentityAlias: %w", err)
		}
		*k = KeygraphID(id)
		return nil
	}
	return fmt.Errorf("unable to scan into IdentityAlias")
}
