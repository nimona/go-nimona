package crypto

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"strings"

	"nimona.io/internal/encoding/base58"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
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
	// p := k.Public().(ed25519.PublicKey)
	return PrivateKey(k)
}

func NewPublicKey(publicKey ed25519.PublicKey) PublicKey {
	s := "ed25519." + base58.Encode(publicKey)
	return PublicKey(s)
}

func parse25519PublicKey(s string) (ed25519.PublicKey, error) {
	if strings.HasPrefix(s, "ed25519.") == false {
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
	if strings.HasPrefix(s, "ed25519.prv.") == false {
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
	out := make([]byte, 32)
	for i, b := range i.ed25519() {
		out[i] = b
	}
	return out
}

func (i PrivateKey) Sign(message []byte) []byte {
	return ed25519.Sign(i.ed25519(), message)
}

func (i PrivateKey) raw() crypto.PrivateKey {
	return i
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

func (r PublicKey) raw() crypto.PublicKey {
	return ed25519.PublicKey(r)
}

func (r PublicKey) ToObject() object.Object {
	o := object.New()
	o.Set("@type:s", "ed25519")
	o.Set("x:s", strings.Replace(string(r), "ed25519.", "", 1))
	return o
}

func (r PublicKey) Equals(w PublicKey) bool {
	return string(r) == string(w)
	// return subtle.ConstantTimeCompare(w.ed25519(), r.ed25519()) == 1
}

func (r *PublicKey) FromObject(o object.Object) error {
	v := o.Get("x:s")
	s, ok := v.(string)
	if !ok {
		return errors.New("invalid x type")
	}
	*r = PublicKey("ed25519." + s)
	return nil
}
