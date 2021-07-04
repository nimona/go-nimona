package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/teserakt-io/golang-ed25519/extra25519"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/curve25519"
)

// https://blog.filippo.io/using-ed25519-keys-for-encryption
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519
// http://moderncrypto.org/mail-archive/curves/2014/000205.html
// https://signal.org/docs/specifications/xeddsa
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519

// we are opting for ed to x at this point based on FiloSottile's age spec

type (
	KeyAlgorithm uint64
)

const (
	Ed25519Private KeyAlgorithm = 0x1300 // well known value
	Ed25519Public  KeyAlgorithm = 0xED   // well known value
)

type (
	PublicKey struct {
		Algorithm KeyAlgorithm
		RawKey    ed25519.PublicKey // TODO use crypto.PublicKey
	}
	PrivateKey struct {
		Algorithm KeyAlgorithm
		RawKey    ed25519.PrivateKey // TODO use crypto.PrivateKey
	}
)

var (
	EmptyPublicKey  = PublicKey{}
	EmptyPrivateKey = PrivateKey{}
)

func (k PublicKey) String() string {
	if k.IsEmpty() {
		return ""
	}
	return encodeToCID(uint64(k.Algorithm), k.RawKey)
}

func (k PublicKey) IsEmpty() bool {
	return k.RawKey == nil
}

func (k PublicKey) MarshalText() ([]byte, error) {
	s, err := k.MarshalString()
	return []byte(s), err
}

func (k PublicKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *PublicKey) UnmarshalText(s []byte) error {
	return k.UnmarshalString(string(s))
}

func (k *PublicKey) UnmarshalString(s string) error {
	c, err := cid.Decode(s)
	if err != nil {
		return fmt.Errorf("decoding cid, %w", err)
	}

	if c.Type() != uint64(Ed25519Public) {
		return ErrUnsupportedKeyAlgorithm
	}

	k.Algorithm = Ed25519Public
	k.RawKey = ed25519.PublicKey(c.Hash())

	return nil
}

func (k *PublicKey) UnmarshalJSON(s []byte) error {
	v := ""
	if err := json.Unmarshal(s, &v); err != nil {
		return err
	}
	return k.UnmarshalString(v)
}

func (k PrivateKey) IsEmpty() bool {
	return k.RawKey == nil
}

func (k PrivateKey) String() string {
	return encodeToCID(uint64(k.Algorithm), k.RawKey)
}

func (k PrivateKey) Seed() []byte {
	return k.RawKey.Seed()
}

func (k PrivateKey) BIP39() string {
	m, _ := bip39.NewMnemonic(k.Seed())
	return m
}

func (k PrivateKey) MarshalText() ([]byte, error) {
	s, err := k.MarshalString()
	return []byte(s), err
}

func (k PrivateKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

func (k *PrivateKey) UnmarshalText(s []byte) error {
	return k.UnmarshalString(string(s))
}

func (k *PrivateKey) UnmarshalString(s string) error {
	c, err := cid.Decode(s)
	if err != nil {
		return err
	}

	if c.Type() != uint64(Ed25519Private) {
		return ErrUnsupportedKeyAlgorithm
	}

	k.Algorithm = Ed25519Private
	k.RawKey = ed25519.PrivateKey(c.Hash())

	return nil
}

func (k *PrivateKey) UnmarshalJSON(s []byte) error {
	v := ""
	if err := json.Unmarshal(s, &v); err != nil {
		return err
	}
	return k.UnmarshalString(v)
}

func (k PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		Algorithm: Ed25519Public,
		RawKey:    k.RawKey.Public().(ed25519.PublicKey),
	}
}

func NewEd25519PrivateKey() (PrivateKey, error) {
	_, k, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return EmptyPrivateKey, err
	}
	return PrivateKey{
		Algorithm: Ed25519Private,
		RawKey:    k,
	}, nil
}

// TODO check validity and return error
func NewEd25519PrivateKeyFromSeed(seed []byte) PrivateKey {
	b := ed25519.NewKeyFromSeed(seed)
	return PrivateKey{
		Algorithm: Ed25519Private,
		RawKey:    b,
	}
}

var (
	charRegex  = regexp.MustCompile(`[^a-zA-Z ]+`)
	spaceRegex = regexp.MustCompile(`\s+`)
)

// TODO check validity and return error
func NewEd25519PrivateKeyFromBIP39(mnemonic string) (PrivateKey, error) {
	mnemonicClean := mnemonic
	mnemonicClean = charRegex.ReplaceAllString(mnemonicClean, " ")
	mnemonicClean = spaceRegex.ReplaceAllString(mnemonicClean, " ")
	mnemonicClean = strings.TrimSpace(mnemonicClean)
	seed, err := bip39.EntropyFromMnemonic(mnemonicClean)
	if err != nil {
		return EmptyPrivateKey, fmt.Errorf("error parsing mnemonic, %w", err)
	}
	return NewEd25519PrivateKeyFromSeed(seed), nil
}

// TODO check validity and return error
func NewEd25519PublicKeyFromRaw(raw ed25519.PublicKey) PublicKey {
	return PublicKey{
		Algorithm: Ed25519Public,
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
	priv PrivateKey,
	pub PublicKey,
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
	priv PrivateKey,
	pub PublicKey,
) (PrivateKey, []byte, error) {
	ca := privateEd25519KeyToCurve25519(priv.RawKey)
	cB := publicEd25519KeyToCurve25519(pub.RawKey)
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return EmptyPrivateKey, nil, err
	}
	return priv, ss, nil
}

// CalculateEphemeralSharedKey creates a new ec25519 key pair, calculates a
// shared secret given a public key, and returns the created public key and
// secret
func CalculateEphemeralSharedKey(
	pub PublicKey,
) (PrivateKey, []byte, error) {
	priv, err := NewEd25519PrivateKey()
	if err != nil {
		return EmptyPrivateKey, nil, err
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

func (k PublicKey) Equals(w PublicKey) bool {
	return k.Algorithm == w.Algorithm &&
		k.RawKey.Equal(w.RawKey)
}

func encodeToCID(cidCode uint64, raw []byte) string {
	c := cid.NewCidV1(cidCode, raw)
	// nolint: errcheck // cannot error
	s, _ := c.StringOfBase(multibase.Base58BTC)
	return s
}
