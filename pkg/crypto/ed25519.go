package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/multiformats/go-varint"
	"github.com/teserakt-io/golang-ed25519/extra25519"
	"golang.org/x/crypto/curve25519"
)

// https://blog.filippo.io/using-ed25519-keys-for-encryption
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519
// http://moderncrypto.org/mail-archive/curves/2014/000205.html
// https://signal.org/docs/specifications/xeddsa
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519

// we are opting for ed to x at this point based on FiloSottile's age spec

type (
	PublicKey struct {
		Type      KeyType
		Algorithm KeyAlgorithm
		RawKey    ed25519.PublicKey
	}
	PrivateKey struct {
		Type      KeyType
		Algorithm KeyAlgorithm
		RawKey    ed25519.PrivateKey
	}
)

func (k PublicKey) String() string {
	return encodeToCID(uint64(k.Algorithm), uint64(k.Type), k.RawKey)
}

func (k PublicKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *PublicKey) UnmarshalString(s string) error {
	c, err := cid.Decode(s)
	if err != nil {
		return err
	}

	if c.Type() != uint64(Ed25519Public) {
		return ErrUnsupportedKeyAlgorithm
	}

	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return err
	}

	k.Algorithm = Ed25519Public
	k.RawKey = ed25519.PublicKey(h.Digest)
	k.Type = KeyType(h.Code)

	return nil
}

func (k *PublicKey) UnmarshalJSON(s []byte) error {
	v := ""
	if err := json.Unmarshal(s, &v); err != nil {
		return err
	}
	return k.UnmarshalString(v)
}

func (k PrivateKey) String() string {
	return encodeToCID(uint64(k.Algorithm), uint64(k.Type), k.RawKey)
}

func (k PrivateKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *PrivateKey) UnmarshalString(s string) error {
	c, err := cid.Decode(s)
	if err != nil {
		return err
	}

	if c.Type() != uint64(Ed25519Private) {
		return ErrUnsupportedKeyAlgorithm
	}

	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return err
	}

	k.Algorithm = Ed25519Private
	k.RawKey = ed25519.PrivateKey(h.Digest)
	k.Type = KeyType(h.Code)

	return nil
}

func (k *PrivateKey) UnmarshalJSON(s []byte) error {
	v := ""
	if err := json.Unmarshal(s, &v); err != nil {
		return err
	}
	return k.UnmarshalString(v)
}

func (k PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{
		Algorithm: Ed25519Public,
		Type:      k.Type,
		RawKey:    k.RawKey.Public().(ed25519.PublicKey),
	}
}

func NewEd25519PrivateKey(keyType KeyType) (*PrivateKey, error) {
	_, k, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &PrivateKey{
		Algorithm: Ed25519Private,
		Type:      keyType,
		RawKey:    k,
	}, nil
}

func NewEd25519PrivateKeyFromSeed(
	seed []byte,
	keyType KeyType,
) *PrivateKey {
	b := ed25519.NewKeyFromSeed(seed)
	return &PrivateKey{
		Algorithm: Ed25519Private,
		Type:      keyType,
		RawKey:    b,
	}
}

func NewEd25519PublicKeyFromRaw(
	raw ed25519.PublicKey,
	keyType KeyType,
) *PublicKey {
	return &PublicKey{
		Algorithm: Ed25519Public,
		Type:      keyType,
		RawKey:    raw,
	}
}

func publicEd25519KeyToCurve25519(pub ed25519.PublicKey) []byte {
	var edPk [ed25519.PublicKeySize]byte
	var curveKey [32]byte
	copy(edPk[:], pub)
	if !extra25519.PublicKeyToCurve25519(&curveKey, &edPk) {
		panic("could not convert ed25519 public key to curve25519")
	}
	return curveKey[:]
}

func privateEd25519KeyToCurve25519(priv ed25519.PrivateKey) []byte {
	var edSk [ed25519.PrivateKeySize]byte
	var curveKey [32]byte
	copy(edSk[:], priv)
	extra25519.PrivateKeyToCurve25519(&curveKey, &edSk)
	return curveKey[:]
}

// CalculateSharedKey calculates a shared secret given a private an public key
func CalculateSharedKey(
	priv *PrivateKey,
	pub *PublicKey,
) ([]byte, error) {
	if priv.Algorithm != Ed25519Private || pub.Algorithm != Ed25519Public {
		return nil, ErrUnsupportedKeyAlgorithm
	}
	ca := privateEd25519KeyToCurve25519(priv.RawKey)
	cB := publicEd25519KeyToCurve25519(pub.RawKey)
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return nil, fmt.Errorf("error getting x25519, %w", err)
	}
	return ss, nil
}

// NewSharedKey calculates a shared secret given a private and a public key,
// and returns it
func NewSharedKey(
	priv *PrivateKey,
	pub *PublicKey,
) (*PrivateKey, []byte, error) {
	ca := privateEd25519KeyToCurve25519(priv.RawKey)
	cB := publicEd25519KeyToCurve25519(pub.RawKey)
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return nil, nil, err
	}
	return priv, ss, nil
}

// CalculateEphemeralSharedKey creates a new ec25519 key pair, calculates a
// shared secret given a public key, and returns the created public key and
// secret
func CalculateEphemeralSharedKey(
	pub *PublicKey,
) (*PrivateKey, []byte, error) {
	priv, err := NewEd25519PrivateKey(PeerKey)
	if err != nil {
		return nil, nil, err
	}
	return NewSharedKey(priv, pub)
}

func (k PrivateKey) Sign(message []byte) []byte {
	return ed25519.Sign(k.RawKey, message)
}

func (k PublicKey) Verify(message []byte, signature []byte) error {
	ok := ed25519.Verify(k.RawKey, message, signature)
	if !ok {
		return ErrInvalidSignature
	}
	return nil
}

func (k PublicKey) Equals(w *PublicKey) bool {
	return k.Algorithm == w.Algorithm && k.Type == w.Type && k.RawKey.Equal(w.RawKey)
}

func encodeToCID(cidCode, multihashCode uint64, raw []byte) string {
	mh := make(
		[]byte,
		varint.UvarintSize(multihashCode)+
			varint.UvarintSize(uint64(len(raw)))+
			len(raw),
	)
	n := varint.PutUvarint(mh, multihashCode)
	n += varint.PutUvarint(mh[n:], uint64(len(raw)))
	copy(mh[n:], raw)
	c := cid.NewCidV1(cidCode, mh)
	// nolint: errcheck // cannot error
	s, _ := multibase.Encode(multibase.Base32, c.Bytes())
	return s
}
