package objectmanager

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"
)

// LookupOptions
type (
	LookupOption  func(*LookupOptions)
	LookupOptions struct {
		// Lookups are used to perform db queries for these filters
		// TODO find a better name for this
		Lookups struct {
			ObjectHashes []tilde.Digest
			StreamHashes []tilde.Digest
			ContentTypes []string
			Owners       []peer.ID
		}
		// filters are the lookups equivalents for matching objects for pubsub
		Filters []ObjectFilter
	}
)

func newLookupOptions(lookupOptions ...LookupOption) LookupOptions {
	options := &LookupOptions{
		Lookups: struct {
			ObjectHashes []tilde.Digest
			StreamHashes []tilde.Digest
			ContentTypes []string
			Owners       []peer.ID
		}{
			ObjectHashes: []tilde.Digest{},
			StreamHashes: []tilde.Digest{},
			ContentTypes: []string{},
			Owners:       []peer.ID{},
		},
		Filters: []ObjectFilter{},
	}
	for _, lookupOption := range lookupOptions {
		lookupOption(options)
	}
	return *options
}

func FilterByHash(hs ...tilde.Digest) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.ObjectHashes = append(opts.Lookups.ObjectHashes, hs...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, h := range hs {
				if !h.IsEmpty() && o != nil && o.Hash().Equal(h) {
					return true
				}
			}
			return false
		})
	}
}

func FilterByOwner(hs ...peer.ID) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.Owners = append(opts.Lookups.Owners, hs...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, h := range hs {
				owner := o.Metadata.Owner
				if !owner.IsEmpty() && !h.IsEmpty() && owner.Equals(h) {
					return true
				}
			}
			return false
		})
	}
}

func FilterByStreamHash(hs ...tilde.Digest) LookupOption {
	return func(opts *LookupOptions) {
		opts.Lookups.StreamHashes = append(opts.Lookups.StreamHashes, hs...)
		opts.Filters = append(opts.Filters, func(o *object.Object) bool {
			for _, h := range hs {
				if !h.IsEmpty() && o != nil && o.Metadata.Root.Equal(h) {
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
