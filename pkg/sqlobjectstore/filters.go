package sqlobjectstore

import (
	"github.com/gobwas/glob"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
)

// FilterOptions
type (
	FilterOption  func(*FilterOptions)
	FilterOptions struct {
		// Filters are used to perform db queries for these filters
		// TODO find a better name for this
		Filters struct {
			ObjectHashes []object.Hash
			StreamHashes []object.Hash
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
			ObjectHashes []object.Hash
			StreamHashes []object.Hash
			ContentTypes []string
			Owners       []crypto.PublicKey
			OrderBy      string
			OrderDir     string
			Limit        *int
			Offset       *int
		}{
			ObjectHashes: []object.Hash{},
			StreamHashes: []object.Hash{},
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

func FilterByHash(h object.Hash) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.ObjectHashes = append(opts.Filters.ObjectHashes, h)
	}
}

func FilterByOwner(h crypto.PublicKey) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.Owners = append(opts.Filters.Owners, h)
	}
}

func FilterByStreamHash(h object.Hash) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.StreamHashes = append(opts.Filters.StreamHashes, h)
	}
}

func FilterByObjectType(typePatterns ...string) FilterOption {
	patterns := make([]glob.Glob, len(typePatterns))
	for i, typePattern := range typePatterns {
		g, err := glob.Compile(typePattern, '.', '/', '#')
		if err != nil {
			panic(errors.Wrap(err, errors.New("invalid pattern")))
		}
		patterns[i] = g
	}
	return func(opts *FilterOptions) {
		opts.Filters.ContentTypes = append(opts.Filters.ContentTypes, typePatterns...)
	}
}
