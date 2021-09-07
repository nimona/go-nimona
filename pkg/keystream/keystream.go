// State is a simplified implementation of KERI.
// It is not a full nor faithful implementation of the spec and is not intended
// to be for the foreseeable future.
// It is just an attempt to implement some of the basic aspects of KERI using
// nimona's streams and use the stream's root hash as an identifier.

package keystream

import (
	"fmt"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/tilde"
)

const (
	ErrUnsupportedVersion = errors.Error("unsupported version")
	ErrInvalidVersion     = errors.Error("invalid version")
)

// - ~ (tilde) is used to denote that this is not a real KERI implementation
//   but rather a version re-designed to work with nimona's tilde objects.
// - KERI is the identifier of KERI events
// - 0 is the major version code
// - 0 the minor version code
// - Serialization encoding and size are no longer used
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
		NextKeyDigest tilde.Digest     `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		// Seals         DelegateSeal   `nimona:"dr:m"`
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
		NextKeyDigest tilde.Digest     `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		// DelegatorSeal *DelegatorSeal `nimona:"da:m"`
		DelegateSeal DelegateSeal `nimona:"dr:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
	// TODO(geoah): implement RotationInteraction
	RotationInteraction struct{}
	// nolint: lll
	DelegationInteraction struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=keri.DelegationInteraction/v0"`
		Version  string          `nimona:"v:s"`
		// Prefix   string          `nimona:"i:s"`
		// Sequence int `nimona:"s:i"`
		// EventType        string          `nimona:"t:s"`
		// EventDigest      string          `nimona:"d:s"`
		// PriorEventDigest string          `nimona:"p:s"`
		// SigThreshold      *SigThreshold  `nimona:"kt"` // [][]*big.Rat
		// Key           crypto.PublicKey `nimona:"k:s"`
		// NextKeyDigest tilde.Digest       `nimona:"n:s"`
		// WitnessThreshold  string    `nimona:"wt:s"`
		// Witnesses         []string  `nimona:"w:as"`
		// AddWitness        []string  `nimona:"wa:as"`
		// RemoveWitness     []string  `nimona:"wr:as"`
		// Config []*Config `nimona:"c:am"`
		DelegateSeal DelegateSeal `nimona:"dr:m"`
		// DelegatorSeal *DelegatorSeal `nimona:"da:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
)

// components
type (
	Trait        string
	DelegateSeal struct {
		Root        tilde.Digest `nimona:"rd:r"`
		Permissions Permissions  `nimona:"p:m"`
		// Type      SealType `nimona:"-"`
		// Prefix    string `nimona:"i:s"`
		// Sequence  string `nimona:"s:s"`
		// EventType string `nimona:"t:s"`
		// Digest    string `nimona:"d:s"`
	}
	DelegatorSeal struct {
		Metadata object.Metadata `nimona:"@metadata:m,type=keri.DelegatorSeal/v0"`
		Root     tilde.Digest    `nimona:"rd:r"`
		// Type      SealType `nimona:"-"`
		// Delegation  tilde.Digest      `nimona:"d:s"`
		Permissions Permissions `nimona:"p:m"`
		// Prefix    string `nimona:"i:s"`
		Sequence uint64 `nimona:"s:u"`
		// EventType string `nimona:"t:s"`
		// Digest    string `nimona:"d:s"`
	}
	// TODO: implement permissions
	Permissions struct{}
	Config      struct {
		Trait Trait `nimona:"trait:s"`
	}
)

const (
	TraitEstOnly       Trait = "EO"  //  Only allow establishment events
	TraitDoNotDelegate Trait = "DND" //  Dot not allow delegated identifiers
	TraitNoBackers     Trait = "NB"  // Do not allow any backers for registry
)

func (inc *Inception) apply(s *State) error {
	if inc.Version != Version {
		return ErrUnsupportedVersion
	}

	if s.Sequence != 0 {
		return fmt.Errorf("invalid keystream sequence")
	}

	if inc.Metadata.Sequence != 0 {
		return fmt.Errorf("invalid event sequence")
	}

	if inc.Key.IsEmpty() {
		return fmt.Errorf("key cannot be empty")
	}

	if inc.NextKeyDigest.IsEmpty() {
		return fmt.Errorf("next key digest cannot be empty")
	}

	o, err := object.Marshal(inc)
	if err != nil {
		return fmt.Errorf("error trying to get inc hash, %w", err)
	}

	s.Root = o.Hash()
	if inc.DelegatorSeal != nil {
		s.DelegatorRoot = inc.DelegatorSeal.Root
		s.Delegator = did.DID{
			Method:   did.MethodNimona,
			Identity: string(s.DelegatorRoot),
		}
	}
	s.Version = inc.Version
	s.ActiveKey = inc.Key
	s.NextKeyDigest = inc.NextKeyDigest
	s.RotatedKeys = []crypto.PublicKey{}

	return nil
}

func (rot *Rotation) apply(s *State) error {
	if rot.Version != s.Version {
		return ErrInvalidVersion
	}

	if rot.Metadata.Sequence != s.Sequence+1 {
		return fmt.Errorf("invalid event sequence")
	}

	if rot.Key.IsEmpty() {
		return fmt.Errorf("key cannot be empty")
	}

	if rot.Key.Hash() != s.NextKeyDigest {
		return fmt.Errorf("current key digest doesn't match previous next key")
	}

	if rot.NextKeyDigest.IsEmpty() {
		return fmt.Errorf("next key digest cannot be empty")
	}

	s.RotatedKeys = append(s.RotatedKeys, s.ActiveKey)
	s.ActiveKey = rot.Key
	s.NextKeyDigest = rot.NextKeyDigest
	s.Sequence = rot.Metadata.Sequence
	return nil
}

func (del *DelegationInteraction) apply(s *State) error {
	if del.Version != Version {
		return ErrUnsupportedVersion
	}

	if del.Metadata.Sequence != s.Sequence+1 {
		return fmt.Errorf("invalid event sequence")
	}

	s.DelegateRoots = append(s.DelegateRoots, del.DelegateSeal.Root)
	s.Delegates = append(s.Delegates, did.DID{
		Method:   did.MethodNimona,
		Identity: string(del.DelegateSeal.Root),
	})

	s.Sequence = del.Metadata.Sequence
	return nil
}

// state and key manager
type (
	applier interface {
		apply(s *State) error
	}
	// State of a single KERI stream
	State struct {
		Version       string
		Root          tilde.Digest
		ActiveKey     crypto.PublicKey
		NextKeyDigest tilde.Digest
		RotatedKeys   []crypto.PublicKey
		Sequence      uint64
		// Delegator
		DelegatorRoot tilde.Digest
		Delegator     did.DID
		// Delegates
		DelegateRoots []tilde.Digest
		Delegates     []did.DID
	}
)

func (s State) GetDID() did.DID {
	return did.DID{
		Method:   did.MethodNimona,
		Identity: string(s.Root),
	}
}

func (s *State) GetIdentity() tilde.Digest {
	return s.Root
}

func FromStream(
	or object.ReadCloser,
) (*State, error) {
	s := &State{}

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

		// TODO: before or after applying each event we should be verifying
		// that any seals are actually valid by fetching the stream.
	}

	return s, nil
}
