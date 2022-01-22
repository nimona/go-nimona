package resolver

import (
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
		// lookups
		DID    did.DID
		Digest tilde.Digest
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

// LookupByDigest matches content digest
func LookupByDigest(digest tilde.Digest) LookupOption {
	return func(opts *LookupOptions) {
		opts.Digest = digest
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				for _, d := range p.Digests {
					if d.Equal(digest) {
						return true
					}
				}
				return false
			},
		)
	}
}

// LookupByDID matches the owner or connection public key
func LookupByDID(id did.DID) LookupOption {
	return func(opts *LookupOptions) {
		opts.DID = id
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				owner := p.Metadata.Owner
				if owner == did.Empty {
					return true
				}
				if owner.Equals(id) {
					return true
				}
				return false
			},
		)
	}
}
