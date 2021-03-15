package objectmanager

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

// LookupOptions
type (
	LookupOption  func(*LookupOptions)
	LookupOptions struct {
		// Lookups are used to perform db queries for these filters
		// TODO find a better name for this
		Lookups struct {
			ObjectCIDs   []object.CID
			StreamCIDs   []object.CID
			ContentTypes []string
			Owners       []crypto.PublicKey
		}
		// filters are the lookups equivalents for matching objects for pubsub
		Filters []ObjectFilter
	}
)

func newLookupOptions(lookupOptions ...LookupOption) LookupOptions {
	options := &LookupOptions{
		Lookups: struct {
			ObjectCIDs   []object.CID
			StreamCIDs   []object.CID
			ContentTypes []string
			Owners       []crypto.PublicKey
		}{
			ObjectCIDs:   []object.CID{},
			StreamCIDs:   []object.CID{},
			ContentTypes: []string{},
			Owners:       []crypto.PublicKey{},
		},
		Filters: []ObjectFilter{},
	}
	for _, lookupOption := range lookupOptions {
		lookupOption(options)
	}
	return *options
}

func FilterByCID(hs ...object.CID) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.ObjectCIDs = append(opts.Lookups.ObjectCIDs, hs...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, h := range hs {
				if !h.IsEmpty() && o != nil && o.CID() == h {
					return true
				}
			}
			return false
		})
	}
}

func FilterByOwner(hs ...crypto.PublicKey) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.Owners = append(opts.Lookups.Owners, hs...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, h := range hs {
				owner := o.Metadata.Owner
				if !owner.IsEmpty() && owner.Equals(h) {
					return true
				}
			}
			return false
		})
	}
}

func FilterByStreamCID(hs ...object.CID) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.StreamCIDs = append(opts.Lookups.StreamCIDs, hs...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, h := range hs {
				if !h.IsEmpty() && o != nil && h == o.Metadata.Stream {
					return true
				}
			}
			return false
		})
	}
}

func FilterByObjectType(typePatterns ...string) LookupOption {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(fmt.Errorf("invalid pattern: %w", err))
		}
		patterns[i] = g
	}
	return func(opts *LookupOptions) {
		opts.Lookups.ContentTypes = append(opts.Lookups.ContentTypes, typePatterns...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, pattern := range patterns {
				if pattern.Match(o.Type) {
					return true
				}
			}
			return false
		})
	}
}
