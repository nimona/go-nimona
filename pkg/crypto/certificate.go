package crypto

import "time"

func NewCertificate(subject PublicKey, issuer PrivateKey) *Certificate {
	c := &Certificate{
		Subject: subject,
		Created: time.Now().Format(time.RFC3339),
		Expires: time.Now().Add(time.Hour * 24 * 365).Format(time.RFC3339),
	}
	s, _ := NewSignature(issuer, c.ToObject())
	c.Signature = s
	return c
}

func NewSelfSignedCertificate(k PrivateKey) *Certificate {
	return NewCertificate(k.PublicKey(), k)
}
