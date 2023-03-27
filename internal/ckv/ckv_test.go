package ckv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPutAndGet(t *testing.T) {
	type TestStruct struct {
		Name  string
		Value int
	}

	tests := []struct {
		name       string
		collection func() *Collection[TestStruct]
		putData    []struct {
			key   []string
			value TestStruct
		}
		getData []struct {
			key      []string
			expected []TestStruct
		}
	}{{
		name: "Basic Put and Get",
		collection: func() *Collection[TestStruct] {
			store := NewMemoryStore()
			return NewCollection(store, "t0", TestStruct{})
		},
		putData: []struct {
			key   []string
			value TestStruct
		}{{
			key: []string{"A"},
			value: TestStruct{
				Name:  "Alice",
				Value: 1,
			},
		}, {
			key: []string{"B"},
			value: TestStruct{
				Name:  "Bob",
				Value: 2,
			},
		}},
		getData: []struct {
			key      []string
			expected []TestStruct
		}{{
			key: []string{"A"},
			expected: []TestStruct{{
				Name:  "Alice",
				Value: 1,
			}},
		}, {
			key: []string{"B"},
			expected: []TestStruct{{
				Name:  "Bob",
				Value: 2,
			}},
		}, {
			key:      []string{"C"},
			expected: nil,
		}},
	}, {
		name: "Partial Key Match",
		collection: func() *Collection[TestStruct] {
			store := NewMemoryStore()
			return NewCollection(store, "t0", TestStruct{})
		},
		putData: []struct {
			key   []string
			value TestStruct
		}{{
			key: []string{"A", "1"},
			value: TestStruct{
				Name:  "Alice",
				Value: 1,
			},
		}, {
			key: []string{"A", "2"},
			value: TestStruct{
				Name:  "Ava",
				Value: 3,
			},
		}, {
			key: []string{"B", "1"},
			value: TestStruct{
				Name:  "Bob",
				Value: 2,
			},
		}},
		getData: []struct {
			key      []string
			expected []TestStruct
		}{{
			key: []string{"A"},
			expected: []TestStruct{{
				Name:  "Alice",
				Value: 1,
			}, {
				Name:  "Ava",
				Value: 3,
			}},
		}, {
			key: []string{"A", "1"},
			expected: []TestStruct{{
				Name:  "Alice",
				Value: 1,
			}},
		}, {
			key: []string{"B", "1"},
			expected: []TestStruct{{
				Name:  "Bob",
				Value: 2,
			}},
		}},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := tt.collection()
			for _, put := range tt.putData {
				collection.Put(put.key, put.value)
			}
			for _, get := range tt.getData {
				result, err := collection.Get(get.key)
				require.NoError(t, err)
				require.Equal(t, get.expected, result)
			}
		})
	}
}
