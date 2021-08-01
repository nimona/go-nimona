// KeyStream is a simplified implementation of KERI.
// It is not a full nor faithful implementation of the spec and is not intended
// to be for the foreseeable future.
// It is just an attempt to implement some of the basic aspects of KERI using
// nimona's streams and use the stream's root hash as an identifier.

package keystream

import (
	"fmt"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

const (
	ErrUnsupportedVersion = errors.Error("unsupported version")
	ErrInvalidVersion     = errors.Error("invalid version")
)

// - KERI is the identifier of KERI events
// - 1 is the hex major version code
// - 0 the hex minor version code
// - CBOR, is the code for the serialized encoding format of the event
// - 0001c2 is the hex size of the serialized event
// const Version = "KERI10CBOR0001c2"
const (
	Version       = "~KERI00"
	InceptionType = "keri.Inception/v0"
	RotationType  = "keri.Rotation/v0"
)

// events
type (
	Inception struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=keri.Inception/v0"`
		Version  string          `nimona:"v:s"`
		// Prefix   string          `nimona:"i:s"`
		// Sequence int `nimona:"s:i"`
		// EventType        string          `nimona:"t:s"`
		// EventDigest      string          `nimona:"d:s"`
		// PriorEventDigest string          `nimona:"p:s"`
		// SigThreshold      *SigThreshold  `nimona:"kt"` // [][]*big.Rat
		Key           crypto.PublicKey `nimona:"k:s"`
		NextKeyDigest chore.Hash       `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		// Seals         []*Seal   `nimona:"a:am"`
		DelegatorSeal *DelegatorSeal `nimona:"da:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
	Rotation struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=keri.Rotation/v0"`
		Version  string          `nimona:"v:s"`
		// Prefix   string          `nimona:"i:s"`
		// Sequence int `nimona:"s:i"`
		// EventType        string          `nimona:"t:s"`
		// EventDigest      string          `nimona:"d:s"`
		// PriorEventDigest string          `nimona:"p:s"`
		// SigThreshold      *SigThreshold  `nimona:"kt"` // [][]*big.Rat
		Key           crypto.PublicKey `nimona:"k:s"`
		NextKeyDigest chore.Hash       `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		// Seals         []*Seal   `nimona:"a:am"`
		DelegatorSeal *DelegatorSeal `nimona:"da:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
	InceptionDelegation struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=keri.Rotation/v0"`
		Version  string          `nimona:"v:s"`
		// Prefix   string          `nimona:"i:s"`
		// Sequence int `nimona:"s:i"`
		// EventType        string          `nimona:"t:s"`
		// EventDigest      string          `nimona:"d:s"`
		// PriorEventDigest string          `nimona:"p:s"`
		// SigThreshold      *SigThreshold  `nimona:"kt"` // [][]*big.Rat
		// Key           crypto.PublicKey `nimona:"k:s"`
		// NextKeyDigest chore.Hash       `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		Seals []*Seal `nimona:"a:am"`
		// DelegatorSeal *DelegatorSeal `nimona:"da:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
	RotationDelegation struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=keri.Rotation/v0"`
		Version  string          `nimona:"v:s"`
		// Prefix   string          `nimona:"i:s"`
		// Sequence int `nimona:"s:i"`
		// EventType        string          `nimona:"t:s"`
		// EventDigest      string          `nimona:"d:s"`
		// PriorEventDigest string          `nimona:"p:s"`
		// SigThreshold      *SigThreshold  `nimona:"kt"` // [][]*big.Rat
		// Key           crypto.PublicKey `nimona:"k:s"`
		// NextKeyDigest chore.Hash       `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		Seals []*Seal `nimona:"a:am"`
		// DelegatorSeal *DelegatorSeal `nimona:"da:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
)

// components
type (
	Trait string
	Seal  struct { // Type      SealType `nimona:"-"`
		Root        chore.Hash      `nimona:"rd:s"`
		Permissions object.Policies `nimona:"p:am"`
		// Prefix    string `nimona:"i:s"`
		// Sequence  string `nimona:"s:s"`
		// EventType string `nimona:"t:s"`
		// Digest    string `nimona:"d:s"`
	}
	DelegatorSeal struct {
		Root chore.Hash `nimona:"rd:s"`
		// Type      SealType `nimona:"-"`
		// Delegation  chore.Hash      `nimona:"d:s"`
		// Permissions object.Policies `nimona:"p:am"`
		// Prefix    string `nimona:"i:s"`
		Sequence uint64 `nimona:"s:u"`
		// EventType string `nimona:"t:s"`
		// Digest    string `nimona:"d:s"`
	}
	Config struct {
		Trait Trait `nimona:"trait:s"`
	}
)

const (
	TraitEstOnly       Trait = "EO"  //  Only allow establishment events
	TraitDoNotDelegate Trait = "DND" //  Dot not allow delegated identifiers
	TraitNoBackers     Trait = "NB"  // Do not allow any backers for registry
)

func (inc *Inception) apply(s *KeyStream) error {
	if inc.Version != Version {
		return ErrUnsupportedVersion
	}

	o, err := object.Marshal(inc)
	if err != nil {
		return fmt.Errorf("error trying to get inc hash, %w", err)
	}

	s.Root = o.Hash()
	s.Delegator = inc.DelegatorSeal.Root
	s.Version = inc.Version
	s.ActiveKey = inc.Key
	s.RotatedKeys = []crypto.PublicKey{}

	return nil
}

func (rot *Rotation) apply(s *KeyStream) error {
	if rot.Version != s.Version {
		return ErrInvalidVersion
	}

	// TODO if hash(rot.Key) != s.NextKeyHash { err }

	s.RotatedKeys = append(s.RotatedKeys, s.ActiveKey)
	s.ActiveKey = rot.Key
	return nil
}

// state and key manager
type (
	applier interface {
		apply(s *KeyStream) error
	}
	// KeyStream of a single KERI stream
	KeyStream struct {
		Version     string
		Root        chore.Hash
		Delegator   chore.Hash
		ActiveKey   crypto.PublicKey
		NextKeyHash chore.Hash
		RotatedKeys []crypto.PublicKey
	}
)

func (s *KeyStream) GetIdentity() chore.Hash {
	return s.Root
}

func FromStream(
	or object.ReadCloser,
) (*KeyStream, error) {
	s := &KeyStream{}

	for {
		o, err := or.Read()
		if err == object.ErrReaderDone {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading objects, %w", err)
		}

		var v applier
		switch o.Type {
		case InceptionType:
			v = &Inception{}
		case RotationType:
			v = &Rotation{}
		}

		err = object.Unmarshal(o, v)
		if err != nil {
			return nil, fmt.Errorf("error unmarshling object, %w", err)
		}

		err = v.apply(s)
		if err != nil {
			return nil, fmt.Errorf("error applying event, %w", err)
		}
	}

	return s, nil
}
