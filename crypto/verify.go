package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"math/big"
)

var (
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.New("could not verify signature")
)

func Verify(signature *Signature, digest []byte) error {
	mKey := signature.Key.Materialize()
	pKey, ok := mKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("only ecdsa public keys are currently supported")
	}

	hash := sha256.Sum256(digest)
	rBytes := new(big.Int).SetBytes(signature.Signature[0:32])
	sBytes := new(big.Int).SetBytes(signature.Signature[32:64])

	if ok := ecdsa.Verify(pKey, hash[:], rBytes, sBytes); !ok {
		return ErrCouldNotVerify
	}

	return nil
}
