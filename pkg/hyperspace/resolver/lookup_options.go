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

// LookupByContentHash matches content hashes
func LookupByContentHash(hash object.Hash) LookupOption {
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

// LookupByOwner matches the peer key
func LookupByOwner(keys ...crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		for _, key := range keys {
			opts.Lookups = append(opts.Lookups, key.String())
		}
		opts.Filters = append(
			opts.Filters,
			func(p *hyperspace.Announcement) bool {
				for _, key := range keys {
					// TODO check announcement signature
					owner := p.Peer.PublicKey
					if owner.Equals(key) {
						return true
					}
					// TODO(geoah) should certs and sigs be considered owners?
					// for _, c := range p.Certificates {
					// 	sig := c.Metadata.Signature
					// 	if sig.Signer.Equals(key) {
					// 		return true
					// 	}
					// }
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

// LookupByContentType matches content hashes
// TODO support capabilities
// func LookupByContentType(contentType string) LookupOption {
// 	return func(opts *LookupOptions) {
// 		opts.Lookups = append(opts.Lookups, contentType)
// 		opts.Filters = append(
// 			opts.Filters,
// 			func(p *hyperspace.Announcement) bool {
// 				for _, t := range p.ContentTypes {
// 					if contentType == t {
// 						return true
// 					}
// 				}
// 				return false
// 			},
// 		)
// 	}
// }

// LookupByCertificateSigner matches certificate signers
// TODO support
// func LookupByCertificateSigner(certSigner crypto.PublicKey) LookupOption {
// 	return func(opts *LookupOptions) {
// 		opts.Lookups = append(opts.Lookups, certSigner.String())
// 		opts.Filters = append(
// 			opts.Filters,
// 			func(p *hyperspace.Announcement) bool {
// 				for _, c := range p.Certificates {
// 					sig := c.Metadata.Signature
// 					if certSigner.Equals(sig.Signer) {
// 						return true
// 					}
// 				}
// 				return false
// 			},
// 		)
// 	}
// }
