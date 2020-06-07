package peer

import (
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func NewCertificate(
	subject crypto.PublicKey,
	issuer crypto.PrivateKey,
) Certificate {
	c := Certificate{
		Policy: object.Policy{
			Subjects: []string{
				subject.String(),
			},
		},
		Created: time.Now().Format(time.RFC3339),
		Expires: time.Now().Add(time.Hour * 24 * 365).Format(time.RFC3339),
	}
	s, _ := object.NewSignature(issuer, c.ToObject())
	c.Signatures = append(c.Signatures, s)
	return c
}

func NewSelfSignedCertificate(k crypto.PrivateKey) Certificate {
	return NewCertificate(k.PublicKey(), k)
}
