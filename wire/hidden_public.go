package wire

import (
	"github.com/keybase/saltpack"
)

type HiddenPublicKey struct {
	publicKey saltpack.BoxPublicKey
}

// CreateEphemeralKey creates a random ephemeral key.
func (pk *HiddenPublicKey) CreateEphemeralKey() (saltpack.BoxSecretKey, error) {
	return pk.publicKey.CreateEphemeralKey()
}

// ToKID outputs the "key ID" that corresponds to this key.
// You can do whatever you'd like here, but probably it makes sense just
// to output the public key as is.
func (pk *HiddenPublicKey) ToKID() []byte {
	return pk.publicKey.ToKID()
}

// ToRawBoxKeyPointer returns this public key as a *[32]byte,
// for use with nacl.box.Seal
func (pk *HiddenPublicKey) ToRawBoxKeyPointer() *saltpack.RawBoxKey {
	return pk.publicKey.ToRawBoxKeyPointer()
}

// HideIdentity returns true if we should hide the identity of this
// key in our output message format.
func (pk *HiddenPublicKey) HideIdentity() bool {
	return true
}
