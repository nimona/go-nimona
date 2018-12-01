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

// Verify signature given the signer's key and payload digest
func Verify(sig *Signature, key *Key, digest []byte) error {
	mKey := key.Materialize()
	pKey, ok := mKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("only ecdsa public keys are currently supported")
	}

	hash := sha256.Sum256(digest)
	r := new(big.Int).SetBytes(sig.R)
	s := new(big.Int).SetBytes(sig.S)

	if ok := ecdsa.Verify(pKey, hash[:], r, s); !ok {
		return ErrCouldNotVerify
	}

	return nil
}
