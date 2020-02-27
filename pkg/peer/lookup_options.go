package peer

import (
	"nimona.io/pkg/bloom"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

// LookupOptions
type (
	LookupFilter  func(*Peer) bool
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

func (l LookupOptions) Match(p *Peer) bool {
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

// LookupByContentHash matches content hashes
func LookupByContentHash(hash object.Hash) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups = append(opts.Lookups, hash.String())
		opts.Filters = append(
			opts.Filters,
			func(p *Peer) bool {
				return bloom.Bloom(p.Bloom).Contains(
					bloom.New(hash.String()),
				)
			},
		)
	}
}

// LookupByOwner matches the peer key
func LookupByOwner(keys ...crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		for _, key := range keys {
			opts.Lookups = append(opts.Lookups, key.String())
		}
		opts.Filters = append(
			opts.Filters,
			func(p *Peer) bool {
				for _, key := range keys {
					for _, owner := range p.Owners {
						if owner.Equals(key) {
							return true
						}
					}
					// TODO(geoah) should certs and sigs be considered owners?
					for _, c := range p.Certificates {
						for _, s := range c.Signatures {
							if s.Signer.Equals(key) {
								return true
							}
						}
					}
					for _, s := range p.Signatures {
						if s.Signer.Equals(key) {
							return true
						}
					}
				}
				return false
			},
		)
	}
}

// LookupByContentType matches content hashes
func LookupByContentType(contentType string) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups = append(opts.Lookups, contentType)
		opts.Filters = append(
			opts.Filters,
			func(p *Peer) bool {
				for _, t := range p.ContentTypes {
					if contentType == t {
						return true
					}
				}
				return false
			},
		)
	}
}

// LookupByCertificateSigner matches certificate signers
func LookupByCertificateSigner(certSigner crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups = append(opts.Lookups, certSigner.String())
		opts.Filters = append(
			opts.Filters,
			func(p *Peer) bool {
				for _, c := range p.Certificates {
					for _, s := range c.Signatures {
						if certSigner.Equals(s.Signer) {
							return true
						}
					}
				}
				return false
			},
		)
	}
}
