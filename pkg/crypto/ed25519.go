package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"

	"github.com/teserakt-io/golang-ed25519/extra25519"
	"golang.org/x/crypto/curve25519"

	"nimona.io/internal/encoding/base58"
	"nimona.io/pkg/errors"
)

// https://blog.filippo.io/using-ed25519-keys-for-encryption
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519
// http://moderncrypto.org/mail-archive/curves/2014/000205.html
// https://signal.org/docs/specifications/xeddsa
// https://libsodium.gitbook.io/doc/advanced/ed25519-curve25519

// we are opting for ed to x at this point based on FiloSottile's age spec

type (
	PrivateKey string
	PublicKey  string
)

const (
	EmptyPrivateKey = PrivateKey("")
	EmptyPublicKey  = PublicKey("")
)

func GenerateEd25519PrivateKey() (PrivateKey, error) {
	_, k, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	s := "ed25519.prv." + base58.Encode(k)
	return PrivateKey(s), nil
}

func NewPrivateKey(seed []byte) PrivateKey {
	k := ed25519.NewKeyFromSeed(seed)
	s := "ed25519.prv." + base58.Encode(k)
	return PrivateKey(s)
}

func NewPublicKey(publicKey ed25519.PublicKey) PublicKey {
	s := "ed25519." + base58.Encode(publicKey)
	return PublicKey(s)
}

func parse25519PublicKey(s string) (ed25519.PublicKey, error) {
	if !strings.HasPrefix(s, "ed25519.") {
		return nil, errors.Error("invalid key type")
	}
	b58 := strings.Replace(s, "ed25519.", "", 1)
	b, err := base58.Decode(b58)
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not decode key"))
	}

	return ed25519.PublicKey(b), nil
}

func parse25519PrivateKey(s string) (ed25519.PrivateKey, error) {
	if !strings.HasPrefix(s, "ed25519.prv.") {
		return nil, errors.Error("invalid key type")
	}
	b58 := strings.Replace(s, "ed25519.prv.", "", 1)
	b, err := base58.Decode(b58)
	if err != nil {
		return nil, errors.Wrap(err, errors.New("could not decode key"))
	}

	return ed25519.PrivateKey(b), nil
}

func (i PrivateKey) ed25519() ed25519.PrivateKey {
	k, _ := parse25519PrivateKey(string(i))
	return k
}

func (i PrivateKey) PublicKey() PublicKey {
	return NewPublicKey(i.ed25519().Public().(ed25519.PublicKey))
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

// NewEphemeralSharedKey creates a new ec25519 key pair, calculates a shared
// secret given a public key, and returns the created public key and secret
func NewEphemeralSharedKey(pub PublicKey) (*PublicKey, []byte, error) {
	priv, err := GenerateEd25519PrivateKey()
	if err != nil {
		return nil, nil, err
	}
	ca := privateEd25519KeyToCurve25519(priv.ed25519())
	cB := publicEd25519KeyToCurve25519(pub.ed25519())
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return nil, nil, err
	}
	A := priv.PublicKey()
	return &A, ss, nil
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
	k, _ := parse25519PublicKey(string(r))
	return k
}

func (r PublicKey) IsEmpty() bool {
	return r == ""
}

func (r PublicKey) Bytes() []byte {
	out := make([]byte, 32)
	for i, b := range r.ed25519() {
		out[i] = b
	}
	return out
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

func (r PublicKey) Equals(w PublicKey) bool {
	return string(r) == string(w)
}
