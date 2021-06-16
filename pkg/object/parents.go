package object

import (
	"nimona.io/pkg/chore"
)

type (
	Parents map[string]chore.HashArray
)

func (ps Parents) All() []chore.Hash {
	var unique []chore.Hash

	for _, ip := range ps {
		skip := false
		for _, iip := range ip {
			for _, u := range unique {
				if iip == u {
					skip = true
					break
				}
			}
			if !skip {
				unique = append(unique, iip)
			}
		}
	}

	return unique
}
