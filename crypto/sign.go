package crypto

import "nimona.io/go/encoding"

// Sign any block (container) with given key and return a signature block (container)
func Sign(o *encoding.Object, key *Key) (*encoding.Object, error) {
	sig, err := NewSignature(key, AlgorithmObjectHash, o)
	if err != nil {
		return nil, err
	}

	o.SetSignerKey(key.ToObject())
	o.SetSignature(sig.ToObject())

	return o, nil
}
