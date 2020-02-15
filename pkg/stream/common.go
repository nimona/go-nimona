package stream

import (
	"strings"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

// TODO(geoah) remove all helpers

func GetParents(o object.Object) []object.Hash {
	return o.Header.Parents
}

func GetStream(o object.Object) object.Hash {
	return o.Header.Stream
}

func GetPolicy(o object.Object) object.Policy {
	return o.Header.Policy
}

func GetOwners(o object.Object) []crypto.PublicKey {
	return o.Header.Owners
}

func GetSigner(o object.Object) crypto.PublicKey {
	return o.Header.Signature.Signer
}

func GetAllowsKeysFromPolicies(os ...object.Object) []crypto.PublicKey {
	// TODO this currently only accepts allow actions
	pks := []crypto.PublicKey{}
	for _, o := range os {
		p := GetPolicy(o)
		for _, a := range p.Actions {
			switch strings.ToLower(a) {
			case "allow":
				for _, s := range p.Subjects {
					pks = append(pks, crypto.PublicKey(s))
				}
			}
		}
	}
	return pks
}

func GetStreamLeaves(os []object.Object) []object.Object {
	hm := map[string]bool{} // map[hash]isParent
	om := map[string]object.Object{}
	for _, o := range os {
		h := object.NewHash(o).String()
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
		if !isParent {
			os = append(os, om[h])
		}
	}

	return os
}
