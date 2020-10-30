package stream

import (
	"strings"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func GetAllowsKeysFromPolicies(os ...*object.Object) []crypto.PublicKey {
	// TODO this currently only accepts allow actions
	pkm := map[crypto.PublicKey]struct{}{}
	for _, o := range os {
		owner := o.Metadata.Owner
		if !owner.IsEmpty() {
			pkm[owner] = struct{}{}
		}
		p := o.Metadata.Policy
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
