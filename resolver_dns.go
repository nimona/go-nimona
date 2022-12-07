package nimona

import (
	"errors"
	"net"
	"sort"
	"strings"
)

var ErrResolverUnsupported = errors.New("resolver unsupported")

type ResolverDNS struct{}

func (r *ResolverDNS) Resolve(nid string) ([]NodeAddr, error) {
	// check if the nid can be resolved by this resolver
	if !strings.HasPrefix(nid, PeerHandlePrefix) {
		return nil, ErrResolverUnsupported
	}

	// remove the prefix from the nid
	hostname := strings.TrimPrefix(nid, PeerHandlePrefix)

	// look up all the TXT DNS entries for the given hostname
	nodeAddrs := []NodeAddr{}
	entries, err := net.LookupTXT(hostname)
	if err != nil {
		// return an empty slice if there was an error looking up the entries
		return nodeAddrs, nil
	}

	// parse the entries that start with "nimona="
	for _, entry := range entries {
		if strings.HasPrefix(entry, "nimona.node.addr=") {
			entry = strings.TrimPrefix(entry, "nimona=")
			nodeAddr := NodeAddr{}
			err := nodeAddr.Parse(entry)
			if err != nil {
				continue
			}
			nodeAddrs = append(nodeAddrs, nodeAddr)
		}
	}

	// sort addresses by hostname
	sort.Slice(nodeAddrs, func(i, j int) bool {
		return nodeAddrs[i].Host < nodeAddrs[j].Host
	})

	return nodeAddrs, nil
}
