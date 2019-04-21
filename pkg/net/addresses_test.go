package net

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddresses(t *testing.T) {
	tests := []struct {
		name         string
		addresses    Addresses
		toAdd        []string
		toRemove     []string
		toBlacklist  []string
		expectedList []string
	}{
		{
			name:      "add",
			addresses: Addresses{},
			toAdd: []string{
				"a",
				"b",
			},
			expectedList: []string{
				"a",
				"b",
			},
		},
		{
			name:      "add",
			addresses: Addresses{},
			toAdd: []string{
				"a",
				"b",
			},
			toRemove: []string{
				"a",
				"b",
			},
			expectedList: []string{},
		},
		{
			name:      "add",
			addresses: Addresses{},
			toAdd: []string{
				"a",
				"b",
				"c",
			},
			toRemove: []string{
				"b",
			},
			expectedList: []string{
				"a",
				"c",
			},
		},
		{
			name:      "add",
			addresses: Addresses{},
			toAdd: []string{
				"a",
				"b",
				"c",
			},
			toRemove: []string{
				"b",
			},
			toBlacklist: []string{
				"c",
			},
			expectedList: []string{
				"a",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.addresses.Add(tt.toAdd...)
			tt.addresses.Remove(tt.toRemove...)
			tt.addresses.Blacklist(tt.toBlacklist...)
			aList := tt.addresses.List()
			sort.Strings(tt.expectedList)
			sort.Strings(aList)
			assert.Equal(t, tt.expectedList, aList)
		})
	}
}
