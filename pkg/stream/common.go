package stream

import (
	"strings"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func GetAllowsKeysFromPolicies(os ...object.Object) []crypto.PublicKey {
	// TODO this currently only accepts allow actions
	pkm := map[crypto.PublicKey]struct{}{}
	for _, o := range os {
		for _, s := range o.GetOwners() {
			pkm[s] = struct{}{}
		}
		p := o.GetPolicy()
		for _, a := range p.Actions {
			if strings.EqualFold(a, "allow") {
				for _, s := range p.Subjects {
					pkm[crypto.PublicKey(s)] = struct{}{}
				}
			}
		}
	}
	pks := []crypto.PublicKey{}
	for k := range pkm {
		pks = append(pks, k)
	}
	return pks
}

func GetStreamLeaves(os []*object.Object) []*object.Object {
	hm := map[string]bool{} // map[hash]isParent
	om := map[string]*object.Object{}
	for _, o := range os {
		h := o.Hash().String()
		if _, ok := hm[h]; !ok {
			hm[h] = false
		}
		for _, p := range o.GetParents() {
			hm[p.String()] = true
		}
		om[h] = o
	}

	os = []*object.Object{}
	for h, isParent := range hm {
		if !isParent {
			os = append(os, om[h])
		}
	}

	return os
}
