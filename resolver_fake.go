package nimona

import (
	"fmt"
)

type ResolverFake struct {
	identities map[string]*IdentityInfo
}

func (r *ResolverFake) ResolveIdentityAlias(alias IdentityAlias) (*IdentityInfo, error) {
	if info, ok := r.identities[alias.Hostname]; ok {
		return info, nil
	}
	return nil, fmt.Errorf("identity not found")
}
