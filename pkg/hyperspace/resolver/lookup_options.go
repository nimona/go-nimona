package resolver

import (
	"fmt"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/did"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/tilde"
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

// LookupByHash matches content hashes
func LookupByHash(hash tilde.Digest) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups = append(opts.Lookups, hash.String())
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				return hyperspace.Bloom(p.PeerVector).Test(
					hyperspace.New(hash.String()),
				)
			},
		)
	}
}

// LookupByPeerKey matches the peer key
func LookupByPeerKey(keys ...crypto.PublicKey) LookupOption {
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
					if sig.Key.Equals(key) {
						return true
					}
				}
				return false
			},
		)
	}
}

// LookupByOwner matches the owner
func LookupByOwner(owners ...did.DID) LookupOption {
	return func(opts *LookupOptions) {
		for _, o := range owners {
			opts.Lookups = append(opts.Lookups, o.String())
		}
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				fmt.Println("!!!")
				fmt.Println("!!!")
				fmt.Println("!!!")
				fmt.Println("!!!")
				for _, o := range owners {
					owner := p.Metadata.Owner
					fmt.Println(">>", owner, o)
					if owner == did.Empty {
						continue
					}
					if owner.Equals(o) {
						return true
					}
				}
				return false
			},
		)
	}
}
