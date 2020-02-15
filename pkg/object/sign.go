package object

import (
	"nimona.io/pkg/crypto"
)

// Sign any object (container) with given key and return a signature object (container)
// TODO(geoah) remove Sign method and let devs deal with setting the signature
func Sign(o *Object, key crypto.PrivateKey) error {
	sig, err := NewSignature(key, *o)
	if err != nil {
		return err
	}

	o.Header.Signature = sig
	return nil
}
