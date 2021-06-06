// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package object

import (
	crypto "nimona.io/pkg/crypto"
)

type (
	Certificate struct {
		Metadata    Metadata                `nimona:"@metadata:m"`
		Nonce       string                  `nimona:"nonce:s"`
		Subject     crypto.PublicKey        `nimona:"subject:s"`
		Permissions []CertificatePermission `nimona:"permissions:am"`
		Starts      string                  `nimona:"starts:s"`
		Expires     string                  `nimona:"expires:s"`
	}
	CertificatePermission struct {
		Metadata Metadata `nimona:"@metadata:m"`
		Types    []string `nimona:"types:as"`
		Actions  []string `nimona:"actions:as"`
	}
	CertificateRequest struct {
		Metadata               Metadata                `nimona:"@metadata:m"`
		Nonce                  string                  `nimona:"nonce:s"`
		VendorName             string                  `nimona:"vendorName:s"`
		VendorURL              string                  `nimona:"vendorURL:s"`
		ApplicationName        string                  `nimona:"applicationName:s"`
		ApplicationDescription string                  `nimona:"applicationDescription:s"`
		ApplicationURL         string                  `nimona:"applicationURL:s"`
		Permissions            []CertificatePermission `nimona:"permissions:am"`
	}
	CertificateResponse struct {
		Metadata    Metadata           `nimona:"@metadata:m"`
		Signed      bool               `nimona:"signed:b"`
		Notes       string             `nimona:"notes:s"`
		Request     CertificateRequest `nimona:"request:m"`
		Certificate Certificate        `nimona:"certificate:m"`
	}
)

func (e *Certificate) Type() string {
	return "nimona.io/Certificate"
}

func (e *Certificate) MarshalObject() (*Object, error) {
	o, err := Marshal(e)
	if err != nil {
		return nil, err
	}
	o.Type = "nimona.io/Certificate"
	return o, nil
}

func (e *Certificate) UnmarshalObject(o *Object) error {
	return Unmarshal(o, e)
}

func (e *CertificatePermission) Type() string {
	return "nimona.io/CertificatePermission"
}

func (e *CertificatePermission) MarshalObject() (*Object, error) {
	o, err := Marshal(e)
	if err != nil {
		return nil, err
	}
	o.Type = "nimona.io/CertificatePermission"
	return o, nil
}

func (e *CertificatePermission) UnmarshalObject(o *Object) error {
	return Unmarshal(o, e)
}

func (e *CertificateRequest) Type() string {
	return "nimona.io/CertificateRequest"
}

func (e *CertificateRequest) MarshalObject() (*Object, error) {
	o, err := Marshal(e)
	if err != nil {
		return nil, err
	}
	o.Type = "nimona.io/CertificateRequest"
	return o, nil
}

func (e *CertificateRequest) UnmarshalObject(o *Object) error {
	return Unmarshal(o, e)
}

func (e *CertificateResponse) Type() string {
	return "nimona.io/CertificateResponse"
}

func (e *CertificateResponse) MarshalObject() (*Object, error) {
	o, err := Marshal(e)
	if err != nil {
		return nil, err
	}
	o.Type = "nimona.io/CertificateResponse"
	return o, nil
}

func (e *CertificateResponse) UnmarshalObject(o *Object) error {
	return Unmarshal(o, e)
}
