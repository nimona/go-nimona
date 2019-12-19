package stream

import (
	json "encoding/json"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
)

type (
	common struct {
		Type      string            `json:"@ctx:s,omitempty"`
		Context   string            `json:"@type:s,omitempty"`
		Stream    object.Hash       `json:"stream:s,omitempty"`
		Parents   []object.Hash     `json:"parents:as,omitempty"`
		Policies  []*Policy         `json:"policies:ao,omitempty"`
		Signature *crypto.Signature `json:"@signature:o,omitempty"`
		Identity  crypto.PublicKey  `json:"@identity:o"`
	}
)

func toCommon(o object.Object) *common {
	c := &common{}
	b, _ := json.Marshal(map[string]interface{}(o)) // nolint: errcheck
	json.Unmarshal(b, c)                            // nolint: errcheck
	return c
}

func Parents(o object.Object) []object.Hash {
	return toCommon(o).Parents
}

func Stream(o object.Object) object.Hash {
	return toCommon(o).Stream
}

func Policies(o object.Object) []*Policy {
	return toCommon(o).Policies
}

func Identity(o object.Object) crypto.PublicKey {
	return toCommon(o).Identity
}

func GetStreamLeaves(os []object.Object) []object.Object {
	hm := map[string]bool{} // map[hash]isParent
	om := map[string]object.Object{}
	for _, o := range os {
		h := hash.New(o).String()
		if _, ok := hm[h]; !ok {
			hm[h] = false
		}
		for _, p := range Parents(o) {
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
