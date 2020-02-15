package object

import (
	"nimona.io/pkg/crypto"
)

// Sign any object (container) with given key and return a signature object (container)
func Sign(o Object, key crypto.PrivateKey) error {
	sig, err := NewSignature(key, o)
	if err != nil {
		return err
	}

	o.SetSignature(sig.ToObject())
	return nil
}
