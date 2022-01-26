package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
	"github.com/teserakt-io/golang-ed25519/extra25519"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/curve25519"

	"nimona.io/pkg/did"
	"nimona.io/pkg/multiheader"
	"nimona.io/pkg/tilde"
)

// https://blog.filippo.io/using-ed25519-keys-for-encryption
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519
// http://moderncrypto.org/mail-archive/curves/2014/000205.html
// https://signal.org/docs/specifications/xeddsa
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519

// we are opting for ed to x at this point based on FiloSottile's age spec

type (
	KeyAlgorithm multicodec.Code
)

const (
	Ed25519Private KeyAlgorithm = KeyAlgorithm(multicodec.Ed25519Priv)
	Ed25519Public  KeyAlgorithm = KeyAlgorithm(multicodec.Ed25519Pub)
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
	b := multiheader.Encode(multicodec.Code(k.Algorithm), k.RawKey)
	// nolint: errcheck // cannot error
	s, _ := multibase.Encode(multibase.Base58BTC, b)
	return s
}

func (k PublicKey) DID() did.DID {
	return did.DID{
		Method:       did.MethodNimona,
		IdentityType: did.IdentityTypePeer,
		Identity:     k.String(),
	}
}

func (k PublicKey) IsEmpty() bool {
	return k.RawKey == nil
}

func (k PublicKey) Hash() tilde.Digest {
	return tilde.String(k.String()).Hash()
}

func (k PublicKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalText implements encoding.TextUnmarshaler mainly for use
// by envconfig.
func (k *PublicKey) UnmarshalText(b []byte) error {
	return k.UnmarshalString(string(b))
}

func (k *PublicKey) UnmarshalString(s string) error {
	_, b, err := multibase.Decode(s)
	if err != nil {
		return fmt.Errorf("unable to decode multibase, %w", err)
	}

	c, r, err := multiheader.Decode(b)
	if err != nil {
		return fmt.Errorf("unable to decode multiheader, %w", err)
	}

	if c != multicodec.Ed25519Pub {
		return ErrUnsupportedKeyAlgorithm
	}

	k.Algorithm = Ed25519Public
	k.RawKey = ed25519.PublicKey(r)

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
	b := multiheader.Encode(multicodec.Code(k.Algorithm), k.RawKey)
	// nolint: errcheck // cannot error
	s, _ := multibase.Encode(multibase.Base58BTC, b)
	return s
}

func (k PrivateKey) Seed() []byte {
	return k.RawKey.Seed()
}

func (k PrivateKey) BIP39() string {
	m, _ := bip39.NewMnemonic(k.Seed())
	return m
}

func (k PrivateKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(k.String())
}

// UnmarshalText implements encoding.TextUnmarshaler mainly for use
// by envconfig.
func (k *PrivateKey) UnmarshalText(b []byte) error {
	return k.UnmarshalString(string(b))
}

func (k *PrivateKey) UnmarshalString(s string) error {
	_, b, err := multibase.Decode(s)
	if err != nil {
		return fmt.Errorf("unable to decode multibase, %w", err)
	}

	c, r, err := multiheader.Decode(b)
	if err != nil {
		return fmt.Errorf("unable to decode multiheader, %w", err)
	}

	if c != multicodec.Ed25519Priv {
		return ErrUnsupportedKeyAlgorithm
	}

	k.Algorithm = Ed25519Private
	k.RawKey = ed25519.PrivateKey(r)

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

func PublicKeyFromDID(d did.DID) (*PublicKey, error) {
	pk := &PublicKey{}
	err := pk.UnmarshalString(d.Identity)
	if err != nil {
		return nil, err
	}
	return pk, nil
}
