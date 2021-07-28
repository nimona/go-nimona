package sqlobjectstore

import (
	"fmt"

	"github.com/gobwas/glob"

	"nimona.io/pkg/chore"
	"nimona.io/pkg/did"
)

// FilterOptions
type (
	FilterOption  func(*FilterOptions)
	FilterOptions struct {
		// Filters are used to perform db queries for these filters
		// TODO find a better name for this
		Filters struct {
			ObjectHashes []chore.Hash
			StreamHashes []chore.Hash
			ContentTypes []string
			Owners       []string
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
			ObjectHashes []chore.Hash
			StreamHashes []chore.Hash
			ContentTypes []string
			Owners       []string
			OrderBy      string
			OrderDir     string
			Limit        *int
			Offset       *int
		}{
			ObjectHashes: []chore.Hash{},
			StreamHashes: []chore.Hash{},
			ContentTypes: []string{},
			Owners:       []string{},
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

func FilterByHash(hs ...chore.Hash) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.ObjectHashes = append(opts.Filters.ObjectHashes, hs...)
	}
}

func FilterByOwner(owners ...did.DID) FilterOption {
	hs := []string{}
	for _, owner := range owners {
		hs = append(hs, owner.String())
	}
	return func(opts *FilterOptions) {
		opts.Filters.Owners = append(opts.Filters.Owners, hs...)
	}
}

func FilterByStreamHash(hs ...chore.Hash) FilterOption {
	return func(opts *FilterOptions) {
		opts.Filters.StreamHashes = append(opts.Filters.StreamHashes, hs...)
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
