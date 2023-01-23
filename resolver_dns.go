package nimona

import (
	"errors"
	"net"
	"sort"
	"strings"
)

var ErrResolverUnsupported = errors.New("resolver unsupported")

type ResolverDNS struct{}

// TODO add context
func (r *ResolverDNS) Resolve(nID NetworkAlias) ([]PeerAddr, error) {
	// look up all the TXT DNS entries for the given hostname
	peerAddrs := []PeerAddr{}
	entries, err := net.LookupTXT(nID.Hostname)
	if err != nil {
		// return an empty slice if there was an error looking up the entries
		return peerAddrs, nil
	}

	// parse the entries that start with "nimona="
	for _, entry := range entries {
		if strings.HasPrefix(entry, "nimona.node.addr=") {
			entry = strings.TrimPrefix(entry, "nimona=")
			peerAddr, err := ParsePeerAddr(entry)
			if err != nil {
				continue
			}
			peerAddrs = append(peerAddrs, *peerAddr)
		}
	}

	// sort addresses by hostname
	sort.Slice(peerAddrs, func(i, j int) bool {
		return peerAddrs[i].String() < peerAddrs[j].String()
	})

	return peerAddrs, nil
}
