package kv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FileStore(t *testing.T) {
	type TestValue struct {
		Foo string `json:"foo"`
	}

	// Create a new Memory instance
	store, err := NewFileStore[string, TestValue](
		t.TempDir(),
	)
	assert.NoError(t, err)

	// Set a new key-value pair with a struct as value
	testValue := TestValue{Foo: "bar"}
	err = store.Set("structValue", testValue)
	assert.NoError(t, err)

	// Get the value back and check that it matches the original value
	returnedValue, err := store.Get("structValue")
	assert.NoError(t, err)
	assert.Equal(t, testValue, returnedValue)
}

func Test_FileStorePrefix(t *testing.T) {
	type testKey struct {
		Foo string
		Bar string
	}
	type testPair struct {
		key   testKey
		value string
	}
	tests := []struct {
		name    string
		insert  []testPair
		prefix  testKey
		results []string
	}{{
		name: "simple",
		insert: []testPair{{
			key:   testKey{Foo: "foo", Bar: "bar"},
			value: "value-for-foo-bar",
		}, {
			key:   testKey{Foo: "foo", Bar: "baz"},
			value: "value-for-foo-baz",
		}, {
			key:   testKey{Foo: "not-foo", Bar: "qux"},
			value: "value-for-not-foo-qux",
		}},
		prefix: testKey{Foo: "foo"},
		results: []string{
			"value-for-foo-bar",
			"value-for-foo-baz",
		},
	}, {
		name: "simple, includes root",
		insert: []testPair{{
			key:   testKey{Foo: "foo"},
			value: "value-for-foo",
		}, {
			key:   testKey{Foo: "foo", Bar: "bar"},
			value: "value-for-foo-bar",
		}, {
			key:   testKey{Foo: "foo", Bar: "baz"},
			value: "value-for-foo-baz",
		}, {
			key:   testKey{Foo: "not-foo", Bar: "qux"},
			value: "value-for-not-foo-qux",
		}},
		prefix: testKey{Foo: "foo"},
		results: []string{
			"value-for-foo",
			"value-for-foo-bar",
			"value-for-foo-baz",
		},
	}, {
		name: "exact",
		insert: []testPair{{
			key:   testKey{Foo: "foo", Bar: "bar"},
			value: "value-for-foo-bar",
		}, {
			key:   testKey{Foo: "foo", Bar: "baz"},
			value: "value-for-foo-baz",
		}},
		prefix: testKey{Foo: "foo", Bar: "bar"},
		results: []string{
			"value-for-foo-bar",
		},
	}, {
		name: "empty",
		insert: []testPair{{
			key:   testKey{Foo: "foo", Bar: "bar"},
			value: "value-for-foo-bar",
		}, {
			key:   testKey{Foo: "foo", Bar: "baz"},
			value: "value-for-foo-baz",
		}},
		prefix:  testKey{Foo: "not-foo"},
		results: []string{},
	}}
	for _, test := range tests {
		// Create a new Memory instance
		store, err := NewFileStore[testKey, string](
			t.TempDir(),
		)
		assert.NoError(t, err)

		// Set a new key-value pair with a struct as value
		for _, pair := range test.insert {
			err := store.Set(pair.key, pair.value)
			assert.NoError(t, err)
		}

		// Get the value back and check that it matches the original value
		returnedValues, err := store.GetPrefix(test.prefix)
		assert.NoError(t, err)
		assert.Len(t, returnedValues, len(test.results))
		for _, result := range test.results {
			assert.Contains(t, returnedValues, result)
		}
	}
}
