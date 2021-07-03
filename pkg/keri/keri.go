// This is not a legit or even correct implementation of KERI.
// It is just an attempt to implement some of the basic aspects of KERI using
// nimona's streams and use the stream's root hash as an identifier.

package keri

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

// KERI events from kerigo
// https://github.com/decentralized-identity/kerigo/blob/master/pkg/event/event.go
//
// type Event struct {
// 	Version           string         `json:"v"`
// 	Prefix            string         `json:"i,omitempty"`
// 	Sequence          string         `json:"s,omitempty"`
// 	EventType         string         `json:"t"`
// 	EventDigest       string         `json:"d,omitempty"`
// 	PriorEventDigest  string         `json:"p,omitempty"`
// 	SigThreshold      *SigThreshold  `json:"kt,omitempty"`
// 	Keys              []string       `json:"k,omitempty"`
// 	Next              string         `json:"n,omitempty"`
// 	WitnessThreshold  string         `json:"wt,omitempty"`
// 	Witnesses         []string       `json:"w,omitempty"`
// 	AddWitness        []string       `json:"wa,omitempty"`
// 	RemoveWitness     []string       `json:"wr,omitempty"`
// 	Config            []prefix.Trait `json:"c,omitempty" cbor:",omitempty"`
// 	Seals             SealArray      `json:"a,omitempty"`
// 	DelegatorSeal     *Seal          `json:"da,omitempty"`
// 	LastEvent         *Seal          `json:"e,omitempty"`
// 	LastEstablishment *Seal          `json:"ee,omitempty"`
// 	_dig              string
// }

// sub-structures
type (
	Identity chore.Hash
	Digest   struct{}
	Seal     struct { // Type      SealType `nimona:"-"`
		// Root      string `nimona:"rd:s"`
		// Prefix    string `nimona:"i:s"`
		// Sequence  string `nimona:"s:s"`
		// EventType string `nimona:"t:s"`
		// Digest    string `nimona:"d:s"`
	}
	Config struct {
		Trait string `nimona:"trait:s"`
	}
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
		Config []*Config `nimona:"c:am"`
		// Seals         []*Seal   `nimona:"a:am"`
		// DelegatorSeal *Seal     `nimona:"da:m"`
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
		Config []*Config `nimona:"c:am"`
		// Seals         []*Seal   `nimona:"a:am"`
		// DelegatorSeal *Seal     `nimona:"da:m"`
		// LastEvent         *Seal     `nimona:"e:m"`
		// LastEstablishment *Seal     `nimona:"ee:m"`
	}
)

func (inc *Inception) apply(s *State) error {
	if inc.Version != Version {
		return ErrUnsupportedVersion
	}

	o, err := object.Marshal(inc)
	if err != nil {
		return fmt.Errorf("error trying to get inc hash, %w", err)
	}

	s.RootHash = o.Hash()
	s.Version = inc.Version
	s.ActiveKey = inc.Key
	s.RotatedKeys = []crypto.PublicKey{}

	return nil
}

func (rot *Rotation) apply(s *State) error {
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
		apply(s *State) error
	}
	// State of a single KERI stream
	State struct {
		Version      string
		RootHash     chore.Hash
		ActiveKey    crypto.PublicKey
		NextKeyHash  chore.Hash
		RotatedKeys  []crypto.PublicKey
		LatestObject *object.Object
	}
)

func (s *State) GetIdentity() Identity {
	return Identity(s.RootHash)
}

func CreateState(
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

		s.LatestObject = o
	}

	return s, nil
}
