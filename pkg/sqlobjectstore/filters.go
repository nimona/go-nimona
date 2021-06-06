package sqlobjectstore

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object/value"
)

// FilterOptions
type (
	FilterOption  func(*FilterOptions)
	FilterOptions struct {
		// Filters are used to perform db queries for these filters
		// TODO find a better name for this
		Filters struct {
			ObjectCIDs   []value.CID
			StreamCIDs   []value.CID
			ContentTypes []string
			Owners       []crypto.PublicKey
			OrderBy      string
			OrderDir     string
			Limit        *int
			Offset       *int
		}
	}
)

func newFilterOptions(filterOptions ...FilterOption) FilterOptions {
	options := &FilterOptions{
		Filters: struct {
			ObjectCIDs   []value.CID
			StreamCIDs   []value.CID
			ContentTypes []string
			Owners       []crypto.PublicKey
			OrderBy      string
			OrderDir     string
			Limit        *int
			Offset       *int
		}{
			ObjectCIDs:   []value.CID{},
			StreamCIDs:   []value.CID{},
			ContentTypes: []string{},
			Owners:       []crypto.PublicKey{},
			OrderBy:      "Created",
			OrderDir:     "ASC",
		},
	}
	for _, filterOption := range filterOptions {
		filterOption(options)
	}
	return *options
}

func FilterOrderBy(orderBy string) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.OrderBy = orderBy
	}
}

func FilterOrderDir(orderDir string) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.OrderDir = orderDir
	}
}

func FilterLimit(limit, offset int) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.Limit = &limit
		opts.Filters.Offset = &offset
	}
}

func FilterByCID(hs ...value.CID) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.ObjectCIDs = append(opts.Filters.ObjectCIDs, hs...)
	}
}

func FilterByOwner(hs ...crypto.PublicKey) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.Owners = append(opts.Filters.Owners, hs...)
	}
}

func FilterByStreamCID(hs ...value.CID) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.StreamCIDs = append(opts.Filters.StreamCIDs, hs...)
	}
}

func FilterByObjectType(typePatterns ...string) FilterOption {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(fmt.Errorf("invalid pattern: %w", err))
		}
		patterns[i] = g
	}
	return func(opts *FilterOptions) {
		opts.Filters.ContentTypes = append(opts.Filters.ContentTypes, typePatterns...)
	}
}
