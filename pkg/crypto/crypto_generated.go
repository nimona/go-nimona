// Code generated by nimona.io/tools/codegen. DO NOT EDIT.

package crypto

import (
	json "encoding/json"

	object "nimona.io/pkg/object"
)

type (
	Signature struct {
		PublicKey *PublicKey `json:"publicKey:o"`
		Algorithm string     `json:"algorithm:s"`
		R         []byte     `json:"r:d"`
		S         []byte     `json:"s:d"`
	}
	PrivateKey struct {
		PublicKey *PublicKey `json:"publicKey:o"`
		KeyType   string     `json:"keyType:s"`
		Algorithm string     `json:"algorithm:s"`
		Curve     string     `json:"curve:s"`
		X         []byte     `json:"x:d"`
		Y         []byte     `json:"y:d"`
		D         []byte     `json:"d:d"`
	}
	PublicKey struct {
		KeyType   string     `json:"keyType:s"`
		Algorithm string     `json:"algorithm:s"`
		Curve     string     `json:"curve:s"`
		X         []byte     `json:"x:d"`
		Y         []byte     `json:"y:d"`
		Signature *Signature `json:"@signature:o"`
	}
)

func (e *Signature) ContextName() string {
	return "nimona.io/crypto"
}

func (e *Signature) GetType() string {
	return "Signature"
}

func (e *Signature) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "Signature",
		"@struct:s": "Signature",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *Signature) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *PrivateKey) ContextName() string {
	return "nimona.io/crypto"
}

func (e *PrivateKey) GetType() string {
	return "PrivateKey"
}

func (e *PrivateKey) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "PrivateKey",
		"@struct:s": "PrivateKey",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *PrivateKey) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}

func (e *PublicKey) ContextName() string {
	return "nimona.io/crypto"
}

func (e *PublicKey) GetType() string {
	return "PublicKey"
}

func (e *PublicKey) ToObject() object.Object {
	m := map[string]interface{}{
		"@ctx:s":    "PublicKey",
		"@struct:s": "PublicKey",
	}
	b, _ := json.Marshal(e)
	json.Unmarshal(b, &m)
	return object.Object(m)
}

func (e *PublicKey) FromObject(o object.Object) error {
	b, _ := json.Marshal(map[string]interface{}(o))
	return json.Unmarshal(b, e)
}
