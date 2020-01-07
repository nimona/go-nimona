package sql

import (
	"github.com/gobwas/glob"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

// LookupOptions
type (
	LookupOption  func(*LookupOptions)
	LookupOptions struct {
		// Lookups are used to perform db queries for these filters
		// TODO find a better name for this
		Lookups struct {
			ObjectHashes []object.Hash
			StreamHashes []object.Hash
			ContentTypes []string
			Signers      []crypto.PublicKey
			Identities   []crypto.PublicKey
		}
		// filters are the lookups equivalents for matching objects for pubsub
		Filters []SqlStoreFilter
	}
)

func newLookupOptions(lookupOptions ...LookupOption) LookupOptions {
	options := &LookupOptions{
		Lookups: struct {
			ObjectHashes []object.Hash
			StreamHashes []object.Hash
			ContentTypes []string
			Signers      []crypto.PublicKey
			Identities   []crypto.PublicKey
		}{
			ObjectHashes: []object.Hash{},
			StreamHashes: []object.Hash{},
			ContentTypes: []string{},
			Signers:      []crypto.PublicKey{},
			Identities:   []crypto.PublicKey{},
		},
		Filters: []SqlStoreFilter{},
	}
	for _, lookupOption := range lookupOptions {
		lookupOption(options)
	}
	return *options
}

func FilterByHash(h object.Hash) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.ObjectHashes = append(opts.Lookups.ObjectHashes, h)
		opts.Filters = append(opts.Filters, func(o object.Object) bool {
			return hash.New(o) == h
		})
	}
}

func FilterBySigner(h crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.Signers = append(opts.Lookups.Signers, h)
		opts.Filters = append(opts.Filters, func(o object.Object) bool {
			return !h.IsEmpty() && h.Equals(stream.GetSigner(o))
		})
	}
}

func FilterByIdentity(h crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.Identities = append(opts.Lookups.Identities, h)
		opts.Filters = append(opts.Filters, func(o object.Object) bool {
			return !h.IsEmpty() && h.Equals(stream.GetIdentity(o))
		})
	}
}

func FilterByStreamHash(h object.Hash) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.StreamHashes = append(opts.Lookups.StreamHashes, h)
		opts.Filters = append(opts.Filters, func(o object.Object) bool {
			os := o.Get("@stream:s")
			switch oh := os.(type) {
			case object.Hash:
				return h.IsEqual(oh)
			case string:
				return h.String() == os
			default:
				return false
			}
		})
	}
}

func FilterByObjectType(typePatterns ...string) LookupOption {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(errors.Wrap(err, errors.New("invalid pattern")))
		}
		patterns[i] = g
	}
	return func(opts *LookupOptions) {
		opts.Lookups.ContentTypes = append(opts.Lookups.ContentTypes, typePatterns...)
		opts.Filters = append(opts.Filters, func(o object.Object) bool {
			for _, pattern := range patterns {
				if pattern.Match(o.GetType()) {
					return true
				}
			}
			return false
		})
	}
}
