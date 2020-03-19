package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"

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
	return PrivateKey(k)
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

// func (i PrivateKey) Shared(r PublicKey) []byte {
// this requires a curve25519
// 	var shared [32]byte
// 	ib := i.Bytes()
// 	rb := r.Bytes()
// 	curve25519.ScalarMult(&shared, &ib, &rb)
// 	return shared[:]
// }

func (i PrivateKey) IsEmpty() bool {
	return i == ""
}

func (i PrivateKey) Bytes() []byte {
	k := i.ed25519().Seed()
	out := make([]byte, len(k))
	copy(out, k)
	return out
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
