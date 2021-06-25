package object

import (
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
)

func NewCertificate(
	issuer crypto.PrivateKey,
	req CertificateRequest,
	sign bool,
	notes string,
) (*CertificateResponse, error) {
	if req.Metadata.Signature.IsEmpty() {
		return nil, errors.Error("missing signature")
	}
	now := time.Now().UTC()
	exp := now.Add(time.Hour * 24 * 365)
	nowString := now.Format(time.RFC3339)
	expString := exp.Format(time.RFC3339)
	crt := &Certificate{
		Metadata: Metadata{
			Owner:    issuer.PublicKey(),
			Datetime: nowString,
		},
		Nonce:       req.Nonce,
		Subject:     req.Metadata.Owner,
		Permissions: req.Permissions,
		Starts:      nowString,
		Expires:     expString,
	}
	co, err := Marshal(crt)
	if err != nil {
		return nil, err
	}
	crtSig, err := NewSignature(issuer, co)
	if err != nil {
		return nil, err
	}
	crt.Metadata.Signature = crtSig
	res := &CertificateResponse{
		Metadata: Metadata{
			Owner:    issuer.PublicKey(),
			Datetime: nowString,
		},
		Signed:      sign,
		Notes:       notes,
		Request:     req,
		Certificate: *crt,
	}
	reso, err := Marshal(res)
	if err != nil {
		return nil, err
	}
	resSig, err := NewSignature(issuer, reso)
	if err != nil {
		return nil, err
	}
	res.Metadata.Signature = resSig
	return res, nil
}
