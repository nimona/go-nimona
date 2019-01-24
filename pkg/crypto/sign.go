package crypto

import "nimona.io/pkg/object"

// Sign any block (container) with given key and return a signature block (container)
func Sign(o *object.Object, key *Key) error {
	o.SetSignerKey(key.GetPublicKey().ToObject())

	sig, err := NewSignature(key, AlgorithmObjectHash, o)
	if err != nil {
		return err
	}

	o.SetSignature(sig.ToObject())

	return nil
}
