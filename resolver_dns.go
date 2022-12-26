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
func (r *ResolverDNS) Resolve(nID NetworkID) ([]NodeAddr, error) {
	// look up all the TXT DNS entries for the given hostname
	nodeAddrs := []NodeAddr{}
	entries, err := net.LookupTXT(nID.Hostname)
	if err != nil {
		// return an empty slice if there was an error looking up the entries
		return nodeAddrs, nil
	}

	// parse the entries that start with "nimona="
	for _, entry := range entries {
		if strings.HasPrefix(entry, "nimona.node.addr=") {
			entry = strings.TrimPrefix(entry, "nimona=")
			nodeAddr, err := ParseNodeAddr(entry)
			if err != nil {
				continue
			}
			nodeAddrs = append(nodeAddrs, *nodeAddr)
		}
	}

	// sort addresses by hostname
	sort.Slice(nodeAddrs, func(i, j int) bool {
		return nodeAddrs[i].String() < nodeAddrs[j].String()
	})

	return nodeAddrs, nil
}
