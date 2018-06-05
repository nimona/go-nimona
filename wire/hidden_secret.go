package wire

import (
	"github.com/keybase/saltpack"
)

// HiddenSecretKey that wraps BoxPublicKey to hide the key when used in
// saltpack's encrypt method.
type HiddenSecretKey struct {
	SecretKey saltpack.BoxSecretKey
}

// Box boxes up data, sent from this secret key, and to the receiver
// specified.
func (sk *HiddenSecretKey) Box(receiver saltpack.BoxPublicKey, nonce saltpack.Nonce, msg []byte) []byte {
	return sk.SecretKey.Box(receiver, nonce, msg)
}

// Unbox opens up the box, using this secret key as the receiver key
// abd the give public key as the sender key.
func (sk *HiddenSecretKey) Unbox(sender saltpack.BoxPublicKey, nonce saltpack.Nonce, msg []byte) ([]byte, error) {
	return sk.SecretKey.Unbox(sender, nonce, msg)
}

// GetPublicKey gets the public key associated with this secret key.
func (sk *HiddenSecretKey) GetPublicKey() saltpack.BoxPublicKey {
	return &HiddenPublicKey{sk.SecretKey.GetPublicKey()}
}

// Precompute computes a DH with the given key
func (sk *HiddenSecretKey) Precompute(peer saltpack.BoxPublicKey) saltpack.BoxPrecomputedSharedKey {
	return sk.SecretKey.Precompute(peer)
}
