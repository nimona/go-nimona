package object

import (
	"nimona.io/pkg/tilde"
)

type (
	Parents map[string]tilde.DigestArray
)

func (ps Parents) All() []tilde.Digest {
	var unique []tilde.Digest

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
