package object

import (
	"time"

	"nimona.io/pkg/crypto"
)

func NewCertificate(
	issuer crypto.PrivateKey,
	req CertificateRequest,
) (*Certificate, error) {
	c := &Certificate{
		Metadata: Metadata{
			Owner: issuer.PublicKey(),
			Datetime: time.Now().
				UTC().
				Format(time.RFC3339),
		},
		Nonce:                  req.Nonce,
		VendorName:             req.VendorName,
		ApplicationName:        req.ApplicationName,
		ApplicationDescription: req.ApplicationDescription,
		ApplicationURL:         req.ApplicationURL,
		Subject:                req.Metadata.Owner,
		Permissions:            req.Permissions,
		Starts: time.Now().
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
