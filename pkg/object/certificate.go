package object

import (
	"time"

	"nimona.io/pkg/crypto"
)

func NewCertificate(
	issuer crypto.PrivateKey,
	subjects ...crypto.PublicKey,
) (*Certificate, error) {
	c := &Certificate{
		Metadata: Metadata{
			Owner: issuer.PublicKey(),
		},
		Subjects: subjects,
		Created: time.Now().
			UTC().
			Format(time.RFC3339),
		Expires: time.Now().
			UTC().
			Add(time.Hour * 24 * 365).
			Format(time.RFC3339),
	}
	s, err := NewSignature(issuer, c.ToObject())
	if err != nil {
		return nil, err
	}
	c.Metadata.Signature = s
	return c, nil
}

func NewCertificateSelfSigned(k crypto.PrivateKey) (*Certificate, error) {
	return NewCertificate(k, k.PublicKey())
}
