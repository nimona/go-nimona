package stream

import (
	"encoding/json"
	"strings"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
)

type (
	common struct {
		Context   string            `json:"@ctx:s,omitempty"`
		Type      string            `json:"@type:s,omitempty"`
		Stream    object.Hash       `json:"@stream:s,omitempty"`
		Parents   []object.Hash     `json:"@parents:as,omitempty"`
		Policy    *Policy           `json:"@policy:o,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:s"`
	}
)

func toCommon(o object.Object) *common {
	c := &common{}
	b, _ := json.Marshal(map[string]interface{}(o)) // nolint: errcheck
	json.Unmarshal(b, c)                            // nolint: errcheck
	return c
}

func GetParents(o object.Object) []object.Hash {
	return toCommon(o).Parents
}

func GetStream(o object.Object) object.Hash {
	return toCommon(o).Stream
}

func GetPolicy(o object.Object) *Policy {
	return toCommon(o).Policy
}

func GetIdentity(o object.Object) crypto.PublicKey {
	return toCommon(o).Identity
}

func GetSigner(o object.Object) crypto.PublicKey {
	c := toCommon(o)
	if c.Signature == nil || c.Signature.Signer.IsEmpty() {
		return ""
	}
	return c.Signature.Signer
}

func GetAllowsKeysFromPolicies(os ...object.Object) []crypto.PublicKey {
	// TODO this currently only accepts allow actions
	pks := []crypto.PublicKey{}
	for _, o := range os {
		p := GetPolicy(o)
		if p == nil {
			continue
		}
		switch strings.ToLower(p.Action) {
		case "allow":
			for _, s := range p.Subjects {
				pks = append(pks, crypto.PublicKey(s))
			}
		}
	}
	return pks
}

func GetStreamLeaves(os []object.Object) []object.Object {
	hm := map[string]bool{} // map[hash]isParent
	om := map[string]object.Object{}
	for _, o := range os {
		h := hash.New(o).String()
		if _, ok := hm[h]; !ok {
			hm[h] = false
		}
		for _, p := range GetParents(o) {
			hm[p.String()] = true
		}
		om[h] = o
	}

	os = []object.Object{}
	for h, isParent := range hm {
		if isParent == false {
			os = append(os, om[h])
		}
	}

	return os
}
