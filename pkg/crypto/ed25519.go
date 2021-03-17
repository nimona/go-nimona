package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/multiformats/go-varint"
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
	Ed25519PublicKey struct {
		t KeyType
		k ed25519.PublicKey
	}
	Ed25519PrivateKey struct {
		t KeyType
		k ed25519.PrivateKey
		p Ed25519PublicKey
	}
)

func (k Ed25519PublicKey) Type() KeyType {
	return k.t
}

func (k Ed25519PublicKey) String() string {
	return encodeToCID(cidEd25519Public, uint64(k.t), k.k)
}

func (k Ed25519PublicKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k *Ed25519PublicKey) UnmarshalString(s string) error {
	return k.String(), nil
}

func (k Ed25519PrivateKey) Type() KeyType {
	return k.t
}

func (k Ed25519PrivateKey) String() string {
	return encodeToCID(cidEd25519Private, uint64(k.t), k.k)
}

func (k Ed25519PrivateKey) MarshalString() (string, error) {
	return k.String(), nil
}

func (k *Ed25519PrivateKey) UnmarshalString(s string) error {
	c, err := cid.Decode(s)
	if err != nil {
		return err
	}

	if c.Type() != cidEd25519Private {
		return errors.Error("invalid or unsupported private key type")
	}

	h, err := multihash.Decode(c.Hash())
	if err != nil {
		return err
	}
	
	return nil
}

func (k Ed25519PrivateKey) String() PublicKey {
	return encodeToCID(cidEd25519Private, uint64(k.t), k.k)
}

// func ed25519PrivateToPrivateKey(k ed25519.PrivateKey) (PrivateKey, error) {
// 	h, err := multihash.Encode(k, multihash.IDENTITY)
// 	if err != nil {
// 		panic(err)
// 	}
// 	c := cid.NewCidV1(cidEd25519Private, h)
// 	s, err := multibase.Encode(multibase.Base32, c.Bytes())
// 	if err != nil {
// 		panic(err)
// 	}
// 	return PrivateKey(s), nil
// }

// func ed25519PrivateFromPrivateKey(k PrivateKey) (ed25519.PrivateKey, error) {
// 	c, err := cid.Decode(string(k))
// 	if err != nil {
// 		return nil, err
// 	}
// 	if c.Type() != cidEd25519Private {
// 		return nil, errors.Error("invalid or unsupported private key type")
// 	}
// 	h, err := multihash.Decode(c.Hash())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ed25519.PrivateKey(h.Digest), nil
// }

// func ed25519PublicFromPublicKey(k PublicKey) (ed25519.PublicKey, error) {
// 	c, err := cid.Decode(string(k))
// 	if err != nil {
// 		return nil, err
// 	}
// 	if c.Type() != cidEd25519Public {
// 		return nil, errors.Error("invalid or unsupported public key type")
// 	}
// 	h, err := multihash.Decode(c.Hash())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return ed25519.PublicKey(h.Digest), nil
// }

// func ed25519PublicToPublicKey(k ed25519.PublicKey) (PublicKey, error) {
// 	h, err := multihash.Encode(k, multihash.IDENTITY)
// 	if err != nil {
// 		panic(err)
// 	}
// 	c := cid.NewCidV1(cidEd25519Public, h)
// 	s, err := multibase.Encode(multibase.Base32, c.Bytes())
// 	if err != nil {
// 		panic(err)
// 	}
// 	return PublicKey(s), nil
// }

func GenerateEd25519PrivateKey(keyType KeyType) (PrivateKey, error) {
	_, k, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &Ed25519PrivateKey{
		t: keyType,
		k: k,
		p: &Ed25519PublicKey{
			t: keyType,
		},
	}, nil
}

// func NewPrivateKey(seed []byte) PrivateKey {
// 	b := ed25519.NewKeyFromSeed(seed)
// 	k, err := ed25519PrivateToPrivateKey(b)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return k
// }

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
	ed25519Priv, ok := priv.(*Ed25519PrivateKey)
	if !ok {
		return nil, ErrOnlyEd25519KeysSupported
	}
	ed25519Pub, ok := pub.(*Ed25519PublicKey)
	if !ok {
		return nil, ErrOnlyEd25519KeysSupported
	}
	ca := privateEd25519KeyToCurve25519(ed25519Priv.k)
	cB := publicEd25519KeyToCurve25519(ed25519Pub.k)
	ss, err := curve25519.X25519(ca, cB)
	if err != nil {
		return nil, fmt.Errorf("error getting x25519, %w", err)
	}
	return ss, nil
}

// // NewSharedKey calculates a shared secret given a private and a public key,
// // and returns it
// func NewSharedKey(priv PrivateKey, pub PublicKey) (*PrivateKey, []byte, error) {
// 	ca := privateEd25519KeyToCurve25519(priv.ed25519())
// 	cB := publicEd25519KeyToCurve25519(pub.ed25519())
// 	ss, err := curve25519.X25519(ca, cB)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	return &priv, ss, nil
// }

// CalculateEphemeralSharedKey creates a new ec25519 key pair, calculates a
// shared secret given a public key, and returns the created public key and
// secret
func CalculateEphemeralSharedKey(pub PublicKey) (*PrivateKey, []byte, error) {
	priv, err := GenerateEd25519PrivateKey()
	if err != nil {
		return nil, nil, err
	}
	return NewSharedKey(priv, pub)
}

// func (i PrivateKey) Sign(message []byte) []byte {
// 	return ed25519.Sign(i.ed25519(), message)
// }

// func (r PublicKey) Verify(message []byte, signature []byte) error {
// 	ok := ed25519.Verify(r.ed25519(), message, signature)
// 	if !ok {
// 		return errors.Error("invalid signature")
// 	}
// 	return nil
// }

// // TODO invalid public keys should not be evaluated for equality
// func (r PublicKey) Equals(w PublicKey) bool {
// 	if r == w {
// 		return true
// 	}
// 	ew := w.ed25519()
// 	if ew == nil {
// 		return false
// 	}
// 	rw := r.ed25519()
// 	if rw == nil {
// 		return false
// 	}
// 	return ew.Equal(rw)
// }

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
