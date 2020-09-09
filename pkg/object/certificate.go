package object

import (
	"time"

	"nimona.io/pkg/crypto"
)

func NewCertificate(
	subject crypto.PublicKey,
	issuer crypto.PrivateKey,
) (*Certificate, error) {
	c := &Certificate{
		Metadata: Metadata{
			Owner: issuer.PublicKey(),
			Policy: Policy{
				Subjects: []string{
					subject.String(),
				},
			},
		},
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
	return NewCertificate(k.PublicKey(), k)
}
