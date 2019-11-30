package discovery

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery/bloom"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

// LookupOptions
type (
	LookupFilter  func(*peer.Peer) bool
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
			func(p *peer.Peer) bool {
				return bloom.Bloom(p.Bloom).Contains(
					bloom.New(hash.String()),
				)
			},
		)
	}
}

// LookupByKey matches the peer key
func LookupByKey(key crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups = append(opts.Lookups, key.String())
		opts.Filters = append(
			opts.Filters,
			func(p *peer.Peer) bool {
				return p.Signature.Signer.Equals(key)
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
			func(p *peer.Peer) bool {
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
			func(p *peer.Peer) bool {
				for _, c := range p.Certificates {
					if certSigner.Equals(c.Signature.Signer) {
						return true
					}
				}
				return false
			},
		)
	}
}

func ParseLookupOptions(opts ...LookupOption) *LookupOptions {
	options := &LookupOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

func matchPeerWithLookupFilters(p *peer.Peer, fs ...LookupFilter) bool {
	for _, f := range fs {
		if f(p) == false {
			return false
		}
	}
	return true
}
