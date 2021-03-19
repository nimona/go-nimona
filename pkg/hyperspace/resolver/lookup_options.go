package resolver

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/object"
)

// LookupOptions
type (
	LookupFilter  func(*hyperspace.Announcement) bool
	LookupOption  func(*LookupOptions)
	LookupOptions struct {
		Local bool
		// Lookups are strings we are looking for, will be used to create a
		// bloom filter when forwarding the lookup request to providers
		Lookups []string
		// filters are the lookups equivalents for matching local peers
		Filters []LookupFilter
	}
)

func ParseLookupOptions(opts ...LookupOption) *LookupOptions {
	options := &LookupOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

func (l LookupOptions) Match(p *hyperspace.Announcement) bool {
	for _, f := range l.Filters {
		if !f(p) {
			return false
		}
	}
	return true
}

// LookupOnlyLocal forces the discoverer to only look at its cache
func LookupOnlyLocal() LookupOption {
	return func(opts *LookupOptions) {
		opts.Local = true
	}
}

// LookupByCID matches content cids
func LookupByCID(cid object.CID) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups = append(opts.Lookups, cid.String())
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				return hyperspace.Bloom(p.PeerVector).Test(
					hyperspace.New(cid.String()),
				)
			},
		)
	}
}

// LookupByPeerKey matches the peer key
func LookupByPeerKey(keys ...*crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		for _, key := range keys {
			opts.Lookups = append(opts.Lookups, key.String())
		}
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				for _, key := range keys {
					// TODO check announcement signature
					owner := p.ConnectionInfo.PublicKey
					if owner.Equals(key) {
						return true
					}
					sig := p.Metadata.Signature
					if sig.Signer.Equals(key) {
						return true
					}
				}
				return false
			},
		)
	}
}
