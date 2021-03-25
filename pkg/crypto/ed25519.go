package crypto

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/teserakt-io/golang-ed25519/extra25519"
	"golang.org/x/crypto/curve25519"

	"nimona.io/pkg/errors"
)

// https://blog.filippo.io/using-ed25519-keys-for-encryption
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519
// http://moderncrypto.org/mail-archive/curves/2014/000205.html
// https://signal.org/docs/specifications/xeddsa
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519

// we are opting for ed to x at this point based on FiloSottile's age spec

type (
	KeyType    uint64
	PrivateKey string
	PublicKey  string
)

const (
	EmptyPrivateKey = PrivateKey("")
	EmptyPublicKey  = PublicKey("")

	cidEd25519Private = 0x1300
	cidEd25519Public  = 0xed

	UnknownKey  KeyType = 0
	PeerKey     KeyType = 0x6E00 // codec code for nimona _peer_ key
	IdentityKey KeyType = 0x6E01 // codec code for nimona _identity_ key
)

func ed25519PrivateToPrivateKey(
	k ed25519.PrivateKey,
	t KeyType,
) (PrivateKey, error) {
	h, err := multihash.Encode(k, uint64(t))
	if err != nil {
		panic(err)
	}
	c := cid.NewCidV1(cidEd25519Private, h)
	s, err := multibase.Encode(multibase.Base32, c.Bytes())
	if err != nil {
		panic(err)
	}
	return PrivateKey(s), nil
}

func ed25519PrivateFromPrivateKey(k PrivateKey) (ed25519.PrivateKey, error) {
	c, err := cid.Decode(string(k))
	if err != nil {
		return nil, err
	}
	if c.Type() != cidEd25519Private {
		return nil, errors.Error("invalid or unsupported private key type")
	}
	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return nil, err
	}
	return ed25519.PrivateKey(h.Digest), nil
}

func ed25519PublicFromPublicKey(k PublicKey) (ed25519.PublicKey, error) {
	c, err := cid.Decode(string(k))
	if err != nil {
		return nil, err
	}
	if c.Type() != cidEd25519Public {
		return nil, errors.Error("invalid or unsupported public key type")
	}
	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return nil, err
	}
	return ed25519.PublicKey(h.Digest), nil
}

func ed25519PublicToPublicKey(
	k ed25519.PublicKey,
	t KeyType,
) (PublicKey, error) {
	h, err := multihash.Encode(k, uint64(t))
	if err != nil {
		panic(err)
	}
	c := cid.NewCidV1(cidEd25519Public, h)
	s, err := multibase.Encode(multibase.Base32, c.Bytes())
	if err != nil {
		panic(err)
	}
	return PublicKey(s), nil
}

func GenerateEd25519PrivateKey(keyType KeyType) (PrivateKey, error) {
	_, b, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	return ed25519PrivateToPrivateKey(b, keyType)
}

func NewPrivateKey(seed []byte, keyType KeyType) PrivateKey {
	b := ed25519.NewKeyFromSeed(seed)
	k, err := ed25519PrivateToPrivateKey(b, keyType)
	if err != nil {
		panic(err)
	}
	return k
}

func NewPublicKey(publicKey ed25519.PublicKey, keyType KeyType) PublicKey {
	k, err := ed25519PublicToPublicKey(publicKey, keyType)
	if err != nil {
		panic(err)
	}
	return k
}

func (i PrivateKey) ed25519() ed25519.PrivateKey {
	k, _ := ed25519PrivateFromPrivateKey(i)
	return k
}

func (i PrivateKey) Type() KeyType {
	c, err := cid.Decode(string(i))
	if err != nil {
		return UnknownKey
	}
	if c.Type() != cidEd25519Private {
		return UnknownKey
	}
	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return UnknownKey
	}
	return KeyType(h.Code)
}

func (i PrivateKey) PublicKey() PublicKey {
	return NewPublicKey(
		i.ed25519().Public().(ed25519.PublicKey),
		i.Type(),
	)
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
func CalculateSharedKey(priv PrivateKey, pub PublicKey) ([]byte, error) {
	ca := privateEd25519KeyToCurve25519(priv.ed25519())
	cB := publicEd25519KeyToCurve25519(pub.ed25519())
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return nil, err
	}
	return ss, nil
}

// NewSharedKey calculates a shared secret given a private and a public key,
// and returns it
func NewSharedKey(priv PrivateKey, pub PublicKey) (*PrivateKey, []byte, error) {
	ca := privateEd25519KeyToCurve25519(priv.ed25519())
	cB := publicEd25519KeyToCurve25519(pub.ed25519())
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return nil, nil, err
	}
	return &priv, ss, nil
}

// NewEphemeralSharedKey creates a new ec25519 key pair, calculates a shared
// secret given a public key, and returns the created public key and secret
func NewEphemeralSharedKey(pub PublicKey) (*PrivateKey, []byte, error) {
	priv, err := GenerateEd25519PrivateKey(PeerKey)
	if err != nil {
		return nil, nil, err
	}
	return NewSharedKey(priv, pub)
}

func (i PrivateKey) IsEmpty() bool {
	return i == ""
}

func (i PrivateKey) Bytes() []byte {
	if i.IsEmpty() {
		return nil
	}
	return i.ed25519().Seed()
}

func (i PrivateKey) Sign(message []byte) []byte {
	return ed25519.Sign(i.ed25519(), message)
}

func (i PrivateKey) String() string {
	return string(i)
}

func (r PublicKey) ed25519() ed25519.PublicKey {
	k, _ := ed25519PublicFromPublicKey(r)
	return k
}

func (r PublicKey) Type() KeyType {
	c, err := cid.Decode(string(r))
	if err != nil {
		return UnknownKey
	}
	if c.Type() != cidEd25519Public {
		return UnknownKey
	}
	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return UnknownKey
	}
	return KeyType(h.Code)
}

func (r PublicKey) IsEmpty() bool {
	return r == ""
}

func (r PublicKey) Bytes() []byte {
	if r.IsEmpty() {
		return nil
	}
	return r.ed25519()
}

func (r PublicKey) String() string {
	return string(r)
}

func (r PublicKey) Address() string {
	return "peer:" + r.String()
}

func (r PublicKey) Verify(message []byte, signature []byte) error {
	ok := ed25519.Verify(r.ed25519(), message, signature)
	if !ok {
		return errors.Error("invalid signature")
	}
	return nil
}

// TODO invalid public keys should not be evaluated for equality
func (r PublicKey) Equals(w PublicKey) bool {
	if r == w {
		return true
	}
	ew := w.ed25519()
	if ew == nil {
		return false
	}
	rw := r.ed25519()
	if rw == nil {
		return false
	}
	return ew.Equal(rw)
}
