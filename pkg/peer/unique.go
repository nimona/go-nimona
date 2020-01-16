package peer

import (
	"nimona.io/pkg/crypto"
)

func Unique(peers []*Peer) []*Peer {
	m := map[crypto.PublicKey]*Peer{}
	for _, p := range peers {
		pk := p.PublicKey()
		xp, ok := m[pk]
		if !ok {
			m[pk] = p
			continue
		}
		if p.Version > xp.Version {
			m[pk] = p
			continue
		}
	}
	r := []*Peer{}
	for _, p := range m {
		r = append(r, p)
	}
	return r
}
