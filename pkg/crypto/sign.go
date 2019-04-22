package crypto

import "nimona.io/pkg/object"

// Sign any object (container) with given key and return a signature object (container)
func Sign(o *object.Object, key *PrivateKey) error {
	o.SetSignerKey(key.PublicKey.ToObject())

	sig, err := NewSignature(key, AlgorithmObjectHash, o)
	if err != nil {
		return err
	}

	o.SetSignature(sig.ToObject())

	return nil
}
